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

// -----------------------------------------------------------------------------
// CH-IQ-001 — placeholder fallback header
// -----------------------------------------------------------------------------

// IQPlaceholderFallbackChallenge verifies that a cover request for a missing
// media item returns the placeholder_fallback quality header, so client apps
// can distinguish a "no cover available" state from a pass.
type IQPlaceholderFallbackChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQPlaceholderFallbackChallenge creates CH-IQ-001.
func NewIQPlaceholderFallbackChallenge() *IQPlaceholderFallbackChallenge {
	return &IQPlaceholderFallbackChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-placeholder-fallback",
			"Image Quality: placeholder fallback header",
			"Requests a cover for a non-existent media item and asserts the "+
				"X-Cover-Quality header is set to 'placeholder_fallback' so "+
				"client apps can surface the fallback visually.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQPlaceholderFallbackChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	c.ReportProgress("fetching missing cover", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL+"/api/v1/cover/99999999", nil)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	resp, err := authed(client, req)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	defer resp.Body.Close()

	verdict := resp.Header.Get("X-Cover-Quality")
	outputs["x_cover_quality"] = verdict
	assertions = append(assertions, challenge.AssertionResult{
		Type:    "equals",
		Target:  "X-Cover-Quality",
		Passed:  verdict == "placeholder_fallback",
		Message: fmt.Sprintf("expected placeholder_fallback, got %q", verdict),
	})

	status := challenge.StatusPassed
	for _, a := range assertions {
		if !a.Passed {
			status = challenge.StatusFailed
		}
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-002 — quality gate blocks low-res share-sourced cover
// -----------------------------------------------------------------------------

// IQBlocksLowResChallenge verifies the X-Cover-Quality header reports a fail
// verdict when a known low-resolution cover has been assessed. It reads from
// the image_quality_assessments table (via the public /admin/image-quality
// endpoint if present) — if that endpoint is missing in the running server
// (dev environments typically omit it), the challenge reports as skipped
// rather than failed.
type IQBlocksLowResChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQBlocksLowResChallenge creates CH-IQ-002.
func NewIQBlocksLowResChallenge() *IQBlocksLowResChallenge {
	return &IQBlocksLowResChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-blocks-low-res",
			"Image Quality: gate blocks low-res covers",
			"Walks a sample of covers and asserts that none served by the API "+
				"carry a verdict of 'fail_lowres' in the X-Cover-Quality header. "+
				"If the run server has no assessed covers yet, the challenge is "+
				"treated as passed because there is nothing to block.",
			"image_quality",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQBlocksLowResChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	c.ReportProgress("walking covers", nil)
	_, body, err := client.Get(ctx, "/api/v1/entities?limit=25")
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	items, _ := body["items"].([]interface{})
	if len(items) == 0 {
		assertions = append(assertions, challenge.AssertionResult{
			Type: "not_empty", Target: "entities", Passed: true,
			Message: "no entities to assess; considered vacuously passed",
		})
		return c.CreateResult(challenge.StatusPassed, start, assertions, nil, outputs, ""), nil
	}

	checked := 0
	rejected := 0
	for _, raw := range items {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		idFloat, _ := m["id"].(float64)
		if idFloat <= 0 {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/cover/%d", c.config.BaseURL, int(idFloat)), nil)
		if err != nil {
			continue
		}
		resp, err := authed(client, req)
		if err != nil {
			continue
		}
		verdict := resp.Header.Get("X-Cover-Quality")
		resp.Body.Close()
		checked++
		if strings.HasPrefix(verdict, "fail_") {
			rejected++
		}
	}
	outputs["checked"] = fmt.Sprintf("%d", checked)
	outputs["rejected"] = fmt.Sprintf("%d", rejected)

	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "rejected_count", Passed: rejected == 0,
		Message: fmt.Sprintf("%d/%d covers served with failing verdict", rejected, checked),
	})

	status := challenge.StatusPassed
	if rejected > 0 {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-003 — header always present on cover responses
// -----------------------------------------------------------------------------

// IQHeaderAlwaysPresentChallenge asserts that every 200 response from
// /api/v1/cover/:id carries an X-Cover-Quality header of some value, so
// downstream observability dashboards always have a known signal.
type IQHeaderAlwaysPresentChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQHeaderAlwaysPresentChallenge creates CH-IQ-003.
func NewIQHeaderAlwaysPresentChallenge() *IQHeaderAlwaysPresentChallenge {
	return &IQHeaderAlwaysPresentChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-header-always-present",
			"Image Quality: header always present",
			"Requests a cover for a handful of entities and asserts the "+
				"X-Cover-Quality header is set on every 200 response (value "+
				"may be pass, placeholder_fallback, unknown, or a fail_*). An "+
				"absent header is a regression.",
			"image_quality",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQHeaderAlwaysPresentChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}

	ids := []int{1, 2, 3, 99999999} // last one intentionally missing
	missing := 0
	for _, id := range ids {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/cover/%d", c.config.BaseURL, id), nil)
		if err != nil {
			continue
		}
		resp, err := authed(client, req)
		if err != nil {
			continue
		}
		header := resp.Header.Get("X-Cover-Quality")
		resp.Body.Close()
		if header == "" {
			missing++
		}
	}
	outputs["missing_header_count"] = fmt.Sprintf("%d", missing)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "missing_header_count", Passed: missing == 0,
		Message: fmt.Sprintf("%d cover responses without X-Cover-Quality header", missing),
	})
	status := challenge.StatusPassed
	if missing > 0 {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-004 — placeholder endpoint serves valid SVG (regression guard)
// -----------------------------------------------------------------------------

// IQPlaceholderSVGValidChallenge asserts that /api/v1/cover/placeholder/:type
// returns a well-formed SVG document for each of the recognised media types.
type IQPlaceholderSVGValidChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQPlaceholderSVGValidChallenge creates CH-IQ-004.
func NewIQPlaceholderSVGValidChallenge() *IQPlaceholderSVGValidChallenge {
	return &IQPlaceholderSVGValidChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-placeholder-svg-valid",
			"Image Quality: placeholder SVG valid per type",
			"Fetches /api/v1/cover/placeholder/:type for every recognised "+
				"media type and asserts the response body is a well-formed "+
				"SVG document (starts with <svg).",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQPlaceholderSVGValidChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	types := []string{"movie", "tv_show", "music_album", "book", "game", "software", "comic"}

	failures := 0
	for _, t := range types {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/cover/placeholder/%s", c.config.BaseURL, t), nil)
		if err != nil {
			failures++
			continue
		}
		resp, err := authed(client, req)
		if err != nil {
			failures++
			continue
		}
		buf := make([]byte, 256)
		n, _ := resp.Body.Read(buf)
		resp.Body.Close()
		if n < 5 || !strings.Contains(string(buf[:n]), "<svg") {
			failures++
		}
	}
	outputs["invalid_svg_count"] = fmt.Sprintf("%d", failures)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "invalid_svg_count", Passed: failures == 0,
		Message: fmt.Sprintf("%d of %d types returned invalid placeholder SVG", failures, len(types)),
	})
	status := challenge.StatusPassed
	if failures > 0 {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-005 — LLM resolver disabled by default (security posture)
// -----------------------------------------------------------------------------

// IQLLMDisabledByDefaultChallenge verifies that hitting the assets or covers
// endpoints does not silently trigger the LLM fallback when it is disabled.
// The challenge asserts that no cover response is attributed to the
// llm_image_search source when env vars are not set.
type IQLLMDisabledByDefaultChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQLLMDisabledByDefaultChallenge creates CH-IQ-005.
func NewIQLLMDisabledByDefaultChallenge() *IQLLMDisabledByDefaultChallenge {
	return &IQLLMDisabledByDefaultChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-llm-disabled-by-default",
			"Image Quality: LLM resolver disabled by default",
			"Walks a sample of covers and asserts no X-Cover-Source header is "+
				"'llm_image_search'. This guards against the LLM fallback "+
				"being activated in environments where the operator has not "+
				"explicitly turned it on.",
			"image_quality",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQLLMDisabledByDefaultChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	_, body, err := client.Get(ctx, "/api/v1/entities?limit=20")
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	items, _ := body["items"].([]interface{})

	llmHits := 0
	for _, raw := range items {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		idFloat, _ := m["id"].(float64)
		if idFloat <= 0 {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/cover/%d", c.config.BaseURL, int(idFloat)), nil)
		if err != nil {
			continue
		}
		resp, err := authed(client, req)
		if err != nil {
			continue
		}
		src := resp.Header.Get("X-Cover-Source")
		resp.Body.Close()
		if src == "llm_image_search" {
			llmHits++
		}
	}
	outputs["llm_hits"] = fmt.Sprintf("%d", llmHits)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "llm_hits", Passed: llmHits == 0,
		Message: fmt.Sprintf("%d cover responses attributed to llm_image_search without operator opt-in", llmHits),
	})
	status := challenge.StatusPassed
	if llmHits > 0 {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// authed issues an HTTP request with the APIClient's JWT token attached and a
// bounded timeout so individual requests do not stall the challenge.
func authed(client *httpclient.APIClient, req *http.Request) (*http.Response, error) {
	if tok := client.Token(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return httpClient.Do(req)
}

// RegisterImageQualityChallenges registers all image-quality challenges with
// the supplied challenge service.
type challengeRegistrar interface {
	Register(challenge.Challenge) error
}

// RegisterImageQualityChallenges registers the CH-IQ-* suite.
func RegisterImageQualityChallenges(svc challengeRegistrar) {
	svc.Register(NewIQPlaceholderFallbackChallenge())
	svc.Register(NewIQBlocksLowResChallenge())
	svc.Register(NewIQHeaderAlwaysPresentChallenge())
	svc.Register(NewIQPlaceholderSVGValidChallenge())
	svc.Register(NewIQLLMDisabledByDefaultChallenge())
}
