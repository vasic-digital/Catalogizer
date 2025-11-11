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

type BrowseHandlerTestSuite struct {
	suite.Suite
	handler  *BrowseHandler
	fileRepo *repository.FileRepository
	router   *gin.Engine
}

func (suite *BrowseHandlerTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func (suite *BrowseHandlerTestSuite) SetupTest() {
	// Initialize handler with nil repository
	suite.fileRepo = nil
	suite.handler = NewBrowseHandler(suite.fileRepo)

	// Setup test router
	suite.router = gin.New()
	suite.router.GET("/api/browse/roots", suite.handler.GetStorageRoots)
	suite.router.GET("/api/browse/:storage_root", suite.handler.BrowseDirectory)
	suite.router.GET("/api/browse/file/:id", suite.handler.GetFileInfo)
	suite.router.GET("/api/browse/:storage_root/sizes", suite.handler.GetDirectorySizes)
	suite.router.GET("/api/browse/:storage_root/duplicates", suite.handler.GetDirectoryDuplicates)
}

// Test handler initialization
func (suite *BrowseHandlerTestSuite) TestNewBrowseHandler() {
	handler := NewBrowseHandler(nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.fileRepo)
}

func (suite *BrowseHandlerTestSuite) TestNewBrowseHandler_WithRepository() {
	repo := &repository.FileRepository{}
	handler := NewBrowseHandler(repo)
	assert.NotNil(suite.T(), handler)
	assert.Equal(suite.T(), repo, handler.fileRepo)
}

// Test HTTP method restrictions (these don't call repository)
func (suite *BrowseHandlerTestSuite) TestGetStorageRoots_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/browse/roots", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// POST not allowed, only GET
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *BrowseHandlerTestSuite) TestBrowseDirectory_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/browse/main", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// POST not allowed, only GET
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *BrowseHandlerTestSuite) TestGetFileInfo_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/browse/file/123", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// POST not allowed, only GET
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test required path parameters (route matching)
func (suite *BrowseHandlerTestSuite) TestBrowseDirectory_RequiresStorageRoot() {
	req := httptest.NewRequest("GET", "/api/browse/", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Missing storage_root should result in not found (no route match)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *BrowseHandlerTestSuite) TestGetFileInfo_RequiresID() {
	req := httptest.NewRequest("GET", "/api/browse/file/", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Missing ID should result in not found or redirect
	assert.True(suite.T(), w.Code == http.StatusNotFound || w.Code == http.StatusMovedPermanently)
}

// Test input validation (these fail before repository calls)
func (suite *BrowseHandlerTestSuite) TestGetFileInfo_InvalidID_NotANumber() {
	req := httptest.NewRequest("GET", "/api/browse/file/abc", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Invalid ID should return bad request
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid file ID")
}

func (suite *BrowseHandlerTestSuite) TestGetFileInfo_InvalidID_SpecialCharacters() {
	req := httptest.NewRequest("GET", "/api/browse/file/file@123", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Invalid ID with special characters should return bad request
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid file ID")
}

func (suite *BrowseHandlerTestSuite) TestGetFileInfo_InvalidID_Decimal() {
	req := httptest.NewRequest("GET", "/api/browse/file/123.45", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Decimal ID should return bad request
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

//Note: Tests for valid IDs and storage roots that would succeed (return 2xx/5xx)
// cannot be tested without a working repository, as the handlers don't validate
// parameters before calling repository methods. These tests focus on input
// validation that fails early (4xx errors) and route matching (404 errors).

// Run the test suite
func TestBrowseHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BrowseHandlerTestSuite))
}
