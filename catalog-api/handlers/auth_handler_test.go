package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/models"
	"catalogizer/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(req models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	args := m.Called(req, ipAddress, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*models.LoginResponse, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) Logout(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockAuthService) LogoutAll(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) ChangePassword(userID int, currentPassword, newPassword string) error {
	args := m.Called(userID, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) CheckPermission(userID int, permission string) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

// Tests

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	expectedResponse := &models.LoginResponse{
		Token:        "jwt_token_here",
		RefreshToken: "refresh_token_here",
		User: &models.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		},
	}

	mockService.On("Login", loginReq, mock.Anything, mock.Anything).Return(expectedResponse, nil)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	var response models.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Token, response.Token)
	assert.Equal(t, expectedResponse.User.Username, response.User.Username)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockService.On("Login", loginReq, mock.Anything, mock.Anything).
		Return(nil, errors.New("invalid credentials"))

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidRequestBody(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_Login_InvalidMethod(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	refreshReq := struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: "valid_refresh_token",
	}

	expectedResponse := &models.LoginResponse{
		Token:        "new_jwt_token",
		RefreshToken: "new_refresh_token",
		User: &models.User{
			ID:       1,
			Username: "testuser",
		},
	}

	mockService.On("RefreshToken", refreshReq.RefreshToken).Return(expectedResponse, nil)

	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.RefreshToken(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Token, response.Token)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	refreshReq := struct {
		RefreshToken string `json:"refresh_token"`
	}{
		RefreshToken: "invalid_refresh_token",
	}

	mockService.On("RefreshToken", refreshReq.RefreshToken).
		Return(nil, errors.New("invalid refresh token"))

	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.RefreshToken(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	token := "valid_jwt_token"
	mockService.On("Logout", token).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.Logout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Logout_NoToken(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Logout_ServiceError(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	token := "valid_jwt_token"
	mockService.On("Logout", token).Return(errors.New("logout failed"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.Logout(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

// Helper function tests
func TestExtractToken(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedToken  string
	}{
		{
			name:          "Valid Bearer token",
			authHeader:    "Bearer test_token_123",
			expectedToken: "test_token_123",
		},
		{
			name:          "No Bearer prefix",
			authHeader:    "test_token_123",
			expectedToken: "",
		},
		{
			name:          "Empty header",
			authHeader:    "",
			expectedToken: "",
		},
		{
			name:          "Only Bearer",
			authHeader:    "Bearer",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token := extractToken(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			expectedIP: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			expectedIP: "192.168.1.2",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "192.168.1.3:12345",
			expectedIP: "192.168.1.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			ip := getClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// Benchmark tests
func BenchmarkAuthHandler_Login(b *testing.B) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	expectedResponse := &models.LoginResponse{
		Token:        "jwt_token_here",
		RefreshToken: "refresh_token_here",
		User: &models.User{
			ID:       1,
			Username: "testuser",
		},
	}

	mockService.On("Login", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedResponse, nil)

	body, _ := json.Marshal(loginReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.Login(rr, req)
	}
}
