package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/services"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockErrorReportingService for testing
type MockErrorReportingService struct {
	mock.Mock
}

func (m *MockErrorReportingService) ReportError(userID int, request *models.ErrorReportRequest) (*models.ErrorReport, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorReport), args.Error(1)
}

func (m *MockErrorReportingService) ReportCrash(userID int, request *models.CrashReportRequest) (*models.CrashReport, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashReport), args.Error(1)
}

func (m *MockErrorReportingService) GetErrorReport(reportID int, userID int) (*models.ErrorReport, error) {
	args := m.Called(reportID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorReport), args.Error(1)
}

func (m *MockErrorReportingService) GetCrashReport(reportID int, userID int) (*models.CrashReport, error) {
	args := m.Called(reportID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashReport), args.Error(1)
}

func (m *MockErrorReportingService) GetErrorReportsByUser(userID int, filters *models.ErrorReportFilters) ([]models.ErrorReport, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ErrorReport), args.Error(1)
}

func (m *MockErrorReportingService) ListErrorReports(userID int, filters *models.ErrorReportFilters) ([]models.ErrorReport, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ErrorReport), args.Error(1)
}

func (m *MockErrorReportingService) GetCrashReportsByUser(userID int, filters *models.CrashReportFilters) ([]models.CrashReport, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CrashReport), args.Error(1)
}

func (m *MockErrorReportingService) ListCrashReports(userID int, filters *models.CrashReportFilters) ([]models.CrashReport, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CrashReport), args.Error(1)
}

func (m *MockErrorReportingService) UpdateErrorStatus(reportID int, userID int, status string) error {
	args := m.Called(reportID, userID, status)
	return args.Error(0)
}

func (m *MockErrorReportingService) UpdateCrashStatus(reportID int, userID int, status string) error {
	args := m.Called(reportID, userID, status)
	return args.Error(0)
}

func (m *MockErrorReportingService) GetErrorStatistics(userID int) (*models.ErrorStatistics, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorStatistics), args.Error(1)
}

func (m *MockErrorReportingService) GetErrorStatisticsWithDate(userID *int, startDate, endDate time.Time) (*models.ErrorStatistics, error) {
	args := m.Called(userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ErrorStatistics), args.Error(1)
}

func (m *MockErrorReportingService) GetCrashStatistics(userID int) (*models.CrashStatistics, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashStatistics), args.Error(1)
}

func (m *MockErrorReportingService) GetCrashStatisticsWithDate(userID *int, startDate, endDate time.Time) (*models.CrashStatistics, error) {
	args := m.Called(userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CrashStatistics), args.Error(1)
}

func (m *MockErrorReportingService) GetSystemHealth() (*models.SystemHealth, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemHealth), args.Error(1)
}

func (m *MockErrorReportingService) UpdateConfiguration(config *services.ErrorReportingConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockErrorReportingService) UpdateConfigurationWithUserID(userID int, config map[string]interface{}) error {
	args := m.Called(userID, config)
	return args.Error(0)
}

func (m *MockErrorReportingService) GetConfiguration() (*services.ErrorReportingConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ErrorReportingConfig), args.Error(1)
}

func (m *MockErrorReportingService) GetConfigurationNoUserID() (*services.ErrorReportingConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ErrorReportingConfig), args.Error(1)
}

func (m *MockErrorReportingService) CleanupOldReports(olderThan time.Time) error {
	args := m.Called(olderThan)
	return args.Error(0)
}

func (m *MockErrorReportingService) CleanupOldReportsWithUserID(userID int, olderThan time.Time) error {
	args := m.Called(userID, olderThan)
	return args.Error(0)
}

func (m *MockErrorReportingService) ExportReports(userID int, filters *models.ExportFilters) ([]byte, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockErrorReportingAuthService for testing (different name to avoid collision)
type MockErrorReportingAuthService struct {
	mock.Mock
}

func (m *MockErrorReportingAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func TestErrorReportingHandler_ReportError(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		hasPermission  bool
		permissionErr  error
		requestData    *models.ErrorReportRequest
		mockResponse   *models.ErrorReport
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			hasPermission:  true,
			permissionErr:  nil,
			requestData:    &models.ErrorReportRequest{Level: "error", Message: "Test error"},
			mockResponse:   &models.ErrorReport{ID: 1, Level: "error"},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:          "Permission denied",
			userID:        1,
			hasPermission:  false,
			permissionErr:  nil,
			requestData:    &models.ErrorReportRequest{Level: "error", Message: "Test error"},
			mockResponse:   nil,
			serviceError:   nil,
			expectedStatus: 403,
			expectedError:  true,
		},
		{
			name:          "Service error",
			userID:        1,
			hasPermission:  true,
			permissionErr:  nil,
			requestData:    &models.ErrorReportRequest{Level: "error", Message: "Test error"},
			mockResponse:   nil,
			serviceError:   errors.New("service error"),
			expectedStatus: 500,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockErrorService := new(MockErrorReportingService)
			mockAuthService := new(MockErrorReportingAuthService)
			
			handler := &ErrorReportingHandler{
				errorReportingService: mockErrorService,
				authService:           mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionReportCreate).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				mockErrorService.On("ReportError", tt.userID, tt.requestData).Return(tt.mockResponse, tt.serviceError)
			}

			body, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/error-report", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.ReportError(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var report models.ErrorReport
				err := json.Unmarshal(rr.Body.Bytes(), &report)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, report.ID)
			}

			mockAuthService.AssertExpectations(t)
			mockErrorService.AssertExpectations(t)
		})
	}
}

func TestErrorReportingHandler_ReportCrash(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		hasPermission  bool
		permissionErr  error
		requestData    *models.CrashReportRequest
		mockResponse   *models.CrashReport
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			hasPermission:  true,
			permissionErr:  nil,
			requestData:    &models.CrashReportRequest{Signal: "SIGSEGV", Message: "Test crash"},
			mockResponse:   &models.CrashReport{ID: 1, Signal: "SIGSEGV"},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:          "Permission denied",
			userID:        1,
			hasPermission:  false,
			permissionErr:  nil,
			requestData:    &models.CrashReportRequest{Signal: "SIGSEGV", Message: "Test crash"},
			mockResponse:   nil,
			serviceError:   nil,
			expectedStatus: 403,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockErrorService := new(MockErrorReportingService)
			mockAuthService := new(MockErrorReportingAuthService)
			
			handler := &ErrorReportingHandler{
				errorReportingService: mockErrorService,
				authService:           mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionReportCreate).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				mockErrorService.On("ReportCrash", tt.userID, tt.requestData).Return(tt.mockResponse, tt.serviceError)
			}

			body, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/crash-report", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.ReportCrash(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var report models.CrashReport
				err := json.Unmarshal(rr.Body.Bytes(), &report)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, report.ID)
			}

			mockAuthService.AssertExpectations(t)
			mockErrorService.AssertExpectations(t)
		})
	}
}

func TestErrorReportingHandler_GetErrorReport(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		reportID       int
		hasPermission  bool
		permissionErr  error
		mockResponse   *models.ErrorReport
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			reportID:      1,
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:   &models.ErrorReport{ID: 1},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:          "Permission denied",
			userID:        1,
			reportID:      1,
			hasPermission:  false,
			permissionErr:  nil,
			mockResponse:   nil,
			serviceError:   nil,
			expectedStatus: 403,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockErrorService := new(MockErrorReportingService)
			mockAuthService := new(MockErrorReportingAuthService)
			
			handler := &ErrorReportingHandler{
				errorReportingService: mockErrorService,
				authService:           mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionReportView).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				mockErrorService.On("GetErrorReport", tt.reportID, tt.userID).Return(tt.mockResponse, tt.serviceError)
			}

			req := httptest.NewRequest("GET", "/error-report/"+strconv.Itoa(tt.reportID), nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(tt.reportID), "report_id": strconv.Itoa(tt.reportID)})
			rr := httptest.NewRecorder()

			handler.GetErrorReport(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var report models.ErrorReport
				err := json.Unmarshal(rr.Body.Bytes(), &report)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, report.ID)
			}

			mockAuthService.AssertExpectations(t)
			mockErrorService.AssertExpectations(t)
		})
	}
}

func TestErrorReportingHandler_NewErrorReportingHandler(t *testing.T) {
	mockErrorService := new(MockErrorReportingService)
	mockAuthService := new(MockErrorReportingAuthService)
	
	handler := NewErrorReportingHandler(mockErrorService, mockAuthService)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockErrorService, handler.errorReportingService)
	assert.Equal(t, mockAuthService, handler.authService)
}