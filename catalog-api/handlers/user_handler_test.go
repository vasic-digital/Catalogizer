package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
	"catalogizer/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*models.User, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Count() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) GetRole(roleID int) (*models.Role, error) {
	args := m.Called(roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

// MockAuthServiceForUser is a mock implementation of AuthService for UserHandler
type MockAuthServiceForUser struct {
	mock.Mock
}

func (m *MockAuthServiceForUser) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthServiceForUser) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockAuthServiceForUser) GenerateSecureToken(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

func (m *MockAuthServiceForUser) HashData(data string) string {
	args := m.Called(data)
	return args.String(0)
}

func (m *MockAuthServiceForUser) GetCurrentUser(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthServiceForUser) ResetPassword(userID int, newPassword string) error {
	args := m.Called(userID, newPassword)
	return args.Error(0)
}

func (m *MockAuthServiceForUser) LockAccount(userID int, lockUntil time.Time) error {
	args := m.Called(userID, lockUntil)
	return args.Error(0)
}

func (m *MockAuthServiceForUser) UnlockAccount(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

// Test setup helpers

func setupUserHandler() (*UserHandler, *MockUserRepository, *MockAuthServiceForUser) {
	mockUserRepo := new(MockUserRepository)
	mockAuthService := new(MockAuthServiceForUser)

	handler := &UserHandler{
		userRepo:    (*repository.UserRepository)(unsafe_cast_user_repo(mockUserRepo)),
		authService: (*services.AuthService)(unsafe_cast_auth_service(mockAuthService)),
	}

	return handler, mockUserRepo, mockAuthService
}

// Helper function to bypass type safety for testing
func unsafe_cast_user_repo(m *MockUserRepository) interface{} {
	return m
}

func unsafe_cast_auth_service(m *MockAuthServiceForUser) interface{} {
	return m
}

// CreateUser Tests

func TestUserHandler_CreateUser_Success(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserCreate).Return(true, nil)
	mockAuthService.On("ValidatePassword", "SecureP@ss123").Return(nil)
	mockAuthService.On("GenerateSecureToken", 16).Return("random_salt_16bytes", nil)
	mockAuthService.On("HashData", "SecureP@ss123random_salt_16bytes").Return("hashed_password")

	mockUserRepo.On("Create", mock.MatchedBy(func(u *models.User) bool {
		return u.Username == "newuser" && u.Email == "newuser@example.com"
	})).Return(123, nil)

	isActive := true
	reqBody := models.CreateUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "SecureP@ss123",
		RoleID:    2,
		FirstName: stringPtr("New"),
		LastName:  stringPtr("User"),
		IsActive:  &isActive,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response models.User
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 123, response.ID)
	assert.Equal(t, "newuser", response.Username)
	assert.Empty(t, response.PasswordHash, "Password hash should be cleared")
	assert.Empty(t, response.Salt, "Salt should be cleared")

	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_CreateUser_Unauthorized(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	mockAuthService.On("GetCurrentUser", "").Return(nil, models.ErrUnauthorized)

	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InsufficientPermissions(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "user"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserCreate).Return(false, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_InvalidRequestBody(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserCreate).Return(true, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_WeakPassword(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserCreate).Return(true, nil)
	mockAuthService.On("ValidatePassword", "weak").Return(errors.New("password too weak"))

	reqBody := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "weak",
		RoleID:   2,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestUserHandler_CreateUser_DuplicateUsername(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserCreate).Return(true, nil)
	mockAuthService.On("ValidatePassword", "SecureP@ss123").Return(nil)
	mockAuthService.On("GenerateSecureToken", 16).Return("random_salt", nil)
	mockAuthService.On("HashData", mock.Anything).Return("hashed_password")

	mockUserRepo.On("Create", mock.Anything).Return(0, errors.New("UNIQUE constraint failed: users.username"))

	reqBody := models.CreateUserRequest{
		Username: "existinguser",
		Email:    "newuser@example.com",
		Password: "SecureP@ss123",
		RoleID:   2,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_CreateUser_MethodNotAllowed(t *testing.T) {
	handler, _, _ := setupUserHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

// GetUser Tests

func TestUserHandler_GetUser_OwnProfile(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "testuser"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)

	mockUser := &models.User{
		ID:       123,
		Username: "testuser",
		Email:    "test@example.com",
		RoleID:   2,
	}

	mockRole := &models.Role{
		ID:   2,
		Name: "User",
	}

	mockUserRepo.On("GetByID", 123).Return(mockUser, nil)
	mockUserRepo.On("GetRole", 2).Return(mockRole, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.User
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 123, response.ID)
	assert.Equal(t, "testuser", response.Username)
	assert.NotNil(t, response.Role)

	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_GetUser_OtherUserWithPermission(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserView).Return(true, nil)

	mockUser := &models.User{
		ID:       456,
		Username: "otheruser",
		Email:    "other@example.com",
		RoleID:   2,
	}

	mockRole := &models.Role{
		ID:   2,
		Name: "User",
	}

	mockUserRepo.On("GetByID", 456).Return(mockUser, nil)
	mockUserRepo.On("GetRole", 2).Return(mockRole, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/456", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_GetUser_InsufficientPermissions(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "user"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 123, models.PermissionUserView).Return(false, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/456", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserView).Return(true, nil)

	mockUserRepo.On("GetByID", 999).Return(nil, errors.New("user not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/invalid", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockAuthService.AssertExpectations(t)
}

// UpdateUser Tests

func TestUserHandler_UpdateUser_OwnProfile(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "testuser"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)

	existingUser := &models.User{
		ID:       123,
		Username: "testuser",
		Email:    "old@example.com",
		RoleID:   2,
	}

	mockUserRepo.On("GetByID", 123).Return(existingUser, nil)
	mockUserRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "new@example.com"
	})).Return(nil)

	mockRole := &models.Role{ID: 2, Name: "User"}
	mockUserRepo.On("GetRole", 2).Return(mockRole, nil)

	newEmail := "new@example.com"
	reqBody := models.UpdateUserRequest{
		Email: &newEmail,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/users/123", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_WithSettings(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "testuser"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)

	existingUser := &models.User{
		ID:       123,
		Username: "testuser",
		RoleID:   2,
	}

	mockUserRepo.On("GetByID", 123).Return(existingUser, nil)
	mockUserRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.Settings != ""
	})).Return(nil)

	mockRole := &models.Role{ID: 2, Name: "User"}
	mockUserRepo.On("GetRole", 2).Return(mockRole, nil)

	settings := map[string]interface{}{
		"theme":         "dark",
		"notifications": true,
	}
	reqBody := models.UpdateUserRequest{
		Settings: &settings,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/users/123", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_AdminChangingRole(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserUpdate).Return(true, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserManage).Return(true, nil)

	existingUser := &models.User{
		ID:       456,
		Username: "targetuser",
		RoleID:   2,
	}

	newRoleID := 3
	isActive := false

	mockUserRepo.On("GetByID", 456).Return(existingUser, nil)
	mockUserRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
		return u.RoleID == 3 && u.IsActive == false
	})).Return(nil)

	mockRole := &models.Role{ID: 3, Name: "Moderator"}
	mockUserRepo.On("GetRole", 3).Return(mockRole, nil)

	reqBody := models.UpdateUserRequest{
		RoleID:   &newRoleID,
		IsActive: &isActive,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/users/456", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// DeleteUser Tests

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserDelete).Return(true, nil)

	mockUserRepo.On("Delete", 456).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/456", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_CannotDeleteSelf(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "user"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 123, models.PermissionUserDelete).Return(true, nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/123", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Cannot delete your own account")
	mockAuthService.AssertExpectations(t)
}

// ListUsers Tests

func TestUserHandler_ListUsers_Success(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserView).Return(true, nil)

	mockUsers := []*models.User{
		{ID: 1, Username: "user1", RoleID: 1},
		{ID: 2, Username: "user2", RoleID: 2},
	}

	mockRole1 := &models.Role{ID: 1, Name: "Admin"}
	mockRole2 := &models.Role{ID: 2, Name: "User"}

	mockUserRepo.On("List", 50, 0).Return(mockUsers, nil)
	mockUserRepo.On("GetRole", 1).Return(mockRole1, nil)
	mockUserRepo.On("GetRole", 2).Return(mockRole2, nil)
	mockUserRepo.On("Count").Return(2, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["total_count"])

	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestUserHandler_ListUsers_WithPagination(t *testing.T) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserView).Return(true, nil)

	mockUsers := []*models.User{}

	mockUserRepo.On("List", 20, 40).Return(mockUsers, nil)
	mockUserRepo.On("Count").Return(100, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users?limit=20&offset=40", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, float64(20), response["limit"])
	assert.Equal(t, float64(40), response["offset"])

	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// ResetPassword Tests

func TestUserHandler_ResetPassword_Success(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserManage).Return(true, nil)
	mockAuthService.On("ValidatePassword", "NewSecureP@ss123").Return(nil)
	mockAuthService.On("ResetPassword", 456, "NewSecureP@ss123").Return(nil)

	reqBody := map[string]string{
		"new_password": "NewSecureP@ss123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users/456/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ResetPassword(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
}

// LockAccount Tests

func TestUserHandler_LockAccount_Success(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserManage).Return(true, nil)
	mockAuthService.On("LockAccount", 456, mock.AnythingOfType("time.Time")).Return(nil)

	lockUntil := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	reqBody := map[string]string{
		"lock_until": lockUntil,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users/456/lock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.LockAccount(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
}

func TestUserHandler_LockAccount_CannotLockSelf(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "user"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 123, models.PermissionUserManage).Return(true, nil)

	lockUntil := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	reqBody := map[string]string{
		"lock_until": lockUntil,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/users/123/lock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer valid_token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.LockAccount(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Cannot lock your own account")
	mockAuthService.AssertExpectations(t)
}

// UnlockAccount Tests

func TestUserHandler_UnlockAccount_Success(t *testing.T) {
	handler, _, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", "valid_token").Return(currentUser, nil)
	mockAuthService.On("CheckPermission", 1, models.PermissionUserManage).Return(true, nil)
	mockAuthService.On("UnlockAccount", 456).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/users/456/unlock", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	rr := httptest.NewRecorder()

	handler.UnlockAccount(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockAuthService.AssertExpectations(t)
}

// Helper Tests

func TestParseTime_ValidRFC3339(t *testing.T) {
	timeStr := "2024-01-15T10:30:00Z"
	parsedTime, err := parseTime(timeStr)

	assert.NoError(t, err)
	assert.Equal(t, 2024, parsedTime.Year())
	assert.Equal(t, time.January, parsedTime.Month())
	assert.Equal(t, 15, parsedTime.Day())
}

func TestParseTime_InvalidFormat(t *testing.T) {
	timeStr := "2024-01-15 10:30:00"
	_, err := parseTime(timeStr)

	assert.Error(t, err)
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

// Benchmark Tests

func BenchmarkUserHandler_CreateUser(b *testing.B) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", mock.Anything).Return(currentUser, nil)
	mockAuthService.On("CheckPermission", mock.Anything, mock.Anything).Return(true, nil)
	mockAuthService.On("ValidatePassword", mock.Anything).Return(nil)
	mockAuthService.On("GenerateSecureToken", mock.Anything).Return("salt", nil)
	mockAuthService.On("HashData", mock.Anything).Return("hashed")

	mockUserRepo.On("Create", mock.Anything).Return(123, nil)

	reqBody := models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecureP@ss123",
		RoleID:   2,
	}

	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer token")
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.CreateUser(rr, req)
	}
}

func BenchmarkUserHandler_GetUser(b *testing.B) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 123, Username: "testuser"}

	mockAuthService.On("GetCurrentUser", mock.Anything).Return(currentUser, nil)

	mockUser := &models.User{ID: 123, Username: "testuser", RoleID: 2}
	mockRole := &models.Role{ID: 2, Name: "User"}

	mockUserRepo.On("GetByID", 123).Return(mockUser, nil)
	mockUserRepo.On("GetRole", 2).Return(mockRole, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/users/123", nil)
		req.Header.Set("Authorization", "Bearer token")
		rr := httptest.NewRecorder()
		handler.GetUser(rr, req)
	}
}

func BenchmarkUserHandler_ListUsers(b *testing.B) {
	handler, mockUserRepo, mockAuthService := setupUserHandler()

	currentUser := &models.User{ID: 1, Username: "admin"}

	mockAuthService.On("GetCurrentUser", mock.Anything).Return(currentUser, nil)
	mockAuthService.On("CheckPermission", mock.Anything, mock.Anything).Return(true, nil)

	mockUsers := []*models.User{
		{ID: 1, Username: "user1", RoleID: 1},
		{ID: 2, Username: "user2", RoleID: 2},
	}

	mockRole := &models.Role{ID: 1, Name: "Admin"}

	mockUserRepo.On("List", mock.Anything, mock.Anything).Return(mockUsers, nil)
	mockUserRepo.On("GetRole", mock.Anything).Return(mockRole, nil)
	mockUserRepo.On("Count").Return(2, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set("Authorization", "Bearer token")
		rr := httptest.NewRecorder()
		handler.ListUsers(rr, req)
	}
}
