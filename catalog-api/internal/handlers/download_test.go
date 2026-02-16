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

type DownloadHandlerTestSuite struct {
	suite.Suite
	handler *DownloadHandler
	router  *gin.Engine
	logger  *zap.Logger
}

func (suite *DownloadHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.logger = zap.NewNop()
}

func (suite *DownloadHandlerTestSuite) SetupTest() {
	// Initialize handler with nil services and reasonable defaults
	suite.handler = NewDownloadHandler(nil, nil, "/tmp", 1024*1024*100, 32768, suite.logger)

	suite.router = gin.New()
	suite.router.GET("/api/v1/download/file/:id", suite.handler.DownloadFile)
	suite.router.GET("/api/v1/download/directory/*path", suite.handler.DownloadDirectory)
	suite.router.POST("/api/v1/download/archive", suite.handler.DownloadArchive)
}

// Constructor tests

func (suite *DownloadHandlerTestSuite) TestNewDownloadHandler() {
	handler := NewDownloadHandler(nil, nil, "/tmp", 1024, 4096, suite.logger)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.catalogService)
	assert.Nil(suite.T(), handler.smbService)
	assert.Equal(suite.T(), "/tmp", handler.tempDir)
	assert.Equal(suite.T(), int64(1024), handler.maxArchiveSize)
	assert.Equal(suite.T(), 4096, handler.chunkSize)
}

func (suite *DownloadHandlerTestSuite) TestNewDownloadHandler_NilLogger() {
	handler := NewDownloadHandler(nil, nil, "/tmp", 0, 0, nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.logger)
}

// DownloadFile tests

func (suite *DownloadHandlerTestSuite) TestDownloadFile_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/download/file/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid file ID", resp["error"])
}

func (suite *DownloadHandlerTestSuite) TestDownloadFile_FloatID() {
	req := httptest.NewRequest("GET", "/api/v1/download/file/1.5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadFile_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/v1/download/file/1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// DownloadDirectory tests

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_InvalidFormat() {
	req := httptest.NewRequest("GET", "/api/v1/download/directory/test?format=rar", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid format")
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_DefaultFormat() {
	// With nil services, the recursive directory listing returns empty
	req := httptest.NewRequest("GET", "/api/v1/download/directory/testpath", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Returns 404 because the placeholder getDirectoryContentsRecursive returns empty
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_ZipFormat() {
	req := httptest.NewRequest("GET", "/api/v1/download/directory/testpath?format=zip", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_TarFormat() {
	req := httptest.NewRequest("GET", "/api/v1/download/directory/testpath?format=tar", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_TarGzFormat() {
	req := httptest.NewRequest("GET", "/api/v1/download/directory/testpath?format=tar.gz", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// DownloadArchive tests

func (suite *DownloadHandlerTestSuite) TestDownloadArchive_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/download/archive", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", resp["error"])
}

func (suite *DownloadHandlerTestSuite) TestDownloadArchive_EmptyPaths() {
	body := `{"paths": [], "format": "zip"}`
	req := httptest.NewRequest("POST", "/api/v1/download/archive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "No paths specified", resp["error"])
}

func (suite *DownloadHandlerTestSuite) TestDownloadArchive_InvalidFormat() {
	body := `{"paths": ["/test/path"], "format": "7z"}`
	req := httptest.NewRequest("POST", "/api/v1/download/archive", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), resp["error"], "Invalid format")
}

func (suite *DownloadHandlerTestSuite) TestDownloadArchive_EmptyBody() {
	req := httptest.NewRequest("POST", "/api/v1/download/archive", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestDownloadHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(DownloadHandlerTestSuite))
}
