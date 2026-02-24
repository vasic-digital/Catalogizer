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

// wsConn wraps a websocket.Conn with a write mutex to prevent concurrent writes.
// gorilla/websocket connections do not support concurrent writers, so all
// WriteMessage calls must be serialized per connection.
type wsConn struct {
	conn *websocket.Conn
	wmu  sync.Mutex
}

// writeMessage safely writes a message to the WebSocket connection.
func (wc *wsConn) writeMessage(messageType int, data []byte) error {
	wc.wmu.Lock()
	defer wc.wmu.Unlock()
	return wc.conn.WriteMessage(messageType, data)
}

// WebSocketHandler manages WebSocket connections for real-time updates.
type WebSocketHandler struct {
	clients  map[*wsConn]bool
	mu       sync.Mutex
	upgrader *websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients: make(map[*wsConn]bool),
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

	wc := &wsConn{conn: conn}

	h.mu.Lock()
	h.clients[wc] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, wc)
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
		wc.writeMessage(websocket.TextMessage, data)
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
			h.handleMessage(wc, message)
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

	for wc := range h.clients {
		if err := wc.writeMessage(websocket.TextMessage, data); err != nil {
			log.Printf("WebSocket broadcast error: %v", err)
		}
	}
}

// handleMessage processes incoming WebSocket messages.
func (h *WebSocketHandler) handleMessage(wc *wsConn, message []byte) {
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
			wc.writeMessage(websocket.TextMessage, data)
		}
	case "unsubscribe":
		channel, _ := msg["channel"].(string)
		ack := map[string]interface{}{
			"type":      "unsubscribed",
			"channel":   channel,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if data, err := json.Marshal(ack); err == nil {
			wc.writeMessage(websocket.TextMessage, data)
		}
	case "ping":
		pong := map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		if data, err := json.Marshal(pong); err == nil {
			wc.writeMessage(websocket.TextMessage, data)
		}
	}
}
