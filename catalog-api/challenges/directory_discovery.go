package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// DirectoryDiscoveryChallenge validates that all configured
// Cyrillic-named directories exist and are accessible on the
// SMB share.
type DirectoryDiscoveryChallenge struct {
	challenge.BaseChallenge
	endpoint *Endpoint
	client   *filesystem.SmbClient
}

// NewDirectoryDiscoveryChallenge creates CH-002.
func NewDirectoryDiscoveryChallenge(ep *Endpoint) *DirectoryDiscoveryChallenge {
	return &DirectoryDiscoveryChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"first-catalog-dir-discovery",
			"Directory Discovery",
			fmt.Sprintf(
				"Validates that all %d configured directories exist "+
					"and are accessible on %s/%s",
				len(ep.Directories), ep.Host, ep.Share,
			),
			"integration",
			[]challenge.ID{"first-catalog-smb-connect"},
		),
		endpoint: ep,
	}
}

// Execute runs the directory discovery challenge.
func (c *DirectoryDiscoveryChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"endpoint_id":     c.endpoint.ID,
		"directory_count": fmt.Sprintf("%d", len(c.endpoint.Directories)),
	}

	client, err := newSMBClient(ctx, c.endpoint)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "smb_connection",
			Passed:  false,
			Message: fmt.Sprintf("SMB connection failed: %v", err),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	c.client = client

	totalEntries := 0
	for _, dir := range c.endpoint.Directories {
		// Check directory exists
		info, err := client.GetFileInfo(ctx, dir.Path)
		dirExists := err == nil && info != nil && info.IsDir
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "not_empty",
			Target:   fmt.Sprintf("dir_exists_%s", dir.ContentType),
			Expected: fmt.Sprintf("directory '%s' exists", dir.Path),
			Actual:   fmt.Sprintf("exists=%v", dirExists),
			Passed:   dirExists,
			Message:  challenge.Ternary(dirExists, fmt.Sprintf("Directory '%s' exists", dir.Path), fmt.Sprintf("Directory '%s' not found: %v", dir.Path, err)),
		})

		if !dirExists {
			continue
		}

		// Check directory is readable and non-empty
		entries, err := client.ListDirectory(ctx, dir.Path)
		entryCount := len(entries)
		readable := err == nil && entryCount > 0
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "min_count",
			Target:   fmt.Sprintf("dir_entries_%s", dir.ContentType),
			Expected: "> 0",
			Actual:   fmt.Sprintf("%d", entryCount),
			Passed:   readable,
			Message:  challenge.Ternary(readable, fmt.Sprintf("Directory '%s' has %d entries", dir.Path, entryCount), fmt.Sprintf("Directory '%s' empty or unreadable: entries=%d, err=%v", dir.Path, entryCount, err)),
		})

		outputs[fmt.Sprintf("%s_entry_count", dir.ContentType)] = fmt.Sprintf("%d", entryCount)
		totalEntries += entryCount
	}

	outputs["total_entries"] = fmt.Sprintf("%d", totalEntries)

	metrics := map[string]challenge.MetricValue{
		"total_entries": {
			Name:  "total_entries",
			Value: float64(totalEntries),
			Unit:  "count",
		},
		"discovery_time": {
			Name:  "discovery_time",
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

	return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}

// Cleanup disconnects the SMB client.
func (c *DirectoryDiscoveryChallenge) Cleanup(ctx context.Context) error {
	if c.client != nil {
		c.client.Disconnect(ctx)
	}
	return c.BaseChallenge.Cleanup(ctx)
}
