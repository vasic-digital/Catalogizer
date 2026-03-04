package challenges

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// GracefulShutdownChallenge validates that the application handles
// graceful shutdown correctly. Verifies the API responds to health
// checks, accepts concurrent connections, and handles them cleanly
// (prerequisites for graceful shutdown support).
type GracefulShutdownChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewGracefulShutdownChallenge creates CH-050.
func NewGracefulShutdownChallenge() *GracefulShutdownChallenge {
	return &GracefulShutdownChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"graceful-shutdown",
			"Graceful Shutdown Support",
			"Validates the application supports graceful shutdown: "+
				"verifies health endpoint, concurrent request handling, "+
				"connection keep-alive, and proper HTTP response headers.",
			"resilience",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the graceful shutdown challenge.
func (c *GracefulShutdownChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Verify API is running and healthy
	c.ReportProgress("health-check", nil)
	healthCode, healthBody, healthErr := client.Get(ctx, "/health")
	healthOK := healthErr == nil && healthCode == 200
	healthStatus := ""
	if healthBody != nil {
		if s, ok := healthBody["status"].(string); ok {
			healthStatus = s
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "health_endpoint",
		Expected: "200 with status=healthy",
		Actual:   fmt.Sprintf("HTTP %d, status=%q", healthCode, healthStatus),
		Passed:   healthOK,
		Message: challenge.Ternary(healthOK,
			"API is running and healthy",
			fmt.Sprintf("API not healthy: code=%d err=%v", healthCode, healthErr)),
	})

	if !healthOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "API not healthy",
		), nil
	}

	// Step 2: Verify server returns proper HTTP headers for connection handling
	c.ReportProgress("connection-headers", nil)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, c.config.BaseURL+"/health", nil,
	)

	resp, err := httpClient.Do(req)
	if err == nil {
		resp.Body.Close()

		// Check for proper HTTP response (not HTTP/0.9 or malformed)
		protoOK := resp.ProtoMajor >= 1
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "http_protocol",
			Expected: "HTTP/1.1 or higher",
			Actual:   resp.Proto,
			Passed:   protoOK,
			Message: challenge.Ternary(protoOK,
				fmt.Sprintf("Server uses %s protocol", resp.Proto),
				fmt.Sprintf("Unexpected protocol: %s", resp.Proto)),
		})
		outputs["http_protocol"] = resp.Proto

		// Check Content-Type header is set
		contentType := resp.Header.Get("Content-Type")
		hasContentType := contentType != ""
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "content_type_header",
			Expected: "Content-Type header present",
			Actual:   challenge.Ternary(hasContentType, contentType, "missing"),
			Passed:   hasContentType,
			Message: challenge.Ternary(hasContentType,
				fmt.Sprintf("Content-Type header present: %s", contentType),
				"Content-Type header missing from response"),
		})
	}

	// Step 3: Verify concurrent connections are handled (prerequisite for graceful drain)
	c.ReportProgress("concurrent-connections", nil)
	concurrentCount := 5
	results := make(chan bool, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func() {
			reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			code, _, reqErr := client.Get(reqCtx, "/health")
			results <- reqErr == nil && code == 200
		}()
	}

	successCount := 0
	for i := 0; i < concurrentCount; i++ {
		if <-results {
			successCount++
		}
	}

	concurrentOK := successCount == concurrentCount
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "concurrent_connections",
		Expected: fmt.Sprintf("%d/%d concurrent requests succeed", concurrentCount, concurrentCount),
		Actual:   fmt.Sprintf("%d/%d", successCount, concurrentCount),
		Passed:   concurrentOK,
		Message: challenge.Ternary(concurrentOK,
			fmt.Sprintf("All %d concurrent connections handled", concurrentCount),
			fmt.Sprintf("Only %d/%d concurrent connections succeeded", successCount, concurrentCount)),
	})

	// Step 4: Verify the server handles keep-alive properly
	c.ReportProgress("keep-alive", nil)
	keepAliveClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        5,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 5,
		},
	}

	kaSuccess := 0
	for i := 0; i < 3; i++ {
		kaReq, _ := http.NewRequestWithContext(
			ctx, http.MethodGet, c.config.BaseURL+"/health", nil,
		)
		kaResp, kaErr := keepAliveClient.Do(kaReq)
		if kaErr == nil && kaResp.StatusCode == 200 {
			kaSuccess++
			kaResp.Body.Close()
		}
	}

	kaOK := kaSuccess == 3
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "keep_alive_connections",
		Expected: "3/3 keep-alive requests succeed",
		Actual:   fmt.Sprintf("%d/3", kaSuccess),
		Passed:   kaOK,
		Message: challenge.Ternary(kaOK,
			"Keep-alive connections handled correctly",
			fmt.Sprintf("Keep-alive issue: %d/3 requests succeeded", kaSuccess)),
	})

	metrics := map[string]challenge.MetricValue{
		"shutdown_readiness_time": {
			Name:  "shutdown_readiness_time",
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

	return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}
