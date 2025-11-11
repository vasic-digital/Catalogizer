package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/internal/models"
	"catalogizer/internal/services"
	"catalogizer/repository"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRecommendationService is a mock implementation of RecommendationService
type MockRecommendationService struct {
	mock.Mock
}

func (m *MockRecommendationService) GetSimilarItems(ctx context.Context, req *services.SimilarItemsRequest) (*services.SimilarItemsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.SimilarItemsResponse), args.Error(1)
}

// MockDeepLinkingService is a mock implementation of DeepLinkingService
type MockDeepLinkingService struct {
	mock.Mock
}

func (m *MockDeepLinkingService) GenerateDeepLinks(ctx context.Context, req *services.DeepLinkRequest) (*services.DeepLinkResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.DeepLinkResponse), args.Error(1)
}

func (m *MockDeepLinkingService) GenerateSmartLink(ctx context.Context, req *services.DeepLinkRequest) (*services.SmartLinkResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.SmartLinkResponse), args.Error(1)
}

func (m *MockDeepLinkingService) GenerateBatchLinks(ctx context.Context, requests []*services.DeepLinkRequest) ([]*services.DeepLinkResponse, error) {
	args := m.Called(ctx, requests)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*services.DeepLinkResponse), args.Error(1)
}

func (m *MockDeepLinkingService) TrackLinkEvent(ctx context.Context, event *services.LinkTrackingEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockDeepLinkingService) GetLinkAnalytics(ctx context.Context, trackingID string) (*services.LinkAnalytics, error) {
	args := m.Called(ctx, trackingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.LinkAnalytics), args.Error(1)
}

// MockFileRepositoryForRec is a mock implementation of FileRepository for RecommendationHandler
type MockFileRepositoryForRec struct {
	mock.Mock
}

func (m *MockFileRepositoryForRec) GetFileByID(ctx context.Context, id int64) (*models.FileWithMetadata, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileWithMetadata), args.Error(1)
}

// Test setup helper

func setupRecommendationHandler() (*RecommendationHandler, *MockRecommendationService, *MockDeepLinkingService, *MockFileRepositoryForRec) {
	mockRecService := new(MockRecommendationService)
	mockDeepLinkService := new(MockDeepLinkingService)
	mockFileRepo := new(MockFileRepositoryForRec)

	handler := &RecommendationHandler{
		recommendationService: (*services.RecommendationService)(unsafe_cast_rec_service(mockRecService)),
		deepLinkingService:    (*services.DeepLinkingService)(unsafe_cast_deeplink_service(mockDeepLinkService)),
		fileRepository:        (*repository.FileRepository)(unsafe_cast_file_repo(mockFileRepo)),
	}

	return handler, mockRecService, mockDeepLinkService, mockFileRepo
}

func unsafe_cast_rec_service(m *MockRecommendationService) interface{} {
	return m
}

func unsafe_cast_deeplink_service(m *MockDeepLinkingService) interface{} {
	return m
}

func unsafe_cast_file_repo(m *MockFileRepositoryForRec) interface{} {
	return m
}

// GetSimilarItems Tests

func TestRecommendationHandler_GetSimilarItems_Success(t *testing.T) {
	handler, mockRecService, _, mockFileRepo := setupRecommendationHandler()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File: models.File{
			ID:         123,
			Name:       "Test Movie.mp4",
			Size:       1024000000,
			MimeType:   "video/mp4",
			CreatedAt:  now,
			ModifiedAt: now,
		},
		Metadata: []models.FileMetadata{
			{FileID: 123, Key: "title", Value: "Test Movie"},
			{FileID: 123, Key: "year", Value: float64(2024)},
		},
	}

	mockResponse := &services.SimilarItemsResponse{
		LocalItems:    []services.LocalSimilarItem{{MediaID: "456", Score: 0.85}},
		ExternalItems: []services.ExternalSimilarItem{{Title: "Similar Movie", Score: 0.75}},
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockRecService.On("GetSimilarItems", mock.Anything, mock.MatchedBy(func(req *services.SimilarItemsRequest) bool {
		return req.MediaID == "123" && req.MaxLocalItems == 10 && req.MaxExternalItems == 5
	})).Return(mockResponse, nil)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/media/{id}/similar", handler.GetSimilarItems)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/123/similar", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response services.SimilarItemsResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.LocalItems, 1)

	mockFileRepo.AssertExpectations(t)
	mockRecService.AssertExpectations(t)
}

func TestRecommendationHandler_GetSimilarItems_WithFilters(t *testing.T) {
	handler, mockRecService, _, mockFileRepo := setupRecommendationHandler()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File:     models.File{ID: 123, Name: "Test.mp4", CreatedAt: now, ModifiedAt: now},
		Metadata: []models.FileMetadata{},
	}

	mockResponse := &services.SimilarItemsResponse{
		LocalItems: []services.LocalSimilarItem{},
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockRecService.On("GetSimilarItems", mock.Anything, mock.MatchedBy(func(req *services.SimilarItemsRequest) bool {
		return req.Filters != nil &&
			req.Filters.GenreFilter != nil &&
			req.Filters.YearRange != nil &&
			req.Filters.RatingRange != nil &&
			req.Filters.ExcludeWatched == true &&
			req.SimilarityThreshold == 0.5
	})).Return(mockResponse, nil)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/media/{id}/similar", handler.GetSimilarItems)

	url := "/api/v1/media/123/similar?genre=Action&year_start=2020&year_end=2024&" +
		"min_rating=7.0&max_rating=10.0&language=en&exclude_watched=true&" +
		"min_confidence=0.8&similarity_threshold=0.5&max_local=20&max_external=10&include_external=true"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockFileRepo.AssertExpectations(t)
	mockRecService.AssertExpectations(t)
}

func TestRecommendationHandler_GetSimilarItems_InvalidMediaID(t *testing.T) {
	handler, _, _, mockFileRepo := setupRecommendationHandler()

	mockFileRepo.On("GetFileByID", mock.Anything, int64(999)).Return(nil, errors.New("file not found"))

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/media/{id}/similar", handler.GetSimilarItems)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/999/similar", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockFileRepo.AssertExpectations(t)
}

func TestRecommendationHandler_GetSimilarItems_ServiceError(t *testing.T) {
	handler, mockRecService, _, mockFileRepo := setupRecommendationHandler()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File:     models.File{ID: 123, Name: "Test.mp4", CreatedAt: now, ModifiedAt: now},
		Metadata: []models.FileMetadata{},
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockRecService.On("GetSimilarItems", mock.Anything, mock.Anything).Return(nil, errors.New("recommendation service unavailable"))

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/media/{id}/similar", handler.GetSimilarItems)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/123/similar", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockFileRepo.AssertExpectations(t)
	mockRecService.AssertExpectations(t)
}

// PostSimilarItems Tests

func TestRecommendationHandler_PostSimilarItems_Success(t *testing.T) {
	handler, mockRecService, _, _ := setupRecommendationHandler()

	mockResponse := &services.SimilarItemsResponse{
		LocalItems: []services.LocalSimilarItem{{MediaID: "456", Score: 0.85}},
	}

	mockRecService.On("GetSimilarItems", mock.Anything, mock.Anything).Return(mockResponse, nil)

	reqBody := services.SimilarItemsRequest{
		MediaID:         "123",
		MaxLocalItems:   10,
		MaxExternalItems: 5,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/similar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PostSimilarItems(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockRecService.AssertExpectations(t)
}

func TestRecommendationHandler_PostSimilarItems_InvalidBody(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/similar", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PostSimilarItems(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRecommendationHandler_PostSimilarItems_MissingRequiredFields(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	reqBody := services.SimilarItemsRequest{
		// Missing both MediaID and MediaMetadata
		MaxLocalItems: 10,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/similar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.PostSimilarItems(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Either media_id or media_metadata is required")
}

// GenerateDeepLinks Tests

func TestRecommendationHandler_GenerateDeepLinks_Success(t *testing.T) {
	handler, _, mockDeepLinkService, _ := setupRecommendationHandler()

	mockResponse := &services.DeepLinkResponse{
		TrackingID: "track123",
		Links: map[string]string{
			"android": "catalogizer://media/123",
			"ios":     "catalogizer://media/123",
			"web":     "https://catalogizer.com/media/123",
		},
	}

	mockDeepLinkService.On("GenerateDeepLinks", mock.Anything, mock.MatchedBy(func(req *services.DeepLinkRequest) bool {
		return req.MediaID == "123" && req.Action == "play"
	})).Return(mockResponse, nil)

	reqBody := services.DeepLinkRequest{
		MediaID: "123",
		Action:  "play",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GenerateDeepLinks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response services.DeepLinkResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "track123", response.TrackingID)

	mockDeepLinkService.AssertExpectations(t)
}

func TestRecommendationHandler_GenerateDeepLinks_DefaultAction(t *testing.T) {
	handler, _, mockDeepLinkService, _ := setupRecommendationHandler()

	mockResponse := &services.DeepLinkResponse{}

	mockDeepLinkService.On("GenerateDeepLinks", mock.Anything, mock.MatchedBy(func(req *services.DeepLinkRequest) bool {
		return req.Action == "detail" // Default action
	})).Return(mockResponse, nil)

	reqBody := services.DeepLinkRequest{
		MediaID: "123",
		// Action not specified
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GenerateDeepLinks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockDeepLinkService.AssertExpectations(t)
}

// GetMediaWithSimilarItems Tests

func TestRecommendationHandler_GetMediaWithSimilarItems_Success(t *testing.T) {
	handler, mockRecService, mockDeepLinkService, mockFileRepo := setupRecommendationHandler()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File:     models.File{ID: 123, Name: "Test.mp4", CreatedAt: now, ModifiedAt: now},
		Metadata: []models.FileMetadata{},
	}

	mockSimilarResponse := &services.SimilarItemsResponse{
		LocalItems: []services.LocalSimilarItem{{MediaID: "456", Score: 0.85}},
	}

	mockLinksResponse := &services.DeepLinkResponse{
		TrackingID: "track123",
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockRecService.On("GetSimilarItems", mock.Anything, mock.Anything).Return(mockSimilarResponse, nil)
	mockDeepLinkService.On("GenerateDeepLinks", mock.Anything, mock.Anything).Return(mockLinksResponse, nil)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/media/{id}/detail-with-similar", handler.GetMediaWithSimilarItems)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/media/123/detail-with-similar", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response MediaDetailWithSimilarResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "123", response.MediaID)
	assert.NotNil(t, response.SimilarItems)
	assert.NotNil(t, response.Links)

	mockFileRepo.AssertExpectations(t)
	mockRecService.AssertExpectations(t)
	mockDeepLinkService.AssertExpectations(t)
}

// TrackLinkClick Tests

func TestRecommendationHandler_TrackLinkClick_Success(t *testing.T) {
	handler, _, mockDeepLinkService, _ := setupRecommendationHandler()

	mockDeepLinkService.On("TrackLinkEvent", mock.Anything, mock.MatchedBy(func(event *services.LinkTrackingEvent) bool {
		return event.TrackingID == "track123" && event.UserAgent != "" && event.IPAddress != ""
	})).Return(nil)

	reqBody := services.LinkTrackingEvent{
		TrackingID: "track123",
		EventType:  "click",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/track", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	handler.TrackLinkClick(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	mockDeepLinkService.AssertExpectations(t)
}

func TestRecommendationHandler_TrackLinkClick_MissingTrackingID(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	reqBody := services.LinkTrackingEvent{
		EventType: "click",
		// TrackingID missing
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/track", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.TrackLinkClick(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Tracking ID is required")
}

// GetLinkAnalytics Tests

func TestRecommendationHandler_GetLinkAnalytics_Success(t *testing.T) {
	handler, _, mockDeepLinkService, _ := setupRecommendationHandler()

	mockAnalytics := &services.LinkAnalytics{
		TrackingID:  "track123",
		TotalClicks: 100,
		UniqueUsers: 75,
	}

	mockDeepLinkService.On("GetLinkAnalytics", mock.Anything, "track123").Return(mockAnalytics, nil)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/links/{tracking_id}/analytics", handler.GetLinkAnalytics)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/links/track123/analytics", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response services.LinkAnalytics
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "track123", response.TrackingID)
	assert.Equal(t, 100, response.TotalClicks)

	mockDeepLinkService.AssertExpectations(t)
}

// BatchGenerateLinks Tests

func TestRecommendationHandler_BatchGenerateLinks_Success(t *testing.T) {
	handler, _, mockDeepLinkService, _ := setupRecommendationHandler()

	mockResponses := []*services.DeepLinkResponse{
		{TrackingID: "track1"},
		{TrackingID: "track2"},
	}

	mockDeepLinkService.On("GenerateBatchLinks", mock.Anything, mock.MatchedBy(func(requests []*services.DeepLinkRequest) bool {
		return len(requests) == 2
	})).Return(mockResponses, nil)

	requests := []*services.DeepLinkRequest{
		{MediaID: "123", Action: "detail"},
		{MediaID: "456", Action: "play"},
	}

	body, _ := json.Marshal(requests)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.BatchGenerateLinks(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["processed"])
	assert.Equal(t, float64(2), response["requested"])

	mockDeepLinkService.AssertExpectations(t)
}

func TestRecommendationHandler_BatchGenerateLinks_EmptyRequest(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	requests := []*services.DeepLinkRequest{}

	body, _ := json.Marshal(requests)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.BatchGenerateLinks(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "At least one request is required")
}

func TestRecommendationHandler_BatchGenerateLinks_TooManyRequests(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	// Create 51 requests (exceeds limit of 50)
	requests := make([]*services.DeepLinkRequest, 51)
	for i := 0; i < 51; i++ {
		requests[i] = &services.DeepLinkRequest{MediaID: "123"}
	}

	body, _ := json.Marshal(requests)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.BatchGenerateLinks(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Too many requests")
}

// GenerateSmartLink Tests

func TestRecommendationHandler_GenerateSmartLink_Success(t *testing.T) {
	handler, _, mockDeepLinkService, _ := setupRecommendationHandler()

	mockResponse := &services.SmartLinkResponse{
		ShortURL:   "https://short.link/abc123",
		TrackingID: "track123",
	}

	mockDeepLinkService.On("GenerateSmartLink", mock.Anything, mock.Anything).Return(mockResponse, nil)

	reqBody := services.DeepLinkRequest{
		MediaID: "123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/links/smart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.GenerateSmartLink(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response services.SmartLinkResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "https://short.link/abc123", response.ShortURL)

	mockDeepLinkService.AssertExpectations(t)
}

// GetRecommendationTrends Tests

func TestRecommendationHandler_GetRecommendationTrends_Success(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/trends?media_type=video&period=week&limit=10", nil)
	rr := httptest.NewRecorder()

	handler.GetRecommendationTrends(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response RecommendationTrends
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "week", response.Period)
	assert.Equal(t, "video", response.MediaType)
	assert.Len(t, response.Items, 10)
}

func TestRecommendationHandler_GetRecommendationTrends_DefaultPeriod(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations/trends", nil)
	rr := httptest.NewRecorder()

	handler.GetRecommendationTrends(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response RecommendationTrends
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "week", response.Period) // Default period
}

// Helper Function Tests

func TestExtractLinkContext_AllHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?utm_source=email&utm_medium=newsletter&utm_campaign=spring2024", nil)
	req.Header.Set("X-User-ID", "user123")
	req.Header.Set("X-Device-ID", "device456")
	req.Header.Set("X-Session-ID", "session789")
	req.Header.Set("User-Agent", "Catalogizer-Android/1.0")
	req.Header.Set("Referer", "https://catalogizer.com/browse")

	context := extractLinkContext(req)

	assert.Equal(t, "user123", context.UserID)
	assert.Equal(t, "device456", context.DeviceID)
	assert.Equal(t, "session789", context.SessionID)
	assert.Equal(t, "android", context.Platform)
	assert.Equal(t, "https://catalogizer.com/browse", context.ReferrerPage)
	assert.NotNil(t, context.UTMParams)
	assert.Equal(t, "email", context.UTMParams.Source)
	assert.Equal(t, "newsletter", context.UTMParams.Medium)
	assert.Equal(t, "spring2024", context.UTMParams.Campaign)
}

func TestExtractLinkContext_PlatformDetection(t *testing.T) {
	tests := []struct {
		userAgent        string
		expectedPlatform string
	}{
		{"Mozilla/5.0 (Android 10)", "android"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0)", "ios"},
		{"Mozilla/5.0 (iPad; CPU OS 14_0)", "ios"},
		{"Catalogizer-Desktop/1.0", "desktop"},
		{"Mozilla/5.0 (Windows NT 10.0)", "web"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedPlatform, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("User-Agent", tt.userAgent)

			context := extractLinkContext(req)

			assert.Equal(t, tt.expectedPlatform, context.Platform)
		})
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1, 172.16.0.1")

	ip := getClientIP(req)

	assert.Equal(t, "192.168.1.1", ip) // Should return first IP
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.2")

	ip := getClientIP(req)

	assert.Equal(t, "192.168.1.2", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.3:12345"

	ip := getClientIP(req)

	assert.Equal(t, "192.168.1.3:12345", ip) // Fallback to RemoteAddr
}

func TestConvertFileToMediaMetadata_CompleteMetadata(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	now := time.Now()
	fileWithMetadata := &models.FileWithMetadata{
		File: models.File{
			ID:         123,
			Name:       "Original Name.mp4",
			Size:       1024000000,
			MimeType:   "video/mp4",
			CreatedAt:  now,
			ModifiedAt: now,
		},
		Metadata: []models.FileMetadata{
			{FileID: 123, Key: "title", Value: "Test Movie"},
			{FileID: 123, Key: "description", Value: "A great movie"},
			{FileID: 123, Key: "genre", Value: "Action"},
			{FileID: 123, Key: "year", Value: float64(2024)},
			{FileID: 123, Key: "rating", Value: 8.5},
			{FileID: 123, Key: "duration", Value: float64(120)},
			{FileID: 123, Key: "language", Value: "en"},
			{FileID: 123, Key: "country", Value: "US"},
			{FileID: 123, Key: "director", Value: "John Doe"},
			{FileID: 123, Key: "media_type", Value: "movie"},
		},
	}

	metadata := handler.convertFileToMediaMetadata(fileWithMetadata)

	assert.NotNil(t, metadata)
	assert.Equal(t, int64(123), metadata.ID)
	assert.Equal(t, "Test Movie", metadata.Title) // Should use metadata title, not file name
	assert.Equal(t, "A great movie", metadata.Description)
	assert.Equal(t, "Action", metadata.Genre)
	assert.Equal(t, 2024, *metadata.Year)
	assert.Equal(t, 8.5, *metadata.Rating)
	assert.Equal(t, 120, *metadata.Duration)
	assert.Equal(t, "en", metadata.Language)
	assert.Equal(t, "US", metadata.Country)
	assert.Equal(t, "John Doe", metadata.Director)
	assert.Equal(t, "movie", metadata.MediaType)
}

func TestConvertFileToMediaMetadata_InferMediaType(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	tests := []struct {
		mimeType         string
		expectedMediaType string
	}{
		{"video/mp4", "movie"},
		{"audio/mp3", "music"},
		{"application/pdf", "ebook"},
		{"application/epub+zip", "ebook"},
		{"text/plain", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			now := time.Now()
			fileWithMetadata := &models.FileWithMetadata{
				File: models.File{
					ID:         123,
					Name:       "test.file",
					MimeType:   tt.mimeType,
					CreatedAt:  now,
					ModifiedAt: now,
				},
				Metadata: []models.FileMetadata{},
			}

			metadata := handler.convertFileToMediaMetadata(fileWithMetadata)

			assert.Equal(t, tt.expectedMediaType, metadata.MediaType)
		})
	}
}

func TestConvertFileToMediaMetadata_NilInput(t *testing.T) {
	handler, _, _, _ := setupRecommendationHandler()

	metadata := handler.convertFileToMediaMetadata(nil)

	assert.Nil(t, metadata)
}

// Benchmark Tests

func BenchmarkRecommendationHandler_GetSimilarItems(b *testing.B) {
	handler, mockRecService, _, mockFileRepo := setupRecommendationHandler()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File:     models.File{ID: 123, Name: "Test.mp4", CreatedAt: now, ModifiedAt: now},
		Metadata: []models.FileMetadata{},
	}

	mockResponse := &services.SimilarItemsResponse{
		LocalItems: []services.LocalSimilarItem{{MediaID: "456", Score: 0.85}},
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockRecService.On("GetSimilarItems", mock.Anything, mock.Anything).Return(mockResponse, nil)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/media/{id}/similar", handler.GetSimilarItems)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/media/123/similar", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
	}
}

func BenchmarkConvertFileToMediaMetadata(b *testing.B) {
	handler, _, _, _ := setupRecommendationHandler()

	now := time.Now()
	fileWithMetadata := &models.FileWithMetadata{
		File: models.File{
			ID:         123,
			Name:       "Test.mp4",
			Size:       1024000000,
			MimeType:   "video/mp4",
			CreatedAt:  now,
			ModifiedAt: now,
		},
		Metadata: []models.FileMetadata{
			{FileID: 123, Key: "title", Value: "Test Movie"},
			{FileID: 123, Key: "year", Value: float64(2024)},
			{FileID: 123, Key: "rating", Value: 8.5},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.convertFileToMediaMetadata(fileWithMetadata)
	}
}
