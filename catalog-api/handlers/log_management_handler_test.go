package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"catalogizer/models"
	"catalogizer/services"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogManagementService for testing
type MockLogManagementService struct {
	mock.Mock
}

func (m *MockLogManagementService) CollectLogs(userID int, request *models.LogCollectionRequest) (*models.LogCollection, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogCollection), args.Error(1)
}

func (m *MockLogManagementService) GetLogCollection(collectionID int, userID int) (*models.LogCollection, error) {
	args := m.Called(collectionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogCollection), args.Error(1)
}

func (m *MockLogManagementService) GetLogCollectionsByUser(userID int, limit, offset int) ([]models.LogCollection, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.LogCollection), args.Error(1)
}

func (m *MockLogManagementService) GetLogEntries(collectionID int, userID int, filters *models.LogEntryFilters) ([]models.LogEntry, error) {
	args := m.Called(collectionID, userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.LogEntry), args.Error(1)
}

func (m *MockLogManagementService) CreateLogShare(userID int, request *models.LogShareRequest) (*models.LogShare, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogShare), args.Error(1)
}

func (m *MockLogManagementService) GetLogShare(token string) (*models.LogShare, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogShare), args.Error(1)
}

func (m *MockLogManagementService) RevokeLogShare(shareID int, userID int) error {
	args := m.Called(shareID, userID)
	return args.Error(0)
}

func (m *MockLogManagementService) ExportLogs(collectionID int, userID int, format string) ([]byte, error) {
	args := m.Called(collectionID, userID, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockLogManagementService) StreamLogs(userID int, filters *models.LogStreamFilters) (<-chan models.LogEntry, error) {
	args := m.Called(userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan models.LogEntry), args.Error(1)
}

func (m *MockLogManagementService) AnalyzeLogs(collectionID int, userID int) (*models.LogAnalysis, error) {
	args := m.Called(collectionID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogAnalysis), args.Error(1)
}

func (m *MockLogManagementService) GetLogStatistics(userID int) (*models.LogStatistics, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LogStatistics), args.Error(1)
}

func (m *MockLogManagementService) GetConfiguration() *services.LogManagementConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*services.LogManagementConfig)
}

func (m *MockLogManagementService) UpdateConfiguration(config *services.LogManagementConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockLogManagementService) CleanupOldLogs() error {
	args := m.Called()
	return args.Error(0)
}

// MockAuthService for testing
type MockLogManagementAuthService struct {
	mock.Mock
}

func (m *MockLogManagementAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

// TestLogManagementHandler_CreateLogCollection
func TestLogManagementHandler_CreateLogCollection(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		requestBody    interface{}
		hasPermission  bool
		permissionErr  error
		mockResponse   *models.LogCollection
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Success",
			userID:         1,
			requestBody:    models.LogCollectionRequest{Name: "Test Collection", Description: "Test Description"},
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:   &models.LogCollection{ID: 1, Name: "Test Collection"},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "Permission denied",
			userID:         1,
			requestBody:    models.LogCollectionRequest{Name: "Test Collection"},
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
			mockLogService := new(MockLogManagementService)
			mockAuthService := new(MockLogManagementAuthService)

			handler := &LogManagementHandler{
				logManagementService: mockLogService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				bodyBytes, _ := json.Marshal(tt.requestBody)
				var request models.LogCollectionRequest
				_ = json.Unmarshal(bodyBytes, &request)
				mockLogService.On("CollectLogs", tt.userID, &request).Return(tt.mockResponse, tt.serviceError)
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/log-collections", bytes.NewReader(bodyBytes))
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.CreateLogCollection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var collection models.LogCollection
				err := json.Unmarshal(rr.Body.Bytes(), &collection)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, collection.ID)
			}

			mockAuthService.AssertExpectations(t)
			mockLogService.AssertExpectations(t)
		})
	}
}

// TestLogManagementHandler_GetLogCollection
func TestLogManagementHandler_GetLogCollection(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		collectionID   int
		hasPermission  bool
		permissionErr  error
		mockResponse   *models.LogCollection
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Success",
			userID:         1,
			collectionID:   1,
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:   &models.LogCollection{ID: 1, Name: "Test Collection"},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "Permission denied",
			userID:         1,
			collectionID:   1,
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
			mockLogService := new(MockLogManagementService)
			mockAuthService := new(MockLogManagementAuthService)

			handler := &LogManagementHandler{
				logManagementService: mockLogService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				mockLogService.On("GetLogCollection", tt.collectionID, tt.userID).Return(tt.mockResponse, tt.serviceError)
			}

			req := httptest.NewRequest("GET", "/log-collections/"+strconv.Itoa(tt.collectionID), nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			req = mux.SetURLVars(req, map[string]string{"id": strconv.Itoa(tt.collectionID)})
			rr := httptest.NewRecorder()

			handler.GetLogCollection(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var collection models.LogCollection
				err := json.Unmarshal(rr.Body.Bytes(), &collection)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, collection.ID)
			}

			mockAuthService.AssertExpectations(t)
			mockLogService.AssertExpectations(t)
		})
	}
}

// TestLogManagementHandler_ListLogCollections
func TestLogManagementHandler_ListLogCollections(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		queryParams    string
		hasPermission  bool
		permissionErr  error
		mockResponse   []models.LogCollection
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Success with default pagination",
			userID:         1,
			queryParams:    "",
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:   []models.LogCollection{{ID: 1, Name: "Collection 1"}, {ID: 2, Name: "Collection 2"}},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "Success with custom pagination",
			userID:         1,
			queryParams:    "?limit=10&offset=5",
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:   []models.LogCollection{{ID: 3, Name: "Collection 3"}},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "Permission denied",
			userID:         1,
			queryParams:    "",
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
			mockLogService := new(MockLogManagementService)
			mockAuthService := new(MockLogManagementAuthService)

			handler := &LogManagementHandler{
				logManagementService: mockLogService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				// Parse expected limit and offset from query params
				limit := 20 // default
				offset := 0 // default
				if tt.queryParams != "" {
					// This is a simplified parsing - in a real test, you'd parse the query string properly
					if strings.Contains(tt.queryParams, "limit=10") {
						limit = 10
					}
					if strings.Contains(tt.queryParams, "offset=5") {
						offset = 5
					}
				}
				mockLogService.On("GetLogCollectionsByUser", tt.userID, limit, offset).Return(tt.mockResponse, tt.serviceError)
			}

			req := httptest.NewRequest("GET", "/log-collections"+tt.queryParams, nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.ListLogCollections(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "collections")
				assert.Contains(t, response, "limit")
				assert.Contains(t, response, "offset")
			}

			mockLogService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

// TestLogManagementHandler_GetLogEntries
func TestLogManagementHandler_GetLogEntries(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		collectionID   string
		queryParams    string
		hasPermission  bool
		permissionErr  error
		mockResponse   []models.LogEntry
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Success",
			userID:         1,
			collectionID:   "1",
			queryParams:    "?level=error&component=api",
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:   []models.LogEntry{{ID: 1, Level: "error", Message: "Test error"}},
			serviceError:   nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:           "Invalid collection ID",
			userID:         1,
			collectionID:   "invalid",
			queryParams:    "",
			hasPermission:  true, // This won't be checked due to early return
			permissionErr:  nil,
			mockResponse:   nil,
			serviceError:   nil,
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name:           "Permission denied",
			userID:         1,
			collectionID:   "1",
			queryParams:    "",
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
			mockLogService := new(MockLogManagementService)
			mockAuthService := new(MockLogManagementAuthService)

			handler := &LogManagementHandler{
				logManagementService: mockLogService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr).Maybe()
			if tt.hasPermission && tt.permissionErr == nil && tt.collectionID != "invalid" {
				collectionID, _ := strconv.Atoi(tt.collectionID)
				mockLogService.On("GetLogEntries", collectionID, tt.userID, mock.AnythingOfType("*models.LogEntryFilters")).Return(tt.mockResponse, tt.serviceError)
			}

			req := httptest.NewRequest("GET", "/log-collections/"+tt.collectionID+"/entries"+tt.queryParams, nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			req = mux.SetURLVars(req, map[string]string{"id": tt.collectionID})
			rr := httptest.NewRecorder()

			handler.GetLogEntries(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "entries")
				assert.Contains(t, response, "filters")
			}

			mockLogService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

// TestLogManagementHandler_NewLogManagementHandler
func TestLogManagementHandler_NewLogManagementHandler(t *testing.T) {
	mockLogService := new(MockLogManagementService)
	mockAuthService := new(MockLogManagementAuthService)

	handler := NewLogManagementHandler(mockLogService, mockAuthService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockLogService, handler.logManagementService)
	assert.Equal(t, mockAuthService, handler.authService)
}
