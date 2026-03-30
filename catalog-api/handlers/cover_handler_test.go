package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"catalogizer/internal/media/models"
	"catalogizer/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupCoverHandler() *CoverHandler {
	logger, _ := zap.NewNop(), error(nil)
	coverArtService := services.NewCoverArtService(nil, logger)
	return NewCoverHandler(coverArtService)
}

func TestCoverHandler_ServePlaceholder_Movie(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/placeholder/movie", nil)
	c.Params = gin.Params{{Key: "type", Value: "movie"}}

	handler.ServePlaceholder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "<svg")
	assert.Contains(t, w.Body.String(), "#E53E3E") // movie color
	assert.Contains(t, w.Body.String(), "Movie")
}

func TestCoverHandler_ServePlaceholder_AllTypes(t *testing.T) {
	handler := setupCoverHandler()

	types := []string{
		"movie", "tv_show", "tv_season", "tv_episode",
		"music_artist", "music_album", "song",
		"game", "software", "book", "comic",
	}

	for _, mediaType := range types {
		t.Run(mediaType, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/v1/cover/placeholder/"+mediaType, nil)
			c.Params = gin.Params{{Key: "type", Value: mediaType}}

			handler.ServePlaceholder(c)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
			assert.Contains(t, w.Body.String(), "<svg")
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestCoverHandler_ServePlaceholder_UnknownType(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/placeholder/unknown", nil)
	c.Params = gin.Params{{Key: "type", Value: "unknown"}}

	handler.ServePlaceholder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "<svg")
	assert.Contains(t, w.Body.String(), "#718096") // gray fallback
}

func TestCoverHandler_ServePlaceholder_EmptyType(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/placeholder/", nil)
	c.Params = gin.Params{{Key: "type", Value: ""}}

	handler.ServePlaceholder(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
	// Empty type defaults to "movie"
	assert.Contains(t, w.Body.String(), "<svg")
}

func TestCoverHandler_ServeCover_NoDB_FallsBackToPlaceholder(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.ServeCover(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/svg+xml", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "<svg")
}

func TestCoverHandler_ServeCover_InvalidID(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	handler.ServeCover(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoverHandler_GetCoverURL_NoDB(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/url/42?type=movie", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	handler.GetCoverURL(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(42), resp["media_item_id"])
	assert.Equal(t, "/api/v1/cover/placeholder/movie", resp["cover_url"])
}

func TestCoverHandler_GetCoverURL_InvalidID(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/url/notanumber", nil)
	c.Params = gin.Params{{Key: "id", Value: "notanumber"}}

	handler.GetCoverURL(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCoverHandler_GetCoverURL_DefaultType(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No type query param — should default to "movie"
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/url/10", nil)
	c.Params = gin.Params{{Key: "id", Value: "10"}}

	handler.GetCoverURL(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/cover/placeholder/movie", resp["cover_url"])
}

func TestCoverHandler_PlaceholderCaching(t *testing.T) {
	handler := setupCoverHandler()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/placeholder/movie", nil)
	c.Params = gin.Params{{Key: "type", Value: "movie"}}

	handler.ServePlaceholder(c)

	// Verify Cache-Control header is set
	cacheControl := w.Header().Get("Cache-Control")
	assert.True(t, strings.Contains(cacheControl, "86400"), "expected 24h cache, got: %s", cacheControl)
}

func TestCoverHandler_WithCoverArtService_Integration(t *testing.T) {
	// Test that the full chain works: handler -> cover art service -> placeholder
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	logger, _ := zap.NewNop(), error(nil)
	coverArtService := services.NewCoverArtService(db, logger)
	handler := NewCoverHandler(coverArtService)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/cover/url/999?type=tv_show", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	handler.GetCoverURL(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// No external_metadata or cover_art for item 999, so should get placeholder
	assert.Equal(t, "/api/v1/cover/placeholder/tv_show", resp["cover_url"])
}

func TestMediaEntityHandler_WithCoverService_ListEntities(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	// Wire up cover art service
	logger, _ := zap.NewNop(), error(nil)
	coverArtService := services.NewCoverArtService(db, logger)
	handler.SetCoverArtService(coverArtService)

	// Create a movie entity
	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "movie")
	_, _ = itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Test Movie", Status: "detected"})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/entities?query=Test", nil)

	handler.ListEntities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	items := resp["items"].([]interface{})
	require.Len(t, items, 1)

	item := items[0].(map[string]interface{})
	coverURL, ok := item["cover_url"]
	assert.True(t, ok, "cover_url field should be present")
	assert.NotEmpty(t, coverURL, "cover_url should not be empty")
	// Should be a placeholder since no metadata exists
	assert.Contains(t, coverURL, "/api/v1/cover/placeholder/")
}

func TestMediaEntityHandler_WithCoverService_GetEntity(t *testing.T) {
	db, cleanup := setupEntityTestDB(t)
	defer cleanup()

	handler, itemRepo := setupEntityHandler(t, db)
	ctx := context.Background()

	logger, _ := zap.NewNop(), error(nil)
	coverArtService := services.NewCoverArtService(db, logger)
	handler.SetCoverArtService(coverArtService)

	_, typeID, _ := itemRepo.GetMediaTypeByName(ctx, "tv_show")
	id, _ := itemRepo.Create(ctx, &models.MediaItem{MediaTypeID: typeID, Title: "Test Show", Status: "detected"})

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

	coverURL, ok := resp["cover_url"]
	assert.True(t, ok, "cover_url field should be present in GetEntity response")
	assert.Contains(t, coverURL, "/api/v1/cover/placeholder/tv_show")
}
