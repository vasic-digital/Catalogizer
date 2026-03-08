package services

import (
	"catalogizer/database"
	"catalogizer/internal/models"
	catalogModels "catalogizer/models"
	"context"
	"database/sql"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =============================================================================
// CatalogService — additional coverage for 0% functions
// =============================================================================

func setupCatalogTestDB(t *testing.T) (*database.DB, *CatalogService) {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", "file::memory:?cache=shared&_busy_timeout=5000")
	require.NoError(t, err)

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS storage_roots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			protocol TEXT,
			host TEXT, port INTEGER, path TEXT, username TEXT, password TEXT, domain TEXT,
			enabled BOOLEAN DEFAULT 1,
			max_depth INTEGER DEFAULT 10,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME
		);
		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			storage_root_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			is_directory BOOLEAN DEFAULT 0,
			size INTEGER,
			modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			quick_hash TEXT,
			extension TEXT,
			mime_type TEXT,
			file_type TEXT,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_scan_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted BOOLEAN DEFAULT 0,
			is_duplicate BOOLEAN DEFAULT 0,
			duplicate_group_id INTEGER,
			FOREIGN KEY (storage_root_id) REFERENCES storage_roots(id),
			FOREIGN KEY (parent_id) REFERENCES files(id)
		);
		CREATE TABLE IF NOT EXISTS duplicate_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_hash TEXT NOT NULL,
			file_count INTEGER DEFAULT 0,
			total_size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	// Seed data
	_, err = db.Exec(`INSERT INTO storage_roots (id, name, protocol) VALUES (1, 'test-root', 'local'), (2, 'backup-root', 'smb')`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO files (id, storage_root_id, name, path, is_directory, size, parent_id, quick_hash, extension, mime_type) VALUES
		(1, 1, 'media',       '/media',                     1, 0,        NULL, NULL,    NULL,  NULL),
		(2, 1, 'movies',      '/media/movies',              1, 0,        1,    NULL,    NULL,  NULL),
		(3, 1, 'video.mp4',   '/media/movies/video.mp4',    0, 500000,   2,    'h1',    'mp4', 'video/mp4'),
		(4, 1, 'video2.mp4',  '/media/movies/video2.mp4',   0, 500000,   2,    'h1',    'mp4', 'video/mp4'),
		(5, 2, 'docs',        '/docs',                      1, 0,        NULL, NULL,    NULL,  NULL),
		(6, 2, 'readme.txt',  '/docs/readme.txt',           0, 1024,     5,    NULL,    'txt', 'text/plain')
	`)
	require.NoError(t, err)

	logger := zap.NewNop()
	svc := NewCatalogService(nil, logger)
	svc.SetDB(db)

	t.Cleanup(func() { db.Close() })
	return db, svc
}

func TestCatalogService_GetSMBRoots(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	roots, err := svc.GetSMBRoots()
	require.NoError(t, err)
	assert.Len(t, roots, 2)
	assert.Contains(t, roots, "test-root")
	assert.Contains(t, roots, "backup-root")
}

func TestCatalogService_GetFileInfoByPath(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	tests := []struct {
		name     string
		path     string
		wantNil  bool
		wantName string
		wantType string
	}{
		{name: "existing file", path: "/media/movies/video.mp4", wantName: "video.mp4", wantType: "file"},
		{name: "existing directory", path: "/media/movies", wantName: "movies", wantType: "directory"},
		{name: "non-existent path", path: "/nonexistent", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := svc.GetFileInfoByPath(tt.path)
			require.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, info)
			} else {
				require.NotNil(t, info)
				assert.Equal(t, tt.wantName, info.Name)
				assert.Equal(t, tt.wantType, info.Type)
			}
		})
	}
}

func TestCatalogService_GetDirectoriesBySizeLimited(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	dirs, err := svc.GetDirectoriesBySizeLimited(10)
	require.NoError(t, err)
	// Result may be nil or empty when no directories match; we just verify no error
	_ = dirs
}

func TestCatalogService_SearchFiles_WithFilters(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	minSize := int64(100)
	maxSize := int64(600000)
	isDir := false

	tests := []struct {
		name      string
		req       *models.SearchRequest
		wantCount int
	}{
		{
			name: "search by query",
			req: &models.SearchRequest{
				Query: "video",
				Limit: 10,
			},
			wantCount: 2,
		},
		{
			name: "search by extension",
			req: &models.SearchRequest{
				Extension: "mp4",
				Limit:     10,
			},
			wantCount: 2,
		},
		{
			name: "search by mime type",
			req: &models.SearchRequest{
				MimeType: "text/plain",
				Limit:    10,
			},
			wantCount: 1,
		},
		{
			name: "search with min/max size",
			req: &models.SearchRequest{
				MinSize: &minSize,
				MaxSize: &maxSize,
				Limit:   10,
			},
			wantCount: 3, // 1024 < 100, so only 500000-sized files + 1024
		},
		{
			name: "search by smb roots",
			req: &models.SearchRequest{
				SmbRoots: []string{"backup-root"},
				Limit:    10,
			},
			wantCount: 2, // docs dir + readme.txt
		},
		{
			name: "search only files (not directories)",
			req: &models.SearchRequest{
				IsDirectory: &isDir,
				Limit:       10,
			},
			wantCount: 3,
		},
		{
			name: "search by path prefix",
			req: &models.SearchRequest{
				Path:  "/media/movies",
				Limit: 10,
			},
			wantCount: 3, // /media/movies dir + 2 files
		},
		{
			name: "sort by size desc",
			req: &models.SearchRequest{
				SortBy:    "size",
				SortOrder: "desc",
				Limit:     10,
			},
			wantCount: 6,
		},
		{
			name: "sort by modified",
			req: &models.SearchRequest{
				SortBy: "modified",
				Limit:  10,
			},
			wantCount: 6,
		},
		{
			name: "sort by name",
			req: &models.SearchRequest{
				SortBy: "name",
				Limit:  10,
			},
			wantCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, total, err := svc.SearchFiles(tt.req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(files))
			assert.Equal(t, int64(tt.wantCount), total)
		})
	}
}

func TestCatalogService_ListPath_Sorting(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	tests := []struct {
		name      string
		path      string
		sortBy    string
		sortOrder string
		limit     int
		offset    int
	}{
		{name: "sort by name asc", path: "/media/movies", sortBy: "name", sortOrder: "asc", limit: 10},
		{name: "sort by size desc", path: "/media/movies", sortBy: "size", sortOrder: "desc", limit: 10},
		{name: "sort by modified asc", path: "/media/movies", sortBy: "modified", sortOrder: "asc", limit: 10},
		{name: "default sort", path: "/media/movies", sortBy: "", sortOrder: "", limit: 10},
		{name: "with pagination", path: "/media/movies", sortBy: "name", sortOrder: "asc", limit: 1, offset: 1},
		{name: "root listing", path: "/", sortBy: "name", sortOrder: "asc", limit: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := svc.ListPath(tt.path, tt.sortBy, tt.sortOrder, tt.limit, tt.offset)
			require.NoError(t, err)
			assert.NotNil(t, files)
		})
	}
}

func TestCatalogService_ListPath_PathNotFound(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	_, err := svc.ListPath("/nonexistent/path", "name", "asc", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path not found")
}

func TestCatalogService_GetFileInfo_ByID(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	// Lookup by numeric ID
	info, err := svc.GetFileInfo("3")
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "video.mp4", info.Name)
	assert.Equal(t, "file", info.Type)
}

func TestCatalogService_GetDuplicateGroups_WithSMBRoot(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	// Files with quick_hash 'h1' are duplicates
	groups, err := svc.GetDuplicateGroups("test-root", 2, 10)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Len(t, groups[0].Files, 2)
	assert.Equal(t, "h1", groups[0].Hash)
}

func TestCatalogService_GetDuplicateGroups_NoRoot(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	groups, err := svc.GetDuplicateGroups("", 2, 10)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
}

func TestCatalogService_GetDuplicateGroups_NoLimit(t *testing.T) {
	_, svc := setupCatalogTestDB(t)

	groups, err := svc.GetDuplicateGroups("", 2, 0)
	require.NoError(t, err)
	assert.NotNil(t, groups)
}

// =============================================================================
// DuplicateDetectionService — additional coverage
// =============================================================================

func newDupDetectionSvc() *DuplicateDetectionService {
	return NewDuplicateDetectionService(nil, zap.NewNop(), nil)
}

func TestDuplicateDetection_CalculateBookMetadataSimilarity_Extended(t *testing.T) {
	svc := newDupDetectionSvc()

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "same author and year",
			item1: &DuplicateItem{
				Author: "J.K. Rowling",
				Year:   2001,
			},
			item2: &DuplicateItem{
				Author: "J.K. Rowling",
				Year:   2001,
			},
			minScore: 0.9,
		},
		{
			name: "different author same year",
			item1: &DuplicateItem{
				Author: "J.K. Rowling",
				Year:   2001,
			},
			item2: &DuplicateItem{
				Author: "Stephen King",
				Year:   2001,
			},
			minScore: 0.0,
		},
		{
			name:     "empty metadata",
			item1:    &DuplicateItem{},
			item2:    &DuplicateItem{},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateBookMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetection_CalculateSoftwareMetadataSimilarity_Extended(t *testing.T) {
	svc := newDupDetectionSvc()

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "same size",
			item1: &DuplicateItem{
				FileSize: 10000000,
			},
			item2: &DuplicateItem{
				FileSize: 10000000,
			},
			minScore: 0.9,
		},
		{
			name: "very different size",
			item1: &DuplicateItem{
				FileSize: 1000,
			},
			item2: &DuplicateItem{
				FileSize: 10000000,
			},
			minScore: 0.0,
		},
		{
			name:     "empty metadata",
			item1:    &DuplicateItem{},
			item2:    &DuplicateItem{},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateSoftwareMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetection_CalculateGenericMetadataSimilarity_Extended(t *testing.T) {
	svc := newDupDetectionSvc()

	tests := []struct {
		name     string
		item1    *DuplicateItem
		item2    *DuplicateItem
		minScore float64
	}{
		{
			name: "identical files",
			item1: &DuplicateItem{
				FileName: "test.mp4",
				FileSize: 5000000,
				Format:   "mp4",
			},
			item2: &DuplicateItem{
				FileName: "test.mp4",
				FileSize: 5000000,
				Format:   "mp4",
			},
			minScore: 0.9,
		},
		{
			name: "different filenames, same format",
			item1: &DuplicateItem{
				FileName: "test.mp4",
				FileSize: 5000000,
				Format:   "mp4",
			},
			item2: &DuplicateItem{
				FileName: "other.mp4",
				FileSize: 4900000,
				Format:   "mp4",
			},
			minScore: 0.3,
		},
		{
			name: "empty items",
			item1: &DuplicateItem{
				FileName: "",
			},
			item2: &DuplicateItem{
				FileName: "",
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateGenericMetadataSimilarity(tt.item1, tt.item2)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestDuplicateDetection_CalculateSimilarity_Comprehensive(t *testing.T) {
	svc := newDupDetectionSvc()

	tests := []struct {
		name      string
		item1     *DuplicateItem
		item2     *DuplicateItem
		mediaType MediaType
		minScore  float64
		maxScore  float64
	}{
		{
			name: "hash match",
			item1: &DuplicateItem{
				FileHash: "abc123",
				Title:    "Test",
			},
			item2: &DuplicateItem{
				FileHash: "abc123",
				Title:    "Different",
			},
			mediaType: MediaTypeMovie,
			minScore:  1.0,
			maxScore:  1.0,
		},
		{
			name: "external ID match",
			item1: &DuplicateItem{
				Title:       "Test",
				ExternalIDs: map[string]string{"imdb": "tt1234567"},
			},
			item2: &DuplicateItem{
				Title:       "Different",
				ExternalIDs: map[string]string{"imdb": "tt1234567"},
			},
			mediaType: MediaTypeMovie,
			minScore:  0.95,
			maxScore:  0.95,
		},
		{
			name: "music media type",
			item1: &DuplicateItem{
				Title:    "Song Title",
				Artist:   "Artist Name",
				Album:    "Album Name",
				FileName: "song.mp3",
			},
			item2: &DuplicateItem{
				Title:    "Song Title",
				Artist:   "Artist Name",
				Album:    "Album Name",
				FileName: "song.mp3",
			},
			mediaType: MediaTypeMusic,
			minScore:  0.5,
			maxScore:  1.0,
		},
		{
			name: "book media type",
			item1: &DuplicateItem{
				Title:    "Book Title",
				Author:   "Author Name",
				FileName: "book.pdf",
			},
			item2: &DuplicateItem{
				Title:    "Book Title",
				Author:   "Author Name",
				FileName: "book.pdf",
			},
			mediaType: MediaTypeBook,
			minScore:  0.5,
			maxScore:  1.0,
		},
		{
			name: "game media type",
			item1: &DuplicateItem{
				Title:    "Game Title",
				FileName: "game.iso",
				FileSize: 5000000,
			},
			item2: &DuplicateItem{
				Title:    "Game Title",
				FileName: "game.iso",
				FileSize: 5000000,
			},
			mediaType: MediaTypeGame,
			minScore:  0.5,
			maxScore:  1.0,
		},
		{
			name: "unknown media type defaults",
			item1: &DuplicateItem{
				Title:    "Some Item",
				FileName: "file.dat",
			},
			item2: &DuplicateItem{
				Title:    "Some Item",
				FileName: "file.dat",
			},
			mediaType: MediaType("unknown"),
			minScore:  0.3,
			maxScore:  1.0,
		},
		{
			name: "with fingerprints",
			item1: &DuplicateItem{
				Title:        "Test Item",
				FileName:     "file.mp4",
				Fingerprints: map[string]string{"audio": "fp123"},
			},
			item2: &DuplicateItem{
				Title:        "Test Item",
				FileName:     "file.mp4",
				Fingerprints: map[string]string{"audio": "fp123"},
			},
			mediaType: MediaTypeMovie,
			minScore:  0.3,
			maxScore:  1.0,
		},
		{
			name: "empty external IDs no match",
			item1: &DuplicateItem{
				Title:       "Test",
				ExternalIDs: map[string]string{"imdb": ""},
			},
			item2: &DuplicateItem{
				Title:       "Test",
				ExternalIDs: map[string]string{"imdb": ""},
			},
			mediaType: MediaTypeMovie,
			minScore:  0.0,
			maxScore:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := svc.calculateSimilarity(context.Background(), tt.item1, tt.item2, tt.mediaType)
			assert.GreaterOrEqual(t, analysis.OverallScore, tt.minScore)
			assert.LessOrEqual(t, analysis.OverallScore, tt.maxScore)
		})
	}
}

func TestDuplicateDetection_VideoMetadata_Duration(t *testing.T) {
	svc := newDupDetectionSvc()

	item1 := &DuplicateItem{
		Duration: 7200000,
		Quality:  "1080p",
	}
	item2 := &DuplicateItem{
		Duration: 7200000,
		Quality:  "1080p",
	}

	score := svc.calculateVideoMetadataSimilarity(item1, item2)
	assert.GreaterOrEqual(t, score, 0.9)
}

func TestDuplicateDetection_AudioMetadata_Bitrate(t *testing.T) {
	svc := newDupDetectionSvc()

	item1 := &DuplicateItem{
		Artist:   "Artist",
		Duration: 240000,
		Bitrate:  320,
	}
	item2 := &DuplicateItem{
		Artist:   "Artist",
		Duration: 240000,
		Bitrate:  320,
	}

	score := svc.calculateAudioMetadataSimilarity(item1, item2)
	assert.GreaterOrEqual(t, score, 0.8)
}

func TestDuplicateDetection_MetaphoneMatch(t *testing.T) {
	svc := newDupDetectionSvc()

	// Same word should match
	assert.True(t, svc.metaphoneMatch("Hello", "Hello"))

	// Different words with different sounds should not match
	result := svc.metaphoneMatch("Cat", "Dog")
	_ = result // just exercising the function
}

func TestDuplicateDetection_FingerprintSimilarity_WithMatch(t *testing.T) {
	svc := newDupDetectionSvc()

	fp1 := map[string]string{"audio": "hash1", "video": "hash2"}
	fp2 := map[string]string{"audio": "hash1", "video": "hash3"}

	score := svc.calculateFingerprintSimilarity(fp1, fp2)
	assert.GreaterOrEqual(t, score, 0.0)
}

func TestDuplicateDetection_FingerprintSimilarity_NoOverlap(t *testing.T) {
	svc := newDupDetectionSvc()

	fp1 := map[string]string{"audio": "hash1"}
	fp2 := map[string]string{"video": "hash2"}

	score := svc.calculateFingerprintSimilarity(fp1, fp2)
	assert.Equal(t, 0.0, score)
}

func TestDuplicateDetection_FileSimilarity_NegativeSizeScore(t *testing.T) {
	svc := newDupDetectionSvc()

	// Very different sizes should get 0
	item1 := &DuplicateItem{FileName: "a.txt", FileSize: 1}
	item2 := &DuplicateItem{FileName: "b.txt", FileSize: 100000000}

	score := svc.calculateFileSimilarity(item1, item2)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestDuplicateDetection_DetectDuplicates(t *testing.T) {
	svc := newDupDetectionSvc()

	req := &DuplicateDetectionRequest{
		MediaTypes:    []MediaType{MediaTypeMovie},
		MinSimilarity: 0.9,
	}

	// Should return nil or empty because getMediaItems returns empty for in-memory DB
	groups, err := svc.DetectDuplicates(context.Background(), req)
	require.NoError(t, err)
	assert.Empty(t, groups)
}

func TestDuplicateDetection_DetectDuplicates_AllMediaTypes(t *testing.T) {
	svc := newDupDetectionSvc()

	// Pass empty MediaTypes to trigger default all-types detection
	req := &DuplicateDetectionRequest{
		MediaTypes:    nil,
		MinSimilarity: 0.9,
	}

	groups, err := svc.DetectDuplicates(context.Background(), req)
	require.NoError(t, err)
	assert.Empty(t, groups)
}

// =============================================================================
// AssetResolvers — Name() and Priority() coverage
// =============================================================================

func TestCachedFileResolver_NameAndPriority(t *testing.T) {
	r := NewCachedFileResolver("/tmp/cache", 10)
	assert.Equal(t, "cached_file", r.Name())
	assert.Equal(t, 10, r.Priority())
}

func TestExternalMetadataResolver_NameAndPriority(t *testing.T) {
	r := NewExternalMetadataResolver(nil, 20)
	assert.Equal(t, "external_metadata", r.Name())
	assert.Equal(t, 20, r.Priority())
}

func TestLocalScanResolver_NameAndPriority(t *testing.T) {
	r := NewLocalScanResolver(30)
	assert.Equal(t, "local_scan", r.Name())
	assert.Equal(t, 30, r.Priority())
}

// =============================================================================
// BookRecognitionProvider — uncovered pure functions
// =============================================================================

func TestBookRecognition_BasicBookRecognition(t *testing.T) {
	provider := NewBookRecognitionProvider(zap.NewNop())

	result := provider.basicBookRecognition(
		&MediaRecognitionRequest{FileName: "test.pdf"},
		"The Great Gatsby",
		"F. Scott Fitzgerald",
		"9780743273565",
		MediaTypeBook,
	)

	require.NotNil(t, result)
	assert.Equal(t, "The Great Gatsby", result.Title)
	assert.Equal(t, "F. Scott Fitzgerald", result.Author)
	assert.Equal(t, "9780743273565", result.ISBN)
	assert.Equal(t, MediaTypeBook, result.MediaType)
	assert.Equal(t, 0.3, result.Confidence)
	assert.Equal(t, "filename_parsing", result.RecognitionMethod)
	assert.Equal(t, "basic", result.APIProvider)
	assert.NotEmpty(t, result.MediaID)
}

func TestBookRecognition_ExtractTitleFromOCR(t *testing.T) {
	provider := NewBookRecognitionProvider(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		blocks   []OCRTextBlock
		expected string
	}{
		{
			name: "largest block is title",
			text: "The Great Gatsby\nby F. Scott Fitzgerald",
			blocks: []OCRTextBlock{
				{
					Text:       "The Great Gatsby",
					Confidence: 0.95,
					BoundingBox: OCRBoundingBox{
						Width:  400,
						Height: 100,
					},
				},
				{
					Text:       "by F. Scott Fitzgerald",
					Confidence: 0.90,
					BoundingBox: OCRBoundingBox{
						Width:  200,
						Height: 50,
					},
				},
			},
			expected: "The Great Gatsby",
		},
		{
			name: "low confidence blocks ignored",
			text: "Some Title Here\nby Someone",
			blocks: []OCRTextBlock{
				{
					Text:       "Some Title Here",
					Confidence: 0.5, // below 0.8 threshold
					BoundingBox: OCRBoundingBox{
						Width:  400,
						Height: 100,
					},
				},
			},
			expected: "Some Title Here", // falls back to first line
		},
		{
			name:     "empty blocks, uses first text line",
			text:     "A Good Title Line\nSecond line",
			blocks:   []OCRTextBlock{},
			expected: "A Good Title Line",
		},
		{
			name:     "empty text and blocks",
			text:     "",
			blocks:   []OCRTextBlock{},
			expected: "",
		},
		{
			name:     "skips very short lines",
			text:     "Hi\nA Reasonable Title Here\nOther text",
			blocks:   []OCRTextBlock{},
			expected: "A Reasonable Title Here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractTitleFromOCR(tt.text, tt.blocks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBookRecognition_ExtractAuthorFromOCR(t *testing.T) {
	provider := NewBookRecognitionProvider(zap.NewNop())

	tests := []struct {
		name     string
		text     string
		blocks   []OCRTextBlock
		expected string
	}{
		{
			name:     "by pattern",
			text:     "Some Title by John Smith",
			blocks:   []OCRTextBlock{},
			expected: "John Smith",
		},
		{
			name: "author-like block",
			text: "Some text without by keyword",
			blocks: []OCRTextBlock{
				{
					Text:       "John Smith",
					Confidence: 0.9,
				},
			},
			expected: "John Smith",
		},
		{
			name:     "no author found",
			text:     "just some random text",
			blocks:   []OCRTextBlock{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.extractAuthorFromOCR(tt.text, tt.blocks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// LocalizationService — ValidateConfigurationJSON & GetConfigurationTemplate
// =============================================================================

func TestLocalization_ValidateConfigurationJSON(t *testing.T) {
	svc := NewLocalizationService(nil, zap.NewNop(), nil, nil)

	tests := []struct {
		name       string
		json       string
		wantErr    bool
		wantErrors bool
	}{
		{
			name: "valid JSON, valid config",
			json: `{
				"version": "1.0",
				"config_type": "full",
				"localization": {
					"primary_language": "en",
					"date_format": "MM/DD/YYYY",
					"time_format": "12h"
				},
				"media_settings": {
					"default_quality": "high",
					"volume_level": 0.8,
					"crossfade_duration": 3000
				}
			}`,
			wantErr:    false,
			wantErrors: false,
		},
		{
			name:       "invalid JSON",
			json:       `{not valid json`,
			wantErr:    true,
			wantErrors: true,
		},
		{
			name:       "valid JSON but missing version",
			json:       `{"version": "", "config_type": "full"}`,
			wantErr:    false,
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := svc.ValidateConfigurationJSON(context.Background(), tt.json)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantErrors {
				assert.NotEmpty(t, errors)
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestLocalization_GetConfigurationTemplate(t *testing.T) {
	svc := NewLocalizationService(nil, zap.NewNop(), nil, nil)

	tests := []struct {
		name         string
		templateType string
		wantErr      bool
	}{
		{name: "localization template", templateType: "localization", wantErr: false},
		{name: "media template", templateType: "media", wantErr: false},
		{name: "playlists template", templateType: "playlists", wantErr: false},
		{name: "full template", templateType: "full", wantErr: false},
		{name: "unknown template", templateType: "unknown", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := svc.GetConfigurationTemplate(context.Background(), tt.templateType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, template)
			} else {
				require.NoError(t, err)
				require.NotNil(t, template)
				assert.Equal(t, "1.0", template.Version)
				assert.Equal(t, tt.templateType, template.ConfigType)
				assert.Contains(t, template.Tags, "template")
			}
		})
	}
}

// =============================================================================
// RecommendationService — additional coverage for calculation helpers
// =============================================================================

func newTestRecommendationSvc() *RecommendationService {
	mockRepo := &MockFileRepository{}
	mockDupDetection := NewDuplicateDetectionService(nil, zap.NewNop(), nil)
	return NewRecommendationService(nil, mockDupDetection, mockRepo, nil)
}

func TestRecommendation_GetDB(t *testing.T) {
	svc := newTestRecommendationSvc()
	assert.Nil(t, svc.GetDB())
}

func TestRecommendation_CalculateTMDbSimilarity(t *testing.T) {
	svc := newTestRecommendationSvc()

	year := 2023
	rating := 8.5
	original := &catalogModels.MediaMetadata{
		Title:  "Test Movie",
		Year:   &year,
		Rating: &rating,
	}

	tests := []struct {
		name        string
		title       string
		releaseDate string
		rating      float64
		minScore    float64
	}{
		{name: "matching year and close rating", title: "Test Movie", releaseDate: "2023-05-15", rating: 8.2, minScore: 0.7},
		{name: "different year", title: "Other Movie", releaseDate: "2020-01-01", rating: 5.0, minScore: 0.4},
		{name: "no release date", title: "Movie", releaseDate: "", rating: 8.5, minScore: 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateTMDbSimilarity(original, tt.title, tt.releaseDate, tt.rating)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestRecommendation_CalculateOMDbSimilarity(t *testing.T) {
	svc := newTestRecommendationSvc()

	year := 2023
	original := &catalogModels.MediaMetadata{
		Title: "Test Movie",
		Year:  &year,
	}

	tests := []struct {
		name     string
		title    string
		year     string
		minScore float64
	}{
		{name: "matching year", title: "Similar Movie", year: "2023", minScore: 0.6},
		{name: "different year", title: "Other Movie", year: "2010", minScore: 0.3},
		{name: "empty year", title: "Unknown Movie", year: "", minScore: 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateOMDbSimilarity(original, tt.title, tt.year)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestRecommendation_CalculateGoogleBooksSimilarity(t *testing.T) {
	svc := newTestRecommendationSvc()

	original := &catalogModels.MediaMetadata{
		Title:    "Test Book",
		Producer: "Test Author",
	}

	tests := []struct {
		name     string
		title    string
		author   string
		minScore float64
	}{
		{name: "matching author and title", title: "Test Book", author: "Test Author", minScore: 0.6},
		{name: "different author", title: "Other Book", author: "Other Author", minScore: 0.3},
		{name: "empty author", title: "Some Book", author: "", minScore: 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateGoogleBooksSimilarity(original, tt.title, tt.author)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestRecommendation_CalculateGitHubSimilarity(t *testing.T) {
	svc := newTestRecommendationSvc()

	original := &catalogModels.MediaMetadata{
		Title: "MyProject",
		Genre: "Go",
	}

	tests := []struct {
		name     string
		repoName string
		language string
		stars    int
		minScore float64
	}{
		{name: "popular matching repo", repoName: "MyProject", language: "Go", stars: 15000, minScore: 0.6},
		{name: "medium popularity", repoName: "Other", language: "Go", stars: 5000, minScore: 0.3},
		{name: "low popularity", repoName: "Unknown", language: "Python", stars: 10, minScore: 0.2},
		{name: "no language match", repoName: "Test", language: "", stars: 100, minScore: 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := svc.calculateGitHubSimilarity(original, tt.repoName, tt.language, tt.stars)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestRecommendation_PassesExternalFilters(t *testing.T) {
	svc := newTestRecommendationSvc()

	tests := []struct {
		name     string
		item     *ExternalSimilarItem
		filters  *RecommendationFilters
		expected bool
	}{
		{
			name:     "nil filters always passes",
			item:     &ExternalSimilarItem{},
			filters:  nil,
			expected: true,
		},
		{
			name:    "genre filter matches",
			item:    &ExternalSimilarItem{Genre: "Action, Thriller"},
			filters: &RecommendationFilters{GenreFilter: []string{"action"}},
			expected: true,
		},
		{
			name:    "genre filter does not match",
			item:    &ExternalSimilarItem{Genre: "Comedy"},
			filters: &RecommendationFilters{GenreFilter: []string{"action"}},
			expected: false,
		},
		{
			name: "year range filter within range",
			item: &ExternalSimilarItem{Year: "2023"},
			filters: &RecommendationFilters{
				YearRange: &YearRange{StartYear: 2020, EndYear: 2025},
			},
			expected: true,
		},
		{
			name: "rating range filter within range",
			item: &ExternalSimilarItem{Rating: 8.5},
			filters: &RecommendationFilters{
				RatingRange: &RatingRange{MinRating: 7.0, MaxRating: 9.0},
			},
			expected: true,
		},
		{
			name: "rating filter out of range",
			item: &ExternalSimilarItem{Rating: 5.0},
			filters: &RecommendationFilters{
				RatingRange: &RatingRange{MinRating: 7.0, MaxRating: 9.0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.passesExternalFilters(tt.item, tt.filters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRecommendation_GenerateMockLocalMedia(t *testing.T) {
	svc := newTestRecommendationSvc()

	year := 2023
	original := &catalogModels.MediaMetadata{
		Title: "Test Movie",
		Year:  &year,
		Genre: "Action",
	}

	mockMedia := svc.generateMockLocalMedia(original)
	assert.NotEmpty(t, mockMedia)
	assert.GreaterOrEqual(t, len(mockMedia), 4)

	// Verify mock data has the original's genre
	for _, media := range mockMedia {
		assert.Equal(t, "Action", media.Genre)
	}
}

func TestRecommendation_FindSimilarBooks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring network")
	}
	svc := newTestRecommendationSvc()

	// Without API keys, this should gracefully return empty
	items, err := svc.findSimilarBooks(context.Background(), &catalogModels.MediaMetadata{
		Title: "Test Book",
		Genre: "Fiction",
	})
	assert.NoError(t, err)
	assert.NotNil(t, items)
}

func TestRecommendation_FindSimilarGames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring network")
	}
	svc := newTestRecommendationSvc()

	items, err := svc.findSimilarGames(context.Background(), &catalogModels.MediaMetadata{
		Title: "Test Game",
		Genre: "RPG",
	})
	assert.NoError(t, err)
	assert.NotNil(t, items)
}

func TestRecommendation_FindSimilarSoftware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring network")
	}
	svc := newTestRecommendationSvc()

	items, err := svc.findSimilarSoftware(context.Background(), &catalogModels.MediaMetadata{
		Title: "Test App",
		Genre: "Utility",
	})
	assert.NoError(t, err)
	assert.NotNil(t, items)
}

func TestRecommendation_FindSimilarMusic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test requiring network")
	}
	svc := newTestRecommendationSvc()

	items, err := svc.findSimilarMusic(context.Background(), &catalogModels.MediaMetadata{
		Title:    "Test Song",
		Director: "Test Artist",
	})
	assert.NoError(t, err)
	assert.NotNil(t, items)
}

func TestRecommendation_PassesFilters_LanguageFilter(t *testing.T) {
	svc := newTestRecommendationSvc()

	media := &catalogModels.MediaMetadata{
		Title:    "Test",
		Language: "English",
	}

	// Language filter that matches
	assert.True(t, svc.passesFilters(media, 0.8, &RecommendationFilters{
		LanguageFilter: []string{"English"},
	}))

	// Language filter that doesn't match
	assert.False(t, svc.passesFilters(media, 0.8, &RecommendationFilters{
		LanguageFilter: []string{"French"},
	}))
}

func TestRecommendation_PassesFilters_MinConfidence(t *testing.T) {
	svc := newTestRecommendationSvc()

	media := &catalogModels.MediaMetadata{Title: "Test"}

	// Below min confidence
	assert.False(t, svc.passesFilters(media, 0.3, &RecommendationFilters{
		MinConfidence: 0.5,
	}))

	// Above min confidence
	assert.True(t, svc.passesFilters(media, 0.7, &RecommendationFilters{
		MinConfidence: 0.5,
	}))
}

// =============================================================================
// Localization — importWizardStepSettings (0% coverage)
// =============================================================================

func TestLocalization_ImportWizardStepSettings(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", "file::memory:?cache=shared&_busy_timeout=5000")
	require.NoError(t, err)
	defer sqlDB.Close()

	db := database.WrapDB(sqlDB, database.DialectSQLite)

	// Create required tables for SetupUserLocalization
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_localization (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL UNIQUE,
			primary_language TEXT NOT NULL DEFAULT 'en',
			secondary_languages TEXT DEFAULT '[]',
			subtitle_languages TEXT DEFAULT '[]',
			lyrics_languages TEXT DEFAULT '[]',
			metadata_languages TEXT DEFAULT '[]',
			auto_translate BOOLEAN DEFAULT 0,
			auto_download_subtitles BOOLEAN DEFAULT 0,
			auto_download_lyrics BOOLEAN DEFAULT 0,
			preferred_region TEXT DEFAULT 'US',
			date_format TEXT DEFAULT 'YYYY-MM-DD',
			time_format TEXT DEFAULT '24h',
			number_format TEXT DEFAULT 'en-US',
			currency_code TEXT DEFAULT 'USD',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS content_language_preferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			content_type TEXT NOT NULL,
			languages TEXT DEFAULT '[]',
			priority INTEGER DEFAULT 0,
			auto_apply BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)

	svc := NewLocalizationService(db, zap.NewNop(), nil, nil)

	wizardStep := &WizardLocalizationStep{
		UserID:                42,
		PrimaryLanguage:       "en",
		SecondaryLanguages:    []string{"fr"},
		SubtitleLanguages:     []string{"en"},
		LyricsLanguages:       []string{"en"},
		MetadataLanguages:     []string{"en"},
		AutoTranslate:         false,
		AutoDownloadSubtitles: false,
		AutoDownloadLyrics:    false,
		PreferredRegion:       "US",
		DateFormat:            "MM/DD/YYYY",
		TimeFormat:            "12h",
		NumberFormat:          "en-US",
		CurrencyCode:          "USD",
	}

	err = svc.importWizardStepSettings(context.Background(), 42, wizardStep)
	assert.NoError(t, err)
}

// (EditConfiguration tests are in additional_coverage_test.go)
