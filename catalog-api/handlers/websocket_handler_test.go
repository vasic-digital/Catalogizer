package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketHandler_HandleConnection(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create WebSocket handler
	handler := NewWebSocketHandler()

	// Create a test HTTP server
	router := gin.New()
	router.GET("/ws", handler.HandleConnection)
	server := httptest.NewServer(router)
	defer server.Close()

	// Connect to WebSocket endpoint
	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	// Read welcome message
	messageType, message, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Contains(t, string(message), "Connected to Catalogizer real-time updates")

	// Send ping message
	err = conn.WriteJSON(map[string]interface{}{
		"type": "ping",
	})
	require.NoError(t, err)

	// Expect pong response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	messageType, message, err = conn.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Contains(t, string(message), "pong")
}

func TestWebSocketHandler_BroadcastToClients(t *testing.T) {
	handler := NewWebSocketHandler()

	// Create two test connections
	router := gin.New()
	router.GET("/ws", handler.HandleConnection)
	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn2.Close()

	// Wait for welcome messages (discard)
	conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
	conn1.ReadMessage()
	conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	conn2.ReadMessage()

	// Broadcast a message
	handler.BroadcastToClients(map[string]interface{}{
		"type":    "test",
		"message": "hello",
	})

	// Both connections should receive the broadcast
	conn1.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msg1, err := conn1.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg1), "hello")

	conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msg2, err := conn2.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg2), "hello")
}
