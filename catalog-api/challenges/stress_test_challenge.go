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

// StressTestChallenge validates that the API can handle concurrent
// requests to health and entity endpoints without errors or
// excessive latency. Sends parallel requests and measures
// success rate and response times.
type StressTestChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewStressTestChallenge creates CH-026.
func NewStressTestChallenge() *StressTestChallenge {
	return &StressTestChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"stress-test",
			"API Stress Test",
			"Sends concurrent requests to health and entity endpoints, "+
				"verifies all respond within timeout and measures "+
				"success rate under load.",
			"stress",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the API stress test challenge.
func (c *StressTestChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
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

	// Step 1: Concurrent health endpoint requests
	concurrency := 10
	requestsPerEndpoint := 20
	var healthSuccesses int64
	var healthFailures int64
	var healthTotalMs int64

	c.ReportProgress("stress-health", map[string]any{
		"concurrency":  concurrency,
		"total_requests": requestsPerEndpoint,
	})

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			perWorker := requestsPerEndpoint / concurrency
			for j := 0; j < perWorker; j++ {
				reqStart := time.Now()
				code, _, reqErr := client.Get(ctx, "/health")
				elapsed := time.Since(reqStart).Milliseconds()
				atomic.AddInt64(&healthTotalMs, elapsed)
				if reqErr == nil && code == 200 {
					atomic.AddInt64(&healthSuccesses, 1)
				} else {
					atomic.AddInt64(&healthFailures, 1)
				}
			}
		}()
	}
	wg.Wait()

	healthTotal := healthSuccesses + healthFailures
	healthRate := float64(0)
	if healthTotal > 0 {
		healthRate = float64(healthSuccesses) / float64(healthTotal) * 100
	}
	healthAvgMs := float64(0)
	if healthTotal > 0 {
		healthAvgMs = float64(healthTotalMs) / float64(healthTotal)
	}

	healthPassed := healthRate >= 95.0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "health_endpoint_success_rate",
		Expected: ">=95% success rate",
		Actual:   fmt.Sprintf("%.1f%% (%d/%d)", healthRate, healthSuccesses, healthTotal),
		Passed:   healthPassed,
		Message: challenge.Ternary(healthPassed,
			fmt.Sprintf("Health endpoint: %.1f%% success rate, avg %.0fms", healthRate, healthAvgMs),
			fmt.Sprintf("Health endpoint degraded: %.1f%% success rate", healthRate)),
	})
	outputs["health_success_rate"] = fmt.Sprintf("%.1f%%", healthRate)
	outputs["health_avg_latency_ms"] = fmt.Sprintf("%.0f", healthAvgMs)

	// Step 2: Concurrent entity endpoint requests
	var entitySuccesses int64
	var entityFailures int64
	var entityTotalMs int64

	c.ReportProgress("stress-entities", map[string]any{
		"concurrency":  concurrency,
		"total_requests": requestsPerEndpoint,
	})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			perWorker := requestsPerEndpoint / concurrency
			for j := 0; j < perWorker; j++ {
				reqStart := time.Now()
				code, _, reqErr := client.Get(ctx, "/api/v1/entities?limit=5")
				elapsed := time.Since(reqStart).Milliseconds()
				atomic.AddInt64(&entityTotalMs, elapsed)
				if reqErr == nil && code == 200 {
					atomic.AddInt64(&entitySuccesses, 1)
				} else {
					atomic.AddInt64(&entityFailures, 1)
				}
			}
		}()
	}
	wg.Wait()

	entityTotal := entitySuccesses + entityFailures
	entityRate := float64(0)
	if entityTotal > 0 {
		entityRate = float64(entitySuccesses) / float64(entityTotal) * 100
	}
	entityAvgMs := float64(0)
	if entityTotal > 0 {
		entityAvgMs = float64(entityTotalMs) / float64(entityTotal)
	}

	entityPassed := entityRate >= 95.0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "entity_endpoint_success_rate",
		Expected: ">=95% success rate",
		Actual:   fmt.Sprintf("%.1f%% (%d/%d)", entityRate, entitySuccesses, entityTotal),
		Passed:   entityPassed,
		Message: challenge.Ternary(entityPassed,
			fmt.Sprintf("Entity endpoint: %.1f%% success rate, avg %.0fms", entityRate, entityAvgMs),
			fmt.Sprintf("Entity endpoint degraded: %.1f%% success rate", entityRate)),
	})
	outputs["entity_success_rate"] = fmt.Sprintf("%.1f%%", entityRate)
	outputs["entity_avg_latency_ms"] = fmt.Sprintf("%.0f", entityAvgMs)

	// Step 3: Verify average latency is within acceptable bounds (< 2000ms)
	maxAvgLatency := float64(2000)
	overallAvg := (healthAvgMs + entityAvgMs) / 2
	latencyPassed := overallAvg < maxAvgLatency
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "max_latency",
		Target:   "overall_avg_latency",
		Expected: fmt.Sprintf("<%.0fms", maxAvgLatency),
		Actual:   fmt.Sprintf("%.0fms", overallAvg),
		Passed:   latencyPassed,
		Message: challenge.Ternary(latencyPassed,
			fmt.Sprintf("Average latency %.0fms is within bounds", overallAvg),
			fmt.Sprintf("Average latency %.0fms exceeds %.0fms threshold", overallAvg, maxAvgLatency)),
	})

	metrics := map[string]challenge.MetricValue{
		"health_success_rate": {
			Name:  "health_success_rate",
			Value: healthRate,
			Unit:  "percent",
		},
		"entity_success_rate": {
			Name:  "entity_success_rate",
			Value: entityRate,
			Unit:  "percent",
		},
		"health_avg_latency": {
			Name:  "health_avg_latency",
			Value: healthAvgMs,
			Unit:  "ms",
		},
		"entity_avg_latency": {
			Name:  "entity_avg_latency",
			Value: entityAvgMs,
			Unit:  "ms",
		},
		"total_requests": {
			Name:  "total_requests",
			Value: float64(healthTotal + entityTotal),
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

	return c.CreateResult(
		status, start, assertions, metrics, outputs, "",
	), nil
}
