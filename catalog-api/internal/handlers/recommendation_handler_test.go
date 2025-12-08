package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"catalogizer/internal/services"
	"catalogizer/models"
	"catalogizer/repository"
)

func TestNewRecommendationHandler(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}

	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	assert.NotNil(t, handler)
	assert.Equal(t, recommendationService, handler.recommendationService)
	assert.Equal(t, deepLinkingService, handler.deepLinkingService)
	assert.Equal(t, fileRepository, handler.fileRepository)
}

func TestRecommendationHandler_GetSimilarItems_InvalidMediaID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/media/123/similar", nil)

	handler.GetSimilarItems(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Media ID is required")
}

func TestRecommendationHandler_PostSimilarItems_InvalidJSON(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/media/similar", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.PostSimilarItems(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestRecommendationHandler_PostSimilarItems_MissingMediaID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/media/similar", bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")

	handler.PostSimilarItems(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Either media_id or media_metadata is required")
}

func TestRecommendationHandler_GenerateDeepLinks_InvalidJSON(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/generate", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.GenerateDeepLinks(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestRecommendationHandler_GenerateDeepLinks_MissingMediaID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/generate", bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")

	handler.GenerateDeepLinks(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Either media_id or media_metadata is required")
}

func TestRecommendationHandler_GetMediaWithSimilarItems_InvalidMediaID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/media/123/detail-with-similar", nil)

	handler.GetMediaWithSimilarItems(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Media ID is required")
}

func TestRecommendationHandler_TrackLinkClick_InvalidJSON(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/track", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.TrackLinkClick(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestRecommendationHandler_TrackLinkClick_MissingTrackingID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/track", bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")

	handler.TrackLinkClick(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Tracking ID is required")
}

func TestRecommendationHandler_GetLinkAnalytics_MissingTrackingID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/links/123/analytics", nil)

	handler.GetLinkAnalytics(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Tracking ID is required")
}

func TestRecommendationHandler_BatchGenerateLinks_InvalidJSON(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/batch", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.BatchGenerateLinks(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestRecommendationHandler_BatchGenerateLinks_EmptyRequests(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/batch", bytes.NewBufferString(`[]`))
	r.Header.Set("Content-Type", "application/json")

	handler.BatchGenerateLinks(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "At least one request is required")
}

func TestRecommendationHandler_BatchGenerateLinks_TooManyRequests(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	// Create 51 requests (over the limit)
	requests := make([]*services.DeepLinkRequest, 51)
	for i := range requests {
		requests[i] = &services.DeepLinkRequest{
			MediaID: "test",
			Action:  "detail",
		}
	}

	requestBody, _ := json.Marshal(requests)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/batch", bytes.NewBuffer(requestBody))
	r.Header.Set("Content-Type", "application/json")

	handler.BatchGenerateLinks(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Too many requests")
}

func TestRecommendationHandler_GenerateSmartLink_InvalidJSON(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/smart", bytes.NewBufferString("invalid json"))
	r.Header.Set("Content-Type", "application/json")

	handler.GenerateSmartLink(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestRecommendationHandler_GenerateSmartLink_MissingMediaID(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/links/smart", bytes.NewBufferString(`{}`))
	r.Header.Set("Content-Type", "application/json")

	handler.GenerateSmartLink(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Either media_id or media_metadata is required")
}

func TestRecommendationHandler_GetRecommendationTrends(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/recommendations/trends?media_type=movie&period=week&limit=10", nil)

	handler.GetRecommendationTrends(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RecommendationTrends
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "week", response.Period)
	assert.Equal(t, "movie", response.MediaType)
	assert.NotEmpty(t, response.Items)
}

// Helper function tests
func TestExtractLinkContext(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Referer", "https://example.com")
	r.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)")
	r.Header.Set("X-User-ID", "123")
	r.Header.Set("X-Device-ID", "device123")
	r.Header.Set("X-Session-ID", "session123")
	r.URL.RawQuery = "utm_source=google&utm_medium=email&utm_campaign=summer"

	context := extractLinkContext(r)

	assert.Equal(t, "https://example.com", context.ReferrerPage)
	assert.Equal(t, "123", context.UserID)
	assert.Equal(t, "device123", context.DeviceID)
	assert.Equal(t, "session123", context.SessionID)
	assert.Equal(t, "ios", context.Platform)
	assert.NotNil(t, context.UTMParams)
	assert.Equal(t, "google", context.UTMParams.Source)
	assert.Equal(t, "email", context.UTMParams.Medium)
	assert.Equal(t, "summer", context.UTMParams.Campaign)
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1",
			},
			remote:   "127.0.0.1:1234",
			expected: "192.168.1.100",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.200",
			},
			remote:   "127.0.0.1:1234",
			expected: "192.168.1.200",
		},
		{
			name:     "fallback to RemoteAddr",
			headers:  map[string]string{},
			remote:   "192.168.1.50:8080",
			expected: "192.168.1.50:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test", nil)
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			r.RemoteAddr = tt.remote

			result := getClientIP(r)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMockTrendingItems(t *testing.T) {
	items := generateMockTrendingItems("video", 5)

	assert.Len(t, items, 5)
	assert.Equal(t, "Trending Movie 1", items[0].Title)
	assert.Equal(t, "Action/Adventure", items[0].Subtitle)
	assert.Greater(t, items[0].TrendScore, 0.0)
	assert.Greater(t, items[0].RecommendationCount, 0)
	assert.Greater(t, items[0].ViewCount, 0)
	assert.Greater(t, items[0].Rating, 0.0)

	// Check that scores decrease
	assert.Greater(t, items[0].TrendScore, items[1].TrendScore)
	assert.Greater(t, items[0].RecommendationCount, items[1].RecommendationCount)
}

func TestRecommendationHandler_convertFileToMediaMetadata(t *testing.T) {
	recommendationService := &services.RecommendationService{}
	deepLinkingService := &services.DeepLinkingService{}
	fileRepository := &repository.FileRepository{}
	handler := NewRecommendationHandler(recommendationService, deepLinkingService, fileRepository)

	// Test with nil file
	result := handler.convertFileToMediaMetadata(nil)
	assert.Nil(t, result)

	// Test with file that has metadata
	fileWithMetadata := &models.FileWithMetadata{
		File: models.File{
			ID:   1,
			Name: "test.mp4",
			Size: 1000000,
		},
		Metadata: []models.FileMetadata{
			{Key: "media_type", Value: "video"},
			{Key: "title", Value: "Test Movie"},
			{Key: "year", Value: "2023"},
			{Key: "rating", Value: "8.5"},
		},
	}

	result = handler.convertFileToMediaMetadata(fileWithMetadata)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Movie", result.Title)
	assert.Equal(t, "video", result.MediaType)
	assert.Equal(t, &[]int{2023}[0], result.Year)
	assert.Equal(t, &[]float64{8.5}[0], result.Rating)
}
