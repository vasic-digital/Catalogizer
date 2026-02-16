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

type SimpleRecommendationHandlerTestSuite struct {
	suite.Suite
	handler *SimpleRecommendationHandler
	router  *gin.Engine
}

func (suite *SimpleRecommendationHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *SimpleRecommendationHandlerTestSuite) SetupTest() {
	suite.handler = NewSimpleRecommendationHandler()

	suite.router = gin.New()
	suite.router.GET("/api/v1/recommendations/simple", suite.handler.GetSimpleRecommendation)
	suite.router.GET("/api/v1/recommendations/test", suite.handler.GetTest)
}

// Constructor test

func (suite *SimpleRecommendationHandlerTestSuite) TestNewSimpleRecommendationHandler() {
	handler := NewSimpleRecommendationHandler()
	assert.NotNil(suite.T(), handler)
}

// GetSimpleRecommendation tests

func (suite *SimpleRecommendationHandlerTestSuite) TestGetSimpleRecommendation_Success() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/simple", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Simple recommendation works!", resp["message"])
}

func (suite *SimpleRecommendationHandlerTestSuite) TestGetSimpleRecommendation_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/v1/recommendations/simple", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// GetTest tests

func (suite *SimpleRecommendationHandlerTestSuite) TestGetTest_ReturnsError() {
	req := httptest.NewRequest("GET", "/api/v1/recommendations/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), false, resp["success"])
	assert.Contains(suite.T(), resp["error"], "Test error")
}

func (suite *SimpleRecommendationHandlerTestSuite) TestGetTest_MethodNotAllowed() {
	req := httptest.NewRequest("DELETE", "/api/v1/recommendations/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func TestSimpleRecommendationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SimpleRecommendationHandlerTestSuite))
}
