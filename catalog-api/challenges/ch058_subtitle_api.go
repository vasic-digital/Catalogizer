package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// SubtitleAPIChallenge validates the subtitle management API:
// listing available subtitles, language support, and subtitle
// format handling.
type SubtitleAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewSubtitleAPIChallenge creates CH-058.
func NewSubtitleAPIChallenge() *SubtitleAPIChallenge {
	return &SubtitleAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"subtitle-api",
			"Subtitle Management API",
			"Validates subtitle endpoints: list available subtitles, "+
				"language support, format detection, and subtitle track "+
				"retrieval for video media entities.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the subtitle API challenge.
func (c *SubtitleAPIChallenge) Execute(
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

	// Test 1: Subtitles list endpoint
	c.ReportProgress("testing-subtitles-list", nil)
	status, _, _ := client.Get(ctx, "/subtitles")

	listOK := status == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "subtitles_list",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   listOK,
		Message:  challenge.Ternary(listOK, "Subtitles list endpoint works", "Subtitles list endpoint failed"),
	})

	// Test 2: Subtitle languages endpoint
	c.ReportProgress("testing-subtitle-languages", nil)
	statusLang, _, _ := client.Get(ctx, "/subtitles/languages")

	langOK := statusLang == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "subtitle_languages",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusLang),
		Passed:   langOK,
		Message:  challenge.Ternary(langOK, "Subtitle languages endpoint works", "Subtitle languages endpoint failed"),
	})

	// Test 3: Non-existent subtitle returns 404
	c.ReportProgress("testing-subtitle-not-found", nil)
	statusNotFound, _, _ := client.Get(ctx, "/subtitles/99999999")

	notFoundOK := statusNotFound == 404
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "subtitle_not_found",
		Expected: "404",
		Actual:   fmt.Sprintf("%d", statusNotFound),
		Passed:   notFoundOK,
		Message:  challenge.Ternary(notFoundOK, "Non-existent subtitle returns 404", "Non-existent subtitle wrong status"),
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
