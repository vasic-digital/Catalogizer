package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"net"
	"time"

	"digital.vasic.challenges/pkg/challenge"
)

// SMBConnectivityChallenge validates basic SMB connectivity
// to the configured endpoint: TCP dial, NTLM auth, share mount,
// and root directory listing.
type SMBConnectivityChallenge struct {
	challenge.BaseChallenge
	endpoint *Endpoint
	client   *filesystem.SmbClient
}

// NewSMBConnectivityChallenge creates CH-001.
func NewSMBConnectivityChallenge(ep *Endpoint) *SMBConnectivityChallenge {
	return &SMBConnectivityChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"first-catalog-smb-connect",
			"SMB Connectivity",
			fmt.Sprintf(
				"Validates SMB connectivity to %s:%d/%s "+
					"(TCP dial, NTLM auth, share mount, root listing)",
				ep.Host, ep.Port, ep.Share,
			),
			"integration",
			nil,
		),
		endpoint: ep,
	}
}

// Execute runs the SMB connectivity challenge.
func (c *SMBConnectivityChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{
		"endpoint_id": c.endpoint.ID,
		"share_name":  c.endpoint.Share,
	}

	// Step 1: TCP dial
	addr := fmt.Sprintf("%s:%d", c.endpoint.Host, c.endpoint.Port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "tcp_connection",
		Expected: "connection established",
		Actual:   fmt.Sprintf("dial %s", addr),
		Passed:   err == nil,
		Message:  ternary(err == nil, "TCP connection succeeded", fmt.Sprintf("TCP dial failed: %v", err)),
	})
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	conn.Close()

	// Step 2: Full SMB connection (NTLM auth + share mount)
	client, err := newSMBClient(ctx, c.endpoint)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "not_empty",
		Target:   "smb_session",
		Expected: "SMB session with NTLM auth",
		Actual:   fmt.Sprintf("connect to %s/%s", c.endpoint.Host, c.endpoint.Share),
		Passed:   err == nil,
		Message:  ternary(err == nil, "SMB session and share mount succeeded", fmt.Sprintf("SMB connection failed: %v", err)),
	})
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	c.client = client

	// Step 3: Root directory listing
	entries, err := client.ListDirectory(ctx, ".")
	entryCount := len(entries)
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "min_count",
		Target:   "root_entry_count",
		Expected: "> 0",
		Actual:   fmt.Sprintf("%d", entryCount),
		Passed:   err == nil && entryCount > 0,
		Message:  ternary(err == nil && entryCount > 0, fmt.Sprintf("Root listing returned %d entries", entryCount), fmt.Sprintf("Root listing failed: entries=%d, err=%v", entryCount, err)),
	})
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	outputs["root_entry_count"] = fmt.Sprintf("%d", entryCount)

	metrics := map[string]challenge.MetricValue{
		"connect_time": {
			Name:  "connect_time",
			Value: float64(time.Since(start).Milliseconds()),
			Unit:  "ms",
		},
		"root_entries": {
			Name:  "root_entries",
			Value: float64(entryCount),
			Unit:  "count",
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
func (c *SMBConnectivityChallenge) Cleanup(ctx context.Context) error {
	if c.client != nil {
		c.client.Disconnect(ctx)
	}
	return c.BaseChallenge.Cleanup(ctx)
}

func ternary(cond bool, t, f string) string {
	if cond {
		return t
	}
	return f
}
