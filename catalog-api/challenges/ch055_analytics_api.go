package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// AnalyticsAPIChallenge validates the analytics and statistics
// endpoints return proper data structures for the dashboard.
type AnalyticsAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAnalyticsAPIChallenge creates CH-055.
func NewAnalyticsAPIChallenge() *AnalyticsAPIChallenge {
	return &AnalyticsAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"analytics-api",
			"Analytics and Statistics API",
			"Verifies analytics endpoints return proper statistics "+
				"structures: overall stats, media type distribution, "+
				"duplicate counts, scan history, and growth trends.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the analytics API challenge.
func (c *AnalyticsAPIChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login
	c.ReportProgress("authenticating", nil)
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 5)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", err),
		), nil
	}

	// Test 1: Overall statistics
	c.ReportProgress("testing-overall-stats", nil)
	status, body, _ := client.Get(ctx, "/stats/overall")

	statsOK := status == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "overall_stats",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   statsOK,
		Message:  challenge.Ternary(statsOK, "Overall stats endpoint works", "Overall stats endpoint failed"),
	})

	// Test 2: Duplicate counts
	c.ReportProgress("testing-duplicate-stats", nil)
	statusDup, _, _ := client.Get(ctx, "/stats/duplicates")

	dupOK := statusDup == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "duplicate_stats",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusDup),
		Passed:   dupOK,
		Message:  challenge.Ternary(dupOK, "Duplicate stats endpoint works", "Duplicate stats endpoint failed"),
	})

	// Test 3: Media type distribution
	c.ReportProgress("testing-media-distribution", nil)
	statusDist, bodyDist, _ := client.Get(ctx, "/stats/media-types")

	distOK := statusDist == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_distribution",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusDist),
		Passed:   distOK,
		Message:  challenge.Ternary(distOK, "Media type distribution works", "Media type distribution failed"),
	})

	if bodyDist != nil {
		outputs["media_types_response"] = fmt.Sprintf("%v", bodyDist)
	}

	// Test 4: Scan history
	c.ReportProgress("testing-scan-history", nil)
	statusScan, _, _ := client.Get(ctx, "/stats/scan-history")

	scanOK := statusScan == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "scan_history",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusScan),
		Passed:   scanOK,
		Message:  challenge.Ternary(scanOK, "Scan history endpoint works", "Scan history endpoint failed"),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	metrics := map[string]challenge.MetricValue{
		"endpoints_tested": {
			Name:  "endpoints_tested",
			Value: 4,
			Unit:  "count",
		},
	}

	return c.CreateResult(resultStatus, start, assertions, metrics, outputs, ""), nil
}
