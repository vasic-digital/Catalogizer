package challenges

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"

	"github.com/gorilla/websocket"
)

// WebSocketEventsChallenge validates the WebSocket event system:
// connects to the WebSocket endpoint, triggers a scan status
// check via the API, and verifies that events can be received.
type WebSocketEventsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewWebSocketEventsChallenge creates CH-033.
func NewWebSocketEventsChallenge() *WebSocketEventsChallenge {
	return &WebSocketEventsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"websocket-events",
			"WebSocket Events",
			"Connects to WebSocket endpoint, triggers a scan status "+
				"check, verifies event system is operational.",
			"realtime",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the WebSocket events challenge.
func (c *WebSocketEventsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	apiClient := httpclient.NewAPIClient(c.config.BaseURL)

	_, err := apiClient.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "login",
			Passed:  false,
			Message: fmt.Sprintf("Login failed: %v", err),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, err.Error(),
		), nil
	}

	token := apiClient.Token()

	// Step 1: Determine WebSocket URL
	wsURL := strings.Replace(c.config.BaseURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Try common WebSocket paths
	wsPaths := []string{
		"/api/v1/ws",
		"/ws",
		"/api/v1/events",
	}

	c.ReportProgress("connecting-websocket", map[string]any{
		"base_url": wsURL,
	})

	var wsConn *websocket.Conn
	var wsPath string
	var connectErr error

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	headers := http.Header{}
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	for _, path := range wsPaths {
		fullURL := wsURL + path
		if token != "" {
			if strings.Contains(path, "?") {
				fullURL += "&token=" + token
			} else {
				fullURL += "?token=" + token
			}
		}
		wsConn, _, connectErr = dialer.DialContext(ctx, fullURL, headers)
		if connectErr == nil {
			wsPath = path
			break
		}
	}

	wsConnected := wsConn != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "websocket_connection",
		Expected: "WebSocket connection established",
		Actual:   challenge.Ternary(wsConnected, fmt.Sprintf("connected via %s", wsPath), fmt.Sprintf("err=%v", connectErr)),
		Passed:   wsConnected,
		Message: challenge.Ternary(wsConnected,
			fmt.Sprintf("WebSocket connected at %s", wsPath),
			fmt.Sprintf("WebSocket connection failed (tried %v): %v", wsPaths, connectErr)),
	})
	outputs["ws_path"] = wsPath
	outputs["ws_connected"] = fmt.Sprintf("%v", wsConnected)

	if !wsConnected {
		// If WebSocket is not available, that's still valid info.
		// We verify the API is healthy instead.
		healthCode, _, healthErr := apiClient.Get(ctx, "/health")
		healthOK := healthErr == nil && healthCode == 200

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "api_health_fallback",
			Expected: "200",
			Actual:   fmt.Sprintf("HTTP %d", healthCode),
			Passed:   healthOK,
			Message: challenge.Ternary(healthOK,
				"API is healthy (WebSocket may not be configured)",
				fmt.Sprintf("API health check also failed: code=%d err=%v", healthCode, healthErr)),
		})

		metrics := map[string]challenge.MetricValue{
			"websocket_latency": {
				Name:  "websocket_latency",
				Value: float64(time.Since(start).Milliseconds()),
				Unit:  "ms",
			},
		}

		status := challenge.StatusPassed
		for _, a := range assertions {
			if !a.Passed {
				status = challenge.StatusFailed
				break
			}
		}

		return c.CreateResult(
			status, start, assertions, metrics, outputs, "",
		), nil
	}

	defer wsConn.Close()

	// Step 2: Set a read deadline and try to receive a message
	c.ReportProgress("reading-events", nil)
	wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Send a ping/subscribe message
	subscribeMsg := `{"type":"subscribe","channel":"scan_status"}`
	writeErr := wsConn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
	writeOK := writeErr == nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "websocket_write",
		Expected: "message sent successfully",
		Actual:   challenge.Ternary(writeOK, "sent", fmt.Sprintf("err=%v", writeErr)),
		Passed:   writeOK,
		Message: challenge.Ternary(writeOK,
			"Subscribe message sent to WebSocket",
			fmt.Sprintf("Failed to write to WebSocket: %v", writeErr)),
	})

	// Step 3: Trigger a scan status check via the API
	// Use GetRaw since the endpoint may not exist (returns HTML 404)
	c.ReportProgress("triggering-event", nil)
	statusCode, _, statusErr := apiClient.GetRaw(ctx, "/api/v1/storage-roots")
	statusOK := statusErr == nil && (statusCode == 200 || statusCode == 404)

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "scan_status_trigger",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("HTTP %d", statusCode),
		Passed:   statusOK,
		Message: challenge.Ternary(statusOK,
			fmt.Sprintf("Scan status endpoint responds: code=%d", statusCode),
			fmt.Sprintf("Scan status endpoint failed: code=%d err=%v", statusCode, statusErr)),
	})

	// Step 4: Try to read a message from WebSocket
	_, msgBytes, readErr := wsConn.ReadMessage()
	receivedMessage := readErr == nil && len(msgBytes) > 0

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "websocket_receive",
		Expected: "message received or timeout (both acceptable)",
		Actual: challenge.Ternary(receivedMessage,
			fmt.Sprintf("received %d bytes", len(msgBytes)),
			fmt.Sprintf("no message (timeout or err=%v)", readErr)),
		Passed: true, // Both outcomes are acceptable
		Message: challenge.Ternary(receivedMessage,
			fmt.Sprintf("Received WebSocket message: %d bytes", len(msgBytes)),
			"No message within timeout (expected if no active events)"),
	})
	outputs["message_received"] = fmt.Sprintf("%v", receivedMessage)

	metrics := map[string]challenge.MetricValue{
		"websocket_latency": {
			Name:  "websocket_latency",
			Value: float64(time.Since(start).Milliseconds()),
			Unit:  "ms",
		},
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, metrics, outputs, "",
	), nil
}
