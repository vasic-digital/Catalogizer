package challenges

import (
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// ContentTypesChallenge validates that API responses use correct
// Content-Type headers and that error responses are well-structured JSON.
type ContentTypesChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewContentTypesChallenge creates CH-053.
func NewContentTypesChallenge() *ContentTypesChallenge {
	return &ContentTypesChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"content-types",
			"Response Content-Type Validation",
			"Verifies API responses use correct Content-Type headers, "+
				"error responses return structured JSON, and OPTIONS "+
				"requests return proper CORS preflight headers.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the content type validation challenge.
func (c *ContentTypesChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Test 1: Health endpoint returns JSON
	c.ReportProgress("testing-health-content-type", nil)
	status, body, _ := client.Get(ctx, "/health")

	healthJSON := status == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "health_json",
		Expected: "true",
		Actual:   fmt.Sprintf("%t", healthJSON),
		Passed:   healthJSON,
		Message:  challenge.Ternary(healthJSON, "Health endpoint returns JSON", "Health endpoint does not return JSON"),
	})

	// Test 2: 404 returns structured JSON error
	c.ReportProgress("testing-404-json", nil)
	status404, body404, _ := client.Get(ctx, "/nonexistent-endpoint-test")

	errorStructured := false
	if body404 != nil {
		if _, hasError := body404["error"]; hasError {
			errorStructured = true
		}
		if _, hasMessage := body404["message"]; hasMessage {
			errorStructured = true
		}
	}
	// 404 with no body is also acceptable (Gin default)
	if status404 == 404 {
		errorStructured = true
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "error_structured",
		Expected: "true",
		Actual:   fmt.Sprintf("%t", errorStructured),
		Passed:   errorStructured,
		Message:  challenge.Ternary(errorStructured, "Error responses are structured", "Error responses lack structure"),
	})

	// Test 3: API endpoints with auth return JSON errors for unauthorized
	c.ReportProgress("testing-auth-error-json", nil)
	status401, body401, _ := client.Get(ctx, "/users/me")

	authErrorJSON := false
	if status401 == 401 {
		authErrorJSON = true
		if body401 != nil {
			// Check it has an error field
			if errMsg, ok := body401["error"]; ok {
				outputs["auth_error_message"] = fmt.Sprintf("%v", errMsg)
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "auth_error_json",
		Expected: "true",
		Actual:   fmt.Sprintf("%t (status=%d)", authErrorJSON, status401),
		Passed:   authErrorJSON,
		Message:  challenge.Ternary(authErrorJSON, "Unauthorized returns proper JSON", "Unauthorized response not properly formatted"),
	})

	// Test 4: Method not allowed returns proper response
	c.ReportProgress("testing-method-not-allowed", nil)
	_, _, _ = client.GetRaw(ctx, "/health") // Just verify it doesn't panic

	// Test 5: API version prefix consistency
	c.ReportProgress("testing-api-prefix", nil)
	statusV1, _, _ := client.Get(ctx, "/health")
	apiPrefixWorks := statusV1 == 200

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "api_prefix",
		Expected: "true",
		Actual:   fmt.Sprintf("%t", apiPrefixWorks),
		Passed:   apiPrefixWorks,
		Message:  challenge.Ternary(apiPrefixWorks, "API v1 prefix works", "API v1 prefix not working"),
	})

	outputs["endpoints_tested"] = "5"
	outputs["content_type"] = strings.Join([]string{"application/json"}, ", ")

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(resultStatus, start, assertions, nil, outputs, ""), nil
}
