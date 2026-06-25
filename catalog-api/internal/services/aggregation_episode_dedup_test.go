package services

import (
	"context"
	"fmt"
	"os"
	"testing"

	mediamodels "catalogizer/internal/media/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestAggregationEpisodeDedup_NoDuplicatePerSeasonEpisode is the §11.4.115
// RED-on-broken-artifact / §11.4.135 standing regression guard for DEFECT-H
// (duplicate tv_episode rows for a single S01E01).
//
// Root cause (qa-results FINDING_device_ui_data.md FINDING 1): the episode-build
// path is invoked TWICE for the same episode — once via the representative-file
// g.parsed path (aggregation_service.go L624) and again via
// buildTVHierarchyFromFiles (L627), whose internal seenEpisode map is blind to
// the L624 insert — and the tv_episode Create at L1185 had NO existence check, so
// two rows (ids 9 AND 10) were produced for one episode. The fix makes
// buildTVHierarchy idempotent via findExistingEpisode keyed on
// (parent_id=seasonID, episode_number) — the single convergence point both paths
// share.
//
// Dedup key rationale: (parent_id, episode_number), NOT title. DEFECT-I rewrites
// episode titles, so title is not a stable identity key; the (season, episode)
// pair is.
//
// RED_MODE=1 reproduces the PRE-FIX logic: an unconditional episode Create on the
// SECOND invocation (exactly what the old L1185 did) → 2 rows → the
// "exactly ONE episode" assertion FAILS, proving the test catches the defect.
//
// RED_MODE=0 (GREEN, default) drives the REAL fixed buildTVHierarchy twice for the
// same S01E01 (mirroring the L624 + L1040 double-invocation) → the second call
// hits findExistingEpisode and UPDATEs instead of inserting → exactly 1 row.
func TestAggregationEpisodeDedup_NoDuplicatePerSeasonEpisode(t *testing.T) {
	db := setupGranularityDB(t)
	ctx := context.Background()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, zap.NewNop(), itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	// --- arrange: a tv_show with a Season 1 child, no episodes yet -------------
	_, showTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)
	_, epTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_episode")
	require.NoError(t, err)

	showID, err := itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: showTypeID,
		Title:       "Dedup Test Show",
		Status:      "detected",
	})
	require.NoError(t, err)

	seasonNum := 1
	episodeNum := 1
	parsed := ParsedTitle{Title: "Dedup Test Show", Season: &seasonNum, Episode: &episodeNum}

	// First episode-build invocation (mirrors the representative-file g.parsed
	// path at L624). This creates Season 1 + Episode 1 on BOTH code versions.
	svc.buildTVHierarchy(ctx, showID, showTypeID, parsed)

	redMode := os.Getenv("RED_MODE") == "1"
	if redMode {
		// Reproduce the PRE-FIX second invocation: resolve the season the first
		// call created, then do an UNCONDITIONAL Create (the old L1185 logic with
		// no existence check). On the buggy code this is exactly what the second
		// path (buildTVHierarchyFromFiles -> buildTVHierarchy) did, yielding a
		// duplicate row.
		seasons, err := itemRepo.GetChildren(ctx, showID)
		require.NoError(t, err)
		require.Len(t, seasons, 1, "exactly one Season 1 must exist after the first build")
		seasonID := seasons[0].ID

		_, err = itemRepo.Create(ctx, &mediamodels.MediaItem{
			MediaTypeID:   epTypeID,
			Title:         fmt.Sprintf("Episode %d", episodeNum),
			Status:        "detected",
			ParentID:      &seasonID,
			SeasonNumber:  &seasonNum,
			EpisodeNumber: &episodeNum,
		})
		require.NoError(t, err)
	} else {
		// Second episode-build invocation via the REAL fixed path (mirrors the
		// buildTVHierarchyFromFiles -> buildTVHierarchy call at L1040). The fix's
		// findExistingEpisode guard must converge on the existing row.
		svc.buildTVHierarchy(ctx, showID, showTypeID, parsed)
	}

	// --- assert: exactly ONE episode per (season, episode_number) -------------
	seasons, err := itemRepo.GetChildren(ctx, showID)
	require.NoError(t, err)
	require.Len(t, seasons, 1, "season build must stay idempotent (one Season 1)")
	seasonID := seasons[0].ID

	eps, err := itemRepo.GetChildren(ctx, seasonID)
	require.NoError(t, err)
	dupCount := 0
	for _, e := range eps {
		if e.EpisodeNumber != nil && *e.EpisodeNumber == episodeNum {
			dupCount++
		}
	}
	require.Equalf(t, 1, dupCount,
		"exactly ONE tv_episode row must exist for (season=%d, episode=%d); DEFECT-H produced %d (duplicate insert across the two invocation paths)",
		seasonNum, episodeNum, dupCount)
}

// TestAggregationEpisodeDedup_EndToEndNoDuplicates is the GREEN end-to-end guard:
// the REAL AggregateAfterScan (which fires BOTH episode-build paths — L624 and
// L627) over the granularity fixture must produce exactly one episode per
// (season, episode_number) for Breaking Bad. This exercises the real
// double-invocation convergence rather than the direct buildTVHierarchy calls.
func TestAggregationEpisodeDedup_EndToEndNoDuplicates(t *testing.T) {
	db := setupGranularityDB(t)
	ctx := context.Background()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, zap.NewNop(), itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := seedGranularityFixture(t, db)
	require.NoError(t, svc.AggregateAfterScan(ctx, rootID))

	_, showTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)
	show, err := itemRepo.GetByTitle(ctx, "Breaking Bad", showTypeID)
	require.NoError(t, err)
	require.NotNil(t, show, "Breaking Bad tv_show must exist")

	seasons, err := itemRepo.GetChildren(ctx, show.ID)
	require.NoError(t, err)
	for _, season := range seasons {
		eps, err := itemRepo.GetChildren(ctx, season.ID)
		require.NoError(t, err)
		seen := map[int]int{}
		for _, e := range eps {
			if e.EpisodeNumber != nil {
				seen[*e.EpisodeNumber]++
			}
		}
		for epNum, cnt := range seen {
			require.Equalf(t, 1, cnt,
				"season %v episode %d must have exactly one row (DEFECT-H duplicate); got %d",
				ptrIntVal(season.SeasonNumber), epNum, cnt)
		}
	}
}

func ptrIntVal(p *int) int {
	if p == nil {
		return -1
	}
	return *p
}
