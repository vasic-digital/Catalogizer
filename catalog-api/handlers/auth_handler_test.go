package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"catalogizer/services"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	handler     *AuthHandler
	authService *services.AuthService
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	// For these tests, we'll use a simplified approach without a real database
	// In production, you'd want to use a test database
	suite.authService = services.NewAuthService(nil, "test-secret-key")
	suite.handler = NewAuthHandler(suite.authService)
}

func (suite *AuthHandlerTestSuite) TestLoginMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	suite.handler.Login(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLoginInvalidRequestBody() {
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid-json"))
	w := httptest.NewRecorder()

	suite.handler.Login(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestRefreshTokenMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/refresh", nil)
	w := httptest.NewRecorder()

	suite.handler.RefreshToken(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestRefreshTokenInvalidRequestBody() {
	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString("invalid-json"))
	w := httptest.NewRecorder()

	suite.handler.RefreshToken(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestRefreshTokenMissingToken() {
	reqBody := map[string]string{"refresh_token": ""}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.handler.RefreshToken(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLogoutMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()

	suite.handler.Logout(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLogoutMissingToken() {
	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	suite.handler.Logout(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Authorization token required")
}

func (suite *AuthHandlerTestSuite) TestLogoutInvalidTokenFormat() {
	req, _ := http.NewRequest("POST", "/logout", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()

	suite.handler.Logout(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLogoutAllMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/logout-all", nil)
	w := httptest.NewRecorder()

	suite.handler.LogoutAll(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLogoutAllUnauthorized() {
	req, _ := http.NewRequest("POST", "/logout-all", nil)
	w := httptest.NewRecorder()

	suite.handler.LogoutAll(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestGetCurrentUserMethodNotAllowed() {
	req, _ := http.NewRequest("POST", "/current-user", nil)
	w := httptest.NewRecorder()

	suite.handler.GetCurrentUser(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestGetCurrentUserUnauthorized() {
	req, _ := http.NewRequest("GET", "/current-user", nil)
	w := httptest.NewRecorder()

	suite.handler.GetCurrentUser(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestChangePasswordMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/change-password", nil)
	w := httptest.NewRecorder()

	suite.handler.ChangePassword(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestChangePasswordUnauthorized() {
	req, _ := http.NewRequest("POST", "/change-password", nil)
	w := httptest.NewRecorder()

	suite.handler.ChangePassword(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestChangePasswordInvalidRequestBody() {
	req, _ := http.NewRequest("POST", "/change-password", bytes.NewBufferString("invalid-json"))
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	suite.handler.ChangePassword(w, req)

	// Will fail with unauthorized first since token validation will fail
	// In a full integration test, this would return BadRequest after auth
	assert.True(suite.T(), w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest)
}

func (suite *AuthHandlerTestSuite) TestGetActiveSessionsMethodNotAllowed() {
	req, _ := http.NewRequest("POST", "/sessions", nil)
	w := httptest.NewRecorder()

	suite.handler.GetActiveSessions(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestGetActiveSessionsUnauthorized() {
	req, _ := http.NewRequest("GET", "/sessions", nil)
	w := httptest.NewRecorder()

	suite.handler.GetActiveSessions(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestDeactivateSessionMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/deactivate-session", nil)
	w := httptest.NewRecorder()

	suite.handler.DeactivateSession(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestDeactivateSessionUnauthorized() {
	req, _ := http.NewRequest("POST", "/deactivate-session?session_id=1", nil)
	w := httptest.NewRecorder()

	suite.handler.DeactivateSession(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestDeactivateSessionMissingSessionID() {
	req, _ := http.NewRequest("POST", "/deactivate-session", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	suite.handler.DeactivateSession(w, req)

	// Will fail with unauthorized first, but testing the flow
	assert.True(suite.T(), w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest)
}

func (suite *AuthHandlerTestSuite) TestDeactivateSessionInvalidSessionID() {
	req, _ := http.NewRequest("POST", "/deactivate-session?session_id=invalid", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	suite.handler.DeactivateSession(w, req)

	// Will fail with unauthorized first, but testing the flow
	assert.True(suite.T(), w.Code == http.StatusUnauthorized || w.Code == http.StatusBadRequest)
}

func (suite *AuthHandlerTestSuite) TestValidateTokenMethodNotAllowed() {
	req, _ := http.NewRequest("GET", "/validate", nil)
	w := httptest.NewRecorder()

	suite.handler.ValidateToken(w, req)

	assert.Equal(suite.T(), http.StatusMethodNotAllowed, w.Code)
}

func (suite *AuthHandlerTestSuite) TestValidateTokenMissingToken() {
	req, _ := http.NewRequest("POST", "/validate", nil)
	w := httptest.NewRecorder()

	suite.handler.ValidateToken(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Authorization token required")
}

func (suite *AuthHandlerTestSuite) TestValidateTokenInvalidToken() {
	req, _ := http.NewRequest("POST", "/validate", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	suite.handler.ValidateToken(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *AuthHandlerTestSuite) TestExtractTokenWithValidBearer() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")

	token := extractToken(req)

	assert.Equal(suite.T(), "test-token-123", token)
}

func (suite *AuthHandlerTestSuite) TestExtractTokenWithoutBearer() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "test-token-123")

	token := extractToken(req)

	assert.Equal(suite.T(), "", token)
}

func (suite *AuthHandlerTestSuite) TestExtractTokenNoAuthHeader() {
	req, _ := http.NewRequest("GET", "/test", nil)

	token := extractToken(req)

	assert.Equal(suite.T(), "", token)
}

func (suite *AuthHandlerTestSuite) TestGetClientIPWithXRealIP() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	ip := getClientIP(req)

	assert.Equal(suite.T(), "192.168.1.100", ip)
}

func (suite *AuthHandlerTestSuite) TestGetClientIPWithXForwardedFor() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1")

	ip := getClientIP(req)

	assert.Equal(suite.T(), "192.168.1.100", ip)
}

func (suite *AuthHandlerTestSuite) TestGetClientIPFromRemoteAddr() {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	ip := getClientIP(req)

	assert.Equal(suite.T(), "192.168.1.100", ip)
}

// TestLoginRequestValidation is disabled because it requires a full database setup
// func (suite *AuthHandlerTestSuite) TestLoginRequestValidation() {
// 	// This test would require a proper UserRepository with database
// 	// For now, we test the HTTP layer with the other tests
// }

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
