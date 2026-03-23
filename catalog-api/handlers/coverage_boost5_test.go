package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"catalogizer/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// AuthHandler — Logout with invalid token triggers service error (53.8%)
// ---------------------------------------------------------------------------

func TestAuthHandler_Logout_GetMethod(t *testing.T) {
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	req, _ := http.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAuthHandler_Logout_EmptyToken(t *testing.T) {
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	req, _ := http.NewRequest("POST", "/logout", nil)
	// No Authorization header
	w := httptest.NewRecorder()

	handler.Logout(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization token required")
}

// ---------------------------------------------------------------------------
// AuthHandler — RefreshToken method not allowed (53.8%)
// ---------------------------------------------------------------------------

func TestAuthHandler_RefreshToken_GetMethod(t *testing.T) {
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	req, _ := http.NewRequest("GET", "/refresh", nil)
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestAuthHandler_RefreshToken_MalformedBody(t *testing.T) {
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// AuthHandler — RegisterGin missing required fields (15.4%)
// ---------------------------------------------------------------------------

func TestAuthHandler_RegisterGin_MissingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.POST("/register", func(c *gin.Context) {
		handler.RegisterGin(c, nil)
	})

	body := map[string]string{
		"username":   "user1",
		"password":   "StrongPass123!",
		"first_name": "Test",
		"last_name":  "User",
		// missing email
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RegisterGin_ShortPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.POST("/register", func(c *gin.Context) {
		handler.RegisterGin(c, nil)
	})

	body := map[string]string{
		"username":   "user1",
		"email":      "user1@example.com",
		"password":   "short",
		"first_name": "Test",
		"last_name":  "User",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// AuthHandler — GetPermissionsGin with no token (50%)
// ---------------------------------------------------------------------------

func TestAuthHandler_GetPermissionsGin_EmptyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.GET("/permissions", handler.GetPermissionsGin)

	req := httptest.NewRequest("GET", "/permissions", nil)
	// No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// WebSocketHandler — NewWebSocketHandler (0%)
// ---------------------------------------------------------------------------

func TestNewWebSocketHandler_Default(t *testing.T) {
	handler := NewWebSocketHandler(zap.NewNop())
	defer handler.Stop()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.clients)
	assert.NotNil(t, handler.upgrader)

	stats := handler.GetStats()
	assert.Equal(t, int64(0), stats.ActiveConnections)
	assert.Equal(t, int64(1000), stats.MaxConnections)
}

func TestNewWebSocketHandler_NilLogger(t *testing.T) {
	handler := NewWebSocketHandler(nil)
	defer handler.Stop()

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
}

// ---------------------------------------------------------------------------
// WebSocketHandler — cleanupStaleConnections (0%)
// ---------------------------------------------------------------------------

func TestWebSocketHandler_CleanupStaleConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testWebSocketConfig()
	cfg.PongWait = 50 * time.Millisecond
	handler := NewWebSocketHandlerWithConfig(zap.NewNop(), cfg)

	router := gin.New()
	router.GET("/ws", handler.HandleConnection)
	server := httptest.NewServer(router)

	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	stats := handler.GetStats()
	assert.Equal(t, int64(1), stats.ActiveConnections)

	// Wait for PongWait to expire
	time.Sleep(150 * time.Millisecond)

	handler.cleanupStaleConnections()

	stats = handler.GetStats()
	assert.Equal(t, int64(0), stats.ActiveConnections)

	handler.Stop()
	conn.Close()
	server.Close()
}

// ---------------------------------------------------------------------------
// WebSocketHandler — BroadcastToClientsContext (0%)
// ---------------------------------------------------------------------------

func TestWebSocketHandler_BroadcastToClientsContext_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewWebSocketHandlerWithConfig(zap.NewNop(), testWebSocketConfig())

	router := gin.New()
	router.GET("/ws", handler.HandleConnection)
	server := httptest.NewServer(router)

	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	ctx := context.Background()
	err = handler.BroadcastToClientsContext(ctx, map[string]interface{}{
		"type":    "test",
		"message": "context broadcast",
	})
	require.NoError(t, err)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "context broadcast")

	handler.Stop()
	conn.Close()
	server.Close()
}

func TestWebSocketHandler_BroadcastToClientsContext_InvalidJSON(t *testing.T) {
	handler := NewWebSocketHandlerWithConfig(zap.NewNop(), testWebSocketConfig())
	defer handler.Stop()

	err := handler.BroadcastToClientsContext(context.Background(), map[string]interface{}{
		"channel": make(chan int),
	})
	assert.Error(t, err)
}

func TestWebSocketHandler_BroadcastToClientsContext_NoClients(t *testing.T) {
	handler := NewWebSocketHandlerWithConfig(zap.NewNop(), testWebSocketConfig())
	defer handler.Stop()

	err := handler.BroadcastToClientsContext(context.Background(), map[string]interface{}{
		"type": "test",
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// WebSocketHandler — Stop idempotency
// ---------------------------------------------------------------------------

func TestWebSocketHandler_StopIdempotent(t *testing.T) {
	handler := NewWebSocketHandler(zap.NewNop())

	handler.Stop()
	handler.Stop()
	handler.Stop()
}

// ---------------------------------------------------------------------------
// WebSocketHandler — HandleConnection max connections
// ---------------------------------------------------------------------------

func TestWebSocketHandler_MaxConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testWebSocketConfig()
	cfg.MaxConnections = 1
	handler := NewWebSocketHandlerWithConfig(zap.NewNop(), cfg)

	router := gin.New()
	router.GET("/ws", handler.HandleConnection)
	server := httptest.NewServer(router)

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if resp != nil {
			assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		}
	}

	handler.Stop()
	conn1.Close()
	server.Close()
}

// ---------------------------------------------------------------------------
// AuthHandler — GetAuthStatusGin additional path (66.7%)
// ---------------------------------------------------------------------------

func TestAuthHandler_GetAuthStatusGin_NoTokenReturnsGuest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.GET("/auth/status", handler.GetAuthStatusGin)

	req := httptest.NewRequest("GET", "/auth/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without token, returns guest status
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, false, resp["authenticated"])
}

// ---------------------------------------------------------------------------
// AuthHandler — LoginGin missing credentials (90.9%)
// ---------------------------------------------------------------------------

func TestAuthHandler_LoginGin_MalformedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.POST("/login", handler.LoginGin)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---------------------------------------------------------------------------
// AuthHandler — LogoutGin missing token (88.9%)
// ---------------------------------------------------------------------------

func TestAuthHandler_LogoutGin_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.POST("/logout", handler.LogoutGin)

	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// AuthHandler — GetCurrentUserGin missing token (88.9%)
// ---------------------------------------------------------------------------

func TestAuthHandler_GetCurrentUserGin_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.GET("/me", handler.GetCurrentUserGin)

	req := httptest.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// AuthHandler — RefreshTokenGin invalid body (88.9%)
// ---------------------------------------------------------------------------

func TestAuthHandler_RefreshTokenGin_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authService := services.NewAuthService(nil, "test-secret")
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.POST("/refresh", handler.RefreshTokenGin)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
