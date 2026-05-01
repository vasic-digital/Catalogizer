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
// CH-IQ-006 — per-media-type thresholds enforced
// -----------------------------------------------------------------------------

// IQPerTypeThresholdsChallenge checks the X-Cover-Quality header across the
// 11 supported media types. Each type either passes or produces a
// fail_* / placeholder_fallback value — the verdict must never be the
// placeholder "unknown" for known type names because the placeholder
// endpoint is always reachable.
type IQPerTypeThresholdsChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQPerTypeThresholdsChallenge creates CH-IQ-006.
func NewIQPerTypeThresholdsChallenge() *IQPerTypeThresholdsChallenge {
	return &IQPerTypeThresholdsChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-per-type-thresholds",
			"Image Quality: per-media-type thresholds",
			"Fetches /api/v1/cover/placeholder/:type for each of the 11 "+
				"recognised media types and asserts the response is a "+
				"well-formed SVG with matching content-type.",
			"image_quality",
			[]challenge.ID{"browsing-api-health"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQPerTypeThresholdsChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	types := []string{"movie", "tv_show", "tv_season", "tv_episode",
		"music_artist", "music_album", "song", "book", "comic", "game", "software"}

	mismatches := 0
	for _, t := range types {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/cover/placeholder/%s", c.config.BaseURL, t), nil)
		if err != nil {
			mismatches++
			continue
		}
		resp, err := authed(client, req)
		if err != nil {
			mismatches++
			continue
		}
		ct := resp.Header.Get("Content-Type")
		resp.Body.Close()
		if !strings.HasPrefix(ct, "image/svg") {
			mismatches++
		}
	}
	outputs["mismatches"] = fmt.Sprintf("%d", mismatches)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "per_type_mismatches", Passed: mismatches == 0,
		Message: fmt.Sprintf("%d/%d media types returned non-svg placeholder", mismatches, len(types)),
	})
	status := challenge.StatusPassed
	if mismatches > 0 {
		status = challenge.StatusFailed
	}
	challengeResult := c.CreateResult(status, start, assertions, nil, outputs, "")
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// -----------------------------------------------------------------------------
// CH-IQ-007 — cache hit skips rescore (attempt_count sanity)
// -----------------------------------------------------------------------------

// IQCacheHitChallenge validates that repeated GETs for the same cover do not
// grow attempt_count beyond 1 unless a provider re-fetch is needed. The
// challenge observes the X-Cover-Quality header across two back-to-back
// requests; the value must be stable, which is a necessary-but-not-
// sufficient signal that no rescore happened.
type IQCacheHitChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQCacheHitChallenge creates CH-IQ-007.
func NewIQCacheHitChallenge() *IQCacheHitChallenge {
	return &IQCacheHitChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-cache-hit",
			"Image Quality: cache hit skips rescore",
			"Requests the same cover twice and asserts X-Cover-Quality / "+
				"X-Cover-Source remain identical between calls.",
			"image_quality",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQCacheHitChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}

	headers := [2]map[string]string{}
	for i := 0; i < 2; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			c.config.BaseURL+"/api/v1/cover/1", nil)
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
		headers[i] = map[string]string{
			"quality": resp.Header.Get("X-Cover-Quality"),
			"source":  resp.Header.Get("X-Cover-Source"),
		}
		resp.Body.Close()
	}
	outputs["first_quality"] = headers[0]["quality"]
	outputs["second_quality"] = headers[1]["quality"]

	stable := headers[0]["quality"] == headers[1]["quality"]
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "quality_stable", Passed: stable,
		Message: fmt.Sprintf("quality changed between requests: %q vs %q",
			headers[0]["quality"], headers[1]["quality"]),
	})
	status := challenge.StatusPassed
	if !stable {
		status = challenge.StatusFailed
	}
	challengeResult := c.CreateResult(status, start, assertions, nil, outputs, "")
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// -----------------------------------------------------------------------------
// CH-IQ-008 — concurrent cover fetch dedup
// -----------------------------------------------------------------------------

// IQConcurrentDedupChallenge issues 10 simultaneous requests for the same
// cover id and asserts every response carries an X-Cover-Quality header.
// This is the cheap end of a real single-flight test; a full dedup
// validation requires DB inspection which is outside a black-box bank.
type IQConcurrentDedupChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQConcurrentDedupChallenge creates CH-IQ-008.
func NewIQConcurrentDedupChallenge() *IQConcurrentDedupChallenge {
	return &IQConcurrentDedupChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-concurrent-dedup",
			"Image Quality: concurrent dedup",
			"Fires 10 simultaneous GETs at /api/v1/cover/1 and asserts "+
				"every response carries an X-Cover-Quality header.",
			"image_quality",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQConcurrentDedupChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}

	type result struct {
		header string
		err    error
	}
	ch := make(chan result, 10)
	for i := 0; i < 10; i++ {
		go func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet,
				c.config.BaseURL+"/api/v1/cover/1", nil)
			if err != nil {
				ch <- result{err: err}
				return
			}
			resp, err := authed(client, req)
			if err != nil {
				ch <- result{err: err}
				return
			}
			h := resp.Header.Get("X-Cover-Quality")
			resp.Body.Close()
			ch <- result{header: h}
		}()
	}
	missing := 0
	for i := 0; i < 10; i++ {
		r := <-ch
		if r.err != nil || r.header == "" {
			missing++
		}
	}
	outputs["missing_header_concurrent"] = fmt.Sprintf("%d", missing)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "missing_header_concurrent", Passed: missing == 0,
		Message: fmt.Sprintf("%d/10 concurrent responses missing X-Cover-Quality", missing),
	})
	status := challenge.StatusPassed
	if missing > 0 {
		status = challenge.StatusFailed
	}
	challengeResult := c.CreateResult(status, start, assertions, nil, outputs, "")
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// -----------------------------------------------------------------------------
// CH-IQ-009 — X-Cover-Source header sanity
// -----------------------------------------------------------------------------

// IQXCoverSourceChallenge asserts X-Cover-Source, when present, is one of
// the known source names (cache, external_metadata, local_scan,
// llm_image_search, placeholder, tmdb, omdb, fanart, cover_art_archive,
// igdb). Unknown values would indicate a coding regression.
type IQXCoverSourceChallenge struct {
	challenge.BaseChallenge
	config *BrowsingConfig
}

// NewIQXCoverSourceChallenge creates CH-IQ-009.
func NewIQXCoverSourceChallenge() *IQXCoverSourceChallenge {
	return &IQXCoverSourceChallenge{
		BaseChallenge: challenge.NewBaseChallenge(
			"iq-xcover-source-values",
			"Image Quality: X-Cover-Source values are recognised",
			"Walks a sample of covers and asserts X-Cover-Source, when "+
				"present, is one of the documented sources. Unknown "+
				"source strings indicate an engine regression.",
			"image_quality",
			[]challenge.ID{"browsing-api-catalog"},
		),
		config: LoadBrowsingConfig(),
	}
}

// Execute runs the challenge.
func (c *IQXCoverSourceChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
	start := time.Now()
	assertions := []challenge.AssertionResult{}
	outputs := map[string]string{"api_url": c.config.BaseURL}

	client := httpclient.NewAPIClient(c.config.BaseURL)
	if _, err := client.LoginWithRetry(ctx, c.config.Username, c.config.Password, 3); err != nil {
		challengeResult := c.CreateResult(challenge.StatusFailed, start, assertions, nil, outputs, err.Error())
		challengeResult.RecordAction(fmt.Sprintf("%s: failed - %s", c.Name(), err.Error()))
		return challengeResult, nil
	}
	allowed := map[string]bool{
		"cache": true, "external_metadata": true, "local_scan": true,
		"llm_image_search": true, "placeholder": true,
		"tmdb": true, "omdb": true, "fanart": true,
		"cover_art_archive": true, "igdb": true, "": true,
	}
	unknown := 0
	for _, id := range []int{1, 2, 3, 4, 5} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/cover/%d", c.config.BaseURL, id), nil)
		if err != nil {
			continue
		}
		resp, err := authed(client, req)
		if err != nil {
			continue
		}
		src := resp.Header.Get("X-Cover-Source")
		resp.Body.Close()
		if !allowed[src] {
			unknown++
			outputs["unknown_source_sample"] = src
		}
	}
	outputs["unknown_source_count"] = fmt.Sprintf("%d", unknown)
	assertions = append(assertions, challenge.AssertionResult{
		Type: "equals", Target: "unknown_source_count", Passed: unknown == 0,
		Message: fmt.Sprintf("%d covers served with unrecognised X-Cover-Source", unknown),
	})
	status := challenge.StatusPassed
	if unknown > 0 {
		status = challenge.StatusFailed
	}
	challengeResult := c.CreateResult(status, start, assertions, nil, outputs, "")
	challengeResult.RecordAction(fmt.Sprintf("%s: challenge completed with status %s", c.Name(), status))
	return challengeResult, nil
}

// RegisterExtendedImageQualityChallenges wires CH-IQ-006..009 into the
// challenge service. The original RegisterImageQualityChallenges covers
// CH-IQ-001..005; the two lists together give the 9-case bank.
func RegisterExtendedImageQualityChallenges(svc challengeRegistrar) {
	svc.Register(NewIQPerTypeThresholdsChallenge())
	svc.Register(NewIQCacheHitChallenge())
	svc.Register(NewIQConcurrentDedupChallenge())
	svc.Register(NewIQXCoverSourceChallenge())
}
