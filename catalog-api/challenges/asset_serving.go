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

// AssetServingChallenge (CH-012) validates the asset endpoint works end-to-end.
type AssetServingChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewAssetServingChallenge creates CH-012.
func NewAssetServingChallenge() *AssetServingChallenge {
	return &AssetServingChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"asset-serving",
			"Asset Serving",
			"Validates asset request, serving, and by-entity lookup endpoints",
			"e2e",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the asset serving challenge.
func (c *AssetServingChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"api_url": c.config.BaseURL,
	}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Step 1: Authenticate
	_, loginErr := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3)
	loginOK := loginErr == nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "login",
		Expected: "successful login",
		Actual:   challenge.Ternary(loginOK, "logged in", fmt.Sprintf("err=%v", loginErr)),
		Passed:   loginOK,
		Message:  challenge.Ternary(loginOK, "Login succeeded", fmt.Sprintf("Login failed: %v", loginErr)),
	})
	if !loginOK {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, "login failed"), nil
	}

	// Step 2: POST /api/v1/assets/request — create an asset request
	reqBody := `{"type":"image","source_hint":"","entity_type":"file","entity_id":"1"}`
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
		Type:     "not_empty",
		Target:   "asset_request",
		Expected: "HTTP 200 with asset_id",
		Actual:   fmt.Sprintf("HTTP %d, asset_id=%q", postCode, assetID),
		Passed:   postOK && assetID != "",
		Message:  challenge.Ternary(postOK && assetID != "", "Asset request created", fmt.Sprintf("Asset request failed: code=%d err=%v", postCode, postErr)),
	})

	if assetID == "" {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, "no asset_id returned"), nil
	}

	// Step 3: GET /api/v1/assets/:id — verify response
	assetReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/assets/"+assetID, nil)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	assetResp, assetErr := httpClient.Do(assetReq)
	assetOK := assetErr == nil && assetResp != nil && assetResp.StatusCode == http.StatusOK
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "asset_get",
		Expected: "HTTP 200",
		Actual:   challenge.Ternary(assetOK, "HTTP 200", fmt.Sprintf("err=%v", assetErr)),
		Passed:   assetOK,
		Message:  challenge.Ternary(assetOK, "Asset endpoint returned 200", "Asset endpoint failed"),
	})
	if assetResp != nil {
		defer assetResp.Body.Close()
	}

	// Step 4: Verify Content-Type is an image MIME type
	if assetOK {
		ct := assetResp.Header.Get("Content-Type")
		isImage := strings.HasPrefix(ct, "image/")
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "contains",
			Target:   "content_type",
			Expected: "image/*",
			Actual:   ct,
			Passed:   isImage,
			Message:  challenge.Ternary(isImage, fmt.Sprintf("Content-Type is %s", ct), fmt.Sprintf("Expected image/*, got %s", ct)),
		})

		// Step 5: Verify X-Asset-Status header
		assetStatus := assetResp.Header.Get("X-Asset-Status")
		statusOK := assetStatus == "pending" || assetStatus == "ready"
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "x_asset_status",
			Expected: "pending or ready",
			Actual:   assetStatus,
			Passed:   statusOK,
			Message:  challenge.Ternary(statusOK, fmt.Sprintf("X-Asset-Status: %s", assetStatus), fmt.Sprintf("Invalid X-Asset-Status: %q", assetStatus)),
		})
	}

	// Step 6: GET /api/v1/assets/by-entity/file/1 — verify returns asset metadata
	entityCode, _, entityErr := client.GetRaw(ctx, "/api/v1/assets/by-entity/file/1")
	entityOK := entityErr == nil && entityCode == http.StatusOK
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "by_entity",
		Expected: "HTTP 200",
		Actual:   fmt.Sprintf("HTTP %d", entityCode),
		Passed:   entityOK,
		Message:  challenge.Ternary(entityOK, "By-entity endpoint returned asset list", fmt.Sprintf("By-entity failed: code=%d err=%v", entityCode, entityErr)),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}
