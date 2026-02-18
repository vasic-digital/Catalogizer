package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityBrowsingChallenge validates entity browsing endpoints
// including filtered listing, entity detail, and paginated
// browse by media type.
type EntityBrowsingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityBrowsingChallenge creates CH-017.
func NewEntityBrowsingChallenge() *EntityBrowsingChallenge {
	return &EntityBrowsingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-browsing",
			"Entity Browsing",
			"Validates entity browsing endpoints: filtered listing by "+
				"type, entity detail with media_type and file_count, "+
				"and paginated browse by media type",
			"e2e",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity browsing challenge.
func (c *EntityBrowsingChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Authenticate first
	_, err := client.Login(ctx, c.config.Username, c.config.Password)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "login",
			Passed:  false,
			Message: fmt.Sprintf("Login failed: %v", err),
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions,
			nil, outputs, err.Error(),
		), nil
	}

	// 1. GET /api/v1/entities?type=movie - returns items
	movieCode, movieBody, movieErr := client.Get(
		ctx, "/api/v1/entities?type=movie",
	)
	movieResponds := movieErr == nil && movieCode == 200
	movieCount := 0
	if movieBody != nil {
		if arr, ok := movieBody["items"].([]interface{}); ok {
			movieCount = len(arr)
		} else if arr, ok := movieBody["data"].([]interface{}); ok {
			movieCount = len(arr)
		} else if arr, ok := movieBody["entities"].([]interface{}); ok {
			movieCount = len(arr)
		}
	}
	movieOK := movieResponds && movieCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entities_by_type_movie",
		Expected: "> 0 movie entities",
		Actual: fmt.Sprintf(
			"HTTP %d, %d movies", movieCode, movieCount,
		),
		Passed: movieOK,
		Message: challenge.Ternary(movieOK,
			fmt.Sprintf("Movie entities: %d items", movieCount),
			fmt.Sprintf(
				"Movie filter failed: code=%d count=%d err=%v",
				movieCode, movieCount, movieErr,
			)),
	})
	outputs["movie_count"] = fmt.Sprintf("%d", movieCount)

	// 2. GET /api/v1/entities/:id - returns full entity detail
	//    with media_type and file_count
	entityID := int64(0)
	if movieBody != nil {
		for _, key := range []string{"items", "data", "entities"} {
			if arr, ok := movieBody[key].([]interface{}); ok && len(arr) > 0 {
				if first, ok := arr[0].(map[string]interface{}); ok {
					if id, ok := first["id"].(float64); ok {
						entityID = int64(id)
					}
				}
				break
			}
		}
	}

	detailOK := false
	hasMediaType := false
	hasFileCount := false
	detailCode := 0
	if entityID > 0 {
		var detailBody map[string]interface{}
		var detailErr error
		detailCode, detailBody, detailErr = client.Get(
			ctx, fmt.Sprintf("/api/v1/entities/%d", entityID),
		)
		detailOK = detailErr == nil && detailCode == 200 &&
			detailBody != nil
		if detailOK {
			if _, ok := detailBody["media_type"].(string); ok {
				hasMediaType = true
			}
			if _, ok := detailBody["file_count"]; ok {
				hasFileCount = true
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_detail",
		Expected: "HTTP 200 with media_type and file_count",
		Actual: fmt.Sprintf(
			"HTTP %d, media_type=%v, file_count=%v",
			detailCode, hasMediaType, hasFileCount,
		),
		Passed: detailOK && hasMediaType && hasFileCount,
		Message: challenge.Ternary(
			detailOK && hasMediaType && hasFileCount,
			fmt.Sprintf(
				"Entity %d detail has media_type and file_count",
				entityID,
			),
			fmt.Sprintf(
				"Entity detail incomplete: code=%d media_type=%v "+
					"file_count=%v",
				detailCode, hasMediaType, hasFileCount,
			)),
	})

	// 3. GET /api/v1/entities/browse/movie - paginated results
	browseCode, browseBody, browseErr := client.Get(
		ctx, "/api/v1/entities/browse/movie",
	)
	browseResponds := browseErr == nil && browseCode == 200
	browseCount := 0
	if browseBody != nil {
		if arr, ok := browseBody["items"].([]interface{}); ok {
			browseCount = len(arr)
		} else if arr, ok := browseBody["data"].([]interface{}); ok {
			browseCount = len(arr)
		} else if arr, ok := browseBody["entities"].([]interface{}); ok {
			browseCount = len(arr)
		}
	}
	browseOK := browseResponds && browseCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entity_browse_movie",
		Expected: "> 0 items in paginated browse",
		Actual: fmt.Sprintf(
			"HTTP %d, %d items", browseCode, browseCount,
		),
		Passed: browseOK,
		Message: challenge.Ternary(browseOK,
			fmt.Sprintf(
				"Paginated browse returned %d movie items",
				browseCount,
			),
			fmt.Sprintf(
				"Browse movie failed: code=%d count=%d err=%v",
				browseCode, browseCount, browseErr,
			)),
	})
	outputs["browse_movie_count"] = fmt.Sprintf("%d", browseCount)

	metrics := map[string]challenge.MetricValue{
		"browse_time": {
			Name:  "browse_time",
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
