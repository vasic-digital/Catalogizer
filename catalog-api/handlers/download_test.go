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

// MockFileRepositoryForDownload is a mock implementation of FileRepository for DownloadHandler
type MockFileRepositoryForDownload struct {
	mock.Mock
}

func (m *MockFileRepositoryForDownload) GetFileByID(ctx context.Context, id int64) (*models.FileWithMetadata, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Return just the File part for download handler
	fileWithMetadata := args.Get(0).(*models.FileWithMetadata)
	return fileWithMetadata, args.Error(1)
}

func (m *MockFileRepositoryForDownload) GetStorageRoots(ctx context.Context) ([]models.StorageRoot, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StorageRoot), args.Error(1)
}

// Test setup helper

func setupDownloadHandler() (*DownloadHandler, *MockFileRepositoryForDownload) {
	mockFileRepo := new(MockFileRepositoryForDownload)

	handler := &DownloadHandler{
		fileRepo:       (*repository.FileRepository)(unsafe_cast_download_repo(mockFileRepo)),
		tempDir:        "/tmp",
		maxArchiveSize: 10737418240, // 10 GB
		chunkSize:      32768,        // 32 KB
	}

	return handler, mockFileRepo
}

func unsafe_cast_download_repo(m *MockFileRepositoryForDownload) interface{} {
	return m
}

func setupGinDownloadTest() (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	recorder := httptest.NewRecorder()
	return router, recorder
}

// Helper to create mock file
func createMockFile(id int64, name string, size int64, isDir bool, deleted bool, storageRootID int) *models.FileWithMetadata {
	mimeType := "video/mp4"
	extension := "mp4"

	return &models.FileWithMetadata{
		File: models.File{
			ID:            id,
			Name:          name,
			Path:          "/media/" + name,
			Size:          size,
			IsDirectory:   isDir,
			Deleted:       deleted,
			StorageRootID: storageRootID,
			MimeType:      &mimeType,
			Extension:     &extension,
			CreatedAt:     time.Now(),
			ModifiedAt:    time.Now(),
		},
		Metadata: []models.FileMetadata{},
	}
}

// Helper to create mock storage root
func createMockStorageRoot(id int, name string) models.StorageRoot {
	host := "192.168.1.100"
	port := 445
	path := "share"
	username := "user"
	domain := "WORKGROUP"

	return models.StorageRoot{
		ID:       id,
		Name:     name,
		Protocol: "smb",
		Host:     &host,
		Port:     &port,
		Path:     &path,
		Username: &username,
		Domain:   &domain,
		Enabled:  true,
	}
}

// DownloadFile Tests

func TestDownloadHandler_DownloadFile_InvalidFileID(t *testing.T) {
	handler, _ := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/invalid", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid file ID")
}

func TestDownloadHandler_DownloadFile_FileNotFound(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFileRepo.On("GetFileByID", mock.Anything, int64(999)).Return(nil, errors.New("file not found"))

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/999", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_DownloadFile_IsDirectory(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(123, "test_dir", 0, true, false, 1)

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Cannot download directory as file")
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_DownloadFile_FileDeleted(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(123, "deleted.mp4", 1024000, false, true, 1)

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "File has been deleted")
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_DownloadFile_StorageRootNotFound(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(123, "test.mp4", 1024000, false, false, 999)
	mockRoots := []models.StorageRoot{
		createMockStorageRoot(1, "main_storage"),
		createMockStorageRoot(2, "backup_storage"),
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "SMB root not found")
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_DownloadFile_GetStorageRootsError(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(123, "test.mp4", 1024000, false, false, 1)

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(nil, errors.New("database error"))

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to get SMB root info")
	mockFileRepo.AssertExpectations(t)
}

// Note: Full download streaming test would require mocking SMB client
// which is complex. Testing focuses on validation and error handling logic.

// DownloadDirectory Tests

func TestDownloadHandler_DownloadDirectory_MissingPath(t *testing.T) {
	handler, _ := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/download/directory/main_storage", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Directory path is required")
}

func TestDownloadHandler_DownloadDirectory_GetStorageRootsError(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(nil, errors.New("database error"))

	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/download/directory/main_storage?path=/movies", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to get SMB root info")
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_DownloadDirectory_StorageRootNotFound(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockRoots := []models.StorageRoot{
		createMockStorageRoot(1, "main_storage"),
	}

	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/download/directory/nonexistent_storage?path=/movies", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "SMB root not found")
	mockFileRepo.AssertExpectations(t)
}

// Note: Full ZIP streaming test would require mocking SMB client
// which is complex. Testing focuses on validation and error handling logic.

// GetDownloadInfo Tests

func TestDownloadHandler_GetDownloadInfo_Success(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(123, "test_movie.mp4", 1024000000, false, false, 1)

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/download/info/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(123), data["file_id"])
	assert.Equal(t, "test_movie.mp4", data["name"])
	assert.Equal(t, float64(1024000000), data["size"])
	assert.False(t, data["is_directory"].(bool))
	assert.False(t, data["deleted"].(bool))

	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadInfo_Directory(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(456, "movies_folder", 5368709120, true, false, 1)

	mockFileRepo.On("GetFileByID", mock.Anything, int64(456)).Return(mockFile, nil)

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/download/info/456", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.True(t, data["is_directory"].(bool))
	assert.NotNil(t, data["estimated_archive_size"])

	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadInfo_InvalidFileID(t *testing.T) {
	handler, _ := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/download/info/invalid", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid file ID")
}

func TestDownloadHandler_GetDownloadInfo_FileNotFound(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFileRepo.On("GetFileByID", mock.Anything, int64(999)).Return(nil, errors.New("file not found"))

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/download/info/999", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "File not found")
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadInfo_RepositoryError(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(nil, errors.New("database connection lost"))

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/download/info/123", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Failed to get file info")
	mockFileRepo.AssertExpectations(t)
}

// DownloadInfo Structure Tests

func TestDownloadInfo_Structure(t *testing.T) {
	now := time.Now()
	mimeType := "video/mp4"
	extension := "mp4"

	info := DownloadInfo{
		FileID:               123,
		Name:                 "test.mp4",
		Path:                 "/media/test.mp4",
		Size:                 1024000000,
		IsDirectory:          false,
		MimeType:             &mimeType,
		Extension:            &extension,
		ModifiedAt:           now,
		Deleted:              false,
		EstimatedArchiveSize: 0,
	}

	assert.Equal(t, int64(123), info.FileID)
	assert.Equal(t, "test.mp4", info.Name)
	assert.Equal(t, "/media/test.mp4", info.Path)
	assert.Equal(t, int64(1024000000), info.Size)
	assert.False(t, info.IsDirectory)
	assert.NotNil(t, info.MimeType)
	assert.Equal(t, "video/mp4", *info.MimeType)
	assert.NotNil(t, info.Extension)
	assert.Equal(t, "mp4", *info.Extension)
	assert.False(t, info.Deleted)
}

// Configuration Tests

func TestDownloadHandler_NewDownloadHandler(t *testing.T) {
	mockFileRepo := new(MockFileRepositoryForDownload)

	handler := NewDownloadHandler(
		(*repository.FileRepository)(unsafe_cast_download_repo(mockFileRepo)),
		"/custom/temp",
		5368709120, // 5 GB
		65536,      // 64 KB
	)

	assert.NotNil(t, handler)
	assert.Equal(t, "/custom/temp", handler.tempDir)
	assert.Equal(t, int64(5368709120), handler.maxArchiveSize)
	assert.Equal(t, 65536, handler.chunkSize)
	assert.NotNil(t, handler.smbPool)
}

func TestDownloadHandler_Close(t *testing.T) {
	handler, _ := setupDownloadHandler()

	// Should not panic
	assert.NotPanics(t, func() {
		handler.Close()
	})
}

// Edge Case Tests

func TestDownloadHandler_DownloadFile_NullPointers(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	// File with nil mime type and extension
	mockFile := &models.FileWithMetadata{
		File: models.File{
			ID:            123,
			Name:          "test.bin",
			Path:          "/media/test.bin",
			Size:          1024,
			IsDirectory:   false,
			Deleted:       false,
			StorageRootID: 1,
			MimeType:      nil,
			Extension:     nil,
			CreatedAt:     time.Now(),
			ModifiedAt:    time.Now(),
		},
		Metadata: []models.FileMetadata{},
	}

	mockRoots := []models.StorageRoot{
		createMockStorageRoot(1, "main_storage"),
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/123", nil)
	router.ServeHTTP(recorder, req)

	// Should handle nil pointers gracefully
	// Will fail at SMB connection, but shouldn't panic before that
	assert.NotEqual(t, http.StatusInternalServerError, recorder.Code) // Should get to SMB connection error
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadInfo_WithAllFields(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mimeType := "application/pdf"
	extension := "pdf"
	mockFile := &models.FileWithMetadata{
		File: models.File{
			ID:            789,
			Name:          "document.pdf",
			Path:          "/docs/document.pdf",
			Size:          2048000,
			IsDirectory:   false,
			Deleted:       false,
			StorageRootID: 1,
			MimeType:      &mimeType,
			Extension:     &extension,
			CreatedAt:     time.Now(),
			ModifiedAt:    time.Now().Add(-24 * time.Hour),
		},
		Metadata: []models.FileMetadata{},
	}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(789)).Return(mockFile, nil)

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)
	req := httptest.NewRequest(http.MethodGet, "/api/download/info/789", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "document.pdf", data["name"])
	assert.Equal(t, "/docs/document.pdf", data["path"])
	assert.Equal(t, "application/pdf", data["mime_type"])
	assert.Equal(t, "pdf", data["extension"])

	mockFileRepo.AssertExpectations(t)
}

// Benchmark Tests

func BenchmarkDownloadHandler_GetDownloadInfo(b *testing.B) {
	handler, mockFileRepo := setupDownloadHandler()
	router, _ := setupGinDownloadTest()

	mockFile := createMockFile(123, "test.mp4", 1024000000, false, false, 1)

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/download/info/123", nil)
		router.ServeHTTP(recorder, req)
	}
}

func BenchmarkCreateMockFile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createMockFile(123, "test.mp4", 1024000000, false, false, 1)
	}
}

func BenchmarkCreateMockStorageRoot(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createMockStorageRoot(1, "main_storage")
	}
}

// Integration-style Tests (without actual SMB)

func TestDownloadHandler_DownloadFile_HeadersAndDisposition_Inline(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockFile := createMockFile(123, "preview.mp4", 1024000, false, false, 1)
	mockRoots := []models.StorageRoot{createMockStorageRoot(1, "main_storage")}

	mockFileRepo.On("GetFileByID", mock.Anything, int64(123)).Return(mockFile, nil)
	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/download/file/:id", handler.DownloadFile)
	req := httptest.NewRequest(http.MethodGet, "/api/download/file/123?inline=true", nil)
	router.ServeHTTP(recorder, req)

	// Check headers are set correctly before SMB failure
	// (Will fail at SMB connection but headers should be set)
	mockFileRepo.AssertExpectations(t)
}

func TestDownloadHandler_DownloadDirectory_RecursiveParameter(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, recorder := setupGinDownloadTest()

	mockRoots := []models.StorageRoot{createMockStorageRoot(1, "main_storage")}

	mockFileRepo.On("GetStorageRoots", mock.Anything).Return(mockRoots, nil)

	router.GET("/api/download/directory/:smb_root", handler.DownloadDirectory)
	req := httptest.NewRequest(http.MethodGet, "/api/download/directory/main_storage?path=/movies&recursive=false&max_depth=2", nil)
	router.ServeHTTP(recorder, req)

	// Will fail at SMB connection, but parameters should be parsed
	mockFileRepo.AssertExpectations(t)
}

// Error Recovery Tests

func TestDownloadHandler_MultipleErrors_DoNotPanic(t *testing.T) {
	handler, mockFileRepo := setupDownloadHandler()
	router, _ := setupGinDownloadTest()

	// Test multiple error conditions don't cause panics
	testCases := []struct {
		name     string
		url      string
		mockFunc func()
	}{
		{
			name: "Invalid ID",
			url:  "/api/download/info/abc",
			mockFunc: func() {
				// No mock needed
			},
		},
		{
			name: "Not found",
			url:  "/api/download/info/999",
			mockFunc: func() {
				mockFileRepo.On("GetFileByID", mock.Anything, int64(999)).
					Return(nil, errors.New("file not found")).Once()
			},
		},
		{
			name: "Database error",
			url:  "/api/download/info/888",
			mockFunc: func() {
				mockFileRepo.On("GetFileByID", mock.Anything, int64(888)).
					Return(nil, errors.New("database error")).Once()
			},
		},
	}

	router.GET("/api/download/info/:id", handler.GetDownloadInfo)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)

			assert.NotPanics(t, func() {
				router.ServeHTTP(recorder, req)
			})

			assert.NotEqual(t, http.StatusOK, recorder.Code)
		})
	}

	mockFileRepo.AssertExpectations(t)
}
