package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SearchHandlerTestSuite struct {
	suite.Suite
	handler  *SearchHandler
	fileRepo *repository.FileRepository
	router   *gin.Engine
}

func (suite *SearchHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *SearchHandlerTestSuite) SetupTest() {
	suite.fileRepo = nil
	suite.handler = NewSearchHandler(suite.fileRepo)

	suite.router = gin.New()
	suite.router.GET("/api/search", suite.handler.SearchFiles)
	suite.router.GET("/api/search/duplicates", suite.handler.SearchDuplicates)
	suite.router.POST("/api/search/advanced", suite.handler.AdvancedSearch)
}

// Test handler initialization
func (suite *SearchHandlerTestSuite) TestNewSearchHandler() {
	handler := NewSearchHandler(nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.fileRepo)
}

func (suite *SearchHandlerTestSuite) TestNewSearchHandler_WithRepository() {
	repo := &repository.FileRepository{}
	handler := NewSearchHandler(repo)
	assert.NotNil(suite.T(), handler)
	assert.Equal(suite.T(), repo, handler.fileRepo)
}

// Test HTTP method restrictions
func (suite *SearchHandlerTestSuite) TestSearchFiles_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/search", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *SearchHandlerTestSuite) TestSearchDuplicates_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/search/duplicates", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *SearchHandlerTestSuite) TestAdvancedSearch_MethodNotAllowed() {
	req := httptest.NewRequest("GET", "/api/search/advanced", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test date validation in SearchFiles (these validate BEFORE repository calls)
func (suite *SearchHandlerTestSuite) TestSearchFiles_InvalidModifiedAfterDate() {
	req := httptest.NewRequest("GET", "/api/search?modified_after=invalid-date", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid modified_after date format")
}

func (suite *SearchHandlerTestSuite) TestSearchFiles_InvalidModifiedBeforeDate() {
	req := httptest.NewRequest("GET", "/api/search?modified_before=2024-13-45", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid modified_before date format")
}

func (suite *SearchHandlerTestSuite) TestSearchFiles_InvalidDateFormats() {
	invalidDates := []string{
		"2024-01-01",           // Missing time
		"01/01/2024",           // Wrong format
		"2024-13-01T00:00:00Z", // Invalid month
		"2024-01-32T00:00:00Z", // Invalid day
		"not-a-date",           // Completely invalid
	}

	for _, date := range invalidDates {
		req := httptest.NewRequest("GET", "/api/search?modified_after="+date, nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusBadRequest, w.Code,
			"Date %s should be rejected", date)
	}
}

// Note: Valid date tests omitted - they pass validation but fail at repository level

// Test AdvancedSearch input validation
func (suite *SearchHandlerTestSuite) TestAdvancedSearch_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/search/advanced", bytes.NewBufferString("invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid request body")
}

// Note: Tests that would require a working repository are omitted.
// These tests focus on input validation that fails BEFORE repository calls.

// Run the test suite
func TestSearchHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SearchHandlerTestSuite))
}
