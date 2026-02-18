package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"catalogizer/config"
	"catalogizer/database"
	"catalogizer/internal/media/models"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEntityTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "entity_handler_test_*.db")
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

func setupEntityHandler(t *testing.T, db *database.DB) (*MediaEntityHandler, *repository.MediaItemRepository) {
	t.Helper()
	itemRepo := repository.NewMediaItemRepository(db)
	fileRepo := repository.NewMediaFileRepository(db)
	extMetaRepo := repository.NewExternalMetadataRepository(db)
	userMetaRepo := repository.NewUserMetadataRepository(db)

	handler := NewMediaEntityHandler(itemRepo, fileRepo, extMetaRepo, userMetaRepo)
	return handler, itemRepo
}

func TestMediaEntityHandler_GetEntityTypes(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, _ := setupEntityHandler(t, db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/types", nil)

	handler.GetEntityTypes(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	types, ok := resp["types"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(types), 11) // 11 seeded types
}

func TestMediaEntityHandler_GetEntityStats(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, err := itemRepo.GetMediaTypeByName(ctx, "movie")
	require.NoError(t, err)

	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Movie 1", Status: "detected"})
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Movie 2", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/stats", nil)

	handler.GetEntityStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["total_entities"])
}

func TestMediaEntityHandler_ListEntities(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "The Matrix", Status: "detected"})
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Inception", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities?query=Matrix", nil)

	handler.ListEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["total"])
}

func TestMediaEntityHandler_GetEntity(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	year := 1999
	id, _ := itemRepo.Create(ctx, &models.MediaItem{
		MediaTypeID: typeID,
		Title:       "The Matrix",
		Year:        &year,
		Status:      "detected",
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	_ = id

	handler.GetEntity(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "The Matrix", resp["title"])
	assert.Equal(t, "movie", resp["media_type"])
}
