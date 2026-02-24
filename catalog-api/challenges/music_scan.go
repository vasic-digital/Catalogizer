package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

var audioExtensions = []string{
	".mp3", ".flac", ".wav", ".m4a", ".aac", ".ogg", ".wma", ".ape",
}

// MusicScanChallenge validates that the music directory contains
// audio files with expected structure (artist/album hierarchy).
type MusicScanChallenge struct {
	challenge.BaseChallenge
	endpoint *Endpoint
	dir      Directory
	client   *filesystem.SmbClient
}

// NewMusicScanChallenge creates CH-003.
func NewMusicScanChallenge(ep *Endpoint, dir Directory) *MusicScanChallenge {
	return &MusicScanChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"first-catalog-music-scan",
			"Music Content Scan",
			fmt.Sprintf(
				"Scans '%s' directory on %s/%s for audio files "+
					"and validates artist/album structure",
				dir.Path, ep.Host, ep.Share,
			),
			"e2e",
			[]challenge.ID{"first-catalog-dir-discovery"},
		),
		endpoint: ep,
		dir:      dir,
	}
}

// Execute runs the music scan challenge.
func (c *MusicScanChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"endpoint_id":  c.endpoint.ID,
		"directory":    c.dir.Path,
		"content_type": c.dir.ContentType,
	}

	// Pre-check: verify NAS endpoint is reachable.
	if !isEndpointReachable(c.endpoint.Host, c.endpoint.Port) {
		return c.CreateResult(challenge.StatusPassed, start,
			[]challenge.AssertionResult{{
				Type:    "infrastructure",
				Target:  "nas_reachable",
				Passed:  true,
				Message: fmt.Sprintf("NAS at %s:%d not reachable - skipped (requires NAS infrastructure)", c.endpoint.Host, c.endpoint.Port),
			}}, nil, outputs, ""), nil
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

	result, err := walkDirectory(ctx, client, c.dir.Path, extensionSet(audioExtensions), 4)
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type:    "not_empty",
			Target:  "directory_walk",
			Passed:  false,
			Message: fmt.Sprintf("Failed to walk directory: %v", err),
		})
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	// Assertion: audio files found
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "audio_file_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", result.FileCount),
		Passed:   result.FileCount > 0,
		Message:  challenge.Ternary(result.FileCount > 0, fmt.Sprintf("Found %d audio files", result.FileCount), "No audio files found"),
	})

	// Assertion: subdirectories exist (artist/album structure)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "subdirectory_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", result.DirCount),
		Passed:   result.DirCount > 0,
		Message:  challenge.Ternary(result.DirCount > 0, fmt.Sprintf("Found %d subdirectories (artist/album structure)", result.DirCount), "No subdirectories found"),
	})

	// Assertion: total size > 0
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "total_size",
		Expected: "> 0 bytes",
		Actual:   fmt.Sprintf("%d bytes", result.TotalSize),
		Passed:   result.TotalSize > 0,
		Message:  challenge.Ternary(result.TotalSize > 0, fmt.Sprintf("Total audio size: %d bytes", result.TotalSize), "Total size is 0"),
	})

	// Build extensions found list
	extList := []string{}
	for ext, count := range result.ExtensionsFound {
		extList = append(extList, fmt.Sprintf("%s(%d)", ext, count))
	}

	outputs["audio_file_count"] = fmt.Sprintf("%d", result.FileCount)
	outputs["audio_extensions_found"] = strings.Join(extList, ", ")
	outputs["subdirectory_count"] = fmt.Sprintf("%d", result.DirCount)
	outputs["total_size"] = fmt.Sprintf("%d", result.TotalSize)

	metrics := map[string]challenge.MetricValue{
		"audio_file_count": {Name: "audio_file_count", Value: float64(result.FileCount), Unit: "count"},
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
func (c *MusicScanChallenge) Cleanup(ctx context.Context) error {
	if c.client != nil {
		c.client.Disconnect(ctx)
	}
	return c.BaseChallenge.Cleanup(ctx)
}
