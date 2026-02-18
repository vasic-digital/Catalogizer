package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

var videoExtensions = []string{
	".mkv", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".m4v", ".ts",
}

// SeriesScanChallenge validates that the TV series directory
// contains video files with show/season structure.
type SeriesScanChallenge struct {
	challenge.BaseChallenge
	endpoint *Endpoint
	dir      Directory
	client   *filesystem.SmbClient
}

// NewSeriesScanChallenge creates CH-004.
func NewSeriesScanChallenge(ep *Endpoint, dir Directory) *SeriesScanChallenge {
	return &SeriesScanChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"first-catalog-series-scan",
			"TV Series Content Scan",
			fmt.Sprintf(
				"Scans '%s' directory on %s/%s for video files "+
					"and validates show/season structure",
				dir.Path, ep.Host, ep.Share,
			),
			"e2e",
			[]challenge.ID{"first-catalog-dir-discovery"},
		),
		endpoint: ep,
		dir:      dir,
	}
}

// Execute runs the series scan challenge.
func (c *SeriesScanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"endpoint_id":  c.endpoint.ID,
		"directory":    c.dir.Path,
		"content_type": c.dir.ContentType,
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

	result, err := walkDirectory(ctx, client, c.dir.Path, extensionSet(videoExtensions), 3)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "directory_walk",
			Passed:  false,
			Message: fmt.Sprintf("Failed to walk directory: %v", err),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	// Assertion: video files found
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "video_file_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", result.FileCount),
		Passed:   result.FileCount > 0,
		Message:  challenge.Ternary(result.FileCount > 0, fmt.Sprintf("Found %d video files", result.FileCount), "No video files found"),
	})

	// Assertion: subdirectories exist (show/season structure)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "show_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", result.DirCount),
		Passed:   result.DirCount > 0,
		Message:  challenge.Ternary(result.DirCount > 0, fmt.Sprintf("Found %d subdirectories (show/season structure)", result.DirCount), "No subdirectories found"),
	})

	// Assertion: total size > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "total_size",
		Expected: "> 0 bytes",
		Actual:   fmt.Sprintf("%d bytes", result.TotalSize),
		Passed:   result.TotalSize > 0,
		Message:  challenge.Ternary(result.TotalSize > 0, fmt.Sprintf("Total video size: %d bytes", result.TotalSize), "Total size is 0"),
	})

	extList := []string{}
	for ext, count := range result.ExtensionsFound {
		extList = append(extList, fmt.Sprintf("%s(%d)", ext, count))
	}

	outputs["video_file_count"] = fmt.Sprintf("%d", result.FileCount)
	outputs["video_extensions_found"] = strings.Join(extList, ", ")
	outputs["show_count"] = fmt.Sprintf("%d", result.DirCount)
	outputs["total_size"] = fmt.Sprintf("%d", result.TotalSize)

	metrics := map[string]challenge.MetricValue{
		"video_file_count": {Name: "video_file_count", Value: float64(result.FileCount), Unit: "count"},
		"show_count": {Name: "show_count", Value: float64(result.DirCount), Unit: "count"},
		"total_size": {Name: "total_size", Value: float64(result.TotalSize), Unit: "bytes"},
		"scan_time": {Name: "scan_time", Value: float64(time.Since(start).Milliseconds()), Unit: "ms"},
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
func (c *SeriesScanChallenge) Cleanup(ctx context.Context) error {
	if c.client != nil {
		c.client.Disconnect(ctx)
	}
	return c.BaseChallenge.Cleanup(ctx)
}
