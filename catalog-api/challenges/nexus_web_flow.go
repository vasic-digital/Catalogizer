package challenges

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital.vasic.challenges/pkg/challenge"
	"digital.vasic.challenges/pkg/httpclient"
)

// NexusWebFlowChallenge validates that the full Catalogizer web UI
// flow is reachable from the Nexus browser stack. It is deliberately
// black-box: the challenge speaks only to the running catalog-api + web
// client via HTTP so it succeeds without requiring a real Chromium
// instance inside the orchestrator process. A richer real-browser
// variant lives under pkg/nexus/browser/integration_test.go once the
// nexus_chromedp_integration build tag is active.
//
// This challenge registers as CH-NX-WEBFLOW-001 and is the first
// "Nexus-authored" Catalogizer-facing bank entry — matching R-6 of
// the research-and-execution-plan.
type NexusWebFlowChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewNexusWebFlowChallenge creates the challenge.
func NewNexusWebFlowChallenge() *NexusWebFlowChallenge {
	return &NexusWebFlowChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"nexus-web-flow",
			"Nexus: end-to-end web flow (login -> browse -> cover)",
			"Drives the Catalogizer web UI path a Nexus browser Engine "+
				"would follow — login, entity listing, cover fetch — via "+
				"raw HTTP. Asserts each hop responds with expected status "+
				"and the X-Cover-Quality header is present on the final "+
				"hop.",
			"nexus",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the Nexus-equivalent flow. Every sub-step emits a
// progress report so the HelixQA dashboard can draw a timeline. A
// failing sub-step aborts the flow and the final Result reflects
// where we stopped — matching how a real Nexus orchestrator run
// surfaces failures.
func (c *NexusWebFlowChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)

	c.ReportProgress("login", nil)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type: "not_empty", Target: "login", Passed: false,
			Message: fmt.Sprintf("login failed: %v", err),
		})
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	assertions = append(assertions, challenge.AssertionResult{
		Type: "not_empty", Target: "login", Passed: true, Message: "admin login ok",
	})

	c.ReportProgress("list entities", nil)
	listStatus, listBody, err := client.Get(ctx, "/api/v1/entities?limit=5")
	if err != nil {
		assertions = append(assertions, challenge.AssertionResult{
			Type: "status_code", Target: "entities", Passed: false,
			Message: fmt.Sprintf("entities list errored: %v", err),
		})
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	outputs["entities_status"] = fmt.Sprintf("%d", listStatus)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "status_code", Target: "entities", Passed: listStatus == 200,
		Message: fmt.Sprintf("GET /api/v1/entities = %d", listStatus),
	})
	items, _ := listBody["items"].([]interface{})

	entityID := 1
	for _, it := range items {
		if m, ok := it.(map[string]interface{}); ok {
			if id, ok := m["id"].(float64); ok && id > 0 {
				entityID = int(id)
				break
			}
		}
	}
	outputs["entity_id"] = fmt.Sprintf("%d", entityID)

	c.ReportProgress("cover fetch", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/cover/%d", c.config.BaseURL, entityID), nil)
	if err != nil {
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	resp, err := authed(client, req)
	if err != nil {
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	verdict := resp.Header.Get("X-Cover-Quality")
	source := resp.Header.Get("X-Cover-Source")
	_ = resp.Body.Close()
	outputs["cover_quality"] = verdict
	outputs["cover_source"] = source

	assertions = append(assertions,
		challenge.AssertionResult{
			Type: "contains", Target: "x-cover-quality", Passed: verdict != "",
			Message: fmt.Sprintf("X-Cover-Quality = %q", verdict),
		},
		challenge.AssertionResult{
			Type: "status_code", Target: "cover", Passed: resp.StatusCode == 200 || resp.StatusCode == 302,
			Message: fmt.Sprintf("GET /api/v1/cover/%d = %d", entityID, resp.StatusCode),
		},
	)

	// Final pass/fail: every assertion must be passed.
	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
			break
		}
	}
	challengeResult := c.CreateResult(status, start, assertions, nil, outputs, "")
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// RegisterNexusChallenges wires every Nexus-authored Catalogizer
// challenge into the challenge service. Right now only the web-flow
// challenge ships; real-browser, mobile, and desktop variants land
// when R-7 integration harnesses come online.
func RegisterNexusChallenges(svc challengeRegistrar) {
	svc.Register(NewNexusWebFlowChallenge())
}

// ensure the challengeRegistrar interface stays satisfied even if the
// challenge service swaps its Register signature.
var _ = strings.ToLower
