package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"catalogizer/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConfigurationService for testing
type MockConfigurationService struct {
	mock.Mock
}

func (m *MockConfigurationService) GetWizardStep(stepID string) (*models.WizardStep, error) {
	args := m.Called(stepID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WizardStep), args.Error(1)
}

func (m *MockConfigurationService) ValidateWizardStep(stepID string, data map[string]interface{}) (*models.ValidationResult, error) {
	args := m.Called(stepID, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ValidationResult), args.Error(1)
}

func (m *MockConfigurationService) SaveWizardProgress(userID int, stepID string, data map[string]interface{}) error {
	args := m.Called(userID, stepID, data)
	return args.Error(0)
}

func (m *MockConfigurationService) GetWizardProgress(userID int) (*models.WizardProgress, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.WizardProgress), args.Error(1)
}

func (m *MockConfigurationService) CompleteWizard(userID int, data map[string]interface{}) (*models.SystemConfiguration, error) {
	args := m.Called(userID, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemConfiguration), args.Error(1)
}

func (m *MockConfigurationService) GetConfigurationSchema() (*models.ConfigurationSchema, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConfigurationSchema), args.Error(1)
}

func (m *MockConfigurationService) TestConfiguration(config *models.Configuration) (*models.ValidationResult, error) {
	args := m.Called(config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ValidationResult), args.Error(1)
}

func (m *MockConfigurationService) GetConfiguration() (*models.Configuration, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Configuration), args.Error(1)
}

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(tokenString string) (*models.User, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func TestConfigurationHandler_GetWizardStep(t *testing.T) {

	tests := []struct {
		name           string
		stepID         string
		mockResponse    *models.WizardStep
		mockError      error
		expectedStatus  int
		expectedError   bool
	}{
		{
			name:          "Valid step",
			stepID:        "step1",
			mockResponse:   &models.WizardStep{ID: "step1", Name: "Step 1"},
			mockError:      nil,
			expectedStatus: 200,
			expectedError:   false,
		},
		{
			name:          "Invalid step",
			stepID:        "invalid",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus:  404,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}

			mockConfigService.On("GetWizardStep", tt.stepID).Return(tt.mockResponse, tt.mockError)

			req := httptest.NewRequest("GET", "/wizard/step/"+tt.stepID, nil)
			req = mux.SetURLVars(req, map[string]string{"step_id": tt.stepID})
			rr := httptest.NewRecorder()

			handler.GetWizardStep(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var step models.WizardStep
				err := json.Unmarshal(rr.Body.Bytes(), &step)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, step.ID)
			}

			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_ValidateWizardStep(t *testing.T) {
	testData := map[string]interface{}{
		"field1": "value1",
		"field2": float64(123), // JSON unmarshaling converts numbers to float64
	}

	tests := []struct {
		name           string
		stepID         string
		requestData    map[string]interface{}
		mockResponse   *models.ValidationResult
		mockError      error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Valid data",
			stepID:        "step1",
			requestData:   testData,
			mockResponse:   &models.ValidationResult{IsValid: true},
			mockError:      nil,
			expectedStatus: 200,
			expectedError:   false,
		},
		{
			name:          "Invalid data",
			stepID:        "step1",
			requestData:   testData,
			mockResponse:   &models.ValidationResult{IsValid: false},
			mockError:      nil,
			expectedStatus: 200,
			expectedError:   false,
		},
		{
			name:          "Service error",
			stepID:        "step1",
			requestData:   testData,
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: 500,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}

			mockConfigService.On("ValidateWizardStep", tt.stepID, tt.requestData).Return(tt.mockResponse, tt.mockError)

			body, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/wizard/step/"+tt.stepID+"/validate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"step_id": tt.stepID})
			
			// Add user context for authentication
			ctx := context.WithValue(req.Context(), "user_id", int(1))
			req = req.WithContext(ctx)
			
			rr := httptest.NewRecorder()

			handler.ValidateWizardStep(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var result models.ValidationResult
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.IsValid, result.IsValid)
			}

			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_SaveWizardProgress(t *testing.T) {
	testData := map[string]interface{}{
		"field1": "value1",
		"field2": float64(123), // JSON unmarshaling converts numbers to float64
	}

	tests := []struct {
		name           string
		userID         int
		stepID         string
		requestData    map[string]interface{}
		mockError      error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			stepID:        "step1",
			requestData:   testData,
			mockError:     nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:          "Service error",
			userID:        1,
			stepID:        "step1",
			requestData:   map[string]interface{}{"field1": "value1", "field2": float64(123)},
			mockError:     assert.AnError,
			expectedStatus: 500,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test case
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}
			
			mockConfigService.On("SaveWizardProgress", tt.userID, tt.stepID, tt.requestData).Return(tt.mockError)

			body, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/wizard/step/"+tt.stepID+"/progress", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			req = mux.SetURLVars(req, map[string]string{"step_id": tt.stepID})
			rr := httptest.NewRecorder()

			handler.SaveWizardProgress(rr, req)

			t.Logf("Response status: %d", rr.Code)
			t.Logf("Response body: %s", rr.Body.String())

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Progress saved successfully", response["message"])
			}

			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_GetWizardProgress(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		mockResponse   *models.WizardProgress
		mockError      error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			mockResponse:   &models.WizardProgress{UserID: 1, CurrentStep: "step2"},
			mockError:      nil,
			expectedStatus: 200,
			expectedError:   false,
		},
		{
			name:          "Not found",
			userID:        1,
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: 404,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}

			mockConfigService.On("GetWizardProgress", tt.userID).Return(tt.mockResponse, tt.mockError)

			req := httptest.NewRequest("GET", "/wizard/progress", nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.GetWizardProgress(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var progress models.WizardProgress
				err := json.Unmarshal(rr.Body.Bytes(), &progress)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.UserID, progress.UserID)
			}

			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_CompleteWizard(t *testing.T) {
	testData := map[string]interface{}{
		"field1": "value1",
		"field2": float64(123), // JSON unmarshaling converts numbers to float64
	}

	tests := []struct {
		name           string
		userID         int
		requestData    map[string]interface{}
		mockResponse   *models.SystemConfiguration // Change to SystemConfiguration to match interface
		mockError      error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			userID:        1,
			requestData:   testData,
			mockResponse:   &models.SystemConfiguration{Version: "1.0"},
			mockError:      nil,
			expectedStatus: 200,
			expectedError:   false,
		},
		{
			name:          "Service error",
			userID:        1,
			requestData:   testData,
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedStatus: 500,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}

			mockConfigService.On("CompleteWizard", tt.userID, tt.requestData).Return(tt.mockResponse, tt.mockError)

			body, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/wizard/complete", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.CompleteWizard(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Wizard completed successfully", response["message"])
			}

			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_GetConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		hasPermission  bool
		permissionErr  error
		mockResponse   *models.ConfigurationSchema
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Success",
			userID:         1,
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:    &models.ConfigurationSchema{Version: "1.0"},
			serviceError:    nil,
			expectedStatus:  200,
			expectedError:   false,
		},
		{
			name:           "Permission denied",
			userID:         1,
			hasPermission:  false,
			permissionErr:  nil,
			mockResponse:    nil,
			serviceError:    nil,
			expectedStatus:  403,
			expectedError:   true,
		},
		{
			name:           "Permission error",
			userID:         1,
			hasPermission:  false,
			permissionErr:  assert.AnError,
			mockResponse:    nil,
			serviceError:    nil,
			expectedStatus:  403,
			expectedError:   true,
		},
		{
			name:           "Service error",
			userID:         1,
			hasPermission:  true,
			permissionErr:  nil,
			mockResponse:    nil,
			serviceError:    assert.AnError,
			expectedStatus:  500,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, "system.configure").Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				mockConfigService.On("GetConfigurationSchema").Return(tt.mockResponse, tt.serviceError)
			}

			req := httptest.NewRequest("GET", "/configuration", nil)
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.GetConfiguration(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var schema models.ConfigurationSchema
				err := json.Unmarshal(rr.Body.Bytes(), &schema)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.Version, schema.Version)
			}

			mockAuthService.AssertExpectations(t)
			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_TestConfiguration(t *testing.T) {
	testConfig := &models.Configuration{ID: "config1"}

	tests := []struct {
		name           string
		userID         int
		hasPermission  bool
		permissionErr  error
		requestData    *models.Configuration
		mockResponse   *models.ValidationResult
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Success",
			userID:         1,
			hasPermission:  true,
			permissionErr:  nil,
			requestData:    testConfig,
			mockResponse:   &models.ValidationResult{IsValid: true},
			serviceError:    nil,
			expectedStatus:  200,
			expectedError:   false,
		},
		{
			name:           "Permission denied",
			userID:         1,
			hasPermission:  false,
			permissionErr:  nil,
			requestData:    testConfig,
			mockResponse:    nil,
			serviceError:    nil,
			expectedStatus:  403,
			expectedError:   true,
		},
		{
			name:           "Service error",
			userID:         1,
			hasPermission:  true,
			permissionErr:  nil,
			requestData:    testConfig,
			mockResponse:    nil,
			serviceError:    assert.AnError,
			expectedStatus:  500,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockConfigService := new(MockConfigurationService)
			mockAuthService := new(MockAuthService)
			
			handler := &ConfigurationHandler{
				configurationService: mockConfigService,
				authService:          mockAuthService,
			}

			mockAuthService.On("CheckPermission", tt.userID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr)
			if tt.hasPermission && tt.permissionErr == nil {
				mockConfigService.On("TestConfiguration", tt.requestData).Return(tt.mockResponse, tt.serviceError)
			}

			body, _ := json.Marshal(tt.requestData)
			req := httptest.NewRequest("POST", "/configuration/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.WithValue(context.Background(), "user_id", tt.userID))
			rr := httptest.NewRecorder()

			handler.TestConfiguration(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var result models.ValidationResult
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.IsValid, result.IsValid)
			}

			mockAuthService.AssertExpectations(t)
			mockConfigService.AssertExpectations(t)
		})
	}
}

func TestConfigurationHandler_NewConfigurationHandler(t *testing.T) {
	mockConfigService := &MockConfigurationService{}
	mockAuthService := &MockAuthService{}
	
	handler := NewConfigurationHandler(mockConfigService, mockAuthService)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockConfigService, handler.configurationService)
	assert.Equal(t, mockAuthService, handler.authService)
}