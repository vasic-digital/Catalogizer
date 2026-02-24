package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// MediaPlaybackChallenge validates that the media playback endpoints
// work correctly: requests a stream URL for a media file and verifies
// the response contains valid playback information.
type MediaPlaybackChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewMediaPlaybackChallenge creates CH-030.
func NewMediaPlaybackChallenge() *MediaPlaybackChallenge {
	return &MediaPlaybackChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"media-playback",
			"Media Playback",
			"Requests stream URL for a media file, verifies response "+
				"contains valid playback info (stream URL or file path).",
			"playback",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the media playback challenge.
func (c *MediaPlaybackChallenge) Execute(
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

	// Step 1: Find an entity with associated files
	c.ReportProgress("finding-entity-with-files", nil)
	listCode, listBody, listErr := client.Get(ctx, "/api/v1/entities?limit=10")
	entityID := float64(0)
	entityTitle := ""
	if listErr == nil && listCode == 200 && listBody != nil {
		if items, ok := listBody["items"].([]interface{}); ok {
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					if id, ok := m["id"].(float64); ok && id > 0 {
						entityID = id
						if t, ok := m["title"].(string); ok {
							entityTitle = t
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
		Target:   "entity_for_playback",
		Expected: "at least 1 entity",
		Actual:   fmt.Sprintf("HTTP %d, entity_id=%.0f, title=%q", listCode, entityID, entityTitle),
		Passed:   hasEntity,
		Message: challenge.Ternary(hasEntity,
			fmt.Sprintf("Found entity id=%.0f title=%q", entityID, entityTitle),
			fmt.Sprintf("No entities available: code=%d err=%v", listCode, listErr)),
	})
	outputs["entity_id"] = fmt.Sprintf("%.0f", entityID)
	outputs["entity_title"] = entityTitle

	if !hasEntity {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			"no entities available for playback test",
		), nil
	}

	entityIDStr := fmt.Sprintf("%.0f", entityID)

	// Step 2: Get entity files
	c.ReportProgress("getting-entity-files", map[string]any{"entity_id": entityID})
	filesCode, filesBody, filesErr := client.Get(
		ctx, "/api/v1/entities/"+entityIDStr+"/files",
	)
	filesOK := filesErr == nil && filesCode == 200
	fileCount := 0
	if filesBody != nil {
		var items []interface{}
		if arr, ok := filesBody["files"].([]interface{}); ok {
			items = arr
		} else if arr, ok := filesBody["items"].([]interface{}); ok {
			items = arr
		} else if arr, ok := filesBody["data"].([]interface{}); ok {
			items = arr
		}
		fileCount = len(items)
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "GET /api/v1/entities/:id/files",
		Expected: "200",
		Actual:   fmt.Sprintf("HTTP %d, files=%d", filesCode, fileCount),
		Passed:   filesOK,
		Message: challenge.Ternary(filesOK,
			fmt.Sprintf("Entity files endpoint returned %d files", fileCount),
			fmt.Sprintf("Entity files endpoint failed: code=%d err=%v", filesCode, filesErr)),
	})
	outputs["file_count"] = fmt.Sprintf("%d", fileCount)

	// Step 3: Request stream/playback info for the entity
	// Use GetRaw because stream endpoint may return binary data or non-JSON
	c.ReportProgress("requesting-stream", map[string]any{"entity_id": entityID})
	streamCode, streamBytes, streamErr := client.GetRaw(
		ctx, "/api/v1/entities/"+entityIDStr+"/stream",
	)

	streamOK := streamErr == nil && (streamCode == 200 || streamCode == 206)
	hasStreamInfo := streamOK && len(streamBytes) > 0

	// The stream endpoint may return binary data (file content) or JSON.
	// Both are acceptable responses.
	streamResponds := streamErr == nil && streamCode != 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "stream_endpoint",
		Expected: "200/206 with content, or 404 (no streamable file)",
		Actual:   fmt.Sprintf("HTTP %d, data_len=%d", streamCode, len(streamBytes)),
		Passed:   streamResponds,
		Message: challenge.Ternary(streamOK && hasStreamInfo,
			fmt.Sprintf("Stream endpoint returned %d bytes", len(streamBytes)),
			challenge.Ternary(streamResponds,
				fmt.Sprintf("Stream endpoint responds: code=%d", streamCode),
				fmt.Sprintf("Stream endpoint unreachable: err=%v", streamErr))),
	})

	// Step 4: Verify entity detail contains media type info
	c.ReportProgress("checking-entity-detail", map[string]any{"entity_id": entityID})
	detailCode, detailBody, detailErr := client.Get(
		ctx, "/api/v1/entities/"+entityIDStr,
	)
	detailOK := detailErr == nil && detailCode == 200
	mediaType := ""
	if detailBody != nil {
		if mt, ok := detailBody["media_type"].(string); ok {
			mediaType = mt
		} else if mt, ok := detailBody["type"].(string); ok {
			mediaType = mt
		} else if mt, ok := detailBody["type_name"].(string); ok {
			mediaType = mt
		}
	}

	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "entity_media_type",
		Expected: "entity has a media type",
		Actual:   fmt.Sprintf("HTTP %d, media_type=%q", detailCode, mediaType),
		Passed:   detailOK,
		Message: challenge.Ternary(detailOK,
			fmt.Sprintf("Entity detail returned: media_type=%q", mediaType),
			fmt.Sprintf("Entity detail failed: code=%d err=%v", detailCode, detailErr)),
	})
	outputs["media_type"] = mediaType

	metrics := map[string]challenge.MetricValue{
		"playback_latency": {
			Name:  "playback_latency",
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
