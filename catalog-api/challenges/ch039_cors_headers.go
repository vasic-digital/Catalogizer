package challenges

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// CORSHeadersChallenge validates that CORS headers are correctly
// set on API responses. Checks for Access-Control-Allow-Origin,
// Access-Control-Allow-Methods, and Access-Control-Allow-Headers.
type CORSHeadersChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCORSHeadersChallenge creates CH-039.
func NewCORSHeadersChallenge() *CORSHeadersChallenge {
	return &CORSHeadersChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"cors-headers",
			"CORS Headers",
			"Validates CORS headers are correctly set: checks "+
				"Access-Control-Allow-Origin, Allow-Methods, and "+
				"Allow-Headers on API responses and preflight requests.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the CORS headers challenge.
func (c *CORSHeadersChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Send a regular GET to /health and check CORS headers
	c.ReportProgress("get-cors", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, c.config.BaseURL+"/health", nil,
	)
	req.Header.Set("Origin", "http://localhost:3000")

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "health_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Health endpoint unreachable: %v", err),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, err.Error(),
		), nil
	}
	resp.Body.Close()

	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	hasOrigin := allowOrigin != ""
	outputs["allow_origin"] = allowOrigin

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "access_control_allow_origin",
		Expected: "non-empty Access-Control-Allow-Origin",
		Actual:   challenge.Ternary(hasOrigin, allowOrigin, "missing"),
		Passed:   hasOrigin,
		Message: challenge.Ternary(hasOrigin,
			fmt.Sprintf("CORS Allow-Origin present: %s", allowOrigin),
			"Access-Control-Allow-Origin header missing"),
	})

	// Step 2: Send OPTIONS preflight request
	c.ReportProgress("preflight", nil)
	optReq, _ := http.NewRequestWithContext(
		ctx, http.MethodOptions, c.config.BaseURL+"/api/v1/auth/login", nil,
	)
	optReq.Header.Set("Origin", "http://localhost:3000")
	optReq.Header.Set("Access-Control-Request-Method", "POST")
	optReq.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	optResp, optErr := httpClient.Do(optReq)
	if optErr != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "preflight_request",
			Expected: "preflight response",
			Actual:   fmt.Sprintf("err=%v", optErr),
			Passed:   false,
			Message:  fmt.Sprintf("Preflight request failed: %v", optErr),
		})
	} else {
		optResp.Body.Close()

		// Preflight should return 200 or 204
		preflightOK := optResp.StatusCode == 200 || optResp.StatusCode == 204
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "preflight_status",
			Expected: "200 or 204",
			Actual:   fmt.Sprintf("HTTP %d", optResp.StatusCode),
			Passed:   preflightOK,
			Message: challenge.Ternary(preflightOK,
				fmt.Sprintf("Preflight returned HTTP %d", optResp.StatusCode),
				fmt.Sprintf("Unexpected preflight status: %d", optResp.StatusCode)),
		})

		allowMethods := optResp.Header.Get("Access-Control-Allow-Methods")
		hasMethods := allowMethods != ""
		outputs["allow_methods"] = allowMethods
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "access_control_allow_methods",
			Expected: "non-empty Access-Control-Allow-Methods",
			Actual:   challenge.Ternary(hasMethods, allowMethods, "missing"),
			Passed:   hasMethods,
			Message: challenge.Ternary(hasMethods,
				fmt.Sprintf("CORS Allow-Methods present: %s", allowMethods),
				"Access-Control-Allow-Methods header missing from preflight"),
		})

		allowHeaders := optResp.Header.Get("Access-Control-Allow-Headers")
		hasHeaders := allowHeaders != ""
		outputs["allow_headers"] = allowHeaders
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "access_control_allow_headers",
			Expected: "non-empty Access-Control-Allow-Headers",
			Actual:   challenge.Ternary(hasHeaders, allowHeaders, "missing"),
			Passed:   hasHeaders,
			Message: challenge.Ternary(hasHeaders,
				fmt.Sprintf("CORS Allow-Headers present: %s", allowHeaders),
				"Access-Control-Allow-Headers header missing from preflight"),
		})
	}

	metrics := map[string]challenge.MetricValue{
		"cors_check_time": {
			Name:  "cors_check_time",
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
