package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewAnalyticsHandler(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	assert.NotNil(t, handler)
}

func TestNewReportingHandler(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.ReportingService{}
	handler := NewReportingHandler(svc, logger)

	assert.NotNil(t, handler)
}

func TestNewFavoritesHandler(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.FavoritesService{}
	handler := NewFavoritesHandler(svc, logger)

	assert.NotNil(t, handler)
}

func TestAnalyticsHandler_LogMediaAccess_InvalidBody(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	router := setupTestRouter()
	router.POST("/analytics/access", handler.LogMediaAccess)

	req := httptest.NewRequest(http.MethodPost, "/analytics/access", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_LogEvent_InvalidBody(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	router := setupTestRouter()
	router.POST("/analytics/event", handler.LogEvent)

	req := httptest.NewRequest(http.MethodPost, "/analytics/event", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_GetUserAnalytics_InvalidUserID(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	router := setupTestRouter()
	router.GET("/analytics/user/:user_id", handler.GetUserAnalytics)

	req := httptest.NewRequest(http.MethodGet, "/analytics/user/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_GetMediaAnalytics_InvalidMediaID(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	router := setupTestRouter()
	router.GET("/analytics/media/:media_id", handler.GetMediaAnalytics)

	req := httptest.NewRequest(http.MethodGet, "/analytics/media/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_CreateReport_InvalidBody(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	router := setupTestRouter()
	router.POST("/analytics/reports", handler.CreateReport)

	req := httptest.NewRequest(http.MethodPost, "/analytics/reports", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_CreateReport_MissingReportType(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.AnalyticsService{}
	handler := NewAnalyticsHandler(svc, logger)

	router := setupTestRouter()
	router.POST("/analytics/reports", handler.CreateReport)

	body := map[string]interface{}{
		"params": map[string]interface{}{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/analytics/reports", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReportingHandler_GetUsageReport_InvalidDateFormat(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.ReportingService{}
	handler := NewReportingHandler(svc, logger)

	router := setupTestRouter()
	router.GET("/reports/usage", handler.GetUsageReport)

	req := httptest.NewRequest(http.MethodGet, "/reports/usage?start_date=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFavoritesHandler_ListFavorites_Unauthorized(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.FavoritesService{}
	handler := NewFavoritesHandler(svc, logger)

	router := setupTestRouter()
	router.GET("/favorites", handler.ListFavorites)

	req := httptest.NewRequest(http.MethodGet, "/favorites", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFavoritesHandler_AddFavorite_Unauthorized(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.FavoritesService{}
	handler := NewFavoritesHandler(svc, logger)

	router := setupTestRouter()
	router.POST("/favorites", handler.AddFavorite)

	body := map[string]interface{}{
		"entity_id":   123,
		"entity_type": "movie",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/favorites", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFavoritesHandler_AddFavorite_InvalidBody(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.FavoritesService{}
	handler := NewFavoritesHandler(svc, logger)

	router := setupTestRouter()
	router.POST("/favorites", func(c *gin.Context) {
		c.Set("user_id", 1)
		handler.AddFavorite(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/favorites", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFavoritesHandler_RemoveFavorite_InvalidEntityID(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.FavoritesService{}
	handler := NewFavoritesHandler(svc, logger)

	router := setupTestRouter()
	router.DELETE("/favorites/:entity_type/:entity_id", func(c *gin.Context) {
		c.Set("user_id", 1)
		handler.RemoveFavorite(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/favorites/movie/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFavoritesHandler_CheckFavorite_InvalidEntityID(t *testing.T) {
	logger := zap.NewNop()
	svc := &services.FavoritesService{}
	handler := NewFavoritesHandler(svc, logger)

	router := setupTestRouter()
	router.GET("/favorites/check/:entity_type/:entity_id", func(c *gin.Context) {
		c.Set("user_id", 1)
		handler.CheckFavorite(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/favorites/check/movie/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{"valid date", "2024-01-15", false},
		{"invalid date", "invalid", true},
		{"wrong format", "01/15/2024", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse("2006-01-02", tt.dateStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMediaAccessLog_Fields(t *testing.T) {
	access := models.MediaAccessLog{
		UserID:     1,
		MediaID:    100,
		Action:     "play",
		AccessTime: time.Now(),
	}

	assert.Equal(t, 1, access.UserID)
	assert.Equal(t, 100, access.MediaID)
	assert.Equal(t, "play", access.Action)
	assert.False(t, access.AccessTime.IsZero())
}

func TestFavorite_Struct(t *testing.T) {
	fav := models.Favorite{
		UserID:     1,
		EntityID:   100,
		EntityType: "movie",
	}

	assert.Equal(t, 1, fav.UserID)
	assert.Equal(t, 100, fav.EntityID)
	assert.Equal(t, "movie", fav.EntityType)
}

func TestAnalyticsEvent_Struct(t *testing.T) {
	event := models.AnalyticsEvent{
		UserID:    1,
		EventType: "page_view",
		Timestamp: time.Now(),
	}

	assert.Equal(t, 1, event.UserID)
	assert.Equal(t, "page_view", event.EventType)
	assert.False(t, event.Timestamp.IsZero())
}
