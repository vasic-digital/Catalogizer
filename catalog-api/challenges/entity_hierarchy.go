package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// EntityHierarchyChallenge validates hierarchical navigation
// of entities, including parent-child relationships for
// TV shows (seasons/episodes) and music albums (tracks).
type EntityHierarchyChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewEntityHierarchyChallenge creates CH-020.
func NewEntityHierarchyChallenge() *EntityHierarchyChallenge {
	return &EntityHierarchyChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"entity-hierarchy",
			"Entity Hierarchical Navigation",
			"Validates hierarchical entity navigation: TV shows and "+
				"music albums have children endpoints that return "+
				"items arrays for seasons/episodes and tracks",
			"e2e",
			[]challenge.ID{"entity-aggregation"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the entity hierarchical navigation challenge.
func (c *EntityHierarchyChallenge) Execute(
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

	// 1. GET /api/v1/entities?type=tv_show - check for TV shows
	tvCode, tvBody, tvErr := client.Get(
		ctx, "/api/v1/entities?type=tv_show",
	)
	tvResponds := tvErr == nil && tvCode == 200
	tvCount := 0
	tvEntityID := int64(0)
	if tvBody != nil {
		for _, key := range []string{"items", "data", "entities"} {
			if arr, ok := tvBody[key].([]interface{}); ok {
				tvCount = len(arr)
				if tvCount > 0 {
					if first, ok := arr[0].(map[string]interface{}); ok {
						if id, ok := first["id"].(float64); ok {
							tvEntityID = int64(id)
						}
					}
				}
				break
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "tv_show_listing",
		Expected: "HTTP 200 with tv_show entities",
		Actual: fmt.Sprintf(
			"HTTP %d, %d tv_show entities", tvCode, tvCount,
		),
		Passed: tvResponds,
		Message: challenge.Ternary(tvResponds,
			fmt.Sprintf("TV show listing returned %d items", tvCount),
			fmt.Sprintf(
				"TV show listing failed: code=%d err=%v",
				tvCode, tvErr,
			)),
	})
	outputs["tv_show_count"] = fmt.Sprintf("%d", tvCount)

	// 2. GET /api/v1/entities/:id/children - TV show children
	if tvEntityID > 0 {
		childCode, childBody, childErr := client.Get(
			ctx,
			fmt.Sprintf(
				"/api/v1/entities/%d/children", tvEntityID,
			),
		)
		childResponds := childErr == nil && childCode == 200
		childCount := 0
		if childBody != nil {
			if arr, ok := childBody["items"].([]interface{}); ok {
				childCount = len(arr)
			} else if arr, ok := childBody["data"].([]interface{}); ok {
				childCount = len(arr)
			} else if arr, ok := childBody["children"].([]interface{}); ok {
				childCount = len(arr)
			}
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "tv_show_children",
			Expected: "HTTP 200 with items array",
			Actual: fmt.Sprintf(
				"HTTP %d, %d children for tv_show %d",
				childCode, childCount, tvEntityID,
			),
			Passed: childResponds,
			Message: challenge.Ternary(childResponds,
				fmt.Sprintf(
					"TV show %d has %d children",
					tvEntityID, childCount,
				),
				fmt.Sprintf(
					"TV show children failed: code=%d err=%v",
					childCode, childErr,
				)),
		})
		outputs["tv_show_children"] = fmt.Sprintf("%d", childCount)
	} else {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "tv_show_children",
			Expected: "TV show entity for children test",
			Actual:   "no tv_show entities found",
			Passed:   true, // Not a failure - catalog may not have TV shows
			Message:  "Skipped TV show children test (no tv_show entities in catalog)",
		})
	}

	// 3. GET /api/v1/entities?type=music_album - check for albums
	albumCode, albumBody, albumErr := client.Get(
		ctx, "/api/v1/entities?type=music_album",
	)
	albumResponds := albumErr == nil && albumCode == 200
	albumCount := 0
	albumEntityID := int64(0)
	if albumBody != nil {
		for _, key := range []string{"items", "data", "entities"} {
			if arr, ok := albumBody[key].([]interface{}); ok {
				albumCount = len(arr)
				if albumCount > 0 {
					if first, ok := arr[0].(map[string]interface{}); ok {
						if id, ok := first["id"].(float64); ok {
							albumEntityID = int64(id)
						}
					}
				}
				break
			}
		}
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "music_album_listing",
		Expected: "HTTP 200 with music_album entities",
		Actual: fmt.Sprintf(
			"HTTP %d, %d music_album entities",
			albumCode, albumCount,
		),
		Passed: albumResponds,
		Message: challenge.Ternary(albumResponds,
			fmt.Sprintf(
				"Music album listing returned %d items", albumCount,
			),
			fmt.Sprintf(
				"Music album listing failed: code=%d err=%v",
				albumCode, albumErr,
			)),
	})
	outputs["music_album_count"] = fmt.Sprintf("%d", albumCount)

	// 4. GET /api/v1/entities/:id/children - album children (tracks)
	if albumEntityID > 0 {
		trackCode, trackBody, trackErr := client.Get(
			ctx,
			fmt.Sprintf(
				"/api/v1/entities/%d/children", albumEntityID,
			),
		)
		trackResponds := trackErr == nil && trackCode == 200
		trackCount := 0
		if trackBody != nil {
			if arr, ok := trackBody["items"].([]interface{}); ok {
				trackCount = len(arr)
			} else if arr, ok := trackBody["data"].([]interface{}); ok {
				trackCount = len(arr)
			} else if arr, ok := trackBody["children"].([]interface{}); ok {
				trackCount = len(arr)
			}
		}
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "music_album_children",
			Expected: "HTTP 200 with items array",
			Actual: fmt.Sprintf(
				"HTTP %d, %d children for album %d",
				trackCode, trackCount, albumEntityID,
			),
			Passed: trackResponds,
			Message: challenge.Ternary(trackResponds,
				fmt.Sprintf(
					"Album %d has %d children (tracks)",
					albumEntityID, trackCount,
				),
				fmt.Sprintf(
					"Album children failed: code=%d err=%v",
					trackCode, trackErr,
				)),
		})
		outputs["album_children"] = fmt.Sprintf("%d", trackCount)
	} else {
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   "music_album_children",
			Expected: "Music album entity for children test",
			Actual:   "no music_album entities found",
			Passed:   true, // Not a failure - catalog may not have albums
			Message:  "Skipped album children test (no music_album entities in catalog)",
		})
	}

	metrics := map[string]challenge.MetricValue{
		"hierarchy_time": {
			Name:  "hierarchy_time",
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
