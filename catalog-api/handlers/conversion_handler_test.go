package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConversionService for testing
type MockConversionService struct {
	mock.Mock
}

func (m *MockConversionService) CreateConversionJob(userID int, request *models.ConversionRequest) (*models.ConversionJob, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConversionJob), args.Error(1)
}

func (m *MockConversionService) GetJob(jobID int, userID int) (*models.ConversionJob, error) {
	args := m.Called(jobID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConversionJob), args.Error(1)
}

func (m *MockConversionService) GetUserJobs(userID int, status *string, limit, offset int) ([]models.ConversionJob, error) {
	args := m.Called(userID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ConversionJob), args.Error(1)
}

func (m *MockConversionService) CancelJob(jobID int, userID int) error {
	args := m.Called(jobID, userID)
	return args.Error(0)
}

func (m *MockConversionService) GetSupportedFormats() *models.SupportedFormats {
	args := m.Called()
	return args.Get(0).(*models.SupportedFormats)
}

// MockConversionAuthService for testing
type MockConversionAuthService struct {
	mock.Mock
}

func (m *MockConversionAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func setupConversionHandler() (*ConversionHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	
	mockConversionService := &MockConversionService{}
	mockAuthService := &MockConversionAuthService{}
	handler := NewConversionHandler(mockConversionService, mockAuthService)
	
	router := gin.New()
	router.POST("/jobs", handler.CreateJob)
	router.GET("/jobs", handler.ListJobs)
	router.GET("/jobs/:id", handler.GetJob)
	router.POST("/jobs/:id/cancel", handler.CancelJob)
	router.GET("/formats", handler.GetSupportedFormats)
	
	return handler, router
}

func TestCreateJob(t *testing.T) {
	handler, router := setupConversionHandler()
	
	// Set up mock expectations
	mockConversionService := handler.conversionService.(*MockConversionService)
	mockAuthService := handler.authService.(*MockConversionAuthService)
	
	mockAuthService.On("CheckPermission", 1, models.PermissionConversionCreate).Return(true, nil)
	
	request := &models.ConversionRequest{
		SourcePath:     "/test/source.mp4",
		TargetPath:     "/test/target.mkv",
		SourceFormat:   "mp4",
		TargetFormat:   "mkv",
		ConversionType: models.ConversionTypeVideo,
		Quality:        "medium",
		Priority:       1,
	}
	
	expectedJob := &models.ConversionJob{
		ID:             1,
		UserID:         1,
		SourcePath:     request.SourcePath,
		TargetPath:     request.TargetPath,
		SourceFormat:   request.SourceFormat,
		TargetFormat:   request.TargetFormat,
		ConversionType: request.ConversionType,
		Quality:        request.Quality,
		Status:         models.ConversionStatusPending,
		Priority:       request.Priority,
	}
	
	mockConversionService.On("CreateConversionJob", 1, request).Return(expectedJob, nil)
	
	// Create request body
	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/jobs", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.ConversionJob
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedJob.UserID, response.UserID)
	assert.Equal(t, expectedJob.SourcePath, response.SourcePath)
	assert.Equal(t, expectedJob.TargetPath, response.TargetPath)
	
	// Verify mock expectations
	mockAuthService.AssertExpectations(t)
	mockConversionService.AssertExpectations(t)
}

func TestGetJob(t *testing.T) {
	handler, router := setupConversionHandler()
	
	// Set up mock expectations
	mockConversionService := handler.conversionService.(*MockConversionService)
	
	// GetJob doesn't check permissions, only needs the conversion service
	expectedJob := &models.ConversionJob{
		ID:             1,
		UserID:         1,
		SourcePath:     "/test/source.mp4",
		TargetPath:     "/test/target.mkv",
		SourceFormat:   "mp4",
		TargetFormat:   "mkv",
		ConversionType: models.ConversionTypeVideo,
		Status:         models.ConversionStatusCompleted,
	}
	
	mockConversionService.On("GetJob", 1, 1).Return(expectedJob, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/jobs/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.ConversionJob
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedJob.ID, response.ID)
	assert.Equal(t, expectedJob.SourcePath, response.SourcePath)
	
	// Verify mock expectations
	mockConversionService.AssertExpectations(t)
}

func TestListJobs(t *testing.T) {
	handler, router := setupConversionHandler()
	
	// Set up mock expectations
	mockConversionService := handler.conversionService.(*MockConversionService)
	mockAuthService := handler.authService.(*MockConversionAuthService)
	
	mockAuthService.On("CheckPermission", 1, models.PermissionConversionView).Return(true, nil)
	
	expectedJobs := []models.ConversionJob{
		{
			ID:             1,
			UserID:         1,
			SourcePath:     "/test/source1.mp4",
			TargetPath:     "/test/target1.mkv",
			ConversionType: models.ConversionTypeVideo,
			Status:         models.ConversionStatusCompleted,
		},
		{
			ID:             2,
			UserID:         1,
			SourcePath:     "/test/source2.mp3",
			TargetPath:     "/test/target2.wav",
			ConversionType: models.ConversionTypeAudio,
			Status:         models.ConversionStatusPending,
		},
	}
	
	// When no status query param is provided, status will be an empty string
	// so we need to expect a pointer to an empty string, not nil
	emptyStatus := ""
	mockConversionService.On("GetUserJobs", 1, &emptyStatus, 50, 0).Return(expectedJobs, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/jobs", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response []models.ConversionJob
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	
	// Verify mock expectations
	mockAuthService.AssertExpectations(t)
	mockConversionService.AssertExpectations(t)
}

func TestCancelJob(t *testing.T) {
	handler, router := setupConversionHandler()
	
	// Set up mock expectations
	mockConversionService := handler.conversionService.(*MockConversionService)
	mockAuthService := handler.authService.(*MockConversionAuthService)
	
	mockAuthService.On("CheckPermission", 1, models.PermissionConversionManage).Return(true, nil)
	mockConversionService.On("CancelJob", 1, 1).Return(nil)
	
	// Create request
	req, _ := http.NewRequest("POST", "/jobs/1/cancel", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Job cancelled successfully", response["message"])
	
	// Verify mock expectations
	mockAuthService.AssertExpectations(t)
	mockConversionService.AssertExpectations(t)
}

func TestGetSupportedFormats(t *testing.T) {
	handler, router := setupConversionHandler()
	
	// Set up mock expectations
	mockConversionService := handler.conversionService.(*MockConversionService)
	mockAuthService := handler.authService.(*MockConversionAuthService)
	
	mockAuthService.On("CheckPermission", 1, models.PermissionConversionView).Return(true, nil)
	
	expectedFormats := &models.SupportedFormats{
		Video: models.VideoFormats{
			Input:  []string{"mp4", "avi", "mkv", "mov"},
			Output: []string{"mp4", "avi", "mkv", "mov", "webm"},
		},
		Audio: models.AudioFormats{
			Input:  []string{"mp3", "wav", "flac"},
			Output: []string{"mp3", "wav", "flac", "aac"},
		},
		Document: models.DocumentFormats{
			Input:  []string{"pdf", "epub", "mobi"},
			Output: []string{"pdf", "epub", "mobi", "txt"},
		},
		Image: models.ImageFormats{
			Input:  []string{"jpg", "png", "gif"},
			Output: []string{"jpg", "png", "gif", "bmp"},
		},
	}
	
	mockConversionService.On("GetSupportedFormats").Return(expectedFormats)
	
	// Create request
	req, _ := http.NewRequest("GET", "/formats", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.SupportedFormats
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedFormats.Video.Input, response.Video.Input)
	assert.Equal(t, expectedFormats.Audio.Input, response.Audio.Input)
	
	// Verify mock expectations
	mockAuthService.AssertExpectations(t)
	mockConversionService.AssertExpectations(t)
}