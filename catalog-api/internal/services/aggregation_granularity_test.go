package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/filesystem"
	mediamodels "catalogizer/internal/media/models"
	"catalogizer/models"
	"catalogizer/repository"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// granularityDBCounter gives each test a uniquely-named shared-cache in-memory DB.
var granularityDBCounter int

// setupGranularityDB builds a real in-memory SQLite DB with the full migration
// chain applied (mirrors setupScannerE2EDB). Shared-cache mode is required because
// insertFileRecord opens nested BeginTx/QueryRow operations needing concurrent
// connections.
func setupGranularityDB(t *testing.T) *database.DB {
	t.Helper()
	granularityDBCounter++
	dsn := fmt.Sprintf("file:agggranularity%d?mode=memory&cache=shared&_foreign_keys=1", granularityDBCounter)
	sqlDB, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	db := database.WrapDB(sqlDB, database.DialectSQLite)
	require.NoError(t, db.RunMigrations(context.Background()))
	return db
}

// driveGranularityScan replays exactly what a real protocol scanner does — walks
// the fake tree, computes fullPath := filepath.Join(path, name), and calls the
// REAL insertFileRecord. This exercises the real ensureDirectoryPathExists +
// parent_id arithmetic (incl. its known parenting bug) without a live server, so
// the DB state faithfully reproduces a real share scan.
func driveGranularityScan(t *testing.T, db *database.DB, job ScanJob, status *ScanStatus, path string, nodes []fakeScanNode) {
	t.Helper()
	for _, n := range nodes {
		fullPath := filepath.Join(path, n.name)
		fi := &filesystem.FileInfo{
			Name:    n.name,
			Size:    n.size,
			IsDir:   n.isDir,
			ModTime: time.Now(),
		}
		require.NoError(t, insertFileRecord(context.Background(), db, fullPath, fi, job, status, zap.NewNop()))
		if n.isDir {
			driveGranularityScan(t, db, job, status, fullPath, n.children)
		}
	}
}

// granularityFixtureTree is the EXACT topology from the live SMB fixture
// (data/catalogizer.db) recorded in FINDING_aggregation_granularity.md: a
// CATEGORY directory (movies/, series/, music/, software/) at the top level,
// per-title files (and per-show / per-album subtrees) underneath.
func granularityFixtureTree() []fakeScanNode {
	const vid = int64(3_000_000_000)
	const aud = int64(40_000_000)
	const sw = int64(50_000_000)
	return []fakeScanNode{
		{name: "movies", isDir: true, children: []fakeScanNode{
			{name: "Inception.2010.720p.BluRay.mp4", size: vid},
			{name: "Interstellar.2014.2160p.UHD.mkv", size: vid},
			{name: "The.Matrix.1999.1080p.BluRay.mkv", size: vid},
		}},
		{name: "series", isDir: true, children: []fakeScanNode{
			{name: "Breaking Bad", isDir: true, children: []fakeScanNode{
				{name: "Season 01", isDir: true, children: []fakeScanNode{
					{name: "Breaking.Bad.S01E01.Pilot.mkv", size: vid},
					{name: "Breaking.Bad.S01E02.mkv", size: vid},
				}},
				{name: "Season 02", isDir: true, children: []fakeScanNode{
					{name: "Breaking.Bad.S02E01.mkv", size: vid},
				}},
			}},
		}},
		{name: "music", isDir: true, children: []fakeScanNode{
			{name: "Pink Floyd", isDir: true, children: []fakeScanNode{
				{name: "The Dark Side of the Moon", isDir: true, children: []fakeScanNode{
					{name: "01 - Speak to Me.flac", size: aud},
					{name: "02 - Breathe.flac", size: aud},
				}},
			}},
			{name: "Artist - Song Title.mp3", size: aud},
			{name: "Another Artist - Track Name.flac", size: aud},
		}},
		{name: "software", isDir: true, children: []fakeScanNode{
			{name: "app-installer-v1.0.exe", size: sw},
			{name: "tool-setup-2.3.1.dmg", size: sw},
		}},
	}
}

// seedGranularityFixture runs the real scanner persistence over the fixture tree
// and returns the storage-root id for AggregateAfterScan.
func seedGranularityFixture(t *testing.T, db *database.DB) int64 {
	t.Helper()
	ctx := context.Background()
	rootID, err := db.InsertReturningID(ctx,
		`INSERT INTO storage_roots (name, protocol, path, enabled, max_depth) VALUES (?, 'smb', '/', 1, 10)`,
		"granularityshare")
	require.NoError(t, err)

	job := ScanJob{
		ID:          "job-granularity",
		StorageRoot: &models.StorageRoot{ID: rootID, Name: "granularityshare", Protocol: "smb"},
		Path:        ".",
		MaxDepth:    10,
		Context:     ctx,
	}
	status := &ScanStatus{JobID: job.ID, StartTime: time.Now(), Status: "running"}
	driveGranularityScan(t, db, job, status, job.Path, granularityFixtureTree())

	var fileCount int
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM files WHERE storage_root_id = ? AND deleted = 0`, rootID).Scan(&fileCount))
	require.Greater(t, fileCount, 0, "scanner must have indexed fixture files")
	return rootID
}

// countEntitiesTitled returns how many media_items rows carry one of the given titles.
func countEntitiesTitled(t *testing.T, db *database.DB, titles ...string) int {
	t.Helper()
	ctx := context.Background()
	total := 0
	for _, title := range titles {
		var c int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM media_items WHERE title = ?`, title).Scan(&c))
		total += c
	}
	return total
}

// TestAggregationGranularity_PerTitleEntities is the §11.4.115 RED-on-broken /
// §11.4.135 standing regression guard for DEFECT-E (aggregation granularity).
//
// RED_MODE=1 runs the PRE-FIX, top-level-directory-granular aggregation
// (aggregateAfterScanLegacy) and MUST FAIL: the live defect emits one entity per
// top-level dir titled "movies"/"music"/"software" and drops series/ entirely,
// so "Inception"/"Breaking Bad"/"The Dark Side of the Moon" never exist and the
// per-title assertions below fail.
//
// RED_MODE=0 (default) runs the FIXED title-granular AggregateAfterScan and is
// the GREEN guard: per-title movie entities, a browsable Breaking Bad tv_show
// with season/episode children, a per-album music entity, and ZERO
// category-folder entities.
func TestAggregationGranularity_PerTitleEntities(t *testing.T) {
	db := setupGranularityDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	svc := NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := seedGranularityFixture(t, db)

	redMode := os.Getenv("RED_MODE") == "1"
	if redMode {
		// Reproduce the defect on the pre-fix logic path.
		require.NoError(t, svc.aggregateAfterScanLegacy(ctx, rootID))
	} else {
		require.NoError(t, svc.AggregateAfterScan(ctx, rootID))
	}

	// --- A: movies become DISTINCT per-title entities -------------------------
	_, movieTypeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)
	for _, want := range []struct {
		title string
		year  int
	}{
		{"Inception", 2010},
		{"Interstellar", 2014},
		{"The Matrix", 1999},
	} {
		mi, err := itemRepo.GetByTitle(ctx, want.title, movieTypeID)
		require.NoError(t, err)
		require.NotNilf(t, mi, "movie entity %q must exist (defect: only a 'movies' folder entity is created)", want.title)
		require.NotNilf(t, mi.Year, "movie %q must carry a parsed year", want.title)
		require.Equalf(t, want.year, *mi.Year, "movie %q year", want.title)
	}

	// --- B: Breaking Bad is a browsable tv_show with season+episode children --
	_, showTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)
	show, err := itemRepo.GetByTitle(ctx, "Breaking Bad", showTypeID)
	require.NoError(t, err)
	require.NotNil(t, show, "Breaking Bad tv_show entity must exist (defect: series/ produces NO entity at all)")

	seasons, err := itemRepo.GetChildren(ctx, show.ID)
	require.NoError(t, err)
	seasonByNum := map[int]*mediamodels.MediaItem{}
	for _, s := range seasons {
		require.NotNil(t, s.SeasonNumber, "season entity must carry a season number")
		seasonByNum[*s.SeasonNumber] = s
	}
	require.Containsf(t, seasonByNum, 1, "Breaking Bad must have a Season 1 child")
	require.Containsf(t, seasonByNum, 2, "Breaking Bad must have a Season 2 child")

	// Episodes hang under their season.
	s1Eps, err := itemRepo.GetChildren(ctx, seasonByNum[1].ID)
	require.NoError(t, err)
	require.GreaterOrEqualf(t, len(s1Eps), 2, "Season 1 must expose its episodes (S01E01, S01E02)")
	s2Eps, err := itemRepo.GetChildren(ctx, seasonByNum[2].ID)
	require.NoError(t, err)
	require.GreaterOrEqualf(t, len(s2Eps), 1, "Season 2 must expose its episode (S02E01)")

	// --- C: album granularity (per-album, not one swallowing "music") ---------
	_, albumTypeID, err := itemRepo.GetMediaTypeByName(ctx, "music_album")
	require.NoError(t, err)
	album, err := itemRepo.GetByTitle(ctx, "The Dark Side of the Moon", albumTypeID)
	require.NoError(t, err)
	require.NotNil(t, album, "music_album 'The Dark Side of the Moon' must exist (defect: one 'music' entity swallows everything)")

	// --- D: NO category-folder entities ---------------------------------------
	bogus := countEntitiesTitled(t, db, "movies", "series", "music", "software")
	require.Equalf(t, 0, bogus, "no entity may be titled after a category folder (movies/series/music/software); found %d", bogus)
}
