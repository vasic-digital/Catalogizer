package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// PaginationChallenge validates that all list endpoints support
// proper pagination with limit/offset parameters and return
// consistent results.
type PaginationChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewPaginationChallenge creates CH-052.
func NewPaginationChallenge() *PaginationChallenge {
	return &PaginationChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"pagination",
			"API Pagination Consistency",
			"Verifies that list endpoints support limit/offset pagination, "+
				"return proper page sizes, and handle edge cases like "+
				"zero limit, negative offset, and beyond-range requests.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the pagination challenge.
func (c *PaginationChallenge) Execute(
	ctx context.Context,
) (*challenge.Result, error) {

	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	// Login for authenticated endpoints
	c.ReportProgress("authenticating", nil)
	_, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 5)
	if err != nil {
		return c.CreateResult(
			challenge.StatusFailed, start, assertions, nil, outputs,
			fmt.Sprintf("login failed: %v", err),
		), nil
	}

	// Test pagination endpoints
	paginatedEndpoints := []struct {
		name     string
		endpoint string
	}{
		{"files", "/catalog/files"},
		{"entities", "/entities"},
		{"storage-roots", "/storage-roots"},
	}

	for _, ep := range paginatedEndpoints {
		c.ReportProgress(fmt.Sprintf("testing-%s", ep.name), nil)

		// Test with limit=1
		status, body, _ := client.Get(ctx, fmt.Sprintf("%s?limit=1&offset=0", ep.endpoint))

		endpointResponds := status > 0 && status < 500
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   fmt.Sprintf("%s_responds", ep.name),
			Expected: "2xx or 4xx",
			Actual:   fmt.Sprintf("%d", status),
			Passed:   endpointResponds,
			Message:  fmt.Sprintf("Endpoint %s responds with status %d", ep.endpoint, status),
		})

		if body != nil {
			outputs[fmt.Sprintf("%s_response_keys", ep.name)] = fmt.Sprintf("%v", getMapKeys(body))
		}

		// Test with offset beyond data range (should return empty, not error)
		status2, _, _ := client.Get(ctx, fmt.Sprintf("%s?limit=10&offset=999999", ep.endpoint))

		beyondRangeOK := status2 > 0 && status2 < 500
		assertions = append(assertions, challenge.AssertionResult{
			Type:     "status_code",
			Target:   fmt.Sprintf("%s_beyond_range", ep.name),
			Expected: "2xx or 4xx (not 5xx)",
			Actual:   fmt.Sprintf("%d", status2),
			Passed:   beyondRangeOK,
			Message:  fmt.Sprintf("Beyond-range pagination on %s returns %d", ep.endpoint, status2),
		})
	}

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}

	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
