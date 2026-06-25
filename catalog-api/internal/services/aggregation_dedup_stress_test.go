package services

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"catalogizer/database"
	"catalogizer/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// §11.4.85 stress + chaos coverage for DEFECT-H (episode dedup key
// (parent_id, episode_number)). The companion file
// aggregation_episode_dedup_test.go is the §11.4.115 RED/GREEN polarity guard
// proving the single-pass fix; THIS file proves the SAME invariant holds under
// (a) sustained idempotent re-aggregation and (b) concurrent contention.
//
// Both tests drive the REAL AggregateAfterScan over the granularity fixture via
// the in-memory shared-cache SQLite harness (setupGranularityDB +
// seedGranularityFixture). No live server, no live :8080 DB, no network: the
// TMDB enrichment goroutine early-returns because TMDB_API_KEY is unset in the
// test environment, so AggregateAfterScan is fully self-contained.
//
// Invariant under both regimes: for every (season_id parent, episode_number)
// pair under every season of every tv_show, exactly ONE tv_episode row exists.
// A duplicate (count > 1) is the DEFECT-H regression.

// countEpisodeDuplicates walks every tv_show -> season -> episode chain and
// returns the worst (max) per-(season, episode_number) row count plus a
// human-readable census. A return of 1 means perfectly deduped; >1 means a
// DEFECT-H duplicate exists. This is the captured-evidence primitive both
// regimes assert against — a PASS is backed by these real counts, not merely by
// absence-of-error.
func countEpisodeDuplicates(t *testing.T, db *database.DB) (worst int, census string) {
	t.Helper()
	ctx := context.Background()
	itemRepo := repository.NewMediaItemRepository(db)

	_, showTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)
	show, err := itemRepo.GetByTitle(ctx, "Breaking Bad", showTypeID)
	require.NoError(t, err)
	require.NotNil(t, show, "Breaking Bad tv_show must exist after aggregation")

	worst = 0
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
			if cnt > worst {
				worst = cnt
			}
			census += fmt.Sprintf("S%v E%d=%d ", ptrIntVal(season.SeasonNumber), epNum, cnt)
		}
	}
	return worst, census
}

// TestAggregationDedupStress_RepeatedReaggregationIdempotent is the §11.4.85
// STRESS leg: it calls AggregateAfterScan N times in a row over the SAME seeded
// files and asserts, PER ITERATION, that (1) the tv_episode count stays constant
// and (2) no (parent_id, episode_number) ever duplicates. The first run
// establishes the baseline count; every subsequent re-aggregation must converge
// on the identical count (idempotent), which is exactly the property the
// DEFECT-H fix introduced — without it, repeated runs would accrete duplicate
// episode rows.
func TestAggregationDedupStress_RepeatedReaggregationIdempotent(t *testing.T) {
	db := setupGranularityDB(t)
	ctx := context.Background()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, zap.NewNop(), itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := seedGranularityFixture(t, db)

	// countEpisodes returns the total number of tv_episode rows across all shows
	// (the corpus-wide invariant that must stay flat under re-aggregation).
	countEpisodes := func() int {
		var c int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM media_items mi
			 JOIN media_types mt ON mt.id = mi.media_type_id
			 WHERE mt.name = 'tv_episode'`).Scan(&c))
		return c
	}

	const iterations = 25 // §11.4.85: N >= 20 sustained-load floor
	var baseline int
	type sample struct {
		iter     int
		episodes int
		worstDup int
	}
	evidence := make([]sample, 0, iterations)

	for i := 1; i <= iterations; i++ {
		require.NoErrorf(t, svc.AggregateAfterScan(ctx, rootID),
			"AggregateAfterScan must not error on iteration %d", i)

		got := countEpisodes()
		worst, census := countEpisodeDuplicates(t, db)
		evidence = append(evidence, sample{iter: i, episodes: got, worstDup: worst})

		// Per-iteration captured-evidence assertions (§11.4.85 sustained load).
		require.Equalf(t, 1, worst,
			"iteration %d: every (season, episode_number) must be unique (DEFECT-H dedup); census=[%s]",
			i, census)
		require.Greaterf(t, got, 0, "iteration %d: episodes must exist (sanity)", i)

		if i == 1 {
			baseline = got
		} else {
			require.Equalf(t, baseline, got,
				"iteration %d: re-aggregation must be IDEMPOTENT — episode count drifted from baseline %d to %d (DEFECT-H accretion)",
				i, baseline, got)
		}
	}

	// Emit the per-iteration evidence so a PASS is backed by real captured data.
	t.Logf("§11.4.85 STRESS evidence: baseline=%d episodes over %d iterations", baseline, iterations)
	for _, s := range evidence {
		t.Logf("  iter %2d: tv_episode_count=%d  worst_dup_per_(season,episode)=%d", s.iter, s.episodes, s.worstDup)
	}
	require.Len(t, evidence, iterations, "every iteration must have produced a captured-evidence sample")
}

// TestAggregationDedupStress_ConcurrentReaggregation is the §11.4.85 CHAOS /
// CONTENTION leg: it fires N concurrent AggregateAfterScan goroutines on the
// SAME storage root and asserts (1) no goroutine panics, (2) no deadlock (the
// whole barrier completes), and (3) AFTERWARD no (parent_id, episode_number)
// duplicate exists.
//
// HONEST NOTE on the invariant asserted (§11.4.6): AggregateAfterScan is NOT
// declared safe for unsynchronised concurrent invocation of the SAME root by
// design — there is no application-level mutex around the read-group-write
// sequence, so two goroutines can interleave their getMediaLeafFiles ->
// processTitleGroup phases. We therefore do NOT assert a property the code does
// not provide (e.g. "exactly the single-pass count regardless of races", or
// "every goroutine returns nil"). What the code DOES provide, and what we assert:
//   - the per-episode upsert is keyed on (parent_id, episode_number) via
//     findExistingEpisode, and the shared-cache SQLite layer serialises the
//     actual writes, so the CONVERGED end-state must still hold the dedup
//     invariant: zero duplicate (season, episode) rows once all goroutines
//     have quiesced. This is the user-visible guarantee that matters (the
//     catalog never shows two S01E01 rows), and it is what DEFECT-H broke.
//   - the goroutines must not panic and the barrier must complete (no deadlock).
//
// A per-goroutine error is recorded but tolerated: under genuine write
// contention SQLite may return a transient "database is locked"/"busy" to a
// loser, which is a legitimate outcome of unsynchronised concurrency — NOT a
// dedup-correctness failure. We surface any such errors as captured evidence
// (t.Logf) for honesty rather than asserting them away.
func TestAggregationDedupStress_ConcurrentReaggregation(t *testing.T) {
	db := setupGranularityDB(t)
	ctx := context.Background()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, zap.NewNop(), itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := seedGranularityFixture(t, db)

	// Establish the deduped state once, single-threaded, so the show/season tree
	// exists before the concurrent storm (the storm then re-aggregates it).
	require.NoError(t, svc.AggregateAfterScan(ctx, rootID))
	worst0, census0 := countEpisodeDuplicates(t, db)
	require.Equalf(t, 1, worst0, "pre-storm baseline must already be deduped; census=[%s]", census0)

	const goroutines = 8 // §11.4.85: N >= 8 concurrent-contention floor

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errs   []error
		panics []interface{}
		start  = make(chan struct{}) // release barrier => maximise overlap
	)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					panics = append(panics, r)
					mu.Unlock()
				}
			}()
			<-start // all goroutines block here, then fire together
			if err := svc.AggregateAfterScan(ctx, rootID); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("goroutine %d: %w", id, err))
				mu.Unlock()
			}
		}(i)
	}

	close(start) // unleash the contention storm
	wg.Wait()    // if this never returns, the test deadlocks the runner => FAIL

	// (1) No panic under contention — a panic is always a real defect, never a
	// tolerable concurrency outcome.
	require.Emptyf(t, panics, "AggregateAfterScan must not panic under %d-way contention; panics=%v", goroutines, panics)

	// (2) Captured evidence: surface (do NOT assert away) any transient
	// per-goroutine errors — under unsynchronised SQLite write contention a loser
	// may see a busy/locked error, which is honest concurrency, not a dedup bug.
	t.Logf("§11.4.85 CHAOS evidence: %d goroutines fired; %d returned a (tolerated) contention error", goroutines, len(errs))
	for _, e := range errs {
		t.Logf("  tolerated contention error: %v", e)
	}

	// (3) The load-bearing user-visible invariant: after the storm quiesces,
	// the dedup key (parent_id, episode_number) still holds — zero duplicates.
	worst, census := countEpisodeDuplicates(t, db)
	require.Equalf(t, 1, worst,
		"after %d-way concurrent re-aggregation, every (season, episode_number) must remain unique (DEFECT-H dedup under contention); census=[%s]",
		goroutines, census)
	t.Logf("§11.4.85 CHAOS post-storm census (worst_dup=%d): [%s]", worst, census)
}
