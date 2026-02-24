package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityAggregationChallenge validates that entity aggregation
// endpoints return valid data after a catalog scan has completed.
// Checks stats, types, and entity listing endpoints.
type EntityAggregationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityAggregationChallenge creates CH-016.
func NewEntityAggregationChallenge() *EntityAggregationChallenge {
	return &EntityAggregationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-aggregation",
			"Entity Aggregation",
			"Validates that entity aggregation endpoints return data: "+
				"entity stats show total_entities > 0, entity types "+
				"endpoint returns types with counts, and entity listing "+
				"returns items",
			"e2e",
			[]challenge.ID{"first-catalog-populate"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity aggregation challenge.
func (c *EntityAggregationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Authenticate first
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
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

	// 1. GET /api/v1/entities/stats - total_entities > 0
	statsCode, statsBody, statsErr := client.Get(
		ctx, "/api/v1/entities/stats",
	)
	statsResponds := statsErr == nil && statsCode == 200
	totalEntities := float64(0)
	if statsBody != nil {
		if v, ok := statsBody["total_entities"].(float64); ok {
			totalEntities = v
		} else if data, ok := statsBody["data"].(map[string]interface{}); ok {
			if v, ok := data["total_entities"].(float64); ok {
				totalEntities = v
			}
		}
	}
	statsOK := statsResponds && totalEntities > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entity_stats_total",
		Expected: "total_entities > 0",
		Actual: fmt.Sprintf(
			"HTTP %d, total_entities=%.0f", statsCode, totalEntities,
		),
		Passed: statsOK,
		Message: challenge.Ternary(statsOK,
			fmt.Sprintf(
				"Entity stats: %.0f total entities", totalEntities,
			),
			fmt.Sprintf(
				"Entity stats failed: code=%d total=%.0f err=%v",
				statsCode, totalEntities, statsErr,
			)),
	})
	outputs["total_entities"] = fmt.Sprintf("%.0f", totalEntities)

	// 2. GET /api/v1/entities/types - returns types with counts
	typesCode, typesBody, typesErr := client.Get(
		ctx, "/api/v1/entities/types",
	)
	typesResponds := typesErr == nil && typesCode == 200
	typeCount := 0
	if typesBody != nil {
		if arr, ok := typesBody["types"].([]interface{}); ok {
			typeCount = len(arr)
		} else if data, ok := typesBody["data"].([]interface{}); ok {
			typeCount = len(data)
		}
	}
	typesOK := typesResponds && typeCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entity_types",
		Expected: "> 0 entity types",
		Actual: fmt.Sprintf(
			"HTTP %d, %d types", typesCode, typeCount,
		),
		Passed: typesOK,
		Message: challenge.Ternary(typesOK,
			fmt.Sprintf("Found %d entity types", typeCount),
			fmt.Sprintf(
				"Entity types failed: code=%d count=%d err=%v",
				typesCode, typeCount, typesErr,
			)),
	})
	outputs["entity_type_count"] = fmt.Sprintf("%d", typeCount)

	// 3. GET /api/v1/entities - returns items array with length > 0
	listCode, listBody, listErr := client.Get(
		ctx, "/api/v1/entities",
	)
	listResponds := listErr == nil && listCode == 200
	itemCount := 0
	if listBody != nil {
		if arr, ok := listBody["items"].([]interface{}); ok {
			itemCount = len(arr)
		} else if arr, ok := listBody["data"].([]interface{}); ok {
			itemCount = len(arr)
		} else if arr, ok := listBody["entities"].([]interface{}); ok {
			itemCount = len(arr)
		}
	}
	listOK := listResponds && itemCount > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entity_list",
		Expected: "> 0 entities in listing",
		Actual: fmt.Sprintf(
			"HTTP %d, %d items", listCode, itemCount,
		),
		Passed: listOK,
		Message: challenge.Ternary(listOK,
			fmt.Sprintf("Entity listing returned %d items", itemCount),
			fmt.Sprintf(
				"Entity listing failed: code=%d count=%d err=%v",
				listCode, itemCount, listErr,
			)),
	})
	outputs["entity_list_count"] = fmt.Sprintf("%d", itemCount)

	metrics := map[string]challenge.MetricValue{
		"total_entities": {
			Name:  "total_entities",
			Value: totalEntities,
			Unit:  "count",
		},
		"aggregation_time": {
			Name:  "aggregation_time",
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
