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

type StatsHandlerTestSuite struct {
	suite.Suite
	handler   *StatsHandler
	fileRepo  *repository.FileRepository
	statsRepo *repository.StatsRepository
	router    *gin.Engine
}

func (suite *StatsHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *StatsHandlerTestSuite) SetupTest() {
	suite.fileRepo = nil
	suite.statsRepo = nil
	suite.handler = NewStatsHandler(suite.fileRepo, suite.statsRepo)

	suite.router = gin.New()
	suite.router.GET("/api/stats/overall", suite.handler.GetOverallStats)
	suite.router.GET("/api/stats/smb/:smb_root", suite.handler.GetSmbRootStats)
	suite.router.GET("/api/stats/filetypes", suite.handler.GetFileTypeStats)
	suite.router.GET("/api/stats/sizes", suite.handler.GetSizeDistribution)
}

// Test handler initialization
func (suite *StatsHandlerTestSuite) TestNewStatsHandler() {
	handler := NewStatsHandler(nil, nil)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.fileRepo)
	assert.Nil(suite.T(), handler.statsRepo)
}

func (suite *StatsHandlerTestSuite) TestNewStatsHandler_WithRepositories() {
	fileRepo := &repository.FileRepository{}
	statsRepo := &repository.StatsRepository{}
	handler := NewStatsHandler(fileRepo, statsRepo)
	assert.NotNil(suite.T(), handler)
	assert.Equal(suite.T(), fileRepo, handler.fileRepo)
	assert.Equal(suite.T(), statsRepo, handler.statsRepo)
}

// Test HTTP method restrictions
func (suite *StatsHandlerTestSuite) TestGetOverallStats_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/stats/overall", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *StatsHandlerTestSuite) TestGetSmbRootStats_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/stats/smb/main", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *StatsHandlerTestSuite) TestGetFileTypeStats_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/stats/filetypes", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *StatsHandlerTestSuite) TestGetSizeDistribution_MethodNotAllowed() {
	req := httptest.NewRequest("POST", "/api/stats/sizes", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test required path parameters
func (suite *StatsHandlerTestSuite) TestGetSmbRootStats_RequiresSmbRoot() {
	req := httptest.NewRequest("GET", "/api/stats/smb/", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Missing smb_root should result in not found (no route match)
	assert.True(suite.T(), w.Code == http.StatusNotFound || w.Code == http.StatusMovedPermanently)
}

// Test validation in GetTopDuplicateGroups

func (suite *StatsHandlerTestSuite) TestGetTopDuplicateGroups_InvalidSortBy() {
	// Add route for this endpoint
	suite.router.GET("/api/stats/duplicates/groups", suite.handler.GetTopDuplicateGroups)

	req := httptest.NewRequest("GET", "/api/stats/duplicates/groups?sort_by=invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "sort_by must be 'count' or 'size'")
}

func (suite *StatsHandlerTestSuite) TestGetTopDuplicateGroups_MethodNotAllowed() {
	suite.router.GET("/api/stats/duplicates/groups", suite.handler.GetTopDuplicateGroups)

	req := httptest.NewRequest("POST", "/api/stats/duplicates/groups", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Test struct types

func TestScanHistoryResponse_Fields(t *testing.T) {
	resp := ScanHistoryResponse{
		TotalCount: 100,
		Limit:      50,
		Offset:     10,
	}

	assert.Equal(t, int64(100), resp.TotalCount)
	assert.Equal(t, 50, resp.Limit)
	assert.Equal(t, 10, resp.Offset)
	assert.Nil(t, resp.Scans)
}

func TestOverallStats_Fields(t *testing.T) {
	stats := OverallStats{
		TotalFiles:         1000,
		TotalDirectories:   50,
		TotalSize:          1024 * 1024 * 1024,
		TotalDuplicates:    100,
		DuplicateGroups:    25,
		StorageRootsCount:  3,
		ActiveStorageRoots: 2,
		LastScanTime:       1707000000,
	}

	assert.Equal(t, int64(1000), stats.TotalFiles)
	assert.Equal(t, int64(50), stats.TotalDirectories)
	assert.Equal(t, int64(3), stats.StorageRootsCount)
}

func TestSizeDistribution_Fields(t *testing.T) {
	dist := SizeDistribution{
		Tiny:    100,
		Small:   200,
		Medium:  50,
		Large:   20,
		Huge:    5,
		Massive: 1,
	}

	assert.Equal(t, int64(100), dist.Tiny)
	assert.Equal(t, int64(200), dist.Small)
	assert.Equal(t, int64(1), dist.Massive)
}

// Note: Tests that would pass validation but fail at repository level are omitted.
// These tests focus only on HTTP method restrictions, route matching, and handler initialization.

// Run the test suite
func TestStatsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(StatsHandlerTestSuite))
}
