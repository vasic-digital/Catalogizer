package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// CollectionManagementChallenge validates the full collection CRUD
// lifecycle: create a collection, browse it, update its metadata,
// and delete it.
type CollectionManagementChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCollectionManagementChallenge creates CH-029.
func NewCollectionManagementChallenge() *CollectionManagementChallenge {
	return &CollectionManagementChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"collection-management",
			"Collection Management",
			"Validates the full collection CRUD lifecycle: create collection, "+
				"browse contents, update metadata, delete collection.",
			"workflow",
			[]challenge.ID{"collections-api", "entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the collection management challenge.
func (c *CollectionManagementChallenge) Execute(
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

	// Step 1: Create a test collection
	c.ReportProgress("creating-collection", nil)
	createPayload, _ := json.Marshal(map[string]interface{}{
		"name":        "CH-029 Management Test",
		"description": "Created by challenge CH-029 for CRUD test",
	})
	createCode, createBytes, createErr := client.PostJSON(
		ctx, "/api/v1/collections", string(createPayload),
	)
	createOK := createErr == nil && (createCode == 200 || createCode == 201)
	collectionID := float64(0)
	if createOK && len(createBytes) > 0 {
		var resp map[string]interface{}
		if jsonErr := json.Unmarshal(createBytes, &resp); jsonErr == nil {
			if id, ok := resp["id"].(float64); ok {
				collectionID = id
			} else if data, ok := resp["data"].(map[string]interface{}); ok {
				if id, ok := data["id"].(float64); ok {
					collectionID = id
				}
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "create_collection",
		Expected: "200 or 201 with ID",
		Actual:   fmt.Sprintf("HTTP %d, id=%.0f", createCode, collectionID),
		Passed:   createOK && collectionID > 0,
		Message: challenge.Ternary(createOK && collectionID > 0,
			fmt.Sprintf("Collection created with id=%.0f", collectionID),
			fmt.Sprintf("Collection creation failed: code=%d err=%v", createCode, createErr)),
	})
	outputs["collection_id"] = fmt.Sprintf("%.0f", collectionID)

	if collectionID == 0 {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			"collection creation failed",
		), nil
	}

	collIDStr := fmt.Sprintf("%.0f", collectionID)

	// Step 2: Browse collection contents (GET /:id)
	c.ReportProgress("browsing-collection", map[string]any{
		"collection_id": collectionID,
	})
	browseCode, browseBody, browseErr := client.Get(
		ctx, "/api/v1/collections/"+collIDStr,
	)
	browseOK := browseErr == nil && browseCode == 200
	collName := ""
	if browseBody != nil {
		if n, ok := browseBody["name"].(string); ok {
			collName = n
		} else if data, ok := browseBody["data"].(map[string]interface{}); ok {
			if n, ok := data["name"].(string); ok {
				collName = n
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "browse_collection",
		Expected: "200 with collection data",
		Actual:   fmt.Sprintf("HTTP %d, name=%q", browseCode, collName),
		Passed:   browseOK,
		Message: challenge.Ternary(browseOK,
			fmt.Sprintf("Collection browsed: name=%q", collName),
			fmt.Sprintf("Browse failed: code=%d err=%v", browseCode, browseErr)),
	})

	// Step 3: Update collection metadata (PUT /:id)
	c.ReportProgress("updating-collection", map[string]any{
		"collection_id": collectionID,
	})
	updatePayload, _ := json.Marshal(map[string]interface{}{
		"name":        "CH-029 Updated Collection",
		"description": "Updated by challenge CH-029",
	})
	updateCode, _, updateErr := client.PutJSON(
		ctx,
		"/api/v1/collections/"+collIDStr,
		string(updatePayload),
	)
	updateOK := updateErr == nil && (updateCode == 200 || updateCode == 201 || updateCode == 204)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "update_collection",
		Expected: "200, 201, or 204",
		Actual:   fmt.Sprintf("HTTP %d", updateCode),
		Passed:   updateOK,
		Message: challenge.Ternary(updateOK,
			fmt.Sprintf("Collection updated successfully: code=%d", updateCode),
			fmt.Sprintf("Collection update failed: code=%d err=%v", updateCode, updateErr)),
	})

	// Step 4: Delete collection (DELETE /:id)
	c.ReportProgress("deleting-collection", map[string]any{
		"collection_id": collectionID,
	})
	deleteCode, _, deleteErr := client.Delete(
		ctx,
		"/api/v1/collections/"+collIDStr,
	)
	deleteOK := deleteErr == nil && (deleteCode == 200 || deleteCode == 204)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "delete_collection",
		Expected: "200 or 204",
		Actual:   fmt.Sprintf("HTTP %d", deleteCode),
		Passed:   deleteOK,
		Message: challenge.Ternary(deleteOK,
			fmt.Sprintf("Collection deleted successfully: code=%d", deleteCode),
			fmt.Sprintf("Collection delete failed: code=%d err=%v", deleteCode, deleteErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"collection_management_latency": {
			Name:  "collection_management_latency",
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
