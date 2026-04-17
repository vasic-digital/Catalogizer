package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"digital.vasic.assets/pkg/resolver"

	"github.com/Henry-Sarabia/igdb"
)

// IGDBResolver resolves game cover art from the Internet Game Database
// via Twitch's IGDB API. The resolver activates when IGDB_CLIENT_ID
// and IGDB_APP_ACCESS_TOKEN env vars are set.
//
// Priority 22 — after the existing providers but before the last-resort
// LLM resolver.
type IGDBResolver struct {
	clientID string
	token    string
	priority int
	http     *http.Client
	client   *igdb.Client
}

// NewIGDBResolver returns a resolver bound to Twitch credentials. If
// either env var is missing the resolver remains inert.
func NewIGDBResolver(priority int) *IGDBResolver {
	id := strings.TrimSpace(envFirstNonEmpty("IGDB_CLIENT_ID"))
	tok := strings.TrimSpace(envFirstNonEmpty("IGDB_APP_ACCESS_TOKEN", "IGDB_BEARER_TOKEN"))
	var client *igdb.Client
	if id != "" && tok != "" {
		client = igdb.NewClient(tok, &http.Client{Timeout: 10 * time.Second})
	}
	return &IGDBResolver{
		clientID: id,
		token:    tok,
		priority: priority,
		http:     &http.Client{Timeout: 10 * time.Second},
		client:   client,
	}
}

// Name returns the resolver identifier.
func (r *IGDBResolver) Name() string { return "igdb" }

// Priority returns the chain priority.
func (r *IGDBResolver) Priority() int { return r.priority }

// CanResolve reports whether the resolver is configured + the request
// carries a game / software entity.
func (r *IGDBResolver) CanResolve(_ context.Context, req *resolver.ResolveRequest) bool {
	if r.client == nil {
		return false
	}
	if req == nil {
		return false
	}
	switch strings.ToLower(req.EntityType) {
	case "game", "software":
		return true
	}
	return false
}

// Resolve queries IGDB for covers. It prefers an igdb_id stored in
// Metadata; otherwise it searches by title.
func (r *IGDBResolver) Resolve(ctx context.Context, req *resolver.ResolveRequest) (*resolver.ResolveResult, error) {
	if r.client == nil {
		return nil, errors.New("igdb: resolver not configured")
	}
	id, err := r.resolveGameID(ctx, req)
	if err != nil {
		return nil, err
	}
	covers, err := r.client.Covers.Index(
		igdb.SetFilter("game", igdb.OpEquals, strconv.Itoa(id)),
		igdb.SetLimit(1),
	)
	if err != nil {
		return nil, fmt.Errorf("igdb cover lookup: %w", err)
	}
	if len(covers) == 0 || covers[0].URL == "" {
		return nil, errors.New("igdb: no covers")
	}
	url := normalizeIGDBImageURL(covers[0].URL)
	if err := allowPublicURLLocal(url); err != nil {
		return nil, fmt.Errorf("igdb: unsafe URL: %w", err)
	}
	return r.download(ctx, url)
}

func (r *IGDBResolver) resolveGameID(ctx context.Context, req *resolver.ResolveRequest) (int, error) {
	_ = ctx
	if raw := req.Metadata["igdb_id"]; raw != "" {
		id, err := strconv.Atoi(raw)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	title := req.Metadata["title"]
	if title == "" {
		return 0, errors.New("igdb: need igdb_id or title metadata")
	}
	results, err := r.client.Games.Search(title, igdb.SetLimit(1))
	if err != nil {
		return 0, fmt.Errorf("igdb search: %w", err)
	}
	if len(results) == 0 {
		return 0, errors.New("igdb: no matches")
	}
	return results[0].ID, nil
}

// normalizeIGDBImageURL turns the API's "//images.igdb.com/..."
// protocol-relative URL with t_thumb into a full https://...t_cover_big
// URL for higher-resolution cover art.
func normalizeIGDBImageURL(raw string) string {
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	raw = strings.Replace(raw, "t_thumb", "t_cover_big", 1)
	return raw
}

func (r *IGDBResolver) download(ctx context.Context, url string) (*resolver.ResolveResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("igdb download: HTTP %d", resp.StatusCode)
	}
	return &resolver.ResolveResult{
		Content:     resp.Body,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
	}, nil
}
