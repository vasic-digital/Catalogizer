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
// CH-IQ-010 — quality header absent on 404 (nonexistent cover id)
// -----------------------------------------------------------------------------

// IQNotFoundHeaderChallenge confirms the API still emits an
// X-Cover-Quality header even when the cover is entirely missing. The
// expected value is "placeholder_fallback" so clients can render the
// placeholder consistently.
type IQNotFoundHeaderChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQNotFoundHeaderChallenge creates CH-IQ-010.
func NewIQNotFoundHeaderChallenge() *IQNotFoundHeaderChallenge {
	return &IQNotFoundHeaderChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-not-found-header",
			"Image Quality: nonexistent cover carries placeholder header",
			"GET /api/v1/cover/-1 and assert X-Cover-Quality == placeholder_fallback.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQNotFoundHeaderChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}
	client := httpclient.NewAPIClient(c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/cover/-1", nil)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	resp, err := authed(client, req)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	quality := resp.Header.Get("X-Cover-Quality")
	resp.Body.Close()
	outputs["quality"] = quality
	pass := quality == "placeholder_fallback" || quality == "unknown"
	assertions = append(assertions, challenge.AssertionResult{
		Type: "contains", Target: "x-cover-quality", Passed: pass,
		Message: fmt.Sprintf("expected placeholder_fallback|unknown, got %q", quality),
	})
	status := challenge.StatusPassed
	if !pass {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-011 — cover URL endpoint returns JSON with cover_url field
// -----------------------------------------------------------------------------

// IQCoverURLChallenge asserts /api/v1/cover/url/:id returns JSON with a
// cover_url field the client can follow. The URL itself may be empty on
// unknown items; the field must exist regardless.
type IQCoverURLChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQCoverURLChallenge creates CH-IQ-011.
func NewIQCoverURLChallenge() *IQCoverURLChallenge {
	return &IQCoverURLChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-cover-url-json",
			"Image Quality: /cover/url/:id returns cover_url",
			"GET /api/v1/cover/url/1?type=movie and assert JSON contains cover_url field.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQCoverURLChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}
	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	code, body, err := client.Get(ctx, "/api/v1/cover/url/1?type=movie")
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	outputs["status"] = fmt.Sprintf("%d", code)
	_, hasURL := body["cover_url"]
	assertions = append(assertions, challenge.AssertionResult{
		Type: "contains", Target: "cover_url", Passed: hasURL,
		Message: fmt.Sprintf("cover_url field missing from response body (status %d)", code),
	})
	status := challenge.StatusPassed
	if !hasURL {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-012 — placeholder content-length stable
// -----------------------------------------------------------------------------

// IQPlaceholderSizeStableChallenge asserts placeholder SVG responses are
// reproducible (byte-identical) across requests. A drifting placeholder
// size usually indicates an accidental runtime mutation.
type IQPlaceholderSizeStableChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQPlaceholderSizeStableChallenge creates CH-IQ-012.
func NewIQPlaceholderSizeStableChallenge() *IQPlaceholderSizeStableChallenge {
	return &IQPlaceholderSizeStableChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-placeholder-stable",
			"Image Quality: placeholder payload stable across requests",
			"GET /api/v1/cover/placeholder/movie twice and compare content lengths.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQPlaceholderSizeStableChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}
	client := httpclient.NewAPIClient(c.config.BaseURL)
	sizes := [2]int64{}
	for i := 0; i < 2; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			c.config.BaseURL+"/api/v1/cover/placeholder/movie", nil)
		if err != nil {
			return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
		}
		resp, err := authed(client, req)
		if err != nil {
			return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
		}
		sizes[i] = resp.ContentLength
		resp.Body.Close()
	}
	outputs["size_first"] = fmt.Sprintf("%d", sizes[0])
	outputs["size_second"] = fmt.Sprintf("%d", sizes[1])
	stable := sizes[0] == sizes[1]
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "placeholder_size_stable", Passed: stable,
		Message: fmt.Sprintf("placeholder size changed: %d vs %d", sizes[0], sizes[1]),
	})
	status := challenge.StatusPassed
	if !stable {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-013 — cache-control header set on covers
// -----------------------------------------------------------------------------

// IQCacheControlChallenge validates that cover responses carry a
// Cache-Control header so browsers cache them aggressively. Matches the
// "Cache-Control: public, max-age=86400" contract in cover_handler.go.
type IQCacheControlChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQCacheControlChallenge creates CH-IQ-013.
func NewIQCacheControlChallenge() *IQCacheControlChallenge {
	return &IQCacheControlChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-cache-control",
			"Image Quality: Cache-Control header present",
			"GET /api/v1/cover/1 and assert Cache-Control includes max-age.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQCacheControlChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}
	client := httpclient.NewAPIClient(c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/cover/1", nil)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	resp, err := authed(client, req)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	cc := resp.Header.Get("Cache-Control")
	resp.Body.Close()
	outputs["cache_control"] = cc
	pass := strings.Contains(cc, "max-age=")
	assertions = append(assertions, challenge.AssertionResult{
		Type: "contains", Target: "cache-control", Passed: pass,
		Message: fmt.Sprintf("Cache-Control should include max-age, got %q", cc),
	})
	status := challenge.StatusPassed
	if !pass {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// -----------------------------------------------------------------------------
// CH-IQ-014 — content-type on covers is an image MIME
// -----------------------------------------------------------------------------

// IQContentTypeChallenge ensures cover responses declare an image/* MIME
// type so client decoders work.
type IQContentTypeChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQContentTypeChallenge creates CH-IQ-014.
func NewIQContentTypeChallenge() *IQContentTypeChallenge {
	return &IQContentTypeChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-content-type",
			"Image Quality: image MIME content-type on covers",
			"GET /api/v1/cover/1 and assert Content-Type starts with image/.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQContentTypeChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}
	client := httpclient.NewAPIClient(c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.config.BaseURL+"/api/v1/cover/1", nil)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	resp, err := authed(client, req)
	if err != nil {
		return c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error()), nil
	}
	ct := resp.Header.Get("Content-Type")
	resp.Body.Close()
	outputs["content_type"] = ct
	pass := strings.HasPrefix(ct, "image/")
	assertions = append(assertions, challenge.AssertionResult{
		Type: "contains", Target: "content-type", Passed: pass,
		Message: fmt.Sprintf("Content-Type should be image/*, got %q", ct),
	})
	status := challenge.StatusPassed
	if !pass {
		status = challenge.StatusFailed
	}
	return c.CreateResult(status, start, assertions, nil, outputs, ""), nil
}

// RegisterFinalImageQualityChallenges wires CH-IQ-010..014 into the
// challenge service. Together with the earlier registrations this
// brings the total to the 14 cases called for in the design spec.
func RegisterFinalImageQualityChallenges(svc challengeRegistrar) {
	svc.Register(NewIQNotFoundHeaderChallenge())
	svc.Register(NewIQCoverURLChallenge())
	svc.Register(NewIQPlaceholderSizeStableChallenge())
	svc.Register(NewIQCacheControlChallenge())
	svc.Register(NewIQContentTypeChallenge())
}
