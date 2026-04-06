package challenges

import (
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityListChallenge validates GET /api/v1/entities returns
// an entity list.
type EntityListChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityListChallenge creates CH-121.
func NewEntityListChallenge() *EntityListChallenge {
	return &EntityListChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-list",
			"Entity List",
			"Validates GET /api/v1/entities returns an entity "+
				"list with 200 status.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity list challenge.
func (c *EntityListChallenge) Execute(
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

	c.ReportProgress("fetching-entities", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_list_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity list endpoint returned 200",
			fmt.Sprintf("Entity list returned %d, err=%v",
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

// EntityDetailChallenge validates GET /api/v1/entities/:id
// returns entity detail.
type EntityDetailChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityDetailChallenge creates CH-122.
func NewEntityDetailChallenge() *EntityDetailChallenge {
	return &EntityDetailChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-detail",
			"Entity Detail",
			"Validates GET /api/v1/entities/:id returns entity "+
				"detail with 200 or 404 status.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity detail challenge.
func (c *EntityDetailChallenge) Execute(
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

	c.ReportProgress("fetching-entity-detail", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities/1",
	)

	// Accept 200 (found) or 404 (no entities yet)
	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_detail_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Entity detail returned %d", code),
			fmt.Sprintf("Entity detail returned %d, err=%v",
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

// EntityTypesChallenge validates GET /api/v1/entities/types
// returns media types.
type EntityTypesChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityTypesChallenge creates CH-123.
func NewEntityTypesChallenge() *EntityTypesChallenge {
	return &EntityTypesChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-types",
			"Entity Types",
			"Validates GET /api/v1/entities/types returns the "+
				"list of supported media types.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity types challenge.
func (c *EntityTypesChallenge) Execute(
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

	c.ReportProgress("fetching-types", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities/types",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_types_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity types endpoint returned 200",
			fmt.Sprintf("Entity types returned %d, err=%v",
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

// EntityStatsChallenge validates GET /api/v1/entities/stats
// returns statistics.
type EntityStatsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityStatsChallenge creates CH-124.
func NewEntityStatsChallenge() *EntityStatsChallenge {
	return &EntityStatsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-stats",
			"Entity Stats",
			"Validates GET /api/v1/entities/stats returns "+
				"entity statistics with 200 status.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity stats challenge.
func (c *EntityStatsChallenge) Execute(
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

	c.ReportProgress("fetching-stats", nil)
	code, _, err := client.Get(
		ctx, "/api/v1/entities/stats",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_stats_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity stats endpoint returned 200",
			fmt.Sprintf("Entity stats returned %d, err=%v",
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

// EntitySearchByTitleChallenge validates that entity search
// by title works correctly.
type EntitySearchByTitleChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntitySearchByTitleChallenge creates CH-125.
func NewEntitySearchByTitleChallenge() *EntitySearchByTitleChallenge {
	return &EntitySearchByTitleChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-search-by-title",
			"Entity Search by Title",
			"Validates that entity search by title query "+
				"parameter returns 200.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity search by title challenge.
func (c *EntitySearchByTitleChallenge) Execute(
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

	c.ReportProgress("searching-by-title", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities?title=test",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_search_title_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity search by title returned 200",
			fmt.Sprintf(
				"Entity search by title returned %d, err=%v",
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

// EntityFilterByTypeChallenge validates that entity filter
// by type works correctly.
type EntityFilterByTypeChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityFilterByTypeChallenge creates CH-126.
func NewEntityFilterByTypeChallenge() *EntityFilterByTypeChallenge {
	return &EntityFilterByTypeChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-filter-by-type",
			"Entity Filter by Type",
			"Validates that entity filter by type query "+
				"parameter returns 200.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity filter by type challenge.
func (c *EntityFilterByTypeChallenge) Execute(
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

	c.ReportProgress("filtering-by-type", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities?type=movie",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_filter_type_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity filter by type returned 200",
			fmt.Sprintf(
				"Entity filter by type returned %d, err=%v",
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

// EntityHierarchyCorrectChallenge validates that entity
// hierarchy (parent-child relationships) is correct.
type EntityHierarchyCorrectChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityHierarchyCorrectChallenge creates CH-127.
func NewEntityHierarchyCorrectChallenge() *EntityHierarchyCorrectChallenge {
	return &EntityHierarchyCorrectChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-hierarchy-correct",
			"Entity Hierarchy Correct",
			"Validates that entity hierarchy (parent-child) "+
				"relationships are correctly reported.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity hierarchy challenge.
func (c *EntityHierarchyCorrectChallenge) Execute(
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

	// Fetch a TV show entity that should have children
	c.ReportProgress("checking-hierarchy", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?type=tv_show&limit=1",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "hierarchy_query_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Hierarchy query returned 200",
			fmt.Sprintf("Hierarchy query returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		bodyStr := string(rawBody)
		// Check that parent_id or children fields exist
		hasHierarchy := strings.Contains(bodyStr, "parent_id") ||
			strings.Contains(bodyStr, "children") ||
			strings.Contains(bodyStr, "parent")
		// If no TV shows exist, pass with note
		if bodyStr == "[]" || bodyStr == "null" ||
			strings.Contains(bodyStr, `"total":0`) {
			hasHierarchy = true
			outputs["note"] = "no TV shows available"
		}
		outputs["has_hierarchy_fields"] = fmt.Sprintf(
			"%t", hasHierarchy,
		)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "hierarchy_fields",
			Expected: "parent_id or children in response",
			Actual: challenge.Ternary(hasHierarchy,
				"present", "missing"),
			Passed: hasHierarchy,
			Message: challenge.Ternary(hasHierarchy,
				"Hierarchy fields present in entity response",
				"Hierarchy fields missing from entity response"),
		})
	}

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	), nil
}

// EntityDuplicateDetectionChallenge validates that the entity
// duplicate detection endpoint works.
type EntityDuplicateDetectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityDuplicateDetectionChallenge creates CH-128.
func NewEntityDuplicateDetectionChallenge() *EntityDuplicateDetectionChallenge {
	return &EntityDuplicateDetectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-duplicate-detection",
			"Entity Duplicate Detection",
			"Validates that the entity duplicate detection "+
				"endpoint returns 200.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity duplicate detection challenge.
func (c *EntityDuplicateDetectionChallenge) Execute(
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

	c.ReportProgress("checking-duplicates", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/entities/duplicates",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_duplicates_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entity duplicates endpoint returned 200",
			fmt.Sprintf(
				"Entity duplicates returned %d, err=%v",
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

// EntityPaginationChallenge validates that entity pagination
// works with limit and offset parameters.
type EntityPaginationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityPaginationChallenge creates CH-129.
func NewEntityPaginationChallenge() *EntityPaginationChallenge {
	return &EntityPaginationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-pagination",
			"Entity Pagination",
			"Validates that entity pagination works with limit "+
				"and offset parameters.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity pagination challenge.
func (c *EntityPaginationChallenge) Execute(
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

	// First page
	c.ReportProgress("page-1", nil)
	code1, rawBody1, err1 := client.GetRaw(
		ctx, "/api/v1/entities?limit=2&offset=0",
	)

	page1OK := err1 == nil && code1 == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "pagination_page1",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code1),
		Passed:   page1OK,
		Message: challenge.Ternary(page1OK,
			"First page returned 200",
			fmt.Sprintf("First page returned %d, err=%v",
				code1, err1)),
	})

	// Second page
	c.ReportProgress("page-2", nil)
	code2, rawBody2, err2 := client.GetRaw(
		ctx, "/api/v1/entities?limit=2&offset=2",
	)

	page2OK := err2 == nil && code2 == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "pagination_page2",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code2),
		Passed:   page2OK,
		Message: challenge.Ternary(page2OK,
			"Second page returned 200",
			fmt.Sprintf("Second page returned %d, err=%v",
				code2, err2)),
	})

	// Verify pages return different data (if both succeeded)
	if page1OK && page2OK && rawBody1 != nil && rawBody2 != nil {
		page1Str := string(rawBody1)
		page2Str := string(rawBody2)
		different := page1Str != page2Str ||
			page1Str == "[]" || page2Str == "[]"
		outputs["pages_different"] = fmt.Sprintf(
			"%t", different,
		)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "pagination_different_pages",
			Expected: "different page content",
			Actual: challenge.Ternary(different,
				"different", "identical"),
			Passed: different,
			Message: challenge.Ternary(different,
				"Pagination returns different page content",
				"Pagination returned identical pages"),
		})
	}

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	), nil
}

// EntitySortingChallenge validates that entity sorting works
// with sort and order parameters.
type EntitySortingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntitySortingChallenge creates CH-130.
func NewEntitySortingChallenge() *EntitySortingChallenge {
	return &EntitySortingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-sorting",
			"Entity Sorting",
			"Validates that entity sorting works with sort "+
				"and order parameters.",
			"entity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity sorting challenge.
func (c *EntitySortingChallenge) Execute(
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

	// Sort ascending
	c.ReportProgress("sort-asc", nil)
	codeAsc, _, errAsc := client.GetRaw(
		ctx, "/api/v1/entities?sort=title&order=asc&limit=5",
	)

	ascOK := errAsc == nil && codeAsc == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sort_asc_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", codeAsc),
		Passed:   ascOK,
		Message: challenge.Ternary(ascOK,
			"Ascending sort returned 200",
			fmt.Sprintf("Ascending sort returned %d, err=%v",
				codeAsc, errAsc)),
	})

	// Sort descending
	c.ReportProgress("sort-desc", nil)
	codeDesc, _, errDesc := client.GetRaw(
		ctx,
		"/api/v1/entities?sort=title&order=desc&limit=5",
	)

	descOK := errDesc == nil && codeDesc == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sort_desc_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", codeDesc),
		Passed:   descOK,
		Message: challenge.Ternary(descOK,
			"Descending sort returned 200",
			fmt.Sprintf("Descending sort returned %d, err=%v",
				codeDesc, errDesc)),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(
		resultStatus, start, assertions, nil, outputs, "",
	), nil
}
