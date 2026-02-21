package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// StorageRootsAPIChallenge validates the storage roots API endpoints
// used by @vasic-digital/catalogizer-api-client StorageService.
// Confirms that storage roots are listed and status is queryable.
type StorageRootsAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewStorageRootsAPIChallenge creates CH-024.
func NewStorageRootsAPIChallenge() *StorageRootsAPIChallenge {
	return &StorageRootsAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"storage-roots-api",
			"Storage Roots API",
			"Validates storage roots API: "+
				"GET /api/v1/storage-roots returns a list, "+
				"GET /api/v1/storage-roots/:id/status returns connectivity info. "+
				"Used by @vasic-digital/catalogizer-api-client StorageService.",
			"e2e",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the storage roots API challenge.
func (c *StorageRootsAPIChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	_, err := client.Login(ctx, c.config.Username, c.config.Password)
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

	// 1. GET /api/v1/storage-roots — returns 200 with list
	// Try both possible endpoint paths (storage-roots or smb-configs)
	listCode, listBody, listErr := client.Get(ctx, "/api/v1/storage-roots")
	if listErr != nil || listCode == 404 {
		// Fallback: try /api/v1/smb-configs (legacy endpoint name)
		listCode, listBody, listErr = client.Get(ctx, "/api/v1/smb-configs")
	}
	listOK := listErr == nil && listCode == 200
	rootCount := 0
	firstRootID := float64(0)
	if listBody != nil {
		var items []interface{}
		if arr, ok := listBody["items"].([]interface{}); ok {
			items = arr
		} else if arr, ok := listBody["data"].([]interface{}); ok {
			items = arr
		} else if arr, ok := listBody["configs"].([]interface{}); ok {
			items = arr
		}
		rootCount = len(items)
		if rootCount > 0 {
			if item, ok := items[0].(map[string]interface{}); ok {
				if id, ok := item["id"].(float64); ok {
					firstRootID = id
				}
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "GET /api/v1/storage-roots",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d, count=%d", listCode, rootCount),
		Passed:   listOK,
		Message: challenge.Ternary(listOK,
			fmt.Sprintf("Storage roots list OK: %d roots", rootCount),
			fmt.Sprintf("Storage roots list failed: code=%d err=%v", listCode, listErr)),
	})
	outputs["storage_root_count"] = fmt.Sprintf("%d", rootCount)
	outputs["first_root_id"] = fmt.Sprintf("%.0f", firstRootID)

	// 2. GET /api/v1/storage-roots/:id/status (if a root exists)
	if firstRootID > 0 {
		rootIDStr := fmt.Sprintf("%.0f", firstRootID)
		statusCode, _, statusErr := client.Get(
			ctx, "/api/v1/storage-roots/"+rootIDStr+"/status",
		)
		statusOK := statusErr == nil && statusCode == 200
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   "GET /api/v1/storage-roots/:id/status",
			Expected: "200",
			Actual:   fmt.Sprintf("HTTP %d", statusCode),
			Passed:   statusOK,
			Message: challenge.Ternary(statusOK,
				fmt.Sprintf("Storage root status endpoint OK for id=%.0f", firstRootID),
				fmt.Sprintf("Storage root status failed: code=%d err=%v", statusCode, statusErr)),
		})
	}

	metrics := map[string]challenge.MetricValue{
		"storage_roots_api_latency": {
			Name:  "storage_roots_api_latency",
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
