package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityDuplicatesChallenge validates entity duplicate detection
// endpoints including the global duplicates listing and
// per-entity duplicate lookup.
type EntityDuplicatesChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityDuplicatesChallenge creates CH-019.
func NewEntityDuplicatesChallenge() *EntityDuplicatesChallenge {
	return &EntityDuplicatesChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-duplicates",
			"Entity Duplicate Detection",
			"Validates entity duplicate detection endpoints: global "+
				"duplicates returns groups array (may be empty), and "+
				"per-entity duplicates returns duplicates array",
			"e2e",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity duplicate detection challenge.
func (c *EntityDuplicatesChallenge) Execute(
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

	// 1. GET /api/v1/entities/duplicates - returns groups array
	//    (even if empty - not an error, just means no duplicates)
	dupsCode, dupsBody, dupsErr := client.Get(
		ctx, "/api/v1/entities/duplicates",
	)
	dupsResponds := dupsErr == nil && dupsCode == 200
	groupCount := 0
	if dupsBody != nil {
		if arr, ok := dupsBody["groups"].([]interface{}); ok {
			groupCount = len(arr)
		} else if arr, ok := dupsBody["data"].([]interface{}); ok {
			groupCount = len(arr)
		} else if arr, ok := dupsBody["duplicates"].([]interface{}); ok {
			groupCount = len(arr)
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "global_duplicates",
		Expected: "HTTP 200 with groups array",
		Actual: fmt.Sprintf(
			"HTTP %d, %d duplicate groups", dupsCode, groupCount,
		),
		Passed: dupsResponds,
		Message: challenge.Ternary(dupsResponds,
			fmt.Sprintf(
				"Global duplicates endpoint returned %d groups",
				groupCount,
			),
			fmt.Sprintf(
				"Global duplicates failed: code=%d err=%v",
				dupsCode, dupsErr,
			)),
	})
	outputs["duplicate_group_count"] = fmt.Sprintf("%d", groupCount)

	// Get an entity ID for per-entity duplicate check
	listCode, listBody, listErr := client.Get(
		ctx, "/api/v1/entities?limit=1",
	)
	entityID := int64(0)
	if listErr == nil && listCode == 200 && listBody != nil {
		for _, key := range []string{"items", "data", "entities"} {
			if arr, ok := listBody[key].([]interface{}); ok &&
				len(arr) > 0 {
				if first, ok := arr[0].(map[string]interface{}); ok {
					if id, ok := first["id"].(float64); ok {
						entityID = int64(id)
					}
				}
				break
			}
		}
	}

	if entityID == 0 {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "entity_lookup",
			Expected: "at least 1 entity for duplicate check",
			Actual: fmt.Sprintf(
				"HTTP %d, no entities found", listCode,
			),
			Passed:  false,
			Message: "No entities available for per-entity duplicate test",
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions,
			nil, outputs, "no entities found",
		), nil
	}
	outputs["test_entity_id"] = fmt.Sprintf("%d", entityID)

	// 2. GET /api/v1/entities/:id/duplicates - returns
	//    duplicates array
	edCode, edBody, edErr := client.Get(
		ctx,
		fmt.Sprintf("/api/v1/entities/%d/duplicates", entityID),
	)
	edResponds := edErr == nil && edCode == 200
	edCount := 0
	if edBody != nil {
		if arr, ok := edBody["duplicates"].([]interface{}); ok {
			edCount = len(arr)
		} else if arr, ok := edBody["data"].([]interface{}); ok {
			edCount = len(arr)
		} else if arr, ok := edBody["items"].([]interface{}); ok {
			edCount = len(arr)
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_duplicates",
		Expected: "HTTP 200 with duplicates array",
		Actual: fmt.Sprintf(
			"HTTP %d, %d duplicates for entity %d",
			edCode, edCount, entityID,
		),
		Passed: edResponds,
		Message: challenge.Ternary(edResponds,
			fmt.Sprintf(
				"Entity %d has %d duplicates", entityID, edCount,
			),
			fmt.Sprintf(
				"Entity duplicates failed: code=%d err=%v",
				edCode, edErr,
			)),
	})
	outputs["entity_duplicate_count"] = fmt.Sprintf("%d", edCount)

	metrics := map[string]challenge.MetricValue{
		"duplicate_groups": {
			Name:  "duplicate_groups",
			Value: float64(groupCount),
			Unit:  "count",
		},
		"detection_time": {
			Name:  "detection_time",
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
