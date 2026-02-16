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
	"go.uber.org/zap"
)

type SMBHandlerTestSuite struct {
	suite.Suite
	handler *SMBHandler
	router  *gin.Engine
	logger  *zap.Logger
}

func (suite *SMBHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.logger = zap.NewNop()
}

func (suite *SMBHandlerTestSuite) SetupTest() {
	// Initialize handler with nil manager to test validation paths
	suite.handler = NewSMBHandler(nil, suite.logger)

	suite.router = gin.New()
	suite.router.POST("/api/v1/smb/sources", suite.handler.AddSource)
	suite.router.DELETE("/api/v1/smb/sources/:id", suite.handler.RemoveSource)
	suite.router.GET("/api/v1/smb/sources/status", suite.handler.GetSourcesStatus)
	suite.router.GET("/api/v1/smb/sources/:id", suite.handler.GetSourceDetails)
	suite.router.POST("/api/v1/smb/test-connection", suite.handler.TestConnection)
	suite.router.POST("/api/v1/smb/sources/:id/reconnect", suite.handler.ForceReconnect)
	suite.router.GET("/api/v1/smb/statistics", suite.handler.GetStatistics)
	suite.router.PUT("/api/v1/smb/sources/:id", suite.handler.UpdateSource)
	suite.router.GET("/api/v1/smb/health", suite.handler.GetHealth)
}

// Constructor tests

func (suite *SMBHandlerTestSuite) TestNewSMBHandler() {
	handler := NewSMBHandler(nil, suite.logger)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.smbManager)
	assert.NotNil(suite.T(), handler.logger)
}

// AddSource tests

func (suite *SMBHandlerTestSuite) TestAddSource_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/smb/sources", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", resp["error"])
}

func (suite *SMBHandlerTestSuite) TestAddSource_MissingRequiredFields() {
	body := `{"username": "user"}`
	req := httptest.NewRequest("POST", "/api/v1/smb/sources", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBHandlerTestSuite) TestAddSource_EmptyBody() {
	req := httptest.NewRequest("POST", "/api/v1/smb/sources", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// RemoveSource tests

func (suite *SMBHandlerTestSuite) TestRemoveSource_EmptyID() {
	// Empty ID is not matched by the route
	req := httptest.NewRequest("DELETE", "/api/v1/smb/sources/", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Router redirects or returns 404
	assert.True(suite.T(), w.Code == http.StatusMovedPermanently || w.Code == http.StatusNotFound)
}

// GetSourceDetails tests

func (suite *SMBHandlerTestSuite) TestGetSourceDetails_NonExistentSource() {
	// With nil manager, this will panic - test that handler expects manager
	assert.Panics(suite.T(), func() {
		req := httptest.NewRequest("GET", "/api/v1/smb/sources/nonexistent", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
	})
}

// TestConnection tests

func (suite *SMBHandlerTestSuite) TestTestConnection_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/smb/test-connection", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBHandlerTestSuite) TestTestConnection_MissingFields() {
	body := `{"path": ""}`
	req := httptest.NewRequest("POST", "/api/v1/smb/test-connection", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// ForceReconnect tests

func (suite *SMBHandlerTestSuite) TestForceReconnect_EmptyID() {
	// Gin matches the route with an empty :id parameter, so the handler responds with 400
	req := httptest.NewRequest("POST", "/api/v1/smb/sources//reconnect", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Source ID is required", resp["error"])
}

// UpdateSource tests

func (suite *SMBHandlerTestSuite) TestUpdateSource_InvalidJSON() {
	req := httptest.NewRequest("PUT", "/api/v1/smb/sources/source1", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Helper method tests

func (suite *SMBHandlerTestSuite) TestGenerateStatusSummary_EmptyStatus() {
	status := map[string]interface{}{}
	summary := suite.handler.generateStatusSummary(status)

	assert.Equal(suite.T(), 0, summary["total"])
	assert.Equal(suite.T(), 0, summary["connected"])
	assert.Equal(suite.T(), 0, summary["disconnected"])
	assert.Equal(suite.T(), 0, summary["reconnecting"])
	assert.Equal(suite.T(), 0, summary["offline"])
}

func (suite *SMBHandlerTestSuite) TestGenerateStatusSummary_WithSources() {
	status := map[string]interface{}{
		"source1": map[string]interface{}{"state": "connected"},
		"source2": map[string]interface{}{"state": "disconnected"},
		"source3": map[string]interface{}{"state": "connected"},
		"source4": map[string]interface{}{"state": "offline"},
	}
	summary := suite.handler.generateStatusSummary(status)

	assert.Equal(suite.T(), 4, summary["total"])
	assert.Equal(suite.T(), 2, summary["connected"])
	assert.Equal(suite.T(), 1, summary["disconnected"])
	assert.Equal(suite.T(), 1, summary["offline"])
}

func (suite *SMBHandlerTestSuite) TestCalculateErrorStats_EmptyStatus() {
	status := map[string]interface{}{}
	stats := suite.handler.calculateErrorStats(status)

	assert.Equal(suite.T(), 0, stats["total_errors"])
	assert.Equal(suite.T(), 0, stats["sources_with_errors"])
}

func (suite *SMBHandlerTestSuite) TestCalculateErrorStats_WithErrors() {
	status := map[string]interface{}{
		"source1": map[string]interface{}{"retry_attempts": 3},
		"source2": map[string]interface{}{"retry_attempts": 0},
		"source3": map[string]interface{}{"retry_attempts": 5},
	}
	stats := suite.handler.calculateErrorStats(status)

	assert.Equal(suite.T(), 8, stats["total_errors"])
	assert.Equal(suite.T(), 2, stats["sources_with_errors"])
}

func TestSMBHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SMBHandlerTestSuite))
}
