package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// AssetLazyLoadingChallenge (CH-013) validates the lazy loading lifecycle.
type AssetLazyLoadingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAssetLazyLoadingChallenge creates CH-013.
func NewAssetLazyLoadingChallenge() *AssetLazyLoadingChallenge {
	return &AssetLazyLoadingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"asset-lazy-loading",
			"Asset Lazy Loading",
			"Validates the lazy loading lifecycle: request, poll, verify resolved content",
			"e2e",
			[]challenge.ID{"asset-serving"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the asset lazy loading challenge.
func (c *AssetLazyLoadingChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Authenticate
	_, loginErr := client.Login(ctx, c.config.Username, c.config.Password)
	if loginErr != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type: "not_empty", Target: "login", Expected: "success",
			Actual: fmt.Sprintf("err=%v", loginErr), Passed: false,
			Message: fmt.Sprintf("Login failed: %v", loginErr),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, "login failed"), nil
	}

	// Step 2: Find a file with cover_url metadata
	searchCode, searchData, searchErr := client.GetRaw(ctx, "/api/v1/media/search?query=&limit=10")
	searchOK := searchErr == nil && searchCode == http.StatusOK
	assertions = append(assertions, challenge.AssertionResult{
		Type: "not_empty", Target: "media_search", Expected: "HTTP 200",
		Actual:  fmt.Sprintf("HTTP %d", searchCode),
		Passed:  searchOK,
		Message: challenge.Ternary(searchOK, "Media search returned results", fmt.Sprintf("Search failed: %v", searchErr)),
	})

	var fileID string
	if searchOK && searchData != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(searchData, &result); err == nil {
			if files, ok := result["files"].([]interface{}); ok && len(files) > 0 {
				if file, ok := files[0].(map[string]interface{}); ok {
					if id, ok := file["id"].(float64); ok {
						fileID = fmt.Sprintf("%.0f", id)
					}
				}
			}
		}
	}

	if fileID == "" {
		fileID = "1" // fallback
	}
	outputs["file_id"] = fileID

	// Step 3: Request asset for the file
	reqBody := fmt.Sprintf(`{"type":"image","source_hint":"","entity_type":"file","entity_id":"%s"}`, fileID)
	postCode, postData, postErr := client.PostJSON(ctx, "/api/v1/assets/request", reqBody)
	postOK := postErr == nil && postCode == http.StatusOK
	var assetID string
	if postOK && postData != nil {
		var resp map[string]interface{}
		if err := json.Unmarshal(postData, &resp); err == nil {
			if id, ok := resp["asset_id"].(string); ok {
				assetID = id
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type: "not_empty", Target: "asset_request", Expected: "HTTP 200 with asset_id",
		Actual:  fmt.Sprintf("HTTP %d, asset_id=%q", postCode, assetID),
		Passed:  postOK && assetID != "",
		Message: challenge.Ternary(postOK && assetID != "", "Asset requested", "Asset request failed"),
	})

	if assetID == "" {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, "no asset_id"), nil
	}

	// Step 4: Poll asset endpoint until ready or timeout (30 seconds)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	var finalStatus string
	var contentType string
	var contentSize int

	pollDeadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(pollDeadline) {
		c.ReportProgress("polling asset", map[string]interface{}{
			"asset_id": assetID,
			"elapsed":  time.Since(start).String(),
		})

		assetReq, _ := http.NewRequestWithContext(ctx, http.MethodGet,
			c.config.BaseURL+"/api/v1/assets/"+assetID, nil)
		resp, err := httpClient.Do(assetReq)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		finalStatus = resp.Header.Get("X-Asset-Status")
		contentType = resp.Header.Get("Content-Type")
		contentSize = int(resp.ContentLength)
		resp.Body.Close()

		if finalStatus == "ready" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	// Step 5: Verify final status
	gotStatus := finalStatus == "pending" || finalStatus == "ready"
	assertions = append(assertions, challenge.AssertionResult{
		Type: "not_empty", Target: "poll_status", Expected: "pending or ready",
		Actual:  finalStatus,
		Passed:  gotStatus,
		Message: challenge.Ternary(gotStatus, fmt.Sprintf("Asset status: %s", finalStatus), fmt.Sprintf("Unexpected status: %s", finalStatus)),
	})

	// Step 6: Verify content type is image
	isImage := strings.HasPrefix(contentType, "image/")
	assertions = append(assertions, challenge.AssertionResult{
		Type: "contains", Target: "content_type", Expected: "image/*",
		Actual:  contentType,
		Passed:  isImage,
		Message: challenge.Ternary(isImage, fmt.Sprintf("Content-Type: %s", contentType), fmt.Sprintf("Not image: %s", contentType)),
	})

	// Step 7: Verify asset appears in by-entity lookup
	entityCode, _, entityErr := client.GetRaw(ctx, fmt.Sprintf("/api/v1/assets/by-entity/file/%s", fileID))
	entityOK := entityErr == nil && entityCode == http.StatusOK
	assertions = append(assertions, challenge.AssertionResult{
		Type: "not_empty", Target: "by_entity_lookup", Expected: "HTTP 200",
		Actual:  fmt.Sprintf("HTTP %d", entityCode),
		Passed:  entityOK,
		Message: challenge.Ternary(entityOK, "By-entity lookup returned assets", "By-entity lookup failed"),
	})

	metrics := map[string]challenge.MetricValue{
		"content_size": {
			Name: "content_size", Value: float64(contentSize), Unit: "bytes",
		},
		"poll_duration": {
			Name: "poll_duration", Value: float64(time.Since(start).Milliseconds()), Unit: "ms",
		},
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}
