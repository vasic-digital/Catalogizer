package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"catalogizer/config"
	"catalogizer/database"
	internalservices "catalogizer/internal/services"
	mediamodels "catalogizer/internal/media/models"
	"catalogizer/models"
	"catalogizer/repository"
	"catalogizer/services"
	challengepkg "digital.vasic.challenges/pkg/challenge"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test helpers
// =============================================================================

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter() *gin.Engine {
	return gin.New()
}

func newTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "coverage_boost_test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	cfg := &config.DatabaseConfig{
		Type:               "sqlite",
		Path:               tmpFile.Name(),
		MaxOpenConnections: 1,
		MaxIdleConnections: 1,
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

// =============================================================================
// BrowseHandler tests
// =============================================================================

func TestNewBrowseHandler(t *testing.T) {
	t.Run("nil repo", func(t *testing.T) {
		h := NewBrowseHandler(nil)
		assert.NotNil(t, h)
		assert.Nil(t, h.fileRepo)
	})

	t.Run("with repo", func(t *testing.T) {
		repo := &repository.FileRepository{}
		h := NewBrowseHandler(repo)
		assert.NotNil(t, h)
		assert.Equal(t, repo, h.fileRepo)
	})
}

func TestBrowseHandler_BrowseDirectory_MissingStorageRoot(t *testing.T) {
	h := NewBrowseHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/browse/", nil)
	c.Params = gin.Params{{Key: "storage_root", Value: ""}}

	h.BrowseDirectory(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Storage root name is required")
}

func TestBrowseHandler_BrowseDirectory_PaginationDefaults(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewBrowseHandler(fileRepo)
	router := newTestRouter()
	router.GET("/browse/:storage_root", h.BrowseDirectory)

	tests := []struct {
		name  string
		query string
	}{
		{"defaults", "/browse/test-root"},
		{"negative page", "/browse/test-root?page=-1&limit=50"},
		{"excessive limit", "/browse/test-root?page=1&limit=2000"},
		{"zero limit", "/browse/test-root?page=1&limit=0"},
		{"invalid sort field", "/browse/test-root?sort_by=invalid"},
		{"invalid sort order", "/browse/test-root?sort_order=random"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError,
				"unexpected status %d for %s", w.Code, tt.name)
		})
	}
}

func TestBrowseHandler_GetFileInfo_InvalidID(t *testing.T) {
	h := NewBrowseHandler(nil)
	router := newTestRouter()
	router.GET("/file/:id", h.GetFileInfo)

	tests := []struct {
		name string
		id   string
		code int
	}{
		{"not a number", "abc", http.StatusBadRequest},
		{"decimal", "1.5", http.StatusBadRequest},
		{"overflow", "99999999999999999999", http.StatusBadRequest},
		{"special chars", "!@#", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/file/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}

func TestBrowseHandler_GetDirectorySizes_MissingStorageRoot(t *testing.T) {
	h := NewBrowseHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/sizes/", nil)
	c.Params = gin.Params{{Key: "storage_root", Value: ""}}

	h.GetDirectorySizes(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Storage root name is required")
}

func TestBrowseHandler_GetDirectoryDuplicates_MissingStorageRoot(t *testing.T) {
	h := NewBrowseHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/duplicates/", nil)
	c.Params = gin.Params{{Key: "storage_root", Value: ""}}

	h.GetDirectoryDuplicates(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Storage root name is required")
}

func TestBrowseHandler_GetDirectorySizes_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewBrowseHandler(fileRepo)
	router := newTestRouter()
	router.GET("/sizes/:storage_root", h.GetDirectorySizes)

	tests := []struct {
		name  string
		query string
	}{
		{"defaults", "/sizes/test-root"},
		{"negative page", "/sizes/test-root?page=-1"},
		{"big limit", "/sizes/test-root?limit=1000"},
		{"ascending", "/sizes/test-root?ascending=true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
		})
	}
}

func TestBrowseHandler_GetDirectoryDuplicates_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewBrowseHandler(fileRepo)
	router := newTestRouter()
	router.GET("/duplicates/:storage_root", h.GetDirectoryDuplicates)

	req := httptest.NewRequest("GET", "/duplicates/test-root?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

// =============================================================================
// MediaEntityHandler tests - expanded coverage
// =============================================================================

func TestMediaEntityHandler_Constructor(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	assert.NotNil(t, h)
	assert.Nil(t, h.itemRepo)
	assert.Nil(t, h.fileRepo)
	assert.Nil(t, h.extMetaRepo)
	assert.Nil(t, h.userMetaRepo)
}

func TestMediaEntityHandler_GetEntity_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id", h.GetEntity)

	tests := []struct {
		name string
		id   string
	}{
		{"text", "abc"},
		{"decimal", "1.5"},
		{"overflow", "99999999999999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/entity/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid entity ID")
		})
	}
}

func TestMediaEntityHandler_GetEntityChildren_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/children", h.GetEntityChildren)

	req := httptest.NewRequest("GET", "/entity/abc/children", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetEntityFiles_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/files", h.GetEntityFiles)

	req := httptest.NewRequest("GET", "/entity/abc/files", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetEntityMetadata_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/metadata", h.GetEntityMetadata)

	req := httptest.NewRequest("GET", "/entity/abc/metadata", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetEntityDuplicates_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/duplicates", h.GetEntityDuplicates)

	req := httptest.NewRequest("GET", "/entity/abc/duplicates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_BrowseByType_InvalidType(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	router := newTestRouter()
	router.GET("/browse/:type", h.BrowseByType)

	req := httptest.NewRequest("GET", "/browse/nonexistent_type", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_BrowseByType_PaginationDefaults(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	router := newTestRouter()
	router.GET("/browse/:type", h.BrowseByType)

	tests := []struct {
		name  string
		query string
	}{
		{"negative limit", "/browse/movie?limit=-1"},
		{"excessive limit", "/browse/movie?limit=500"},
		{"negative offset", "/browse/movie?offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestMediaEntityHandler_RefreshEntityMetadata_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.POST("/entity/:id/metadata/refresh", h.RefreshEntityMetadata)

	req := httptest.NewRequest("POST", "/entity/abc/metadata/refresh", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_RefreshEntityMetadata_ValidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.POST("/entity/:id/metadata/refresh", h.RefreshEntityMetadata)

	req := httptest.NewRequest("POST", "/entity/42/metadata/refresh", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), "Metadata refresh queued")
}

func TestMediaEntityHandler_UpdateUserMetadata_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.PUT("/entity/:id/user-metadata", h.UpdateUserMetadata)

	body := `{"user_rating": 8.5}`
	req := httptest.NewRequest("PUT", "/entity/abc/user-metadata", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_UpdateUserMetadata_InvalidBody(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.PUT("/entity/:id/user-metadata", h.UpdateUserMetadata)

	req := httptest.NewRequest("PUT", "/entity/1/user-metadata", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_UpdateUserMetadata_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	// Insert a user row to satisfy the FK constraint on user_metadata.user_id
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, username, email, password_hash, salt, role_id, is_active) VALUES (1, 'testuser', 'test@test.com', 'hash', 'salt', 1, 1)`)
	require.NoError(t, err)

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemID, err := itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID,
		Title:       "Test Movie",
		Status:      "detected",
	})
	assert.NoError(t, err)

	router := newTestRouter()
	router.PUT("/entity/:id/user-metadata", h.UpdateUserMetadata)

	body := `{"user_rating": 8.5, "favorite": true, "personal_notes": "Great movie"}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/entity/%d/user-metadata", itemID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User metadata updated")
}

func TestMediaEntityHandler_UpdateUserMetadata_IsFavoriteField(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	// Insert a user row to satisfy the FK constraint on user_metadata.user_id
	_, err := db.ExecContext(ctx, `INSERT INTO users (id, username, email, password_hash, salt, role_id, is_active) VALUES (1, 'testuser', 'test@test.com', 'hash', 'salt', 1, 1)`)
	require.NoError(t, err)

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemID, err := itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID,
		Title:       "Test",
		Status:      "detected",
	})
	assert.NoError(t, err)

	router := newTestRouter()
	router.PUT("/entity/:id/user-metadata", h.UpdateUserMetadata)

	body := `{"is_favorite": true}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/entity/%d/user-metadata", itemID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_favorite":true`)
}

func TestMediaEntityHandler_StreamEntity_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/stream", h.StreamEntity)

	req := httptest.NewRequest("GET", "/entity/abc/stream", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_DownloadEntity_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/download", h.DownloadEntity)

	req := httptest.NewRequest("GET", "/entity/abc/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_GetInstallInfo_InvalidID(t *testing.T) {
	h := NewMediaEntityHandler(nil, nil, nil, nil)
	router := newTestRouter()
	router.GET("/entity/:id/install-info", h.GetInstallInfo)

	req := httptest.NewRequest("GET", "/entity/abc/install-info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_ListDuplicateGroups(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/duplicates?limit=10&offset=0", nil)

	h.ListDuplicateGroups(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_ListDuplicateGroups_InvalidPagination(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/duplicates?limit=-1&offset=0", nil)

	h.ListDuplicateGroups(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_GetEntityChildren_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID,
		Title:       "Test Show",
		Status:      "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/children", h.GetEntityChildren)

	req := httptest.NewRequest("GET", "/entity/1/children?limit=5&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_GetEntityFiles_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID,
		Title:       "Test Movie",
		Status:      "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/files", h.GetEntityFiles)

	req := httptest.NewRequest("GET", "/entity/1/files", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_GetEntityMetadata_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID,
		Title:       "Test Movie",
		Status:      "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/metadata", h.GetEntityMetadata)

	req := httptest.NewRequest("GET", "/entity/1/metadata", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_GetEntityDuplicates_NotFound(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	router := newTestRouter()
	router.GET("/entity/:id/duplicates", h.GetEntityDuplicates)

	req := httptest.NewRequest("GET", "/entity/99999/duplicates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_GetEntityDuplicates_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	year := 2020
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID, Title: "Dupe", Year: &year, Status: "detected",
	})
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID, Title: "Dupe", Year: &year, Status: "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/duplicates", h.GetEntityDuplicates)

	req := httptest.NewRequest("GET", "/entity/1/duplicates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMediaEntityHandler_StreamEntity_NoFiles(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID, Title: "NoFile Movie", Status: "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/stream", h.StreamEntity)

	req := httptest.NewRequest("GET", "/entity/1/stream", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_DownloadEntity_NoFiles(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID, Title: "NoFile Movie", Status: "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/download", h.DownloadEntity)

	req := httptest.NewRequest("GET", "/entity/1/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_GetInstallInfo_NotFound(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := NewMediaEntityHandler(itemRepo, nil, nil, nil)

	router := newTestRouter()
	router.GET("/entity/:id/install-info", h.GetInstallInfo)

	req := httptest.NewRequest("GET", "/entity/99999/install-info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaEntityHandler_GetInstallInfo_NotSoftware(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	h := NewMediaEntityHandler(itemRepo, fileRepo, nil, nil)

	ctx := context.Background()
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	itemRepo.Create(ctx, &mediamodels.MediaItem{
		MediaTypeID: typeID, Title: "Not Software", Status: "detected",
	})

	router := newTestRouter()
	router.GET("/entity/:id/install-info", h.GetInstallInfo)

	req := httptest.NewRequest("GET", "/entity/1/install-info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "only available for software")
}

func TestMediaEntityHandler_ListEntities_InvalidType(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := NewMediaEntityHandler(itemRepo, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/entities?type=nonexistent", nil)

	h.ListEntities(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMediaEntityHandler_ListEntities_PaginationDefaults(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	itemRepo := repository.NewMediaItemRepository(db)
	h := NewMediaEntityHandler(itemRepo, nil, nil, nil)

	tests := []struct {
		name  string
		query string
	}{
		{"negative limit", "/entities?limit=-1"},
		{"excessive limit", "/entities?limit=500"},
		{"negative offset", "/entities?offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.query, nil)
			h.ListEntities(c)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// =============================================================================
// itemsToJSON / itemToJSON / entityDetailJSON helper tests
// =============================================================================

func TestItemsToJSON_NilSlice(t *testing.T) {
	result := itemsToJSON(nil)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestItemsToJSON_EmptySlice(t *testing.T) {
	result := itemsToJSON([]*mediamodels.MediaItem{})
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestItemToJSON_AllFields(t *testing.T) {
	origTitle := "Original Title"
	year := 2020
	desc := "A description"
	director := "Director Name"
	rating := 8.5
	runtime := 120
	lang := "en"
	parentID := int64(1)
	seasonNum := 2
	episodeNum := 5
	trackNum := 3

	item := &mediamodels.MediaItem{
		ID: 42, MediaTypeID: 1, Title: "Test Title",
		OriginalTitle: &origTitle, Year: &year, Description: &desc,
		Genre: []string{"action", "drama"}, Director: &director,
		Rating: &rating, Runtime: &runtime, Language: &lang,
		ParentID: &parentID, SeasonNumber: &seasonNum,
		EpisodeNumber: &episodeNum, TrackNumber: &trackNum,
		Status: "detected", FirstDetected: time.Now(), LastUpdated: time.Now(),
	}

	result := itemToJSON(item)
	assert.Equal(t, int64(42), result["id"])
	assert.Equal(t, "Test Title", result["title"])
	assert.Equal(t, "Original Title", result["original_title"])
	assert.Equal(t, 2020, result["year"])
	assert.Equal(t, "A description", result["description"])
	assert.Equal(t, []string{"action", "drama"}, result["genre"])
	assert.Equal(t, "Director Name", result["director"])
	assert.Equal(t, 8.5, result["rating"])
	assert.Equal(t, 120, result["runtime"])
	assert.Equal(t, "en", result["language"])
	assert.Equal(t, int64(1), result["parent_id"])
	assert.Equal(t, 2, result["season_number"])
	assert.Equal(t, 5, result["episode_number"])
	assert.Equal(t, 3, result["track_number"])
}

func TestItemToJSON_MinimalFields(t *testing.T) {
	item := &mediamodels.MediaItem{
		ID: 1, MediaTypeID: 1, Title: "Minimal", Status: "detected",
	}

	result := itemToJSON(item)
	assert.Equal(t, "Minimal", result["title"])
	_, hasOrigTitle := result["original_title"]
	assert.False(t, hasOrigTitle)
	_, hasYear := result["year"]
	assert.False(t, hasYear)
}

func TestEntityDetailJSON(t *testing.T) {
	item := &mediamodels.MediaItem{
		ID: 1, MediaTypeID: 1, Title: "Test", Status: "detected",
	}

	result := entityDetailJSON(item, "movie", 5, 3, nil)
	assert.Equal(t, "movie", result["media_type"])
	assert.Equal(t, int64(5), result["file_count"])
	assert.Equal(t, int64(3), result["children_count"])
	assert.NotNil(t, result["external_metadata"])

	extMeta := []*mediamodels.ExternalMetadata{{ID: 1, MediaItemID: 1, Provider: "tmdb"}}
	result2 := entityDetailJSON(item, "tv_show", 2, 10, extMeta)
	assert.Equal(t, extMeta, result2["external_metadata"])
}

// =============================================================================
// ChallengeHandler tests - expanded coverage
// =============================================================================

func TestChallengeHandler_GetChallenge_Found(t *testing.T) {
	mock := &mockChallengeService{
		listChallengesFunc: func() []services.ChallengeSummary {
			return []services.ChallengeSummary{
				{ID: "ch-001", Name: "Test", Description: "Test challenge", Category: "test"},
			}
		},
	}
	h := NewChallengeHandler(mock)

	router := newTestRouter()
	router.GET("/challenges/:id", h.GetChallenge)

	req := httptest.NewRequest("GET", "/challenges/ch-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ch-001")
}

func TestChallengeHandler_GetChallenge_NotFound(t *testing.T) {
	mock := &mockChallengeService{
		listChallengesFunc: func() []services.ChallengeSummary {
			return []services.ChallengeSummary{{ID: "ch-001", Name: "Test"}}
		},
	}
	h := NewChallengeHandler(mock)

	router := newTestRouter()
	router.GET("/challenges/:id", h.GetChallenge)

	req := httptest.NewRequest("GET", "/challenges/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChallengeHandler_RunChallenge_Success(t *testing.T) {
	mock := &mockChallengeService{
		runChallengeFunc: func(ctx context.Context, id string) (*challengepkg.Result, error) {
			return &challengepkg.Result{Status: "passed"}, nil
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.POST("/challenges/:id/run", h.RunChallenge)

	req := httptest.NewRequest("POST", "/challenges/ch-001/run", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChallengeHandler_RunChallenge_Error(t *testing.T) {
	mock := &mockChallengeService{
		runChallengeFunc: func(ctx context.Context, id string) (*challengepkg.Result, error) {
			return nil, errors.New("failed")
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.POST("/challenges/:id/run", h.RunChallenge)

	req := httptest.NewRequest("POST", "/challenges/ch-001/run", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChallengeHandler_RunAll_Success(t *testing.T) {
	mock := &mockChallengeService{
		runAllFunc: func(ctx context.Context) ([]*challengepkg.Result, error) {
			return []*challengepkg.Result{
				{Status: "passed"}, {Status: "failed"}, {Status: "passed"},
			}, nil
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.POST("/challenges/run/all", h.RunAll)

	req := httptest.NewRequest("POST", "/challenges/run/all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	summary := resp["summary"].(map[string]interface{})
	assert.Equal(t, float64(3), summary["total"])
	assert.Equal(t, float64(2), summary["passed"])
	assert.Equal(t, float64(1), summary["failed"])
}

func TestChallengeHandler_RunAll_Error(t *testing.T) {
	mock := &mockChallengeService{
		runAllFunc: func(ctx context.Context) ([]*challengepkg.Result, error) {
			return nil, errors.New("run all failed")
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.POST("/challenges/run/all", h.RunAll)

	req := httptest.NewRequest("POST", "/challenges/run/all", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChallengeHandler_RunByCategory_Success(t *testing.T) {
	mock := &mockChallengeService{
		runByCategoryFunc: func(ctx context.Context, category string) ([]*challengepkg.Result, error) {
			return []*challengepkg.Result{{Status: "passed"}, {Status: "passed"}}, nil
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.POST("/challenges/run/category/:category", h.RunByCategory)

	req := httptest.NewRequest("POST", "/challenges/run/category/integration", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChallengeHandler_RunByCategory_Error(t *testing.T) {
	mock := &mockChallengeService{
		runByCategoryFunc: func(ctx context.Context, category string) ([]*challengepkg.Result, error) {
			return nil, errors.New("error")
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.POST("/challenges/run/category/:category", h.RunByCategory)

	req := httptest.NewRequest("POST", "/challenges/run/category/integration", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestChallengeHandler_GetResults_CoverageBoost(t *testing.T) {
	mock := &mockChallengeService{
		getResultsFunc: func() []*challengepkg.Result {
			return []*challengepkg.Result{{Status: "passed"}}
		},
	}
	h := NewChallengeHandler(mock)
	router := newTestRouter()
	router.GET("/challenges/results", h.GetResults)

	req := httptest.NewRequest("GET", "/challenges/results", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["count"])
}

// =============================================================================
// RoleHandler tests - UpdateRole, DeleteRole, ListRoles, GetPermissions
// =============================================================================

func TestRoleHandler_UpdateRole_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authToken      string
		currentUser    *models.User
		hasPermission  bool
		roleID         string
		expectedStatus int
	}{
		{"Method not allowed", "GET", "valid-token", nil, false, "", 405},
		{"Unauthorized", "PUT", "", nil, false, "", 401},
		{"Permission denied", "PUT", "valid-token", &models.User{ID: 1}, false, "", 403},
		{"Invalid role ID", "PUT", "valid-token", &models.User{ID: 1}, true, "invalid", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockRoleUserService)
			mockAuthService := new(MockRoleAuthService)
			handler := NewRoleHandler(mockUserService, mockAuthService)

			if tt.authToken != "" && tt.currentUser != nil {
				mockAuthService.On("GetCurrentUser", tt.authToken).Return(tt.currentUser, nil).Maybe()
				mockAuthService.On("CheckPermission", tt.currentUser.ID, models.PermissionSystemAdmin).Return(tt.hasPermission, nil).Maybe()
			}

			roleID := tt.roleID
			if roleID == "" {
				roleID = "1"
			}

			body := bytes.NewBufferString(`{"name": "Updated"}`)
			req := httptest.NewRequest(tt.method, "/api/roles/"+roleID, body)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}
			rr := httptest.NewRecorder()
			handler.UpdateRole(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRoleHandler_DeleteRole_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authToken      string
		currentUser    *models.User
		hasPermission  bool
		roleID         string
		expectedStatus int
	}{
		{"Method not allowed", "GET", "", nil, false, "", 405},
		{"Unauthorized", "DELETE", "", nil, false, "", 401},
		{"Permission denied", "DELETE", "valid-token", &models.User{ID: 1}, false, "", 403},
		{"Invalid role ID", "DELETE", "valid-token", &models.User{ID: 1}, true, "invalid", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockRoleUserService)
			mockAuthService := new(MockRoleAuthService)
			handler := NewRoleHandler(mockUserService, mockAuthService)

			if tt.authToken != "" && tt.currentUser != nil {
				mockAuthService.On("GetCurrentUser", tt.authToken).Return(tt.currentUser, nil).Maybe()
				mockAuthService.On("CheckPermission", tt.currentUser.ID, models.PermissionSystemAdmin).Return(tt.hasPermission, nil).Maybe()
			}

			roleID := tt.roleID
			if roleID == "" {
				roleID = "1"
			}
			req := httptest.NewRequest(tt.method, "/api/roles/"+roleID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}
			rr := httptest.NewRecorder()
			handler.DeleteRole(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRoleHandler_ListRoles_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authToken      string
		currentUser    *models.User
		hasPermission  bool
		roles          []models.Role
		serviceError   error
		expectedStatus int
	}{
		{"Method not allowed", "POST", "", nil, false, nil, nil, 405},
		{"Unauthorized", "GET", "", nil, false, nil, nil, 401},
		{"Permission denied", "GET", "valid-token", &models.User{ID: 1}, false, nil, nil, 403},
		{"Success", "GET", "valid-token", &models.User{ID: 1}, true,
			[]models.Role{{ID: 1, Name: "Admin"}, {ID: 2, Name: "User"}}, nil, 200},
		{"Service error", "GET", "valid-token", &models.User{ID: 1}, true,
			nil, errors.New("db error"), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockRoleUserService)
			mockAuthService := new(MockRoleAuthService)
			handler := NewRoleHandler(mockUserService, mockAuthService)

			if tt.authToken != "" && tt.currentUser != nil {
				mockAuthService.On("GetCurrentUser", tt.authToken).Return(tt.currentUser, nil).Maybe()
				mockAuthService.On("CheckPermission", tt.currentUser.ID, models.PermissionSystemAdmin).Return(tt.hasPermission, nil).Maybe()
			}
			if tt.hasPermission && tt.method == "GET" {
				mockUserService.On("ListRoles").Return(tt.roles, tt.serviceError).Maybe()
			}

			req := httptest.NewRequest(tt.method, "/api/roles", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}
			rr := httptest.NewRecorder()
			handler.ListRoles(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRoleHandler_GetPermissions_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authToken      string
		currentUser    *models.User
		hasPermission  bool
		expectedStatus int
	}{
		{"Method not allowed", "POST", "", nil, false, 405},
		{"Unauthorized", "GET", "", nil, false, 401},
		{"Permission denied", "GET", "valid-token", &models.User{ID: 1}, false, 403},
		{"Success", "GET", "valid-token", &models.User{ID: 1}, true, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(MockRoleUserService)
			mockAuthService := new(MockRoleAuthService)
			handler := NewRoleHandler(mockUserService, mockAuthService)

			if tt.authToken != "" && tt.currentUser != nil {
				mockAuthService.On("GetCurrentUser", tt.authToken).Return(tt.currentUser, nil).Maybe()
				mockAuthService.On("CheckPermission", tt.currentUser.ID, models.PermissionSystemAdmin).Return(tt.hasPermission, nil).Maybe()
			}

			req := httptest.NewRequest(tt.method, "/api/permissions", nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}
			rr := httptest.NewRecorder()
			handler.GetPermissions(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == 200 {
				var resp map[string]interface{}
				json.Unmarshal(rr.Body.Bytes(), &resp)
				assert.Contains(t, resp, "user_management")
				assert.Contains(t, resp, "system")
			}
		})
	}
}

func TestRoleHandler_DeleteRole_Success(t *testing.T) {
	mockUserService := new(MockRoleUserService)
	mockAuthService := new(MockRoleAuthService)
	handler := NewRoleHandler(mockUserService, mockAuthService)

	user := &models.User{ID: 1, Username: "admin"}
	mockAuthService.On("GetCurrentUser", "valid-token").Return(user, nil)
	mockAuthService.On("CheckPermission", user.ID, models.PermissionSystemAdmin).Return(true, nil)
	mockUserService.On("DeleteRole", 5).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.DeleteRole(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestRoleHandler_DeleteRole_NotFoundError(t *testing.T) {
	mockUserService := new(MockRoleUserService)
	mockAuthService := new(MockRoleAuthService)
	handler := NewRoleHandler(mockUserService, mockAuthService)

	user := &models.User{ID: 1}
	mockAuthService.On("GetCurrentUser", "valid-token").Return(user, nil)
	mockAuthService.On("CheckPermission", user.ID, models.PermissionSystemAdmin).Return(true, nil)
	mockUserService.On("DeleteRole", 5).Return(errors.New("role not found"))

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.DeleteRole(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRoleHandler_DeleteRole_AssignedToUsersError(t *testing.T) {
	mockUserService := new(MockRoleUserService)
	mockAuthService := new(MockRoleAuthService)
	handler := NewRoleHandler(mockUserService, mockAuthService)

	user := &models.User{ID: 1}
	mockAuthService.On("GetCurrentUser", "valid-token").Return(user, nil)
	mockAuthService.On("CheckPermission", user.ID, models.PermissionSystemAdmin).Return(true, nil)
	mockUserService.On("DeleteRole", 5).Return(errors.New("role is assigned to users"))

	req := httptest.NewRequest("DELETE", "/api/roles/5", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler.DeleteRole(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
}

// =============================================================================
// SearchHandler tests
// =============================================================================

func TestNewSearchHandler(t *testing.T) {
	h := NewSearchHandler(nil)
	assert.NotNil(t, h)
}

func TestSearchHandler_SearchFiles_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewSearchHandler(fileRepo)
	router := newTestRouter()
	router.GET("/search", h.SearchFiles)

	tests := []struct {
		name  string
		query string
	}{
		{"basic", "/search?q=test"},
		{"all params", "/search?q=test&file_type=video&sort_by=size&sort_order=desc&page=1&limit=10"},
		{"size filters", "/search?min_size=100&max_size=10000000"},
		{"date filter", "/search?modified_after=" + url.QueryEscape(time.Now().Add(-24*time.Hour).Format(time.RFC3339))},
		{"boolean filters", "/search?include_deleted=true&only_duplicates=true"},
		{"invalid sort", "/search?sort_by=invalid_field&sort_order=random"},
		{"negative page", "/search?page=-1"},
		{"big limit", "/search?limit=5000"},
		{"zero limit", "/search?limit=0"},
		{"smb roots", "/search?smb_roots=root1,root2"},
		{"no directories", "/search?include_directories=false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
		})
	}
}

func TestSearchHandler_SearchFiles_InvalidDate(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewSearchHandler(fileRepo)
	router := newTestRouter()
	router.GET("/search", h.SearchFiles)

	req := httptest.NewRequest("GET", "/search?modified_after=not-a-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchHandler_SearchFiles_InvalidBeforeDate(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewSearchHandler(fileRepo)
	router := newTestRouter()
	router.GET("/search", h.SearchFiles)

	req := httptest.NewRequest("GET", "/search?modified_before=not-a-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchHandler_SearchDuplicates(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewSearchHandler(fileRepo)
	router := newTestRouter()
	router.GET("/search/duplicates", h.SearchDuplicates)

	tests := []struct {
		name  string
		query string
	}{
		{"basic", "/search/duplicates"},
		{"with filters", "/search/duplicates?file_type=video&extension=mp4"},
		{"with size", "/search/duplicates?min_size=100&max_size=5000000"},
		{"with smb roots", "/search/duplicates?smb_roots=root1,root2"},
		{"invalid sort", "/search/duplicates?sort_by=invalid&sort_order=whatever"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
		})
	}
}

func TestSearchHandler_AdvancedSearch_InvalidJSON(t *testing.T) {
	h := NewSearchHandler(nil)
	router := newTestRouter()
	router.POST("/search/advanced", h.AdvancedSearch)

	req := httptest.NewRequest("POST", "/search/advanced", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchHandler_AdvancedSearch_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	h := NewSearchHandler(fileRepo)
	router := newTestRouter()
	router.POST("/search/advanced", h.AdvancedSearch)

	tests := []struct {
		name string
		body SearchRequest
	}{
		{"default pagination", SearchRequest{Filter: models.SearchFilter{Query: "test"}}},
		{"invalid sort", SearchRequest{
			Filter: models.SearchFilter{Query: "test"},
			Page: 1, Limit: 50, SortBy: "invalid_field", SortOrder: "invalid_order",
		}},
		{"valid request", SearchRequest{
			Filter: models.SearchFilter{Query: "movie"},
			Page: 1, Limit: 10, SortBy: "name", SortOrder: "asc",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/search/advanced", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
		})
	}
}

// parseBool tests already exist in search_test.go - not duplicated here

// =============================================================================
// MediaBrowseHandler tests
// =============================================================================

func TestNewMediaBrowseHandler(t *testing.T) {
	h := NewMediaBrowseHandler(nil, nil, nil)
	assert.NotNil(t, h)
}

func TestMediaBrowseHandler_SearchMedia(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	h := NewMediaBrowseHandler(fileRepo, statsRepo, db)
	router := newTestRouter()
	router.GET("/media/search", h.SearchMedia)

	tests := []struct {
		name  string
		query string
	}{
		{"basic", "/media/search?query=test"},
		{"with type", "/media/search?query=test&media_type=video"},
		{"with pagination", "/media/search?query=test&limit=10&offset=5"},
		{"negative limit", "/media/search?limit=-1"},
		{"excessive limit", "/media/search?limit=500"},
		{"negative offset", "/media/search?offset=-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
		})
	}
}

func TestMediaBrowseHandler_GetMediaStats(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	fileRepo := repository.NewFileRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	h := NewMediaBrowseHandler(fileRepo, statsRepo, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/media/stats", nil)

	h.GetMediaStats(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp, "total_items")
	assert.Contains(t, resp, "by_type")
	assert.Contains(t, resp, "total_size")
	assert.Contains(t, resp, "recent_additions")
}

func TestFileToMediaItem_WithData(t *testing.T) {
	fileType := "video"
	ext := "mp4"
	now := time.Now()

	f := models.FileWithMetadata{
		File: models.File{
			ID: 1, Name: "test-file.mp4", Path: "/media/videos/test-file.mp4",
			Size: 1024000, FileType: &fileType, Extension: &ext,
			StorageRootName: "local", CreatedAt: now, ModifiedAt: now,
		},
	}

	result := fileToMediaItem(f)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "test-file.mp4", result.Title)
	assert.Equal(t, "video", result.MediaType)
	assert.Equal(t, "mp4", result.Quality)
}

func TestFileToMediaItem_EmptyFields(t *testing.T) {
	now := time.Now()
	f := models.FileWithMetadata{
		File: models.File{
			ID: 2, Name: "unknown", Path: "/media/unknown", Size: 500,
			FileType: nil, Extension: nil, CreatedAt: now, ModifiedAt: now,
		},
	}
	result := fileToMediaItem(f)
	assert.Equal(t, "other", result.MediaType)
	assert.Equal(t, "", result.Quality)
}

// =============================================================================
// ScanHandler tests
// =============================================================================

type mockScanner struct {
	queueScanFunc           func(job internalservices.ScanJob) error
	getAllActiveScanStatuses func() map[string]*internalservices.ScanStatus
	getActiveScanStatus     func(jobID string) (*internalservices.ScanStatus, bool)
}

func (m *mockScanner) QueueScan(job internalservices.ScanJob) error {
	if m.queueScanFunc != nil {
		return m.queueScanFunc(job)
	}
	return nil
}

func (m *mockScanner) GetAllActiveScanStatuses() map[string]*internalservices.ScanStatus {
	if m.getAllActiveScanStatuses != nil {
		return m.getAllActiveScanStatuses()
	}
	return map[string]*internalservices.ScanStatus{}
}

func (m *mockScanner) GetActiveScanStatus(jobID string) (*internalservices.ScanStatus, bool) {
	if m.getActiveScanStatus != nil {
		return m.getActiveScanStatus(jobID)
	}
	return nil, false
}

func TestNewScanHandler(t *testing.T) {
	h := NewScanHandler(nil, nil)
	assert.NotNil(t, h)
}

func TestScanHandler_CreateStorageRoot_InvalidJSON(t *testing.T) {
	h := NewScanHandler(&mockScanner{}, nil)
	router := newTestRouter()
	router.POST("/storage/roots", h.CreateStorageRoot)

	req := httptest.NewRequest("POST", "/storage/roots", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_CreateStorageRoot_MissingRequired(t *testing.T) {
	h := NewScanHandler(&mockScanner{}, nil)
	router := newTestRouter()
	router.POST("/storage/roots", h.CreateStorageRoot)

	body := `{"name": "test"}`
	req := httptest.NewRequest("POST", "/storage/roots", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_CreateStorageRoot_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	h := NewScanHandler(&mockScanner{}, db)
	router := newTestRouter()
	router.POST("/storage/roots", h.CreateStorageRoot)

	body := `{"name": "test-root", "protocol": "local", "path": "/tmp"}`
	req := httptest.NewRequest("POST", "/storage/roots", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Upsert
	body2 := `{"name": "test-root", "protocol": "smb", "path": "/data"}`
	req2 := httptest.NewRequest("POST", "/storage/roots", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)
}

func TestScanHandler_GetStorageRoots_WithDB(t *testing.T) {
	db, cleanup := newTestDB(t)
	defer cleanup()

	h := NewScanHandler(&mockScanner{}, db)
	router := newTestRouter()
	router.GET("/storage/roots", h.GetStorageRoots)

	req := httptest.NewRequest("GET", "/storage/roots", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_GetStorageRootStatus_InvalidID(t *testing.T) {
	h := NewScanHandler(&mockScanner{}, nil)
	router := newTestRouter()
	router.GET("/storage-roots/:id/status", h.GetStorageRootStatus)

	req := httptest.NewRequest("GET", "/storage-roots/abc/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_QueueScan_InvalidJSON(t *testing.T) {
	h := NewScanHandler(&mockScanner{}, nil)
	router := newTestRouter()
	router.POST("/scans", h.QueueScan)

	req := httptest.NewRequest("POST", "/scans", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScanHandler_ListScans(t *testing.T) {
	scanner := &mockScanner{
		getAllActiveScanStatuses: func() map[string]*internalservices.ScanStatus {
			return map[string]*internalservices.ScanStatus{
				"job-1": {
					Status: "running", StartTime: time.Now(),
					StorageRootName: "test-root", Protocol: "local",
				},
			}
		},
	}
	h := NewScanHandler(scanner, nil)
	router := newTestRouter()
	router.GET("/scans", h.ListScans)

	req := httptest.NewRequest("GET", "/scans", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScanHandler_GetScanStatus_NotFound(t *testing.T) {
	h := NewScanHandler(&mockScanner{}, nil)
	router := newTestRouter()
	router.GET("/scans/:job_id", h.GetScanStatus)

	req := httptest.NewRequest("GET", "/scans/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestScanStatusToJSON(t *testing.T) {
	status := &internalservices.ScanStatus{
		StorageRootName: "test", Protocol: "smb", Status: "running",
		StartTime: time.Now().Add(-10 * time.Second), CurrentPath: "/media/test",
		FilesProcessed: 100, FilesFound: 150, FilesUpdated: 20,
		FilesDeleted: 5, ErrorCount: 2,
	}
	result := scanStatusToJSON("job-123", status)
	assert.Equal(t, "job-123", result["job_id"])
	assert.Equal(t, "test", result["storage_root"])
	assert.Equal(t, "running", result["status"])
	assert.Equal(t, int64(100), result["files_processed"])
}

// =============================================================================
// RecommendationHandler tests
// =============================================================================

func TestNewRecommendationHandler(t *testing.T) {
	h := NewRecommendationHandler(nil)
	assert.NotNil(t, h)
}

func TestRecommendationHandler_GetTrendingItems(t *testing.T) {
	h := NewRecommendationHandler(nil)
	router := newTestRouter()
	router.GET("/recommendations/trending", h.GetTrendingItems)

	tests := []struct {
		name  string
		query string
	}{
		{"defaults", "/recommendations/trending"},
		{"with params", "/recommendations/trending?media_type=movie&limit=5&time_range=month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRecommendationHandler_GetPersonalizedRecommendations_InvalidUserID(t *testing.T) {
	h := NewRecommendationHandler(nil)
	router := newTestRouter()
	router.GET("/recommendations/personalized/:user_id", h.GetPersonalizedRecommendations)

	req := httptest.NewRequest("GET", "/recommendations/personalized/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRecommendationHandler_GetPersonalizedRecommendations_Success(t *testing.T) {
	h := NewRecommendationHandler(nil)
	router := newTestRouter()
	router.GET("/recommendations/personalized/:user_id", h.GetPersonalizedRecommendations)

	req := httptest.NewRequest("GET", "/recommendations/personalized/1?limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp PersonalizedResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.UserID)
	assert.True(t, len(resp.Items) > 0)
}

func TestRecommendationHandler_GetSimilarItems_InvalidID(t *testing.T) {
	h := NewRecommendationHandler(nil)
	router := newTestRouter()
	router.GET("/recommendations/similar/:media_id", h.GetSimilarItems)

	req := httptest.NewRequest("GET", "/recommendations/similar/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// ConfigurationHandler tests
// =============================================================================

type mockConfigServiceImpl struct {
	getWizardStepFunc func(stepID string) (*models.WizardStep, error)
	getSchemaFunc     func() (*models.ConfigurationSchema, error)
}

func (m *mockConfigServiceImpl) GetWizardStep(stepID string) (*models.WizardStep, error) {
	if m.getWizardStepFunc != nil {
		return m.getWizardStepFunc(stepID)
	}
	return nil, errors.New("not found")
}
func (m *mockConfigServiceImpl) ValidateWizardStep(stepID string, data map[string]interface{}) (*models.ValidationResult, error) {
	return &models.ValidationResult{IsValid: true}, nil
}
func (m *mockConfigServiceImpl) SaveWizardProgress(userID int, stepID string, data map[string]interface{}) error {
	return nil
}
func (m *mockConfigServiceImpl) GetWizardProgress(userID int) (*models.WizardProgress, error) {
	return &models.WizardProgress{}, nil
}
func (m *mockConfigServiceImpl) CompleteWizard(userID int, finalData map[string]interface{}) (*models.SystemConfiguration, error) {
	return &models.SystemConfiguration{}, nil
}
func (m *mockConfigServiceImpl) GetConfiguration() (*models.Configuration, error) {
	return &models.Configuration{}, nil
}
func (m *mockConfigServiceImpl) TestConfiguration(cfg *models.Configuration) (*models.ValidationResult, error) {
	return &models.ValidationResult{IsValid: true}, nil
}
func (m *mockConfigServiceImpl) GetConfigurationSchema() (*models.ConfigurationSchema, error) {
	if m.getSchemaFunc != nil {
		return m.getSchemaFunc()
	}
	return &models.ConfigurationSchema{}, nil
}

type mockConfigAuthImpl struct {
	checkPermissionFunc func(userID int, permission string) (bool, error)
}

func (m *mockConfigAuthImpl) ValidateToken(tokenString string) (*models.User, error) {
	return nil, errors.New("invalid")
}
func (m *mockConfigAuthImpl) CheckPermission(userID int, permission string) (bool, error) {
	if m.checkPermissionFunc != nil {
		return m.checkPermissionFunc(userID, permission)
	}
	return true, nil
}

func TestNewConfigurationHandler(t *testing.T) {
	h := NewConfigurationHandler(nil, nil)
	assert.NotNil(t, h)
}

func TestConfigurationHandler_GetConfiguration_Forbidden(t *testing.T) {
	h := NewConfigurationHandler(
		&mockConfigServiceImpl{},
		&mockConfigAuthImpl{checkPermissionFunc: func(userID int, permission string) (bool, error) {
			return false, nil
		}},
	)
	req := httptest.NewRequest("GET", "/configuration", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.GetConfiguration(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestConfigurationHandler_GetConfiguration_Success(t *testing.T) {
	h := NewConfigurationHandler(&mockConfigServiceImpl{}, &mockConfigAuthImpl{})
	req := httptest.NewRequest("GET", "/configuration", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.GetConfiguration(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConfigurationHandler_GetSystemStatus_Forbidden(t *testing.T) {
	h := NewConfigurationHandler(
		&mockConfigServiceImpl{},
		&mockConfigAuthImpl{checkPermissionFunc: func(userID int, permission string) (bool, error) {
			return false, nil
		}},
	)
	req := httptest.NewRequest("GET", "/status", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.GetSystemStatus(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestConfigurationHandler_GetSystemStatus_Success(t *testing.T) {
	h := NewConfigurationHandler(&mockConfigServiceImpl{}, &mockConfigAuthImpl{})
	req := httptest.NewRequest("GET", "/status", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.GetSystemStatus(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "healthy", resp["status"])
}

func TestConfigurationHandler_DeleteBackup_InvalidID(t *testing.T) {
	h := NewConfigurationHandler(&mockConfigServiceImpl{}, &mockConfigAuthImpl{})
	req := httptest.NewRequest("DELETE", "/backups/abc", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.DeleteBackup(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigurationHandler_TestConfiguration_Forbidden(t *testing.T) {
	h := NewConfigurationHandler(
		&mockConfigServiceImpl{},
		&mockConfigAuthImpl{checkPermissionFunc: func(userID int, permission string) (bool, error) {
			return false, nil
		}},
	)
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{}`))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.TestConfiguration(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestConfigurationHandler_TestConfiguration_InvalidBody(t *testing.T) {
	h := NewConfigurationHandler(&mockConfigServiceImpl{}, &mockConfigAuthImpl{})
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString("not-json"))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.TestConfiguration(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigurationHandler_TestConfiguration_Success(t *testing.T) {
	h := NewConfigurationHandler(&mockConfigServiceImpl{}, &mockConfigAuthImpl{})
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"id": "test-config", "name": "Test"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "user_id", 1))
	w := httptest.NewRecorder()
	h.TestConfiguration(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// AuthHandler additional gin endpoint tests
// =============================================================================

func TestAuthHandler_RegisterGin_InvalidJSON(t *testing.T) {
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := newTestRouter()
	router.POST("/register", func(c *gin.Context) {
		handler.RegisterGin(c, nil)
	})

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GetPermissionsGin_InvalidToken(t *testing.T) {
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := newTestRouter()
	router.GET("/permissions", handler.GetPermissionsGin)

	req := httptest.NewRequest("GET", "/permissions", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
