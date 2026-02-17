package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

var softwareExtensions = []string{
	".exe", ".msi", ".dmg", ".pkg", ".iso", ".img",
	".deb", ".rpm", ".apk", ".appimage",
	".zip", ".rar", ".7z", ".tar.gz",
}

// SoftwareScanChallenge validates that the software/installations
// directory contains installer and archive files.
type SoftwareScanChallenge struct {
	challenge.BaseChallenge
	endpoint *Endpoint
	dir      Directory
	client   *filesystem.SmbClient
}

// NewSoftwareScanChallenge creates CH-006.
func NewSoftwareScanChallenge(ep *Endpoint, dir Directory) *SoftwareScanChallenge {
	return &SoftwareScanChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"first-catalog-software-scan",
			"Software Content Scan",
			fmt.Sprintf(
				"Scans '%s' directory on %s/%s for installer "+
					"and archive files",
				dir.Path, ep.Host, ep.Share,
			),
			"e2e",
			[]challenge.ID{"first-catalog-dir-discovery"},
		),
		endpoint: ep,
		dir:      dir,
	}
}

// Execute runs the software scan challenge.
func (c *SoftwareScanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
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

	result, err := walkDirectory(ctx, client, c.dir.Path, extensionSet(softwareExtensions), 3)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "directory_walk",
			Passed:  false,
			Message: fmt.Sprintf("Failed to walk directory: %v", err),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	// Assertion: installer/archive files found
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "installer_file_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", result.FileCount),
		Passed:   result.FileCount > 0,
		Message:  ternary(result.FileCount > 0, fmt.Sprintf("Found %d installer/archive files", result.FileCount), "No installer/archive files found"),
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

	outputs["installer_file_count"] = fmt.Sprintf("%d", result.FileCount)
	outputs["extensions_found"] = strings.Join(extList, ", ")
	outputs["subdirectory_count"] = fmt.Sprintf("%d", result.DirCount)
	outputs["total_size"] = fmt.Sprintf("%d", result.TotalSize)

	metrics := map[string]challenge.MetricValue{
		"installer_file_count": {Name: "installer_file_count", Value: float64(result.FileCount), Unit: "count"},
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
func (c *SoftwareScanChallenge) Cleanup(ctx context.Context) error {
	if c.client != nil {
		c.client.Disconnect(ctx)
	}
	return c.BaseChallenge.Cleanup(ctx)
}
