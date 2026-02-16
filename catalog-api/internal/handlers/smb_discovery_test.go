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

type SMBDiscoveryHandlerTestSuite struct {
	suite.Suite
	handler *SMBDiscoveryHandler
	router  *gin.Engine
	logger  *zap.Logger
}

func (suite *SMBDiscoveryHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.logger = zap.NewNop()
}

func (suite *SMBDiscoveryHandlerTestSuite) SetupTest() {
	// Initialize handler with nil service to test validation paths
	suite.handler = NewSMBDiscoveryHandler(nil, suite.logger)

	suite.router = gin.New()
	suite.router.POST("/api/v1/smb/discover", suite.handler.DiscoverShares)
	suite.router.POST("/api/v1/smb/test", suite.handler.TestConnection)
	suite.router.POST("/api/v1/smb/browse", suite.handler.BrowseShare)
	suite.router.GET("/api/v1/smb/discover", suite.handler.DiscoverSharesGET)
	suite.router.GET("/api/v1/smb/test", suite.handler.TestConnectionGET)
}

// Constructor tests

func (suite *SMBDiscoveryHandlerTestSuite) TestNewSMBDiscoveryHandler() {
	handler := NewSMBDiscoveryHandler(nil, suite.logger)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.service)
	assert.NotNil(suite.T(), handler.logger)
}

// DiscoverShares (POST) tests

func (suite *SMBDiscoveryHandlerTestSuite) TestDiscoverShares_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/smb/discover", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"].(string), "Invalid request")
}

func (suite *SMBDiscoveryHandlerTestSuite) TestDiscoverShares_MissingRequiredFields() {
	body := `{"host": ""}`
	req := httptest.NewRequest("POST", "/api/v1/smb/discover", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBDiscoveryHandlerTestSuite) TestDiscoverShares_EmptyBody() {
	req := httptest.NewRequest("POST", "/api/v1/smb/discover", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestConnection (POST) tests

func (suite *SMBDiscoveryHandlerTestSuite) TestTestConnection_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/smb/test", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBDiscoveryHandlerTestSuite) TestTestConnection_MissingFields() {
	body := `{"host": "192.168.1.1"}`
	req := httptest.NewRequest("POST", "/api/v1/smb/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// BrowseShare tests

func (suite *SMBDiscoveryHandlerTestSuite) TestBrowseShare_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/smb/browse", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBDiscoveryHandlerTestSuite) TestBrowseShare_MissingFields() {
	body := `{"host": "192.168.1.1"}`
	req := httptest.NewRequest("POST", "/api/v1/smb/browse", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// DiscoverSharesGET tests

func (suite *SMBDiscoveryHandlerTestSuite) TestDiscoverSharesGET_MissingParams() {
	req := httptest.NewRequest("GET", "/api/v1/smb/discover", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"].(string), "host, username, and password are required")
}

func (suite *SMBDiscoveryHandlerTestSuite) TestDiscoverSharesGET_MissingUsername() {
	req := httptest.NewRequest("GET", "/api/v1/smb/discover?host=192.168.1.1&password=pass", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBDiscoveryHandlerTestSuite) TestDiscoverSharesGET_MissingPassword() {
	req := httptest.NewRequest("GET", "/api/v1/smb/discover?host=192.168.1.1&username=user", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestConnectionGET tests

func (suite *SMBDiscoveryHandlerTestSuite) TestTestConnectionGET_MissingParams() {
	req := httptest.NewRequest("GET", "/api/v1/smb/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"].(string), "host, share, username, and password are required")
}

func (suite *SMBDiscoveryHandlerTestSuite) TestTestConnectionGET_MissingShare() {
	req := httptest.NewRequest("GET", "/api/v1/smb/test?host=192.168.1.1&username=user&password=pass", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *SMBDiscoveryHandlerTestSuite) TestTestConnectionGET_PartialParams() {
	req := httptest.NewRequest("GET", "/api/v1/smb/test?host=192.168.1.1&share=share1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestSMBDiscoveryHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SMBDiscoveryHandlerTestSuite))
}
