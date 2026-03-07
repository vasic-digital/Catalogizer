package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// SyncAPIChallenge validates the synchronization API endpoints:
// sync status, sync history, device registration, and conflict
// detection mechanisms.
type SyncAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSyncAPIChallenge creates CH-057.
func NewSyncAPIChallenge() *SyncAPIChallenge {
	return &SyncAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"sync-api",
			"Synchronization API",
			"Validates sync endpoints: sync status, sync history, "+
				"device registration, and conflict detection. Ensures "+
				"multi-device synchronization infrastructure is functional.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the sync API challenge.
func (c *SyncAPIChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login
	c.ReportProgress("authenticating", nil)
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 5)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", err),
		), nil
	}

	// Test 1: Sync status endpoint
	c.ReportProgress("testing-sync-status", nil)
	status, body, _ := client.Get(ctx, "/sync/status")

	statusOK := status == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_status",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   statusOK,
		Message:  challenge.Ternary(statusOK, "Sync status endpoint works", "Sync status endpoint failed"),
	})

	// Test 2: Sync history endpoint
	c.ReportProgress("testing-sync-history", nil)
	statusHistory, _, _ := client.Get(ctx, "/sync/history")

	historyOK := statusHistory == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_history",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusHistory),
		Passed:   historyOK,
		Message:  challenge.Ternary(historyOK, "Sync history endpoint works", "Sync history endpoint failed"),
	})

	// Test 3: Sync devices endpoint
	c.ReportProgress("testing-sync-devices", nil)
	statusDevices, _, _ := client.Get(ctx, "/sync/devices")

	devicesOK := statusDevices == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_devices",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusDevices),
		Passed:   devicesOK,
		Message:  challenge.Ternary(devicesOK, "Sync devices endpoint works", "Sync devices endpoint failed"),
	})

	// Test 4: Sync conflicts endpoint
	c.ReportProgress("testing-sync-conflicts", nil)
	statusConflicts, _, _ := client.Get(ctx, "/sync/conflicts")

	conflictsOK := statusConflicts == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "sync_conflicts",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusConflicts),
		Passed:   conflictsOK,
		Message:  challenge.Ternary(conflictsOK, "Sync conflicts endpoint works", "Sync conflicts endpoint failed"),
	})

	resultStatus := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			resultStatus = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(resultStatus, start, assertions, nil, outputs, ""), nil
}
