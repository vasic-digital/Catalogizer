# Module 6: Real-Time Features - Script

**Duration**: 30 minutes
**Module**: 6 - Real-Time Features

---

## Scene 1: WebSocket Server (0:00 - 15:00)

**[Visual: Architecture diagram showing: Client <-> WebSocket Handler <-> Event Bus <-> Scanner/Watcher]**

**Narrator**: Welcome to Module 6. Real-time communication transforms Catalogizer from a request-response application into a live, reactive system. Scan progress, file changes, and system notifications all stream to connected clients in real time through WebSocket.

**[Visual: Open `catalog-api/handlers/websocket_handler.go`]**

**Narrator**: The WebSocket handler manages client connections. It uses the Gorilla WebSocket library with a connection upgrader that converts HTTP requests into persistent WebSocket connections.

```go
// catalog-api/handlers/websocket_handler.go
type WebSocketHandler struct {
    clients  map[*wsConn]bool
    mu       sync.Mutex
    upgrader *websocket.Upgrader
}

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
```

**[Visual: Highlight the `wsConn` wrapper]**

**Narrator**: A critical detail: Gorilla WebSocket connections do not support concurrent writers. The `wsConn` wrapper adds a write mutex to serialize all outgoing messages per connection. Without this, concurrent broadcasts would corrupt the wire protocol.

```go
// catalog-api/handlers/websocket_handler.go
type wsConn struct {
    conn *websocket.Conn
    wmu  sync.Mutex
}

func (wc *wsConn) writeMessage(messageType int, data []byte) error {
    wc.wmu.Lock()
    defer wc.wmu.Unlock()
    return wc.conn.WriteMessage(messageType, data)
}
```

**[Visual: Show `HandleConnection` method]**

**Narrator**: `HandleConnection` upgrades the HTTP request, registers the client in the connection map, and starts a read loop. The read loop processes incoming messages (like subscription requests) and detects disconnections. When the connection drops, the client is removed from the map.

```go
// catalog-api/handlers/websocket_handler.go
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
    conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }

    wc := &wsConn{conn: conn}
    // Register client, start read loop...
}
```

**[Visual: Show broadcast pattern]**

**Narrator**: Broadcasting iterates over all connected clients and sends the message to each one. Failed sends remove the client from the map. This fan-out pattern is simple and effective for moderate client counts.

**[Visual: Show message protocol with JSON]**

**Narrator**: Messages use a JSON protocol with a `type` field for routing and a `payload` field for data. Types include `scan_progress`, `scan_complete`, `file_change`, `entity_created`, and `system_notification`. The frontend's WebSocket handler dispatches on the type field to update the appropriate React Query cache keys.

```json
{
  "type": "scan_progress",
  "payload": {
    "job_id": "scan-123",
    "storage_root": "NAS Media",
    "files_processed": 12500,
    "files_found": 85000,
    "current_path": "/movies/Action/",
    "status": "running"
  }
}
```

**[Visual: Show the frontend WebSocket client from `@vasic-digital/websocket-client`]**

**Narrator**: The frontend uses the `@vasic-digital/websocket-client` submodule, which provides automatic reconnection with exponential backoff, React hooks for subscription, and connection state management. When the connection drops, the client retries with increasing delays and re-subscribes to all active topics on reconnect.

---

## Scene 2: Live Updates (15:00 - 30:00)

**[Visual: Open `catalog-api/internal/media/realtime/watcher.go`]**

**Narrator**: The real-time watcher monitors filesystem changes on mounted storage. The `SMBChangeWatcher` uses `fsnotify` to detect file creates, modifies, deletes, and moves.

```go
// catalog-api/internal/media/realtime/watcher.go
type SMBChangeWatcher struct {
    mediaDB       *database.MediaDatabase
    analyzer      *analyzer.MediaAnalyzer
    logger        *zap.Logger
    watchers      map[string]*fsnotify.Watcher
    changeQueue   chan ChangeEvent
    workers       int
    debounceMap   map[string]*debounceEntry
    debounceDelay time.Duration
    stopCh        chan struct{}
}
```

**[Visual: Show `ChangeEvent` struct]**

**Narrator**: Each filesystem change produces a `ChangeEvent` with the path, SMB root, operation type, timestamp, file size, and whether it is a directory. These events flow into a buffered channel for worker processing.

```go
// catalog-api/internal/media/realtime/watcher.go
type ChangeEvent struct {
    Path      string
    SmbRoot   string
    Operation string // created, modified, deleted, moved
    Timestamp time.Time
    Size      int64
    IsDir     bool
}
```

**[Visual: Show debounce logic]**

**Narrator**: File operations often produce burst events -- a file write triggers multiple modified events within milliseconds. The watcher implements debouncing with a 2-second delay. Each path gets a timer; new events for the same path reset the timer. Only after 2 seconds of silence is the event processed.

**[Visual: Show the event bus in tests]**

**Narrator**: The event bus pattern uses Go channels for pub/sub routing. Publishers send events to the bus. Subscribers register callbacks for specific event types. In the test suite, a full EventBus implementation validates the same pub/sub semantics used by the production watchers.

```go
// catalog-api/internal/media/realtime/event_bus_test.go
type EventType string

const (
    EventFileCreated  EventType = "created"
    EventFileModified EventType = "modified"
    EventFileDeleted  EventType = "deleted"
    EventFileMoved    EventType = "moved"
)

type BusEvent struct {
    Type      EventType
    Path      string
    Timestamp time.Time
    Payload   interface{}
}
```

**[Visual: Show scan progress streaming flow]**

**Narrator**: Scan progress follows a clear pipeline. The `UniversalScanner` updates the `ScanStatus` struct atomically. A periodic goroutine reads the status every 5 seconds and publishes it to the WebSocket handler. The handler broadcasts to all connected clients. The frontend receives the update, and React Query automatically refreshes the scan status UI.

**[Visual: Show the EnhancedChangeWatcher]**

**Narrator**: The `EnhancedChangeWatcher` in `internal/media/realtime/enhanced_watcher.go` extends the basic watcher with deeper analysis integration. When a change is detected, it triggers the media analyzer to re-analyze the affected directory, potentially updating entity metadata and cover art.

**[Visual: Show real-time notification types]**

**Narrator**: The system supports several notification types: scan progress (percentage, files processed, current path), scan completion (total files, duration, entities created), file change alerts (new, modified, deleted files), entity updates (new entities, metadata changes), and system notifications (storage health, cache status).

**[Visual: Show presence indicators concept]**

**Narrator**: Each WebSocket connection is tracked as an active client. The frontend can request the count of connected clients to display presence indicators -- showing how many users are currently online. This is a lightweight feature built on top of the existing connection management.

**[Visual: Course title card]**

**Narrator**: Real-time features tie the entire system together. The scanner reports progress, the watcher detects changes, the event bus routes notifications, and the WebSocket handler delivers them to every connected client. In Module 7, we dive deep into the protocol implementations that make multi-storage scanning possible.

---

## Key Code Examples

### WebSocket Route Registration
```go
// main.go
wsHandler := handlers.NewWebSocketHandler()
router.GET("/api/v1/ws", wsHandler.HandleConnection)
```

### Frontend WebSocket Integration
```typescript
// catalog-web/src/lib/websocket.ts
import { useWebSocket } from '@vasic-digital/websocket-client';

function ScanProgress() {
  const { lastMessage, connectionStatus } = useWebSocket('/api/v1/ws');

  useEffect(() => {
    if (lastMessage?.type === 'scan_progress') {
      // Invalidate React Query cache to refresh UI
      queryClient.invalidateQueries({ queryKey: ['scan-status'] });
    }
  }, [lastMessage]);
}
```

### Event Flow
```
fsnotify event
    -> ChangeEvent (debounced, 2s)
    -> changeQueue channel
    -> Worker goroutine
    -> MediaAnalyzer (re-analysis)
    -> WebSocketHandler.Broadcast()
    -> All connected clients
    -> React Query invalidation
    -> UI update
```

---

## Quiz Questions

1. Why does the `wsConn` wrapper include a write mutex?
   **Answer**: Gorilla WebSocket connections do not support concurrent writers. Multiple goroutines (broadcast, ping, direct send) could write simultaneously, corrupting the WebSocket frame protocol. The write mutex serializes all outgoing messages per connection.

2. What is the purpose of debouncing in the file watcher?
   **Answer**: File operations (especially writes) produce rapid burst events -- a single file save can trigger multiple "modified" events within milliseconds. Debouncing waits 2 seconds after the last event for a given path before processing. This prevents redundant analysis work and reduces WebSocket traffic.

3. How do real-time scan progress updates reach the frontend UI?
   **Answer**: The UniversalScanner updates ScanStatus atomically. A periodic goroutine reads the status and publishes it as a JSON message to the WebSocket handler. The handler broadcasts to all connected clients. The frontend receives the message, and the WebSocket handler invalidates React Query cache keys, triggering automatic UI re-render.

4. What reconnection strategy does the frontend WebSocket client use?
   **Answer**: The `@vasic-digital/websocket-client` submodule implements exponential backoff reconnection. When the connection drops, it retries with increasing delays. On reconnect, it automatically re-subscribes to all previously active topics, ensuring no subscription state is lost.
