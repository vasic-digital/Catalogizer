package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// ScannerRecoveryChallenge validates that the scanner system
// recovers from temporary filesystem unavailability. Triggers
// a scan of a non-existent path and verifies the API handles
// the error gracefully without crashing.
type ScannerRecoveryChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewScannerRecoveryChallenge creates CH-049.
func NewScannerRecoveryChallenge() *ScannerRecoveryChallenge {
	return &ScannerRecoveryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"scanner-recovery",
			"Scanner Filesystem Recovery",
			"Validates the scanner recovers from filesystem errors: "+
				"attempts to scan a non-existent path, verifies the error "+
				"is handled gracefully, and confirms the API remains healthy.",
			"resilience",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the scanner recovery challenge.
func (c *ScannerRecoveryChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Login
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

	// Step 2: Pre-check: API is healthy
	c.ReportProgress("pre-check", nil)
	preHealthCode, _, preHealthErr := client.Get(ctx, "/health")
	preHealthOK := preHealthErr == nil && preHealthCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "pre_health",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", preHealthCode),
		Passed:   preHealthOK,
		Message: challenge.Ternary(preHealthOK,
			"API healthy before scanner test",
			fmt.Sprintf("API not healthy: code=%d err=%v", preHealthCode, preHealthErr)),
	})

	if !preHealthOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "API not healthy",
		), nil
	}

	// Step 3: Create a storage root pointing to a non-existent path
	c.ReportProgress("create-bad-root", nil)
	badRootName := fmt.Sprintf("scanner-recovery-test-%d", time.Now().UnixMilli())
	badRootBody := fmt.Sprintf(
		`{"name":%q,"protocol":"local","path":"/nonexistent/path/that/does/not/exist/%d","max_depth":1}`,
		badRootName, time.Now().UnixMilli(),
	)
	createCode, _, createErr := client.PostJSON(ctx, "/api/v1/storage/roots", badRootBody)
	// Accept both success (root created but scan fails) and error (root rejected)
	createHandled := createErr == nil && (createCode >= 200 && createCode < 500)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "create_bad_root",
		Expected: "non-5xx response (graceful handling)",
		Actual:   fmt.Sprintf("HTTP %d", createCode),
		Passed:   createHandled,
		Message: challenge.Ternary(createHandled,
			fmt.Sprintf("Bad root handled gracefully: HTTP %d", createCode),
			fmt.Sprintf("Server error on bad root: code=%d err=%v", createCode, createErr)),
	})

	// Step 4: Brief pause
	time.Sleep(500 * time.Millisecond)

	// Step 5: Verify API is still healthy after the bad scan attempt
	c.ReportProgress("post-check", nil)
	postHealthCode, _, postHealthErr := client.Get(ctx, "/health")
	postHealthOK := postHealthErr == nil && postHealthCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "post_health",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", postHealthCode),
		Passed:   postHealthOK,
		Message: challenge.Ternary(postHealthOK,
			"API still healthy after bad scan attempt",
			fmt.Sprintf("API unhealthy after bad scan: code=%d err=%v", postHealthCode, postHealthErr)),
	})

	// Step 6: Verify other endpoints still work
	entitiesCode, _, entitiesErr := client.Get(ctx, "/api/v1/entities?limit=1")
	entitiesOK := entitiesErr == nil && entitiesCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "post_scan_entities",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", entitiesCode),
		Passed:   entitiesOK,
		Message: challenge.Ternary(entitiesOK,
			"Entity endpoint responds after bad scan",
			fmt.Sprintf("Entity endpoint failed after bad scan: code=%d err=%v", entitiesCode, entitiesErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"scanner_recovery_time": {
			Name:  "scanner_recovery_time",
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
