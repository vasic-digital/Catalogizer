package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"digital.vasic.assets/pkg/resolver"

	"github.com/odwrtw/fanarttv"
)

// FanartTVResolver resolves movie / tv poster art from Fanart.tv. The
// resolver is environment-driven: it reads FANART_TV_API_KEY and
// remains inert when the key is unset, so fresh installs do not need
// Fanart.tv credentials to run.
//
// Priority 11 — tried after ExternalMetadataResolver (2) but before
// OMDB and OpenLibrary, matching the design spec's "standard providers
// first, LLM last" order.
type FanartTVResolver struct {
	apiKey   string
	priority int
	client   *http.Client
	fanart   *fanarttv.Client
}

// NewFanartTVResolver returns a resolver that activates only when the
// FANART_TV_API_KEY env var is set. The client timeout is bounded so a
// slow upstream cannot stall the asset chain.
func NewFanartTVResolver(priority int) *FanartTVResolver {
	key := strings.TrimSpace(envFirstNonEmpty("FANART_TV_API_KEY"))
	var client *fanarttv.Client
	if key != "" {
		client = fanarttv.New(key)
	}
	return &FanartTVResolver{
		apiKey:   key,
		priority: priority,
		client:   &http.Client{Timeout: 10 * time.Second},
		fanart:   client,
	}
}

// Name returns the resolver identifier.
func (r *FanartTVResolver) Name() string { return "fanarttv" }

// Priority returns the chain priority.
func (r *FanartTVResolver) Priority() int { return r.priority }

// CanResolve reports whether the resolver is configured and the
// request carries the platform-appropriate id. The Fanart.tv v3 API
// requires an IMDB id (tt-prefixed) for movies and a TheTVDB numeric
// id for TV content — TMDB ids are not accepted by either endpoint.
// Fixes B1 from docs/nexus/remaining-work.md.
func (r *FanartTVResolver) CanResolve(_ context.Context, req *resolver.ResolveRequest) bool {
	if r.apiKey == "" || r.fanart == nil {
		return false
	}
	if req == nil || req.Metadata == nil {
		return false
	}
	switch strings.ToLower(req.EntityType) {
	case "movie", "movies":
		return req.Metadata["imdb_id"] != ""
	case "tv_show", "tv_season", "tv_episode":
		return req.Metadata["tvdb_id"] != ""
	}
	return false
}

// Resolve queries Fanart.tv and returns the first high-res poster.
func (r *FanartTVResolver) Resolve(ctx context.Context, req *resolver.ResolveRequest) (*resolver.ResolveResult, error) {
	if r.apiKey == "" || r.fanart == nil {
		return nil, errors.New("fanarttv: resolver not configured")
	}
	id := r.providerID(req)
	if id == "" {
		return nil, fmt.Errorf("fanarttv: no platform-appropriate id in metadata for %s (movies need imdb_id, TV needs tvdb_id)", req.EntityType)
	}
	posters, err := r.lookupPosters(ctx, req.EntityType, id)
	if err != nil {
		return nil, fmt.Errorf("fanarttv lookup: %w", err)
	}
	if len(posters) == 0 {
		return nil, errors.New("fanarttv: no posters returned")
	}
	return r.download(ctx, posters[0])
}

// providerID picks the correct id for the entity type. Movies must
// supply IMDB ids (imdb_id / "tt"-prefixed); TV content must supply
// TheTVDB numeric ids (tvdb_id). TMDB ids cannot be used with the
// Fanart.tv v3 API — see docs/nexus/remaining-work.md B1.
func (r *FanartTVResolver) providerID(req *resolver.ResolveRequest) string {
	if req == nil || req.Metadata == nil {
		return ""
	}
	switch strings.ToLower(req.EntityType) {
	case "movie", "movies":
		return req.Metadata["imdb_id"]
	case "tv_show", "tv_season", "tv_episode":
		return req.Metadata["tvdb_id"]
	}
	return ""
}

// lookupPosters chooses the correct Fanart.tv endpoint based on entity
// type. Movies and TV shows have dedicated methods on the client.
func (r *FanartTVResolver) lookupPosters(ctx context.Context, entityType, id string) ([]string, error) {
	_ = ctx
	switch strings.ToLower(entityType) {
	case "movie", "movies":
		art, err := r.fanart.GetMovieImages(id)
		if err != nil {
			return nil, err
		}
		return imageInfoURLs(art.Posters), nil
	case "tv_show", "tv_season", "tv_episode":
		art, err := r.fanart.GetShowImages(id)
		if err != nil {
			return nil, err
		}
		return imageInfoURLs(art.Posters), nil
	default:
		return nil, fmt.Errorf("fanarttv: entity type %q not supported", entityType)
	}
}

func (r *FanartTVResolver) download(ctx context.Context, url string) (*resolver.ResolveResult, error) {
	if err := allowPublicURLLocal(url); err != nil {
		return nil, fmt.Errorf("fanarttv: unsafe URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fanarttv download: HTTP %d", resp.StatusCode)
	}
	return &resolver.ResolveResult{
		Content:     resp.Body,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
	}, nil
}

// imageInfoURLs extracts non-empty URLs from a Fanart.tv ImageInfo
// slice in their original (most-liked-first) order.
func imageInfoURLs(entries []*fanarttv.ImageInfo) []string {
	out := make([]string, 0, len(entries))
	for _, p := range entries {
		if p == nil || p.URL == "" {
			continue
		}
		out = append(out, p.URL)
	}
	return out
}

// envFirstNonEmpty returns the first non-empty value of the listed
// environment variables. It lives in this file because every new
// resolver reuses the same fallback pattern.
func envFirstNonEmpty(names ...string) string {
	for _, n := range names {
		if v := strings.TrimSpace(osGetenv(n)); v != "" {
			return v
		}
	}
	return ""
}

var osGetenv = defaultOsGetenv

// allowPublicURLLocal is a small SSRF guard local to the provider
// resolvers. It mirrors the one in the LLM resolver so provider
// URLs never reach a loopback / private range (a compromised upstream
// should not be able to pivot through our fetch path).
func allowPublicURLLocal(raw string) error {
	lower := strings.ToLower(raw)
	for _, bad := range []string{"file:", "javascript:", "data:", "vbscript:"} {
		if strings.HasPrefix(lower, bad) {
			return fmt.Errorf("unsupported scheme in %q", raw)
		}
	}
	if strings.Contains(lower, "://127.0.0.1") ||
		strings.Contains(lower, "://localhost") ||
		strings.Contains(lower, "://10.") ||
		strings.Contains(lower, "://192.168.") ||
		strings.Contains(lower, "://169.254.") {
		return fmt.Errorf("unsafe private host in %q", raw)
	}
	return nil
}

// noopCloser is retained for older callers that expected a ReadCloser
// from raw bytes. The Fanart.tv resolver streams the body directly.
type noopCloser struct{ io.Reader }

func (noopCloser) Close() error { return nil }

// defaultOsGetenv shells out to os.Getenv by default. Tests override
// osGetenv to inject deterministic values.
func defaultOsGetenv(k string) string {
	return getenv(k)
}
