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
	wf.conn.SetReadDeadline(time.Now().Add(timeout))
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

// setupWebSocketStressServer creates a minimal WebSocket echo server for stress testing
func setupWebSocketStressServer(t *testing.T) *httptest.Server {
	t.Helper()

	var activeConns int64

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
		defer func() {
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
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
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
				io.ReadFull(conn, ext)
				length = int(ext[0])<<8 | int(ext[1])
			}

			var maskKey [4]byte
			if masked {
				io.ReadFull(conn, maskKey[:])
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
	t.Cleanup(func() { ts.Close() })
	return ts
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
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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

// =============================================================================
// STRESS TEST: WebSocket Concurrent Connections
// =============================================================================

func TestWebSocketStress_ConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupWebSocketStressServer(t)

	t.Run("50ConcurrentConnections", func(t *testing.T) {
		connCount := 50
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

				key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("key-%d", idx)))
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

				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
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
		t.Logf("Connected: %d/%d, Errors: %d", connected, connCount, errors)
		assert.Greater(t, connected, int64(connCount*8/10), "At least 80% of connections should succeed")

		// Send and receive on each connected client
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
		assert.Greater(t, echoSuccess, connected*8/10, "At least 80% of connected clients should echo")

		// Cleanup
		for _, wf := range frames {
			if wf != nil {
				wf.close()
			}
		}
	})
}

// =============================================================================
// STRESS TEST: WebSocket Message Throughput
// =============================================================================

func TestWebSocketStress_MessageThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupWebSocketStressServer(t)

	t.Run("HighMessageRate", func(t *testing.T) {
		wf := dialWebSocket(t, ts.URL)
		defer wf.close()

		messageCount := 500
		var sentCount int64
		var receivedCount int64

		start := time.Now()

		// Send all messages
		for i := 0; i < messageCount; i++ {
			msg := fmt.Sprintf("msg-%d", i)
			err := wf.sendText(msg)
			if err != nil {
				break
			}
			atomic.AddInt64(&sentCount, 1)

			// Read echo immediately after each send
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

		t.Logf("Sent: %d, Received: %d, Duration: %v, Rate: %.0f msg/s", sent, received, elapsed, float64(received)/elapsed.Seconds())
		assert.Equal(t, sent, received, "All sent messages should be echoed back")
		assert.Equal(t, int64(messageCount), sent, "All messages should be sent")
	})
}

// =============================================================================
// STRESS TEST: WebSocket Rapid Connect/Disconnect
// =============================================================================

func TestWebSocketStress_RapidConnectDisconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupWebSocketStressServer(t)

	t.Run("100RapidCycles", func(t *testing.T) {
		cycles := 100
		var successCount int64
		var errorCount int64

		var wg sync.WaitGroup
		for i := 0; i < cycles; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				addr := strings.TrimPrefix(ts.URL, "http://")
				conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					return
				}

				key := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("rapid-%d", idx)))
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
				err = wf.sendText("hello")
				if err != nil {
					conn.Close()
					atomic.AddInt64(&errorCount, 1)
					return
				}

				msg, err := wf.readText(3 * time.Second)
				if err != nil || msg != "hello" {
					conn.Close()
					atomic.AddInt64(&errorCount, 1)
					return
				}

				wf.close()
				atomic.AddInt64(&successCount, 1)
			}(i)
		}

		wg.Wait()

		success := atomic.LoadInt64(&successCount)
		errors := atomic.LoadInt64(&errorCount)
		t.Logf("Rapid connect/disconnect: %d success, %d errors", success, errors)
		successRate := float64(success) / float64(cycles) * 100
		assert.Greater(t, successRate, 85.0, "Should complete >85% of rapid connect/disconnect cycles")
	})
}

// =============================================================================
// STRESS TEST: WebSocket Concurrent Message Sending
// =============================================================================

func TestWebSocketStress_ConcurrentSending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := setupWebSocketStressServer(t)

	t.Run("10ConnectionsSendingConcurrently", func(t *testing.T) {
		connCount := 10
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

		t.Logf("Concurrent sending: %d/%d sent, %d/%d received", sent, expected, received, expected)
		assert.Greater(t, sent, expected*8/10, "At least 80% of messages should be sent")
		assert.Equal(t, sent, received, "All sent messages should be echoed")
	})
}
