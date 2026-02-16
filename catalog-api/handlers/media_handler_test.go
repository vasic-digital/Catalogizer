package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MediaHandlerTestSuite struct {
	suite.Suite
	handler *AndroidTVMediaHandler
	router  *gin.Engine
}

func (suite *MediaHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *MediaHandlerTestSuite) SetupTest() {
	// Initialize handler with nil database to test guard paths
	suite.handler = NewAndroidTVMediaHandler(nil)

	suite.router = gin.New()
	suite.router.GET("/api/v1/media/:id", suite.handler.GetMediaByID)
	suite.router.PUT("/api/v1/media/:id/progress", suite.handler.UpdateWatchProgress)
	suite.router.PUT("/api/v1/media/:id/favorite", suite.handler.UpdateFavoriteStatus)
}

// Constructor tests

func (suite *MediaHandlerTestSuite) TestNewAndroidTVMediaHandler() {
	handler := NewAndroidTVMediaHandler(nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.db)
}

// GetMediaByID tests

func (suite *MediaHandlerTestSuite) TestGetMediaByID_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/media/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), false, body["success"])
	assert.Contains(suite.T(), body["error"], "Invalid media ID")
}

func (suite *MediaHandlerTestSuite) TestGetMediaByID_NegativeID() {
	req := httptest.NewRequest("GET", "/api/v1/media/-1", nil)
	w := httptest.NewRecorder()

	// Negative ID parses fine as int64, but db is nil so it will panic
	assert.Panics(suite.T(), func() {
		suite.router.ServeHTTP(w, req)
	})
}

func (suite *MediaHandlerTestSuite) TestGetMediaByID_EmptyID() {
	// With the route pattern :id, an empty ID results in 404 from the router
	req := httptest.NewRequest("GET", "/api/v1/media/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Router returns 301 redirect or 404 for trailing slash
	assert.True(suite.T(), w.Code == http.StatusMovedPermanently || w.Code == http.StatusNotFound)
}

func (suite *MediaHandlerTestSuite) TestGetMediaByID_FloatID() {
	req := httptest.NewRequest("GET", "/api/v1/media/1.5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaHandlerTestSuite) TestGetMediaByID_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/v1/media/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// UpdateWatchProgress tests

func (suite *MediaHandlerTestSuite) TestUpdateWatchProgress_InvalidID() {
	body := bytes.NewBufferString(`{"progress": 0.5}`)
	req := httptest.NewRequest("PUT", "/api/v1/media/abc/progress", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid media ID")
}

func (suite *MediaHandlerTestSuite) TestUpdateWatchProgress_InvalidJSON() {
	req := httptest.NewRequest("PUT", "/api/v1/media/1/progress", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid request body")
}

func (suite *MediaHandlerTestSuite) TestUpdateWatchProgress_MissingProgressField() {
	body := bytes.NewBufferString(`{"something": 0.5}`)
	req := httptest.NewRequest("PUT", "/api/v1/media/1/progress", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Progress field is required")
}

func (suite *MediaHandlerTestSuite) TestUpdateWatchProgress_ProgressTooHigh() {
	body := bytes.NewBufferString(`{"progress": 1.5}`)
	req := httptest.NewRequest("PUT", "/api/v1/media/1/progress", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Progress must be between 0.0 and 1.0")
}

func (suite *MediaHandlerTestSuite) TestUpdateWatchProgress_ProgressNegative() {
	body := bytes.NewBufferString(`{"progress": -0.1}`)
	req := httptest.NewRequest("PUT", "/api/v1/media/1/progress", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaHandlerTestSuite) TestUpdateWatchProgress_EmptyBody() {
	req := httptest.NewRequest("PUT", "/api/v1/media/1/progress", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// UpdateFavoriteStatus tests

func (suite *MediaHandlerTestSuite) TestUpdateFavoriteStatus_InvalidID() {
	body := bytes.NewBufferString(`{"favorite": true}`)
	req := httptest.NewRequest("PUT", "/api/v1/media/abc/favorite", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid media ID")
}

func (suite *MediaHandlerTestSuite) TestUpdateFavoriteStatus_InvalidJSON() {
	req := httptest.NewRequest("PUT", "/api/v1/media/1/favorite", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *MediaHandlerTestSuite) TestUpdateFavoriteStatus_MissingFavoriteField() {
	body := bytes.NewBufferString(`{"something": true}`)
	req := httptest.NewRequest("PUT", "/api/v1/media/1/favorite", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Favorite field is required")
}

func (suite *MediaHandlerTestSuite) TestUpdateFavoriteStatus_EmptyBody() {
	req := httptest.NewRequest("PUT", "/api/v1/media/1/favorite", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestMediaHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MediaHandlerTestSuite))
}
