package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/services"
	"catalogizer/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// aggregationTestDBCounter provides unique names for shared-cache in-memory databases.
var aggregationTestDBCounter int

// setupAggregationTestDB creates an in-memory SQLite database with all
// migrations applied and returns the wrapped database.DB ready for use.
// Uses shared-cache mode so multiple connections share the same in-memory
// database, which is required by the aggregation service (it opens nested
// queries that need concurrent connections).
func setupAggregationTestDB(t *testing.T) *database.DB {
	t.Helper()
	aggregationTestDBCounter++
	dsn := fmt.Sprintf("file:aggtest%d?mode=memory&cache=shared&_foreign_keys=1", aggregationTestDBCounter)
	sqlDB, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	// Run the full migration chain so all tables exist
	ctx := context.Background()
	require.NoError(t, db.RunMigrations(ctx))

	t.Cleanup(func() { sqlDB.Close() })
	return db
}

// insertStorageRoot inserts a storage root and returns its ID.
func insertStorageRoot(t *testing.T, db *database.DB, name, protocol string) int64 {
	t.Helper()
	id, err := db.InsertReturningID(context.Background(),
		`INSERT INTO storage_roots (name, protocol, path, enabled) VALUES (?, ?, ?, 1)`,
		name, protocol, "/media")
	require.NoError(t, err)
	return id
}

// insertFileRecord inserts a file record and returns its ID.
func insertFileRecord(t *testing.T, db *database.DB, storageRootID, parentID int64, path, name, ext string, size int64, isDir bool) int64 {
	t.Helper()
	var parentVal interface{}
	if parentID > 0 {
		parentVal = parentID
	}

	id, err := db.InsertReturningID(context.Background(),
		`INSERT INTO files (storage_root_id, path, name, extension, size, is_directory, deleted, modified_at, parent_id)
		 VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		storageRootID, path, name, ext, size, isDir, time.Now(), parentVal)
	require.NoError(t, err)
	return id
}

// TestAggregationPipeline_MovieScan validates the full pipeline:
// scan files -> aggregate -> verify entities and links.
func TestAggregationPipeline_MovieScan(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	// Create repositories
	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	// Setup: storage root + movie directory with video files
	rootID := insertStorageRoot(t, db, "movies-nas", "smb")

	// Top-level directory: "The Matrix (1999)"
	matrixDirID := insertFileRecord(t, db, rootID, 0,
		"/media/movies/The Matrix (1999)", "The Matrix (1999)", "", 0, true)

	// Child files in that directory
	insertFileRecord(t, db, rootID, matrixDirID,
		"/media/movies/The Matrix (1999)/The.Matrix.1999.1080p.mkv",
		"The.Matrix.1999.1080p.mkv", ".mkv", 4_000_000_000, false)

	insertFileRecord(t, db, rootID, matrixDirID,
		"/media/movies/The Matrix (1999)/The.Matrix.1999.srt",
		"The.Matrix.1999.srt", ".srt", 50_000, false)

	// Run aggregation
	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	// Verify: media_items was created for "The Matrix"
	count, err := itemRepo.Count(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1), "at least one media item should exist")

	// Verify: the movie type was detected
	_, movieTypeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	item, err := itemRepo.GetByTitle(ctx, "The Matrix", movieTypeID)
	require.NoError(t, err)
	require.NotNil(t, item, "The Matrix media item should exist")
	assert.Equal(t, "The Matrix", item.Title)
	assert.NotNil(t, item.Year)
	assert.Equal(t, 1999, *item.Year)
	assert.Equal(t, "detected", item.Status)

	// Verify: media_files links exist (mkv + srt linked to the item)
	files, err := fileRepo.GetFilesByItem(ctx, item.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1, "files should be linked to The Matrix")

	// Verify: first linked file is marked as primary
	hasPrimary := false
	for _, f := range files {
		if f.IsPrimary {
			hasPrimary = true
			break
		}
	}
	assert.True(t, hasPrimary, "one file should be marked as primary")
}

// TestAggregationPipeline_TVShowHierarchy validates that TV shows create
// a parent-child hierarchy: show -> season -> episode.
func TestAggregationPipeline_TVShowHierarchy(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := insertStorageRoot(t, db, "tv-nas", "smb")

	// Top-level directory: "Breaking Bad S01E01"
	bbDirID := insertFileRecord(t, db, rootID, 0,
		"/media/tv/Breaking Bad S01E01", "Breaking Bad S01E01", "", 0, true)

	insertFileRecord(t, db, rootID, bbDirID,
		"/media/tv/Breaking Bad S01E01/Breaking.Bad.S01E01.mkv",
		"Breaking.Bad.S01E01.mkv", ".mkv", 1_500_000_000, false)

	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	// Verify: tv_show entity was created
	_, tvShowTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)

	show, err := itemRepo.GetByTitle(ctx, "Breaking Bad", tvShowTypeID)
	require.NoError(t, err)
	require.NotNil(t, show, "Breaking Bad tv_show entity should exist")
	assert.Equal(t, "Breaking Bad", show.Title)
	assert.Nil(t, show.ParentID, "top-level show should have no parent")

	// Verify: season child was created under the show
	_, tvSeasonTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_season")
	require.NoError(t, err)

	season, err := itemRepo.GetByTitle(ctx, "Season 1", tvSeasonTypeID)
	require.NoError(t, err)
	require.NotNil(t, season, "Season 1 entity should exist")
	assert.NotNil(t, season.ParentID)
	assert.Equal(t, show.ID, *season.ParentID, "season should be child of show")
	assert.NotNil(t, season.SeasonNumber)
	assert.Equal(t, 1, *season.SeasonNumber)

	// Verify hierarchy via GetChildren
	children, err := itemRepo.GetChildren(ctx, show.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(children), 1, "show should have at least one season child")

	// Verify: episode child was created under the season
	_, tvEpTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_episode")
	require.NoError(t, err)

	ep, err := itemRepo.GetByTitle(ctx, "Episode 1", tvEpTypeID)
	require.NoError(t, err)
	require.NotNil(t, ep, "Episode 1 entity should exist")
	assert.NotNil(t, ep.ParentID)
	assert.Equal(t, season.ID, *ep.ParentID, "episode should be child of season")
	assert.NotNil(t, ep.EpisodeNumber)
	assert.Equal(t, 1, *ep.EpisodeNumber)
}

// TestAggregationPipeline_MusicAlbum verifies music directory aggregation.
func TestAggregationPipeline_MusicAlbum(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := insertStorageRoot(t, db, "music-nas", "smb")

	// Top-level directory: music album with audio files only
	albumDirID := insertFileRecord(t, db, rootID, 0,
		"/media/music/Pink Floyd - The Wall (1979)",
		"Pink Floyd - The Wall (1979)", "", 0, true)

	insertFileRecord(t, db, rootID, albumDirID,
		"/media/music/Pink Floyd - The Wall (1979)/01 - In the Flesh.flac",
		"01 - In the Flesh.flac", ".flac", 30_000_000, false)

	insertFileRecord(t, db, rootID, albumDirID,
		"/media/music/Pink Floyd - The Wall (1979)/02 - The Thin Ice.flac",
		"02 - The Thin Ice.flac", ".flac", 25_000_000, false)

	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	_, albumTypeID, err := itemRepo.GetMediaTypeByName(ctx, "music_album")
	require.NoError(t, err)

	// The title parser extracts "The Wall" as the album title from "Pink Floyd - The Wall (1979)"
	// Search for any music_album items to verify one was created
	items, total, err := itemRepo.GetByType(ctx, albumTypeID, 10, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1), "at least one music album should exist")
	assert.GreaterOrEqual(t, len(items), 1)

	// Verify files were linked
	files, err := fileRepo.GetFilesByItem(ctx, items[0].ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1, "audio files should be linked to the album")
}

// TestAggregationPipeline_EmptyDirectorySkipped verifies that directories
// with no child files are skipped during aggregation.
func TestAggregationPipeline_EmptyDirectorySkipped(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := insertStorageRoot(t, db, "empty-nas", "smb")

	// Directory with no child files
	insertFileRecord(t, db, rootID, 0,
		"/media/empty/SomeDir", "SomeDir", "", 0, true)

	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	count, err := itemRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "no media items should be created for an empty directory")
}

// TestAggregationPipeline_MultipleMovies verifies that multiple directories
// each produce separate media items.
func TestAggregationPipeline_MultipleMovies(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := insertStorageRoot(t, db, "multi-nas", "smb")

	// Movie 1
	dir1 := insertFileRecord(t, db, rootID, 0,
		"/media/movies/Inception (2010)", "Inception (2010)", "", 0, true)
	insertFileRecord(t, db, rootID, dir1,
		"/media/movies/Inception (2010)/Inception.2010.mkv",
		"Inception.2010.mkv", ".mkv", 3_000_000_000, false)

	// Movie 2
	dir2 := insertFileRecord(t, db, rootID, 0,
		"/media/movies/Interstellar (2014)", "Interstellar (2014)", "", 0, true)
	insertFileRecord(t, db, rootID, dir2,
		"/media/movies/Interstellar (2014)/Interstellar.2014.mkv",
		"Interstellar.2014.mkv", ".mkv", 5_000_000_000, false)

	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	_, movieTypeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	inception, err := itemRepo.GetByTitle(ctx, "Inception", movieTypeID)
	require.NoError(t, err)
	require.NotNil(t, inception)
	assert.Equal(t, 2010, *inception.Year)

	interstellar, err := itemRepo.GetByTitle(ctx, "Interstellar", movieTypeID)
	require.NoError(t, err)
	require.NotNil(t, interstellar)
	assert.Equal(t, 2014, *interstellar.Year)

	// Both should be distinct entities
	assert.NotEqual(t, inception.ID, interstellar.ID)
}

// TestAggregationPipeline_IdempotentRerun verifies that running aggregation
// twice on the same data does not duplicate entities.
func TestAggregationPipeline_IdempotentRerun(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := insertStorageRoot(t, db, "idempotent-nas", "smb")

	dirID := insertFileRecord(t, db, rootID, 0,
		"/media/movies/Blade Runner (1982)", "Blade Runner (1982)", "", 0, true)
	insertFileRecord(t, db, rootID, dirID,
		"/media/movies/Blade Runner (1982)/Blade.Runner.1982.mkv",
		"Blade.Runner.1982.mkv", ".mkv", 2_000_000_000, false)

	// First run
	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	countAfterFirst, err := itemRepo.Count(ctx)
	require.NoError(t, err)

	// Second run
	err = svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	countAfterSecond, err := itemRepo.Count(ctx)
	require.NoError(t, err)

	assert.Equal(t, countAfterFirst, countAfterSecond,
		"re-running aggregation should not duplicate entities")
}

// TestAggregationPipeline_MediaTypeCorrectness verifies that different
// media types are correctly assigned based on file extensions and directory names.
func TestAggregationPipeline_MediaTypeCorrectness(t *testing.T) {
	db := setupAggregationTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := services.NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	rootID := insertStorageRoot(t, db, "mixed-nas", "smb")

	// Movie directory (video files)
	movieDir := insertFileRecord(t, db, rootID, 0,
		"/media/Gladiator (2000)", "Gladiator (2000)", "", 0, true)
	insertFileRecord(t, db, rootID, movieDir,
		"/media/Gladiator (2000)/Gladiator.mp4",
		"Gladiator.mp4", ".mp4", 2_000_000_000, false)

	// TV show directory
	tvDir := insertFileRecord(t, db, rootID, 0,
		"/media/The Wire S02E03", "The Wire S02E03", "", 0, true)
	insertFileRecord(t, db, rootID, tvDir,
		"/media/The Wire S02E03/The.Wire.S02E03.mkv",
		"The.Wire.S02E03.mkv", ".mkv", 1_000_000_000, false)

	err := svc.AggregateAfterScan(ctx, rootID)
	require.NoError(t, err)

	// Check movie
	_, movieTypeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)
	gladiator, err := itemRepo.GetByTitle(ctx, "Gladiator", movieTypeID)
	require.NoError(t, err)
	require.NotNil(t, gladiator, "Gladiator should be detected as a movie")
	assert.Equal(t, movieTypeID, gladiator.MediaTypeID)

	// Check TV show
	_, tvShowTypeID, err := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	require.NoError(t, err)
	wire, err := itemRepo.GetByTitle(ctx, "The Wire", tvShowTypeID)
	require.NoError(t, err)
	require.NotNil(t, wire, "The Wire should be detected as a tv_show")
	assert.Equal(t, tvShowTypeID, wire.MediaTypeID)
}
