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
	if args.Get(0) == nil {
		return nil
	}
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

func (m *MockConversionAuthService) GetCurrentUser(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// TestCreateJob tests the conversion handler's CreateJob method
func TestCreateJob(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int
		hasPermission  bool
		permissionErr  error
		requestData    *models.ConversionRequest
		mockResponse   *models.ConversionJob
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			hasPermission: true,
			permissionErr: nil,
			requestData: &models.ConversionRequest{
				SourcePath:   "/input/test.pdf",
				TargetPath:   "/output/test.docx",
				SourceFormat: "pdf",
				TargetFormat: "docx",
				Quality:      "high",
			},
			mockResponse:   &models.ConversionJob{ID: 123, Status: "pending"},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockConversionService := &MockConversionService{}
			mockAuthService := &MockConversionAuthService{}

			// Setup expectations
			mockAuthService.On("CheckPermission", tt.userID, models.PermissionConversionCreate).Return(tt.hasPermission, tt.permissionErr)
			mockAuthService.On("GetCurrentUser", "test-token").Return(&models.User{ID: tt.userID}, nil)
			mockConversionService.On("CreateConversionJob", tt.userID, tt.requestData).Return(tt.mockResponse, tt.serviceError)

			// Create handler with mocks
			handler := NewConversionHandler(mockConversionService, mockAuthService)

			// Setup request
			body, _ := json.Marshal(tt.requestData)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/conversion/jobs", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Authorization", "Bearer test-token")
			c.Set("user_id", tt.userID)

			// Call handler
			handler.CreateJob(c)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError && tt.mockResponse != nil {
				var responseJob models.ConversionJob
				err := json.Unmarshal(w.Body.Bytes(), &responseJob)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, responseJob.ID)
			}

			mockAuthService.AssertExpectations(t)
			mockConversionService.AssertExpectations(t)
		})
	}
}
