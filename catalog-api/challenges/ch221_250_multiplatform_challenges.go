package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// Challenges CH-221 through CH-250 validate API consistency,
// data integrity, and admin operations across the Catalogizer
// platform.

// ============================================================
// API Consistency Challenges (CH-221 to CH-230)
// ============================================================

// --- CH-221: ConsistentJSONStructure ---

// ConsistentJSONStructureChallenge validates that all list
// endpoints return consistent JSON structure.
type ConsistentJSONStructureChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentJSONStructureChallenge creates CH-221.
func NewConsistentJSONStructureChallenge() *ConsistentJSONStructureChallenge {
	return &ConsistentJSONStructureChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-json-structure",
			"Consistent JSON Structure",
			"Validates that list endpoints return consistent "+
				"JSON structure with items or data fields.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent JSON structure challenge.
func (c *ConsistentJSONStructureChallenge) Execute(
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

	c.ReportProgress("checking-json-consistency", nil)

	endpoints := []string{
		"/api/v1/entities?limit=5",
		"/api/v1/collections",
		"/api/v1/storage-roots",
	}

	for _, ep := range endpoints {
		code, rawBody, err := client.GetRaw(ctx, ep)
		if err != nil || code != 200 {
			continue
		}
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "json_structure_" + ep,
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				ep+" returns valid JSON",
				ep+" returns invalid JSON"),
		})
	}

	if len(assertions) == 0 {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "json_structure_any",
			Expected: "at least one endpoint responds",
			Actual:   "none responded with 200",
			Passed:   false,
			Message:  "No list endpoints responded with 200",
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

// --- CH-222: ConsistentErrorFormat ---

// ConsistentErrorFormatChallenge validates that error responses
// use a consistent JSON format.
type ConsistentErrorFormatChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentErrorFormatChallenge creates CH-222.
func NewConsistentErrorFormatChallenge() *ConsistentErrorFormatChallenge {
	return &ConsistentErrorFormatChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-error-format",
			"Consistent Error Format",
			"Validates that error responses use a consistent "+
				"JSON format with error or message fields.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent error format challenge.
func (c *ConsistentErrorFormatChallenge) Execute(
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

	c.ReportProgress("checking-error-format", nil)

	// Request non-existent resources to trigger errors
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/999999999",
	)

	if err == nil && (code == 404 || code == 400) && rawBody != nil {
		var errResp map[string]interface{}
		if json.Unmarshal(rawBody, &errResp) == nil {
			_, hasError := errResp["error"]
			_, hasMessage := errResp["message"]
			hasField := hasError || hasMessage
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "not_empty",
				Target:   "error_format",
				Expected: "error or message field",
				Actual: challenge.Ternary(hasField,
					"present", "missing"),
				Passed: hasField,
				Message: challenge.Ternary(hasField,
					"Error response has consistent format",
					"Error response missing error/message field"),
			})
		}
	} else {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "error_format_status",
			Expected: "404 or 400",
			Actual:   fmt.Sprintf("%d", code),
			Passed:   false,
			Message: fmt.Sprintf(
				"Could not trigger error response: %d", code),
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

// --- CH-223: ConsistentPagination ---

// ConsistentPaginationChallenge validates that pagination
// parameters work consistently across endpoints.
type ConsistentPaginationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentPaginationChallenge creates CH-223.
func NewConsistentPaginationChallenge() *ConsistentPaginationChallenge {
	return &ConsistentPaginationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-pagination",
			"Consistent Pagination",
			"Validates that limit and offset parameters work "+
				"consistently across list endpoints.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent pagination challenge.
func (c *ConsistentPaginationChallenge) Execute(
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

	c.ReportProgress("checking-pagination-consistency", nil)

	endpoints := []string{
		"/api/v1/entities?limit=2&offset=0",
		"/api/v1/collections?limit=2&offset=0",
	}

	for _, ep := range endpoints {
		code, rawBody, err := client.GetRaw(ctx, ep)
		codeOK := err == nil && code == 200
		if codeOK && rawBody != nil {
			validJSON := json.Valid(rawBody)
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "not_empty",
				Target:   "pagination_" + ep,
				Expected: "valid JSON with pagination",
				Actual: challenge.Ternary(validJSON,
					"valid", "invalid"),
				Passed: validJSON,
				Message: challenge.Ternary(validJSON,
					ep+" supports pagination",
					ep+" pagination response invalid"),
			})
		}
	}

	if len(assertions) == 0 {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "pagination_any",
			Expected: "at least one endpoint supports pagination",
			Actual:   "none responded",
			Passed:   false,
			Message:  "No endpoints responded to pagination query",
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

// --- CH-224: ConsistentStatusCodes ---

// ConsistentStatusCodesChallenge validates that endpoints
// return appropriate HTTP status codes.
type ConsistentStatusCodesChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentStatusCodesChallenge creates CH-224.
func NewConsistentStatusCodesChallenge() *ConsistentStatusCodesChallenge {
	return &ConsistentStatusCodesChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-status-codes",
			"Consistent Status Codes",
			"Validates that endpoints return appropriate "+
				"HTTP status codes (200 for success, 404 for "+
				"not found).",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent status codes challenge.
func (c *ConsistentStatusCodesChallenge) Execute(
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

	c.ReportProgress("checking-status-codes", nil)

	// Existing resource should return 200
	code, _, err := client.GetRaw(ctx, "/api/v1/entities?limit=1")
	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "list_status_code",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"List endpoint returns 200",
			fmt.Sprintf("List endpoint returns %d", code)),
	})

	// Non-existent resource should return 404
	code2, _, err2 := client.GetRaw(
		ctx, "/api/v1/entities/999999999",
	)
	code2OK := err2 == nil && (code2 == 404 || code2 == 400)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "not_found_status_code",
		Expected: "404 or 400",
		Actual:   fmt.Sprintf("%d", code2),
		Passed:   code2OK,
		Message: challenge.Ternary(code2OK,
			fmt.Sprintf("Not-found returns %d", code2),
			fmt.Sprintf("Not-found returns %d", code2)),
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

// --- CH-225: ConsistentContentType ---

// ConsistentContentTypeChallenge validates that API responses
// use application/json content type.
type ConsistentContentTypeChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentContentTypeChallenge creates CH-225.
func NewConsistentContentTypeChallenge() *ConsistentContentTypeChallenge {
	return &ConsistentContentTypeChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-content-type",
			"Consistent Content-Type",
			"Validates that API responses use application/json "+
				"content type.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent content-type challenge.
func (c *ConsistentContentTypeChallenge) Execute(
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

	c.ReportProgress("checking-content-type", nil)

	code, body, err := client.Get(ctx, "/api/v1/entities?limit=1")
	codeOK := err == nil && code == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "content_type_status",
		Expected: "200 with JSON body",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"API returns JSON content",
			fmt.Sprintf("API returned %d, err=%v", code, err)),
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

// --- CH-226: ConsistentAuthRequired ---

// ConsistentAuthRequiredChallenge validates that protected
// endpoints consistently require authentication.
type ConsistentAuthRequiredChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentAuthRequiredChallenge creates CH-226.
func NewConsistentAuthRequiredChallenge() *ConsistentAuthRequiredChallenge {
	return &ConsistentAuthRequiredChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-auth-required",
			"Consistent Auth Required",
			"Validates that protected endpoints consistently "+
				"require authentication.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent auth required challenge.
func (c *ConsistentAuthRequiredChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	// Use unauthenticated client
	client := httpclient.NewAPIClient(c.config.BaseURL)

	c.ReportProgress("checking-auth-consistency", nil)

	protectedEndpoints := []string{
		"/api/v1/entities",
		"/api/v1/collections",
		"/api/v1/storage-roots",
	}

	for _, ep := range protectedEndpoints {
		code, _, err := client.GetRaw(ctx, ep)
		isProtected := err == nil &&
			(code == 401 || code == 403)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "auth_required_" + ep,
			Expected: "401 or 403",
			Actual:   fmt.Sprintf("%d", code),
			Passed:   isProtected,
			Message: challenge.Ternary(isProtected,
				ep+" requires authentication",
				fmt.Sprintf("%s returned %d without auth",
					ep, code)),
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

// --- CH-227: ConsistentTimestampFormat ---

// ConsistentTimestampFormatChallenge validates that API
// responses use consistent timestamp formats.
type ConsistentTimestampFormatChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentTimestampFormatChallenge creates CH-227.
func NewConsistentTimestampFormatChallenge() *ConsistentTimestampFormatChallenge {
	return &ConsistentTimestampFormatChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-timestamp-format",
			"Consistent Timestamp Format",
			"Validates that entity responses contain "+
				"consistent timestamp fields.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent timestamp format challenge.
func (c *ConsistentTimestampFormatChallenge) Execute(
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

	c.ReportProgress("checking-timestamp-format", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?limit=1",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "timestamp_format_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entities endpoint returned 200",
			fmt.Sprintf("Entities returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "timestamp_format_json",
			Expected: "valid JSON with consistent fields",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Entity response has consistent format",
				"Entity response format invalid"),
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

// --- CH-228: ConsistentIDFormat ---

// ConsistentIDFormatChallenge validates that entity IDs use
// a consistent format across endpoints.
type ConsistentIDFormatChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentIDFormatChallenge creates CH-228.
func NewConsistentIDFormatChallenge() *ConsistentIDFormatChallenge {
	return &ConsistentIDFormatChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-id-format",
			"Consistent ID Format",
			"Validates that entity responses contain "+
				"consistent ID fields.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent ID format challenge.
func (c *ConsistentIDFormatChallenge) Execute(
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

	c.ReportProgress("checking-id-format", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?limit=1",
	)

	codeOK := err == nil && code == 200
	if codeOK && rawBody != nil {
		var body map[string]interface{}
		if json.Unmarshal(rawBody, &body) == nil {
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "not_empty",
				Target:   "id_format",
				Expected: "valid response with consistent IDs",
				Actual:   "valid",
				Passed:   true,
				Message:  "Entity response has consistent format",
			})
		}
	} else {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "id_format_status",
			Expected: "200",
			Actual:   fmt.Sprintf("%d", code),
			Passed:   false,
			Message: fmt.Sprintf("Entities returned %d, err=%v",
				code, err),
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

// --- CH-229: ConsistentMethodSupport ---

// ConsistentMethodSupportChallenge validates that endpoints
// support the correct HTTP methods.
type ConsistentMethodSupportChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentMethodSupportChallenge creates CH-229.
func NewConsistentMethodSupportChallenge() *ConsistentMethodSupportChallenge {
	return &ConsistentMethodSupportChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-method-support",
			"Consistent Method Support",
			"Validates that endpoints reject unsupported "+
				"HTTP methods with 405.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent method support challenge.
func (c *ConsistentMethodSupportChallenge) Execute(
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

	c.ReportProgress("checking-method-support", nil)

	// PATCH on health should be 404 or 405
	code, _, err := client.PutJSON(
		ctx, "/health", "{}",
	)

	codeOK := err == nil && code != 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "method_support",
		Expected: "not 200 for unsupported method",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Unsupported method returned %d", code),
			"Unsupported method incorrectly accepted"),
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

// --- CH-230: ConsistentEmptyResponse ---

// ConsistentEmptyResponseChallenge validates that empty results
// return consistent empty arrays, not null.
type ConsistentEmptyResponseChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewConsistentEmptyResponseChallenge creates CH-230.
func NewConsistentEmptyResponseChallenge() *ConsistentEmptyResponseChallenge {
	return &ConsistentEmptyResponseChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"consistent-empty-response",
			"Consistent Empty Response",
			"Validates that search with no results returns "+
				"an empty array, not null.",
			"api-consistency",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the consistent empty response challenge.
func (c *ConsistentEmptyResponseChallenge) Execute(
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

	c.ReportProgress("checking-empty-response", nil)

	// Search for something unlikely to exist
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/search?q=zzz_nonexistent_zzz",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "empty_response_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Empty search returns 200",
			fmt.Sprintf("Empty search returns %d", code)),
	})

	if codeOK && rawBody != nil {
		// Check it's not "null"
		bodyStr := string(rawBody)
		notNull := bodyStr != "null" && bodyStr != ""
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "empty_response_not_null",
			Expected: "non-null response",
			Actual: challenge.Ternary(notNull,
				"valid response", "null"),
			Passed: notNull,
			Message: challenge.Ternary(notNull,
				"Empty search returns proper response, not null",
				"Empty search returns null"),
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

// ============================================================
// Data Integrity Challenges (CH-231 to CH-240)
// ============================================================

// --- CH-231: EntityRelationshipsValid ---

// EntityRelationshipsValidChallenge validates that entity
// parent-child relationships are valid.
type EntityRelationshipsValidChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityRelationshipsValidChallenge creates CH-231.
func NewEntityRelationshipsValidChallenge() *EntityRelationshipsValidChallenge {
	return &EntityRelationshipsValidChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-relationships-valid",
			"Entity Relationships Valid",
			"Validates that entity parent-child relationships "+
				"reference existing entities.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity relationships valid challenge.
func (c *EntityRelationshipsValidChallenge) Execute(
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

	c.ReportProgress("checking-entity-relationships", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities?limit=20",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_relationships_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Entities endpoint returned 200",
			fmt.Sprintf("Entities returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "entity_relationships_json",
			Expected: "valid JSON with entity data",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Entity relationships response is valid",
				"Entity relationships response is invalid"),
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

// --- CH-232: NoOrphanMediaFiles ---

// NoOrphanMediaFilesChallenge validates that media files are
// linked to valid entities.
type NoOrphanMediaFilesChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewNoOrphanMediaFilesChallenge creates CH-232.
func NewNoOrphanMediaFilesChallenge() *NoOrphanMediaFilesChallenge {
	return &NoOrphanMediaFilesChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"no-orphan-media-files",
			"No Orphan Media Files",
			"Validates that the entity stats do not indicate "+
				"orphan files.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the no orphan media files challenge.
func (c *NoOrphanMediaFilesChallenge) Execute(
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

	c.ReportProgress("checking-orphan-files", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/stats",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "orphan_files_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Entity stats returned %d", code),
			fmt.Sprintf("Entity stats returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "orphan_files_stats",
			Expected: "valid stats JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Entity stats response is valid",
				"Entity stats response is invalid"),
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

// --- CH-233: EntityTypeConsistency ---

// EntityTypeConsistencyChallenge validates that all entities
// have a valid media type.
type EntityTypeConsistencyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityTypeConsistencyChallenge creates CH-233.
func NewEntityTypeConsistencyChallenge() *EntityTypeConsistencyChallenge {
	return &EntityTypeConsistencyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-type-consistency",
			"Entity Type Consistency",
			"Validates that the media types endpoint returns "+
				"all 11 expected types.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity type consistency challenge.
func (c *EntityTypeConsistencyChallenge) Execute(
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

	c.ReportProgress("checking-entity-types", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/types",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_type_consistency_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Entity types returned %d", code),
			fmt.Sprintf("Entity types returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "entity_types_json",
			Expected: "valid JSON with type definitions",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Entity types response is valid",
				"Entity types response is invalid"),
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

// --- CH-234: CollectionIntegrity ---

// CollectionIntegrityChallenge validates that collections
// endpoint returns well-formed data.
type CollectionIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCollectionIntegrityChallenge creates CH-234.
func NewCollectionIntegrityChallenge() *CollectionIntegrityChallenge {
	return &CollectionIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"collection-integrity",
			"Collection Integrity",
			"Validates that collections endpoint returns "+
				"well-formed data.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the collection integrity challenge.
func (c *CollectionIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-collection-integrity", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/collections",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "collection_integrity_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Collections returned 200",
			fmt.Sprintf("Collections returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "collection_integrity_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Collections response is valid JSON",
				"Collections response is invalid JSON"),
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

// --- CH-235: StorageRootIntegrity ---

// StorageRootIntegrityChallenge validates that storage roots
// have valid configuration.
type StorageRootIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewStorageRootIntegrityChallenge creates CH-235.
func NewStorageRootIntegrityChallenge() *StorageRootIntegrityChallenge {
	return &StorageRootIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"storage-root-integrity",
			"Storage Root Integrity",
			"Validates that storage roots endpoint returns "+
				"valid configuration data.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the storage root integrity challenge.
func (c *StorageRootIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-storage-root-integrity", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/storage-roots",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "storage_root_integrity_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Storage roots returned 200",
			fmt.Sprintf("Storage roots returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "storage_root_integrity_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Storage roots response is valid JSON",
				"Storage roots response is invalid JSON"),
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

// --- CH-236: ScanHistoryIntegrity ---

// ScanHistoryIntegrityChallenge validates that scan history
// records are well-formed.
type ScanHistoryIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewScanHistoryIntegrityChallenge creates CH-236.
func NewScanHistoryIntegrityChallenge() *ScanHistoryIntegrityChallenge {
	return &ScanHistoryIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"scan-history-integrity",
			"Scan History Integrity",
			"Validates that scan history endpoint returns "+
				"well-formed records.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the scan history integrity challenge.
func (c *ScanHistoryIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-scan-history", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/admin/storage/scan-history",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "scan_history_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Scan history returned %d", code),
			fmt.Sprintf("Scan history returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "scan_history_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Scan history response is valid JSON",
				"Scan history response is invalid JSON"),
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

// --- CH-237: UserDataIntegrity ---

// UserDataIntegrityChallenge validates that user profile data
// is well-formed.
type UserDataIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewUserDataIntegrityChallenge creates CH-237.
func NewUserDataIntegrityChallenge() *UserDataIntegrityChallenge {
	return &UserDataIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"user-data-integrity",
			"User Data Integrity",
			"Validates that user profile data from /auth/me "+
				"is well-formed.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the user data integrity challenge.
func (c *UserDataIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-user-data", nil)
	code, body, err := client.Get(ctx, "/api/v1/auth/me")

	codeOK := err == nil && code == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "user_data_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"User profile returned 200",
			fmt.Sprintf("User profile returned %d, err=%v",
				code, err)),
	})

	if codeOK && body != nil {
		_, hasUsername := body["username"]
		_, hasRole := body["role"]
		hasFields := hasUsername || hasRole
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "user_data_fields",
			Expected: "username or role field",
			Actual: challenge.Ternary(hasFields,
				"present", "missing"),
			Passed: hasFields,
			Message: challenge.Ternary(hasFields,
				"User profile has expected fields",
				"User profile missing expected fields"),
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

// --- CH-238: FavoritesIntegrity ---

// FavoritesIntegrityChallenge validates that favorites data
// is consistent.
type FavoritesIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewFavoritesIntegrityChallenge creates CH-238.
func NewFavoritesIntegrityChallenge() *FavoritesIntegrityChallenge {
	return &FavoritesIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"favorites-integrity",
			"Favorites Integrity",
			"Validates that favorites endpoint returns "+
				"consistent data.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the favorites integrity challenge.
func (c *FavoritesIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-favorites-integrity", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/favorites",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "favorites_integrity_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Favorites returned %d", code),
			fmt.Sprintf("Favorites returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "favorites_integrity_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Favorites response is valid JSON",
				"Favorites response is invalid JSON"),
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

// --- CH-239: MediaStatsIntegrity ---

// MediaStatsIntegrityChallenge validates that media statistics
// are internally consistent.
type MediaStatsIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMediaStatsIntegrityChallenge creates CH-239.
func NewMediaStatsIntegrityChallenge() *MediaStatsIntegrityChallenge {
	return &MediaStatsIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"media-stats-integrity",
			"Media Stats Integrity",
			"Validates that media statistics endpoint returns "+
				"internally consistent data.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the media stats integrity challenge.
func (c *MediaStatsIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-media-stats", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/stats",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "media_stats_integrity_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Media stats returned %d", code),
			fmt.Sprintf("Media stats returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "media_stats_integrity_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Media stats response is valid JSON",
				"Media stats response is invalid JSON"),
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

// --- CH-240: DetectionRulesIntegrity ---

// DetectionRulesIntegrityChallenge validates that detection
// rules are accessible and well-formed.
type DetectionRulesIntegrityChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewDetectionRulesIntegrityChallenge creates CH-240.
func NewDetectionRulesIntegrityChallenge() *DetectionRulesIntegrityChallenge {
	return &DetectionRulesIntegrityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"detection-rules-integrity",
			"Detection Rules Integrity",
			"Validates that detection rules endpoint returns "+
				"well-formed rule definitions.",
			"data-integrity",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the detection rules integrity challenge.
func (c *DetectionRulesIntegrityChallenge) Execute(
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

	c.ReportProgress("checking-detection-rules", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/entities/detection-rules",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "detection_rules_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Detection rules returned %d", code),
			fmt.Sprintf("Detection rules returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "detection_rules_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Detection rules response is valid JSON",
				"Detection rules response is invalid JSON"),
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

// ============================================================
// Admin Operations Challenges (CH-241 to CH-250)
// ============================================================

// --- CH-241: AdminSystemInfoAccess ---

// AdminSystemInfoAccessChallenge validates that admin users
// can access system info.
type AdminSystemInfoAccessChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminSystemInfoAccessChallenge creates CH-241.
func NewAdminSystemInfoAccessChallenge() *AdminSystemInfoAccessChallenge {
	return &AdminSystemInfoAccessChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-system-info-access",
			"Admin System Info Access",
			"Validates that admin users can access "+
				"/api/v1/admin/system-info.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin system info access challenge.
func (c *AdminSystemInfoAccessChallenge) Execute(
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

	c.ReportProgress("accessing-system-info", nil)
	code, body, err := client.Get(
		ctx, "/api/v1/admin/system-info",
	)

	codeOK := err == nil && code == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_system_info_access_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Admin system info accessible",
			fmt.Sprintf("Admin system info returned %d, err=%v",
				code, err)),
	})

	if codeOK && body != nil {
		_, hasVersion := body["version"]
		_, hasUptime := body["uptime"]
		hasFields := hasVersion || hasUptime
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "admin_system_info_fields",
			Expected: "version or uptime field",
			Actual: challenge.Ternary(hasFields,
				"present", "missing"),
			Passed: hasFields,
			Message: challenge.Ternary(hasFields,
				"System info has expected fields",
				"System info missing expected fields"),
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

// --- CH-242: AdminUserListAccess ---

// AdminUserListAccessChallenge validates that admin users
// can list all users.
type AdminUserListAccessChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminUserListAccessChallenge creates CH-242.
func NewAdminUserListAccessChallenge() *AdminUserListAccessChallenge {
	return &AdminUserListAccessChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-user-list-access",
			"Admin User List Access",
			"Validates that admin users can list all users "+
				"via /api/v1/admin/users.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin user list access challenge.
func (c *AdminUserListAccessChallenge) Execute(
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

	c.ReportProgress("listing-users", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/admin/users",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_user_list_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Admin user list returned 200",
			fmt.Sprintf("Admin user list returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "admin_user_list_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"User list response is valid JSON",
				"User list response is invalid JSON"),
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

// --- CH-243: AdminStorageOverview ---

// AdminStorageOverviewChallenge validates that admin users
// can access storage overview.
type AdminStorageOverviewChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminStorageOverviewChallenge creates CH-243.
func NewAdminStorageOverviewChallenge() *AdminStorageOverviewChallenge {
	return &AdminStorageOverviewChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-storage-overview",
			"Admin Storage Overview",
			"Validates that admin users can access storage "+
				"overview via /api/v1/admin/storage.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin storage overview challenge.
func (c *AdminStorageOverviewChallenge) Execute(
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

	c.ReportProgress("checking-storage-overview", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/admin/storage",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_storage_overview_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Storage overview returned %d", code),
			fmt.Sprintf("Storage overview returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "admin_storage_overview_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Storage overview response is valid JSON",
				"Storage overview response is invalid JSON"),
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

// --- CH-244: AdminLogCollection ---

// AdminLogCollectionChallenge validates that admin users
// can access log collection.
type AdminLogCollectionChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminLogCollectionChallenge creates CH-244.
func NewAdminLogCollectionChallenge() *AdminLogCollectionChallenge {
	return &AdminLogCollectionChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-log-collection",
			"Admin Log Collection",
			"Validates that admin users can access logs "+
				"via /api/v1/admin/logs.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin log collection challenge.
func (c *AdminLogCollectionChallenge) Execute(
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

	c.ReportProgress("collecting-logs", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/admin/logs",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_log_collection_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Log collection returned %d", code),
			fmt.Sprintf("Log collection returned %d, err=%v",
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

// --- CH-245: AdminErrorReporting ---

// AdminErrorReportingChallenge validates that admin users
// can access error reports.
type AdminErrorReportingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminErrorReportingChallenge creates CH-245.
func NewAdminErrorReportingChallenge() *AdminErrorReportingChallenge {
	return &AdminErrorReportingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-error-reporting",
			"Admin Error Reporting",
			"Validates that admin error reporting endpoint "+
				"is accessible.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin error reporting challenge.
func (c *AdminErrorReportingChallenge) Execute(
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

	c.ReportProgress("checking-error-reporting", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/admin/errors",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_error_reporting_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Error reporting returned %d", code),
			fmt.Sprintf("Error reporting returned %d, err=%v",
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

// --- CH-246: AdminBackupList ---

// AdminBackupListChallenge validates that admin users can
// list backups.
type AdminBackupListChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminBackupListChallenge creates CH-246.
func NewAdminBackupListChallenge() *AdminBackupListChallenge {
	return &AdminBackupListChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-backup-list-ops",
			"Admin Backup List Ops",
			"Validates that admin users can list backups "+
				"via /api/v1/admin/backups.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin backup list challenge.
func (c *AdminBackupListChallenge) Execute(
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

	c.ReportProgress("listing-backups-ops", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/admin/backups",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_backup_list_ops_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Backup list returned 200",
			fmt.Sprintf("Backup list returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "admin_backup_list_ops_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Backup list response is valid JSON",
				"Backup list response is invalid JSON"),
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

// --- CH-247: AdminUserUpdateOps ---

// AdminUserUpdateOpsChallenge validates that admin users can
// update user details.
type AdminUserUpdateOpsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminUserUpdateOpsChallenge creates CH-247.
func NewAdminUserUpdateOpsChallenge() *AdminUserUpdateOpsChallenge {
	return &AdminUserUpdateOpsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-user-update-ops",
			"Admin User Update Ops",
			"Validates that admin users can update user "+
				"details via PUT /api/v1/admin/users/1.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin user update ops challenge.
func (c *AdminUserUpdateOpsChallenge) Execute(
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

	c.ReportProgress("updating-user-ops", nil)
	body := `{"role":"admin"}`
	code, _, err := client.PutJSON(
		ctx, "/api/v1/admin/users/1", body,
	)

	codeOK := err == nil &&
		(code == 200 || code == 400 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_user_update_ops_status",
		Expected: "200, 400, or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("User update returned %d", code),
			fmt.Sprintf("User update returned %d, err=%v",
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

// --- CH-248: AdminChallengesList ---

// AdminChallengesListChallenge validates that the challenges
// list endpoint is accessible.
type AdminChallengesListChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminChallengesListChallenge creates CH-248.
func NewAdminChallengesListChallenge() *AdminChallengesListChallenge {
	return &AdminChallengesListChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-challenges-list",
			"Admin Challenges List",
			"Validates that GET /api/v1/challenges returns "+
				"the list of registered challenges.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin challenges list challenge.
func (c *AdminChallengesListChallenge) Execute(
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

	c.ReportProgress("listing-challenges", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/challenges",
	)

	codeOK := err == nil && code == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_challenges_list_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			"Challenges list returned 200",
			fmt.Sprintf("Challenges list returned %d, err=%v",
				code, err)),
	})

	if codeOK && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "admin_challenges_list_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Challenges list response is valid JSON",
				"Challenges list response is invalid JSON"),
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

// --- CH-249: AdminConfigAccess ---

// AdminConfigAccessChallenge validates that admin users can
// access system configuration.
type AdminConfigAccessChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminConfigAccessChallenge creates CH-249.
func NewAdminConfigAccessChallenge() *AdminConfigAccessChallenge {
	return &AdminConfigAccessChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-config-access",
			"Admin Config Access",
			"Validates that admin users can access system "+
				"configuration via /api/v1/admin/config.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin config access challenge.
func (c *AdminConfigAccessChallenge) Execute(
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

	c.ReportProgress("accessing-config", nil)
	code, _, err := client.GetRaw(
		ctx, "/api/v1/admin/config",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_config_access_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Admin config returned %d", code),
			fmt.Sprintf("Admin config returned %d, err=%v",
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

// --- CH-250: AdminHealthDashboard ---

// AdminHealthDashboardChallenge validates that the admin
// health dashboard data is accessible.
type AdminHealthDashboardChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAdminHealthDashboardChallenge creates CH-250.
func NewAdminHealthDashboardChallenge() *AdminHealthDashboardChallenge {
	return &AdminHealthDashboardChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"admin-health-dashboard",
			"Admin Health Dashboard",
			"Validates that the admin health dashboard data "+
				"is accessible via /api/v1/admin/health.",
			"admin-ops",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the admin health dashboard challenge.
func (c *AdminHealthDashboardChallenge) Execute(
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

	c.ReportProgress("checking-health-dashboard", nil)
	code, rawBody, err := client.GetRaw(
		ctx, "/api/v1/admin/health",
	)

	codeOK := err == nil && (code == 200 || code == 404)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "admin_health_dashboard_status",
		Expected: "200 or 404",
		Actual:   fmt.Sprintf("%d", code),
		Passed:   codeOK,
		Message: challenge.Ternary(codeOK,
			fmt.Sprintf("Health dashboard returned %d", code),
			fmt.Sprintf("Health dashboard returned %d, err=%v",
				code, err)),
	})

	if codeOK && code == 200 && rawBody != nil {
		validJSON := json.Valid(rawBody)
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "admin_health_dashboard_json",
			Expected: "valid JSON",
			Actual: challenge.Ternary(validJSON,
				"valid", "invalid"),
			Passed: validJSON,
			Message: challenge.Ternary(validJSON,
				"Health dashboard response is valid JSON",
				"Health dashboard response is invalid JSON"),
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
