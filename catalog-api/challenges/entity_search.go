package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntitySearchChallenge validates that the entity search and browse
// endpoints work correctly — used by @vasic-digital/media-browser
// and @vasic-digital/catalogizer-api-client EntityService.
type EntitySearchChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntitySearchChallenge creates CH-023.
func NewEntitySearchChallenge() *EntitySearchChallenge {
	return &EntitySearchChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-search",
			"Entity Search",
			"Validates entity search and browse endpoints: "+
				"GET /api/v1/entities?query=<term> returns filtered results, "+
				"GET /api/v1/entities/browse/:type returns type-filtered results, "+
				"pagination fields (total, limit, offset) are present. "+
				"Used by @vasic-digital/media-browser EntityBrowser.",
			"e2e",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity search challenge.
func (c *EntitySearchChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "login",
			Passed:  false,
			Message: fmt.Sprintf("Login failed: %v", err),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs, err.Error(),
		), nil
	}

	// 1. GET /api/v1/entities?limit=5 — pagination fields present
	listCode, listBody, listErr := client.Get(ctx, "/api/v1/entities?limit=5")
	listOK := listErr == nil && listCode == 200
	hasPagination := false
	var totalEntities float64
	if listBody != nil {
		_, hasItems := listBody["items"]
		_, hasTotal := listBody["total"]
		_, hasLimit := listBody["limit"]
		_, hasOffset := listBody["offset"]
		hasPagination = hasItems && hasTotal && hasLimit && hasOffset
		if v, ok := listBody["total"].(float64); ok {
			totalEntities = v
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_list_pagination",
		Expected: "items, total, limit, offset fields present",
		Actual:   fmt.Sprintf("HTTP %d, has_pagination=%v, total=%.0f", listCode, hasPagination, totalEntities),
		Passed:   listOK && hasPagination,
		Message: challenge.Ternary(listOK && hasPagination,
			fmt.Sprintf("Entity list has pagination: total=%.0f", totalEntities),
			fmt.Sprintf("Entity list pagination missing: code=%d has_pagination=%v err=%v", listCode, hasPagination, listErr)),
	})
	outputs["total_entities"] = fmt.Sprintf("%.0f", totalEntities)

	// 2. GET /api/v1/entities/types — use first type name for browse test
	typesCode, typesBody, typesErr := client.Get(ctx, "/api/v1/entities/types")
	firstTypeName := ""
	if typesErr == nil && typesCode == 200 && typesBody != nil {
		var types []interface{}
		if arr, ok := typesBody["types"].([]interface{}); ok {
			types = arr
		} else if arr, ok := typesBody["data"].([]interface{}); ok {
			types = arr
		}
		if len(types) > 0 {
			if t, ok := types[0].(map[string]interface{}); ok {
				if n, ok := t["name"].(string); ok {
					firstTypeName = n
				}
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_types_for_browse",
		Expected: "at least one media type",
		Actual:   fmt.Sprintf("HTTP %d, first_type=%q", typesCode, firstTypeName),
		Passed:   firstTypeName != "",
		Message: challenge.Ternary(firstTypeName != "",
			fmt.Sprintf("First media type: %q", firstTypeName),
			fmt.Sprintf("No media types returned: code=%d err=%v", typesCode, typesErr)),
	})

	// 3. GET /api/v1/entities/browse/:type — returns paginated results for a type
	if firstTypeName != "" {
		browseCode, browseBody, browseErr := client.Get(
			ctx, "/api/v1/entities/browse/"+firstTypeName+"?limit=5",
		)
		browseOK := browseErr == nil && browseCode == 200
		browseCount := 0
		if browseBody != nil {
			if arr, ok := browseBody["items"].([]interface{}); ok {
				browseCount = len(arr)
			}
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "GET /api/v1/entities/browse/:type",
			Expected: "200",
			Actual:   fmt.Sprintf("HTTP %d, items=%d", browseCode, browseCount),
			Passed:   browseOK,
			Message: challenge.Ternary(browseOK,
				fmt.Sprintf("Browse by type %q returned %d items", firstTypeName, browseCount),
				fmt.Sprintf("Browse by type failed: code=%d type=%q err=%v", browseCode, firstTypeName, browseErr)),
		})
		outputs["browse_type"] = firstTypeName
		outputs["browse_count"] = fmt.Sprintf("%d", browseCount)
	}

	// 4. GET /api/v1/entities?query=<term> — search returns valid response
	searchCode, searchBody, searchErr := client.Get(ctx, "/api/v1/entities?query=a&limit=5")
	searchOK := searchErr == nil && searchCode == 200
	searchCount := 0
	if searchBody != nil {
		if arr, ok := searchBody["items"].([]interface{}); ok {
			searchCount = len(arr)
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "GET /api/v1/entities?query=a",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d, results=%d", searchCode, searchCount),
		Passed:   searchOK,
		Message: challenge.Ternary(searchOK,
			fmt.Sprintf("Entity search returned %d results", searchCount),
			fmt.Sprintf("Entity search failed: code=%d err=%v", searchCode, searchErr)),
	})
	outputs["search_results"] = fmt.Sprintf("%d", searchCount)

	metrics := map[string]challenge.MetricValue{
		"entity_search_latency": {
			Name:  "entity_search_latency",
			Value: float64(time.Since(start).Milliseconds()),
			Unit:  "ms",
		},
		"total_entities": {
			Name:  "total_entities",
			Value: totalEntities,
			Unit:  "count",
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
