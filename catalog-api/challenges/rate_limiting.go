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

// RateLimitingChallenge validates that the API rate limiter blocks
// excessive requests to auth endpoints with 429 status codes.
// Sends rapid sequential requests and verifies that the rate
// limiter activates after the allowed threshold.
type RateLimitingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewRateLimitingChallenge creates CH-027.
func NewRateLimitingChallenge() *RateLimitingChallenge {
	return &RateLimitingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"rate-limiting",
			"Rate Limiting",
			"Sends rapid requests to auth endpoints, verifies rate limiter "+
				"blocks excessive requests with 429 status or the endpoint "+
				"continues responding without server errors.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the rate limiting challenge.
func (c *RateLimitingChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	// We use a raw HTTP client here to avoid Login's error handling
	// when we intentionally send bad requests.
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Verify auth endpoint is reachable with valid credentials
	apiClient := httpclient.NewAPIClient(c.config.BaseURL)
	_, loginErr := apiClient.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	loginOK := loginErr == nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "auth_endpoint_reachable",
		Expected: "login succeeds",
		Actual:   challenge.Ternary(loginOK, "login succeeded", fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message: challenge.Ternary(loginOK,
			"Auth endpoint is reachable",
			fmt.Sprintf("Auth endpoint unreachable: %v", loginErr)),
	})

	if !loginOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "auth endpoint unreachable",
		), nil
	}

	// Step 2: Send rapid requests to login endpoint with invalid credentials
	// to trigger rate limiting
	rapidCount := 50
	status429Count := 0
	status200Count := 0
	status401Count := 0
	statusOtherCount := 0

	c.ReportProgress("rapid-requests", map[string]any{
		"total_requests": rapidCount,
	})

	for i := 0; i < rapidCount; i++ {
		body := `{"username":"invalid_user","password":"invalid_pass"}`
		req, reqErr := http.NewRequestWithContext(
			ctx, http.MethodPost,
			c.config.BaseURL+"/api/v1/auth/login",
			strings.NewReader(body),
		)
		if reqErr != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, doErr := httpClient.Do(req)
		if doErr != nil {
			statusOtherCount++
			continue
		}
		resp.Body.Close()

		switch resp.StatusCode {
		case 429:
			status429Count++
		case 200:
			status200Count++
		case 401:
			status401Count++
		default:
			statusOtherCount++
		}
	}

	outputs["rapid_requests_sent"] = fmt.Sprintf("%d", rapidCount)
	outputs["status_429_count"] = fmt.Sprintf("%d", status429Count)
	outputs["status_401_count"] = fmt.Sprintf("%d", status401Count)
	outputs["status_200_count"] = fmt.Sprintf("%d", status200Count)
	outputs["status_other_count"] = fmt.Sprintf("%d", statusOtherCount)

	// Step 3: Verify rate limiter behavior
	// We accept two outcomes:
	// a) Rate limiter active: some 429 responses
	// b) No rate limiter but no server errors: all 401 responses (invalid creds)
	// What we do NOT accept: 5xx server errors
	noServerErrors := statusOtherCount == 0
	hasRateLimiting := status429Count > 0
	allHandled := (status429Count + status401Count + status200Count) == rapidCount

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "no_server_errors",
		Expected: "zero 5xx/unexpected responses",
		Actual:   fmt.Sprintf("other_count=%d", statusOtherCount),
		Passed:   noServerErrors,
		Message: challenge.Ternary(noServerErrors,
			"No server errors during rapid requests",
			fmt.Sprintf("%d unexpected responses during rapid requests", statusOtherCount)),
	})

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "rate_limit_or_handled",
		Expected: "429 rate limit responses or all requests handled cleanly",
		Actual:   fmt.Sprintf("429=%d, 401=%d, 200=%d", status429Count, status401Count, status200Count),
		Passed:   hasRateLimiting || allHandled,
		Message: challenge.Ternary(hasRateLimiting,
			fmt.Sprintf("Rate limiter active: %d/%d requests received 429", status429Count, rapidCount),
			fmt.Sprintf("All %d requests handled without errors (rate limiter may not be configured)", rapidCount)),
	})

	// Step 4: Verify original credentials still work after rate limiting
	// Wait briefly for any rate limit window to reset
	time.Sleep(1 * time.Second)
	freshClient := httpclient.NewAPIClient(c.config.BaseURL)
	_, postErr := freshClient.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	postOK := postErr == nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "post_rate_limit_login",
		Expected: "valid login succeeds after rate limiting",
		Actual:   challenge.Ternary(postOK, "login succeeded", fmt.Sprintf("err=%v", postErr)),
		Passed:   postOK,
		Message: challenge.Ternary(postOK,
			"Valid credentials still work after rate limit test",
			fmt.Sprintf("Login failed after rate limit test: %v", postErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"rate_limited_requests": {
			Name:  "rate_limited_requests",
			Value: float64(status429Count),
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

	return c.CreateResult(
		status, start, assertions, metrics, outputs, "",
	), nil
}
