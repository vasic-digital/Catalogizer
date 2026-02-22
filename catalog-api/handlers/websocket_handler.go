package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler manages WebSocket connections for real-time updates.
type WebSocketHandler struct {
	clients  map[*websocket.Conn]bool
	mu       sync.Mutex
	upgrader *websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients: make(map[*websocket.Conn]bool),
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// HandleConnection upgrades an HTTP request to a WebSocket connection
// and manages the connection lifecycle.
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	// Send welcome message
	welcome := map[string]interface{}{
		"type":      "connected",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"message":   "Connected to Catalogizer real-time updates",
	}
	if data, err := json.Marshal(welcome); err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Read loop - handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			h.handleMessage(conn, message)
		}
	}
}

// BroadcastToClients sends a message to all connected WebSocket clients.
func (h *WebSocketHandler) BroadcastToClients(msg map[string]interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("WebSocket broadcast error: %v", err)
		}
	}
}

// handleMessage processes incoming WebSocket messages.
func (h *WebSocketHandler) handleMessage(conn *websocket.Conn, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	msgType, _ := msg["type"].(string)
	switch msgType {
	case "subscribe":
		channel, _ := msg["channel"].(string)
		ack := map[string]interface{}{
			"type":      "subscribed",
			"channel":   channel,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if data, err := json.Marshal(ack); err == nil {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	case "unsubscribe":
		channel, _ := msg["channel"].(string)
		ack := map[string]interface{}{
			"type":      "unsubscribed",
			"channel":   channel,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if data, err := json.Marshal(ack); err == nil {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	case "ping":
		pong := map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if data, err := json.Marshal(pong); err == nil {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}
