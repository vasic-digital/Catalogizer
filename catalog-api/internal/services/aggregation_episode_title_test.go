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

// TestAggregationEpisodeTitle_UsesParsedHumanTitle is the §11.4.115
// RED-on-broken-artifact / §11.4.135 standing regression guard for DEFECT-I
// Layer B (real episode titles discarded by the hardcoded "Episode %d").
//
// Root cause (qa-results FINDING_device_ui_data.md FINDING 2): buildTVHierarchy
// hardcoded the episode title as fmt.Sprintf("Episode %d", *parsed.Episode),
// discarding the real title parsed by digital.vasic.entities/pkg/parser into
// ParsedTitle.EpisodeTitle (Layer A). So "Breaking.Bad.S01E01.Pilot.mkv"
// (EpisodeTitle="Pilot") displayed as "Episode 1" on the device UI.
//
// The fix sets the episode Title to parsed.EpisodeTitle when non-empty, and
// falls back to "Episode %d" ONLY when the filename carries no human title
// (EpisodeTitle empty). The dedup key remains the STABLE (parent_id,
// episode_number) pair (DEFECT-H) — title is display-only, never identity.
//
// The §11.4.146 case-space is covered in one test: a TITLED episode
// (S01E01.Pilot -> "Pilot") AND an UNTITLED episode (S01E02 -> "Episode 2")
// driven through the same real buildTVHierarchy path.
//
// RED_MODE=1 reproduces the PRE-FIX logic: it forces the episode title to the
// "Episode %d" hardcode regardless of the parsed EpisodeTitle (exactly what the
// old code did). On the buggy code this is what shipped, so the "Pilot"
// assertion below FAILS — proving the test catches the defect (no tautology).
//
// RED_MODE=0 (GREEN, default) drives the REAL fixed buildTVHierarchy, which must
// title the S01E01 episode "Pilot" while still titling the untitled S01E02
// "Episode 2".
func TestAggregationEpisodeTitle_UsesParsedHumanTitle(t *testing.T) {
	db := setupGranularityDB(t)
	ctx := context.Background()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, zap.NewNop(), itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	_, showTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)
	_, epTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_episode")
	require.NoError(t, err)

	showID, err := itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: showTypeID,
		Title:       "Breaking Bad",
		Status:      "detected",
	})
	require.NoError(t, err)

	// The TITLED case (S01E01.Pilot -> EpisodeTitle="Pilot") and the UNTITLED
	// case (S01E02 -> EpisodeTitle=""), derived from the live fixture filenames
	// via the REAL entities parser (digital.vasic.entities/pkg/parser).
	titledParsed := ParseTVShow("Breaking.Bad.S01E01.Pilot.mkv")
	require.NotNil(t, titledParsed.Season, "S01E01 must parse a season")
	require.NotNil(t, titledParsed.Episode, "S01E01 must parse an episode")
	require.Equal(t, "Pilot", titledParsed.EpisodeTitle,
		"Layer A must populate EpisodeTitle=Pilot for the titled fixture filename")

	untitledParsed := ParseTVShow("Breaking.Bad.S01E02.mkv")
	require.NotNil(t, untitledParsed.Season, "S01E02 must parse a season")
	require.NotNil(t, untitledParsed.Episode, "S01E02 must parse an episode")
	require.Equal(t, "", untitledParsed.EpisodeTitle,
		"the untitled fixture filename must yield an empty EpisodeTitle (no human title)")

	redMode := os.Getenv("RED_MODE") == "1"
	if redMode {
		// Reproduce the PRE-FIX logic: strip the parsed human title so the
		// build path falls back to the "Episode %d" hardcode exactly as the old
		// code did. On the buggy code this is what shipped for EVERY episode,
		// so the "Pilot" assertion below FAILS.
		titledParsed.EpisodeTitle = ""
		untitledParsed.EpisodeTitle = ""
	}

	// Drive the REAL fixed buildTVHierarchy for both episodes.
	svc.buildTVHierarchy(ctx, showID, showTypeID, titledParsed)
	svc.buildTVHierarchy(ctx, showID, showTypeID, untitledParsed)

	// --- assert: titled episode carries its human title ----------------------
	gotTitled := episodeTitleFor(t, itemRepo, showID, epTypeID, *titledParsed.Season, *titledParsed.Episode)
	require.Equalf(t, "Pilot", gotTitled,
		"S01E01 must display its real human title 'Pilot' (DEFECT-I: hardcode showed %q)", gotTitled)

	// --- assert: untitled episode falls back to the "Episode N" placeholder ---
	gotUntitled := episodeTitleFor(t, itemRepo, showID, epTypeID, *untitledParsed.Season, *untitledParsed.Episode)
	require.Equalf(t, fmt.Sprintf("Episode %d", *untitledParsed.Episode), gotUntitled,
		"S01E02 carries no human title so it must fall back to 'Episode %d'; got %q",
		*untitledParsed.Episode, gotUntitled)
}

// episodeTitleFor resolves the title of the tv_episode child of showID with the
// given (season, episode) number. It walks show -> season -> episode (the real
// hierarchy buildTVHierarchy creates) and fails the test if the episode is
// absent.
func episodeTitleFor(t *testing.T, itemRepo *repository.MediaItemRepository, showID, epTypeID int64, seasonNum, episodeNum int) string {
	t.Helper()
	ctx := context.Background()

	seasons, err := itemRepo.GetChildren(ctx, showID)
	require.NoError(t, err)
	for _, season := range seasons {
		if season.SeasonNumber == nil || *season.SeasonNumber != seasonNum {
			continue
		}
		eps, err := itemRepo.GetChildren(ctx, season.ID)
		require.NoError(t, err)
		for _, ep := range eps {
			if ep.MediaTypeID == epTypeID && ep.EpisodeNumber != nil && *ep.EpisodeNumber == episodeNum {
				return ep.Title
			}
		}
	}
	require.FailNowf(t, "episode not found",
		"no tv_episode for (season=%d, episode=%d) under show %d", seasonNum, episodeNum, showID)
	return ""
}
