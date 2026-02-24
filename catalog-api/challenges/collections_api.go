package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// CollectionsAPIChallenge validates that the collections API endpoints
// used by @vasic-digital/collection-manager are functional:
// list returns an array, create returns a new collection with an ID,
// and get-by-ID returns the correct collection.
type CollectionsAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCollectionsAPIChallenge creates CH-021.
func NewCollectionsAPIChallenge() *CollectionsAPIChallenge {
	return &CollectionsAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"collections-api",
			"Collections API",
			"Validates the collections CRUD lifecycle: "+
				"list returns array, create returns new collection with ID, "+
				"get-by-ID returns correct collection. "+
				"Used by @vasic-digital/collection-manager.",
			"e2e",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the collections API challenge.
func (c *CollectionsAPIChallenge) Execute(
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

	// 1. GET /api/v1/collections — expect 200 with items array
	listCode, listBody, listErr := client.Get(ctx, "/api/v1/collections")
	listOK := listErr == nil && listCode == 200
	listCount := 0
	if listBody != nil {
		if arr, ok := listBody["items"].([]interface{}); ok {
			listCount = len(arr)
		} else if arr, ok := listBody["data"].([]interface{}); ok {
			listCount = len(arr)
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "GET /api/v1/collections",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d", listCode),
		Passed:   listOK,
		Message: challenge.Ternary(listOK,
			fmt.Sprintf("Collections list OK: %d collections", listCount),
			fmt.Sprintf("Collections list failed: code=%d err=%v", listCode, listErr)),
	})
	outputs["initial_collection_count"] = fmt.Sprintf("%d", listCount)

	// 2. POST /api/v1/collections — create a test collection
	createPayload, _ := json.Marshal(map[string]interface{}{
		"name":        "CH-021 Test Collection",
		"description": "Created by challenge CH-021",
		"is_public":   false,
		"is_smart":    false,
	})
	createCode, createBytes, createErr := client.PostJSON(
		ctx, "/api/v1/collections", string(createPayload),
	)
	createOK := createErr == nil && (createCode == 200 || createCode == 201)
	newCollectionID := float64(0)
	if createOK && len(createBytes) > 0 {
		var resp map[string]interface{}
		if jsonErr := json.Unmarshal(createBytes, &resp); jsonErr == nil {
			if id, ok := resp["id"].(float64); ok {
				newCollectionID = id
			} else if data, ok := resp["data"].(map[string]interface{}); ok {
				if id, ok := data["id"].(float64); ok {
					newCollectionID = id
				}
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "POST /api/v1/collections",
		Expected: "200 or 201",
		Actual:   fmt.Sprintf("HTTP %d, id=%.0f", createCode, newCollectionID),
		Passed:   createOK && newCollectionID > 0,
		Message: challenge.Ternary(createOK && newCollectionID > 0,
			fmt.Sprintf("Collection created with id=%.0f", newCollectionID),
			fmt.Sprintf("Collection create failed: code=%d err=%v", createCode, createErr)),
	})
	outputs["created_collection_id"] = fmt.Sprintf("%.0f", newCollectionID)

	// 3. GET /api/v1/collections/:id — verify the created collection exists
	if newCollectionID > 0 {
		collID := fmt.Sprintf("%.0f", newCollectionID)
		getCode, getBody, getErr := client.Get(ctx, "/api/v1/collections/"+collID)
		getName := ""
		if getBody != nil {
			if n, ok := getBody["name"].(string); ok {
				getName = n
			} else if data, ok := getBody["data"].(map[string]interface{}); ok {
				if n, ok := data["name"].(string); ok {
					getName = n
				}
			}
		}
		getOK := getErr == nil && getCode == 200 && getName == "CH-021 Test Collection"
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "GET /api/v1/collections/:id",
			Expected: "200 with correct name",
			Actual:   fmt.Sprintf("HTTP %d, name=%q", getCode, getName),
			Passed:   getOK,
			Message: challenge.Ternary(getOK,
				fmt.Sprintf("Collection GET OK: name=%q", getName),
				fmt.Sprintf("Collection GET failed: code=%d name=%q err=%v", getCode, getName, getErr)),
		})
	}

	metrics := map[string]challenge.MetricValue{
		"collections_api_latency": {
			Name:  "collections_api_latency",
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
