package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DownloadHandlerTestSuite struct {
	suite.Suite
	handler  *DownloadHandler
	fileRepo *repository.FileRepository
	router   *gin.Engine
}

func (suite *DownloadHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *DownloadHandlerTestSuite) SetupTest() {
	suite.fileRepo = nil
	suite.handler = NewDownloadHandler(suite.fileRepo, "/tmp", 1024*1024*100, 4096)

	suite.router = gin.New()
	suite.router.GET("/api/download/file/:id", suite.handler.DownloadFile)
	suite.router.GET("/api/download/directory/:smb_root", suite.handler.DownloadDirectory)
	suite.router.GET("/api/download/info/:id", suite.handler.GetDownloadInfo)
}

// Test handler initialization
func (suite *DownloadHandlerTestSuite) TestNewDownloadHandler() {
	handler := NewDownloadHandler(nil, "/tmp", 1024*1024, 4096)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.fileRepo)
	assert.NotNil(suite.T(), handler.smbPool)
	assert.Equal(suite.T(), "/tmp", handler.tempDir)
	assert.Equal(suite.T(), int64(1024*1024), handler.maxArchiveSize)
	assert.Equal(suite.T(), 4096, handler.chunkSize)
}

func (suite *DownloadHandlerTestSuite) TestNewDownloadHandler_WithRepository() {
	fileRepo := &repository.FileRepository{}
	handler := NewDownloadHandler(fileRepo, "/custom", 2048, 8192)
	assert.NotNil(suite.T(), handler)
	assert.Equal(suite.T(), fileRepo, handler.fileRepo)
	assert.Equal(suite.T(), "/custom", handler.tempDir)
	assert.Equal(suite.T(), int64(2048), handler.maxArchiveSize)
	assert.Equal(suite.T(), 8192, handler.chunkSize)
}

// Test HTTP method restrictions
func (suite *DownloadHandlerTestSuite) TestDownloadFile_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/download/file/123", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/download/directory/main?path=/test", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestGetDownloadInfo_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/download/info/123", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test DownloadFile input validation
func (suite *DownloadHandlerTestSuite) TestDownloadFile_InvalidFileID_NotANumber() {
	req := httptest.NewRequest("GET", "/api/download/file/abc", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadFile_InvalidFileID_Empty() {
	req := httptest.NewRequest("GET", "/api/download/file/", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Empty ID should result in not found (no route match)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadFile_InvalidFileID_SpecialCharacters() {
	invalidIDs := []string{
		"123abc",
		"!@#$",
		"12.34",
		"12e5",
	}

	for _, id := range invalidIDs {
		req := httptest.NewRequest("GET", "/api/download/file/"+id, nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code,
			"File ID %s should be rejected", id)
	}
}

// Test GetDownloadInfo input validation
func (suite *DownloadHandlerTestSuite) TestGetDownloadInfo_InvalidFileID_NotANumber() {
	req := httptest.NewRequest("GET", "/api/download/info/abc", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestGetDownloadInfo_InvalidFileID_SpecialCharacters() {
	invalidIDs := []string{
		"test",
		"!@#",
		"12.5",
	}

	for _, id := range invalidIDs {
		req := httptest.NewRequest("GET", "/api/download/info/"+id, nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code,
			"File ID %s should be rejected", id)
	}
}

// Test DownloadDirectory input validation
func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_MissingPathParameter() {
	req := httptest.NewRequest("GET", "/api/download/directory/main", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Missing path query parameter
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_EmptyPathParameter() {
	req := httptest.NewRequest("GET", "/api/download/directory/main?path=", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Empty path parameter
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *DownloadHandlerTestSuite) TestDownloadDirectory_MissingSmbRoot() {
	req := httptest.NewRequest("GET", "/api/download/directory/?path=/test", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Missing smb_root should result in route not found
	assert.True(suite.T(), w.Code == http.StatusNotFound || w.Code == http.StatusMovedPermanently)
}

// Test query parameter handling
func (suite *DownloadHandlerTestSuite) TestDownloadFile_InlineQueryParameter() {
	// Test that inline parameter is accepted (though will fail at repo level)
	// This just validates the query parameter doesn't cause parsing errors
	req := httptest.NewRequest("GET", "/api/download/file/abc?inline=true", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Should still fail on invalid ID, not on query parameter
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Note: Tests that would pass validation but fail at repository/SMB level are omitted.
// These tests focus only on HTTP method restrictions, ID parsing, and required parameter validation.

// Run the test suite
func TestDownloadHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(DownloadHandlerTestSuite))
}
