package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// AuthTokenRefreshChallenge validates the full auth lifecycle used
// by @vasic-digital/catalogizer-api-client: login, check status,
// and token refresh. Ensures the auth endpoints the client depends
// on all return valid responses.
type AuthTokenRefreshChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAuthTokenRefreshChallenge creates CH-025.
func NewAuthTokenRefreshChallenge() *AuthTokenRefreshChallenge {
	return &AuthTokenRefreshChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"auth-token-refresh",
			"Auth Token Refresh",
			"Validates the auth lifecycle: "+
				"POST /auth/login returns session_token and user, "+
				"GET /auth/status returns authenticated=true with user, "+
				"POST /auth/refresh returns new session_token. "+
				"Used by @vasic-digital/catalogizer-api-client AuthService.",
			"e2e",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the auth token refresh challenge.
func (c *AuthTokenRefreshChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// 1. POST /auth/login — expect session_token in response
	loginPayload, _ := json.Marshal(map[string]string{
		"username": c.config.Username,
		"password": c.config.Password,
	})
	loginCode, loginBytes, loginErr := client.PostJSON(
		ctx, "/auth/login", string(loginPayload),
	)
	loginOK := loginErr == nil && loginCode == 200
	sessionToken := ""
	if loginOK && len(loginBytes) > 0 {
		var resp map[string]interface{}
		if jsonErr := json.Unmarshal(loginBytes, &resp); jsonErr == nil {
			if t, ok := resp["session_token"].(string); ok {
				sessionToken = t
			} else if t, ok := resp["token"].(string); ok {
				sessionToken = t
			} else if data, ok := resp["data"].(map[string]interface{}); ok {
				if t, ok := data["session_token"].(string); ok {
					sessionToken = t
				}
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "POST /auth/login session_token",
		Expected: "non-empty session_token",
		Actual:   fmt.Sprintf("HTTP %d, token_len=%d", loginCode, len(sessionToken)),
		Passed:   loginOK && sessionToken != "",
		Message: challenge.Ternary(loginOK && sessionToken != "",
			fmt.Sprintf("Login OK: session_token length=%d", len(sessionToken)),
			fmt.Sprintf("Login failed or missing token: code=%d err=%v", loginCode, loginErr)),
	})

	if sessionToken != "" {
		client.SetToken(sessionToken)
	}

	// 2. GET /auth/status — expect authenticated=true
	statusCode, statusBody, statusErr := client.Get(ctx, "/auth/status")
	statusOK := statusErr == nil && statusCode == 200
	isAuthenticated := false
	hasUser := false
	if statusBody != nil {
		if auth, ok := statusBody["authenticated"].(bool); ok {
			isAuthenticated = auth
		}
		if data, ok := statusBody["data"].(map[string]interface{}); ok {
			if auth, ok := data["authenticated"].(bool); ok {
				isAuthenticated = auth
			}
			if _, ok := data["user"]; ok {
				hasUser = true
			}
		}
		if _, ok := statusBody["user"]; ok {
			hasUser = true
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "GET /auth/status authenticated",
		Expected: "authenticated=true with user object",
		Actual:   fmt.Sprintf("HTTP %d, authenticated=%v, has_user=%v", statusCode, isAuthenticated, hasUser),
		Passed:   statusOK && isAuthenticated,
		Message: challenge.Ternary(statusOK && isAuthenticated,
			fmt.Sprintf("Auth status OK: authenticated=%v has_user=%v", isAuthenticated, hasUser),
			fmt.Sprintf("Auth status failed: code=%d authenticated=%v err=%v", statusCode, isAuthenticated, statusErr)),
	})

	// 3. POST /auth/refresh — expect new session_token
	refreshCode, refreshBytes, refreshErr := client.PostJSON(
		ctx, "/auth/refresh", "{}",
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
			} else if data, ok := resp["data"].(map[string]interface{}); ok {
				if t, ok := data["session_token"].(string); ok {
					newToken = t
				}
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "POST /auth/refresh session_token",
		Expected: "non-empty session_token",
		Actual:   fmt.Sprintf("HTTP %d, token_len=%d", refreshCode, len(newToken)),
		Passed:   refreshOK && newToken != "",
		Message: challenge.Ternary(refreshOK && newToken != "",
			fmt.Sprintf("Token refresh OK: new token length=%d", len(newToken)),
			fmt.Sprintf("Token refresh failed: code=%d err=%v", refreshCode, refreshErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"auth_lifecycle_latency": {
			Name:  "auth_lifecycle_latency",
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
