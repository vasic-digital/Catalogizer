package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntitySearchLatencyChallenge validates that the entity search
// endpoint responds within 500ms on average. Performs multiple
// search queries and measures response time.
type EntitySearchLatencyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntitySearchLatencyChallenge creates CH-043.
func NewEntitySearchLatencyChallenge() *EntitySearchLatencyChallenge {
	return &EntitySearchLatencyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-search-latency",
			"Entity Search Latency",
			"Validates entity search responds within 500ms on average. "+
				"Tests multiple search queries with different terms and "+
				"measures response time.",
			"performance",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity search latency challenge.
func (c *EntitySearchLatencyChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login
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

	// Warm up
	client.Get(ctx, "/api/v1/entities?limit=1")

	// Search queries to test
	c.ReportProgress("searching", nil)
	searchQueries := []string{
		"/api/v1/entities?query=a&limit=10",
		"/api/v1/entities?query=the&limit=10",
		"/api/v1/entities?query=test&limit=10",
		"/api/v1/entities?limit=20",
		"/api/v1/entities?query=music&limit=10",
	}

	latencies := make([]float64, 0, len(searchQueries))
	failures := 0

	for _, query := range searchQueries {
		reqStart := time.Now()
		code, _, err := client.Get(ctx, query)
		elapsed := float64(time.Since(reqStart).Milliseconds())

		if err == nil && code == 200 {
			latencies = append(latencies, elapsed)
		} else {
			failures++
		}
	}

	// Calculate average
	var totalMs float64
	for _, lat := range latencies {
		totalMs += lat
	}
	avgMs := float64(0)
	if len(latencies) > 0 {
		avgMs = totalMs / float64(len(latencies))
	}

	// Calculate max
	maxMs := float64(0)
	for _, lat := range latencies {
		if lat > maxMs {
			maxMs = lat
		}
	}

	outputs["avg_latency_ms"] = fmt.Sprintf("%.1f", avgMs)
	outputs["max_latency_ms"] = fmt.Sprintf("%.0f", maxMs)
	outputs["queries_tested"] = fmt.Sprintf("%d", len(searchQueries))
	outputs["failures"] = fmt.Sprintf("%d", failures)

	// Queries should mostly succeed
	successCount := len(latencies)
	totalCount := len(searchQueries)
	successPassed := float64(successCount)/float64(totalCount) >= 0.8
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_score",
		Target:   "search_success_rate",
		Expected: ">=80% queries succeed",
		Actual:   fmt.Sprintf("%d/%d succeeded", successCount, totalCount),
		Passed:   successPassed,
		Message: challenge.Ternary(successPassed,
			fmt.Sprintf("Search queries: %d/%d succeeded", successCount, totalCount),
			fmt.Sprintf("Too many search failures: %d/%d failed", failures, totalCount)),
	})

	// Average latency under 500ms
	maxAvgLatency := float64(500)
	latencyPassed := avgMs < maxAvgLatency || len(latencies) == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "max_latency",
		Target:   "entity_search_avg_latency",
		Expected: fmt.Sprintf("<%.0fms average", maxAvgLatency),
		Actual:   fmt.Sprintf("%.1fms", avgMs),
		Passed:   latencyPassed,
		Message: challenge.Ternary(latencyPassed,
			fmt.Sprintf("Entity search avg latency %.1fms < %.0fms threshold", avgMs, maxAvgLatency),
			fmt.Sprintf("Entity search avg latency %.1fms exceeds %.0fms threshold", avgMs, maxAvgLatency)),
	})

	metrics := map[string]challenge.MetricValue{
		"entity_search_avg_latency": {
			Name:  "entity_search_avg_latency",
			Value: avgMs,
			Unit:  "ms",
		},
		"entity_search_max_latency": {
			Name:  "entity_search_max_latency",
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
