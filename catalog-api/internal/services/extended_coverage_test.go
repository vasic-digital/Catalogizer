package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"
	"catalogizer/repository"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"digital.vasic.filesystem/pkg/client"
)

// stubFSClient is a minimal filesystem.FileSystemClient for testing.
// FileExists always returns false (destination does not exist).
type stubFSClient struct{}

func (s *stubFSClient) Connect(ctx context.Context) error                              { return nil }
func (s *stubFSClient) Disconnect(ctx context.Context) error                           { return nil }
func (s *stubFSClient) IsConnected() bool                                              { return true }
func (s *stubFSClient) TestConnection(ctx context.Context) error                       { return nil }
func (s *stubFSClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}
func (s *stubFSClient) WriteFile(ctx context.Context, path string, data io.Reader) error {
	return fmt.Errorf("not implemented")
}
func (s *stubFSClient) GetFileInfo(ctx context.Context, path string) (*client.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}
func (s *stubFSClient) FileExists(ctx context.Context, path string) (bool, error) {
	return false, nil // destination does not exist, so validation passes
}
func (s *stubFSClient) DeleteFile(ctx context.Context, path string) error {
	return fmt.Errorf("not implemented")
}
func (s *stubFSClient) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	return fmt.Errorf("not implemented")
}
func (s *stubFSClient) ListDirectory(ctx context.Context, path string) ([]*client.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}
func (s *stubFSClient) CreateDirectory(ctx context.Context, path string) error {
	return fmt.Errorf("not implemented")
}
func (s *stubFSClient) DeleteDirectory(ctx context.Context, path string) error {
	return fmt.Errorf("not implemented")
}
func (s *stubFSClient) GetProtocol() string    { return "local" }
func (s *stubFSClient) GetConfig() interface{} { return nil }

// ---------------------------------------------------------------------------
// Shared test helpers
// ---------------------------------------------------------------------------

// setupExtendedTestDB creates a per-test in-memory SQLite DB with all tables
// needed by the services under test. Each call gets a unique file-URI so
// concurrent subtests never share state.
func setupExtendedTestDB(t *testing.T) *database.DB {
	id := testDBCounter.Add(1)
	dsn := fmt.Sprintf("file:extdb%d?mode=memory&cache=shared&_busy_timeout=5000", id)
	sqlDB, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(10)

	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	migrations := []string{
		// ------------- storage_roots (needed by files + aggregation) ---------
		`CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			protocol TEXT NOT NULL DEFAULT 'local',
			path TEXT,
			enabled BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// ------------- files (needed by aggregation) -----------------------
		`CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			path TEXT NOT NULL,
			name TEXT NOT NULL,
			extension TEXT,
			size INTEGER NOT NULL DEFAULT 0,
			is_directory BOOLEAN DEFAULT 0,
			deleted BOOLEAN DEFAULT 0,
			parent_id INTEGER,
			modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
			FOREIGN KEY (parent_id) REFERENCES files(id)
		)`,

		// ------------- media types (needed by aggregation) ------------------
		`CREATE TABLE IF NOT EXISTS media_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			detection_patterns TEXT,
			metadata_providers TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// ------------- media items (aggregation entity table) ---------------
		`CREATE TABLE IF NOT EXISTS media_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_type_id INTEGER,
			title TEXT NOT NULL,
			original_title TEXT DEFAULT '',
			year INTEGER,
			description TEXT DEFAULT '',
			genre TEXT,
			director TEXT,
			cast_crew TEXT,
			rating REAL,
			runtime INTEGER,
			language TEXT,
			country TEXT,
			status TEXT DEFAULT 'detected',
			parent_id INTEGER,
			season_number INTEGER,
			episode_number INTEGER,
			track_number INTEGER,
			disc_number INTEGER DEFAULT 0,
			first_detected DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			-- columns used by music/video player services
			path TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL DEFAULT '',
			file_path TEXT DEFAULT '',
			file_size INTEGER DEFAULT 0,
			duration INTEGER DEFAULT 0,
			artist TEXT DEFAULT '',
			album TEXT DEFAULT '',
			album_id INTEGER,
			album_artist TEXT DEFAULT '',
			format TEXT DEFAULT '',
			bitrate INTEGER DEFAULT 0,
			sample_rate INTEGER DEFAULT 0,
			channels INTEGER DEFAULT 0,
			bpm INTEGER,
			key TEXT,
			play_count INTEGER DEFAULT 0,
			last_played DATETIME,
			date_added DATETIME DEFAULT CURRENT_TIMESTAMP,
			user_id INTEGER,
			-- video columns
			resolution TEXT DEFAULT '',
			aspect_ratio TEXT DEFAULT '',
			frame_rate REAL DEFAULT 0.0,
			codec TEXT DEFAULT '',
			hdr BOOLEAN DEFAULT FALSE,
			dolby_vision BOOLEAN DEFAULT FALSE,
			dolby_atmos BOOLEAN DEFAULT FALSE,
			genres TEXT DEFAULT '[]',
			directors TEXT DEFAULT '[]',
			actors TEXT DEFAULT '[]',
			writers TEXT DEFAULT '[]',
			imdb_id TEXT DEFAULT '',
			tmdb_id TEXT DEFAULT '',
			release_date DATETIME,
			user_rating INTEGER DEFAULT 0,
			is_favorite BOOLEAN DEFAULT FALSE,
			watched_percentage REAL DEFAULT 0.0,
			artist_id INTEGER,
			FOREIGN KEY (media_type_id) REFERENCES media_types(id),
			FOREIGN KEY (parent_id) REFERENCES media_items(id)
		)`,

		// ------------- media_files (junction; needed by aggregation) --------
		`CREATE TABLE IF NOT EXISTS media_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER NOT NULL,
			file_id INTEGER,
			quality_info TEXT,
			language TEXT,
			is_primary BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (media_item_id) REFERENCES media_items(id)
		)`,

		// ------------- directory_analyses (aggregation) ---------------------
		`CREATE TABLE IF NOT EXISTS directory_analyses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			directory_path TEXT NOT NULL,
			smb_root TEXT DEFAULT '',
			media_item_id INTEGER,
			confidence_score REAL DEFAULT 0,
			detection_method TEXT DEFAULT '',
			analysis_data TEXT,
			last_analyzed DATETIME DEFAULT CURRENT_TIMESTAMP,
			files_count INTEGER DEFAULT 0,
			total_size INTEGER DEFAULT 0
		)`,

		// ------------- external_metadata -----------------------------------
		`CREATE TABLE IF NOT EXISTS external_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_item_id INTEGER NOT NULL,
			provider TEXT NOT NULL,
			external_id TEXT DEFAULT '',
			data TEXT DEFAULT '',
			rating REAL,
			review_url TEXT,
			cover_url TEXT,
			trailer_url TEXT,
			last_fetched DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// ------------- users (FK target) -----------------------------------
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT OR IGNORE INTO users (id, username, email) VALUES (1, 'testuser', 'test@example.com')`,

		// ------------- video/music session tables --------------------------
		`CREATE TABLE IF NOT EXISTS video_playback_sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			session_data TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS music_playback_sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			session_data TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// ------------- video bookmarks ------------------------------------
		`CREATE TABLE IF NOT EXISTS video_bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			video_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			title TEXT,
			description TEXT,
			thumbnail_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// ------------- playback_positions (for position service) -----------
		`CREATE TABLE IF NOT EXISTS playback_positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			media_item_id INTEGER NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			duration INTEGER NOT NULL DEFAULT 0,
			percent_complete REAL DEFAULT 0.0,
			last_played DATETIME,
			is_completed BOOLEAN DEFAULT 0,
			device_info TEXT,
			playback_quality TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, media_item_id)
		)`,

		// ------------- universal_rename_events (rename tracker) -----------
		// (InitializeTables creates this, but we pre-create for GetStatistics)

		// ------------- seed media types -----------------------------------
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (1, 'movie')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (2, 'tv_show')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (3, 'tv_season')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (4, 'tv_episode')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (5, 'music_album')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (6, 'music_artist')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (7, 'song')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (8, 'game')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (9, 'software')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (10, 'book')`,
		`INSERT OR IGNORE INTO media_types (id, name) VALUES (11, 'comic')`,
	}

	for _, m := range migrations {
		_, err := sqlDB.Exec(m)
		require.NoError(t, err, "migration failed: %s", m)
	}

	return database.WrapDB(sqlDB, database.DialectSQLite)
}

// ==========================================================================
// 1. AggregationService — real-SQLite tests
// ==========================================================================

func TestAggregateAfterScan_CreatesMediaItems(t *testing.T) {
	db := setupExtendedTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	// Seed a storage root
	_, err := db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'test-root', 'local')")
	require.NoError(t, err)

	// Create a top-level directory (parent_id IS NULL)
	_, err = db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, deleted, parent_id, size, modified_at)
		VALUES (10, 1, '/movies/Inception (2010)', 'Inception (2010)', 1, 0, NULL, 0, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	// Create child files inside the directory
	_, err = db.Exec(`INSERT INTO files (id, storage_root_id, path, name, extension, is_directory, deleted, parent_id, size, modified_at)
		VALUES (11, 1, '/movies/Inception (2010)/Inception.mkv', 'Inception.mkv', '.mkv', 0, 0, 10, 5000000000, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO files (id, storage_root_id, path, name, extension, is_directory, deleted, parent_id, size, modified_at)
		VALUES (12, 1, '/movies/Inception (2010)/Inception.srt', 'Inception.srt', '.srt', 0, 0, 10, 50000, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	// Run aggregation
	err = svc.AggregateAfterScan(ctx, 1)
	require.NoError(t, err)

	// Verify a media_item was created for "Inception"
	item, err := itemRepo.GetByTitle(ctx, "Inception", 1) // movie type id = 1
	require.NoError(t, err)
	require.NotNil(t, item, "expected media item 'Inception' to be created")
	assert.Equal(t, "Inception", item.Title)
	require.NotNil(t, item.Year)
	assert.Equal(t, 2010, *item.Year)
	assert.Equal(t, "detected", item.Status)

	// Verify files are linked
	linkedFiles, err := fileRepo.GetFilesByItem(ctx, item.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(linkedFiles), 2, "expected at least 2 linked files")

	// Verify directory analysis was created
	da, err := dirAnalysisRepo.GetByPath(ctx, "/movies/Inception (2010)")
	require.NoError(t, err)
	require.NotNil(t, da)
	assert.Equal(t, 0.8, da.ConfidenceScore, "movie with year should get 0.8 confidence")
	assert.Equal(t, "title_parser", da.DetectionMethod)
}

func TestAggregateAfterScan_NoDirectories(t *testing.T) {
	db := setupExtendedTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	_, err := db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'empty-root', 'local')")
	require.NoError(t, err)

	err = svc.AggregateAfterScan(ctx, 1)
	require.NoError(t, err) // should succeed with 0 entities
}

func TestBuildTVHierarchy(t *testing.T) {
	db := setupExtendedTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	// Seed a storage root
	_, err := db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'tv-root', 'local')")
	require.NoError(t, err)

	// Create a top-level directory for a TV show with season+episode
	_, err = db.Exec(`INSERT INTO files (id, storage_root_id, path, name, is_directory, deleted, parent_id, size, modified_at)
		VALUES (20, 1, '/tv/Breaking Bad S01E01 720p', 'Breaking Bad S01E01 720p', 1, 0, NULL, 0, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO files (id, storage_root_id, path, name, extension, is_directory, deleted, parent_id, size, modified_at)
		VALUES (21, 1, '/tv/Breaking Bad S01E01 720p/episode.mkv', 'episode.mkv', '.mkv', 0, 0, 20, 700000000, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	err = svc.AggregateAfterScan(ctx, 1)
	require.NoError(t, err)

	// Verify the TV show entity was created
	show, err := itemRepo.GetByTitle(ctx, "Breaking Bad", 2) // tv_show type = 2
	require.NoError(t, err)
	require.NotNil(t, show, "expected TV show entity")

	// Verify season was created
	season, err := itemRepo.GetByTitle(ctx, "Season 1", 3) // tv_season type = 3
	require.NoError(t, err)
	require.NotNil(t, season, "expected Season 1 entity")
	require.NotNil(t, season.ParentID)
	assert.Equal(t, show.ID, *season.ParentID)
	require.NotNil(t, season.SeasonNumber)
	assert.Equal(t, 1, *season.SeasonNumber)

	// Verify episode was created
	var epCount int64
	err = db.QueryRow("SELECT COUNT(*) FROM media_items WHERE media_type_id = 4 AND parent_id = ?", season.ID).Scan(&epCount)
	require.NoError(t, err)
	assert.Equal(t, int64(1), epCount, "expected 1 episode entity")
}

func TestBuildTVHierarchy_Direct(t *testing.T) {
	db := setupExtendedTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	dirAnalysisRepo := repository.NewDirectoryAnalysisRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)

	svc := NewAggregationService(db, logger, itemRepo, fileRepo, dirAnalysisRepo, extMetaRepo)

	// Create a TV show parent item directly
	showItem := &models.MediaItem{
		Title:       "Test Show",
		MediaTypeID: 2,
		Status:      "detected",
	}
	showID, err := itemRepo.Create(ctx, showItem)
	require.NoError(t, err)

	season := 2
	episode := 5
	parsed := ParsedTitle{
		Title:   "Test Show",
		Season:  &season,
		Episode: &episode,
	}

	svc.buildTVHierarchy(ctx, showID, 2, parsed)

	// Verify season was created
	seasonItem, err := itemRepo.GetByTitle(ctx, "Season 2", 3)
	require.NoError(t, err)
	require.NotNil(t, seasonItem)
	require.NotNil(t, seasonItem.ParentID)
	assert.Equal(t, showID, *seasonItem.ParentID)

	// Verify episode was created under the season
	var epTitle string
	err = db.QueryRow("SELECT title FROM media_items WHERE media_type_id = 4 AND parent_id = ?", seasonItem.ID).Scan(&epTitle)
	require.NoError(t, err)
	assert.Equal(t, "Episode 5", epTitle)
}

// ==========================================================================
// 2. DuplicateDetectionService — extended tests
// ==========================================================================

func TestDuplicateDetectionService_DetectDuplicates_EmptyDB(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	ctx := context.Background()
	req := &DuplicateDetectionRequest{
		MinSimilarity: 0.8,
		UserID:        1,
	}

	groups, err := svc.DetectDuplicates(ctx, req)
	require.NoError(t, err)
	assert.Empty(t, groups, "no duplicates expected in empty state")
}

func TestDuplicateDetectionService_DetectDuplicates_WithMediaTypes(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	ctx := context.Background()
	req := &DuplicateDetectionRequest{
		MediaTypes:    []MediaType{MediaTypeMovie, MediaTypeMusic},
		MinSimilarity: 0.9,
		UserID:        1,
	}

	groups, err := svc.DetectDuplicates(ctx, req)
	require.NoError(t, err)
	// getMediaItems returns empty by default → no duplicates
	assert.Empty(t, groups)
}

func TestDuplicateDetectionService_DeterminePrimaryItem(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	// determinePrimaryItem is a no-op placeholder, but it must not panic
	group := &DuplicateGroup{
		PrimaryItem: DuplicateItem{
			MediaID:  "1",
			Title:    "Movie A",
			FileSize: 1000,
			Quality:  "1080p",
		},
		DuplicateItems: []DuplicateItem{
			{MediaID: "2", Title: "Movie A", FileSize: 2000, Quality: "4K"},
			{MediaID: "3", Title: "Movie A", FileSize: 500, Quality: "720p"},
		},
	}

	// Should not panic
	svc.determinePrimaryItem(group)
}

func TestDuplicateDetectionService_AnalyzeFieldMatches(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	item1 := &DuplicateItem{
		Title:    "The Matrix",
		Year:     1999,
		Director: "Wachowski",
		FileSize: 5000000000,
		Format:   "mkv",
	}
	item2 := &DuplicateItem{
		Title:    "The Matrix",
		Year:     1999,
		Director: "Wachowski",
		FileSize: 4500000000,
		Format:   "avi",
	}

	analysis := &SimilarityAnalysis{
		MatchingFields:   []string{},
		DifferencesFound: []string{},
	}

	// Should not panic — analyzeFieldMatches is a placeholder
	svc.analyzeFieldMatches(item1, item2, analysis)
}

func TestDuplicateDetectionService_CalculateSimilarity_ComprehensiveScoring(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	// Items with similar titles, no hash, no external IDs → uses weighted scoring
	item1 := &DuplicateItem{
		Title:        "Inception",
		FileName:     "inception_2010.mkv",
		FileSize:     5000000000,
		Year:         2010,
		Director:     "Christopher Nolan",
		Format:       "mkv",
		ExternalIDs:  map[string]string{},
		Fingerprints: map[string]string{},
	}
	item2 := &DuplicateItem{
		Title:        "Inception",
		FileName:     "inception.mkv",
		FileSize:     4800000000,
		Year:         2010,
		Director:     "Christopher Nolan",
		Format:       "mkv",
		ExternalIDs:  map[string]string{},
		Fingerprints: map[string]string{},
	}

	analysis := svc.calculateSimilarity(nil, item1, item2, MediaTypeMovie)
	assert.NotNil(t, analysis)
	assert.Greater(t, analysis.OverallScore, 0.0, "similar items should have positive score")
	assert.Greater(t, analysis.TitleSimilarity, 0.0, "identical titles should score high")
	assert.Greater(t, analysis.MetadataSimilarity, 0.0, "same director/year should score")
}

func TestDuplicateDetectionService_CalculateBookMetadataSimilarity(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	item1 := &DuplicateItem{
		Author:      "Frank Herbert",
		Year:        1965,
		ExternalIDs: map[string]string{},
	}
	item2 := &DuplicateItem{
		Author:      "Frank Herbert",
		Year:        1965,
		ExternalIDs: map[string]string{},
	}
	score := svc.calculateBookMetadataSimilarity(item1, item2)
	assert.Greater(t, score, 0.8, "same author and year should score high")

	// Empty metadata
	empty1 := &DuplicateItem{ExternalIDs: map[string]string{}}
	empty2 := &DuplicateItem{ExternalIDs: map[string]string{}}
	score2 := svc.calculateBookMetadataSimilarity(empty1, empty2)
	assert.Equal(t, 0.0, score2)
}

func TestDuplicateDetectionService_CalculateSoftwareMetadataSimilarity(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	item1 := &DuplicateItem{FileSize: 1000000, Metadata: map[string]interface{}{}}
	item2 := &DuplicateItem{FileSize: 1000000, Metadata: map[string]interface{}{}}
	score := svc.calculateSoftwareMetadataSimilarity(item1, item2)
	assert.GreaterOrEqual(t, score, 0.0)

	// Empty metadata
	empty1 := &DuplicateItem{Metadata: map[string]interface{}{}}
	empty2 := &DuplicateItem{Metadata: map[string]interface{}{}}
	score2 := svc.calculateSoftwareMetadataSimilarity(empty1, empty2)
	assert.Equal(t, 0.0, score2)
}

func TestDuplicateDetectionService_CalculateGenericMetadataSimilarity(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	item1 := &DuplicateItem{FileName: "file.mkv", FileSize: 1000, Format: "mkv"}
	item2 := &DuplicateItem{FileName: "file.mkv", FileSize: 1000, Format: "mkv"}
	score := svc.calculateGenericMetadataSimilarity(item1, item2)
	assert.Greater(t, score, 0.5, "identical files should score high")
}

func TestDuplicateDetectionService_StoreDuplicateGroup(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	group := &DuplicateGroup{ID: "test-group"}
	err := svc.storeDuplicateGroup(context.Background(), group)
	assert.NoError(t, err, "placeholder storeDuplicateGroup should not error")
}

func TestDuplicateDetectionService_DetectDuplicatesForMediaType_SingleItem(t *testing.T) {
	logger := zap.NewNop()
	svc := NewDuplicateDetectionService(nil, logger, nil)

	// getMediaItems returns empty → less than 2 items → empty result
	groups, err := svc.detectDuplicatesForMediaType(context.Background(), MediaTypeMovie, &DuplicateDetectionRequest{
		MinSimilarity: 0.8,
	})
	require.NoError(t, err)
	assert.Empty(t, groups)
}

// ==========================================================================
// 3. VideoPlayerService — real-SQLite session tests
// ==========================================================================

func setupVideoService(t *testing.T) (*VideoPlayerService, *database.DB) {
	db := setupExtendedTestDB(t)
	logger := zap.NewNop()
	positionService := NewPlaybackPositionService(db, logger)
	coverArtService := NewCoverArtService(db, logger)
	svc := NewVideoPlayerService(db, logger, nil, positionService, nil, coverArtService, nil)
	return svc, db
}

func insertVideoItem(t *testing.T, db *database.DB, id int64, title string, duration int64) {
	_, err := db.Exec(`INSERT INTO media_items
		(id, path, title, type, file_path, file_size, duration, resolution, aspect_ratio,
		 frame_rate, bitrate, codec, year, language, country, date_added)
		VALUES (?, ?, ?, 'video', '/test/video.mkv', 5000000000, ?, '1920x1080', '16:9',
		 23.976, 8000000, 'h264', 2020, 'en', 'US', CURRENT_TIMESTAMP)`,
		id, "/videos/"+title, title, duration)
	require.NoError(t, err)
}

func createAndSaveVideoSession(t *testing.T, db *database.DB, sessionID string, userID int64, video *VideoContent) {
	session := &VideoPlaybackSession{
		ID:            sessionID,
		UserID:        userID,
		CurrentVideo:  video,
		Playlist:      []VideoContent{*video},
		PlaylistIndex: 0,
		PlayMode:      VideoPlayModeSingle,
		Volume:        1.0,
		PlaybackSpeed: 1.0,
		PlaybackState: PlaybackStatePlaying,
		Position:      0,
		Duration:      video.Duration,
		VideoQuality:  Quality1080p,
		LastActivity:  time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	sessionData, err := json.Marshal(session)
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(`INSERT INTO video_playback_sessions (id, user_id, session_data, expires_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`, sessionID, userID, string(sessionData), expiresAt)
	require.NoError(t, err)
}

func TestVideoPlayerService_PlayVideo(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	insertVideoItem(t, db, 1, "Test Movie", 7200000)

	session, err := svc.PlayVideo(ctx, &PlayVideoRequest{
		UserID:   1,
		VideoID:  1,
		PlayMode: VideoPlayModeSingle,
		Quality:  Quality1080p,
		AutoPlay: true,
		DeviceInfo: DeviceInfo{
			DeviceID:   "test-device",
			DeviceName: "Test Browser",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, int64(1), session.UserID)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, PlaybackStatePlaying, session.PlaybackState)
	assert.Equal(t, int64(7200000), session.Duration)
	assert.Equal(t, Quality1080p, session.VideoQuality)
	assert.True(t, session.AutoPlay)
	require.NotNil(t, session.CurrentVideo)
	assert.Equal(t, "Test Movie", session.CurrentVideo.Title)
}

func TestVideoPlayerService_PlayVideo_NotFound(t *testing.T) {
	svc, _ := setupVideoService(t)
	ctx := context.Background()

	_, err := svc.PlayVideo(ctx, &PlayVideoRequest{
		UserID:   1,
		VideoID:  999,
		PlayMode: VideoPlayModeSingle,
	})
	assert.Error(t, err)
}

func TestVideoPlayerService_GetVideoSession(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	video := &VideoContent{
		ID:       1,
		Title:    "Session Test Movie",
		Duration: 5400000,
	}
	createAndSaveVideoSession(t, db, "test-session-get", 1, video)

	session, err := svc.GetVideoSession(ctx, "test-session-get")
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, "test-session-get", session.ID)
	assert.Equal(t, int64(1), session.UserID)
}

func TestVideoPlayerService_GetVideoSession_Expired(t *testing.T) {
	svc, _ := setupVideoService(t)
	ctx := context.Background()

	_, err := svc.GetVideoSession(ctx, "nonexistent-session-xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestVideoPlayerService_SeekVideo(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	video := &VideoContent{ID: 1, Title: "Seek Test", Duration: 3600000}
	createAndSaveVideoSession(t, db, "test-session-seek", 1, video)

	// Seek forward
	session, err := svc.SeekVideo(ctx, &VideoSeekRequest{
		SessionID: "test-session-seek",
		Position:  1800000, // 30 minutes
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1800000), session.Position)
	assert.Equal(t, 1, session.ViewingProgress.SeekCount)
	assert.Equal(t, 1, session.ViewingProgress.FastForwardCount)

	// Seek backward
	session, err = svc.SeekVideo(ctx, &VideoSeekRequest{
		SessionID: "test-session-seek",
		Position:  900000,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(900000), session.Position)
	assert.Equal(t, 2, session.ViewingProgress.SeekCount)
	assert.Equal(t, 1, session.ViewingProgress.RewindCount)
}

func TestVideoPlayerService_SeekVideo_ClampedPosition(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	video := &VideoContent{ID: 1, Title: "Clamp Test", Duration: 1000}
	createAndSaveVideoSession(t, db, "test-session-clamp", 1, video)

	// Seek past end → clamped to duration
	session, err := svc.SeekVideo(ctx, &VideoSeekRequest{
		SessionID: "test-session-clamp",
		Position:  99999,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1000), session.Position)

	// Seek negative → clamped to 0
	session, err = svc.SeekVideo(ctx, &VideoSeekRequest{
		SessionID: "test-session-clamp",
		Position:  -500,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), session.Position)
}

func TestVideoPlayerService_CreateVideoBookmark(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	insertVideoItem(t, db, 2, "Bookmark Movie", 7200000)

	// Play the video to create a session
	session, err := svc.PlayVideo(ctx, &PlayVideoRequest{
		UserID:   1,
		VideoID:  2,
		PlayMode: VideoPlayModeSingle,
		Quality:  Quality1080p,
	})
	require.NoError(t, err)

	// Seek to a position first
	session, err = svc.SeekVideo(ctx, &VideoSeekRequest{
		SessionID: session.ID,
		Position:  3600000,
	})
	require.NoError(t, err)

	// Create a bookmark
	bookmark, err := svc.CreateVideoBookmark(ctx, &CreateVideoBookmarkRequest{
		SessionID:   session.ID,
		Title:       "Cool Scene",
		Description: "An interesting part of the movie",
	})
	require.NoError(t, err)
	require.NotNil(t, bookmark)
	assert.Equal(t, "Cool Scene", bookmark.Title)
	assert.Equal(t, "An interesting part of the movie", bookmark.Description)
	assert.Equal(t, int64(3600000), bookmark.Position)
	assert.Equal(t, int64(2), bookmark.VideoID)
}

func TestVideoPlayerService_UpdateVideoPlayback(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	video := &VideoContent{ID: 1, Title: "Update Test", Duration: 5000000}
	createAndSaveVideoSession(t, db, "test-update-session", 1, video)

	// Pause
	pauseState := PlaybackStatePaused
	session, err := svc.UpdateVideoPlayback(ctx, &UpdateVideoPlaybackRequest{
		SessionID: "test-update-session",
		State:     &pauseState,
	})
	require.NoError(t, err)
	assert.Equal(t, PlaybackStatePaused, session.PlaybackState)
	assert.Equal(t, 1, session.ViewingProgress.PauseCount)

	// Update volume
	vol := 0.5
	session, err = svc.UpdateVideoPlayback(ctx, &UpdateVideoPlaybackRequest{
		SessionID: "test-update-session",
		Volume:    &vol,
	})
	require.NoError(t, err)
	assert.Equal(t, 0.5, session.Volume)

	// Mute
	muted := true
	session, err = svc.UpdateVideoPlayback(ctx, &UpdateVideoPlaybackRequest{
		SessionID: "test-update-session",
		IsMuted:   &muted,
	})
	require.NoError(t, err)
	assert.True(t, session.IsMuted)

	// Change playback speed
	speed := 2.0
	session, err = svc.UpdateVideoPlayback(ctx, &UpdateVideoPlaybackRequest{
		SessionID:     "test-update-session",
		PlaybackSpeed: &speed,
	})
	require.NoError(t, err)
	assert.Equal(t, 2.0, session.PlaybackSpeed)

	// Change quality
	q := Quality2160p
	session, err = svc.UpdateVideoPlayback(ctx, &UpdateVideoPlaybackRequest{
		SessionID: "test-update-session",
		Quality:   &q,
	})
	require.NoError(t, err)
	assert.Equal(t, Quality2160p, session.VideoQuality)
	assert.Equal(t, 1, session.ViewingProgress.QualityChanges)
}

func TestVideoPlayerService_NextVideo_EndOfPlaylist(t *testing.T) {
	svc, db := setupVideoService(t)
	ctx := context.Background()

	video := &VideoContent{ID: 1, Title: "Only Video", Duration: 1000}
	createAndSaveVideoSession(t, db, "test-next-end", 1, video)

	// Next on a single-item playlist should stop
	session, err := svc.NextVideo(ctx, "test-next-end")
	require.NoError(t, err)
	assert.Equal(t, PlaybackStateStopped, session.PlaybackState)
}

// ==========================================================================
// 4. MusicPlayerService — real-SQLite session tests
// ==========================================================================

func setupMusicService(t *testing.T) (*MusicPlayerService, *database.DB) {
	db := setupExtendedTestDB(t)
	logger := zap.NewNop()
	svc := NewMusicPlayerService(db, logger, nil, nil, nil, nil, nil, nil)
	return svc, db
}

func insertMusicTrack(t *testing.T, db *database.DB, id int64, title, artist, album string, duration int64) {
	_, err := db.Exec(`INSERT INTO media_items
		(id, path, title, type, artist, album, album_artist, genre, year,
		 track_number, disc_number, duration, file_path, file_size, format,
		 bitrate, sample_rate, channels, play_count, date_added)
		VALUES (?, ?, ?, 'audio', ?, ?, ?, 'Rock', 2020,
		 1, 1, ?, '/music/track.flac', 50000000, 'flac',
		 1411, 44100, 2, 0, CURRENT_TIMESTAMP)`,
		id, "/music/"+title, title, artist, album, artist, duration)
	require.NoError(t, err)
}

func createAndSaveMusicSession(t *testing.T, db *database.DB, sessionID string, userID int64, tracks []MusicTrack, queueIndex int) {
	currentTrack := &tracks[queueIndex]
	session := &MusicPlaybackSession{
		ID:                sessionID,
		UserID:            userID,
		CurrentTrack:      currentTrack,
		Queue:             tracks,
		QueueIndex:        queueIndex,
		PlayMode:          PlayModeQueue,
		RepeatMode:        RepeatModeOff,
		ShuffleEnabled:    false,
		Volume:            1.0,
		IsMuted:           false,
		Crossfade:         false,
		CrossfadeDuration: 3000,
		EqualizerPreset:   "flat",
		EqualizerBands:    make(map[string]float64),
		PlaybackState:     PlaybackStatePlaying,
		Position:          0,
		Duration:          currentTrack.Duration,
		LastActivity:      time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	sessionData, err := json.Marshal(session)
	require.NoError(t, err)

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(`INSERT INTO music_playback_sessions (id, user_id, session_data, expires_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`, sessionID, userID, string(sessionData), expiresAt)
	require.NoError(t, err)
}

func TestMusicPlayerService_PlayTrack(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	insertMusicTrack(t, db, 1, "Bohemian Rhapsody", "Queen", "Night at the Opera", 354000)

	session, err := svc.PlayTrack(ctx, &PlayTrackRequest{
		UserID:   1,
		TrackID:  1,
		PlayMode: PlayModeTrack,
		Quality:  QualityLossless,
		DeviceInfo: DeviceInfo{
			DeviceID:   "test-dev",
			DeviceName: "Test Player",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, int64(1), session.UserID)
	assert.Equal(t, PlaybackStatePlaying, session.PlaybackState)
	assert.Equal(t, int64(354000), session.Duration)
	require.NotNil(t, session.CurrentTrack)
	assert.Equal(t, "Bohemian Rhapsody", session.CurrentTrack.Title)
	assert.Equal(t, "Queen", session.CurrentTrack.Artist)
	assert.Equal(t, 1, len(session.Queue))
}

func TestMusicPlayerService_PlayTrack_NotFound(t *testing.T) {
	svc, _ := setupMusicService(t)
	ctx := context.Background()

	_, err := svc.PlayTrack(ctx, &PlayTrackRequest{
		UserID:   1,
		TrackID:  999,
		PlayMode: PlayModeTrack,
	})
	assert.Error(t, err)
}

func TestMusicPlayerService_GetSession(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Track 1", Duration: 200000},
	}
	createAndSaveMusicSession(t, db, "music-session-1", 1, tracks, 0)

	session, err := svc.GetSession(ctx, "music-session-1")
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, "music-session-1", session.ID)
	assert.Equal(t, int64(1), session.UserID)
}

func TestMusicPlayerService_GetSession_NotFound(t *testing.T) {
	svc, _ := setupMusicService(t)
	ctx := context.Background()

	_, err := svc.GetSession(ctx, "nonexistent-session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestMusicPlayerService_NextTrack(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Track 1", Duration: 200000},
		{ID: 2, Title: "Track 2", Duration: 300000},
		{ID: 3, Title: "Track 3", Duration: 250000},
	}
	createAndSaveMusicSession(t, db, "music-next-1", 1, tracks, 0)

	// Advance to track 2
	session, err := svc.NextTrack(ctx, "music-next-1")
	require.NoError(t, err)
	assert.Equal(t, 1, session.QueueIndex)
	assert.Equal(t, "Track 2", session.CurrentTrack.Title)
	assert.Equal(t, int64(0), session.Position, "position should reset on next track")
}

func TestMusicPlayerService_NextTrack_EndOfQueue(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Only Track", Duration: 200000},
	}
	createAndSaveMusicSession(t, db, "music-next-end", 1, tracks, 0)

	session, err := svc.NextTrack(ctx, "music-next-end")
	require.NoError(t, err)
	assert.Equal(t, PlaybackStateStopped, session.PlaybackState)
}

func TestMusicPlayerService_PreviousTrack(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Track 1", Duration: 200000},
		{ID: 2, Title: "Track 2", Duration: 300000},
	}
	createAndSaveMusicSession(t, db, "music-prev-1", 1, tracks, 1)

	// Previous should go back to track 1 (position is 0, which is <= 3000ms)
	session, err := svc.PreviousTrack(ctx, "music-prev-1")
	require.NoError(t, err)
	assert.Equal(t, 0, session.QueueIndex)
	assert.Equal(t, "Track 1", session.CurrentTrack.Title)
}

func TestMusicPlayerService_PreviousTrack_RestartsIfPastThreshold(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Track 1", Duration: 200000},
		{ID: 2, Title: "Track 2", Duration: 300000},
	}

	// Create session at track index 1 with position > 3000
	session := &MusicPlaybackSession{
		ID:                "music-prev-restart",
		UserID:            1,
		CurrentTrack:      &tracks[1],
		Queue:             tracks,
		QueueIndex:        1,
		PlayMode:          PlayModeQueue,
		RepeatMode:        RepeatModeOff,
		Volume:            1.0,
		EqualizerPreset:   "flat",
		EqualizerBands:    make(map[string]float64),
		PlaybackState:     PlaybackStatePlaying,
		Position:          5000, // > 3000ms
		Duration:          tracks[1].Duration,
		LastActivity:      time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	sessionData, err := json.Marshal(session)
	require.NoError(t, err)
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = db.Exec(`INSERT INTO music_playback_sessions (id, user_id, session_data, expires_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`, "music-prev-restart", 1, string(sessionData), expiresAt)
	require.NoError(t, err)

	// Should restart current track (position > 3000)
	result, err := svc.PreviousTrack(ctx, "music-prev-restart")
	require.NoError(t, err)
	assert.Equal(t, 1, result.QueueIndex, "should stay on same track")
	assert.Equal(t, int64(0), result.Position, "position should reset to 0")
}

func TestMusicPlayerService_Seek(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Seek Track", Duration: 300000},
	}
	createAndSaveMusicSession(t, db, "music-seek-1", 1, tracks, 0)

	session, err := svc.Seek(ctx, &SeekRequest{
		SessionID: "music-seek-1",
		Position:  150000,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(150000), session.Position)

	// Clamp negative
	session, err = svc.Seek(ctx, &SeekRequest{
		SessionID: "music-seek-1",
		Position:  -100,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), session.Position)

	// Clamp past duration
	session, err = svc.Seek(ctx, &SeekRequest{
		SessionID: "music-seek-1",
		Position:  999999,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(300000), session.Position)
}

func TestMusicPlayerService_AddToQueue(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	// Insert tracks in DB
	insertMusicTrack(t, db, 10, "Queue Track 1", "Artist A", "Album X", 200000)
	insertMusicTrack(t, db, 11, "Queue Track 2", "Artist B", "Album Y", 250000)

	// Create session with one track
	initialTracks := []MusicTrack{
		{ID: 1, Title: "Initial Track", Duration: 180000},
	}
	createAndSaveMusicSession(t, db, "music-queue-1", 1, initialTracks, 0)

	// Add tracks to queue
	session, err := svc.AddToQueue(ctx, &QueueRequest{
		SessionID: "music-queue-1",
		TrackIDs:  []int64{10, 11},
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(session.Queue), "queue should have 3 tracks (1 initial + 2 added)")
}

func TestMusicPlayerService_SetRepeatMode(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Repeat Track", Duration: 200000},
	}
	createAndSaveMusicSession(t, db, "music-repeat-1", 1, tracks, 0)

	// Set repeat mode
	repeatAll := RepeatModeAll
	session, err := svc.UpdatePlayback(ctx, &UpdatePlaybackRequest{
		SessionID:  "music-repeat-1",
		RepeatMode: &repeatAll,
	})
	require.NoError(t, err)
	assert.Equal(t, RepeatModeAll, session.RepeatMode)
}

func TestMusicPlayerService_SetShuffleMode(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Track A", Duration: 200000},
		{ID: 2, Title: "Track B", Duration: 250000},
		{ID: 3, Title: "Track C", Duration: 300000},
	}
	createAndSaveMusicSession(t, db, "music-shuffle-1", 1, tracks, 0)

	// Enable shuffle
	shuffleOn := true
	session, err := svc.UpdatePlayback(ctx, &UpdatePlaybackRequest{
		SessionID: "music-shuffle-1",
		Shuffle:   &shuffleOn,
	})
	require.NoError(t, err)
	assert.True(t, session.ShuffleEnabled)
	assert.Equal(t, 3, len(session.Queue), "queue length should be preserved")
}

func TestMusicPlayerService_UpdatePlayback_MultipleFields(t *testing.T) {
	svc, db := setupMusicService(t)
	ctx := context.Background()

	tracks := []MusicTrack{
		{ID: 1, Title: "Update Track", Duration: 200000},
	}
	createAndSaveMusicSession(t, db, "music-update-1", 1, tracks, 0)

	vol := 0.7
	muted := true
	pauseState := PlaybackStatePaused
	session, err := svc.UpdatePlayback(ctx, &UpdatePlaybackRequest{
		SessionID: "music-update-1",
		Volume:    &vol,
		IsMuted:   &muted,
		State:     &pauseState,
	})
	require.NoError(t, err)
	assert.Equal(t, 0.7, session.Volume)
	assert.True(t, session.IsMuted)
	assert.Equal(t, PlaybackStatePaused, session.PlaybackState)
}

func TestMusicPlayerService_ShuffleQueue(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMusicPlayerService(nil, logger, nil, nil, nil, nil, nil, nil)

	session := &MusicPlaybackSession{
		Queue: []MusicTrack{
			{ID: 1, Title: "A"},
			{ID: 2, Title: "B"},
			{ID: 3, Title: "C"},
			{ID: 4, Title: "D"},
			{ID: 5, Title: "E"},
		},
		QueueIndex: 2, // Current track is C
		CurrentTrack: &MusicTrack{ID: 3, Title: "C"},
	}

	svc.shuffleQueue(session)

	assert.Equal(t, 5, len(session.Queue), "all tracks should remain")
	assert.Equal(t, 0, session.QueueIndex, "current track should be at index 0")
	assert.Equal(t, int64(3), session.Queue[0].ID, "current track should be first")
}

func TestMusicPlayerService_ShuffleQueue_SingleTrack(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMusicPlayerService(nil, logger, nil, nil, nil, nil, nil, nil)

	session := &MusicPlaybackSession{
		Queue:      []MusicTrack{{ID: 1, Title: "Only"}},
		QueueIndex: 0,
		CurrentTrack: &MusicTrack{ID: 1, Title: "Only"},
	}

	svc.shuffleQueue(session)
	assert.Equal(t, 1, len(session.Queue))
}

func TestMusicPlayerService_UnshuffleQueue(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMusicPlayerService(nil, logger, nil, nil, nil, nil, nil, nil)

	session := &MusicPlaybackSession{
		Queue: []MusicTrack{
			{ID: 3, Title: "C"},
			{ID: 1, Title: "A"},
			{ID: 5, Title: "E"},
			{ID: 2, Title: "B"},
			{ID: 4, Title: "D"},
		},
		QueueIndex:     0,
		CurrentTrack:   &MusicTrack{ID: 3, Title: "C"},
		ShuffleHistory: []int{0, 1, 2},
	}

	svc.unshuffleQueue(session)

	// Should be sorted by ID
	assert.Equal(t, int64(1), session.Queue[0].ID)
	assert.Equal(t, int64(2), session.Queue[1].ID)
	assert.Equal(t, int64(3), session.Queue[2].ID)
	assert.Equal(t, int64(4), session.Queue[3].ID)
	assert.Equal(t, int64(5), session.Queue[4].ID)
	// Current track (ID 3) should be at index 2
	assert.Equal(t, 2, session.QueueIndex)
	assert.Empty(t, session.ShuffleHistory)
}

// ==========================================================================
// 5. UniversalRenameTracker — real-SQLite tests
// ==========================================================================

func setupRenameTracker(t *testing.T) (*UniversalRenameTracker, *database.DB) {
	db := setupExtendedTestDB(t)
	logger := zap.NewNop()
	tracker := NewUniversalRenameTracker(db, logger)
	return tracker, db
}

func TestUniversalRenameTracker_InitializeTables(t *testing.T) {
	tracker, db := setupRenameTracker(t)

	err := tracker.InitializeTables()
	require.NoError(t, err)

	// Verify table was created by inserting a test row
	_, err = db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'test-root', 'local')")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO universal_rename_events
		(storage_root_id, protocol, old_path, new_path, is_directory, size, detected_at, status)
		VALUES (1, 'local', '/old/path', '/new/path', 0, 1024, CURRENT_TIMESTAMP, 'pending')`)
	require.NoError(t, err)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM universal_rename_events").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUniversalRenameTracker_InitializeTables_Idempotent(t *testing.T) {
	tracker, _ := setupRenameTracker(t)

	// Should succeed on first call
	err := tracker.InitializeTables()
	require.NoError(t, err)

	// Should succeed on second call (IF NOT EXISTS)
	err = tracker.InitializeTables()
	require.NoError(t, err)
}

func TestUniversalRenameTracker_TrackDeleteAndDetectCreate(t *testing.T) {
	tracker, _ := setupRenameTracker(t)
	ctx := context.Background()

	hash := "abc123def456"

	// Track a file deletion
	tracker.TrackDelete(ctx, 42, "/old/file.txt", "test-root", "local",
		1024, &hash, false, nil)

	// Detect the corresponding create
	move, detected := tracker.DetectCreate(ctx, "/new/file.txt", "test-root", "local",
		1024, &hash, false, nil)

	assert.True(t, detected, "should detect the move")
	require.NotNil(t, move)
	assert.Equal(t, "/old/file.txt", move.Path)
	assert.Equal(t, int64(42), move.FileID)
	assert.Equal(t, "local", move.Protocol)
}

func TestUniversalRenameTracker_DetectCreate_NoMatch(t *testing.T) {
	tracker, _ := setupRenameTracker(t)
	ctx := context.Background()

	hash := "different-hash"
	_, detected := tracker.DetectCreate(ctx, "/new/file.txt", "test-root", "local",
		1024, &hash, false, nil)
	assert.False(t, detected, "should not detect move without prior delete")
}

func TestUniversalRenameTracker_ProcessMove(t *testing.T) {
	tracker, db := setupRenameTracker(t)
	ctx := context.Background()

	err := tracker.InitializeTables()
	require.NoError(t, err)

	// Set up storage root and file
	_, err = db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'test-root', 'local')")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO files (id, storage_root_id, path, name, size, is_directory, modified_at)
		VALUES (100, 1, '/old-document.txt', 'old-document.txt', 2048, 0, CURRENT_TIMESTAMP)`)
	require.NoError(t, err)

	move := &UniversalPendingMove{
		Path:        "/old-document.txt",
		StorageRoot: "test-root",
		Protocol:    "local",
		Size:        2048,
		IsDirectory: false,
		DeletedAt:   time.Now(),
		FileID:      100,
	}

	// Use root-level destination so moveFile skips parent directory lookup
	// (filepath.Dir("/new-document.txt") == "/", which is excluded from the lookup).
	err = tracker.ProcessMove(ctx, &stubFSClient{}, move, "/new-document.txt")
	require.NoError(t, err)

	// Verify the file path was updated
	var newPath string
	err = db.QueryRow("SELECT path FROM files WHERE id = 100").Scan(&newPath)
	require.NoError(t, err)
	assert.Equal(t, "/new-document.txt", newPath)

	// Verify rename event was recorded
	var status string
	err = db.QueryRow("SELECT status FROM universal_rename_events ORDER BY id DESC LIMIT 1").Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "processed", status)
}

func TestUniversalRenameTracker_GetStatistics(t *testing.T) {
	tracker, db := setupRenameTracker(t)
	ctx := context.Background()

	err := tracker.InitializeTables()
	require.NoError(t, err)

	// Insert some test data
	_, err = db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'stat-root', 'local')")
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO universal_rename_events
		(storage_root_id, protocol, old_path, new_path, is_directory, size, detected_at, status)
		VALUES (1, 'local', '/a', '/b', 0, 100, CURRENT_TIMESTAMP, 'processed')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO universal_rename_events
		(storage_root_id, protocol, old_path, new_path, is_directory, size, detected_at, status)
		VALUES (1, 'smb', '/c', '/d', 0, 200, CURRENT_TIMESTAMP, 'failed')`)
	require.NoError(t, err)

	// Add a pending move to in-memory state
	hash := "test-hash"
	tracker.TrackDelete(ctx, 1, "/e", "stat-root", "local", 300, &hash, false, nil)

	stats := tracker.GetStatistics()
	require.NotNil(t, stats)

	assert.Equal(t, 1, stats["total_pending_moves"])
	assert.Equal(t, 2, stats["total_renames"])
	assert.Equal(t, 1, stats["successful_renames"])
	assert.Equal(t, 50.0, stats["success_rate"])

	pending := stats["pending_by_protocol"].(map[string]int)
	assert.Equal(t, 1, pending["local"])
}

func TestUniversalRenameTracker_CleanupExpiredMoves(t *testing.T) {
	tracker, _ := setupRenameTracker(t)

	// Manually add an expired move
	tracker.pendingMovesMu.Lock()
	tracker.pendingMoves["test:expired:key"] = &UniversalPendingMove{
		Path:        "/expired/file.txt",
		StorageRoot: "test",
		Protocol:    "local",
		Size:        100,
		DeletedAt:   time.Now().Add(-1 * time.Minute), // well past any move window
	}
	tracker.pendingMovesMu.Unlock()

	// Run cleanup
	tracker.cleanupExpiredMoves()

	// Verify the expired move was removed
	tracker.pendingMovesMu.RLock()
	_, exists := tracker.pendingMoves["test:expired:key"]
	tracker.pendingMovesMu.RUnlock()
	assert.False(t, exists, "expired move should be cleaned up")
}

func TestUniversalRenameTracker_CleanupExpiredMoves_KeepsFresh(t *testing.T) {
	tracker, _ := setupRenameTracker(t)
	ctx := context.Background()

	// Add a fresh move
	hash := "fresh-hash"
	tracker.TrackDelete(ctx, 1, "/fresh/file.txt", "test", "local", 100, &hash, false, nil)

	// Run cleanup
	tracker.cleanupExpiredMoves()

	// Verify the fresh move is still there
	stats := tracker.GetStatistics()
	assert.Equal(t, 1, stats["total_pending_moves"])
}

func TestUniversalRenameTracker_RecordRenameEvent(t *testing.T) {
	tracker, db := setupRenameTracker(t)
	ctx := context.Background()

	err := tracker.InitializeTables()
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'event-root', 'local')")
	require.NoError(t, err)

	hash := "event-hash"
	move := &UniversalPendingMove{
		Path:        "/event/old.txt",
		StorageRoot: "event-root",
		Protocol:    "local",
		Size:        512,
		FileHash:    &hash,
		IsDirectory: false,
		DeletedAt:   time.Now(),
		FileID:      1,
	}

	tx, err := db.Begin()
	require.NoError(t, err)

	eventID, err := tracker.recordUniversalRenameEvent(ctx, tx, move, "/event/new.txt")
	require.NoError(t, err)
	assert.Greater(t, eventID, int64(0))

	err = tx.Commit()
	require.NoError(t, err)

	// Verify the event was recorded
	var protocol, oldPath, newPath, status string
	err = db.QueryRow(`SELECT protocol, old_path, new_path, status FROM universal_rename_events WHERE id = ?`, eventID).
		Scan(&protocol, &oldPath, &newPath, &status)
	require.NoError(t, err)
	assert.Equal(t, "local", protocol)
	assert.Equal(t, "/event/old.txt", oldPath)
	assert.Equal(t, "/event/new.txt", newPath)
	assert.Equal(t, "pending", status)
}

func TestUniversalRenameTracker_StartStop(t *testing.T) {
	tracker, _ := setupRenameTracker(t)

	err := tracker.InitializeTables()
	require.NoError(t, err)

	err = tracker.Start()
	require.NoError(t, err)

	// Should be running; stop it
	tracker.Stop()
	// Should not panic on double-stop if stopCh is already closed
}

func TestUniversalRenameTracker_TrackDelete_UnknownProtocol(t *testing.T) {
	tracker, _ := setupRenameTracker(t)
	ctx := context.Background()

	// Should not panic, just log a warning
	tracker.TrackDelete(ctx, 1, "/test", "root", "unknown_protocol", 100, nil, false, nil)

	stats := tracker.GetStatistics()
	assert.Equal(t, 0, stats["total_pending_moves"], "unknown protocol should not add pending move")
}

func TestUniversalRenameTracker_DetectCreate_UnknownProtocol(t *testing.T) {
	tracker, _ := setupRenameTracker(t)
	ctx := context.Background()

	_, detected := tracker.DetectCreate(ctx, "/test", "root", "unknown_protocol", 100, nil, false, nil)
	assert.False(t, detected)
}
