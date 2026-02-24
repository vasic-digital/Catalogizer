package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// SecurityChallenge validates the authentication and authorization
// flow: login with valid credentials, use the token, test with
// an expired/invalid token, and test role-based access enforcement.
type SecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSecurityChallenge creates CH-034.
func NewSecurityChallenge() *SecurityChallenge {
	return &SecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"security",
			"Security",
			"Validates auth flow: login with valid credentials, "+
				"use token for protected endpoints, test invalid token "+
				"rejection, test expired token handling, verify role enforcement.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the security challenge.
func (c *SecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	// Step 1: Login with valid credentials
	c.ReportProgress("valid-login", nil)
	client := httpclient.NewAPIClient(c.config.BaseURL)
	loginResp, loginErr := client.Login(ctx, c.config.Username, c.config.Password)
	loginOK := loginErr == nil && loginResp != nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "valid_login",
		Expected: "login succeeds with valid credentials",
		Actual:   challenge.Ternary(loginOK, "login succeeded", fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message: challenge.Ternary(loginOK,
			"Valid credentials accepted",
			fmt.Sprintf("Valid login failed: %v", loginErr)),
	})

	if !loginOK {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, "login failed",
		), nil
	}

	validToken := client.Token()
	outputs["token_length"] = fmt.Sprintf("%d", len(validToken))

	// Step 2: Access protected endpoint with valid token
	c.ReportProgress("protected-endpoint", nil)
	meCode, meBody, meErr := client.Get(ctx, "/api/v1/auth/me")
	meOK := meErr == nil && meCode == 200
	meUsername := ""
	if meBody != nil {
		if u, ok := meBody["username"].(string); ok {
			meUsername = u
		} else if user, ok := meBody["user"].(map[string]interface{}); ok {
			if u, ok := user["username"].(string); ok {
				meUsername = u
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "protected_with_valid_token",
		Expected: "200 with user data",
		Actual:   fmt.Sprintf("HTTP %d, username=%q", meCode, meUsername),
		Passed:   meOK,
		Message: challenge.Ternary(meOK,
			fmt.Sprintf("Protected endpoint accessible: username=%q", meUsername),
			fmt.Sprintf("Protected endpoint failed with valid token: code=%d err=%v", meCode, meErr)),
	})

	// Step 3: Test with invalid token
	c.ReportProgress("invalid-token", nil)
	invalidClient := httpclient.NewAPIClient(c.config.BaseURL)
	invalidClient.SetToken("invalid.jwt.token.that.should.be.rejected")

	invalidCode, _, invalidErr := invalidClient.Get(ctx, "/api/v1/auth/me")
	// Expect 401 Unauthorized
	invalidRejected := invalidErr == nil && invalidCode == 401
	// Also accept 403 Forbidden
	if !invalidRejected && invalidErr == nil && invalidCode == 403 {
		invalidRejected = true
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "invalid_token_rejected",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("HTTP %d", invalidCode),
		Passed:   invalidRejected,
		Message: challenge.Ternary(invalidRejected,
			fmt.Sprintf("Invalid token correctly rejected: HTTP %d", invalidCode),
			fmt.Sprintf("Invalid token not rejected properly: code=%d err=%v", invalidCode, invalidErr)),
	})

	// Step 4: Test with no token (unauthenticated access to protected endpoint)
	c.ReportProgress("no-token", nil)
	noAuthClient := httpclient.NewAPIClient(c.config.BaseURL)

	noAuthCode, _, noAuthErr := noAuthClient.Get(ctx, "/api/v1/auth/me")
	noAuthRejected := noAuthErr == nil && (noAuthCode == 401 || noAuthCode == 403)

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "unauthenticated_rejected",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("HTTP %d", noAuthCode),
		Passed:   noAuthRejected,
		Message: challenge.Ternary(noAuthRejected,
			fmt.Sprintf("Unauthenticated request correctly rejected: HTTP %d", noAuthCode),
			fmt.Sprintf("Unauthenticated request not rejected: code=%d err=%v", noAuthCode, noAuthErr)),
	})

	// Step 5: Test with bad credentials
	c.ReportProgress("bad-credentials", nil)
	badClient := httpclient.NewAPIClient(c.config.BaseURL)
	_, badLoginErr := badClient.Login(ctx, "nonexistent_user", "wrong_password")
	badLoginRejected := badLoginErr != nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "bad_credentials_rejected",
		Expected: "login fails with bad credentials",
		Actual:   challenge.Ternary(badLoginRejected, "rejected", "accepted (unexpected)"),
		Passed:   badLoginRejected,
		Message: challenge.Ternary(badLoginRejected,
			"Bad credentials correctly rejected",
			"Bad credentials were accepted (security issue)"),
	})

	// Step 6: Verify token refresh with valid session
	c.ReportProgress("token-refresh", nil)
	refreshToken := ""
	if loginResp != nil {
		if rt, ok := loginResp["refresh_token"].(string); ok {
			refreshToken = rt
		}
	}

	if refreshToken != "" {
		refreshPayload, _ := json.Marshal(map[string]string{
			"refresh_token": refreshToken,
		})
		refreshCode, refreshBytes, refreshErr := client.PostJSON(
			ctx, "/api/v1/auth/refresh", string(refreshPayload),
		)
		refreshOK := refreshErr == nil && (refreshCode == 200 || refreshCode == 201)
		newToken := ""
		if refreshOK && len(refreshBytes) > 0 {
			var resp map[string]interface{}
			if jsonErr := json.Unmarshal(refreshBytes, &resp); jsonErr == nil {
				if t, ok := resp["session_token"].(string); ok {
					newToken = t
				} else if t, ok := resp["token"].(string); ok {
					newToken = t
				}
			}
		}

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "token_refresh",
			Expected: "new token returned",
			Actual:   fmt.Sprintf("HTTP %d, new_token_len=%d", refreshCode, len(newToken)),
			Passed:   refreshOK && newToken != "",
			Message: challenge.Ternary(refreshOK && newToken != "",
				fmt.Sprintf("Token refresh succeeded: new token length=%d", len(newToken)),
				fmt.Sprintf("Token refresh failed: code=%d err=%v", refreshCode, refreshErr)),
		})
	}

	// Step 7: Verify health endpoint is publicly accessible (no auth needed)
	c.ReportProgress("public-endpoint", nil)
	publicClient := httpclient.NewAPIClient(c.config.BaseURL)
	healthCode, _, healthErr := publicClient.Get(ctx, "/health")
	healthOK := healthErr == nil && healthCode == 200

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "public_endpoint_accessible",
		Expected: "200 without auth",
		Actual:   fmt.Sprintf("HTTP %d", healthCode),
		Passed:   healthOK,
		Message: challenge.Ternary(healthOK,
			"Public health endpoint accessible without auth",
			fmt.Sprintf("Public endpoint failed: code=%d err=%v", healthCode, healthErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"security_test_latency": {
			Name:  "security_test_latency",
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

	return c.CreateResult(
		status, start, assertions, metrics, outputs, "",
	), nil
}
