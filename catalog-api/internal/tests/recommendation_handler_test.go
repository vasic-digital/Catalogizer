package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"catalogizer/internal/handlers"
	"catalogizer/internal/services"
	"catalogizer/repository"
	_ "github.com/mattn/go-sqlite3"
)

func setupServices(t *testing.T, db *sql.DB) (*services.RecommendationService, *services.DeepLinkingService, *repository.FileRepository) {
	logger := zaptest.NewLogger(t)
	
	// Create services with minimal dependencies
	cacheService := services.NewCacheService(db, logger)
	translationService := services.NewTranslationService(logger)
	
	// Create services
	recognitionService := services.NewMediaRecognitionService(
		db,
		logger,
		cacheService,
		translationService,
		"http://mock-movie-api.com",
		"http://mock-music-api.com",
		"http://mock-book-api.com",
		"http://mock-game-api.com",
		"http://mock-ocr-api.com",
		"http://mock-fingerprint-api.com",
	)
	
	duplicateService := services.NewDuplicateDetectionService(db, logger, cacheService)
	recommendationService := services.NewRecommendationService(recognitionService, duplicateService)
	deepLinkingService := services.NewDeepLinkingService("https://test.catalogizer.app", "v1")
	
	// Create file repository directly without DB wrapper
	fileRepo := &repository.FileRepository{}
	
	return recommendationService, deepLinkingService, fileRepo
}

func TestRecommendationHandler_GetSimilarItems(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	
	recommendationService, deepLinkingService, fileRepo := setupServices(t, db)
	handler := handlers.NewRecommendationHandler(recommendationService, deepLinkingService, fileRepo)
	
	tests := []struct {
		name           string
		mediaID        string
		expectedStatus int
	}{
		{
			name:           "missing media ID",
			mediaID:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid request format",
			mediaID:        "123",
			expectedStatus: http.StatusNotFound, // Will fail due to no database setup
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/api/v1/media/"+tt.mediaID+"/similar", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.mediaID})
			
			// Create response recorder
			rr := httptest.NewRecorder()
			
			// Call handler
			handler.GetSimilarItems(rr, req)
			
			// Check status
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRecommendationHandler_PostSimilarItems(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	
	recommendationService, deepLinkingService, fileRepo := setupServices(t, db)
	handler := handlers.NewRecommendationHandler(recommendationService, deepLinkingService, fileRepo)
	
	t.Run("valid POST request", func(t *testing.T) {
		// Test request body
		requestBody := map[string]interface{}{
			"media_id": "123",
			"filters": map[string]interface{}{
				"genre": "Action",
			},
			"max_local_items":  5,
			"max_external":     3,
			"similarity_threshold": 0.5,
		}
		
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Create request
		req := httptest.NewRequest("POST", "/api/v1/media/similar", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Call handler
		handler.PostSimilarItems(rr, req)
		
		// Check status - should work even without actual data
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// Check that response has expected structure
		assert.Contains(t, response, "local_items")
		assert.Contains(t, response, "external_items")
		assert.Contains(t, response, "generated_at")
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Create request with invalid JSON
		req := httptest.NewRequest("POST", "/api/v1/media/similar", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Call handler
		handler.PostSimilarItems(rr, req)
		
		// Check status
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
	
	t.Run("missing media_id and metadata", func(t *testing.T) {
		// Test request body without media_id or metadata
		requestBody := map[string]interface{}{
			"filters": map[string]interface{}{
				"genre": "Action",
			},
		}
		
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Create request
		req := httptest.NewRequest("POST", "/api/v1/media/similar", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Call handler
		handler.PostSimilarItems(rr, req)
		
		// Check status
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestRecommendationHandler_GenerateDeepLinks(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	
	recommendationService, deepLinkingService, fileRepo := setupServices(t, db)
	handler := handlers.NewRecommendationHandler(recommendationService, deepLinkingService, fileRepo)
	
	t.Run("generate deep links with media_id", func(t *testing.T) {
		// Test request body
		requestBody := map[string]interface{}{
			"media_id": "123",
			"action":   "detail",
			"platforms": []string{"web", "android", "ios"},
		}
		
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Create request
		req := httptest.NewRequest("POST", "/api/v1/links/generate", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Call handler
		handler.GenerateDeepLinks(rr, req)
		
		// Check status
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// Check that response has expected structure
		assert.Contains(t, response, "links")
		assert.Contains(t, response, "universal_link")
		assert.Contains(t, response, "tracking_id")
	})
	
	t.Run("generate deep links with metadata", func(t *testing.T) {
		// Test request body with metadata instead of media_id
		requestBody := map[string]interface{}{
			"media_metadata": map[string]interface{}{
				"title": "Test Movie",
				"year": 2023,
				"type": "video",
			},
			"action": "play",
		}
		
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)
		
		// Create request
		req := httptest.NewRequest("POST", "/api/v1/links/generate", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Call handler
		handler.GenerateDeepLinks(rr, req)
		
		// Check status
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Parse response
		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// Check that response has expected structure
		assert.Contains(t, response, "links")
		assert.Contains(t, response, "universal_link")
	})
	
	t.Run("invalid request body", func(t *testing.T) {
		// Create request with invalid JSON
		req := httptest.NewRequest("POST", "/api/v1/links/generate", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		rr := httptest.NewRecorder()
		
		// Call handler
		handler.GenerateDeepLinks(rr, req)
		
		// Check status
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}