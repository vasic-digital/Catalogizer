package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"

	"catalogizer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleUserService for testing
type MockRoleUserService struct {
	mock.Mock
}

func (m *MockRoleUserService) CreateRole(role *models.Role) (int, error) {
	args := m.Called(role)
	return args.Int(0), args.Error(1)
}

func (m *MockRoleUserService) GetRole(roleID int) (*models.Role, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleUserService) UpdateRole(role *models.Role) error {
	args := m.Called(role)
	return args.Error(0)
}

func (m *MockRoleUserService) DeleteRole(roleID int) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func (m *MockRoleUserService) ListRoles() ([]models.Role, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Role), args.Error(1)
}

// MockRoleAuthService for testing
type MockRoleAuthService struct {
	mock.Mock
}

func (m *MockRoleAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleAuthService) GetCurrentUser(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// TestRoleHandler_CreateRole
func TestRoleHandler_CreateRole(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authToken      string
		currentUser    *models.User
		hasPermission  bool
		permissionErr  error
		requestBody    interface{}
		mockRoleID     int
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			method:        "POST",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "admin"},
			hasPermission: true,
			permissionErr: nil,
			requestBody: models.CreateRoleRequest{
				Name:        "Test Role",
				Description: func() *string { s := "Test Description"; return &s }(),
				Permissions: []string{"user.view"},
			},
			mockRoleID:     1,
			serviceError:    nil,
			expectedStatus: 201,
			expectedError:   false,
		},
		{
			name:          "Method not allowed",
			method:        "GET",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "admin"},
			hasPermission: true,
			permissionErr: nil,
			requestBody: models.CreateRoleRequest{
				Name: "Test Role",
			},
			mockRoleID:     0,
			serviceError:    nil,
			expectedStatus: 405,
			expectedError:   true,
		},
		{
			name:          "Unauthorized",
			method:        "POST",
			authToken:     "",
			currentUser:   nil,
			hasPermission: false,
			permissionErr: models.ErrUnauthorized,
			requestBody: models.CreateRoleRequest{
				Name: "Test Role",
			},
			mockRoleID:     0,
			serviceError:    nil,
			expectedStatus: 401,
			expectedError:   true,
		},
		{
			name:          "Permission denied",
			method:        "POST",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "user"},
			hasPermission: false,
			permissionErr: nil,
			requestBody: models.CreateRoleRequest{
				Name: "Test Role",
			},
			mockRoleID:     0,
			serviceError:    nil,
			expectedStatus: 403,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockUserService := new(MockRoleUserService)
			mockAuthService := new(MockRoleAuthService)
			
			handler := &RoleHandler{
				userRepo:    mockUserService,
				authService: mockAuthService,
			}

			if tt.authToken != "" && tt.currentUser != nil {
				mockAuthService.On("GetCurrentUser", tt.authToken).Return(tt.currentUser, tt.permissionErr).Maybe()
			}
			
			if tt.currentUser != nil && tt.permissionErr == nil {
				mockAuthService.On("CheckPermission", tt.currentUser.ID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr).Maybe()
			}

			if tt.hasPermission && tt.permissionErr == nil && tt.method == "POST" {
				mockUserService.On("CreateRole", mock.AnythingOfType("*models.Role")).Return(tt.mockRoleID, tt.serviceError)
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/api/roles", bytes.NewReader(bodyBytes))
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}
			rr := httptest.NewRecorder()

			handler.CreateRole(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError && tt.expectedStatus == 201 {
				var role models.Role
				err := json.Unmarshal(rr.Body.Bytes(), &role)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockRoleID, role.ID)
			}

			mockUserService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

// TestRoleHandler_GetRole
func TestRoleHandler_GetRole(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		authToken      string
		currentUser    *models.User
		hasPermission  bool
		permissionErr  error
		roleID         string
		mockRole       *models.Role
		serviceError   error
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "Success",
			method:        "GET",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "admin"},
			hasPermission: true,
			permissionErr: nil,
			roleID:        "1",
			mockRole:      &models.Role{ID: 1, Name: "Test Role"},
			serviceError:  nil,
			expectedStatus: 200,
			expectedError:  false,
		},
		{
			name:          "Method not allowed",
			method:        "POST",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "admin"},
			hasPermission: true,
			permissionErr: nil,
			roleID:        "1",
			mockRole:      nil,
			serviceError:  nil,
			expectedStatus: 405,
			expectedError:  true,
		},
		{
			name:          "Invalid role ID",
			method:        "GET",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "admin"},
			hasPermission: true,
			permissionErr: nil,
			roleID:        "invalid",
			mockRole:      nil,
			serviceError:  nil,
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name:          "Permission denied",
			method:        "GET",
			authToken:     "valid-token",
			currentUser:   &models.User{ID: 1, Username: "user"},
			hasPermission: false,
			permissionErr: nil,
			roleID:        "1",
			mockRole:      nil,
			serviceError:  nil,
			expectedStatus: 403,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mocks for each test to avoid contamination
			mockUserService := new(MockRoleUserService)
			mockAuthService := new(MockRoleAuthService)
			
			handler := &RoleHandler{
				userRepo:    mockUserService,
				authService: mockAuthService,
			}

			if tt.authToken != "" && tt.currentUser != nil {
				mockAuthService.On("GetCurrentUser", tt.authToken).Return(tt.currentUser, tt.permissionErr).Maybe()
			}
			
			if tt.currentUser != nil && tt.permissionErr == nil {
				mockAuthService.On("CheckPermission", tt.currentUser.ID, models.PermissionSystemAdmin).Return(tt.hasPermission, tt.permissionErr).Maybe()
			}

			if tt.hasPermission && tt.permissionErr == nil && tt.method == "GET" && tt.roleID != "invalid" {
				roleID, _ := strconv.Atoi(tt.roleID)
				mockUserService.On("GetRole", roleID).Return(tt.mockRole, tt.serviceError)
			}

			req := httptest.NewRequest(tt.method, "/api/roles/"+tt.roleID, nil)
			if tt.authToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authToken)
			}
			rr := httptest.NewRecorder()

			handler.GetRole(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError && tt.expectedStatus == 200 {
				var role models.Role
				err := json.Unmarshal(rr.Body.Bytes(), &role)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockRole.ID, role.ID)
			}

			mockUserService.AssertExpectations(t)
			mockAuthService.AssertExpectations(t)
		})
	}
}

// TestRoleHandler_NewRoleHandler
func TestRoleHandler_NewRoleHandler(t *testing.T) {
	mockUserService := new(MockRoleUserService)
	mockAuthService := new(MockRoleAuthService)

	handler := NewRoleHandler(mockUserService, mockAuthService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUserService, handler.userRepo)
	assert.Equal(t, mockAuthService, handler.authService)
}