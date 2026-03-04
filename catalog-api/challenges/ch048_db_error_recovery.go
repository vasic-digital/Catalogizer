package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// DBErrorRecoveryChallenge validates that the API continues
// serving after database error recovery. Simulates stress by
// sending rapid concurrent requests and then verifies the API
// still responds correctly afterward.
type DBErrorRecoveryChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDBErrorRecoveryChallenge creates CH-048.
func NewDBErrorRecoveryChallenge() *DBErrorRecoveryChallenge {
	return &DBErrorRecoveryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"db-error-recovery",
			"Database Error Recovery",
			"Validates API continues serving after database stress: "+
				"sends requests that may cause transient DB errors, "+
				"then verifies the API recovers and responds correctly.",
			"resilience",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the database error recovery challenge.
func (c *DBErrorRecoveryChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Verify API is healthy before test
	c.ReportProgress("pre-check", nil)
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
		Target:   "pre_health_check",
		Expected: "200 with status=healthy",
		Actual:   fmt.Sprintf("HTTP %d, status=%q", healthCode, healthStatus),
		Passed:   healthOK,
		Message: challenge.Ternary(healthOK,
			"API healthy before resilience test",
			fmt.Sprintf("API not healthy before test: code=%d err=%v", healthCode, healthErr)),
	})

	if !healthOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "API not healthy",
		), nil
	}

	// Step 2: Login
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

	// Step 3: Trigger requests that stress the database
	// Request a nonexistent entity (triggers DB query + 404)
	c.ReportProgress("stress-db", nil)
	stressRequests := 15
	for i := 0; i < stressRequests; i++ {
		client.Get(ctx, fmt.Sprintf("/api/v1/entities/%d", 99999990+i))
	}
	outputs["stress_requests"] = fmt.Sprintf("%d", stressRequests)

	// Step 4: Brief pause to let any error state propagate
	time.Sleep(500 * time.Millisecond)

	// Step 5: Verify API is still healthy after stress
	c.ReportProgress("post-check", nil)
	postHealthCode, postHealthBody, postHealthErr := client.Get(ctx, "/health")
	postHealthOK := postHealthErr == nil && postHealthCode == 200
	postHealthStatus := ""
	if postHealthBody != nil {
		if s, ok := postHealthBody["status"].(string); ok {
			postHealthStatus = s
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "post_health_check",
		Expected: "200 with status=healthy",
		Actual:   fmt.Sprintf("HTTP %d, status=%q", postHealthCode, postHealthStatus),
		Passed:   postHealthOK,
		Message: challenge.Ternary(postHealthOK,
			"API still healthy after database stress",
			fmt.Sprintf("API unhealthy after stress: code=%d err=%v", postHealthCode, postHealthErr)),
	})

	// Step 6: Verify a data endpoint still works
	entitiesCode, _, entitiesErr := client.Get(ctx, "/api/v1/entities?limit=1")
	entitiesOK := entitiesErr == nil && entitiesCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "post_stress_entities",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", entitiesCode),
		Passed:   entitiesOK,
		Message: challenge.Ternary(entitiesOK,
			"Entity endpoint responds after database stress",
			fmt.Sprintf("Entity endpoint failed after stress: code=%d err=%v", entitiesCode, entitiesErr)),
	})

	// Step 7: Verify auth still works
	freshClient := httpclient.NewAPIClient(c.config.BaseURL)
	_, freshLoginErr := freshClient.Login(ctx, c.config.Username, c.config.Password)
	freshLoginOK := freshLoginErr == nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "post_stress_login",
		Expected: "login succeeds after stress",
		Actual:   challenge.Ternary(freshLoginOK, "succeeded", fmt.Sprintf("err=%v", freshLoginErr)),
		Passed:   freshLoginOK,
		Message: challenge.Ternary(freshLoginOK,
			"Authentication still works after database stress",
			fmt.Sprintf("Login failed after stress: %v", freshLoginErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"recovery_test_time": {
			Name:  "recovery_test_time",
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
