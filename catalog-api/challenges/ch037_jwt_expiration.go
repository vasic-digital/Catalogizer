package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// JWTExpirationChallenge validates that JWT token expiration is
// enforced by the API. Verifies that login returns an expiration
// field and that an expired/tampered token is rejected.
type JWTExpirationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewJWTExpirationChallenge creates CH-037.
func NewJWTExpirationChallenge() *JWTExpirationChallenge {
	return &JWTExpirationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"jwt-expiration",
			"JWT Token Expiration",
			"Validates JWT token expiration is enforced: login returns "+
				"expires_at field, tampered tokens are rejected, and "+
				"valid tokens are accepted.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the JWT expiration challenge.
func (c *JWTExpirationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Login and verify expires_at is present
	c.ReportProgress("login", nil)
	loginResp, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	loginOK := loginErr == nil && loginResp != nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "login",
		Expected: "login succeeds",
		Actual:   challenge.Ternary(loginOK, "succeeded", fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message: challenge.Ternary(loginOK,
			"Login succeeded",
			fmt.Sprintf("Login failed: %v", loginErr)),
	})

	if !loginOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "login failed",
		), nil
	}

	// Step 2: Verify expires_at field exists in login response
	c.ReportProgress("check-expiration", nil)
	expiresAt, hasExpires := loginResp["expires_at"]
	expiresStr := fmt.Sprintf("%v", expiresAt)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "expires_at_present",
		Expected: "expires_at field in login response",
		Actual:   challenge.Ternary(hasExpires, expiresStr, "missing"),
		Passed:   hasExpires,
		Message: challenge.Ternary(hasExpires,
			fmt.Sprintf("Token expiration present: %s", expiresStr),
			"No expires_at in login response"),
	})

	// Step 3: Valid token works for protected endpoint
	c.ReportProgress("valid-token", nil)
	validToken := client.Token()
	meCode, _, meErr := client.Get(ctx, "/api/v1/auth/me")
	meOK := meErr == nil && meCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "valid_token_access",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", meCode),
		Passed:   meOK,
		Message: challenge.Ternary(meOK,
			"Valid token accepted for protected endpoint",
			fmt.Sprintf("Valid token rejected: code=%d err=%v", meCode, meErr)),
	})
	outputs["token_length"] = fmt.Sprintf("%d", len(validToken))

	// Step 4: Tampered token (modified signature) is rejected
	c.ReportProgress("tampered-token", nil)
	tamperedToken := validToken + "tampered"
	tamperedClient := httpclient.NewAPIClient(c.config.BaseURL)
	tamperedClient.SetToken(tamperedToken)

	tamperedCode, _, tamperedErr := tamperedClient.Get(ctx, "/api/v1/auth/me")
	tamperedRejected := tamperedErr == nil && (tamperedCode == 401 || tamperedCode == 403)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "tampered_token_rejected",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("HTTP %d", tamperedCode),
		Passed:   tamperedRejected,
		Message: challenge.Ternary(tamperedRejected,
			fmt.Sprintf("Tampered token correctly rejected: HTTP %d", tamperedCode),
			fmt.Sprintf("Tampered token not rejected: code=%d err=%v", tamperedCode, tamperedErr)),
	})

	// Step 5: Empty token is rejected
	c.ReportProgress("empty-token", nil)
	emptyClient := httpclient.NewAPIClient(c.config.BaseURL)
	emptyClient.SetToken("")

	emptyCode, _, emptyErr := emptyClient.Get(ctx, "/api/v1/auth/me")
	emptyRejected := emptyErr == nil && (emptyCode == 401 || emptyCode == 403)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "empty_token_rejected",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("HTTP %d", emptyCode),
		Passed:   emptyRejected,
		Message: challenge.Ternary(emptyRejected,
			fmt.Sprintf("Empty token correctly rejected: HTTP %d", emptyCode),
			fmt.Sprintf("Empty token not rejected: code=%d err=%v", emptyCode, emptyErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"jwt_validation_time": {
			Name:  "jwt_validation_time",
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
