package challenges

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// BrowsingWebAppChallenge validates that the web application is
// serving correctly and that the full auth + browsing flow works
// end-to-end via HTTP requests (simulating what the browser does).
// Enforces the zero-warning / zero-error policy by checking that
// all frontend modules resolve, all API endpoints respond, and
// the WebSocket endpoint is reachable.
type BrowsingWebAppChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowsingWebAppChallenge creates CH-010.
func NewBrowsingWebAppChallenge() *BrowsingWebAppChallenge {
	return &BrowsingWebAppChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browsing-web-app",
			"Web App Browsing",
			"Validates the web app serves HTML, all JS modules resolve without errors, "+
				"the auth flow works end-to-end, and all API endpoints used by the frontend "+
				"respond correctly (zero-error policy)",
			"e2e",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// viteErrorIndicators are strings that appear in Vite error overlay responses
// when a module fails to resolve or compile. If any of these appear in a
// fetched module's body, the SPA is broken.
var viteErrorIndicators = []string{
	"Failed to resolve import",
	"Failed to fetch dynamically imported module",
	"[plugin:vite:",
	"Internal server error",
	"does not provide an export named",
	"SyntaxError",
	"TransformError",
}

// criticalModules are the frontend entry points and key modules that Vite
// must be able to transform without errors. We fetch each through the Vite
// dev server and verify no error overlay is returned.
var criticalModules = []string{
	"/src/main.tsx",
	"/src/App.tsx",
	"/src/lib/websocket.ts",
}

// Execute runs the web app browsing challenge.
func (c *BrowsingWebAppChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"web_app_url": c.config.WebAppURL,
		"api_url":     c.config.BaseURL,
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	// Step 1: Web app root loads and returns HTML
	rootResp, rootErr := httpClient.Get(c.config.WebAppURL)
	rootOK := false
	rootHTML := ""
	if rootErr == nil && rootResp != nil {
		defer rootResp.Body.Close()
		body, _ := io.ReadAll(rootResp.Body)
		rootHTML = string(body)
		rootOK = rootResp.StatusCode == 200 && (strings.Contains(rootHTML, "<!doctype html") || strings.Contains(rootHTML, "<!DOCTYPE html"))
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "web_app_root",
		Expected: "HTTP 200 with HTML content",
		Actual:   challenge.Ternary(rootOK, fmt.Sprintf("HTTP 200, %d bytes HTML", len(rootHTML)), fmt.Sprintf("err=%v", rootErr)),
		Passed:   rootOK,
		Message:  challenge.Ternary(rootOK, fmt.Sprintf("Web app root returned %d bytes of HTML", len(rootHTML)), fmt.Sprintf("Web app root failed: %v", rootErr)),
	})
	if !rootOK {
		errMsg := "web app not reachable"
		if rootErr != nil {
			errMsg = rootErr.Error()
		}
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, errMsg), nil
	}

	// Step 2: HTML contains expected page title
	titleOK := strings.Contains(rootHTML, "Catalogizer")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "page_title",
		Expected: "title contains 'Catalogizer'",
		Actual:   challenge.Ternary(titleOK, "found", "not found"),
		Passed:   titleOK,
		Message:  challenge.Ternary(titleOK, "Page title contains 'Catalogizer'", "Page title missing 'Catalogizer'"),
	})

	// Step 3: HTML has a React root element (SPA container)
	reactRootOK := strings.Contains(rootHTML, "id=\"root\"") || strings.Contains(rootHTML, "id='root'")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "react_root",
		Expected: "div#root element present",
		Actual:   challenge.Ternary(reactRootOK, "found", "not found"),
		Passed:   reactRootOK,
		Message:  challenge.Ternary(reactRootOK, "React root element found in HTML", "React root element missing - SPA may not mount"),
	})

	// Step 4: Vite dev assets are included (JS/CSS)
	assetsOK := strings.Contains(rootHTML, "src/main.tsx") || strings.Contains(rootHTML, ".js") || strings.Contains(rootHTML, "@vite")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "app_assets",
		Expected: "JavaScript/CSS assets referenced",
		Actual:   challenge.Ternary(assetsOK, "found", "not found"),
		Passed:   assetsOK,
		Message:  challenge.Ternary(assetsOK, "Application assets found in HTML", "No JS/CSS assets found in HTML"),
	})

	// Step 5: Vite module resolution - fetch critical modules and verify no errors
	// This catches broken imports (e.g. unresolved submodule packages) that cause
	// the Vite error overlay to appear in the browser.
	for _, modulePath := range criticalModules {
		moduleURL := c.config.WebAppURL + modulePath
		modResp, modErr := httpClient.Get(moduleURL)
		modOK := false
		modBody := ""
		modStatus := 0
		errDetail := ""

		if modErr == nil && modResp != nil {
			modStatus = modResp.StatusCode
			bodyBytes, _ := io.ReadAll(modResp.Body)
			modResp.Body.Close()
			modBody = string(bodyBytes)

			if modStatus == 200 {
				// Check for Vite error indicators in the response
				hasError := false
				for _, indicator := range viteErrorIndicators {
					if strings.Contains(modBody, indicator) {
						hasError = true
						errDetail = fmt.Sprintf("Vite error: response contains '%s'", indicator)
						break
					}
				}
				modOK = !hasError
			} else {
				errDetail = fmt.Sprintf("HTTP %d", modStatus)
			}
		} else {
			errDetail = fmt.Sprintf("fetch error: %v", modErr)
		}

		target := "vite_module_" + strings.ReplaceAll(strings.TrimPrefix(modulePath, "/src/"), "/", "_")
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   target,
			Expected: fmt.Sprintf("%s resolves without errors", modulePath),
			Actual:   challenge.Ternary(modOK, fmt.Sprintf("HTTP %d, %d bytes, no errors", modStatus, len(modBody)), errDetail),
			Passed:   modOK,
			Message:  challenge.Ternary(modOK, fmt.Sprintf("Module %s resolves cleanly (zero-error policy)", modulePath), fmt.Sprintf("Module %s FAILED: %s (would cause error overlay)", modulePath, errDetail)),
		})
	}

	// Step 6: Full auth flow - login via API (simulating what the web app does)
	apiClient := httpclient.NewAPIClient(c.config.BaseURL)
	loginResp, loginErr := apiClient.Login(ctx, c.config.Username, c.config.Password)
	loginOK := loginErr == nil && loginResp != nil && apiClient.Token() != ""
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "api_login_flow",
		Expected: "login returns valid session_token",
		Actual:   challenge.Ternary(loginOK, fmt.Sprintf("token length=%d", len(apiClient.Token())), fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message:  challenge.Ternary(loginOK, "API login flow succeeded (simulating web app)", fmt.Sprintf("API login failed: %v", loginErr)),
	})
	if !loginOK {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, "login failed"), nil
	}

	// Step 7: GET /api/v1/auth/status returns authenticated=true (web app polls this)
	statusCode, statusBody, statusErr := apiClient.Get(ctx, "/api/v1/auth/status")
	statusOK := statusErr == nil && statusCode == 200 && statusBody != nil
	authenticated := false
	if statusBody != nil {
		if a, ok := statusBody["authenticated"].(bool); ok {
			authenticated = a
		}
	}
	authStatusOK := statusOK && authenticated
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "auth_status_check",
		Expected: "authenticated=true",
		Actual:   fmt.Sprintf("HTTP %d, authenticated=%v", statusCode, authenticated),
		Passed:   authStatusOK,
		Message:  challenge.Ternary(authStatusOK, "Auth status returns authenticated=true", fmt.Sprintf("Auth status check failed: code=%d auth=%v err=%v", statusCode, authenticated, statusErr)),
	})

	// Step 8: GET /api/v1/auth/permissions returns role info (web app uses this)
	permCode, permBody, permErr := apiClient.Get(ctx, "/api/v1/auth/permissions")
	permOK := permErr == nil && permCode == 200 && permBody != nil
	roleName := ""
	isAdmin := false
	if permBody != nil {
		if r, ok := permBody["role"].(string); ok {
			roleName = r
		}
		if a, ok := permBody["is_admin"].(bool); ok {
			isAdmin = a
		}
	}
	permDataOK := permOK && roleName != "" && isAdmin
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "permissions_check",
		Expected: "role=Admin, is_admin=true",
		Actual:   fmt.Sprintf("HTTP %d, role=%s, is_admin=%v", permCode, roleName, isAdmin),
		Passed:   permDataOK,
		Message:  challenge.Ternary(permDataOK, fmt.Sprintf("Permissions: role=%s, is_admin=%v", roleName, isAdmin), fmt.Sprintf("Permissions check failed: code=%d role=%s err=%v", permCode, roleName, permErr)),
	})

	// Step 9: GET /api/v1/auth/me returns user data (web app dashboard uses this)
	meCode, meBody, meErr := apiClient.Get(ctx, "/api/v1/auth/me")
	meOK := meErr == nil && meCode == 200 && meBody != nil
	meUsername := ""
	if meBody != nil {
		if u, ok := meBody["username"].(string); ok {
			meUsername = u
		}
	}
	meDataOK := meOK && meUsername == c.config.Username
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "user_profile_api",
		Expected: fmt.Sprintf("username=%s", c.config.Username),
		Actual:   fmt.Sprintf("HTTP %d, username=%s", meCode, meUsername),
		Passed:   meDataOK,
		Message:  challenge.Ternary(meDataOK, fmt.Sprintf("User profile API returned username=%s", meUsername), fmt.Sprintf("User profile API failed: code=%d err=%v", meCode, meErr)),
	})

	// Step 10: GET /api/v1/stats/overall - dashboard data endpoint
	statsCode, _, statsErr := apiClient.Get(ctx, "/api/v1/stats/overall")
	statsOK := statsErr == nil && statsCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "dashboard_stats",
		Expected: "HTTP 200",
		Actual:   fmt.Sprintf("HTTP %d", statsCode),
		Passed:   statsOK,
		Message:  challenge.Ternary(statsOK, "Dashboard stats endpoint accessible", fmt.Sprintf("Dashboard stats failed: code=%d err=%v", statsCode, statsErr)),
	})

	// Step 11: GET /api/v1/storage/roots - storage page data
	rootsCode, _, rootsErr := apiClient.Get(ctx, "/api/v1/storage/roots")
	rootsOK := rootsErr == nil && rootsCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "storage_page_data",
		Expected: "HTTP 200",
		Actual:   fmt.Sprintf("HTTP %d", rootsCode),
		Passed:   rootsOK,
		Message:  challenge.Ternary(rootsOK, "Storage roots endpoint accessible for web app", fmt.Sprintf("Storage roots failed: code=%d err=%v", rootsCode, rootsErr)),
	})

	// Step 12: GET /api/v1/media/stats - must return real data (not stub zeros)
	mediaStatsCode, mediaStatsBody, mediaStatsErr := apiClient.Get(ctx, "/api/v1/media/stats")
	mediaStatsOK := mediaStatsErr == nil && mediaStatsCode == 200
	mediaTotalItems := float64(0)
	mediaTotalSize := float64(0)
	if mediaStatsBody != nil {
		if v, ok := mediaStatsBody["total_items"].(float64); ok {
			mediaTotalItems = v
		}
		if v, ok := mediaStatsBody["total_size"].(float64); ok {
			mediaTotalSize = v
		}
	}
	mediaStatsDataOK := mediaStatsOK && mediaTotalItems > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "media_stats_endpoint",
		Expected: "HTTP 200 with total_items > 0",
		Actual:   fmt.Sprintf("HTTP %d, total_items=%.0f, total_size=%.0f", mediaStatsCode, mediaTotalItems, mediaTotalSize),
		Passed:   mediaStatsDataOK,
		Message:  challenge.Ternary(mediaStatsDataOK, fmt.Sprintf("Media stats: %.0f items, %.0f bytes", mediaTotalItems, mediaTotalSize), fmt.Sprintf("Media stats returned zero items (endpoint was a stub?): code=%d items=%.0f err=%v", mediaStatsCode, mediaTotalItems, mediaStatsErr)),
	})

	// Step 13: GET /api/v1/media/search - must return actual items from database
	mediaSearchCode, mediaSearchBody, mediaSearchErr := apiClient.Get(ctx, "/api/v1/media/search?limit=24&offset=0&sort_by=name&sort_order=asc")
	mediaSearchOK := mediaSearchErr == nil && mediaSearchCode == 200
	mediaSearchTotal := float64(0)
	mediaSearchItemCount := 0
	if mediaSearchBody != nil {
		if v, ok := mediaSearchBody["total"].(float64); ok {
			mediaSearchTotal = v
		}
		if items, ok := mediaSearchBody["items"].([]interface{}); ok {
			mediaSearchItemCount = len(items)
		}
	}
	mediaSearchDataOK := mediaSearchOK && mediaSearchItemCount > 0 && mediaSearchTotal > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "media_search_endpoint",
		Expected: "HTTP 200 with items > 0 and total > 0",
		Actual:   fmt.Sprintf("HTTP %d, %d items, total=%.0f", mediaSearchCode, mediaSearchItemCount, mediaSearchTotal),
		Passed:   mediaSearchDataOK,
		Message:  challenge.Ternary(mediaSearchDataOK, fmt.Sprintf("Media search: %d items returned, %.0f total", mediaSearchItemCount, mediaSearchTotal), fmt.Sprintf("Media search returned no items (endpoint was a stub?): code=%d items=%d total=%.0f err=%v", mediaSearchCode, mediaSearchItemCount, mediaSearchTotal, mediaSearchErr)),
	})

	// Step 14: WebSocket endpoint reachable (zero-error policy)
	wsCheckURL := strings.Replace(c.config.BaseURL, "http://", "http://", 1)
	wsResp, wsErr := httpClient.Get(wsCheckURL + "/ws")
	wsReachable := wsErr == nil && wsResp != nil
	if wsResp != nil {
		wsResp.Body.Close()
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "websocket_endpoint",
		Expected: "WebSocket endpoint responds",
		Actual:   challenge.Ternary(wsReachable, "reachable", fmt.Sprintf("err=%v", wsErr)),
		Passed:   wsReachable,
		Message:  challenge.Ternary(wsReachable, "WebSocket endpoint is reachable (no connection errors)", fmt.Sprintf("WebSocket endpoint unreachable: %v (would cause console errors)", wsErr)),
	})

	// Step 15: GET /api/v1/challenges - challenges page data
	chCode, chBody, chErr := apiClient.Get(ctx, "/api/v1/challenges")
	chOK := chErr == nil && chCode == 200
	chCount := 0
	if chBody != nil {
		if cnt, ok := chBody["count"].(float64); ok {
			chCount = int(cnt)
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "challenges_page_data",
		Expected: "HTTP 200 with challenges",
		Actual:   fmt.Sprintf("HTTP %d, %d challenges", chCode, chCount),
		Passed:   chOK,
		Message:  challenge.Ternary(chOK, fmt.Sprintf("Challenges page data: %d challenges", chCount), fmt.Sprintf("Challenges endpoint failed: code=%d err=%v", chCode, chErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"web_app_test_time": {
			Name:  "web_app_test_time",
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
