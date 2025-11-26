package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hirochachacha/go-smb2"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"catalogizer/internal/models"
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

func (m *mockCatalogService) SetDB(db *sql.DB) {}
func (m *mockCatalogService) ListPath(path string, sortBy string, sortOrder string, limit, offset int) ([]models.FileInfo, error) {
	if path == "media" {
		return []models.FileInfo{
			{Name: "movies", Path: "/media/movies", IsDirectory: true},
			{Name: "music", Path: "/media/music", IsDirectory: true},
		}, nil
	}
	return []models.FileInfo{}, nil
}
func (m *mockCatalogService) GetFileInfo(pathOrID string) (*models.FileInfo, error) {
	if pathOrID == "media/movies/movie1.mp4" {
		return &models.FileInfo{
			Name:      "movie1.mp4",
			Path:      "/media/movies/movie1.mp4",
			Size:      1000000,
			MediaType: func() *string { s := "movie"; return &s }(),
		}, nil
	}
	return nil, sql.ErrNoRows
}
func (m *mockCatalogService) SearchFiles(req *models.SearchRequest) ([]models.FileInfo, int64, error) {
	return []models.FileInfo{}, 0, nil
}
func (m *mockCatalogService) GetDirectoriesBySize(smbRoot string, limit int) ([]models.DirectoryStats, error) {
	return []models.DirectoryStats{}, nil
}
func (m *mockCatalogService) GetDuplicateGroups(smbRoot string, minCount int, limit int) ([]models.DuplicateGroup, error) {
	return []models.DuplicateGroup{}, nil
}
func (m *mockCatalogService) GetSMBRoots() ([]string, error) { return []string{}, nil }

func (m *mockCatalogService) GetDuplicatesCount() (int64, error) { return 0, nil }
func (m *mockCatalogService) GetDirectoriesBySizeLimited(limit int) ([]models.DirectoryStats, error) {
	return []models.DirectoryStats{}, nil
}

func (m *mockCatalogService) ListDirectory(path string) ([]models.FileInfo, error) {
	if path == "/" {
		return []models.FileInfo{{Name: "media", Path: "/media", IsDirectory: true}}, nil
	}
	if path == "/media" {
		return []models.FileInfo{
			{Name: "movies", Path: "/media/movies", IsDirectory: true},
			{Name: "music", Path: "/media/music", IsDirectory: true},
		}, nil
	}
	return []models.FileInfo{}, nil
}

func (m *mockCatalogService) GetFileInfoByPath(path string) (*models.FileInfo, error) {
	if path == "/media/movies/movie1.mp4" {
		return &models.FileInfo{
			Name:      "movie1.mp4",
			Path:      "/media/movies/movie1.mp4",
			Size:      1000000,
			MediaType: func() *string { s := "movie"; return &s }(),
		}, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockCatalogService) Search(query, fileType string, limit, offset int) ([]models.FileInfo, error) {
	return []models.FileInfo{
		{Name: "movie1.mp4", Path: "/media/movies/movie1.mp4"},
	}, nil
}

func (m *mockCatalogService) SearchDuplicates() ([]models.DuplicateGroup, error) {
	return []models.DuplicateGroup{}, nil
}

type mockSMBService struct{}

func (m *mockSMBService) GetHosts() []string { return []string{} }
func (m *mockSMBService) ListFiles(hostName, path string) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
func (m *mockSMBService) DownloadFile(hostName, remotePath, localPath string) error { return nil }
func (m *mockSMBService) UploadFile(hostName, localPath, remotePath string) error   { return nil }
func (m *mockSMBService) CopyFile(sourceHost, sourcePath, destHost, destPath string) error {
	return nil
}
func (m *mockSMBService) CreateRemoteDir(share *smb2.Share, path string) error { return nil }
func (m *mockSMBService) FileExists(hostName, path string) (bool, error)       { return false, nil }
func (m *mockSMBService) Connect(hostName string) error                        { return nil }
func (m *mockSMBService) ListDirectory(hostName, path string) ([]*models.FileInfo, error) {
	return []*models.FileInfo{}, nil
}
func (m *mockSMBService) IsConnected(hostName string) bool                    { return true }
func (m *mockSMBService) GetFileSize(hostName, path string) (int64, error)    { return 0, nil }
func (m *mockSMBService) CreateDirectory(hostName, path string) error         { return nil }
func (m *mockSMBService) DeleteDirectory(hostName, path string) error         { return nil }
func (m *mockSMBService) DirectoryExists(hostName, path string) (bool, error) { return false, nil }
func (m *mockSMBService) IsValidSMBPath(path string) bool                     { return true }
func (m *mockSMBService) ParseSMBPath(path string) models.SMBPath             { return models.SMBPath{} }

func (suite *CatalogHandlerTestSuite) TestListRoot() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["roots"])
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
	assert.NotNil(suite.T(), response["files"])

	files := response["files"].([]interface{})
	assert.Len(suite.T(), files, 2)
}

func (suite *CatalogHandlerTestSuite) TestGetFileInfo() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog-info/media/movies/movie1.mp4", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.FileInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "movie1.mp4", response.Name)
	assert.Equal(suite.T(), int64(1000000), response.Size)
	assert.NotNil(suite.T(), response.MediaType)
	assert.Equal(suite.T(), "movie", *response.MediaType)
}

func (suite *CatalogHandlerTestSuite) TestGetFileInfoNotFound() {
	req, _ := http.NewRequest("GET", "/api/v1/catalog-info/nonexistent.mp4", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestSearch() {
	req, _ := http.NewRequest("GET", "/api/v1/search?query=movie", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["files"])
}

func (suite *CatalogHandlerTestSuite) TestSearchEmptyQuery() {
	req, _ := http.NewRequest("GET", "/api/v1/search", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestSearchDuplicates() {
	req, _ := http.NewRequest("GET", "/api/v1/search/duplicates?smb_root=test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *CatalogHandlerTestSuite) TestGetDirectoriesBySize() {
	req, _ := http.NewRequest("GET", "/api/v1/stats/directories/by-size?smb_root=test", nil)
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
	assert.Equal(suite.T(), float64(0), response["total_duplicates"])
}

func (suite *CatalogHandlerTestSuite) TestPaginationParameters() {
	req, _ := http.NewRequest("GET", "/api/v1/search?query=test&limit=10&offset=5", nil)
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
