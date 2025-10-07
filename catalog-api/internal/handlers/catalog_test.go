package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	_ "github.com/mattn/go-sqlite3"
)

type CatalogHandlerTestSuite struct {
	suite.Suite
	router  *gin.Engine
	handler *CatalogHandler
	db      *sql.DB
	logger  *zap.Logger
}

func (suite *CatalogHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// Initialize logger
	logger, _ := zap.NewDevelopment()
	suite.logger = logger

	// Initialize in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	suite.Require().NoError(err)
	suite.db = db

	// Setup database
	suite.setupDatabase()

	// Initialize services
	catalogService := &mockCatalogService{db: db}
	smbService := &mockSMBService{}

	// Initialize handler
	suite.handler = NewCatalogHandler(catalogService, smbService, logger)

	// Setup router
	suite.router = gin.Default()
	suite.setupRoutes()
}

func (suite *CatalogHandlerTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *CatalogHandlerTestSuite) setupDatabase() {
	_, err := suite.db.Exec(`
		CREATE TABLE media_items (
			id INTEGER PRIMARY KEY,
			name TEXT,
			path TEXT,
			size INTEGER,
			media_type TEXT
		);
		INSERT INTO media_items VALUES
		(1, 'movie1.mp4', '/media/movies/movie1.mp4', 1000000, 'movie'),
		(2, 'song1.mp3', '/media/music/song1.mp3', 5000000, 'music');
	`)
	suite.Require().NoError(err)
}

func (suite *CatalogHandlerTestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")
	{
		api.GET("/catalog", suite.handler.ListRoot)
		api.GET("/catalog/*path", suite.handler.ListPath)
		api.GET("/catalog-info/*path", suite.handler.GetFileInfo)
		api.GET("/search", suite.handler.Search)
		api.GET("/search/duplicates", suite.handler.SearchDuplicates)
		api.GET("/stats/directories/by-size", suite.handler.GetDirectoriesBySize)
		api.GET("/stats/duplicates/count", suite.handler.GetDuplicatesCount)
	}
}

// Mock services for testing
type mockCatalogService struct {
	db *sql.DB
}

func (m *mockCatalogService) ListDirectory(path string) ([]FileItem, error) {
	if path == "/" {
		return []FileItem{{Name: "media", Type: "directory", Path: "/media"}}, nil
	}
	if path == "/media" {
		return []FileItem{
			{Name: "movies", Type: "directory", Path: "/media/movies"},
			{Name: "music", Type: "directory", Path: "/media/music"},
		}, nil
	}
	return []FileItem{}, nil
}

func (m *mockCatalogService) GetFileInfo(path string) (*FileInfo, error) {
	if path == "/media/movies/movie1.mp4" {
		return &FileInfo{
			Name:      "movie1.mp4",
			Path:      "/media/movies/movie1.mp4",
			Size:      1000000,
			MediaType: "movie",
		}, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockCatalogService) Search(query, mediaType string, limit, offset int) ([]SearchResult, error) {
	return []SearchResult{
		{Name: "movie1.mp4", Path: "/media/movies/movie1.mp4", MediaType: "movie"},
	}, nil
}

func (m *mockCatalogService) SearchDuplicates() ([]DuplicateGroup, error) {
	return []DuplicateGroup{}, nil
}

func (m *mockCatalogService) GetDirectoriesBySize(limit int) ([]DirectoryInfo, error) {
	return []DirectoryInfo{
		{Path: "/media", TotalSize: 6000000, FileCount: 2},
	}, nil
}

func (m *mockCatalogService) GetDuplicatesCount() (int, error) {
	return 0, nil
}

func (m *mockCatalogService) SetDB(db *sql.DB) {}

type mockSMBService struct{}

func (m *mockSMBService) Connect(path string) error { return nil }
func (m *mockSMBService) ListDirectory(path string) ([]FileItem, error) { return []FileItem{}, nil }
func (m *mockSMBService) IsValidSMBPath(path string) bool { return true }
func (m *mockSMBService) ParseSMBPath(path string) SMBPath { return SMBPath{} }

func (suite *CatalogHandlerTestSuite) TestListRoot() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["items"])
}

func (suite *CatalogHandlerTestSuite) TestListPath() {
	// Test media directory
	req, _ := http.NewRequest("GET", "/api/v1/catalog/media", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["items"])

	items := response["items"].([]interface{})
	assert.Len(suite.T(), items, 2)
}

func (suite *CatalogHandlerTestSuite) TestGetFileInfo() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog-info/media/movies/movie1.mp4", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response FileInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "movie1.mp4", response.Name)
	assert.Equal(suite.T(), int64(1000000), response.Size)
	assert.Equal(suite.T(), "movie", response.MediaType)
}

func (suite *CatalogHandlerTestSuite) TestGetFileInfoNotFound() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog-info/nonexistent.mp4", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestSearch() {
	req, _ := http.NewRequest("GET", "/api/v1/search?q=movie", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["results"])
}

func (suite *CatalogHandlerTestSuite) TestSearchEmptyQuery() {
	req, _ := http.NewRequest("GET", "/api/v1/search", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestSearchDuplicates() {
	req, _ := http.NewRequest("GET", "/api/v1/search/duplicates", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestGetDirectoriesBySize() {
	req, _ := http.NewRequest("GET", "/api/v1/stats/directories/by-size", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["directories"])
}

func (suite *CatalogHandlerTestSuite) TestGetDuplicatesCount() {
	req, _ := http.NewRequest("GET", "/api/v1/stats/duplicates/count", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), float64(0), response["count"])
}

func (suite *CatalogHandlerTestSuite) TestPaginationParameters() {
	req, _ := http.NewRequest("GET", "/api/v1/search?q=test&limit=10&offset=5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestInvalidPath() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog/..", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should handle gracefully (either return empty or handle the path traversal)
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
}

func TestCatalogHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CatalogHandlerTestSuite))
}