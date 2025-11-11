package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileRepository is a mock implementation of FileRepository
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) GetStorageRoots(ctx context.Context) ([]*models.StorageRoot, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.StorageRoot), args.Error(1)
}

func (m *MockFileRepository) GetDirectoryContents(ctx context.Context, storageRoot, path string, pagination models.PaginationOptions, sort models.SortOptions) (*models.SearchResult, error) {
	args := m.Called(ctx, storageRoot, path, pagination, sort)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SearchResult), args.Error(1)
}

func (m *MockFileRepository) GetFileByID(ctx context.Context, id int64) (*models.FileWithMetadata, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileWithMetadata), args.Error(1)
}

func (m *MockFileRepository) GetDirectoriesSortedBySize(ctx context.Context, storageRoot string, pagination models.PaginationOptions, ascending bool) ([]*models.DirectoryInfo, error) {
	args := m.Called(ctx, storageRoot, pagination, ascending)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DirectoryInfo), args.Error(1)
}

func (m *MockFileRepository) GetDirectoriesSortedByDuplicates(ctx context.Context, storageRoot string, pagination models.PaginationOptions, ascending bool) ([]*models.DirectoryInfo, error) {
	args := m.Called(ctx, storageRoot, pagination, ascending)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DirectoryInfo), args.Error(1)
}

// Test setup helpers

func setupBrowseHandler() (*BrowseHandler, *MockFileRepository) {
	mockRepo := new(MockFileRepository)
	handler := NewBrowseHandler(&repository.FileRepository{})
	handler.fileRepo = (*repository.FileRepository)(mockRepo)
	return handler, mockRepo
}

func setupGinTest() (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	recorder := httptest.NewRecorder()
	return router, recorder
}

// GetStorageRoots Tests

func TestBrowseHandler_GetStorageRoots_Success(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRoots := []*models.StorageRoot{
		{
			ID:       1,
			Name:     "main_storage",
			Protocol: "smb",
			Host:     "192.168.1.100",
			Port:     445,
			Enabled:  true,
		},
		{
			ID:       2,
			Name:     "backup_storage",
			Protocol: "nfs",
			Host:     "192.168.1.200",
			Port:     2049,
			Enabled:  true,
		},
	}

	mockRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/browse/roots", handler.GetStorageRoots)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/roots", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetStorageRoots_RepositoryError(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRepo.On("GetStorageRoots", mock.Anything).Return(nil, errors.New("database connection failed"))

	router.GET("/api/browse/roots", handler.GetStorageRoots)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/roots", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockRepo.AssertExpectations(t)
}

// BrowseDirectory Tests

func TestBrowseHandler_BrowseDirectory_Success(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	now := time.Now()
	mockResult := &models.SearchResult{
		Files: []*models.File{
			{
				ID:          1,
				Name:        "test_movie.mp4",
				Path:        "/movies/test_movie.mp4",
				Size:        1024000000,
				Extension:   "mp4",
				MimeType:    "video/mp4",
				IsDirectory: false,
				CreatedAt:   now,
				ModifiedAt:  now,
			},
		},
		Total:  1,
		Limit:  100,
		Offset: 0,
	}

	mockRepo.On("GetDirectoryContents",
		mock.Anything,
		"main_storage",
		"/movies",
		models.PaginationOptions{Page: 1, Limit: 100},
		models.SortOptions{Field: "name", Order: "asc"},
	).Return(mockResult, nil)

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage?path=/movies", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_BrowseDirectory_WithPagination(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockResult := &models.SearchResult{
		Files:  []*models.File{},
		Total:  100,
		Limit:  50,
		Offset: 50,
	}

	mockRepo.On("GetDirectoryContents",
		mock.Anything,
		"main_storage",
		"/",
		models.PaginationOptions{Page: 2, Limit: 50},
		models.SortOptions{Field: "name", Order: "asc"},
	).Return(mockResult, nil)

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage?page=2&limit=50", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_BrowseDirectory_WithSorting(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockResult := &models.SearchResult{
		Files:  []*models.File{},
		Total:  0,
		Limit:  100,
		Offset: 0,
	}

	mockRepo.On("GetDirectoryContents",
		mock.Anything,
		"main_storage",
		"/",
		models.PaginationOptions{Page: 1, Limit: 100},
		models.SortOptions{Field: "size", Order: "desc"},
	).Return(mockResult, nil)

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage?sort_by=size&sort_order=desc", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_BrowseDirectory_InvalidSortField(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockResult := &models.SearchResult{
		Files:  []*models.File{},
		Total:  0,
		Limit:  100,
		Offset: 0,
	}

	// Should default to "name" when invalid sort field provided
	mockRepo.On("GetDirectoryContents",
		mock.Anything,
		"main_storage",
		"/",
		models.PaginationOptions{Page: 1, Limit: 100},
		models.SortOptions{Field: "name", Order: "asc"},
	).Return(mockResult, nil)

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage?sort_by=invalid_field", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_BrowseDirectory_InvalidPagination(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockResult := &models.SearchResult{
		Files:  []*models.File{},
		Total:  0,
		Limit:  100,
		Offset: 0,
	}

	// Should default page to 1 and limit to 100 when invalid values provided
	mockRepo.On("GetDirectoryContents",
		mock.Anything,
		"main_storage",
		"/",
		models.PaginationOptions{Page: 1, Limit: 100},
		models.SortOptions{Field: "name", Order: "asc"},
	).Return(mockResult, nil)

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage?page=0&limit=2000", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_BrowseDirectory_MissingStorageRoot(t *testing.T) {
	handler, _ := setupBrowseHandler()
	router, recorder := setupGinTest()

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code) // Gin returns 404 for missing param
}

func TestBrowseHandler_BrowseDirectory_RepositoryError(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRepo.On("GetDirectoryContents",
		mock.Anything,
		"main_storage",
		"/",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("failed to connect to storage"))

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockRepo.AssertExpectations(t)
}

// GetFileInfo Tests

func TestBrowseHandler_GetFileInfo_Success(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File: models.File{
			ID:          123,
			Name:        "test_movie.mp4",
			Path:        "/movies/test_movie.mp4",
			Size:        1024000000,
			Extension:   "mp4",
			MimeType:    "video/mp4",
			IsDirectory: false,
			CreatedAt:   now,
			ModifiedAt:  now,
		},
		Metadata: []models.FileMetadata{
			{
				FileID: 123,
				Key:    "title",
				Value:  "Test Movie",
			},
			{
				FileID: 123,
				Key:    "year",
				Value:  float64(2024),
			},
		},
	}

	mockRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)

	router.GET("/api/browse/file/:id", handler.GetFileInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/file/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetFileInfo_InvalidID(t *testing.T) {
	handler, _ := setupBrowseHandler()
	router, recorder := setupGinTest()

	router.GET("/api/browse/file/:id", handler.GetFileInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/file/invalid_id", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestBrowseHandler_GetFileInfo_NotFound(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRepo.On("GetFileByID", mock.Anything, int64(999)).Return(nil, errors.New("file not found"))

	router.GET("/api/browse/file/:id", handler.GetFileInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/file/999", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetFileInfo_RepositoryError(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRepo.On("GetFileByID", mock.Anything, int64(123)).Return(nil, errors.New("database connection lost"))

	router.GET("/api/browse/file/:id", handler.GetFileInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/file/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockRepo.AssertExpectations(t)
}

// GetDirectorySizes Tests

func TestBrowseHandler_GetDirectorySizes_Success(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{
		{
			Path:           "/movies",
			TotalSize:      10737418240, // 10 GB
			FileCount:      50,
			DuplicateCount: 5,
		},
		{
			Path:           "/tv_shows",
			TotalSize:      5368709120, // 5 GB
			FileCount:      30,
			DuplicateCount: 2,
		},
	}

	mockRepo.On("GetDirectoriesSortedBySize",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 1, Limit: 50},
		false,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/sizes", handler.GetDirectorySizes)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/sizes", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectorySizes_Ascending(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{}

	mockRepo.On("GetDirectoriesSortedBySize",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 1, Limit: 50},
		true,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/sizes", handler.GetDirectorySizes)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/sizes?ascending=true", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectorySizes_WithPagination(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{}

	mockRepo.On("GetDirectoriesSortedBySize",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 3, Limit: 100},
		false,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/sizes", handler.GetDirectorySizes)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/sizes?page=3&limit=100", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectorySizes_InvalidPagination(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{}

	// Should default to page=1, limit=50 when invalid values provided
	mockRepo.On("GetDirectoriesSortedBySize",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 1, Limit: 50},
		false,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/sizes", handler.GetDirectorySizes)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/sizes?page=-1&limit=1000", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectorySizes_MissingStorageRoot(t *testing.T) {
	handler, _ := setupBrowseHandler()
	router, recorder := setupGinTest()

	router.GET("/api/browse/:storage_root/sizes", handler.GetDirectorySizes)
	req := httptest.NewRequest(http.MethodGet, "/api/browse//sizes", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestBrowseHandler_GetDirectorySizes_RepositoryError(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRepo.On("GetDirectoriesSortedBySize",
		mock.Anything,
		"main_storage",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("failed to calculate directory sizes"))

	router.GET("/api/browse/:storage_root/sizes", handler.GetDirectorySizes)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/sizes", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockRepo.AssertExpectations(t)
}

// GetDirectoryDuplicates Tests

func TestBrowseHandler_GetDirectoryDuplicates_Success(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{
		{
			Path:           "/downloads",
			TotalSize:      2147483648, // 2 GB
			FileCount:      100,
			DuplicateCount: 25,
		},
		{
			Path:           "/movies",
			TotalSize:      10737418240, // 10 GB
			FileCount:      50,
			DuplicateCount: 10,
		},
	}

	mockRepo.On("GetDirectoriesSortedByDuplicates",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 1, Limit: 50},
		false,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/duplicates", handler.GetDirectoryDuplicates)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/duplicates", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectoryDuplicates_Ascending(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{}

	mockRepo.On("GetDirectoriesSortedByDuplicates",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 1, Limit: 50},
		true,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/duplicates", handler.GetDirectoryDuplicates)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/duplicates?ascending=true", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectoryDuplicates_WithPagination(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{}

	mockRepo.On("GetDirectoriesSortedByDuplicates",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 2, Limit: 25},
		false,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/duplicates", handler.GetDirectoryDuplicates)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/duplicates?page=2&limit=25", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectoryDuplicates_InvalidPagination(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockDirectories := []*models.DirectoryInfo{}

	// Should default to page=1, limit=50 when invalid values provided
	mockRepo.On("GetDirectoriesSortedByDuplicates",
		mock.Anything,
		"main_storage",
		models.PaginationOptions{Page: 1, Limit: 50},
		false,
	).Return(mockDirectories, nil)

	router.GET("/api/browse/:storage_root/duplicates", handler.GetDirectoryDuplicates)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/duplicates?page=0&limit=600", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	mockRepo.AssertExpectations(t)
}

func TestBrowseHandler_GetDirectoryDuplicates_MissingStorageRoot(t *testing.T) {
	handler, _ := setupBrowseHandler()
	router, recorder := setupGinTest()

	router.GET("/api/browse/:storage_root/duplicates", handler.GetDirectoryDuplicates)
	req := httptest.NewRequest(http.MethodGet, "/api/browse//duplicates", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestBrowseHandler_GetDirectoryDuplicates_RepositoryError(t *testing.T) {
	handler, mockRepo := setupBrowseHandler()
	router, recorder := setupGinTest()

	mockRepo.On("GetDirectoriesSortedByDuplicates",
		mock.Anything,
		"main_storage",
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("failed to analyze duplicates"))

	router.GET("/api/browse/:storage_root/duplicates", handler.GetDirectoryDuplicates)
	req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage/duplicates", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockRepo.AssertExpectations(t)
}

// Benchmark Tests

func BenchmarkBrowseHandler_GetStorageRoots(b *testing.B) {
	handler, mockRepo := setupBrowseHandler()
	router, _ := setupGinTest()

	mockRoots := []*models.StorageRoot{
		{ID: 1, Name: "storage1", Protocol: "smb", Enabled: true},
		{ID: 2, Name: "storage2", Protocol: "nfs", Enabled: true},
	}

	mockRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/browse/roots", handler.GetStorageRoots)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/browse/roots", nil)
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkBrowseHandler_BrowseDirectory(b *testing.B) {
	handler, mockRepo := setupBrowseHandler()
	router, _ := setupGinTest()

	now := time.Now()
	mockResult := &models.SearchResult{
		Files: []*models.File{
			{ID: 1, Name: "file1.mp4", Size: 1024000000, CreatedAt: now, ModifiedAt: now},
			{ID: 2, Name: "file2.mp4", Size: 2048000000, CreatedAt: now, ModifiedAt: now},
		},
		Total:  2,
		Limit:  100,
		Offset: 0,
	}

	mockRepo.On("GetDirectoryContents", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockResult, nil)

	router.GET("/api/browse/:storage_root", handler.BrowseDirectory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/browse/main_storage?path=/", nil)
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkBrowseHandler_GetFileInfo(b *testing.B) {
	handler, mockRepo := setupBrowseHandler()
	router, _ := setupGinTest()

	now := time.Now()
	mockFile := &models.FileWithMetadata{
		File: models.File{
			ID:         123,
			Name:       "test.mp4",
			Size:       1024000000,
			CreatedAt:  now,
			ModifiedAt: now,
		},
		Metadata: []models.FileMetadata{},
	}

	mockRepo.On("GetFileByID", mock.Anything, mock.Anything).Return(mockFile, nil)

	router.GET("/api/browse/file/:id", handler.GetFileInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/browse/file/123", nil)
		router.ServeHTTP(recorder, req)
	}
}
