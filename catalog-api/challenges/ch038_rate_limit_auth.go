package challenges

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// RateLimitAuthChallenge validates that rate limiting is active on
// authentication endpoints. Sends rapid login requests and verifies
// that the server either rate-limits (429) or handles them cleanly
// without server errors.
type RateLimitAuthChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewRateLimitAuthChallenge creates CH-038.
func NewRateLimitAuthChallenge() *RateLimitAuthChallenge {
	return &RateLimitAuthChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"rate-limit-auth",
			"Rate Limiting on Auth Endpoints",
			"Sends rapid login and register requests to verify rate "+
				"limiting is active. Accepts either 429 responses or "+
				"clean handling without 5xx errors.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the rate limit auth challenge.
func (c *RateLimitAuthChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Verify auth endpoint is reachable first
	c.ReportProgress("verify-reachable", nil)
	apiClient := httpclient.NewAPIClient(c.config.BaseURL)
	_, loginErr := apiClient.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	loginOK := loginErr == nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "auth_reachable",
		Expected: "login succeeds",
		Actual:   challenge.Ternary(loginOK, "succeeded", fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message: challenge.Ternary(loginOK,
			"Auth endpoint reachable",
			fmt.Sprintf("Auth endpoint unreachable: %v", loginErr)),
	})

	if !loginOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "auth unreachable",
		), nil
	}

	// Step 2: Rapid login requests with invalid credentials
	c.ReportProgress("rapid-login", nil)
	rapidCount := 30
	status429 := 0
	status401 := 0
	status5xx := 0

	for i := 0; i < rapidCount; i++ {
		body := `{"username":"rate_limit_test","password":"invalid"}`
		req, _ := http.NewRequestWithContext(
			ctx, http.MethodPost,
			c.config.BaseURL+"/api/v1/auth/login",
			strings.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		switch {
		case resp.StatusCode == 429:
			status429++
		case resp.StatusCode == 401:
			status401++
		case resp.StatusCode >= 500:
			status5xx++
		}
	}

	outputs["rapid_requests"] = fmt.Sprintf("%d", rapidCount)
	outputs["status_429"] = fmt.Sprintf("%d", status429)
	outputs["status_401"] = fmt.Sprintf("%d", status401)
	outputs["status_5xx"] = fmt.Sprintf("%d", status5xx)

	// No server errors during rapid requests
	noServerErrors := status5xx == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "no_server_errors",
		Expected: "zero 5xx responses",
		Actual:   fmt.Sprintf("5xx=%d", status5xx),
		Passed:   noServerErrors,
		Message: challenge.Ternary(noServerErrors,
			"No server errors during rapid auth requests",
			fmt.Sprintf("%d server errors during rapid auth requests", status5xx)),
	})

	// Rate limiting active or requests handled cleanly
	rateLimitOrClean := status429 > 0 || (status401+status429) == rapidCount
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "rate_limit_active",
		Expected: "429 responses or all handled cleanly",
		Actual:   fmt.Sprintf("429=%d, 401=%d", status429, status401),
		Passed:   rateLimitOrClean,
		Message: challenge.Ternary(status429 > 0,
			fmt.Sprintf("Rate limiting active: %d/%d requests rate-limited", status429, rapidCount),
			fmt.Sprintf("All %d requests handled cleanly (rate limiter may not be configured)", rapidCount)),
	})

	// Step 3: Verify valid credentials still work after rapid requests
	c.ReportProgress("post-rate-limit", nil)
	time.Sleep(1 * time.Second)
	freshClient := httpclient.NewAPIClient(c.config.BaseURL)
	_, postErr := freshClient.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	postOK := postErr == nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "post_rate_limit_login",
		Expected: "valid login succeeds after rate limiting",
		Actual:   challenge.Ternary(postOK, "succeeded", fmt.Sprintf("err=%v", postErr)),
		Passed:   postOK,
		Message: challenge.Ternary(postOK,
			"Valid credentials still accepted after rate limit test",
			fmt.Sprintf("Login failed after rate limit test: %v", postErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"rate_limited_count": {
			Name:  "rate_limited_count",
			Value: float64(status429),
			Unit:  "count",
		},
		"total_rapid_requests": {
			Name:  "total_rapid_requests",
			Value: float64(rapidCount),
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
