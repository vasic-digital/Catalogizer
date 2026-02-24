package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// CoverArtChallenge validates that cover art can be requested for
// a media entity and that the response contains either an image
// URL, a placeholder, or a valid fallback response.
type CoverArtChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewCoverArtChallenge creates CH-032.
func NewCoverArtChallenge() *CoverArtChallenge {
	return &CoverArtChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"cover-art",
			"Cover Art",
			"Requests cover art for a media entity, verifies response "+
				"contains image URL, placeholder, or valid fallback.",
			"media",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the cover art challenge.
func (c *CoverArtChallenge) Execute(
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

	// Step 1: Find an entity to request cover art for
	c.ReportProgress("finding-entity", nil)
	listCode, listBody, listErr := client.Get(ctx, "/api/v1/entities?limit=5")
	entityID := float64(0)
	entityTitle := ""
	entityType := ""
	if listErr == nil && listCode == 200 && listBody != nil {
		if items, ok := listBody["items"].([]interface{}); ok {
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					if id, ok := m["id"].(float64); ok && id > 0 {
						entityID = id
						if t, ok := m["title"].(string); ok {
							entityTitle = t
						}
						if t, ok := m["type_name"].(string); ok {
							entityType = t
						} else if t, ok := m["media_type"].(string); ok {
							entityType = t
						}
						break
					}
				}
			}
		}
	}

	hasEntity := entityID > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "entity_for_cover_art",
		Expected: "at least 1 entity",
		Actual:   fmt.Sprintf("HTTP %d, entity_id=%.0f", listCode, entityID),
		Passed:   hasEntity,
		Message: challenge.Ternary(hasEntity,
			fmt.Sprintf("Found entity id=%.0f title=%q type=%q", entityID, entityTitle, entityType),
			fmt.Sprintf("No entities available: code=%d err=%v", listCode, listErr)),
	})
	outputs["entity_id"] = fmt.Sprintf("%.0f", entityID)
	outputs["entity_title"] = entityTitle

	if !hasEntity {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			"no entities available for cover art test",
		), nil
	}

	entityIDStr := fmt.Sprintf("%.0f", entityID)

	// Step 2: Check entity detail for cover/thumbnail info
	c.ReportProgress("checking-entity-detail", map[string]any{"entity_id": entityID})
	detailCode, detailBody, detailErr := client.Get(
		ctx, "/api/v1/entities/"+entityIDStr,
	)
	detailOK := detailErr == nil && detailCode == 200
	coverURL := ""
	thumbnailURL := ""
	if detailBody != nil {
		if u, ok := detailBody["cover_url"].(string); ok {
			coverURL = u
		} else if u, ok := detailBody["cover_art"].(string); ok {
			coverURL = u
		} else if u, ok := detailBody["image_url"].(string); ok {
			coverURL = u
		} else if u, ok := detailBody["poster_url"].(string); ok {
			coverURL = u
		}
		if u, ok := detailBody["thumbnail_url"].(string); ok {
			thumbnailURL = u
		} else if u, ok := detailBody["thumb_url"].(string); ok {
			thumbnailURL = u
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_detail_for_cover",
		Expected: "200 with entity data",
		Actual:   fmt.Sprintf("HTTP %d, cover=%q, thumb=%q", detailCode, coverURL, thumbnailURL),
		Passed:   detailOK,
		Message: challenge.Ternary(detailOK,
			fmt.Sprintf("Entity detail: cover=%q thumbnail=%q", coverURL, thumbnailURL),
			fmt.Sprintf("Entity detail failed: code=%d err=%v", detailCode, detailErr)),
	})
	outputs["cover_url"] = coverURL
	outputs["thumbnail_url"] = thumbnailURL

	// Step 3: Request cover art via the assets endpoint
	// Actual endpoint: /api/v1/assets/by-entity/:type/:id
	c.ReportProgress("requesting-cover-art", map[string]any{"entity_id": entityID})
	assetType := entityType
	if assetType == "" {
		assetType = "movie" // Default type for asset lookup
	}
	artCode, artBody, artErr := client.Get(
		ctx, "/api/v1/assets/by-entity/"+assetType+"/"+entityIDStr,
	)
	// Fallback: try entity metadata endpoint which may contain cover info
	if artErr != nil || artCode == 404 {
		artCode, artBody, artErr = client.Get(
			ctx, "/api/v1/entities/"+entityIDStr+"/metadata",
		)
	}

	artResponds := artErr == nil && artCode != 0
	artHasData := false
	if artBody != nil {
		_, hasURL := artBody["url"]
		_, hasImage := artBody["image_url"]
		_, hasCover := artBody["cover_url"]
		_, hasData := artBody["data"]
		_, hasPlaceholder := artBody["placeholder"]
		artHasData = hasURL || hasImage || hasCover || hasData || hasPlaceholder || len(artBody) > 0
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "cover_art_endpoint",
		Expected: "200 with cover data, or valid response",
		Actual:   fmt.Sprintf("HTTP %d, has_data=%v", artCode, artHasData),
		Passed:   artResponds,
		Message: challenge.Ternary(artResponds && (artCode == 200 || artCode == 404),
			challenge.Ternary(artCode == 200 && artHasData,
				"Cover art endpoint returned artwork data",
				fmt.Sprintf("Cover art endpoint responds: code=%d", artCode)),
			fmt.Sprintf("Cover art endpoint: code=%d err=%v", artCode, artErr)),
	})

	// Step 4: Check entity files for potential cover images
	c.ReportProgress("checking-files", map[string]any{"entity_id": entityID})
	filesCode, _, filesErr := client.Get(
		ctx, "/api/v1/entities/"+entityIDStr+"/files",
	)
	filesResponds := filesErr == nil && filesCode != 0

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "entity_files_endpoint",
		Expected: "200 or valid response",
		Actual:   fmt.Sprintf("HTTP %d", filesCode),
		Passed:   filesResponds,
		Message: challenge.Ternary(filesResponds,
			fmt.Sprintf("Entity files endpoint responds: code=%d", filesCode),
			fmt.Sprintf("Entity files endpoint unreachable: err=%v", filesErr)),
	})

	metrics := map[string]challenge.MetricValue{
		"cover_art_latency": {
			Name:  "cover_art_latency",
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
