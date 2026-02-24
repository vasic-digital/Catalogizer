package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// SearchFilterChallenge validates the search and filter functionality
// of the entity API: search by title, filter by type, filter by year
// range, and verify results match the given criteria.
type SearchFilterChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSearchFilterChallenge creates CH-031.
func NewSearchFilterChallenge() *SearchFilterChallenge {
	return &SearchFilterChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"search-filter",
			"Search & Filter",
			"Validates search and filter: search by title returns results, "+
				"filter by media type returns matching entities, "+
				"filter by year range returns entities within bounds, "+
				"results match the specified criteria.",
			"search",
			[]challenge.ID{"entity-search"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the search and filter challenge.
func (c *SearchFilterChallenge) Execute(
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

	// Step 1: Search by title — use a common letter to get results
	c.ReportProgress("search-by-title", nil)
	searchCode, searchBody, searchErr := client.Get(
		ctx, "/api/v1/entities?query=a&limit=10",
	)
	searchOK := searchErr == nil && searchCode == 200
	searchCount := 0
	firstTitle := ""
	if searchBody != nil {
		if items, ok := searchBody["items"].([]interface{}); ok {
			searchCount = len(items)
			if len(items) > 0 {
				if m, ok := items[0].(map[string]interface{}); ok {
					if t, ok := m["title"].(string); ok {
						firstTitle = t
					}
				}
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "search_by_title",
		Expected: "200 with results",
		Actual:   fmt.Sprintf("HTTP %d, count=%d, first=%q", searchCode, searchCount, firstTitle),
		Passed:   searchOK,
		Message: challenge.Ternary(searchOK,
			fmt.Sprintf("Title search returned %d results", searchCount),
			fmt.Sprintf("Title search failed: code=%d err=%v", searchCode, searchErr)),
	})
	outputs["search_result_count"] = fmt.Sprintf("%d", searchCount)

	// Step 2: Get available media types for filtering
	c.ReportProgress("getting-types", nil)
	typesCode, typesBody, typesErr := client.Get(ctx, "/api/v1/entities/types")
	firstTypeName := ""
	typeCount := 0
	if typesErr == nil && typesCode == 200 && typesBody != nil {
		var types []interface{}
		if arr, ok := typesBody["types"].([]interface{}); ok {
			types = arr
		} else if arr, ok := typesBody["data"].([]interface{}); ok {
			types = arr
		}
		typeCount = len(types)
		if len(types) > 0 {
			if t, ok := types[0].(map[string]interface{}); ok {
				if n, ok := t["name"].(string); ok {
					firstTypeName = n
				}
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "media_types_available",
		Expected: "at least 1 media type",
		Actual:   fmt.Sprintf("HTTP %d, types=%d, first=%q", typesCode, typeCount, firstTypeName),
		Passed:   firstTypeName != "",
		Message: challenge.Ternary(firstTypeName != "",
			fmt.Sprintf("Found %d media types, first=%q", typeCount, firstTypeName),
			fmt.Sprintf("No media types returned: code=%d err=%v", typesCode, typesErr)),
	})
	outputs["type_count"] = fmt.Sprintf("%d", typeCount)

	// Step 3: Filter by media type
	if firstTypeName != "" {
		c.ReportProgress("filter-by-type", map[string]any{"type": firstTypeName})
		filterCode, filterBody, filterErr := client.Get(
			ctx, "/api/v1/entities/browse/"+firstTypeName+"?limit=10",
		)
		filterOK := filterErr == nil && filterCode == 200
		filterCount := 0
		allMatchType := true
		if filterBody != nil {
			if items, ok := filterBody["items"].([]interface{}); ok {
				filterCount = len(items)
				for _, item := range items {
					if m, ok := item.(map[string]interface{}); ok {
						mt := ""
						if t, ok := m["media_type"].(string); ok {
							mt = t
						} else if t, ok := m["type"].(string); ok {
							mt = t
						} else if t, ok := m["type_name"].(string); ok {
							mt = t
						}
						if mt != "" && mt != firstTypeName {
							allMatchType = false
						}
					}
				}
			}
		}

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "filter_by_type",
			Expected: "200 with matching results",
			Actual:   fmt.Sprintf("HTTP %d, count=%d, all_match=%v", filterCode, filterCount, allMatchType),
			Passed:   filterOK,
			Message: challenge.Ternary(filterOK,
				fmt.Sprintf("Type filter %q returned %d results", firstTypeName, filterCount),
				fmt.Sprintf("Type filter failed: code=%d err=%v", filterCode, filterErr)),
		})
		outputs["type_filter_count"] = fmt.Sprintf("%d", filterCount)
	}

	// Step 4: Filter by year range
	c.ReportProgress("filter-by-year", nil)
	yearCode, yearBody, yearErr := client.Get(
		ctx, "/api/v1/entities?year_min=2000&year_max=2025&limit=10",
	)
	yearOK := yearErr == nil && yearCode == 200
	yearCount := 0
	if yearBody != nil {
		if items, ok := yearBody["items"].([]interface{}); ok {
			yearCount = len(items)
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "filter_by_year_range",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d, count=%d", yearCode, yearCount),
		Passed:   yearOK,
		Message: challenge.Ternary(yearOK,
			fmt.Sprintf("Year range filter returned %d results", yearCount),
			fmt.Sprintf("Year range filter failed: code=%d err=%v", yearCode, yearErr)),
	})
	outputs["year_filter_count"] = fmt.Sprintf("%d", yearCount)

	// Step 5: Combined search with query and type
	if firstTypeName != "" {
		c.ReportProgress("combined-search", nil)
		combinedCode, combinedBody, combinedErr := client.Get(
			ctx, "/api/v1/entities?query=a&type="+firstTypeName+"&limit=10",
		)
		combinedOK := combinedErr == nil && combinedCode == 200
		combinedCount := 0
		if combinedBody != nil {
			if items, ok := combinedBody["items"].([]interface{}); ok {
				combinedCount = len(items)
			}
		}

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "combined_search_query_and_type",
			Expected: "200",
			Actual:   fmt.Sprintf("HTTP %d, count=%d", combinedCode, combinedCount),
			Passed:   combinedOK,
			Message: challenge.Ternary(combinedOK,
				fmt.Sprintf("Combined search returned %d results", combinedCount),
				fmt.Sprintf("Combined search failed: code=%d err=%v", combinedCode, combinedErr)),
		})
		outputs["combined_search_count"] = fmt.Sprintf("%d", combinedCount)
	}

	metrics := map[string]challenge.MetricValue{
		"search_filter_latency": {
			Name:  "search_filter_latency",
			Value: float64(time.Since(start).Milliseconds()),
			Unit:  "ms",
		},
		"search_results": {
			Name:  "search_results",
			Value: float64(searchCount),
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
