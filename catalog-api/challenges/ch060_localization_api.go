package challenges

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// LocalizationAPIChallenge validates the localization and
// internationalization API: supported languages, translation
// retrieval, and locale management.
type LocalizationAPIChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewLocalizationAPIChallenge creates CH-060.
func NewLocalizationAPIChallenge() *LocalizationAPIChallenge {
	return &LocalizationAPIChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"localization-api",
			"Localization and i18n API",
			"Validates localization endpoints: supported languages, "+
				"translation retrieval, locale preferences, and "+
				"localization statistics.",
			"api",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the localization API challenge.
func (c *LocalizationAPIChallenge) Execute(
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

	// Test 1: Supported languages
	c.ReportProgress("testing-languages", nil)
	status, body, _ := client.Get(ctx, "/localization/languages")

	langOK := status == 200 && body != nil
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "supported_languages",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", status),
		Passed:   langOK,
		Message:  challenge.Ternary(langOK, "Languages endpoint works", "Languages endpoint failed"),
	})

	// Test 2: Translations for a locale
	c.ReportProgress("testing-translations", nil)
	statusTrans, _, _ := client.Get(ctx, "/localization/translations?locale=en")

	transOK := statusTrans == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "translations",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusTrans),
		Passed:   transOK,
		Message:  challenge.Ternary(transOK, "Translations endpoint works", "Translations endpoint failed"),
	})

	// Test 3: Localization stats
	c.ReportProgress("testing-localization-stats", nil)
	statusStats, _, _ := client.Get(ctx, "/localization/stats")

	statsOK := statusStats == 200
	assertions = append(assertions, challenge.AssertionResult{
		Type:     "status_code",
		Target:   "localization_stats",
		Expected: "200",
		Actual:   fmt.Sprintf("%d", statusStats),
		Passed:   statsOK,
		Message:  challenge.Ternary(statsOK, "Localization stats works", "Localization stats failed"),
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
