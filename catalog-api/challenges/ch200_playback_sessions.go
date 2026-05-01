package challenges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// postJSONAndParse wraps the httpclient.APIClient.PostJSON call
// in a convenience that marshals the request body from a map and
// parses the response body into a map so CH-200 can stay focused
// on assertions instead of plumbing.
func postJSONAndParse(
	ctx context.Context,
	client *httpclient.APIClient,
	path string,
	body map[string]interface{},
) (int, map[string]interface{}, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	status, respBody, err := client.PostJSON(ctx, path, string(raw))
	if err != nil {
		return status, nil, err
	}
	var parsed map[string]interface{}
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &parsed)
	}
	return status, parsed, nil
}

// PlaybackSessionsAPIChallenge exercises the new playback
// session tracking lifecycle: start a session, bump progress,
// end it, read back the per-entity progress summary and the
// history list. Validates the five routes added in
// commit "feat(api): playback session lifecycle + entity
// progress/history routes":
//
//	POST /api/v1/playback/sessions/start
//	POST /api/v1/playback/sessions/progress
//	POST /api/v1/playback/sessions/end
//	GET  /api/v1/entities/:id/progress
//	GET  /api/v1/entities/:id/history
type PlaybackSessionsAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPlaybackSessionsAPIChallenge creates CH-200.
func NewPlaybackSessionsAPIChallenge() *PlaybackSessionsAPIChallenge {
	return &PlaybackSessionsAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"playback-sessions-api",
			"Playback Sessions API",
			"Runs a full playback session lifecycle "+
				"(start → progress → end) against /api/v1/playback/"+
				"sessions and verifies the per-entity progress "+
				"summary and history endpoints reflect the write.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the playback sessions API challenge.
func (c *PlaybackSessionsAPIChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	c.ReportProgress("authenticating", nil)
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 5)
	if err != nil {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", err),
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), fmt.Sprintf("login failed: %v", err)))
		return challengeResult, nil
	}

	// Step 1: pick a real media item id (movie #1 is always
	// present in scanned libraries — challenge depends on
	// browsing-api-health so we know the DB has content).
	mediaItemID := 1

	// Step 2: POST /playback/sessions/start
	c.ReportProgress("playback-start", nil)
	startBody := map[string]interface{}{
		"media_item_id":  mediaItemID,
		"position_unit":  "seconds",
		"start_position": 0,
	}
	statusStart, startResp, _ := postJSONAndParse(ctx, client, "/api/v1/playback/sessions/start", startBody)

	startOK := statusStart == 200 && startResp != nil && startResp["session_id"] != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playback_start",
		Expected: "200 with session_id",
		Actual:   fmt.Sprintf("%d", statusStart),
		Passed:   startOK,
		Message:  challenge.Ternary(startOK, "Start returns session_id", "Start failed"),
	})
	if !startOK {
		challengeResult := c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			"playback start did not return a session_id",
		)
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), "playback start did not return a session_id"))
		return challengeResult, nil
	}

	sessionID := int64(0)
	switch v := startResp["session_id"].(type) {
	case float64:
		sessionID = int64(v)
	case int:
		sessionID = int64(v)
	case int64:
		sessionID = v
	}
	outputs["session_id"] = fmt.Sprintf("%d", sessionID)

	// Step 3: POST /playback/sessions/progress at 30s
	c.ReportProgress("playback-progress", nil)
	progressBody := map[string]interface{}{
		"session_id":   sessionID,
		"end_position": 30,
		"total_amount": 30,
	}
	statusProg, _, _ := postJSONAndParse(ctx, client, "/api/v1/playback/sessions/progress", progressBody)
	progOK := statusProg == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playback_progress",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusProg),
		Passed:   progOK,
		Message:  challenge.Ternary(progOK, "Progress accepted", "Progress rejected"),
	})

	// Step 4: POST /playback/sessions/end at 120s
	c.ReportProgress("playback-end", nil)
	endBody := map[string]interface{}{
		"session_id":   sessionID,
		"end_position": 120,
		"total_amount": 120,
		"completed":    false,
	}
	statusEnd, _, _ := postJSONAndParse(ctx, client, "/api/v1/playback/sessions/end", endBody)
	endOK := statusEnd == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "playback_end",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusEnd),
		Passed:   endOK,
		Message:  challenge.Ternary(endOK, "End accepted", "End rejected"),
	})

	// Step 5: GET /entities/:id/progress reflects the write
	c.ReportProgress("read-progress", nil)
	statusProgRead, progBody, _ := client.Get(ctx,
		fmt.Sprintf("/api/v1/entities/%d/progress", mediaItemID))
	progReadOK := statusProgRead == 200 && progBody != nil && progBody["progress"] != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "get_progress",
		Expected: "200 with non-null progress",
		Actual:   fmt.Sprintf("%d", statusProgRead),
		Passed:   progReadOK,
		Message:  challenge.Ternary(progReadOK, "Progress row exists", "No progress row"),
	})

	if progReadOK {
		if pm, ok := progBody["progress"].(map[string]interface{}); ok {
			lastPos, _ := pm["last_position"].(float64)
			totalReps, _ := pm["total_reproductions"].(float64)
			outputs["last_position"] = fmt.Sprintf("%.0f", lastPos)
			outputs["total_reproductions"] = fmt.Sprintf("%.0f", totalReps)

			lastPosOK := int64(lastPos) == 120
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "last_position",
				Target:   "media_progress",
				Expected: "120",
				Actual:   fmt.Sprintf("%.0f", lastPos),
				Passed:   lastPosOK,
				Message:  challenge.Ternary(lastPosOK, "last_position == 120", "last_position mismatch"),
			})

			repsOK := int64(totalReps) >= 1
			assertions = append(assertions, challenge.AssertionResult{
				Type:     "total_reproductions",
				Target:   "media_progress",
				Expected: ">=1",
				Actual:   fmt.Sprintf("%.0f", totalReps),
				Passed:   repsOK,
				Message:  challenge.Ternary(repsOK, "total_reproductions >= 1", "no reproductions counted"),
			})
		}
	}

	// Step 6: GET /entities/:id/history returns the session
	c.ReportProgress("read-history", nil)
	statusHist, histBody, _ := client.Get(ctx,
		fmt.Sprintf("/api/v1/entities/%d/history?limit=10", mediaItemID))
	histOK := statusHist == 200 && histBody != nil
	count := int64(0)
	if cnt, ok := histBody["count"].(float64); ok {
		count = int64(cnt)
	}
	histOK = histOK && count >= 1
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "history_count",
		Target:   "list_history",
		Expected: ">=1",
		Actual:   fmt.Sprintf("%d", count),
		Passed:   histOK,
		Message:  challenge.Ternary(histOK, "History contains the recorded session", "History empty"),
	})
	outputs["history_count"] = fmt.Sprintf("%d", count)

	// Roll up
	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	challengeResult := c.CreateResult(resultStatus, start, assertions, nil, outputs, "")
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), resultStatus))
	return challengeResult, nil
}
