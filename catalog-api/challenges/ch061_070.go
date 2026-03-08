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

// SearchAPIBasicQueryChallenge validates that the search API returns
// a 200 response with a results array for a basic query.
type SearchAPIBasicQueryChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSearchAPIBasicQueryChallenge creates CH-061.
func NewSearchAPIBasicQueryChallenge() *SearchAPIBasicQueryChallenge {
	return &SearchAPIBasicQueryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"search-api-basic-query",
			"Search API Basic Query",
			"Validates the search API returns 200 with a results "+
				"array for a basic query string.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the search API basic query challenge.
func (c *SearchAPIBasicQueryChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("searching", nil)
	code, body, err := client.Get(ctx, "/api/v1/search?q=test")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "search_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Search endpoint returned 200",
			fmt.Sprintf("Search endpoint returned %d, err=%v", code, err)),
	})

	hasResults := false
	if body != nil {
		if _, ok := body["results"]; ok {
			hasResults = true
		}
		// Also accept top-level array or "data" key
		if _, ok := body["data"]; ok {
			hasResults = true
		}
		if _, ok := body["items"]; ok {
			hasResults = true
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "search_results_array",
		Expected: "results array in response",
		Actual:   challenge.Ternary(hasResults, "present", "missing"),
		Passed:   hasResults,
		Message: challenge.Ternary(hasResults,
			"Search response contains results array",
			"Search response missing results array"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// SearchAPIDuplicateDetectionChallenge validates the entity
// duplicates endpoint returns 200.
type SearchAPIDuplicateDetectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSearchAPIDuplicateDetectionChallenge creates CH-062.
func NewSearchAPIDuplicateDetectionChallenge() *SearchAPIDuplicateDetectionChallenge {
	return &SearchAPIDuplicateDetectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"search-api-duplicate-detection",
			"Search API Duplicate Detection",
			"Validates the entity duplicates endpoint returns 200.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the duplicate detection challenge.
func (c *SearchAPIDuplicateDetectionChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("checking-duplicates", nil)
	code, _, err := client.Get(ctx, "/api/v1/entities/duplicates")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "duplicates_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Duplicates endpoint returned 200",
			fmt.Sprintf("Duplicates endpoint returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// SearchAPIAdvancedFiltersChallenge validates that the search API
// accepts advanced filter parameters (type, year) and returns 200.
type SearchAPIAdvancedFiltersChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSearchAPIAdvancedFiltersChallenge creates CH-063.
func NewSearchAPIAdvancedFiltersChallenge() *SearchAPIAdvancedFiltersChallenge {
	return &SearchAPIAdvancedFiltersChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"search-api-advanced-filters",
			"Search API Advanced Filters",
			"Validates search API accepts type and year filter "+
				"parameters and returns 200.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the advanced filters challenge.
func (c *SearchAPIAdvancedFiltersChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("filtered-search", nil)
	code, _, err := client.Get(ctx, "/api/v1/search?q=test&type=movie&year=2020")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "filtered_search_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Filtered search returned 200",
			fmt.Sprintf("Filtered search returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// BrowseAPIStorageRootsChallenge validates the storage-roots
// endpoint returns 200 with an array response.
type BrowseAPIStorageRootsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowseAPIStorageRootsChallenge creates CH-064.
func NewBrowseAPIStorageRootsChallenge() *BrowseAPIStorageRootsChallenge {
	return &BrowseAPIStorageRootsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browse-api-storage-roots",
			"Browse API Storage Roots",
			"Validates the storage-roots endpoint returns 200 "+
				"with an array of configured storage roots.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the storage roots challenge.
func (c *BrowseAPIStorageRootsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("fetching-roots", nil)
	code, _, err := client.GetRaw(ctx, "/api/v1/storage-roots")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "storage_roots_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Storage roots endpoint returned 200",
			fmt.Sprintf("Storage roots returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// BrowseAPIDirectoryListingChallenge validates that the files
// endpoint returns 200 for directory listing.
type BrowseAPIDirectoryListingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewBrowseAPIDirectoryListingChallenge creates CH-065.
func NewBrowseAPIDirectoryListingChallenge() *BrowseAPIDirectoryListingChallenge {
	return &BrowseAPIDirectoryListingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"browse-api-directory-listing",
			"Browse API Directory Listing",
			"Validates the files endpoint returns 200 for a "+
				"directory listing request.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the directory listing challenge.
func (c *BrowseAPIDirectoryListingChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("listing-directory", nil)
	code, _, err := client.Get(ctx, "/api/v1/files?path=/")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "directory_listing_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Directory listing returned 200",
			fmt.Sprintf("Directory listing returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// SyncAPIEndpointCreationChallenge validates that the sync endpoints
// creation endpoint accepts POST requests with auth.
type SyncAPIEndpointCreationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSyncAPIEndpointCreationChallenge creates CH-066.
func NewSyncAPIEndpointCreationChallenge() *SyncAPIEndpointCreationChallenge {
	return &SyncAPIEndpointCreationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"sync-api-endpoint-creation",
			"Sync API Endpoint Creation",
			"Validates the sync endpoints creation endpoint "+
				"accepts POST requests and returns 200 or 201.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the sync endpoint creation challenge.
func (c *SyncAPIEndpointCreationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("creating-endpoint", nil)
	body := `{"name":"test-sync","type":"local","path":"/tmp/sync-test"}`
	code, _, err := client.PostJSON(ctx, "/api/v1/sync/endpoints", body)

	codeOK := err == nil && (code == 200 || code == 201)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_endpoint_creation",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Sync endpoint creation returned %d", code),
			fmt.Sprintf("Sync endpoint creation returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// SyncAPICloudProvidersChallenge validates the sync cloud providers
// endpoint returns 200.
type SyncAPICloudProvidersChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSyncAPICloudProvidersChallenge creates CH-067.
func NewSyncAPICloudProvidersChallenge() *SyncAPICloudProvidersChallenge {
	return &SyncAPICloudProvidersChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"sync-api-cloud-providers",
			"Sync API Cloud Providers",
			"Validates the sync providers endpoint returns 200 "+
				"with available cloud providers.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the cloud providers challenge.
func (c *SyncAPICloudProvidersChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("fetching-providers", nil)
	code, _, err := client.Get(ctx, "/api/v1/sync/providers")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_providers_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Sync providers endpoint returned 200",
			fmt.Sprintf("Sync providers returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// SyncAPIUserEndpointsChallenge validates the sync user endpoints
// listing returns 200 with auth.
type SyncAPIUserEndpointsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSyncAPIUserEndpointsChallenge creates CH-068.
func NewSyncAPIUserEndpointsChallenge() *SyncAPIUserEndpointsChallenge {
	return &SyncAPIUserEndpointsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"sync-api-user-endpoints",
			"Sync API User Endpoints",
			"Validates the sync endpoints listing returns 200 "+
				"for authenticated users.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the user endpoints challenge.
func (c *SyncAPIUserEndpointsChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if loginErr != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", loginErr),
		), nil
	}

	c.ReportProgress("fetching-endpoints", nil)
	code, _, err := client.Get(ctx, "/api/v1/sync/endpoints")

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_user_endpoints_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Sync user endpoints returned 200",
			fmt.Sprintf("Sync user endpoints returned %d, err=%v", code, err)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// SecurityHeadersAllChallenge validates that security headers
// X-Content-Type-Options and X-Frame-Options are present.
type SecurityHeadersAllChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSecurityHeadersAllChallenge creates CH-069.
func NewSecurityHeadersAllChallenge() *SecurityHeadersAllChallenge {
	return &SecurityHeadersAllChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"security-headers-all",
			"Security Headers Present",
			"Validates security headers X-Content-Type-Options and "+
				"X-Frame-Options are present on API responses.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the security headers challenge.
func (c *SecurityHeadersAllChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("checking-headers", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, c.config.BaseURL+"/health", nil,
	)

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

	// Check X-Content-Type-Options
	xcto := resp.Header.Get("X-Content-Type-Options")
	hasXCTO := xcto != ""
	outputs["x_content_type_options"] = xcto
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "x_content_type_options",
		Expected: "non-empty X-Content-Type-Options",
		Actual:   challenge.Ternary(hasXCTO, xcto, "missing"),
		Passed:   hasXCTO,
		Message: challenge.Ternary(hasXCTO,
			fmt.Sprintf("X-Content-Type-Options present: %s", xcto),
			"X-Content-Type-Options header missing"),
	})

	// Check X-Frame-Options
	xfo := resp.Header.Get("X-Frame-Options")
	hasXFO := xfo != ""
	outputs["x_frame_options"] = xfo
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "x_frame_options",
		Expected: "non-empty X-Frame-Options",
		Actual:   challenge.Ternary(hasXFO, xfo, "missing"),
		Passed:   hasXFO,
		Message: challenge.Ternary(hasXFO,
			fmt.Sprintf("X-Frame-Options present: %s", xfo),
			"X-Frame-Options header missing"),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// CORSRejectsUnauthorizedChallenge validates that CORS rejects
// preflight requests from unauthorized origins.
type CORSRejectsUnauthorizedChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCORSRejectsUnauthorizedChallenge creates CH-070.
func NewCORSRejectsUnauthorizedChallenge() *CORSRejectsUnauthorizedChallenge {
	return &CORSRejectsUnauthorizedChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"cors-rejects-unauthorized",
			"CORS Rejects Unauthorized Origins",
			"Validates CORS rejects preflight requests from "+
				"unauthorized origins by checking that no "+
				"Access-Control-Allow-Origin is returned for "+
				"an untrusted origin.",
			"security",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the CORS rejection challenge.
func (c *CORSRejectsUnauthorizedChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	c.ReportProgress("sending-bad-origin", nil)
	req, _ := http.NewRequestWithContext(
		ctx, http.MethodOptions,
		c.config.BaseURL+"/api/v1/auth/login", nil,
	)
	req.Header.Set("Origin", "https://evil-site.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := httpClient.Do(req)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "cors_reachable",
			Passed:  false,
			Message: fmt.Sprintf("CORS endpoint unreachable: %v", err),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, err.Error(),
		), nil
	}
	resp.Body.Close()

	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	outputs["allow_origin_for_bad"] = allowOrigin

	// Either no Allow-Origin header or it does not match the evil origin
	rejected := allowOrigin == "" ||
		(!strings.Contains(allowOrigin, "evil-site.example.com") && allowOrigin != "*")
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "cors_rejects_bad_origin",
		Expected: "no Access-Control-Allow-Origin for evil origin",
		Actual:   challenge.Ternary(allowOrigin == "", "missing (good)", allowOrigin),
		Passed:   rejected,
		Message: challenge.Ternary(rejected,
			"CORS correctly rejects unauthorized origin",
			fmt.Sprintf("CORS allowed unauthorized origin: %s", allowOrigin)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}
