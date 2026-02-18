package challenges

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityMetadataChallenge validates entity metadata enrichment
// endpoints including metadata retrieval, refresh trigger,
// and user-metadata updates.
type EntityMetadataChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityMetadataChallenge creates CH-018.
func NewEntityMetadataChallenge() *EntityMetadataChallenge {
	return &EntityMetadataChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-metadata",
			"Entity Metadata Enrichment",
			"Validates entity metadata endpoints: metadata retrieval "+
				"returns array, metadata refresh returns 202, and "+
				"user-metadata update returns 200",
			"e2e",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity metadata enrichment challenge.
func (c *EntityMetadataChallenge) Execute(
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

	// Get an entity ID to work with
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
			Expected: "at least 1 entity to test metadata",
			Actual: fmt.Sprintf(
				"HTTP %d, no entities found", listCode,
			),
			Passed:  false,
			Message: "No entities available for metadata testing",
		})
		return c.CreateResult(
			challenge.StatusFailed, start, assertions,
			nil, outputs, "no entities found",
		), nil
	}
	outputs["test_entity_id"] = fmt.Sprintf("%d", entityID)

	// 1. GET /api/v1/entities/:id/metadata - returns metadata
	//    array (may be empty if no enrichment has run yet)
	metaCode, metaBody, metaErr := client.Get(
		ctx,
		fmt.Sprintf("/api/v1/entities/%d/metadata", entityID),
	)
	metaResponds := metaErr == nil && metaCode == 200
	metaCount := 0
	if metaBody != nil {
		if arr, ok := metaBody["metadata"].([]interface{}); ok {
			metaCount = len(arr)
		} else if arr, ok := metaBody["data"].([]interface{}); ok {
			metaCount = len(arr)
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_metadata_get",
		Expected: "HTTP 200 with metadata array",
		Actual: fmt.Sprintf(
			"HTTP %d, %d metadata entries", metaCode, metaCount,
		),
		Passed: metaResponds,
		Message: challenge.Ternary(metaResponds,
			fmt.Sprintf(
				"Metadata endpoint returned %d entries for entity %d",
				metaCount, entityID,
			),
			fmt.Sprintf(
				"Metadata GET failed: code=%d err=%v",
				metaCode, metaErr,
			)),
	})
	outputs["metadata_count"] = fmt.Sprintf("%d", metaCount)

	// 2. POST /api/v1/entities/:id/metadata/refresh - returns 202
	refreshCode, _, refreshErr := client.PostJSON(
		ctx,
		fmt.Sprintf(
			"/api/v1/entities/%d/metadata/refresh", entityID,
		),
		"{}",
	)
	refreshOK := refreshErr == nil &&
		(refreshCode == 200 || refreshCode == 202)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_metadata_refresh",
		Expected: "HTTP 200 or 202 accepted",
		Actual:   fmt.Sprintf("HTTP %d", refreshCode),
		Passed:   refreshOK,
		Message: challenge.Ternary(refreshOK,
			fmt.Sprintf(
				"Metadata refresh accepted (HTTP %d) for entity %d",
				refreshCode, entityID,
			),
			fmt.Sprintf(
				"Metadata refresh failed: code=%d err=%v",
				refreshCode, refreshErr,
			)),
	})

	// 3. PUT /api/v1/entities/:id/user-metadata with
	//    {"favorite": true} - returns 200
	umCode := 0
	var umErr error
	umPath := fmt.Sprintf(
		"%s/api/v1/entities/%d/user-metadata",
		c.config.BaseURL, entityID,
	)
	umReq, umReqErr := http.NewRequestWithContext(
		ctx, http.MethodPut, umPath,
		strings.NewReader(`{"favorite": true}`),
	)
	if umReqErr != nil {
		umErr = umReqErr
	} else {
		umReq.Header.Set("Content-Type", "application/json")
		umReq.Header.Set(
			"Authorization", "Bearer "+client.Token(),
		)
		umResp, umDoErr := http.DefaultClient.Do(umReq)
		if umDoErr != nil {
			umErr = umDoErr
		} else {
			io.ReadAll(umResp.Body)
			umResp.Body.Close()
			umCode = umResp.StatusCode
		}
	}
	umOK := umErr == nil && umCode == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_user_metadata_put",
		Expected: "HTTP 200",
		Actual:   fmt.Sprintf("HTTP %d", umCode),
		Passed:   umOK,
		Message: challenge.Ternary(umOK,
			fmt.Sprintf(
				"User metadata updated for entity %d", entityID,
			),
			fmt.Sprintf(
				"User metadata PUT failed: code=%d err=%v",
				umCode, umErr,
			)),
	})

	metrics := map[string]challenge.MetricValue{
		"metadata_time": {
			Name:  "metadata_time",
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
