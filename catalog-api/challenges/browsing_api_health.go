package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// allFirstCatalogDeps lists all 7 first-catalog challenge IDs
// that must complete before browsing challenges can run.
var allFirstCatalogDeps = []challenge.ID{
	"first-catalog-smb-connect",
	"first-catalog-dir-discovery",
	"first-catalog-music-scan",
	"first-catalog-series-scan",
	"first-catalog-movies-scan",
	"first-catalog-software-scan",
	"first-catalog-comics-scan",
}

// BrowsingAPIHealthChallenge validates that the API is running,
// healthy, and that admin credentials produce a valid JWT session.
type BrowsingAPIHealthChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowsingAPIHealthChallenge creates CH-008.
func NewBrowsingAPIHealthChallenge() *BrowsingAPIHealthChallenge {
	return &BrowsingAPIHealthChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browsing-api-health",
			"API Health & Auth",
			"Validates API health endpoint responds, admin login succeeds "+
				"with JWT token, and /auth/me returns user data",
			"e2e",
			allFirstCatalogDeps,
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the API health and auth challenge.
func (c *BrowsingAPIHealthChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url":  c.config.BaseURL,
		"username": c.config.Username,
		"password": c.config.Password,
	}

	client := NewAPIClient(c.config.BaseURL)

	// Step 1: GET /health returns 200 with status=healthy
	healthCode, healthBody, healthErr := client.Get(ctx, "/health")
	healthOK := healthErr == nil && healthCode == 200
	healthStatus := ""
	if healthBody != nil {
		if s, ok := healthBody["status"].(string); ok {
			healthStatus = s
		}
	}
	healthPassed := healthOK && healthStatus == "healthy"

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "health_endpoint",
		Expected: "HTTP 200 with status=healthy",
		Actual:   fmt.Sprintf("HTTP %d, status=%q", healthCode, healthStatus),
		Passed:   healthPassed,
		Message:  ternary(healthPassed, "Health endpoint returned healthy", fmt.Sprintf("Health check failed: code=%d status=%q err=%v", healthCode, healthStatus, healthErr)),
	})
	if !healthPassed {
		errMsg := fmt.Sprintf("health check failed: HTTP %d", healthCode)
		if healthErr != nil {
			errMsg = healthErr.Error()
		}
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, errMsg), nil
	}

	// Step 2: POST /api/v1/auth/login succeeds
	loginResp, loginErr := client.Login(ctx, c.config.Username, c.config.Password)
	loginOK := loginErr == nil && loginResp != nil

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "login_endpoint",
		Expected: "successful login response",
		Actual:   ternary(loginOK, "login succeeded", fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message:  ternary(loginOK, "Admin login succeeded", fmt.Sprintf("Login failed: %v", loginErr)),
	})
	if !loginOK {
		errMsg := "login failed"
		if loginErr != nil {
			errMsg = loginErr.Error()
		}
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, errMsg), nil
	}

	// Step 3: Response contains non-empty token
	token := client.Token()
	tokenOK := token != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "session_token",
		Expected: "non-empty JWT token",
		Actual:   ternary(tokenOK, fmt.Sprintf("token length=%d", len(token)), "empty"),
		Passed:   tokenOK,
		Message:  ternary(tokenOK, "JWT token received", "No token in login response"),
	})

	// Step 4: Response contains refresh_token
	_, hasRefresh := loginResp["refresh_token"]
	refreshOK := hasRefresh
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "refresh_token",
		Expected: "refresh_token present",
		Actual:   ternary(refreshOK, "present", "missing"),
		Passed:   refreshOK,
		Message:  ternary(refreshOK, "Refresh token present in response", "No refresh_token in login response"),
	})

	// Step 5: Response contains expires_in
	_, hasExpires := loginResp["expires_in"]
	expiresOK := hasExpires
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "expires_in",
		Expected: "expires_in present",
		Actual:   ternary(expiresOK, "present", "missing"),
		Passed:   expiresOK,
		Message:  ternary(expiresOK, "Expiration info present in response", "No expires_in in login response"),
	})

	// Step 6: Response user.username matches config
	usernameOK := false
	if user, ok := loginResp["user"].(map[string]interface{}); ok {
		if u, ok := user["username"].(string); ok {
			usernameOK = u == c.config.Username
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "user_username",
		Expected: c.config.Username,
		Actual:   ternary(usernameOK, c.config.Username, "mismatch or missing"),
		Passed:   usernameOK,
		Message:  ternary(usernameOK, fmt.Sprintf("Username matches: %s", c.config.Username), "Username mismatch in login response"),
	})

	// Step 7: GET /api/v1/auth/me returns 200
	meCode, meBody, meErr := client.Get(ctx, "/api/v1/auth/me")
	meOK := meErr == nil && meCode == 200 && meBody != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "auth_me_endpoint",
		Expected: "HTTP 200 with user data",
		Actual:   fmt.Sprintf("HTTP %d", meCode),
		Passed:   meOK,
		Message:  ternary(meOK, "Auth /me endpoint returned user data", fmt.Sprintf("Auth /me failed: code=%d err=%v", meCode, meErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"auth_time": {
			Name:  "auth_time",
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
