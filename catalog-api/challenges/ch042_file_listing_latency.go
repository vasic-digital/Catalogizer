package challenges

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// FileListingLatencyChallenge validates that the file listing
// endpoint responds within 200ms under concurrent load. Sends
// parallel requests and measures average response time.
type FileListingLatencyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewFileListingLatencyChallenge creates CH-042.
func NewFileListingLatencyChallenge() *FileListingLatencyChallenge {
	return &FileListingLatencyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"file-listing-latency",
			"File Listing Latency Under Load",
			"Validates the file listing endpoint responds within "+
				"200ms average under concurrent load. Sends 5 parallel "+
				"workers making requests and measures latency.",
			"performance",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the file listing latency challenge.
func (c *FileListingLatencyChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login first
	c.ReportProgress("login", nil)
	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "login",
			Passed:  false,
			Message: fmt.Sprintf("Login failed: %v", loginErr),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, loginErr.Error(),
		), nil
	}

	// Determine which file endpoint exists
	c.ReportProgress("discover-endpoint", nil)
	fileEndpoints := []string{
		"/api/v1/files?limit=10",
		"/api/v1/entities?limit=10",
	}

	var workingEndpoint string
	for _, ep := range fileEndpoints {
		code, _, err := client.Get(ctx, ep)
		if err == nil && code == 200 {
			workingEndpoint = ep
			break
		}
	}

	if workingEndpoint == "" {
		workingEndpoint = "/api/v1/entities?limit=10"
	}
	outputs["endpoint"] = workingEndpoint

	// Warm up
	client.Get(ctx, workingEndpoint)

	// Concurrent load test
	c.ReportProgress("load-test", nil)
	concurrency := 5
	requestsPerWorker := 4
	var totalLatencyMs int64
	var successes int64
	var failures int64

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				reqStart := time.Now()
				code, _, err := client.Get(ctx, workingEndpoint)
				elapsed := time.Since(reqStart).Milliseconds()
				atomic.AddInt64(&totalLatencyMs, elapsed)

				if err == nil && code == 200 {
					atomic.AddInt64(&successes, 1)
				} else {
					atomic.AddInt64(&failures, 1)
				}
			}
		}()
	}
	wg.Wait()

	totalRequests := successes + failures
	avgMs := float64(0)
	if totalRequests > 0 {
		avgMs = float64(totalLatencyMs) / float64(totalRequests)
	}

	outputs["total_requests"] = fmt.Sprintf("%d", totalRequests)
	outputs["successes"] = fmt.Sprintf("%d", successes)
	outputs["avg_latency_ms"] = fmt.Sprintf("%.1f", avgMs)

	// All requests should succeed
	successRate := float64(0)
	if totalRequests > 0 {
		successRate = float64(successes) / float64(totalRequests) * 100
	}
	successPassed := successRate >= 90
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "file_listing_success_rate",
		Expected: ">=90% success rate",
		Actual:   fmt.Sprintf("%.1f%% (%d/%d)", successRate, successes, totalRequests),
		Passed:   successPassed,
		Message: challenge.Ternary(successPassed,
			fmt.Sprintf("File listing success rate: %.1f%%", successRate),
			fmt.Sprintf("File listing degraded: %.1f%% success rate", successRate)),
	})

	// Average latency under 200ms
	maxAvgLatency := float64(200)
	latencyPassed := avgMs < maxAvgLatency
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "max_latency",
		Target:   "file_listing_avg_latency",
		Expected: fmt.Sprintf("<%.0fms average under load", maxAvgLatency),
		Actual:   fmt.Sprintf("%.1fms", avgMs),
		Passed:   latencyPassed,
		Message: challenge.Ternary(latencyPassed,
			fmt.Sprintf("File listing avg latency %.1fms < %.0fms threshold", avgMs, maxAvgLatency),
			fmt.Sprintf("File listing avg latency %.1fms exceeds %.0fms threshold", avgMs, maxAvgLatency)),
	})

	metrics := map[string]challenge.MetricValue{
		"file_listing_avg_latency": {
			Name:  "file_listing_avg_latency",
			Value: avgMs,
			Unit:  "ms",
		},
		"file_listing_success_rate": {
			Name:  "file_listing_success_rate",
			Value: successRate,
			Unit:  "percent",
		},
		"file_listing_total_requests": {
			Name:  "file_listing_total_requests",
			Value: float64(totalRequests),
			Unit:  "count",
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
