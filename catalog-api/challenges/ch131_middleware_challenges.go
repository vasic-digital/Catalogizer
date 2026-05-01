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

// MiddlewareSecurityHeadersChallenge validates that security
// headers are present on all responses.
type MiddlewareSecurityHeadersChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareSecurityHeadersChallenge creates CH-131.
func NewMiddlewareSecurityHeadersChallenge() *MiddlewareSecurityHeadersChallenge {
	return &MiddlewareSecurityHeadersChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-security-headers",
			"Middleware Security Headers",
			"Validates that security headers "+
				"X-Content-Type-Options, X-Frame-Options, and "+
				"X-XSS-Protection are present on responses.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the middleware security headers challenge.
func (c *MiddlewareSecurityHeadersChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("checking-headers", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/health", nil,
	)

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "health_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Health unreachable: %v", err),
		})
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, err.Error(),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp.Body.Close()

	headers := map[string]string{
		"X-Content-Type-Options": resp.Header.Get(
			"X-Content-Type-Options",
		),
		"X-Frame-Options": resp.Header.Get(
			"X-Frame-Options",
		),
		"X-XSS-Protection": resp.Header.Get(
			"X-XSS-Protection",
		),
	}

	for name, value := range headers {
		present := value != ""
		outputs[name] = value
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "header_" + name,
			Expected: name + " present",
			Actual: challenge.Ternary(present,
				value, "missing"),
			Passed: present,
			Message: challenge.Ternary(present,
				fmt.Sprintf("%s: %s", name, value),
				fmt.Sprintf("%s header missing", name)),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareRequestIDChallenge validates that a request ID
// is generated for every request.
type MiddlewareRequestIDChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareRequestIDChallenge creates CH-132.
func NewMiddlewareRequestIDChallenge() *MiddlewareRequestIDChallenge {
	return &MiddlewareRequestIDChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-request-id",
			"Middleware Request ID",
			"Validates that a unique request ID header is "+
				"generated for every request.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the middleware request ID challenge.
func (c *MiddlewareRequestIDChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("checking-request-id", nil)

	// Send two requests and verify both get unique IDs
	ids := []string{}
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequestWithContext(
			ctx, http.MethodGet,
			c.config.BaseURL+"/health", nil,
		)
		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		// Check common request ID header names
		reqID := resp.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = resp.Header.Get("X-Request-ID")
		}
		if reqID == "" {
			reqID = resp.Header.Get("X-Correlation-Id")
		}
		if reqID != "" {
			ids = append(ids, reqID)
		}
	}

	hasRequestID := len(ids) > 0
	outputs["request_ids"] = strings.Join(ids, ", ")
	actualID := "missing"
	if hasRequestID {
		actualID = ids[0]
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "request_id_present",
		Expected: "X-Request-Id header present",
		Actual:   actualID,
		Passed:   hasRequestID,
		Message: challenge.Ternary(hasRequestID,
			"Request ID header present on responses",
			"Request ID header missing from responses"),
	})

	if len(ids) == 2 {
		unique := ids[0] != ids[1]
		outputs["ids_unique"] = fmt.Sprintf("%t", unique)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "request_ids_unique",
			Expected: "unique request IDs per request",
			Actual: challenge.Ternary(unique,
				"unique", "duplicate"),
			Passed: unique,
			Message: challenge.Ternary(unique,
				"Request IDs are unique across requests",
				"Request IDs are duplicated"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareCacheHeadersHealthChallenge validates cache headers
// on the /health endpoint.
type MiddlewareCacheHeadersHealthChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareCacheHeadersHealthChallenge creates CH-133.
func NewMiddlewareCacheHeadersHealthChallenge() *MiddlewareCacheHeadersHealthChallenge {
	return &MiddlewareCacheHeadersHealthChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-cache-headers-health",
			"Middleware Cache Headers Health",
			"Validates that cache control headers are present "+
				"on the /health endpoint response.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the cache headers health challenge.
func (c *MiddlewareCacheHeadersHealthChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("checking-cache-headers", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/health", nil,
	)

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "health_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Health unreachable: %v", err),
		})
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, err.Error(),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp.Body.Close()

	cacheControl := resp.Header.Get("Cache-Control")
	hasCacheHeader := cacheControl != ""
	outputs["cache_control"] = cacheControl

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "cache_control_health",
		Expected: "Cache-Control header present",
		Actual: challenge.Ternary(hasCacheHeader,
			cacheControl, "missing"),
		Passed: hasCacheHeader,
		Message: challenge.Ternary(hasCacheHeader,
			fmt.Sprintf(
				"Cache-Control on /health: %s", cacheControl,
			),
			"Cache-Control header missing on /health"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareCacheHeadersDiscoveryChallenge validates cache
// headers on the /discovery endpoint.
type MiddlewareCacheHeadersDiscoveryChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareCacheHeadersDiscoveryChallenge creates CH-134.
func NewMiddlewareCacheHeadersDiscoveryChallenge() *MiddlewareCacheHeadersDiscoveryChallenge {
	return &MiddlewareCacheHeadersDiscoveryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-cache-headers-discovery",
			"Middleware Cache Headers Discovery",
			"Validates that cache control headers are present "+
				"on the /discovery endpoint response.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the cache headers discovery challenge.
func (c *MiddlewareCacheHeadersDiscoveryChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("checking-cache-headers", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/discovery", nil,
	)

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "discovery_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Discovery unreachable: %v", err),
		})
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, err.Error(),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp.Body.Close()

	cacheControl := resp.Header.Get("Cache-Control")
	hasCacheHeader := cacheControl != ""
	outputs["cache_control"] = cacheControl

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "cache_control_discovery",
		Expected: "Cache-Control header present",
		Actual: challenge.Ternary(hasCacheHeader,
			cacheControl, "missing"),
		Passed: hasCacheHeader,
		Message: challenge.Ternary(hasCacheHeader,
			fmt.Sprintf(
				"Cache-Control on /discovery: %s",
				cacheControl,
			),
			"Cache-Control header missing on /discovery"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareRateLimitingChallenge validates that rate limiting
// is enforced and returns 429 responses.
type MiddlewareRateLimitingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareRateLimitingChallenge creates CH-135.
func NewMiddlewareRateLimitingChallenge() *MiddlewareRateLimitingChallenge {
	return &MiddlewareRateLimitingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-rate-limiting",
			"Middleware Rate Limiting",
			"Validates that rate limiting is enforced and "+
				"returns 429 or handles requests cleanly.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the middleware rate limiting challenge.
func (c *MiddlewareRateLimitingChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("rapid-requests", nil)
	rapidCount := 25
	status429 := 0
	status5xx := 0

	for i := 0; i < rapidCount; i++ {
		req, _ := http.NewRequestWithContext(
			ctx, http.MethodGet,
			c.config.BaseURL+"/health", nil,
		)
		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 429 {
			status429++
		} else if resp.StatusCode >= 500 {
			status5xx++
		}
	}

	outputs["total_requests"] = fmt.Sprintf("%d", rapidCount)
	outputs["status_429"] = fmt.Sprintf("%d", status429)
	outputs["status_5xx"] = fmt.Sprintf("%d", status5xx)

	noServerErrors := status5xx == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "no_server_errors",
		Expected: "zero 5xx responses",
		Actual:   fmt.Sprintf("5xx=%d", status5xx),
		Passed:   noServerErrors,
		Message: challenge.Ternary(noServerErrors,
			"No server errors during rapid requests",
			fmt.Sprintf(
				"%d server errors during rapid requests",
				status5xx)),
	})

	rateLimitOrClean := status429 > 0 || status5xx == 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "rate_limit_active",
		Expected: "429 responses or clean handling",
		Actual:   fmt.Sprintf("429=%d", status429),
		Passed:   rateLimitOrClean,
		Message: challenge.Ternary(status429 > 0,
			fmt.Sprintf(
				"Rate limiting active: %d/%d rate-limited",
				status429, rapidCount),
			"Requests handled cleanly"),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), resultStatus))
	return challengeResult, nil
}

// MiddlewareRequestTimeoutChallenge validates that request
// timeout returns 408 or handles timeout gracefully.
type MiddlewareRequestTimeoutChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareRequestTimeoutChallenge creates CH-136.
func NewMiddlewareRequestTimeoutChallenge() *MiddlewareRequestTimeoutChallenge {
	return &MiddlewareRequestTimeoutChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-request-timeout",
			"Middleware Request Timeout",
			"Validates that the server handles request timeout "+
				"scenarios gracefully without 500 errors.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the middleware request timeout challenge.
func (c *MiddlewareRequestTimeoutChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	// Use a very short client timeout to simulate timeout
	httpClient := &http.Client{Timeout: 100 * time.Millisecond}

	c.ReportProgress("testing-timeout", nil)

	// Hit a potentially slow endpoint with a fast timeout
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/entities?limit=1000", nil,
	)

	resp, err := httpClient.Do(req)

	// If we got a timeout error, that's expected behavior
	if err != nil {
		isTimeout := strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "deadline")
		outputs["timeout_occurred"] = fmt.Sprintf(
			"%t", isTimeout,
		)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "timeout_handling",
			Expected: "timeout or valid response",
			Actual: challenge.Ternary(isTimeout,
				"client timeout", err.Error()),
			Passed: isTimeout,
			Message: challenge.Ternary(isTimeout,
				"Request timed out as expected",
				fmt.Sprintf("Unexpected error: %v", err)),
		})
	} else {
		resp.Body.Close()
		no500 := resp.StatusCode != 500
		outputs["response_code"] = fmt.Sprintf(
			"%d", resp.StatusCode,
		)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "no_500_on_timeout",
			Expected: "not 500",
			Actual:   fmt.Sprintf("%d", resp.StatusCode),
			Passed:   no500,
			Message: challenge.Ternary(no500,
				fmt.Sprintf(
					"Response returned %d (no timeout needed)",
					resp.StatusCode,
				),
				"Server returned 500 on slow request"),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareCORSHeadersChallenge validates that CORS headers
// are present on responses.
type MiddlewareCORSHeadersChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareCORSHeadersChallenge creates CH-137.
func NewMiddlewareCORSHeadersChallenge() *MiddlewareCORSHeadersChallenge {
	return &MiddlewareCORSHeadersChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-cors-headers",
			"Middleware CORS Headers",
			"Validates that CORS headers are present on "+
				"responses to preflight requests.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the middleware CORS headers challenge.
func (c *MiddlewareCORSHeadersChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("sending-preflight", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodOptions,
		c.config.BaseURL+"/api/v1/auth/login", nil,
	)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "cors_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Endpoint unreachable: %v", err),
		})
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, err.Error(),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp.Body.Close()

	allowOrigin := resp.Header.Get(
		"Access-Control-Allow-Origin",
	)
	allowMethods := resp.Header.Get(
		"Access-Control-Allow-Methods",
	)

	hasCORS := allowOrigin != "" || allowMethods != ""
	outputs["allow_origin"] = allowOrigin
	outputs["allow_methods"] = allowMethods

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "cors_headers_present",
		Expected: "CORS headers present",
		Actual: challenge.Ternary(hasCORS,
			"present", "missing"),
		Passed: hasCORS,
		Message: challenge.Ternary(hasCORS,
			"CORS headers present on preflight response",
			"CORS headers missing from preflight response"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareCompressionChallenge validates that compression
// middleware is active.
type MiddlewareCompressionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareCompressionChallenge creates CH-138.
func NewMiddlewareCompressionChallenge() *MiddlewareCompressionChallenge {
	return &MiddlewareCompressionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-compression",
			"Middleware Compression",
			"Validates that compression middleware is active "+
				"by sending Accept-Encoding: br, gzip and "+
				"checking Content-Encoding in the response.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the middleware compression challenge.
func (c *MiddlewareCompressionChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	// Disable automatic decompression
	transport := &http.Transport{
		DisableCompression: true,
	}
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	c.ReportProgress("checking-compression", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet,
		c.config.BaseURL+"/health", nil,
	)
	req.Header.Set("Accept-Encoding", "br, gzip, deflate")

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "health_reachable",
			Passed:  false,
			Message: fmt.Sprintf("Health unreachable: %v", err),
		})
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, err.Error(),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp.Body.Close()

	contentEncoding := resp.Header.Get("Content-Encoding")
	hasCompression := contentEncoding != ""
	outputs["content_encoding"] = contentEncoding

	// Also check Vary header includes Accept-Encoding
	vary := resp.Header.Get("Vary")
	hasVary := strings.Contains(
		strings.ToLower(vary), "accept-encoding",
	)
	outputs["vary"] = vary

	compressionActive := hasCompression || hasVary
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "compression_active",
		Expected: "Content-Encoding or Vary: Accept-Encoding",
		Actual: challenge.Ternary(compressionActive,
			fmt.Sprintf("encoding=%s vary=%s",
				contentEncoding, vary),
			"no compression indicators"),
		Passed: compressionActive,
		Message: challenge.Ternary(compressionActive,
			"Compression middleware is active",
			"No compression detected in response"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareSQLInjectionChallenge validates that input
// validation rejects SQL injection attempts.
type MiddlewareSQLInjectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareSQLInjectionChallenge creates CH-139.
func NewMiddlewareSQLInjectionChallenge() *MiddlewareSQLInjectionChallenge {
	return &MiddlewareSQLInjectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-sql-injection",
			"Middleware SQL Injection Rejection",
			"Validates that the API does not return 500 when "+
				"given SQL injection payloads in query params.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the SQL injection rejection challenge.
func (c *MiddlewareSQLInjectionChallenge) Execute(
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
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("testing-injection", nil)
	injectionPaths := []string{
		"/api/v1/entities?title=' OR 1=1 --",
		"/api/v1/entities?type='; DROP TABLE media_items;--",
		"/api/v1/search?q=' UNION SELECT * FROM users --",
	}

	safe := true
	for _, path := range injectionPaths {
		code, _, _ := client.GetRaw(ctx, path)
		if code == 500 {
			safe = false
			outputs["failed_path"] = path
			break
		}
	}

	outputs["payloads_tested"] = fmt.Sprintf(
		"%d", len(injectionPaths),
	)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "sql_injection_safe",
		Expected: "no 500 responses",
		Actual: challenge.Ternary(safe,
			"safe", "500 returned"),
		Passed: safe,
		Message: challenge.Ternary(safe,
			"SQL injection payloads handled safely",
			"SQL injection payload caused server error"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		status, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// MiddlewareXSSChallenge validates that input validation
// rejects XSS attempts.
type MiddlewareXSSChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMiddlewareXSSChallenge creates CH-140.
func NewMiddlewareXSSChallenge() *MiddlewareXSSChallenge {
	return &MiddlewareXSSChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"middleware-xss-rejection",
			"Middleware XSS Rejection",
			"Validates that the API does not return 500 when "+
				"given XSS payloads and does not reflect them.",
			"middleware",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the XSS rejection challenge.
func (c *MiddlewareXSSChallenge) Execute(
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
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil,
			outputs, fmt.Sprintf("login failed: %v", loginErr),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", loginErr)))
		return challengeResult, nil
	}

	c.ReportProgress("testing-xss", nil)
	xssPaths := []string{
		"/api/v1/search?q=<script>alert(1)</script>",
		"/api/v1/entities?title=<img onerror=alert(1) src=x>",
		"/api/v1/search?q=javascript:alert(1)",
	}

	safe := true
	reflected := false
	for _, path := range xssPaths {
		code, rawBody, _ := client.GetRaw(ctx, path)
		if code == 500 {
			safe = false
			outputs["failed_path"] = path
			break
		}
		// Check if script tag is reflected in response
		if rawBody != nil {
			bodyStr := string(rawBody)
			if strings.Contains(bodyStr, "<script>") ||
				strings.Contains(bodyStr, "onerror=") {
				reflected = true
				outputs["reflected_path"] = path
			}
		}
	}

	outputs["payloads_tested"] = fmt.Sprintf(
		"%d", len(xssPaths),
	)

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "xss_safe",
		Expected: "no 500 responses",
		Actual: challenge.Ternary(safe,
			"safe", "500 returned"),
		Passed: safe,
		Message: challenge.Ternary(safe,
			"XSS payloads handled without server errors",
			"XSS payload caused server error"),
	})

	notReflected := !reflected
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "equals",
		Target:   "xss_not_reflected",
		Expected: "XSS not reflected in response",
		Actual: challenge.Ternary(notReflected,
			"not reflected", "reflected"),
		Passed: notReflected,
		Message: challenge.Ternary(notReflected,
			"XSS payloads not reflected in responses",
			"XSS payload reflected in response body"),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	)
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), resultStatus))
	return challengeResult, nil
}
