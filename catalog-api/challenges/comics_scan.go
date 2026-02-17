package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

var comicExtensions = []string{
	".cbr", ".cbz", ".cb7", ".cbt", ".pdf", ".epub",
}

// ComicsScanChallenge validates that the comics directory
// contains comic book files.
type ComicsScanChallenge struct {
	challenge.BaseChallenge
	endpoint *Endpoint
	dir      Directory
	client   *filesystem.SmbClient
}

// NewComicsScanChallenge creates CH-007.
func NewComicsScanChallenge(ep *Endpoint, dir Directory) *ComicsScanChallenge {
	return &ComicsScanChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"first-catalog-comics-scan",
			"Comics Content Scan",
			fmt.Sprintf(
				"Scans '%s' directory on %s/%s for comic book files",
				dir.Path, ep.Host, ep.Share,
			),
			"e2e",
			[]challenge.ID{"first-catalog-dir-discovery"},
		),
		endpoint: ep,
		dir:      dir,
	}
}

// Execute runs the comics scan challenge.
func (c *ComicsScanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
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

	result, err := walkDirectory(ctx, client, c.dir.Path, extensionSet(comicExtensions), 3)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "directory_walk",
			Passed:  false,
			Message: fmt.Sprintf("Failed to walk directory: %v", err),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	// Assertion: comic files found
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "comic_file_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", result.FileCount),
		Passed:   result.FileCount > 0,
		Message:  ternary(result.FileCount > 0, fmt.Sprintf("Found %d comic files", result.FileCount), "No comic files found"),
	})

	// Assertion: total size > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "total_size",
		Expected: "> 0 bytes",
		Actual:   fmt.Sprintf("%d bytes", result.TotalSize),
		Passed:   result.TotalSize > 0,
		Message:  ternary(result.TotalSize > 0, fmt.Sprintf("Total size: %d bytes", result.TotalSize), "Total size is 0"),
	})

	extList := []string{}
	for ext, count := range result.ExtensionsFound {
		extList = append(extList, fmt.Sprintf("%s(%d)", ext, count))
	}

	outputs["comic_file_count"] = fmt.Sprintf("%d", result.FileCount)
	outputs["extensions_found"] = strings.Join(extList, ", ")
	outputs["subdirectory_count"] = fmt.Sprintf("%d", result.DirCount)
	outputs["total_size"] = fmt.Sprintf("%d", result.TotalSize)

	metrics := map[string]challenge.MetricValue{
		"comic_file_count": {Name: "comic_file_count", Value: float64(result.FileCount), Unit: "count"},
		"subdirectory_count": {Name: "subdirectory_count", Value: float64(result.DirCount), Unit: "count"},
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
func (c *ComicsScanChallenge) Cleanup(ctx context.Context) error {
	if c.client != nil {
		c.client.Disconnect(ctx)
	}
	return c.BaseChallenge.Cleanup(ctx)
}
