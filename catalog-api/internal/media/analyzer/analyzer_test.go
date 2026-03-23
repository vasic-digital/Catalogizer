package analyzer

import (
	"catalogizer/database"
	"catalogizer/internal/media/detector"
	mediamodels "catalogizer/internal/media/models"
	"catalogizer/internal/media/providers"
	"catalogizer/internal/models"
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mutecomm/go-sqlcipher"
	"go.uber.org/zap"
)

func testLogger(t *testing.T) *zap.Logger {
	t.Helper()
	return zap.NewNop()
}

// setupTestDB creates an in-memory SQLite database with the media schema for testing.
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	rawDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	rawDB.SetMaxOpenConns(1)

	db := database.WrapDB(rawDB, database.DialectSQLite)
	require.NotNil(t, db)

	schema := `
	CREATE TABLE IF NOT EXISTS storage_roots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		protocol TEXT NOT NULL DEFAULT 'smb',
		path TEXT,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		storage_root_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		extension TEXT,
		mime_type TEXT,
		size INTEGER NOT NULL DEFAULT 0,
		is_directory BOOLEAN DEFAULT 0,
		modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_scan_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id)
	);

	CREATE TABLE IF NOT EXISTS media_types (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		detection_patterns TEXT,
		metadata_providers TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS media_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_type_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		original_title TEXT,
		year INTEGER,
		description TEXT,
		genre TEXT,
		director TEXT,
		cast_crew TEXT,
		rating REAL,
		runtime INTEGER,
		language TEXT,
		country TEXT,
		status TEXT DEFAULT 'active',
		first_detected DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_type_id) REFERENCES media_types(id)
	);

	CREATE TABLE IF NOT EXISTS external_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_item_id INTEGER NOT NULL,
		provider TEXT NOT NULL,
		external_id TEXT NOT NULL,
		data TEXT,
		rating REAL,
		review_url TEXT,
		cover_url TEXT,
		trailer_url TEXT,
		last_fetched DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id),
		UNIQUE(media_item_id, provider)
	);

	CREATE TABLE IF NOT EXISTS directory_analysis (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		directory_path TEXT NOT NULL UNIQUE,
		smb_root TEXT NOT NULL,
		media_item_id INTEGER,
		confidence_score REAL NOT NULL,
		detection_method TEXT NOT NULL,
		analysis_data TEXT,
		last_analyzed DATETIME DEFAULT CURRENT_TIMESTAMP,
		files_count INTEGER DEFAULT 0,
		total_size INTEGER DEFAULT 0,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id)
	);

	CREATE TABLE IF NOT EXISTS media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_item_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		smb_root TEXT NOT NULL,
		filename TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		file_extension TEXT,
		quality_info TEXT,
		language TEXT,
		subtitle_tracks TEXT,
		audio_tracks TEXT,
		duration INTEGER,
		checksum TEXT,
		virtual_smb_link TEXT,
		direct_smb_link TEXT,
		last_verified DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (media_item_id) REFERENCES media_items(id)
	);

	INSERT OR IGNORE INTO media_types (id, name, description) VALUES
		(1, 'movie', 'Feature films'),
		(2, 'tv_show', 'Television series'),
		(3, 'music', 'Music albums and tracks'),
		(4, 'game', 'Video games'),
		(5, 'software', 'Applications and utilities'),
		(6, 'comic', 'Comic books');
	`

	_, err = rawDB.Exec(schema)
	require.NoError(t, err)

	t.Cleanup(func() { rawDB.Close() })
	return db
}

// seedStorageRootAndFiles inserts a storage root and files for testing.
func seedStorageRootAndFiles(t *testing.T, db *database.DB, rootName string, files []struct {
	path      string
	name      string
	ext       *string
	mime      *string
	size      int64
	isDir     bool
}) {
	t.Helper()
	_, err := db.Exec("INSERT OR IGNORE INTO storage_roots (name, protocol) VALUES (?, 'smb')", rootName)
	require.NoError(t, err)

	for _, f := range files {
		_, err := db.Exec(
			`INSERT INTO files (storage_root_id, path, name, extension, mime_type, size, is_directory, modified_at)
			 VALUES ((SELECT id FROM storage_roots WHERE name = ? LIMIT 1), ?, ?, ?, ?, ?, ?, ?)`,
			rootName, f.path, f.name, f.ext, f.mime, f.size, f.isDir, time.Now(),
		)
		require.NoError(t, err)
	}
}

// --- Helpers ---

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// ============================================================================
// extractQualityFromFilename tests
// ============================================================================

func TestExtractQualityFromFilename(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	tests := []struct {
		name             string
		filename         string
		extension        *string
		expectResolution *mediamodels.Resolution
		expectSource     *string
		expectVideoCodec *string
		expectAudioCodec *string
		expectHDR        bool
		minScore         int
		maxScore         int
	}{
		{
			name:             "4K UHD BluRay with HDR and HEVC",
			filename:         "Movie.2160p.BluRay.HDR.x265.DTS.mkv",
			extension:        strPtr(".mkv"),
			expectResolution: &mediamodels.Resolution{Width: 3840, Height: 2160},
			expectSource:     strPtr("BluRay"),
			expectVideoCodec: strPtr("H.265/HEVC"),
			expectAudioCodec: strPtr("DTS"),
			expectHDR:        true,
			minScore:         120,
			maxScore:         200,
		},
		{
			name:             "1080p WEB-DL with AAC",
			filename:         "Show.S01E01.1080p.WEB-DL.AAC.x264.mkv",
			extension:        strPtr(".mkv"),
			expectResolution: &mediamodels.Resolution{Width: 1920, Height: 1080},
			expectSource:     strPtr("WEB-DL"),
			expectVideoCodec: strPtr("H.264/AVC"),
			expectAudioCodec: strPtr("AAC"),
			expectHDR:        false,
			minScore:         80,
			maxScore:         100,
		},
		{
			name:             "720p HD with AC3",
			filename:         "Movie.720p.BRRip.AC3.mkv",
			extension:        strPtr(".mkv"),
			expectResolution: &mediamodels.Resolution{Width: 1280, Height: 720},
			expectSource:     strPtr("BluRay"),
			expectVideoCodec: nil,
			expectAudioCodec: strPtr("AC3"),
			expectHDR:        false,
			minScore:         60,
			maxScore:         85,
		},
		{
			name:             "480p DVD quality",
			filename:         "Movie.DVDRip.480p.avi",
			extension:        strPtr(".avi"),
			expectResolution: &mediamodels.Resolution{Width: 720, Height: 480},
			expectSource:     nil,
			expectVideoCodec: nil,
			expectAudioCodec: nil,
			expectHDR:        false,
			minScore:         40,
			maxScore:         50,
		},
		{
			name:             "FLAC lossless audio",
			filename:         "track01.flac",
			extension:        strPtr(".flac"),
			expectResolution: nil,
			expectSource:     nil,
			expectVideoCodec: nil,
			expectAudioCodec: nil,
			expectHDR:        false,
			minScore:         90,
			maxScore:         100,
		},
		{
			name:             "MP3 320k audio",
			filename:         "song.320k.mp3",
			extension:        strPtr(".mp3"),
			expectResolution: nil,
			expectSource:     nil,
			expectVideoCodec: nil,
			expectAudioCodec: nil,
			expectHDR:        false,
			minScore:         70,
			maxScore:         80,
		},
		{
			name:             "plain MP3",
			filename:         "podcast_episode.mp3",
			extension:        strPtr(".mp3"),
			expectResolution: nil,
			expectSource:     nil,
			expectVideoCodec: nil,
			expectAudioCodec: nil,
			expectHDR:        false,
			minScore:         50,
			maxScore:         60,
		},
		{
			name:             "4K with Dolby Vision",
			filename:         "Movie.4K.Dolby.Vision.mkv",
			extension:        strPtr(".mkv"),
			expectResolution: &mediamodels.Resolution{Width: 3840, Height: 2160},
			expectHDR:        true,
			minScore:         110,
			maxScore:         150,
		},
		{
			name:             "WEBRip source",
			filename:         "Show.WEBRip.1080p.mkv",
			extension:        strPtr(".mkv"),
			expectResolution: &mediamodels.Resolution{Width: 1920, Height: 1080},
			expectSource:     strPtr("WEB-RIP"),
			minScore:         80,
			maxScore:         90,
		},
		{
			name:             "no quality indicators",
			filename:         "random_file.txt",
			extension:        strPtr(".txt"),
			expectResolution: nil,
			expectSource:     nil,
			expectVideoCodec: nil,
			expectAudioCodec: nil,
			expectHDR:        false,
			minScore:         0,
			maxScore:         5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qi := ma.extractQualityFromFilename(tc.filename, tc.extension)
			require.NotNil(t, qi)

			if tc.expectResolution != nil {
				require.NotNil(t, qi.Resolution)
				assert.Equal(t, tc.expectResolution.Width, qi.Resolution.Width)
				assert.Equal(t, tc.expectResolution.Height, qi.Resolution.Height)
			}

			if tc.expectSource != nil {
				require.NotNil(t, qi.Source)
				assert.Equal(t, *tc.expectSource, *qi.Source)
			}

			if tc.expectVideoCodec != nil {
				require.NotNil(t, qi.VideoCodec)
				assert.Equal(t, *tc.expectVideoCodec, *qi.VideoCodec)
			}

			if tc.expectAudioCodec != nil {
				require.NotNil(t, qi.AudioCodec)
				assert.Equal(t, *tc.expectAudioCodec, *qi.AudioCodec)
			}

			assert.Equal(t, tc.expectHDR, qi.HDR)
			assert.GreaterOrEqual(t, qi.QualityScore, tc.minScore, "quality score too low")
			assert.LessOrEqual(t, qi.QualityScore, tc.maxScore, "quality score too high")
		})
	}
}

func TestExtractQualityFromFilename_WAVFile(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	ext := ".wav"
	qi := ma.extractQualityFromFilename("recording.wav", &ext)

	require.NotNil(t, qi)
	assert.Equal(t, 90, qi.QualityScore)
	assert.NotNil(t, qi.QualityProfile)
	assert.Equal(t, "Audio_Lossless", *qi.QualityProfile)
}

func TestExtractQualityFromFilename_NilExtension(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("movie.2160p.x265.mkv", nil)

	require.NotNil(t, qi)
	assert.NotNil(t, qi.Resolution)
	assert.Equal(t, 3840, qi.Resolution.Width)
	assert.Equal(t, 2160, qi.Resolution.Height)
	assert.NotNil(t, qi.VideoCodec)
	assert.Equal(t, "H.265/HEVC", *qi.VideoCodec)
}

func TestExtractQualityFromFilename_FHDAlias(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	ext := ".mkv"
	qi := ma.extractQualityFromFilename("Movie.FHD.mkv", &ext)

	require.NotNil(t, qi)
	assert.NotNil(t, qi.Resolution)
	assert.Equal(t, 1920, qi.Resolution.Width)
	assert.Equal(t, 1080, qi.Resolution.Height)
}

func TestExtractQualityFromFilename_UHDAlias(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	ext := ".mkv"
	qi := ma.extractQualityFromFilename("Movie.UHD.BluRay.mkv", &ext)

	require.NotNil(t, qi)
	assert.NotNil(t, qi.Resolution)
	assert.Equal(t, 3840, qi.Resolution.Width)
	assert.Equal(t, 2160, qi.Resolution.Height)
	assert.NotNil(t, qi.Source)
	assert.Equal(t, "BluRay", *qi.Source)
}

func TestExtractQualityFromFilename_H264AVC(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	ext := ".mkv"
	qi := ma.extractQualityFromFilename("Movie.1080p.AVC.mkv", &ext)

	require.NotNil(t, qi)
	assert.NotNil(t, qi.VideoCodec)
	assert.Equal(t, "H.264/AVC", *qi.VideoCodec)
}

func TestExtractQualityFromFilename_AC3Audio(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	ext := ".mkv"
	qi := ma.extractQualityFromFilename("Movie.720p.AC3.mkv", &ext)

	require.NotNil(t, qi)
	assert.NotNil(t, qi.AudioCodec)
	assert.Equal(t, "AC3", *qi.AudioCodec)
}

func TestExtractQualityFromFilename_H265Variant(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.1080p.H265.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	require.NotNil(t, qi.VideoCodec)
	assert.Equal(t, "H.265/HEVC", *qi.VideoCodec)
}

func TestExtractQualityFromFilename_HEVCVariant(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.1080p.HEVC.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	require.NotNil(t, qi.VideoCodec)
	assert.Equal(t, "H.265/HEVC", *qi.VideoCodec)
}

func TestExtractQualityFromFilename_H264Variant(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.1080p.H264.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	require.NotNil(t, qi.VideoCodec)
	assert.Equal(t, "H.264/AVC", *qi.VideoCodec)
}

func TestExtractQualityFromFilename_WebDLVariant(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.1080p.WEBDL.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	require.NotNil(t, qi.Source)
	assert.Equal(t, "WEB-DL", *qi.Source)
}

func TestExtractQualityFromFilename_HDRPlain(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.2160p.HDR.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	assert.True(t, qi.HDR)
}

func TestExtractQualityFromFilename_DTS_Audio(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.1080p.DTS.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	require.NotNil(t, qi.AudioCodec)
	assert.Equal(t, "DTS", *qi.AudioCodec)
}

func TestExtractQualityFromFilename_AAC_Audio(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("Movie.720p.AAC.mkv", strPtr(".mkv"))
	require.NotNil(t, qi)
	require.NotNil(t, qi.AudioCodec)
	assert.Equal(t, "AAC", *qi.AudioCodec)
}

func TestExtractQualityFromFilename_320kMP3(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	qi := ma.extractQualityFromFilename("song.320.mp3", strPtr(".mp3"))
	require.NotNil(t, qi)
	assert.Equal(t, 70, qi.QualityScore)
	require.NotNil(t, qi.QualityProfile)
	assert.Equal(t, "Audio_320k", *qi.QualityProfile)
}

// ============================================================================
// filterMediaFiles tests (using real function with models.FileInfo)
// ============================================================================

func TestFilterMediaFiles_RealFunction(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	mkv := ".mkv"
	mp4 := ".mp4"
	mp3 := ".mp3"
	flac := ".flac"
	txt := ".txt"
	exe := ".exe"
	cbr := ".cbr"

	allFiles := []models.FileInfo{
		{Name: "movie.mkv", Extension: &mkv, Size: 5_000_000_000},
		{Name: "clip.mp4", Extension: &mp4, Size: 200_000_000},
		{Name: "track.mp3", Extension: &mp3, Size: 5_000_000},
		{Name: "album.flac", Extension: &flac, Size: 30_000_000},
		{Name: "notes.txt", Extension: &txt, Size: 1000},
		{Name: "Extras", IsDirectory: true},
		{Name: "setup.exe", Extension: &exe, Size: 50_000_000},
		{Name: "comic.cbr", Extension: &cbr, Size: 20_000_000},
		{Name: "no_ext_file", Extension: nil, Size: 500},
	}

	tests := []struct {
		name      string
		mediaType string
		expectLen int
	}{
		{"movie filters video files", "movie", 2},
		{"tv_show filters video files", "tv_show", 2},
		{"anime filters video files", "anime", 2},
		{"music filters audio files", "music", 2},
		{"audiobook filters audio files", "audiobook", 1},
		{"podcast filters audio files", "podcast", 1},
		{"comic filters comic files", "comic", 1},
		{"software filters executables", "software", 1},
		{"game filters game files", "game", 1},
		{"unknown type returns all files", "unknown_xyz", 9},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filtered := ma.filterMediaFiles(allFiles, tc.mediaType)
			assert.Len(t, filtered, tc.expectLen)
		})
	}
}

func TestFilterMediaFiles_EmptyFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	filtered := ma.filterMediaFiles(nil, "movie")
	assert.Empty(t, filtered)
}

func TestFilterMediaFiles_AllDirectories(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	files := []models.FileInfo{
		{Name: "Season1", IsDirectory: true},
		{Name: "Season2", IsDirectory: true},
	}

	filtered := ma.filterMediaFiles(files, "movie")
	assert.Empty(t, filtered)
}

func TestFilterMediaFiles_CaseInsensitiveExtension(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	MKV := ".MKV"
	files := []models.FileInfo{
		{Name: "Movie.MKV", Extension: &MKV, Size: 5_000_000_000},
	}

	filtered := ma.filterMediaFiles(files, "movie")
	assert.Len(t, filtered, 1)
}

func TestFilterMediaFiles_NilExtensionSkipped(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	files := []models.FileInfo{
		{Name: "noext", Extension: nil, Size: 100},
	}

	filtered := ma.filterMediaFiles(files, "movie")
	assert.Empty(t, filtered)
}

// ============================================================================
// QualityInfo / Resolution helper tests
// ============================================================================

func TestQualityInfo_IsBetterThan(t *testing.T) {
	tests := []struct {
		name     string
		qi       *mediamodels.QualityInfo
		other    *mediamodels.QualityInfo
		expected bool
	}{
		{"higher score is better", &mediamodels.QualityInfo{QualityScore: 100}, &mediamodels.QualityInfo{QualityScore: 80}, true},
		{"lower score is not better", &mediamodels.QualityInfo{QualityScore: 50}, &mediamodels.QualityInfo{QualityScore: 80}, false},
		{"equal scores not better", &mediamodels.QualityInfo{QualityScore: 80}, &mediamodels.QualityInfo{QualityScore: 80}, false},
		{"non-nil is better than nil", &mediamodels.QualityInfo{QualityScore: 50}, nil, true},
		{"nil is not better", nil, &mediamodels.QualityInfo{QualityScore: 50}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.qi.IsBetterThan(tc.other))
		})
	}
}

func TestQualityInfo_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		qi       *mediamodels.QualityInfo
		expected string
	}{
		{"uses quality profile when set", &mediamodels.QualityInfo{QualityProfile: strPtr("4K/UHD")}, "4K/UHD"},
		{"uses resolution display name as fallback", &mediamodels.QualityInfo{Resolution: &mediamodels.Resolution{Width: 1920, Height: 1080}}, "1080p"},
		{"returns Unknown when nothing is set", &mediamodels.QualityInfo{}, "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.qi.GetDisplayName())
		})
	}
}

func TestResolution_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		res      mediamodels.Resolution
		expected string
	}{
		{"4K", mediamodels.Resolution{Width: 3840, Height: 2160}, "4K/UHD"},
		{"1080p", mediamodels.Resolution{Width: 1920, Height: 1080}, "1080p"},
		{"720p", mediamodels.Resolution{Width: 1280, Height: 720}, "720p"},
		{"480p", mediamodels.Resolution{Width: 720, Height: 480}, "480p/DVD"},
		{"low quality", mediamodels.Resolution{Width: 640, Height: 360}, "Low Quality"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.res.GetDisplayName())
		})
	}
}

// ============================================================================
// contains tests
// ============================================================================

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"nil slice", nil, "a", false},
		{"single element found", []string{"x"}, "x", true},
		{"single element not found", []string{"x"}, "y", false},
		{"first element", []string{"a", "b", "c"}, "a", true},
		{"last element", []string{"a", "b", "c"}, "c", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, contains(tc.slice, tc.item))
		})
	}
}

// ============================================================================
// NewMediaAnalyzer tests
// ============================================================================

func TestNewMediaAnalyzer(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	assert.NotNil(t, ma)
	assert.Nil(t, ma.db)
	assert.Nil(t, ma.detector)
	assert.Nil(t, ma.providerManager)
	assert.NotNil(t, ma.logger)
	assert.NotNil(t, ma.analysisQueue)
	assert.Equal(t, 4, ma.workers)
	assert.NotNil(t, ma.stopCh)
	assert.NotNil(t, ma.pendingAnalysis)
	assert.Empty(t, ma.pendingAnalysis)
}

func TestNewMediaAnalyzer_WithAllDependencies(t *testing.T) {
	logger := testLogger(t)
	db := setupTestDB(t)
	det := detector.NewDetectionEngine(logger)
	pm := providers.NewProviderManager(logger)

	ma := NewMediaAnalyzer(db, det, pm, logger)

	assert.NotNil(t, ma)
	assert.NotNil(t, ma.db)
	assert.NotNil(t, ma.detector)
	assert.NotNil(t, ma.providerManager)
	assert.Equal(t, 1000, cap(ma.analysisQueue))
}

// ============================================================================
// Start/Stop lifecycle tests
// ============================================================================

func TestMediaAnalyzer_StartStop(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ma.Start()
	ma.Stop()
}

func TestMediaAnalyzer_StopWithoutStart(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	close(ma.stopCh)
	ma.wg.Wait()
}

func TestMediaAnalyzer_MultipleStartStop(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ma.Start()
	ma.Stop()

	// After stop, create fresh channels for another cycle
	ma.stopCh = make(chan struct{})
	ma.Start()
	ma.Stop()
}

// ============================================================================
// AnalyzeDirectory queue tests
// ============================================================================

func TestMediaAnalyzer_AnalyzeDirectory_QueueRequest(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ctx := context.Background()

	err := ma.AnalyzeDirectory(ctx, "/test/path", "root1", 5)
	assert.NoError(t, err)

	ma.mu.RLock()
	_, exists := ma.pendingAnalysis["/test/path"]
	ma.mu.RUnlock()
	assert.True(t, exists)
}

func TestMediaAnalyzer_AnalyzeDirectory_DuplicateRequest(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ctx := context.Background()

	err := ma.AnalyzeDirectory(ctx, "/test/path", "root1", 5)
	assert.NoError(t, err)

	err = ma.AnalyzeDirectory(ctx, "/test/path", "root1", 10)
	assert.NoError(t, err)

	ma.mu.RLock()
	pending := ma.pendingAnalysis["/test/path"]
	ma.mu.RUnlock()
	assert.NotNil(t, pending)
	assert.Equal(t, 10, pending.Priority)
}

func TestMediaAnalyzer_AnalyzeDirectory_DuplicateWithLowerPriority(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ctx := context.Background()

	err := ma.AnalyzeDirectory(ctx, "/test/path", "root1", 10)
	assert.NoError(t, err)

	err = ma.AnalyzeDirectory(ctx, "/test/path", "root1", 3)
	assert.NoError(t, err)

	ma.mu.RLock()
	pending := ma.pendingAnalysis["/test/path"]
	ma.mu.RUnlock()
	assert.NotNil(t, pending)
	assert.Equal(t, 10, pending.Priority)
}

func TestMediaAnalyzer_AnalyzeDirectory_CancelledContext(t *testing.T) {
	logger := testLogger(t)
	ma := &MediaAnalyzer{
		logger:          logger,
		analysisQueue:   make(chan AnalysisRequest, 0),
		stopCh:          make(chan struct{}),
		pendingAnalysis: make(map[string]*AnalysisRequest),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ma.AnalyzeDirectory(ctx, "/test/path", "root1", 5)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	ma.mu.RLock()
	_, exists := ma.pendingAnalysis["/test/path"]
	ma.mu.RUnlock()
	assert.False(t, exists)
}

func TestMediaAnalyzer_AnalyzeDirectory_MultipleDifferentPaths(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ctx := context.Background()

	require.NoError(t, ma.AnalyzeDirectory(ctx, "/path/a", "root1", 1))
	require.NoError(t, ma.AnalyzeDirectory(ctx, "/path/b", "root1", 2))
	require.NoError(t, ma.AnalyzeDirectory(ctx, "/path/c", "root1", 3))

	ma.mu.RLock()
	assert.Len(t, ma.pendingAnalysis, 3)
	ma.mu.RUnlock()
}

// ============================================================================
// getDirectoryFiles tests (database-dependent)
// ============================================================================

func TestGetDirectoryFiles_ReturnsFiles(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	seedStorageRootAndFiles(t, db, "test-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/movies/Inception/Inception.2160p.mkv", "Inception.2160p.mkv", strPtr(".mkv"), strPtr("video/x-matroska"), 15_000_000_000, false},
		{"/movies/Inception/Inception.srt", "Inception.srt", strPtr(".srt"), strPtr("text/plain"), 50000, false},
		{"/movies/Inception/Extras", "Extras", nil, nil, 0, true},
	})

	files, err := ma.getDirectoryFiles("/movies/Inception/", "test-root")
	require.NoError(t, err)
	assert.Len(t, files, 3)

	// Directories should come first (ORDER BY is_directory DESC)
	assert.True(t, files[0].IsDirectory)
	assert.Equal(t, "Extras", files[0].Name)
}

func TestGetDirectoryFiles_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	// No files seeded for this root
	files, err := ma.getDirectoryFiles("/nonexistent/", "missing-root")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestGetDirectoryFiles_PathFiltering(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	seedStorageRootAndFiles(t, db, "multi-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/movies/A/file1.mkv", "file1.mkv", strPtr(".mkv"), nil, 100, false},
		{"/movies/B/file2.mkv", "file2.mkv", strPtr(".mkv"), nil, 200, false},
		{"/music/C/track.mp3", "track.mp3", strPtr(".mp3"), nil, 50, false},
	})

	files, err := ma.getDirectoryFiles("/movies/A/", "multi-root")
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "file1.mkv", files[0].Name)
}

// ============================================================================
// createDirectoryAnalysis tests (database-dependent)
// ============================================================================

func TestCreateDirectoryAnalysis_BasicInsert(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	detection := &detector.DetectionResult{
		Confidence: 0.95,
		Method:     "filename_pattern",
		AnalysisData: &mediamodels.AnalysisData{
			FileTypes:        map[string]int{".mkv": 2, ".srt": 1},
			SizeDistribution: map[string]int64{"large": 15_000_000_000, "tiny": 50000},
		},
	}

	result, err := ma.createDirectoryAnalysis("/movies/Inception", "test-root", detection)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "/movies/Inception", result.DirectoryPath)
	assert.Equal(t, "test-root", result.SmbRoot)
	assert.Equal(t, 0.95, result.ConfidenceScore)
	assert.Equal(t, "filename_pattern", result.DetectionMethod)
	assert.Equal(t, 3, result.FilesCount)
	assert.Equal(t, int64(15_000_050_000), result.TotalSize)
}

func TestCreateDirectoryAnalysis_NilAnalysisData(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	detection := &detector.DetectionResult{
		Confidence:   0.5,
		Method:       "directory_structure",
		AnalysisData: nil,
	}

	result, err := ma.createDirectoryAnalysis("/movies/Test", "root1", detection)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.FilesCount)
	assert.Equal(t, int64(0), result.TotalSize)
}

func TestCreateDirectoryAnalysis_ReplacesExisting(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	detection1 := &detector.DetectionResult{
		Confidence:   0.5,
		Method:       "first",
		AnalysisData: nil,
	}
	_, err := ma.createDirectoryAnalysis("/same/path", "root", detection1)
	require.NoError(t, err)

	detection2 := &detector.DetectionResult{
		Confidence:   0.9,
		Method:       "second",
		AnalysisData: nil,
	}
	result, err := ma.createDirectoryAnalysis("/same/path", "root", detection2)
	require.NoError(t, err)
	assert.Equal(t, 0.9, result.ConfidenceScore)
	assert.Equal(t, "second", result.DetectionMethod)
}

// ============================================================================
// createOrUpdateMediaItem tests (database-dependent)
// ============================================================================

func TestCreateOrUpdateMediaItem_CreatesNew(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	year := 2010
	movieType := &mediamodels.MediaType{ID: 1, Name: "movie"}

	detection := &detector.DetectionResult{
		MediaTypeID:    1,
		MediaType:      movieType,
		SuggestedTitle: "Inception",
		SuggestedYear:  &year,
	}

	// First create a directory analysis record
	_, err := ma.createDirectoryAnalysis("/movies/Inception", "root1", &detector.DetectionResult{
		Confidence:   0.95,
		Method:       "test",
		AnalysisData: nil,
	})
	require.NoError(t, err)

	dirAnalysis := &mediamodels.DirectoryAnalysis{
		DirectoryPath: "/movies/Inception",
		SmbRoot:       "root1",
	}

	ctx := context.Background()
	mediaItem, err := ma.createOrUpdateMediaItem(ctx, detection, dirAnalysis)
	require.NoError(t, err)
	require.NotNil(t, mediaItem)

	assert.Equal(t, "Inception", mediaItem.Title)
	assert.Equal(t, &year, mediaItem.Year)
	assert.Equal(t, int64(1), mediaItem.MediaTypeID)
	assert.Equal(t, "active", mediaItem.Status)
	assert.NotZero(t, mediaItem.ID)
}

func TestCreateOrUpdateMediaItem_UpdatesExisting(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	ctx := context.Background()

	// Insert a media item first
	genreJSON := "[]"
	castCrewJSON := "{}"
	mediaItemID, err := db.InsertReturningID(ctx,
		`INSERT INTO media_items (media_type_id, title, year, description, genre, director, cast_crew, rating, runtime, language, country, status, first_detected, last_updated)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?)`,
		1, "Inception", 2010, nil, genreJSON, nil, castCrewJSON, nil, nil, nil, nil, time.Now(), time.Now(),
	)
	require.NoError(t, err)

	// Create a directory analysis pointing to that media item
	_, err = db.Exec(
		`INSERT INTO directory_analysis (directory_path, smb_root, media_item_id, confidence_score, detection_method)
		 VALUES (?, ?, ?, ?, ?)`,
		"/movies/Inception", "root1", mediaItemID, 0.95, "test",
	)
	require.NoError(t, err)

	detection := &detector.DetectionResult{
		MediaTypeID:    1,
		MediaType:      &mediamodels.MediaType{ID: 1, Name: "movie"},
		SuggestedTitle: "Inception",
	}
	dirAnalysis := &mediamodels.DirectoryAnalysis{
		DirectoryPath: "/movies/Inception",
	}

	item, err := ma.createOrUpdateMediaItem(ctx, detection, dirAnalysis)
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, mediaItemID, item.ID)
	assert.Equal(t, "Inception", item.Title)
}

// ============================================================================
// updateExistingMediaItem tests
// ============================================================================

func TestUpdateExistingMediaItem_Success(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	ctx := context.Background()

	genreJSON := `["sci-fi","action"]`
	castCrewJSON := `{"director":"Nolan"}`
	mediaItemID, err := db.InsertReturningID(ctx,
		`INSERT INTO media_items (media_type_id, title, year, description, genre, director, cast_crew, rating, runtime, language, country, status, first_detected, last_updated)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active', ?, ?)`,
		1, "Inception", 2010, "A thief", genreJSON, "Nolan", castCrewJSON, 8.8, 148, "en", "US", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	detection := &detector.DetectionResult{MediaTypeID: 1}
	item, err := ma.updateExistingMediaItem(mediaItemID, detection)
	require.NoError(t, err)
	require.NotNil(t, item)

	assert.Equal(t, mediaItemID, item.ID)
	assert.Equal(t, "Inception", item.Title)
	assert.Equal(t, "active", item.Status)
}

func TestUpdateExistingMediaItem_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	detection := &detector.DetectionResult{MediaTypeID: 1}
	item, err := ma.updateExistingMediaItem(99999, detection)
	assert.Error(t, err)
	assert.Nil(t, item)
}

// ============================================================================
// fetchExternalMetadata tests
// ============================================================================

func TestFetchExternalMetadata_NilMediaType(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	item := &mediamodels.MediaItem{
		MediaType: nil,
	}

	result, err := ma.fetchExternalMetadata(context.Background(), item)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "media type not available")
}

func TestFetchExternalMetadata_WithProviderManager_NoResults(t *testing.T) {
	logger := testLogger(t)
	db := setupTestDB(t)
	pm := providers.NewProviderManager(logger)
	ma := &MediaAnalyzer{logger: logger, db: db, providerManager: pm}

	item := &mediamodels.MediaItem{
		ID:        1,
		Title:     "NonexistentMovie12345",
		MediaType: &mediamodels.MediaType{ID: 1, Name: "movie"},
	}

	// All providers are disabled (no API keys), so GetBestMatch returns nil
	result, err := ma.fetchExternalMetadata(context.Background(), item)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

// ============================================================================
// analyzeQuality tests
// ============================================================================

func TestAnalyzeQuality_NilMediaType(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	item := &mediamodels.MediaItem{MediaType: nil}
	result, err := ma.analyzeQuality(nil, item)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "media type not available")
}

func TestAnalyzeQuality_MovieFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	mkv := ".mkv"
	avi := ".avi"
	srt := ".srt"

	files := []models.FileInfo{
		{Name: "Movie.2160p.BluRay.HDR.x265.DTS.mkv", Extension: &mkv, Size: 15_000_000_000},
		{Name: "Movie.1080p.mkv", Extension: &mkv, Size: 4_000_000_000},
		{Name: "Movie.DVDRip.480p.avi", Extension: &avi, Size: 700_000_000},
		{Name: "Movie.srt", Extension: &srt, Size: 50000},
		{Name: "Extras", IsDirectory: true},
	}

	item := &mediamodels.MediaItem{
		MediaType: &mediamodels.MediaType{ID: 1, Name: "movie"},
	}

	result, err := ma.analyzeQuality(files, item)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Only 3 video files should be counted (srt excluded, directory excluded)
	assert.Equal(t, 3, result.TotalFiles)
	assert.Equal(t, int64(15_000_000_000+4_000_000_000+700_000_000), result.TotalSize)
	assert.NotNil(t, result.BestQuality)
	assert.GreaterOrEqual(t, result.BestQuality.QualityScore, 100) // 4K should score highest
	assert.NotEmpty(t, result.AvailableQualities)
}

func TestAnalyzeQuality_MusicFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	flac := ".flac"
	mp3 := ".mp3"

	files := []models.FileInfo{
		{Name: "track01.flac", Extension: &flac, Size: 30_000_000},
		{Name: "track02.320k.mp3", Extension: &mp3, Size: 8_000_000},
		{Name: "track03.mp3", Extension: &mp3, Size: 4_000_000},
	}

	item := &mediamodels.MediaItem{
		MediaType: &mediamodels.MediaType{ID: 3, Name: "music"},
	}

	result, err := ma.analyzeQuality(files, item)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 3, result.TotalFiles)
	assert.NotNil(t, result.BestQuality)
	// FLAC should be the best quality (score 90)
	assert.GreaterOrEqual(t, result.BestQuality.QualityScore, 90)
}

func TestAnalyzeQuality_EmptyFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	item := &mediamodels.MediaItem{
		MediaType: &mediamodels.MediaType{ID: 1, Name: "movie"},
	}

	result, err := ma.analyzeQuality(nil, item)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.TotalFiles)
	assert.Empty(t, result.AvailableQualities)
	assert.Nil(t, result.BestQuality)
}

func TestAnalyzeQuality_UnknownMediaType(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	txt := ".txt"
	files := []models.FileInfo{
		{Name: "readme.txt", Extension: &txt, Size: 100},
	}

	item := &mediamodels.MediaItem{
		MediaType: &mediamodels.MediaType{ID: 99, Name: "unknown_type"},
	}

	// Unknown type passes all files through filterMediaFiles
	result, err := ma.analyzeQuality(files, item)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.TotalFiles)
}

func TestAnalyzeQuality_DuplicateQualityNames(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	mkv := ".mkv"
	files := []models.FileInfo{
		{Name: "Movie.Part1.1080p.mkv", Extension: &mkv, Size: 4_000_000_000},
		{Name: "Movie.Part2.1080p.mkv", Extension: &mkv, Size: 4_000_000_000},
	}

	item := &mediamodels.MediaItem{
		MediaType: &mediamodels.MediaType{ID: 1, Name: "movie"},
	}

	result, err := ma.analyzeQuality(files, item)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Both files have 1080p quality, should appear only once
	assert.Len(t, result.AvailableQualities, 1)
	assert.Equal(t, "1080p", result.AvailableQualities[0])
}

// ============================================================================
// updateMediaFiles tests (database-dependent)
// ============================================================================

func TestUpdateMediaFiles_InsertsRecords(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	ctx := context.Background()

	// Create a media item first
	mediaItemID, err := db.InsertReturningID(ctx,
		`INSERT INTO media_items (media_type_id, title, status, first_detected, last_updated)
		 VALUES (?, ?, 'active', ?, ?)`,
		1, "TestMovie", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	mkv := ".mkv"
	srt := ".srt"

	files := []models.FileInfo{
		{ID: 100, Name: "Movie.1080p.mkv", Path: "/movies/Test/Movie.1080p.mkv", Extension: &mkv, Size: 4_000_000_000},
		{ID: 101, Name: "Movie.srt", Path: "/movies/Test/Movie.srt", Extension: &srt, Size: 50000},
		{Name: "Extras", IsDirectory: true},
	}

	updatedFiles, err := ma.updateMediaFiles(mediaItemID, files, "/movies/Test", "root1")
	require.NoError(t, err)

	// Directory should be skipped
	assert.Len(t, updatedFiles, 2)

	// Verify inserted into database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM media_files WHERE media_item_id = ?", mediaItemID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Check SMB links are generated correctly
	assert.Equal(t, "smb://root1//movies/Test/Movie.1080p.mkv", updatedFiles[0].DirectSmbLink)
	assert.NotNil(t, updatedFiles[0].VirtualSmbLink)
}

func TestUpdateMediaFiles_EmptyFiles(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	updatedFiles, err := ma.updateMediaFiles(1, nil, "/path", "root")
	require.NoError(t, err)
	assert.Empty(t, updatedFiles)
}

func TestUpdateMediaFiles_AllDirectories(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	files := []models.FileInfo{
		{Name: "Dir1", IsDirectory: true},
		{Name: "Dir2", IsDirectory: true},
	}

	updatedFiles, err := ma.updateMediaFiles(1, files, "/path", "root")
	require.NoError(t, err)
	assert.Empty(t, updatedFiles)
}

// ============================================================================
// AnalyzeDirectorySync tests (integration, database-dependent)
// ============================================================================

func TestAnalyzeDirectorySync_NoFiles(t *testing.T) {
	db := setupTestDB(t)
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	ma := NewMediaAnalyzer(db, det, nil, logger)

	ctx := context.Background()
	result, err := ma.AnalyzeDirectorySync(ctx, "/empty/", "missing-root")
	require.NoError(t, err)
	// No files -> detector returns nil -> empty result
	assert.NotNil(t, result)
	assert.Nil(t, result.DirectoryAnalysis)
}

func TestAnalyzeDirectorySync_WithDetectionResult(t *testing.T) {
	db := setupTestDB(t)
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	pm := providers.NewProviderManager(logger)
	ma := NewMediaAnalyzer(db, det, pm, logger)

	// Seed files
	seedStorageRootAndFiles(t, db, "nas-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/movies/Inception (2010)/Inception.2160p.BluRay.x265.mkv", "Inception.2160p.BluRay.x265.mkv", strPtr(".mkv"), nil, 15_000_000_000, false},
		{"/movies/Inception (2010)/Inception.srt", "Inception.srt", strPtr(".srt"), nil, 50000, false},
	})

	// Load detection rules so the engine can detect something
	rules := []mediamodels.DetectionRule{
		{
			ID:               1,
			MediaTypeID:      1,
			RuleName:         "video_files",
			RuleType:         "filename_pattern",
			Pattern:          `["*.mkv","*.mp4","*.avi"]`,
			ConfidenceWeight: 1.0,
			Enabled:          true,
			Priority:         10,
		},
	}
	mediaTypes := []mediamodels.MediaType{
		{ID: 1, Name: "movie"},
	}
	det.LoadRules(rules, mediaTypes)

	ctx := context.Background()
	result, err := ma.AnalyzeDirectorySync(ctx, "/movies/Inception (2010)/", "nas-root")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.DirectoryAnalysis)
	require.NotNil(t, result.MediaItem)

	assert.Equal(t, "/movies/Inception (2010)/", result.DirectoryAnalysis.DirectoryPath)
	assert.NotEmpty(t, result.MediaItem.Title)
	assert.Equal(t, "active", result.MediaItem.Status)
}

func TestAnalyzeDirectorySync_DetectionReturnsNil(t *testing.T) {
	db := setupTestDB(t)
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	// Don't load any rules so detection returns nil
	ma := NewMediaAnalyzer(db, det, nil, logger)

	seedStorageRootAndFiles(t, db, "empty-det-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/data/random.txt", "random.txt", strPtr(".txt"), nil, 100, false},
	})

	ctx := context.Background()
	result, err := ma.AnalyzeDirectorySync(ctx, "/data/", "empty-det-root")
	require.NoError(t, err)
	require.NotNil(t, result)
	// No detection result -> empty AnalysisResult
	assert.Nil(t, result.DirectoryAnalysis)
	assert.Nil(t, result.MediaItem)
}

// ============================================================================
// worker tests
// ============================================================================

func TestWorker_ProcessesRequestWithCallback(t *testing.T) {
	db := setupTestDB(t)
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	pm := providers.NewProviderManager(logger)

	// Load rules so detection succeeds and result.MediaItem is non-nil
	rules := []mediamodels.DetectionRule{
		{
			ID: 1, MediaTypeID: 1, RuleName: "video",
			RuleType: "filename_pattern", Pattern: `["*.mkv","*.mp4"]`,
			ConfidenceWeight: 1.0, Enabled: true, Priority: 10,
		},
	}
	mediaTypes := []mediamodels.MediaType{{ID: 1, Name: "movie"}}
	det.LoadRules(rules, mediaTypes)

	ma := NewMediaAnalyzer(db, det, pm, logger)

	// Seed files so detection has something to work with
	seedStorageRootAndFiles(t, db, "worker-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/worker/test/Movie.1080p.mkv", "Movie.1080p.mkv", strPtr(".mkv"), nil, 4_000_000_000, false},
	})

	var callbackResult *AnalysisResult
	var callbackErr error
	var wg sync.WaitGroup
	wg.Add(1)

	request := AnalysisRequest{
		DirectoryPath: "/worker/test/",
		SmbRoot:       "worker-root",
		Priority:      5,
		Timestamp:     time.Now(),
		Callback: func(result *AnalysisResult, err error) {
			callbackResult = result
			callbackErr = err
			wg.Done()
		},
	}

	ma.workers = 1
	ma.Start()

	ma.analysisQueue <- request

	wg.Wait()

	assert.NoError(t, callbackErr)
	assert.NotNil(t, callbackResult)
	assert.NotNil(t, callbackResult.MediaItem)

	ma.Stop()
}

func TestWorker_ProcessesRequestWithoutCallback(t *testing.T) {
	db := setupTestDB(t)
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	pm := providers.NewProviderManager(logger)

	// Load rules so detection succeeds
	rules := []mediamodels.DetectionRule{
		{
			ID: 1, MediaTypeID: 1, RuleName: "video",
			RuleType: "filename_pattern", Pattern: `["*.mkv"]`,
			ConfidenceWeight: 1.0, Enabled: true, Priority: 10,
		},
	}
	mediaTypes := []mediamodels.MediaType{{ID: 1, Name: "movie"}}
	det.LoadRules(rules, mediaTypes)

	ma := NewMediaAnalyzer(db, det, pm, logger)

	seedStorageRootAndFiles(t, db, "nocb-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/nocb/test/Movie.mkv", "Movie.mkv", strPtr(".mkv"), nil, 2_000_000_000, false},
	})

	ma.mu.Lock()
	ma.pendingAnalysis["/nocb/test/"] = &AnalysisRequest{DirectoryPath: "/nocb/test/"}
	ma.mu.Unlock()

	request := AnalysisRequest{
		DirectoryPath: "/nocb/test/",
		SmbRoot:       "nocb-root",
		Priority:      1,
		Timestamp:     time.Now(),
		Callback:      nil,
	}

	ma.workers = 1
	ma.Start()

	ma.analysisQueue <- request

	// Give the worker time to process
	time.Sleep(200 * time.Millisecond)

	// Verify it was removed from pending
	ma.mu.RLock()
	_, exists := ma.pendingAnalysis["/nocb/test/"]
	ma.mu.RUnlock()
	assert.False(t, exists)

	ma.Stop()
}

func TestWorker_HandlesErrorInAnalysis(t *testing.T) {
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)

	// Use a closed database to trigger a real error (not a panic)
	rawDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	closedDB := database.WrapDB(rawDB, database.DialectSQLite)
	rawDB.Close()

	ma := &MediaAnalyzer{
		db:              closedDB,
		detector:        det,
		logger:          logger,
		analysisQueue:   make(chan AnalysisRequest, 10),
		workers:         1,
		stopCh:          make(chan struct{}),
		pendingAnalysis: make(map[string]*AnalysisRequest),
	}

	var callbackErr error
	var wg sync.WaitGroup
	wg.Add(1)

	request := AnalysisRequest{
		DirectoryPath: "/test/path",
		SmbRoot:       "root",
		Priority:      1,
		Timestamp:     time.Now(),
		Callback: func(result *AnalysisResult, err error) {
			callbackErr = err
			wg.Done()
		},
	}

	ma.wg.Add(1)
	go ma.worker(0)

	ma.analysisQueue <- request
	wg.Wait()

	assert.Error(t, callbackErr)

	close(ma.stopCh)
	ma.wg.Wait()
}

// ============================================================================
// AnalysisRequest / AnalysisResult struct tests
// ============================================================================

func TestAnalysisRequest_FieldsSet(t *testing.T) {
	now := time.Now()
	req := AnalysisRequest{
		DirectoryPath: "/test/path",
		SmbRoot:       "root",
		Priority:      5,
		Timestamp:     now,
	}

	assert.Equal(t, "/test/path", req.DirectoryPath)
	assert.Equal(t, "root", req.SmbRoot)
	assert.Equal(t, 5, req.Priority)
	assert.Equal(t, now, req.Timestamp)
	assert.Nil(t, req.Callback)
}

func TestAnalysisResult_EmptyFields(t *testing.T) {
	result := &AnalysisResult{}

	assert.Nil(t, result.DirectoryAnalysis)
	assert.Nil(t, result.MediaItem)
	assert.Nil(t, result.ExternalMetadata)
	assert.Nil(t, result.QualityAnalysis)
	assert.Nil(t, result.UpdatedFiles)
}

func TestQualityAnalysis_FieldsSet(t *testing.T) {
	qa := &QualityAnalysis{
		BestQuality:        &mediamodels.QualityInfo{QualityScore: 100},
		AvailableQualities: []string{"4K/UHD", "1080p"},
		TotalFiles:         5,
		TotalSize:          20_000_000_000,
		DuplicateCount:     1,
		MissingQualities:   []string{"720p"},
	}

	assert.Equal(t, 100, qa.BestQuality.QualityScore)
	assert.Len(t, qa.AvailableQualities, 2)
	assert.Equal(t, 5, qa.TotalFiles)
	assert.Equal(t, int64(20_000_000_000), qa.TotalSize)
	assert.Equal(t, 1, qa.DuplicateCount)
	assert.Len(t, qa.MissingQualities, 1)
}

// ============================================================================
// Edge case: createDirectoryAnalysis with empty file types
// ============================================================================

func TestCreateDirectoryAnalysis_EmptyFileTypes(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	detection := &detector.DetectionResult{
		Confidence: 0.7,
		Method:     "hybrid",
		AnalysisData: &mediamodels.AnalysisData{
			FileTypes:        map[string]int{},
			SizeDistribution: map[string]int64{},
		},
	}

	result, err := ma.createDirectoryAnalysis("/test/empty", "root", detection)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.FilesCount)
	assert.Equal(t, int64(0), result.TotalSize)
}

// ============================================================================
// End-to-end: AnalyzeDirectorySync with full pipeline
// ============================================================================

func TestAnalyzeDirectorySync_FullPipeline(t *testing.T) {
	db := setupTestDB(t)
	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	pm := providers.NewProviderManager(logger)

	// Load rules for movie detection
	rules := []mediamodels.DetectionRule{
		{
			ID:               1,
			MediaTypeID:      1,
			RuleName:         "video_files",
			RuleType:         "filename_pattern",
			Pattern:          `["*.mkv","*.mp4","*.avi"]`,
			ConfidenceWeight: 1.0,
			Enabled:          true,
			Priority:         10,
		},
	}
	mediaTypes := []mediamodels.MediaType{
		{ID: 1, Name: "movie"},
	}
	det.LoadRules(rules, mediaTypes)

	ma := NewMediaAnalyzer(db, det, pm, logger)

	// Seed a realistic directory
	seedStorageRootAndFiles(t, db, "movie-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/library/Blade Runner 2049 (2017)/Blade.Runner.2049.2017.2160p.BluRay.x265.HDR.DTS.mkv", "Blade.Runner.2049.2017.2160p.BluRay.x265.HDR.DTS.mkv", strPtr(".mkv"), nil, 40_000_000_000, false},
		{"/library/Blade Runner 2049 (2017)/Blade.Runner.2049.2017.1080p.mkv", "Blade.Runner.2049.2017.1080p.mkv", strPtr(".mkv"), nil, 8_000_000_000, false},
		{"/library/Blade Runner 2049 (2017)/subtitle.srt", "subtitle.srt", strPtr(".srt"), nil, 80000, false},
		{"/library/Blade Runner 2049 (2017)/Featurettes", "Featurettes", nil, nil, 0, true},
	})

	ctx := context.Background()
	result, err := ma.AnalyzeDirectorySync(ctx, "/library/Blade Runner 2049 (2017)/", "movie-root")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Directory analysis should be created
	require.NotNil(t, result.DirectoryAnalysis)
	assert.Greater(t, result.DirectoryAnalysis.ConfidenceScore, 0.0)

	// Media item should be created
	require.NotNil(t, result.MediaItem)
	assert.NotEmpty(t, result.MediaItem.Title)
	assert.Equal(t, "active", result.MediaItem.Status)

	// Quality analysis may fail gracefully (media type might not be set on item)
	// but should not cause an error

	// Updated files should include non-directory files
	assert.NotNil(t, result.UpdatedFiles)
}

// ============================================================================
// Concurrency: multiple queue operations
// ============================================================================

func TestAnalyzeDirectory_ConcurrentQueueing(t *testing.T) {
	logger := testLogger(t)
	ma := NewMediaAnalyzer(nil, nil, nil, logger)

	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := "/path/" + string(rune('A'+idx%26))
			_ = ma.AnalyzeDirectory(ctx, path, "root", idx)
		}(i)
	}

	wg.Wait()

	// Should not panic or deadlock
	ma.mu.RLock()
	assert.NotEmpty(t, ma.pendingAnalysis)
	ma.mu.RUnlock()
}

// ============================================================================
// filterMediaFiles: additional media types
// ============================================================================

func TestFilterMediaFiles_SoftwareFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	exe := ".exe"
	iso := ".iso"
	txt := ".txt"

	files := []models.FileInfo{
		{Name: "setup.exe", Extension: &exe, Size: 50_000_000},
		{Name: "disk.iso", Extension: &iso, Size: 4_700_000_000},
		{Name: "readme.txt", Extension: &txt, Size: 1000},
	}

	filtered := ma.filterMediaFiles(files, "software")
	assert.Len(t, filtered, 2)
}

func TestFilterMediaFiles_GameFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	iso := ".iso"
	exe := ".exe"
	rom := ".rom"

	files := []models.FileInfo{
		{Name: "game.iso", Extension: &iso, Size: 4_700_000_000},
		{Name: "game.exe", Extension: &exe, Size: 50_000_000},
		{Name: "rom.rom", Extension: &rom, Size: 2_000_000},
	}

	filtered := ma.filterMediaFiles(files, "game")
	assert.Len(t, filtered, 3)
}

func TestFilterMediaFiles_ComicFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	cbr := ".cbr"
	cbz := ".cbz"
	pdf := ".pdf"

	files := []models.FileInfo{
		{Name: "issue1.cbr", Extension: &cbr, Size: 20_000_000},
		{Name: "issue2.cbz", Extension: &cbz, Size: 25_000_000},
		{Name: "vol1.pdf", Extension: &pdf, Size: 50_000_000},
	}

	filtered := ma.filterMediaFiles(files, "comic")
	assert.Len(t, filtered, 3)
}

func TestFilterMediaFiles_AudiobookFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	mp3 := ".mp3"
	m4b := ".m4b"
	m4a := ".m4a"

	files := []models.FileInfo{
		{Name: "ch01.mp3", Extension: &mp3, Size: 30_000_000},
		{Name: "book.m4b", Extension: &m4b, Size: 200_000_000},
		{Name: "ch02.m4a", Extension: &m4a, Size: 25_000_000},
	}

	filtered := ma.filterMediaFiles(files, "audiobook")
	assert.Len(t, filtered, 3)
}

func TestFilterMediaFiles_PodcastFiles(t *testing.T) {
	ma := &MediaAnalyzer{logger: testLogger(t)}

	mp3 := ".mp3"
	ogg := ".ogg"

	files := []models.FileInfo{
		{Name: "ep01.mp3", Extension: &mp3, Size: 30_000_000},
		{Name: "ep02.ogg", Extension: &ogg, Size: 25_000_000},
	}

	filtered := ma.filterMediaFiles(files, "podcast")
	assert.Len(t, filtered, 2)
}

// ============================================================================
// createOrUpdateMediaItem: edge case with nil year
// ============================================================================

func TestCreateOrUpdateMediaItem_NilYear(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	detection := &detector.DetectionResult{
		MediaTypeID:    1,
		MediaType:      &mediamodels.MediaType{ID: 1, Name: "movie"},
		SuggestedTitle: "Unknown Movie",
		SuggestedYear:  nil,
	}

	_, err := ma.createDirectoryAnalysis("/movies/Unknown", "root", &detector.DetectionResult{
		Confidence:   0.5,
		Method:       "test",
		AnalysisData: nil,
	})
	require.NoError(t, err)

	dirAnalysis := &mediamodels.DirectoryAnalysis{
		DirectoryPath: "/movies/Unknown",
	}

	ctx := context.Background()
	item, err := ma.createOrUpdateMediaItem(ctx, detection, dirAnalysis)
	require.NoError(t, err)
	require.NotNil(t, item)

	assert.Equal(t, "Unknown Movie", item.Title)
	assert.Nil(t, item.Year)
}

// ============================================================================
// updateMediaFiles: quality info on each file
// ============================================================================

func TestUpdateMediaFiles_QualityInfoPersisted(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	ctx := context.Background()
	mediaItemID, err := db.InsertReturningID(ctx,
		`INSERT INTO media_items (media_type_id, title, status, first_detected, last_updated)
		 VALUES (?, ?, 'active', ?, ?)`,
		1, "QualityTest", time.Now(), time.Now(),
	)
	require.NoError(t, err)

	mkv := ".mkv"
	files := []models.FileInfo{
		{ID: 200, Name: "Movie.2160p.BluRay.HDR.x265.DTS.mkv", Path: "/test/Movie.mkv", Extension: &mkv, Size: 15_000_000_000},
	}

	updatedFiles, err := ma.updateMediaFiles(mediaItemID, files, "/test", "root1")
	require.NoError(t, err)
	require.Len(t, updatedFiles, 1)

	// Verify quality info was extracted
	assert.NotNil(t, updatedFiles[0].QualityInfo)
	assert.NotNil(t, updatedFiles[0].QualityInfo.Resolution)
	assert.Equal(t, 3840, updatedFiles[0].QualityInfo.Resolution.Width)
	assert.True(t, updatedFiles[0].QualityInfo.HDR)
}

// ============================================================================
// getDirectoryFiles: scans rows properly
// ============================================================================

func TestGetDirectoryFiles_MultipleFields(t *testing.T) {
	db := setupTestDB(t)
	ma := &MediaAnalyzer{logger: testLogger(t), db: db}

	ext := ".mp4"
	mime := "video/mp4"
	seedStorageRootAndFiles(t, db, "detail-root", []struct {
		path  string
		name  string
		ext   *string
		mime  *string
		size  int64
		isDir bool
	}{
		{"/vids/clip.mp4", "clip.mp4", &ext, &mime, 500_000, false},
	})

	files, err := ma.getDirectoryFiles("/vids/", "detail-root")
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	assert.Equal(t, "clip.mp4", f.Name)
	assert.Equal(t, "/vids/clip.mp4", f.Path)
	assert.Equal(t, int64(500_000), f.Size)
	assert.False(t, f.IsDirectory)
	require.NotNil(t, f.Extension)
	assert.Equal(t, ".mp4", *f.Extension)
	require.NotNil(t, f.MimeType)
	assert.Equal(t, "video/mp4", *f.MimeType)
}

// ============================================================================
// AnalyzeDirectorySync: error on db.Query failure
// ============================================================================

func TestAnalyzeDirectorySync_DBQueryError(t *testing.T) {
	// Use a closed database to trigger error
	rawDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db := database.WrapDB(rawDB, database.DialectSQLite)
	rawDB.Close() // Close to force errors

	logger := testLogger(t)
	det := detector.NewDetectionEngine(logger)
	ma := NewMediaAnalyzer(db, det, nil, logger)

	ctx := context.Background()
	result, err := ma.AnalyzeDirectorySync(ctx, "/test/", "root")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get directory files")
}
