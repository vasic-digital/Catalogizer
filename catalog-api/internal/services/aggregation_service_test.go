package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newMockAggregationService(t *testing.T) (*AggregationService, *sql.DB, sqlmock.Sqlmock) {
	return newMockAggregationServiceWithMatcher(t, sqlmock.QueryMatcherEqual)
}

func newMockAggregationServiceWithRegex(t *testing.T) (*AggregationService, *sql.DB, sqlmock.Sqlmock) {
	return newMockAggregationServiceWithMatcher(t, sqlmock.QueryMatcherRegexp)
}

func newMockAggregationServiceWithMatcher(t *testing.T, matcher sqlmock.QueryMatcher) (*AggregationService, *sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	require.NoError(t, err)
	wrappedDB := database.WrapDB(db, database.DialectSQLite)

	logger := zap.NewNop()
	itemRepo := repository.NewMediaItemRepository(wrappedDB)
	fileRepo := repository.NewMediaFileRepository(wrappedDB)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(wrappedDB)
	extMetaRepo := repository.NewExternalMetadataRepository(wrappedDB)

	service := NewAggregationService(
		wrappedDB,
		logger,
		itemRepo,
		fileRepo,
		dirAnalysisRepo,
		extMetaRepo,
	)
	return service, db, mock
}

func TestNewAggregationService(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()
	wrappedDB := database.WrapDB(db, database.DialectSQLite)

	logger := zap.NewNop()
	itemRepo := repository.NewMediaItemRepository(wrappedDB)
	fileRepo := repository.NewMediaFileRepository(wrappedDB)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(wrappedDB)
	extMetaRepo := repository.NewExternalMetadataRepository(wrappedDB)

	service := NewAggregationService(
		wrappedDB,
		logger,
		itemRepo,
		fileRepo,
		dirAnalysisRepo,
		extMetaRepo,
	)

	assert.NotNil(t, service)
	assert.Equal(t, wrappedDB, service.db)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, itemRepo, service.itemRepo)
	assert.Equal(t, fileRepo, service.fileRepo)
	assert.Equal(t, dirAnalysisRepo, service.dirAnalysisRepo)
	assert.Equal(t, extMetaRepo, service.extMetaRepo)

	// Ensure no unexpected database interactions
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDetectMediaType_VideoDirectory(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "The Matrix (1999)",
		fileCount: 3,
		fileTypes: map[string]int{".mkv": 1, ".srt": 2},
	}

	typeName, parsed := s.detectMediaType(dir)
	assert.Equal(t, "movie", typeName)
	assert.Equal(t, "The Matrix", parsed.Title)
	assert.NotNil(t, parsed.Year)
	assert.Equal(t, 1999, *parsed.Year)
}

func TestDetectMediaType_TVShow(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Breaking Bad S01E01 720p",
		fileCount: 1,
		fileTypes: map[string]int{".mkv": 1},
	}

	typeName, parsed := s.detectMediaType(dir)
	assert.Equal(t, "tv_show", typeName)
	assert.Equal(t, "Breaking Bad", parsed.Title)
	assert.NotNil(t, parsed.Season)
	assert.Equal(t, 1, *parsed.Season)
}

func TestDetectMediaType_TVShowSeasonFolder(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Game of Thrones Season 3",
		fileCount: 10,
		fileTypes: map[string]int{".mkv": 10},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "tv_show", typeName)
}

func TestDetectMediaType_MusicAlbum(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Pink Floyd - The Wall",
		fileCount: 13,
		fileTypes: map[string]int{".flac": 13},
	}

	typeName, parsed := s.detectMediaType(dir)
	assert.Equal(t, "music_album", typeName)
	assert.Equal(t, "Pink Floyd", parsed.Artist)
	assert.Equal(t, "The Wall", parsed.Album)
}

func TestDetectMediaType_Software(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Ubuntu 24.04",
		fileCount: 1,
		fileTypes: map[string]int{".iso": 1},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "software", typeName)
}

func TestDetectMediaType_Comic(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Spider-Man #300",
		fileCount: 1,
		fileTypes: map[string]int{".cbr": 1},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "comic", typeName)
}

func TestDetectMediaType_Book(t *testing.T) {
	s := &AggregationService{}

	dir := directoryInfo{
		name:      "Dune - Frank Herbert",
		fileCount: 1,
		fileTypes: map[string]int{".epub": 1},
	}

	typeName, _ := s.detectMediaType(dir)
	assert.Equal(t, "book", typeName)
}

func TestDetectMediaTypeFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/movies/test.mkv", "video"},
		{"/music/song.mp3", "audio"},
		{"/software/ubuntu.iso", "disc_image"},
		{"/books/novel.epub", "ebook"},
		{"/comics/issue.cbr", "comic"},
		{"/apps/setup.exe", "software"},
		{"/docs/readme.txt", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, DetectMediaTypeFromPath(tt.path))
		})
	}
}

func TestAggregationService_GetTopLevelDirectories(t *testing.T) {
	ctx := context.Background()
	storageRootID := int64(1)

	t.Run("success with directories", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t)
		defer sqlDB.Close()

		// Mock parent directories query
		rows := sqlmock.NewRows([]string{"id", "path", "name"}).
			AddRow(1, "/movies", "movies").
			AddRow(2, "/tv", "tv")
		mock.ExpectQuery(`SELECT id, path, name FROM files
		WHERE storage_root_id = ? AND is_directory = 1 AND deleted = 0 AND parent_id IS NULL
		ORDER BY name`).
			WithArgs(storageRootID).
			WillReturnRows(rows)

		// Mock child files for first directory
		childRows1 := sqlmock.NewRows([]string{"id", "extension", "size"}).
			AddRow(101, ".mp4", 1024*1024).
			AddRow(102, ".srt", 5000)
		mock.ExpectQuery(`SELECT id, extension, size FROM files
			WHERE storage_root_id = ? AND parent_id = ? AND is_directory = 0 AND deleted = 0`).
			WithArgs(storageRootID, 1).
			WillReturnRows(childRows1)

		// Mock child files for second directory (empty)
		childRows2 := sqlmock.NewRows([]string{"id", "extension", "size"})
		mock.ExpectQuery(`SELECT id, extension, size FROM files
			WHERE storage_root_id = ? AND parent_id = ? AND is_directory = 0 AND deleted = 0`).
			WithArgs(storageRootID, 2).
			WillReturnRows(childRows2)

		dirs, err := service.getTopLevelDirectories(ctx, storageRootID)
		require.NoError(t, err)
		// Only first directory has files, second should be excluded
		require.Len(t, dirs, 1)

		// Check first directory
		assert.Equal(t, "/movies", dirs[0].path)
		assert.Equal(t, "movies", dirs[0].name)
		assert.Equal(t, 2, dirs[0].fileCount)
		assert.Equal(t, int64(1024*1024+5000), dirs[0].totalSize)
		assert.Len(t, dirs[0].fileIDs, 2)
		assert.Equal(t, map[string]int{".mp4": 1, ".srt": 1}, dirs[0].fileTypes)
		assert.ElementsMatch(t, []string{".mp4", ".srt"}, dirs[0].extensions)
	})

	t.Run("no directories found", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t)
		defer sqlDB.Close()

		rows := sqlmock.NewRows([]string{"id", "path", "name"})
		mock.ExpectQuery(`SELECT id, path, name FROM files
		WHERE storage_root_id = ? AND is_directory = 1 AND deleted = 0 AND parent_id IS NULL
		ORDER BY name`).
			WithArgs(storageRootID).
			WillReturnRows(rows)

		dirs, err := service.getTopLevelDirectories(ctx, storageRootID)
		require.NoError(t, err)
		assert.Empty(t, dirs)
	})

	t.Run("database error", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t)
		defer sqlDB.Close()

		mock.ExpectQuery(`SELECT id, path, name FROM files
		WHERE storage_root_id = ? AND is_directory = 1 AND deleted = 0 AND parent_id IS NULL
		ORDER BY name`).
			WithArgs(storageRootID).
			WillReturnError(sql.ErrConnDone)

		dirs, err := service.getTopLevelDirectories(ctx, storageRootID)
		assert.Error(t, err)
		assert.Nil(t, dirs)
	})
}

func TestAggregationService_ProcessDirectory(t *testing.T) {
	ctx := context.Background()
	storageRootID := int64(1)

	t.Run("new movie entity", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t) // Use exact matching
		defer sqlDB.Close()

		dir := directoryInfo{
			path:       "/movies/The Matrix (1999)",
			name:       "The Matrix (1999)",
			fileCount:  2,
			totalSize:  1500000000,
			fileIDs:    []int64{101, 102},
			fileTypes:  map[string]int{".mkv": 1, ".srt": 1},
			extensions: []string{".mkv", ".srt"},
		}

		now := time.Now()

		// Mock GetMediaTypeByName for "movie"
		mock.ExpectQuery(`SELECT id, name, description, detection_patterns, metadata_providers,
		created_at, updated_at
	FROM media_types WHERE name = ?`).
			WithArgs("movie").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "detection_patterns", "metadata_providers", "created_at", "updated_at"}).AddRow(1, "movie", "", nil, nil, now, now))

		// Mock GetByTitle - no existing entity
		mock.ExpectQuery(`SELECT id, media_type_id, title, original_title, year, description,
		genre, director, cast_crew, rating, runtime, language, country,
		status, parent_id, season_number, episode_number, track_number,
		first_detected, last_updated
	FROM media_items WHERE title = ? AND media_type_id = ? LIMIT 1`).
			WithArgs("The Matrix", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "media_type_id", "title", "original_title", "year", "description", "genre", "director", "cast_crew", "rating", "runtime", "language", "country", "status", "parent_id", "season_number", "episode_number", "track_number", "first_detected", "last_updated"}))

		// Mock Create media item - exact SQL from repository
		mock.ExpectExec(`INSERT INTO media_items (
		media_type_id, title, original_title, year, description,
		genre, director, cast_crew, rating, runtime, language, country,
		status, parent_id, season_number, episode_number, track_number,
		first_detected, last_updated
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
			WillReturnResult(sqlmock.NewResult(100, 1))

		// Mock LinkFileToItem for two files - exact SQL from repository
		mock.ExpectExec(`
		INSERT INTO media_files (media_item_id, file_id, quality_info, language, is_primary, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`).
			WithArgs(100, 101, nil, nil, true, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`
		INSERT INTO media_files (media_item_id, file_id, quality_info, language, is_primary, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`).
			WithArgs(100, 102, nil, nil, false, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(2, 1))

		// Mock GetByPath for directory analysis (not found) - exact SQL from repository
		mock.ExpectQuery(`SELECT id, directory_path, smb_root, media_item_id, confidence_score,
		detection_method, analysis_data, last_analyzed, files_count, total_size
	FROM directory_analyses WHERE directory_path = ? LIMIT 1`).
			WithArgs("/movies/The Matrix (1999)").
			WillReturnRows(sqlmock.NewRows([]string{"id", "directory_path", "smb_root", "media_item_id", "confidence_score", "detection_method", "analysis_data", "last_analyzed", "files_count", "total_size"}))

		// Mock Create directory analysis - exact SQL from repository
		mock.ExpectExec(`INSERT INTO directory_analyses (
		directory_path, smb_root, media_item_id, confidence_score,
		detection_method, analysis_data, last_analyzed, files_count, total_size
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`).
			WithArgs("/movies/The Matrix (1999)", "", 100, 0.8, "title_parser", "null", sqlmock.AnyArg(), 2, int64(1500000000)).
			WillReturnResult(sqlmock.NewResult(200, 1))

		isNew, err := service.processDirectory(ctx, dir, storageRootID)
		require.NoError(t, err)
		assert.True(t, isNew)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAggregationService_GetStorageRootName(t *testing.T) {
	ctx := context.Background()
	storageRootID := int64(1)

	t.Run("found", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t)
		defer sqlDB.Close()

		mock.ExpectQuery(`SELECT name FROM storage_roots WHERE id = ?`).
			WithArgs(storageRootID).
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("My Movies"))

		name := service.getStorageRootName(ctx, storageRootID)
		assert.Equal(t, "My Movies", name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t)
		defer sqlDB.Close()

		mock.ExpectQuery(`SELECT name FROM storage_roots WHERE id = ?`).
			WithArgs(storageRootID).
			WillReturnRows(sqlmock.NewRows([]string{"name"}))

		name := service.getStorageRootName(ctx, storageRootID)
		assert.Equal(t, "", name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		service, sqlDB, mock := newMockAggregationService(t)
		defer sqlDB.Close()

		mock.ExpectQuery(`SELECT name FROM storage_roots WHERE id = ?`).
			WithArgs(storageRootID).
			WillReturnError(sql.ErrConnDone)

		name := service.getStorageRootName(ctx, storageRootID)
		assert.Equal(t, "", name)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
