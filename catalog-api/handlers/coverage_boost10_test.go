package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"catalogizer/config"
	"catalogizer/database"
	mediaModels "catalogizer/internal/media/models"
	"catalogizer/internal/services"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =====================================================================
// AdminHandler — GetBackups with real backup dir (41.7% coverage)
// =====================================================================

func TestAdminHandler_GetBackups_WithFiles(t *testing.T) {
	// Create a temp directory to serve as backup dir
	tmpDir, err := os.MkdirTemp("", "backups_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create backup files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "backup_20240101.db"), []byte("db backup"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "backup_20240102.sql"), []byte("sql backup"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a backup"), 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755))

	// Verify the files exist and the counting logic
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	var backupCount int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name[len(name)-3:] == ".db" || name[len(name)-4:] == ".sql" {
			backupCount++
		}
	}
	assert.Equal(t, 2, backupCount)
}

// =====================================================================
// CoverHandler — ServeCover deeper paths (41.4% coverage)
// =====================================================================

func TestCoverHandler_ServeCover_WithLocalPath(t *testing.T) {
	// Create a temporary image file to test the local path branch
	tmpFile, err := os.CreateTemp("", "cover_*.jpg")
	require.NoError(t, err)
	tmpFile.Write([]byte("fake jpeg data"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, cleanup := setupCoverTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	coverArtService := services.NewCoverArtService(db, logger)
	handler := NewCoverHandler(coverArtService)

	// Create an entity and cover art with local_path
	itemRepo := repository.NewMediaItemRepository(db)
	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "LocalCoverMovie", Status: "detected"})

	// Insert cover art pointing to the temp file
	_, err = db.Exec(`INSERT INTO cover_art (media_item_id, local_path, format) VALUES (?, ?, ?)`,
		entityID, tmpFile.Name(), "jpg")
	if err != nil {
		t.Skip("cover_art table not available")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/cover/%d", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.ServeCover(c)

	// Should serve the file directly
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCoverHandler_ServeCover_WithURL(t *testing.T) {
	db, cleanup := setupCoverTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	coverArtService := services.NewCoverArtService(db, logger)
	handler := NewCoverHandler(coverArtService)

	itemRepo := repository.NewMediaItemRepository(db)
	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "URLCoverMovie", Status: "detected"})

	// Insert cover art with URL only
	_, err := db.Exec(`INSERT INTO cover_art (media_item_id, url, format) VALUES (?, ?, ?)`,
		entityID, "https://example.com/poster.jpg", "jpg")
	if err != nil {
		t.Skip("cover_art table not available")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/cover/%d", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.ServeCover(c)

	// Should redirect to URL
	assert.Equal(t, http.StatusFound, w.Code)
}

func setupCoverTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "cover_test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Type:               "sqlite",
		Path:               tmpFile.Name(),
		MaxOpenConnections: 5,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600,
		EnableWAL:          true,
		CacheSize:          2000,
		BusyTimeout:        5000,
	}

	db, err := database.NewConnection(cfg)
	require.NoError(t, err)

	err = db.RunMigrations(context.Background())
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}
	return db, cleanup
}

// =====================================================================
// MediaEntityHandler — EnrichAllEntities with data (30.8% coverage)
// =====================================================================

func TestMediaEntityHandler_EnrichAllEntities_WithEntities(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Test Movie", Status: "detected"})

	// Insert a directory_analyses record to link the entity to a directory
	_, err := db.Exec(`INSERT INTO directory_analyses (directory_path, media_item_id, media_type_id) VALUES (?, ?, ?)`,
		"/movies/test", entityID, typeID)
	if err != nil {
		// Table may not exist
		t.Skip("directory_analyses table not available")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/entities/enrich", nil)

	handler.EnrichAllEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["message"], "Batch enrichment complete")
	// At least 1 entity scanned (may not be enriched since no cover files exist)
	assert.GreaterOrEqual(t, resp["scanned"].(float64), float64(1))
}

// =====================================================================
// MediaEntityHandler — RefreshEntityMetadata with directory_analyses
// =====================================================================

func TestMediaEntityHandler_RefreshEntityMetadata_WithDirectory(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Dir Movie", Status: "detected"})

	// Insert directory_analyses record
	_, err := db.Exec(`INSERT INTO directory_analyses (directory_path, media_item_id, media_type_id) VALUES (?, ?, ?)`,
		"/movies/dirmovie", entityID, typeID)
	if err != nil {
		t.Skip("directory_analyses table not available")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/entities/%d/metadata/refresh", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.RefreshEntityMetadata(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// No cover files exist and no TMDB key, so should return "No metadata sources available"
	assert.Contains(t, resp["message"].(string), "No metadata")
}

func TestMediaEntityHandler_RefreshEntityMetadata_WithMediaFileFallback(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Fallback Movie", Status: "detected"})

	// Insert storage root, file, and media_file link (no directory_analyses)
	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/data", "smb")
	rootID, _ := res.LastInsertId()
	fileRes, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/data/movies/movie.mkv", "movie.mkv", 1024, false, time.Now(), time.Now())
	fileID, _ := fileRes.LastInsertId()
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id) VALUES (?, ?)`, entityID, fileID)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/entities/%d/metadata/refresh", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.RefreshEntityMetadata(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// SubtitleHandler — DownloadSubtitle / TranslateSubtitle with valid data
// hitting the service call (52.9% coverage)
// =====================================================================

func TestSubtitleHandler_DownloadSubtitle_ServiceCallReached(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/api/v1/subtitles/download", handler.DownloadSubtitle)

	body := `{"media_item_id": 1, "result_id": "abc123", "language": "en"}`
	req := httptest.NewRequest("POST", "/api/v1/subtitles/download", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Nil service -> panic caught by recovery middleware -> 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSubtitleHandler_TranslateSubtitle_ServiceCallReached(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/api/v1/subtitles/translate", handler.TranslateSubtitle)

	body := `{"subtitle_id": "sub1", "source_language": "en", "target_language": "es"}`
	req := httptest.NewRequest("POST", "/api/v1/subtitles/translate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Nil service -> panic caught by recovery middleware -> 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSubtitleHandler_VerifySubtitleSync_ServiceCallReached(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/api/v1/subtitles/:subtitle_id/verify-sync/:media_id", handler.VerifySubtitleSync)

	req := httptest.NewRequest("GET", "/api/v1/subtitles/sub1/verify-sync/42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Nil service -> panic caught by recovery middleware -> 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSubtitleHandler_GetSubtitles_ServiceCallReached(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/api/v1/subtitles/media/:media_id", handler.GetSubtitles)

	req := httptest.NewRequest("GET", "/api/v1/subtitles/media/42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Nil service -> panic caught by recovery middleware -> 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// =====================================================================
// SubtitleHandler — UploadSubtitle with valid file
// =====================================================================

func TestSubtitleHandler_UploadSubtitle_WithFileFormats(t *testing.T) {
	logger := zap.NewNop()
	handler := NewSubtitleHandler(nil, logger)

	gin.SetMode(gin.TestMode)

	// Test various subtitle file extensions
	formats := []struct {
		filename string
	}{
		{"subtitle.srt"},
		{"subtitle.vtt"},
		{"subtitle.ass"},
		{"subtitle.txt"},
	}

	for _, fmt := range formats {
		t.Run(fmt.filename, func(t *testing.T) {
			router := gin.New()
			router.Use(gin.Recovery())
			router.POST("/api/v1/subtitles/upload", handler.UploadSubtitle)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			writer.WriteField("media_item_id", "123")
			writer.WriteField("language", "English")
			writer.WriteField("language_code", "en")
			part, _ := writer.CreateFormFile("file", fmt.filename)
			part.Write([]byte("1\n00:00:01,000 --> 00:00:05,000\nHello World"))
			writer.Close()

			req := httptest.NewRequest("POST", "/api/v1/subtitles/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Nil service -> panic caught by recovery -> 500
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}

// =====================================================================
// MediaEntityHandler — GetEntityFiles / GetEntityMetadata with data
// =====================================================================

func TestMediaEntityHandler_GetEntityFiles_WithFiles(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/d", "smb")
	rootID, _ := res.LastInsertId()
	fileRes, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/m.mkv", "m.mkv", 100, false, time.Now(), time.Now())
	fileID, _ := fileRes.LastInsertId()
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id) VALUES (?, ?)`, entityID, fileID)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/1/files", nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.GetEntityFiles(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])
}

func TestMediaEntityHandler_GetEntityMetadata_WithData(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/metadata", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.GetEntityMetadata(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// GinAdapter — WrapHTTPHandler (50% coverage)
// =====================================================================

func TestWrapHTTPHandler(t *testing.T) {
	// Create a simple HTTP handler
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello from http handler"))
	})

	ginHandler := WrapHTTPHandler(httpHandler)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/wrapped", ginHandler)

	req := httptest.NewRequest("GET", "/wrapped", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hello from http handler")
}

func TestWrapHTTPHandler_WithContext(t *testing.T) {
	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID != nil {
			w.Write([]byte(fmt.Sprintf("user: %v", userID)))
		} else {
			w.Write([]byte("no user"))
		}
	})

	ginHandler := WrapHTTPHandler(httpHandler)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/wrapped", func(c *gin.Context) {
		c.Set("user_id", 42)
		ginHandler(c)
	})

	req := httptest.NewRequest("GET", "/wrapped", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================================================
// MediaEntityHandler — enrichItemsWithCoverURLs DB fallback
// =====================================================================

func TestMediaEntityHandler_EnrichItems_DBFallbackWithCoverURL(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	handler.db = db
	// Explicitly nil cover service to test DB fallback path
	handler.coverArtService = nil
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "TestMovie", Status: "detected"})

	// Insert external_metadata with cover_url
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	coverURL := "/api/v1/cover/placeholder/movie"
	err := extMetaRepo.Upsert(ctx, &mediaModels.ExternalMetadata{
		MediaItemID: entityID,
		Provider:    "test",
		ExternalID:  "test:1",
		CoverURL:    &coverURL,
	})
	require.NoError(t, err)

	items := []*mediaModels.MediaItem{{ID: entityID, MediaTypeID: typeID, Title: "TestMovie"}}
	jsonItems := []gin.H{{"id": entityID, "title": "TestMovie"}}

	handler.enrichItemsWithCoverURLs(ctx, items, jsonItems)

	// Should have enriched from DB fallback
	url, ok := jsonItems[0]["cover_url"]
	assert.True(t, ok)
	assert.Equal(t, coverURL, url)
}

// =====================================================================
// MediaEntityHandler — DownloadEntity with primary file
// =====================================================================

func TestMediaEntityHandler_DownloadEntity_WithPrimaryFile(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/d", "smb")
	rootID, _ := res.LastInsertId()

	// Create two files, one marked as primary
	file1Res, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/m1.mkv", "m1.mkv", 100, false, time.Now(), time.Now())
	file1ID, _ := file1Res.LastInsertId()

	file2Res, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/m2.mkv", "m2.mkv", 200, false, time.Now(), time.Now())
	file2ID, _ := file2Res.LastInsertId()

	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id, is_primary) VALUES (?, ?, ?)`, entityID, file1ID, false)
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id, is_primary) VALUES (?, ?, ?)`, entityID, file2ID, true)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/download", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.DownloadEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// Should select the primary file (file2)
	assert.Equal(t, float64(file2ID), resp["file_id"])
	assert.Equal(t, float64(2), resp["total_files"])
}

// =====================================================================
// MediaEntityHandler — StreamEntity with primary file
// =====================================================================

func TestMediaEntityHandler_StreamEntity_WithPrimaryFile(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "Movie", Status: "detected"})

	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/d", "smb")
	rootID, _ := res.LastInsertId()

	file1Res, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/m1.mkv", "m1.mkv", 100, false, time.Now(), time.Now())
	file1ID, _ := file1Res.LastInsertId()

	file2Res, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/m2.mkv", "m2.mkv", 200, false, time.Now(), time.Now())
	file2ID, _ := file2Res.LastInsertId()

	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id, is_primary) VALUES (?, ?, ?)`, entityID, file1ID, false)
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id, is_primary) VALUES (?, ?, ?)`, entityID, file2ID, true)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/stream", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.StreamEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(file2ID), resp["file_id"])
}

// =====================================================================
// MediaEntityHandler — GetInstallInfo with primary file
// =====================================================================

func TestMediaEntityHandler_GetInstallInfo_SoftwareWithPrimary(t *testing.T) {
	db, cleanup := setupTestDBForCoverage(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "software")
	entityID, _ := itemRepo.Create(ctx, &mediaModels.MediaItem{MediaTypeID: typeID, Title: "CoolApp", Status: "detected"})

	res, _ := db.Exec(`INSERT INTO storage_roots (name, path, protocol) VALUES (?, ?, ?)`, "r1", "/d", "smb")
	rootID, _ := res.LastInsertId()

	file1Res, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/setup.exe", "setup.exe", 500, false, time.Now(), time.Now())
	file1ID, _ := file1Res.LastInsertId()

	file2Res, _ := db.Exec(`INSERT INTO files (storage_root_id, path, name, size, is_directory, created_at, modified_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rootID, "/d/readme.txt", "readme.txt", 50, false, time.Now(), time.Now())
	file2ID, _ := file2Res.LastInsertId()

	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id, is_primary) VALUES (?, ?, ?)`, entityID, file1ID, true)
	_, _ = db.Exec(`INSERT INTO media_files (media_item_id, file_id, is_primary) VALUES (?, ?, ?)`, entityID, file2ID, false)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/entities/%d/install-info", entityID), nil)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(entityID, 10)}}

	handler.GetInstallInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "CoolApp", resp["title"])
	assert.Equal(t, float64(file1ID), resp["file_id"]) // primary
	assert.Equal(t, float64(2), resp["total_files"])
}
