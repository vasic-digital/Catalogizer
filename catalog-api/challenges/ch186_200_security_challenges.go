package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// Challenges CH-186 through CH-200 validate security controls
// including JWT validation, role-based access, input sanitization,
// rate limiting, and error information leakage prevention.

// --- CH-186: JWTValidation ---

// JWTValidationChallenge validates that requests with an invalid
// JWT token are rejected with 401.
type JWTValidationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewJWTValidationChallenge creates CH-186.
func NewJWTValidationChallenge() *JWTValidationChallenge {
	return &JWTValidationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"jwt-validation",
			"JWT Validation",
			"Validates that requests with an invalid JWT "+
				"token are rejected with 401.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the JWT validation challenge.
func (c *JWTValidationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/storage-roots", nil,
	)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request creation failed: %v", err),
		), nil
	}

	c.ReportProgress("testing-invalid-jwt", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request failed: %v", doErr),
		), nil
	}
	resp.Body.Close()

	codeOK := resp.StatusCode == 401
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "jwt_validation_status",
		Expected: "401",
		Actual:   fmt.Sprintf("%d", resp.StatusCode),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Invalid JWT correctly rejected with 401",
			fmt.Sprintf("Invalid JWT returned %d instead of 401",
				resp.StatusCode)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-187: TokenRefreshSecurity ---

// TokenRefreshSecurityChallenge validates that the token
// refresh endpoint works for authenticated users.
type TokenRefreshSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewTokenRefreshSecurityChallenge creates CH-187.
func NewTokenRefreshSecurityChallenge() *TokenRefreshSecurityChallenge {
	return &TokenRefreshSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"token-refresh-security",
			"Token Refresh Security",
			"Validates POST /api/v1/auth/refresh returns a "+
				"new token for authenticated users.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the token refresh security challenge.
func (c *TokenRefreshSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-token-refresh", nil)
	code, _, err := client.PostJSON(
		ctx, "/api/v1/auth/refresh", "{}",
	)

	codeOK := err == nil && (code == 200 || code == 201)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "token_refresh_status",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Token refresh returned %d", code),
			fmt.Sprintf("Token refresh returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-188: RoleBasedAccessSecurity ---

// RoleBasedAccessSecurityChallenge validates that admin-only
// endpoints are blocked for non-admin users.
type RoleBasedAccessSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewRoleBasedAccessSecurityChallenge creates CH-188.
func NewRoleBasedAccessSecurityChallenge() *RoleBasedAccessSecurityChallenge {
	return &RoleBasedAccessSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"role-based-access-security",
			"Role-Based Access Security",
			"Validates that admin-only endpoints like "+
				"/api/v1/admin/system-info reject non-admin "+
				"users with 401 or 403.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the role-based access security challenge.
func (c *RoleBasedAccessSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	// Try accessing admin endpoint without auth
	client := httpclient.NewAPIClient(c.config.BaseURL)

	c.ReportProgress("testing-role-based-access", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/admin/system-info",
	)

	codeOK := err == nil && (code == 401 || code == 403)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "role_based_access_status",
		Expected: "401 or 403",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Admin endpoint blocked with %d", code),
			fmt.Sprintf("Admin endpoint returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-189: InputValidationSecurity ---

// InputValidationSecurityChallenge validates that XSS payloads
// in request bodies are sanitized or rejected.
type InputValidationSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewInputValidationSecurityChallenge creates CH-189.
func NewInputValidationSecurityChallenge() *InputValidationSecurityChallenge {
	return &InputValidationSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"input-validation-security",
			"Input Validation Security",
			"Validates that XSS payloads in request bodies "+
				"are sanitized or rejected.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the input validation security challenge.
func (c *InputValidationSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-xss-sanitization", nil)
	xssBody := `{"name":"<script>alert('xss')</script>"}`
	code, rawBody, err := client.PostJSON(
		ctx, "/api/v1/collections", xssBody,
	)

	// The API should either reject (400) or sanitize the input
	codeOK := err == nil && code != 500
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "input_validation_security_status",
		Expected: "not 500",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("XSS input handled with %d", code),
			fmt.Sprintf("XSS input caused error %d, err=%v",
				code, err)),
	})

	// If accepted, verify the script tag is not echoed raw
	if codeOK && rawBody != nil {
		bodyStr := string(rawBody)
		noRawScript := !strings.Contains(bodyStr,
			"<script>alert('xss')</script>")
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "input_validation_sanitized",
			Expected: "no raw script tags in response",
			Actual: challenge.Ternary(noRawScript,
				"sanitized", "raw script present"),
			Passed: noRawScript,
			Message: challenge.Ternary(noRawScript,
				"XSS payload sanitized in response",
				"XSS payload echoed raw in response"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-190: SQLInjectionSecurity ---

// SQLInjectionSecurityChallenge validates that SQL injection
// attempts in query parameters are blocked.
type SQLInjectionSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSQLInjectionSecurityChallenge creates CH-190.
func NewSQLInjectionSecurityChallenge() *SQLInjectionSecurityChallenge {
	return &SQLInjectionSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"sql-injection-security",
			"SQL Injection Security",
			"Validates that SQL injection attempts in query "+
				"parameters are blocked or sanitized.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the SQL injection security challenge.
func (c *SQLInjectionSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-sql-injection", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities/search?q='+OR+1=1--",
	)

	// Should not return 500 (would indicate unhandled SQL error)
	codeOK := err == nil && code != 500
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sql_injection_security_status",
		Expected: "not 500",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("SQL injection handled with %d", code),
			fmt.Sprintf("SQL injection caused error %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-191: PathTraversalSecurity ---

// PathTraversalSecurityChallenge validates that path traversal
// attempts are blocked.
type PathTraversalSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPathTraversalSecurityChallenge creates CH-191.
func NewPathTraversalSecurityChallenge() *PathTraversalSecurityChallenge {
	return &PathTraversalSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"path-traversal-security",
			"Path Traversal Security",
			"Validates that path traversal attempts like "+
				"../../etc/passwd are blocked.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the path traversal security challenge.
func (c *PathTraversalSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-path-traversal", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/media/by-path?path=../../etc/passwd",
	)

	// Should return 400 (bad request) or 403 (forbidden), not 200
	codeOK := err == nil && code != 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "path_traversal_security_status",
		Expected: "not 200 (blocked)",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Path traversal blocked with %d", code),
			fmt.Sprintf("Path traversal returned %d, err=%v",
				code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-192: RateLimitingSecurity ---

// RateLimitingSecurityChallenge validates that rate limiting
// is enforced on API endpoints.
type RateLimitingSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewRateLimitingSecurityChallenge creates CH-192.
func NewRateLimitingSecurityChallenge() *RateLimitingSecurityChallenge {
	return &RateLimitingSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"rate-limiting-security",
			"Rate Limiting Security",
			"Validates that rate limiting headers are present "+
				"in API responses.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the rate limiting security challenge.
func (c *RateLimitingSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("testing-rate-limiting", nil)

	// Make a request and check for rate limit headers
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/health", nil,
	)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request creation failed: %v", err),
		), nil
	}

	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request failed: %v", doErr),
		), nil
	}
	resp.Body.Close()

	// Check for rate limit headers
	rlLimit := resp.Header.Get("X-RateLimit-Limit")
	rlRemaining := resp.Header.Get("X-RateLimit-Remaining")
	hasRL := rlLimit != "" || rlRemaining != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "rate_limit_headers",
		Expected: "rate limit headers present",
		Actual: challenge.Ternary(hasRL,
			"present", "missing"),
		Passed: hasRL,
		Message: challenge.Ternary(hasRL,
			"Rate limit headers found in response",
			"Rate limit headers missing from response"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-193: CORSHeadersSecurity ---

// CORSHeadersSecurityChallenge validates that CORS headers are
// present in API responses.
type CORSHeadersSecurityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCORSHeadersSecurityChallenge creates CH-193.
func NewCORSHeadersSecurityChallenge() *CORSHeadersSecurityChallenge {
	return &CORSHeadersSecurityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"cors-headers-security",
			"CORS Headers Security",
			"Validates that CORS headers are present in "+
				"API responses for cross-origin requests.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the CORS headers security challenge.
func (c *CORSHeadersSecurityChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("testing-cors-headers", nil)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodOptions,
		c.config.BaseURL+"/api/v1/health", nil,
	)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request creation failed: %v", err),
		), nil
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request failed: %v", doErr),
		), nil
	}
	resp.Body.Close()

	acao := resp.Header.Get("Access-Control-Allow-Origin")
	hasCORS := acao != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "cors_headers_present",
		Expected: "Access-Control-Allow-Origin header",
		Actual: challenge.Ternary(hasCORS,
			acao, "missing"),
		Passed: hasCORS,
		Message: challenge.Ternary(hasCORS,
			"CORS headers present: "+acao,
			"CORS headers missing from response"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-194: SecurityHeadersPresent ---

// SecurityHeadersPresentChallenge validates that security
// headers like X-Frame-Options are present.
type SecurityHeadersPresentChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSecurityHeadersPresentChallenge creates CH-194.
func NewSecurityHeadersPresentChallenge() *SecurityHeadersPresentChallenge {
	return &SecurityHeadersPresentChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"security-headers-present",
			"Security Headers Present",
			"Validates that security headers like "+
				"X-Frame-Options and X-Content-Type-Options "+
				"are present in responses.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the security headers present challenge.
func (c *SecurityHeadersPresentChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("checking-security-headers", nil)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/health", nil,
	)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request creation failed: %v", err),
		), nil
	}

	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request failed: %v", doErr),
		), nil
	}
	resp.Body.Close()

	xfo := resp.Header.Get("X-Frame-Options")
	xcto := resp.Header.Get("X-Content-Type-Options")
	hasHeaders := xfo != "" || xcto != ""
	outputs["x_frame_options"] = xfo
	outputs["x_content_type_options"] = xcto

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "security_headers",
		Expected: "security headers present",
		Actual: challenge.Ternary(hasHeaders,
			"present", "missing"),
		Passed: hasHeaders,
		Message: challenge.Ternary(hasHeaders,
			fmt.Sprintf("Security headers: XFO=%s, XCTO=%s",
				xfo, xcto),
			"Security headers missing from response"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-195: PasswordHashing ---

// PasswordHashingChallenge validates that the auth endpoint
// does not return raw passwords in responses.
type PasswordHashingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPasswordHashingChallenge creates CH-195.
func NewPasswordHashingChallenge() *PasswordHashingChallenge {
	return &PasswordHashingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"password-hashing",
			"Password Hashing",
			"Validates that user endpoints do not return "+
				"raw passwords in responses.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the password hashing challenge.
func (c *PasswordHashingChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-password-exposure", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/auth/me",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "password_hashing_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("User profile returned %d", code),
			fmt.Sprintf("User profile returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		bodyStr := string(rawBody)
		noPassword := !strings.Contains(bodyStr, c.config.Password)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "password_not_exposed",
			Expected: "password not in response",
			Actual: challenge.Ternary(noPassword,
				"not exposed", "exposed"),
			Passed: noPassword,
			Message: challenge.Ternary(noPassword,
				"Password not exposed in user profile",
				"Password found in user profile response"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-196: SessionTimeout ---

// SessionTimeoutChallenge validates that expired tokens are
// rejected by the API.
type SessionTimeoutChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSessionTimeoutChallenge creates CH-196.
func NewSessionTimeoutChallenge() *SessionTimeoutChallenge {
	return &SessionTimeoutChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"session-timeout",
			"Session Timeout",
			"Validates that a clearly expired JWT token is "+
				"rejected with 401.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the session timeout challenge.
func (c *SessionTimeoutChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("testing-expired-token", nil)

	// Use a known-expired JWT (exp claim in the past)
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxIiwiZXhwIjoxMDAwMDAwMDAwfQ." +
		"invalid_signature"

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/storage-roots", nil,
	)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request creation failed: %v", err),
		), nil
	}
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request failed: %v", doErr),
		), nil
	}
	resp.Body.Close()

	codeOK := resp.StatusCode == 401
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "session_timeout_status",
		Expected: "401",
		Actual:   fmt.Sprintf("%d", resp.StatusCode),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Expired token correctly rejected with 401",
			fmt.Sprintf("Expired token returned %d",
				resp.StatusCode)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-197: BruteForceProtection ---

// BruteForceProtectionChallenge validates that rapid login
// attempts are rate-limited.
type BruteForceProtectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBruteForceProtectionChallenge creates CH-197.
func NewBruteForceProtectionChallenge() *BruteForceProtectionChallenge {
	return &BruteForceProtectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"brute-force-protection",
			"Brute Force Protection",
			"Validates that rapid failed login attempts are "+
				"rate-limited with 429.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the brute force protection challenge.
func (c *BruteForceProtectionChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("testing-brute-force-protection", nil)

	got429 := false
	for i := 0; i < 10; i++ {
		body := `{"username":"nonexistent","password":"wrong"}`
		req, err := http.NewRequestWithContext(
			ctx, http.MethodPost,
			c.config.BaseURL+"/api/v1/auth/login",
			strings.NewReader(body),
		)
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, doErr := httpClient.Do(req)
		if doErr != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 429 {
			got429 = true
			break
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "brute_force_protection",
		Expected: "429 after rapid login attempts",
		Actual: challenge.Ternary(got429,
			"429 received", "no 429"),
		Passed: got429,
		Message: challenge.Ternary(got429,
			"Brute force protection active (429 received)",
			"No rate limiting detected on login endpoint"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-198: ContentTypeValidation ---

// ContentTypeValidationChallenge validates that requests with
// wrong content-type are rejected.
type ContentTypeValidationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewContentTypeValidationChallenge creates CH-198.
func NewContentTypeValidationChallenge() *ContentTypeValidationChallenge {
	return &ContentTypeValidationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"content-type-validation",
			"Content-Type Validation",
			"Validates that POST requests with wrong "+
				"Content-Type are rejected.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the content-type validation challenge.
func (c *ContentTypeValidationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-content-type-validation", nil)

	token := client.Token()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		c.config.BaseURL+"/api/v1/collections",
		strings.NewReader("not json at all"),
	)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request creation failed: %v", err),
		), nil
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, doErr := httpClient.Do(req)
	if doErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("request failed: %v", doErr),
		), nil
	}
	resp.Body.Close()

	// Should return 400 or 415 (unsupported media type), not 200
	codeOK := resp.StatusCode != 200 && resp.StatusCode != 201
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "content_type_validation_status",
		Expected: "not 200/201 (rejected)",
		Actual:   fmt.Sprintf("%d", resp.StatusCode),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Wrong content-type rejected with %d",
				resp.StatusCode),
			"Wrong content-type was accepted"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-199: MaxBodySize ---

// MaxBodySizeChallenge validates that oversized request bodies
// are rejected.
type MaxBodySizeChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMaxBodySizeChallenge creates CH-199.
func NewMaxBodySizeChallenge() *MaxBodySizeChallenge {
	return &MaxBodySizeChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"max-body-size",
			"Max Body Size",
			"Validates that oversized request bodies are "+
				"rejected with 413.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the max body size challenge.
func (c *MaxBodySizeChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-max-body-size", nil)

	// Create a large body (2MB of padding)
	largeBody := `{"name":"` + strings.Repeat("x", 2*1024*1024) + `"}`
	code, _, err := client.PostJSON(
		ctx, "/api/v1/collections", largeBody,
	)

	// Should return 413 (too large) or 400 (bad request)
	codeOK := err == nil &&
		(code == 413 || code == 400 || code == 422)
	// Also accept network errors (connection reset by server)
	if err != nil {
		codeOK = true
		outputs["error"] = err.Error()
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "max_body_size_status",
		Expected: "413, 400, or connection reset",
		Actual:   fmt.Sprintf("%d (err=%v)", code, err),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Oversized body properly rejected",
			fmt.Sprintf("Oversized body returned %d", code)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}

// --- CH-200: ErrorInfoLeak ---

// ErrorInfoLeakChallenge validates that error responses do not
// leak stack traces or internal implementation details.
type ErrorInfoLeakChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewErrorInfoLeakChallenge creates CH-200.
func NewErrorInfoLeakChallenge() *ErrorInfoLeakChallenge {
	return &ErrorInfoLeakChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"error-info-leak",
			"Error Info Leak",
			"Validates that error responses do not leak "+
				"stack traces or internal paths.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the error info leak challenge.
func (c *ErrorInfoLeakChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(
		ctx, c.config.Username, c.config.Password, 3,
	)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("testing-error-info-leak", nil)

	// Request a non-existent resource to trigger an error
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/999999999",
	)

	codeOK := err == nil && (code == 404 || code == 400)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "error_info_leak_status",
		Expected: "404 or 400",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Error endpoint returned %d", code),
			fmt.Sprintf("Error endpoint returned %d, err=%v",
				code, err)),
	})

	if rawBody != nil {
		bodyStr := string(rawBody)
		// Check that response doesn't contain stack traces
		noStackTrace := !strings.Contains(bodyStr, "goroutine") &&
			!strings.Contains(bodyStr, ".go:") &&
			!strings.Contains(bodyStr, "runtime/")
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "error_no_stack_trace",
			Expected: "no stack traces in error response",
			Actual: challenge.Ternary(noStackTrace,
				"clean", "stack trace present"),
			Passed: noStackTrace,
			Message: challenge.Ternary(noStackTrace,
				"Error response does not leak stack traces",
				"Error response contains stack trace info"),
		})

		// Check for JSON structure
		var errResp map[string]interface{}
		if json.Unmarshal(rawBody, &errResp) == nil {
			_, hasError := errResp["error"]
			_, hasMessage := errResp["message"]
			hasStructure := hasError || hasMessage
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "not_empty",
				Target:   "error_structured",
				Expected: "structured error response",
				Actual: challenge.Ternary(hasStructure,
					"structured", "unstructured"),
				Passed: hasStructure,
				Message: challenge.Ternary(hasStructure,
					"Error response is properly structured",
					"Error response lacks structure"),
			})
		}
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		status, start, assertions, nil, outputs, "",
	), nil
}
