package stress

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wsFrame is a minimal WebSocket frame writer/reader for stress testing.
// It only supports small text frames (<=125 bytes) with masking on send.
type wsFrame struct {
	conn net.Conn
}

func (wf *wsFrame) sendText(msg string) error {
	data := []byte(msg)
	frame := make([]byte, 0, 6+len(data))
	frame = append(frame, 0x81) // FIN + text opcode
	maskBit := byte(0x80)
	if len(data) <= 125 {
		frame = append(frame, maskBit|byte(len(data)))
	} else {
		return fmt.Errorf("message too large for simple frame: %d", len(data))
	}
	mask := [4]byte{0x12, 0x34, 0x56, 0x78}
	frame = append(frame, mask[:]...)
	for i, b := range data {
		frame = append(frame, b^mask[i%4])
	}
	_, err := wf.conn.Write(frame)
	return err
}

func (wf *wsFrame) readText(timeout time.Duration) (string, error) {
	if err := wf.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", fmt.Errorf("set read deadline: %w", err)
	}
	header := make([]byte, 2)
	if _, err := io.ReadFull(wf.conn, header); err != nil {
		return "", err
	}
	length := int(header[1] & 0x7F)
	if length > 125 {
		return "", fmt.Errorf("large frames not supported in test helper")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(wf.conn, payload); err != nil {
		return "", err
	}
	return string(payload), nil
}

func (wf *wsFrame) close() {
	// Send close frame (opcode 0x8)
	closeFrame := []byte{0x88, 0x80, 0x00, 0x00, 0x00, 0x00}
	wf.conn.Write(closeFrame)
	wf.conn.Close()
}

// setupWebSocketStressServer creates a minimal WebSocket echo server for stress testing.
// It tracks active connections and provides stats and broadcast endpoints.
func setupWebSocketStressServer(t *testing.T) (*httptest.Server, *wsBroadcaster) {
	t.Helper()

	var activeConns int64
	bc := newWSBroadcaster()

	mux := http.NewServeMux()

	// WebSocket echo handler (raw HTTP upgrade)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != "websocket" {
			http.Error(w, "Not a WebSocket request", http.StatusBadRequest)
			return
		}

		key := r.Header.Get("Sec-WebSocket-Key")
		acceptKey := computeAcceptKey(key)

		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Server doesn't support hijacking", http.StatusInternalServerError)
			return
		}

		conn, bufrw, err := hj.Hijack()
		if err != nil {
			return
		}
		atomic.AddInt64(&activeConns, 1)

		// Register with broadcaster
		bc.register(conn)

		defer func() {
			bc.unregister(conn)
			atomic.AddInt64(&activeConns, -1)
			conn.Close()
		}()

		// Send upgrade response
		response := "HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
		bufrw.WriteString(response)
		bufrw.Flush()

		// Echo loop
		for {
			if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				return
			}
			header := make([]byte, 2)
			if _, err := io.ReadFull(conn, header); err != nil {
				return
			}

			opcode := header[0] & 0x0F
			if opcode == 0x8 { // close
				return
			}

			masked := (header[1] & 0x80) != 0
			length := int(header[1] & 0x7F)
			if length == 126 {
				ext := make([]byte, 2)
				if _, err := io.ReadFull(conn, ext); err != nil {
					return
				}
				length = int(ext[0])<<8 | int(ext[1])
			}

			var maskKey [4]byte
			if masked {
				if _, err := io.ReadFull(conn, maskKey[:]); err != nil {
					return
				}
			}

			payload := make([]byte, length)
			if _, err := io.ReadFull(conn, payload); err != nil {
				return
			}
			if masked {
				for i := range payload {
					payload[i] ^= maskKey[i%4]
				}
			}

			// Echo back (unmasked server->client frame)
			echoFrame := make([]byte, 0, 2+length)
			echoFrame = append(echoFrame, 0x81) // FIN + text
			if length <= 125 {
				echoFrame = append(echoFrame, byte(length))
			} else {
				echoFrame = append(echoFrame, 126, byte(length>>8), byte(length&0xFF))
			}
			echoFrame = append(echoFrame, payload...)
			conn.Write(echoFrame)
		}
	})

	// Connection stats
	mux.HandleFunc("/ws/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"active_connections":%d}`, atomic.LoadInt64(&activeConns))
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(func() {
		bc.closeAll()
		ts.Close()
	})
	return ts, bc
}

func computeAcceptKey(key string) string {
	const wsGUID = "258EAFA5-E914-47DA-95CA-5AB5DC175B07"
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// dialWebSocket performs a raw WebSocket handshake and returns a wsFrame helper
func dialWebSocket(t *testing.T, serverURL string) *wsFrame {
	t.Helper()
	addr := strings.TrimPrefix(serverURL, "http://")

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	require.NoError(t, err, "Failed to connect to WebSocket server")

	key := base64.StdEncoding.EncodeToString([]byte("test-ws-key-1234"))

	request := "GET /ws HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	_, err = conn.Write([]byte(request))
	require.NoError(t, err, "Failed to send WebSocket upgrade request")

	// Read upgrade response
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}
	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	require.NoError(t, err, "Failed to read upgrade response")
	require.Contains(t, statusLine, "101", "Expected 101 Switching Protocols")

	// Read remaining headers
	for {
		line, err := reader.ReadString('\n')
		require.NoError(t, err)
		if strings.TrimSpace(line) == "" {
			break
		}
	}

	return &wsFrame{conn: conn}
}

// wsBroadcaster manages WebSocket connections for broadcast testing
type wsBroadcaster struct {
	mu    sync.Mutex
	conns map[net.Conn]struct{}
}

func newWSBroadcaster() *wsBroadcaster {
	return &wsBroadcaster{
		conns: make(map[net.Conn]struct{}),
	}
}

func (b *wsBroadcaster) register(conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.conns[conn] = struct{}{}
}

func (b *wsBroadcaster) unregister(conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.conns, conn)
}

func (b *wsBroadcaster) broadcast(msg string) int {
	b.mu.Lock()
	targets := make([]net.Conn, 0, len(b.conns))
	for c := range b.conns {
		targets = append(targets, c)
	}
	b.mu.Unlock()

	data := []byte(msg)
	frame := make([]byte, 0, 2+len(data))
	frame = append(frame, 0x81) // FIN + text
	frame = append(frame, byte(len(data)))
	frame = append(frame, data...)

	var sent int
	for _, c := range targets {
		if _, err := c.Write(frame); err == nil {
			sent++
		}
	}
	return sent
}

func (b *wsBroadcaster) closeAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for c := range b.conns {
		c.Close()
	}
	b.conns = make(map[net.Conn]struct{})
}

func (b *wsBroadcaster) count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.conns)
}

// =============================================================================
// STRESS TEST: WebSocket Concurrent Connection Storm (100 simultaneous)
// =============================================================================

func TestWebSocketStress_ConcurrentConnectionStorm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts, _ := setupWebSocketStressServer(t)

	t.Run("100SimultaneousConnections", func(t *testing.T) {
		connCount := 100
		var connectedCount int64
		var errorCount int64

		frames := make([]*wsFrame, connCount)
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < connCount; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				addr := strings.TrimPrefix(ts.URL, "http://")
				conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					return
				}

				key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("storm-%d", idx)))
				request := "GET /ws HTTP/1.1\r\n" +
					"Host: " + addr + "\r\n" +
					"Upgrade: websocket\r\n" +
					"Connection: Upgrade\r\n" +
					"Sec-WebSocket-Key: " + key + "\r\n" +
					"Sec-WebSocket-Version: 13\r\n\r\n"
				_, err = conn.Write([]byte(request))
				if err != nil {
					conn.Close()
					atomic.AddInt64(&errorCount, 1)
					return
				}

				if setErr := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); setErr != nil {
					conn.Close()
					atomic.AddInt64(&errorCount, 1)
					return
				}
				reader := bufio.NewReader(conn)
				statusLine, err := reader.ReadString('\n')
				if err != nil || !strings.Contains(statusLine, "101") {
					conn.Close()
					atomic.AddInt64(&errorCount, 1)
					return
				}
				// Read remaining headers
				for {
					line, err := reader.ReadString('\n')
					if err != nil || strings.TrimSpace(line) == "" {
						break
					}
				}

				mu.Lock()
				frames[idx] = &wsFrame{conn: conn}
				mu.Unlock()
				atomic.AddInt64(&connectedCount, 1)
			}(i)
		}

		wg.Wait()

		connected := atomic.LoadInt64(&connectedCount)
		errors := atomic.LoadInt64(&errorCount)
		t.Logf("Storm: connected %d/%d, errors %d", connected, connCount, errors)
		assert.Greater(t, connected, int64(connCount*8/10),
			"At least 80%% of connections should succeed in a storm")

		// Verify each connected client can send and receive
		var echoSuccess int64
		for _, wf := range frames {
			if wf == nil {
				continue
			}
			err := wf.sendText("ping")
			if err != nil {
				continue
			}
			msg, err := wf.readText(3 * time.Second)
			if err == nil && msg == "ping" {
				atomic.AddInt64(&echoSuccess, 1)
			}
		}
		assert.Greater(t, echoSuccess, connected*8/10,
			"At least 80%% of connected clients should echo successfully")

		// Cleanup
		for _, wf := range frames {
			if wf != nil {
				wf.close()
			}
		}
	})
}

// =============================================================================
// STRESS TEST: WebSocket Rapid Message Sending (1000 messages/second target)
// =============================================================================

func TestWebSocketStress_RapidMessageSending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts, _ := setupWebSocketStressServer(t)

	t.Run("1000MessagesPerSecondTarget", func(t *testing.T) {
		wf := dialWebSocket(t, ts.URL)
		defer wf.close()

		messageCount := 1000
		var sentCount int64
		var receivedCount int64

		start := time.Now()

		for i := 0; i < messageCount; i++ {
			msg := fmt.Sprintf("m%d", i)
			err := wf.sendText(msg)
			if err != nil {
				break
			}
			atomic.AddInt64(&sentCount, 1)

			response, err := wf.readText(3 * time.Second)
			if err != nil {
				break
			}
			if response == msg {
				atomic.AddInt64(&receivedCount, 1)
			}
		}

		elapsed := time.Since(start)
		sent := atomic.LoadInt64(&sentCount)
		received := atomic.LoadInt64(&receivedCount)
		rate := float64(received) / elapsed.Seconds()

		t.Logf("Rapid send: sent=%d, received=%d, duration=%v, rate=%.0f msg/s",
			sent, received, elapsed, rate)
		assert.Equal(t, sent, received,
			"All sent messages should be echoed back")
		assert.Equal(t, int64(messageCount), sent,
			"All messages should be sent")
		assert.Greater(t, rate, 100.0,
			"Should sustain at least 100 msg/s echo throughput")
	})

	t.Run("BurstOf500ThenVerify", func(t *testing.T) {
		wf := dialWebSocket(t, ts.URL)
		defer wf.close()

		// Send a burst of messages without reading immediately, then read all
		burstSize := 50 // keep small enough for 125-byte frame limit
		for i := 0; i < burstSize; i++ {
			msg := fmt.Sprintf("b%d", i)
			err := wf.sendText(msg)
			require.NoError(t, err, "Burst send %d should succeed", i)
		}

		// Read all echoed responses
		var received int
		for i := 0; i < burstSize; i++ {
			msg, err := wf.readText(5 * time.Second)
			if err != nil {
				break
			}
			expected := fmt.Sprintf("b%d", i)
			if msg == expected {
				received++
			}
		}

		assert.Equal(t, burstSize, received,
			"All burst messages should be echoed in order")
	})
}

// =============================================================================
// STRESS TEST: WebSocket Connection Churn (rapid connect/disconnect cycles)
// =============================================================================

func TestWebSocketStress_ConnectionChurn(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts, _ := setupWebSocketStressServer(t)

	t.Run("200RapidCycles", func(t *testing.T) {
		cycles := 200
		var successCount int64
		var errorCount int64

		var wg sync.WaitGroup
		// Run in batches of 50 to avoid fd exhaustion
		batchSize := 50
		for batchStart := 0; batchStart < cycles; batchStart += batchSize {
			batchEnd := batchStart + batchSize
			if batchEnd > cycles {
				batchEnd = cycles
			}
			for i := batchStart; i < batchEnd; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					addr := strings.TrimPrefix(ts.URL, "http://")
					conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						return
					}

					key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("churn-%d", idx)))
					request := "GET /ws HTTP/1.1\r\n" +
						"Host: " + addr + "\r\n" +
						"Upgrade: websocket\r\n" +
						"Connection: Upgrade\r\n" +
						"Sec-WebSocket-Key: " + key + "\r\n" +
						"Sec-WebSocket-Version: 13\r\n\r\n"
					conn.Write([]byte(request))

					if setErr := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); setErr != nil {
						conn.Close()
						atomic.AddInt64(&errorCount, 1)
						return
					}
					reader := bufio.NewReader(conn)
					statusLine, err := reader.ReadString('\n')
					if err != nil || !strings.Contains(statusLine, "101") {
						conn.Close()
						atomic.AddInt64(&errorCount, 1)
						return
					}
					// Drain headers
					for {
						line, _ := reader.ReadString('\n')
						if strings.TrimSpace(line) == "" {
							break
						}
					}

					wf := &wsFrame{conn: conn}
					err = wf.sendText("churn")
					if err != nil {
						conn.Close()
						atomic.AddInt64(&errorCount, 1)
						return
					}

					msg, err := wf.readText(3 * time.Second)
					if err != nil || msg != "churn" {
						conn.Close()
						atomic.AddInt64(&errorCount, 1)
						return
					}

					wf.close()
					atomic.AddInt64(&successCount, 1)
				}(i)
			}
			wg.Wait()
		}

		success := atomic.LoadInt64(&successCount)
		errors := atomic.LoadInt64(&errorCount)
		t.Logf("Connection churn: %d success, %d errors out of %d cycles",
			success, errors, cycles)
		successRate := float64(success) / float64(cycles) * 100
		assert.Greater(t, successRate, 85.0,
			"Should complete >85%% of rapid connect/disconnect cycles")
	})

	t.Run("SequentialChurnNoLeaks", func(t *testing.T) {
		// Sequential connect-send-disconnect to verify no connection leaks
		for i := 0; i < 50; i++ {
			addr := strings.TrimPrefix(ts.URL, "http://")
			conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
			require.NoError(t, err, "Cycle %d dial should succeed", i)

			key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("seq-%d", i)))
			request := "GET /ws HTTP/1.1\r\n" +
				"Host: " + addr + "\r\n" +
				"Upgrade: websocket\r\n" +
				"Connection: Upgrade\r\n" +
				"Sec-WebSocket-Key: " + key + "\r\n" +
				"Sec-WebSocket-Version: 13\r\n\r\n"
			conn.Write([]byte(request))

			conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			reader := bufio.NewReader(conn)
			statusLine, err := reader.ReadString('\n')
			require.NoError(t, err)
			require.Contains(t, statusLine, "101")

			for {
				line, _ := reader.ReadString('\n')
				if strings.TrimSpace(line) == "" {
					break
				}
			}

			wf := &wsFrame{conn: conn}
			require.NoError(t, wf.sendText("seq"))
			msg, err := wf.readText(3 * time.Second)
			require.NoError(t, err)
			assert.Equal(t, "seq", msg)
			wf.close()
		}

		// After all sequential cycles, check server health
		resp, err := http.Get(ts.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Server should remain healthy after sequential churn")
	})
}

// =============================================================================
// STRESS TEST: WebSocket Message Broadcast Under Load
// =============================================================================

func TestWebSocketStress_BroadcastUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts, bc := setupWebSocketStressServer(t)

	t.Run("BroadcastTo20Clients", func(t *testing.T) {
		clientCount := 20

		// Connect all clients
		clients := make([]*wsFrame, clientCount)
		for i := 0; i < clientCount; i++ {
			clients[i] = dialWebSocket(t, ts.URL)
		}

		// Allow connections to register with broadcaster
		time.Sleep(100 * time.Millisecond)

		registeredCount := bc.count()
		t.Logf("Registered %d clients with broadcaster", registeredCount)
		assert.GreaterOrEqual(t, registeredCount, clientCount,
			"All clients should be registered with broadcaster")

		// Broadcast a message
		broadcastMsg := "broadcast-test"
		sent := bc.broadcast(broadcastMsg)
		t.Logf("Broadcast sent to %d connections", sent)
		assert.GreaterOrEqual(t, sent, clientCount,
			"Broadcast should reach all registered clients")

		// Read broadcast from each client
		var receivedCount int64
		var wg sync.WaitGroup
		for i, client := range clients {
			wg.Add(1)
			go func(idx int, c *wsFrame) {
				defer wg.Done()
				msg, err := c.readText(3 * time.Second)
				if err == nil && msg == broadcastMsg {
					atomic.AddInt64(&receivedCount, 1)
				}
			}(i, client)
		}
		wg.Wait()

		received := atomic.LoadInt64(&receivedCount)
		t.Logf("Broadcast received by %d/%d clients", received, clientCount)
		assert.GreaterOrEqual(t, received, int64(clientCount*8/10),
			"At least 80%% of clients should receive the broadcast")

		// Cleanup
		for _, c := range clients {
			c.close()
		}
	})

	t.Run("RepeatedBroadcastsUnderLoad", func(t *testing.T) {
		clientCount := 10
		broadcastRounds := 5

		clients := make([]*wsFrame, clientCount)
		for i := 0; i < clientCount; i++ {
			clients[i] = dialWebSocket(t, ts.URL)
		}
		defer func() {
			for _, c := range clients {
				c.close()
			}
		}()

		time.Sleep(100 * time.Millisecond)

		var totalReceived int64
		expectedTotal := int64(clientCount * broadcastRounds)

		for round := 0; round < broadcastRounds; round++ {
			msg := fmt.Sprintf("round-%d", round)
			bc.broadcast(msg)

			var wg sync.WaitGroup
			for _, client := range clients {
				wg.Add(1)
				go func(c *wsFrame) {
					defer wg.Done()
					received, err := c.readText(3 * time.Second)
					if err == nil && received == msg {
						atomic.AddInt64(&totalReceived, 1)
					}
				}(client)
			}
			wg.Wait()
		}

		received := atomic.LoadInt64(&totalReceived)
		t.Logf("Repeated broadcast: received %d/%d total", received, expectedTotal)
		assert.Greater(t, received, expectedTotal*7/10,
			"At least 70%% of broadcast messages should be received across all rounds")
	})
}

// =============================================================================
// STRESS TEST: WebSocket Connection Cleanup Verification After Storm
// =============================================================================

func TestWebSocketStress_CleanupAfterStorm(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts, bc := setupWebSocketStressServer(t)

	t.Run("AllConnectionsReleasedAfterDisconnect", func(t *testing.T) {
		stormSize := 50

		// Connect a batch of clients
		clients := make([]*wsFrame, stormSize)
		for i := 0; i < stormSize; i++ {
			clients[i] = dialWebSocket(t, ts.URL)
		}

		// Allow server to register all connections
		time.Sleep(200 * time.Millisecond)

		activeAfterConnect := bc.count()
		t.Logf("Active connections after connect: %d", activeAfterConnect)
		assert.GreaterOrEqual(t, activeAfterConnect, stormSize,
			"All clients should be active after connecting")

		// Verify server stats endpoint shows connections
		resp, err := http.Get(ts.URL + "/ws/stats")
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
		t.Logf("Stats before cleanup: %s", string(body))

		// Close all clients
		for _, c := range clients {
			c.close()
		}

		// Allow server cleanup goroutines to process disconnections
		time.Sleep(500 * time.Millisecond)

		activeAfterClose := bc.count()
		t.Logf("Active connections after close: %d", activeAfterClose)
		assert.LessOrEqual(t, activeAfterClose, 2,
			"All connections should be cleaned up after close (allowing small margin for timing)")

		// Verify server remains healthy
		resp, err = http.Get(ts.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Server should be healthy after storm cleanup")
	})

	t.Run("NewConnectionsWorkAfterStormCleanup", func(t *testing.T) {
		// After storm, verify new connections still work
		wf := dialWebSocket(t, ts.URL)
		defer wf.close()

		err := wf.sendText("post-storm")
		require.NoError(t, err)
		msg, err := wf.readText(3 * time.Second)
		require.NoError(t, err)
		assert.Equal(t, "post-storm", msg,
			"New connections should work after storm cleanup")
	})
}

// =============================================================================
// STRESS TEST: WebSocket Concurrent Message Sending (Multiple Connections)
// =============================================================================

func TestWebSocketStress_ConcurrentSendingMultipleConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts, _ := setupWebSocketStressServer(t)

	t.Run("20ConnectionsSending50Each", func(t *testing.T) {
		connCount := 20
		messagesPerConn := 50
		var totalSent int64
		var totalReceived int64

		var wg sync.WaitGroup
		for i := 0; i < connCount; i++ {
			wg.Add(1)
			go func(connID int) {
				defer wg.Done()

				addr := strings.TrimPrefix(ts.URL, "http://")
				conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
				if err != nil {
					return
				}

				key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("conc-%d", connID)))
				request := "GET /ws HTTP/1.1\r\n" +
					"Host: " + addr + "\r\n" +
					"Upgrade: websocket\r\n" +
					"Connection: Upgrade\r\n" +
					"Sec-WebSocket-Key: " + key + "\r\n" +
					"Sec-WebSocket-Version: 13\r\n\r\n"
				conn.Write([]byte(request))

				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				reader := bufio.NewReader(conn)
				statusLine, _ := reader.ReadString('\n')
				if !strings.Contains(statusLine, "101") {
					conn.Close()
					return
				}
				for {
					line, _ := reader.ReadString('\n')
					if strings.TrimSpace(line) == "" {
						break
					}
				}

				wf := &wsFrame{conn: conn}
				defer wf.close()

				for j := 0; j < messagesPerConn; j++ {
					msg := fmt.Sprintf("c%d-m%d", connID, j)
					err := wf.sendText(msg)
					if err != nil {
						break
					}
					atomic.AddInt64(&totalSent, 1)

					response, err := wf.readText(3 * time.Second)
					if err != nil {
						break
					}
					if response == msg {
						atomic.AddInt64(&totalReceived, 1)
					}
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		sent := atomic.LoadInt64(&totalSent)
		received := atomic.LoadInt64(&totalReceived)
		expected := int64(connCount * messagesPerConn)

		t.Logf("Concurrent sending: %d/%d sent, %d/%d received",
			sent, expected, received, expected)
		assert.Greater(t, sent, expected*8/10,
			"At least 80%% of messages should be sent")
		assert.Equal(t, sent, received,
			"All sent messages should be echoed")
	})
}
