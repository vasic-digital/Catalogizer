package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type InternalAuthHandlerTestSuite struct {
	suite.Suite
	handler *AuthHandler
	router  *gin.Engine
	logger  *zap.Logger
}

func (suite *InternalAuthHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.logger = zap.NewNop()
}

func (suite *InternalAuthHandlerTestSuite) SetupTest() {
	// Initialize handler with nil auth service to test validation paths
	suite.handler = NewAuthHandler(nil, suite.logger)

	suite.router = gin.New()
	suite.router.POST("/api/v1/auth/login", suite.handler.Login)
	suite.router.POST("/api/v1/auth/logout", suite.handler.Logout)
	suite.router.POST("/api/v1/auth/register", suite.handler.Register)
	suite.router.GET("/api/v1/auth/profile", suite.handler.GetProfile)
	suite.router.PUT("/api/v1/auth/profile", suite.handler.UpdateProfile)
	suite.router.POST("/api/v1/auth/change-password", suite.handler.ChangePassword)
	suite.router.GET("/api/v1/auth/admin/users", suite.handler.ListUsers)
	suite.router.GET("/api/v1/auth/admin/users/:id", suite.handler.GetUser)
	suite.router.PUT("/api/v1/auth/admin/users/:id", suite.handler.UpdateUser)
	suite.router.GET("/api/v1/auth/status", suite.handler.GetAuthStatus)
	suite.router.GET("/api/v1/auth/permissions", suite.handler.GetPermissions)
	suite.router.GET("/api/v1/auth/init-status", suite.handler.GetInitStatus)
}

// Constructor tests

func (suite *InternalAuthHandlerTestSuite) TestNewAuthHandler() {
	handler := NewAuthHandler(nil, suite.logger)
	assert.NotNil(suite.T(), handler)
	assert.Nil(suite.T(), handler.authService)
	assert.NotNil(suite.T(), handler.logger)
}

// Login tests

func (suite *InternalAuthHandlerTestSuite) TestLogin_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", resp["error"])
}

func (suite *InternalAuthHandlerTestSuite) TestLogin_EmptyBody() {
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *InternalAuthHandlerTestSuite) TestLogin_MissingFields() {
	body := `{"username": ""}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Logout tests

func (suite *InternalAuthHandlerTestSuite) TestLogout_MissingAuthHeader() {
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Missing authorization header", resp["error"])
}

// Register tests

func (suite *InternalAuthHandlerTestSuite) TestRegister_InvalidJSON() {
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", resp["error"])
}

func (suite *InternalAuthHandlerTestSuite) TestRegister_EmptyBody() {
	req := httptest.NewRequest("POST", "/api/v1/auth/register", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// GetProfile tests

func (suite *InternalAuthHandlerTestSuite) TestGetProfile_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not authenticated", resp["error"])
}

// UpdateProfile tests

func (suite *InternalAuthHandlerTestSuite) TestUpdateProfile_Unauthenticated() {
	body := `{"first_name": "John"}`
	req := httptest.NewRequest("PUT", "/api/v1/auth/profile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not authenticated", resp["error"])
}

// ChangePassword tests

func (suite *InternalAuthHandlerTestSuite) TestChangePassword_Unauthenticated() {
	body := `{"current_password": "old", "new_password": "new12345"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// GetUser tests

func (suite *InternalAuthHandlerTestSuite) TestGetUser_InvalidID() {
	req := httptest.NewRequest("GET", "/api/v1/auth/admin/users/abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid user ID", resp["error"])
}

func (suite *InternalAuthHandlerTestSuite) TestGetUser_FloatID() {
	req := httptest.NewRequest("GET", "/api/v1/auth/admin/users/1.5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// UpdateUser tests

func (suite *InternalAuthHandlerTestSuite) TestUpdateUser_InvalidID() {
	body := `{"first_name": "John"}`
	req := httptest.NewRequest("PUT", "/api/v1/auth/admin/users/abc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *InternalAuthHandlerTestSuite) TestUpdateUser_InvalidJSON() {
	req := httptest.NewRequest("PUT", "/api/v1/auth/admin/users/1", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// GetAuthStatus tests

func (suite *InternalAuthHandlerTestSuite) TestGetAuthStatus_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/auth/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), false, resp["authenticated"])
}

// GetPermissions tests

func (suite *InternalAuthHandlerTestSuite) TestGetPermissions_Unauthenticated() {
	req := httptest.NewRequest("GET", "/api/v1/auth/permissions", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Not authenticated", resp["error"])
}

func TestInternalAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(InternalAuthHandlerTestSuite))
}
