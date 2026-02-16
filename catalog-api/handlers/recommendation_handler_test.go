package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RecommendationHandlerTestSuite struct {
	suite.Suite
	handler *RecommendationHandler
	router  *gin.Engine
}

func (suite *RecommendationHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *RecommendationHandlerTestSuite) SetupTest() {
	// Initialize handler with nil service to test validation paths
	suite.handler = NewRecommendationHandler(nil)

	suite.router = gin.New()
	suite.router.GET("/api/v1/recommendations/similar/:media_id", suite.handler.GetSimilarItems)
	suite.router.GET("/api/v1/recommendations/trending", suite.handler.GetTrendingItems)
	suite.router.GET("/api/v1/recommendations/personalized/:user_id", suite.handler.GetPersonalizedRecommendations)
}

// Constructor tests

func (suite *RecommendationHandlerTestSuite) TestNewRecommendationHandler() {
	handler := NewRecommendationHandler(nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.recommendationService)
}

// GetSimilarItems tests

func (suite *RecommendationHandlerTestSuite) TestGetSimilarItems_InvalidMediaID() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/similar/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid media ID")
}

func (suite *RecommendationHandlerTestSuite) TestGetSimilarItems_FloatMediaID() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/similar/1.5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *RecommendationHandlerTestSuite) TestGetSimilarItems_EmptyStringID() {
	// This gets captured by the router as 404 since the param is required
	req := httptest.NewRequest("GET", "/api/v1/recommendations/similar/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.True(suite.T(), w.Code == http.StatusMovedPermanently || w.Code == http.StatusNotFound)
}

func (suite *RecommendationHandlerTestSuite) TestGetSimilarItems_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/v1/recommendations/similar/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// GetTrendingItems tests

func (suite *RecommendationHandlerTestSuite) TestGetTrendingItems_DefaultParams() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/trending", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp TrendingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "week", resp.TimeRange)
	assert.NotNil(suite.T(), resp.Items)
}

func (suite *RecommendationHandlerTestSuite) TestGetTrendingItems_WithMediaType() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/trending?media_type=movie", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp TrendingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "movie", resp.MediaType)
}

func (suite *RecommendationHandlerTestSuite) TestGetTrendingItems_WithTimeRange() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/trending?time_range=month", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp TrendingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "month", resp.TimeRange)
}

func (suite *RecommendationHandlerTestSuite) TestGetTrendingItems_WithCustomLimit() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/trending?limit=3", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp TrendingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), len(resp.Items), 5) // Capped at 5 in mock
}

// GetPersonalizedRecommendations tests

func (suite *RecommendationHandlerTestSuite) TestGetPersonalizedRecommendations_InvalidUserID() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/personalized/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid user ID")
}

func (suite *RecommendationHandlerTestSuite) TestGetPersonalizedRecommendations_ValidUserID() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/personalized/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp PersonalizedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), resp.UserID)
	assert.NotNil(suite.T(), resp.Items)
	assert.LessOrEqual(suite.T(), len(resp.Items), 3)
}

func (suite *RecommendationHandlerTestSuite) TestGetPersonalizedRecommendations_FloatUserID() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/personalized/1.5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestRecommendationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RecommendationHandlerTestSuite))
}
