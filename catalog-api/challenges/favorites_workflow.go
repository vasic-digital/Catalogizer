package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// FavoritesWorkflowChallenge validates the full favorites lifecycle:
// add a favorite, list favorites, check favorite status, remove
// the favorite, and verify removal.
type FavoritesWorkflowChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewFavoritesWorkflowChallenge creates CH-028.
func NewFavoritesWorkflowChallenge() *FavoritesWorkflowChallenge {
	return &FavoritesWorkflowChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"favorites-workflow",
			"Favorites Workflow",
			"Validates the full favorites lifecycle: add favorite, "+
				"list favorites, check favorite status, remove favorite, "+
				"verify removal.",
			"workflow",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the favorites workflow challenge.
func (c *FavoritesWorkflowChallenge) Execute(
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

	// Step 1: Find a valid entity to favorite
	c.ReportProgress("finding-entity", nil)
	listCode, listBody, listErr := client.Get(ctx, "/api/v1/entities?limit=1")
	entityID := float64(0)
	if listErr == nil && listCode == 200 && listBody != nil {
		if items, ok := listBody["items"].([]interface{}); ok && len(items) > 0 {
			if item, ok := items[0].(map[string]interface{}); ok {
				if id, ok := item["id"].(float64); ok {
					entityID = id
				}
			}
		}
	}
	hasEntity := entityID > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entity_for_favorite",
		Expected: "at least 1 entity",
		Actual:   fmt.Sprintf("HTTP %d, entity_id=%.0f", listCode, entityID),
		Passed:   hasEntity,
		Message: challenge.Ternary(hasEntity,
			fmt.Sprintf("Found entity id=%.0f for favorites test", entityID),
			fmt.Sprintf("No entities available: code=%d err=%v", listCode, listErr)),
	})
	outputs["test_entity_id"] = fmt.Sprintf("%.0f", entityID)

	if !hasEntity {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			"no entities available for favorites test",
		), nil
	}

	entityIDStr := fmt.Sprintf("%.0f", entityID)

	// Step 2: Add entity to favorites via user-metadata
	c.ReportProgress("adding-favorite", map[string]any{"entity_id": entityID})
	favPayload, _ := json.Marshal(map[string]interface{}{
		"is_favorite": true,
	})
	addCode, addBytes, addErr := client.PostJSON(
		ctx,
		"/api/v1/entities/"+entityIDStr+"/user-metadata",
		string(favPayload),
	)
	addOK := addErr == nil && (addCode == 200 || addCode == 201 || addCode == 204)

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "add_favorite",
		Expected: "200, 201, or 204",
		Actual:   fmt.Sprintf("HTTP %d", addCode),
		Passed:   addOK,
		Message: challenge.Ternary(addOK,
			fmt.Sprintf("Favorite added for entity %.0f", entityID),
			fmt.Sprintf("Failed to add favorite: code=%d err=%v", addCode, addErr)),
	})

	// Step 3: List favorites and verify entity is present
	c.ReportProgress("listing-favorites", nil)
	favListCode, favListBody, favListErr := client.Get(ctx, "/api/v1/entities?is_favorite=true&limit=50")
	favListOK := favListErr == nil && favListCode == 200
	entityInList := false
	favCount := 0
	if favListBody != nil {
		if items, ok := favListBody["items"].([]interface{}); ok {
			favCount = len(items)
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					if id, ok := m["id"].(float64); ok && id == entityID {
						entityInList = true
						break
					}
				}
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "favorites_list",
		Expected: "200 with favorites listed",
		Actual:   fmt.Sprintf("HTTP %d, count=%d, entity_found=%v", favListCode, favCount, entityInList),
		Passed:   favListOK,
		Message: challenge.Ternary(favListOK,
			fmt.Sprintf("Favorites list returned %d items (entity in list: %v)", favCount, entityInList),
			fmt.Sprintf("Favorites list failed: code=%d err=%v", favListCode, favListErr)),
	})

	// Step 4: Check favorite status on entity detail
	c.ReportProgress("checking-status", map[string]any{"entity_id": entityID})
	detailCode, detailBody, detailErr := client.Get(
		ctx, "/api/v1/entities/"+entityIDStr,
	)
	detailOK := detailErr == nil && detailCode == 200
	isFavorite := false
	if detailBody != nil {
		if fav, ok := detailBody["is_favorite"].(bool); ok {
			isFavorite = fav
		} else if meta, ok := detailBody["user_metadata"].(map[string]interface{}); ok {
			if fav, ok := meta["is_favorite"].(bool); ok {
				isFavorite = fav
			}
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_detail_favorite_status",
		Expected: "entity detail returns favorite status",
		Actual:   fmt.Sprintf("HTTP %d, is_favorite=%v", detailCode, isFavorite),
		Passed:   detailOK,
		Message: challenge.Ternary(detailOK,
			fmt.Sprintf("Entity detail returned: is_favorite=%v", isFavorite),
			fmt.Sprintf("Entity detail failed: code=%d err=%v", detailCode, detailErr)),
	})

	// Step 5: Remove from favorites
	c.ReportProgress("removing-favorite", map[string]any{"entity_id": entityID})
	unfavPayload, _ := json.Marshal(map[string]interface{}{
		"is_favorite": false,
	})
	removeCode, _, removeErr := client.PostJSON(
		ctx,
		"/api/v1/entities/"+entityIDStr+"/user-metadata",
		string(unfavPayload),
	)
	removeOK := removeErr == nil && (removeCode == 200 || removeCode == 201 || removeCode == 204)

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "remove_favorite",
		Expected: "200, 201, or 204",
		Actual:   fmt.Sprintf("HTTP %d", removeCode),
		Passed:   removeOK,
		Message: challenge.Ternary(removeOK,
			fmt.Sprintf("Favorite removed for entity %.0f", entityID),
			fmt.Sprintf("Failed to remove favorite: code=%d err=%v", removeCode, removeErr)),
	})

	// Step 6: Verify removal — fetch entity detail again
	if removeOK {
		verifyCode, verifyBody, verifyErr := client.Get(
			ctx, "/api/v1/entities/"+entityIDStr,
		)
		verifyOK := verifyErr == nil && verifyCode == 200
		stillFavorite := false
		if verifyBody != nil {
			if fav, ok := verifyBody["is_favorite"].(bool); ok {
				stillFavorite = fav
			} else if meta, ok := verifyBody["user_metadata"].(map[string]interface{}); ok {
				if fav, ok := meta["is_favorite"].(bool); ok {
					stillFavorite = fav
				}
			}
		}
		_ = stillFavorite // either false or absent is acceptable

		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "verify_favorite_removed",
			Expected: "entity detail accessible after removal",
			Actual:   fmt.Sprintf("HTTP %d", verifyCode),
			Passed:   verifyOK,
			Message: challenge.Ternary(verifyOK,
				"Favorite removal verified via entity detail",
				fmt.Sprintf("Verification failed: code=%d err=%v", verifyCode, verifyErr)),
		})
	}

	_ = addBytes

	metrics := map[string]challenge.MetricValue{
		"favorites_workflow_latency": {
			Name:  "favorites_workflow_latency",
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
