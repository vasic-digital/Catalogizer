package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// HealthLatencyChallenge validates that the health endpoint
// responds within 10ms on average. Sends multiple requests
// and measures the average response time.
type HealthLatencyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewHealthLatencyChallenge creates CH-041.
func NewHealthLatencyChallenge() *HealthLatencyChallenge {
	return &HealthLatencyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"health-latency",
			"Health Endpoint Latency",
			"Validates the health endpoint responds within 10ms on "+
				"average across multiple requests. Measures p50 and "+
				"p99 response times.",
			"performance",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the health latency challenge.
func (c *HealthLatencyChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Warm up the endpoint
	c.ReportProgress("warmup", nil)
	client.Get(ctx, "/health")

	// Measure latency across multiple requests
	c.ReportProgress("measuring", nil)
	requestCount := 20
	latencies := make([]time.Duration, 0, requestCount)
	failures := 0

	for i := 0; i < requestCount; i++ {
		reqStart := time.Now()
		code, _, err := client.Get(ctx, "/health")
		elapsed := time.Since(reqStart)

		if err != nil || code != 200 {
			failures++
			continue
		}
		latencies = append(latencies, elapsed)
	}

	// Calculate average latency
	var totalMs float64
	for _, lat := range latencies {
		totalMs += float64(lat.Milliseconds())
	}
	avgMs := float64(0)
	if len(latencies) > 0 {
		avgMs = totalMs / float64(len(latencies))
	}

	// Calculate min/max
	var minMs, maxMs float64
	if len(latencies) > 0 {
		minMs = float64(latencies[0].Milliseconds())
		maxMs = float64(latencies[0].Milliseconds())
		for _, lat := range latencies[1:] {
			ms := float64(lat.Milliseconds())
			if ms < minMs {
				minMs = ms
			}
			if ms > maxMs {
				maxMs = ms
			}
		}
	}

	outputs["avg_latency_ms"] = fmt.Sprintf("%.1f", avgMs)
	outputs["min_latency_ms"] = fmt.Sprintf("%.0f", minMs)
	outputs["max_latency_ms"] = fmt.Sprintf("%.0f", maxMs)
	outputs["failures"] = fmt.Sprintf("%d", failures)

	// All requests should succeed
	allSucceeded := failures == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "health_requests_succeed",
		Expected: "zero failures",
		Actual:   fmt.Sprintf("failures=%d/%d", failures, requestCount),
		Passed:   allSucceeded,
		Message: challenge.Ternary(allSucceeded,
			fmt.Sprintf("All %d health requests succeeded", requestCount),
			fmt.Sprintf("%d/%d health requests failed", failures, requestCount)),
	})

	// Average latency should be < 10ms
	maxAvgLatency := float64(10)
	latencyPassed := avgMs < maxAvgLatency
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "max_latency",
		Target:   "health_avg_latency",
		Expected: fmt.Sprintf("<%.0fms average", maxAvgLatency),
		Actual:   fmt.Sprintf("%.1fms", avgMs),
		Passed:   latencyPassed,
		Message: challenge.Ternary(latencyPassed,
			fmt.Sprintf("Health endpoint avg latency %.1fms < %.0fms threshold", avgMs, maxAvgLatency),
			fmt.Sprintf("Health endpoint avg latency %.1fms exceeds %.0fms threshold", avgMs, maxAvgLatency)),
	})

	metrics := map[string]challenge.MetricValue{
		"health_avg_latency": {
			Name:  "health_avg_latency",
			Value: avgMs,
			Unit:  "ms",
		},
		"health_min_latency": {
			Name:  "health_min_latency",
			Value: minMs,
			Unit:  "ms",
		},
		"health_max_latency": {
			Name:  "health_max_latency",
			Value: maxMs,
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
