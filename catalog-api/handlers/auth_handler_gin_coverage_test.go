package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newGinAuthTestHandler returns an AuthHandler backed by an AuthService
// with a nil DB. All service calls will return "database not initialized"
// errors — ideal for exercising the handler error paths without a test DB.
func newGinAuthTestHandler() *AuthHandler {
	return NewAuthHandler(services.NewAuthService(nil, "test-secret"))
}

func newGinContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	c.Request = req
	return c, w
}

func TestAuthHandler_LoginGin_InvalidJSON(t *testing.T) {
	h := newGinAuthTestHandler()
	c, w := newGinContext("POST", "/api/v1/auth/login", "{not json")
	h.LoginGin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request format")
}

// LoginGin's success path requires a real DB-backed AuthService — covered
// by the existing auth_handler_test.go suite. Here we only exercise the
// BadRequest path via TestAuthHandler_LoginGin_InvalidJSON.

func TestAuthHandler_RefreshTokenGin_InvalidJSON(t *testing.T) {
	h := newGinAuthTestHandler()
	c, w := newGinContext("POST", "/api/v1/auth/refresh", "[[")
	h.RefreshTokenGin(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

// RefreshTokenGin's service path requires a real DB — BadRequest path
// is covered by TestAuthHandler_RefreshTokenGin_InvalidJSON above.

func TestAuthHandler_LogoutGin_MissingToken(t *testing.T) {
	h := newGinAuthTestHandler()
	c, w := newGinContext("POST", "/api/v1/auth/logout", "")
	h.LogoutGin(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token required")
}

func TestAuthHandler_GetCurrentUserGin_MissingToken(t *testing.T) {
	h := newGinAuthTestHandler()
	c, w := newGinContext("GET", "/api/v1/auth/me", "")
	h.GetCurrentUserGin(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetAuthStatusGin_NoTokenReturnsNotAuthenticated(t *testing.T) {
	h := newGinAuthTestHandler()
	c, w := newGinContext("GET", "/api/v1/auth/status", "")
	h.GetAuthStatusGin(c)
	// No token → 200 with authenticated=false
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, false, resp["authenticated"])
}

func TestAuthHandler_GetPermissionsGin_MissingToken(t *testing.T) {
	h := newGinAuthTestHandler()
	c, w := newGinContext("GET", "/api/v1/auth/permissions", "")
	h.GetPermissionsGin(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExtractTokenFromGin_Bearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer abc123")
	require.Equal(t, "abc123", extractTokenFromGin(c))
}

func TestExtractTokenFromGin_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	require.Equal(t, "", extractTokenFromGin(c))
}

func TestExtractTokenFromGin_NonBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	// Function only recognizes "Bearer " prefix; anything else returns empty.
	got := extractTokenFromGin(c)
	require.Equal(t, "", got)
}
