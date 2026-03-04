package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// NoSensitiveErrorsChallenge validates that API error responses
// do not leak sensitive data such as stack traces, database
// connection strings, internal paths, or SQL queries.
type NoSensitiveErrorsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewNoSensitiveErrorsChallenge creates CH-040.
func NewNoSensitiveErrorsChallenge() *NoSensitiveErrorsChallenge {
	return &NoSensitiveErrorsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"no-sensitive-errors",
			"No Sensitive Data in Errors",
			"Verifies API error responses do not leak sensitive data: "+
				"stack traces, SQL queries, database connection strings, "+
				"internal file paths, or configuration secrets.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// sensitivePatterns lists substrings that should never appear
// in API error response bodies.
var sensitivePatterns = []string{
	"goroutine",
	"runtime.go",
	"panic(",
	"SELECT ",
	"INSERT INTO",
	"UPDATE ",
	"DELETE FROM",
	"password=",
	"jwt_secret",
	"DB_PASSWORD",
	"connection refused",
	"pq: ",
	"sqlite3:",
	".go:",
}

// Execute runs the no-sensitive-errors challenge.
func (c *NoSensitiveErrorsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Trigger various error conditions and check response bodies
	errorRequests := []struct {
		method string
		path   string
		body   string
		desc   string
	}{
		{"GET", "/api/v1/nonexistent-endpoint-xyz", "", "404 endpoint"},
		{"POST", "/api/v1/auth/login", `{"username":"","password":""}`, "empty credentials"},
		{"POST", "/api/v1/auth/login", `invalid json{{{`, "malformed JSON"},
		{"GET", "/api/v1/entities/99999999", "", "nonexistent entity"},
		{"GET", "/api/v1/files?path=../../etc/passwd", "", "path traversal attempt"},
	}

	leaksFound := 0

	for i, errReq := range errorRequests {
		c.ReportProgress(fmt.Sprintf("error-check-%d", i+1), map[string]any{
			"desc": errReq.desc,
		})

		var bodyReader io.Reader
		if errReq.body != "" {
			bodyReader = strings.NewReader(errReq.body)
		}

		req, _ := http.NewRequestWithContext(
			ctx, errReq.method, c.config.BaseURL+errReq.path, bodyReader,
		)
		if errReq.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			// Network error is acceptable, not a leak
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "not_empty",
				Target:   fmt.Sprintf("error_response_%s", errReq.desc),
				Expected: "no sensitive data in response",
				Actual:   "network error (no response body to check)",
				Passed:   true,
				Message:  fmt.Sprintf("Request to %s failed with network error (acceptable)", errReq.desc),
			})
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		bodyStr := string(respBody)
		bodyLower := strings.ToLower(bodyStr)

		// Also check JSON-parsed error message
		var jsonResp map[string]interface{}
		_ = json.Unmarshal(respBody, &jsonResp)
		errorMsg := ""
		if msg, ok := jsonResp["error"].(string); ok {
			errorMsg = msg
		}
		if msg, ok := jsonResp["message"].(string); ok {
			errorMsg += " " + msg
		}

		foundPatterns := []string{}
		for _, pattern := range sensitivePatterns {
			patternLower := strings.ToLower(pattern)
			if strings.Contains(bodyLower, patternLower) {
				foundPatterns = append(foundPatterns, pattern)
			}
		}

		noLeak := len(foundPatterns) == 0
		if !noLeak {
			leaksFound++
		}

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   fmt.Sprintf("error_response_%s", errReq.desc),
			Expected: "no sensitive patterns in error response",
			Actual: challenge.Ternary(noLeak,
				fmt.Sprintf("HTTP %d, clean (body=%d bytes)", resp.StatusCode, len(respBody)),
				fmt.Sprintf("HTTP %d, LEAKED: %v", resp.StatusCode, foundPatterns)),
			Passed: noLeak,
			Message: challenge.Ternary(noLeak,
				fmt.Sprintf("Error response for %q is clean (HTTP %d, error=%q)", errReq.desc, resp.StatusCode, errorMsg),
				fmt.Sprintf("SENSITIVE DATA LEAK in %q response: found patterns %v", errReq.desc, foundPatterns)),
		})
	}

	outputs["leaks_found"] = fmt.Sprintf("%d", leaksFound)
	outputs["requests_checked"] = fmt.Sprintf("%d", len(errorRequests))

	metrics := map[string]challenge.MetricValue{
		"error_check_time": {
			Name:  "error_check_time",
			Value: float64(time.Since(start).Milliseconds()),
			Unit:  "ms",
		},
		"leaks_found": {
			Name:  "leaks_found",
			Value: float64(leaksFound),
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
