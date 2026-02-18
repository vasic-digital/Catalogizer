package resolver

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// HTTPResolver fetches assets from HTTP/HTTPS URLs.
type HTTPResolver struct {
	client   *http.Client
	priority int
}

// NewHTTPResolver creates an HTTPResolver with the given priority.
func NewHTTPResolver(priority int) *HTTPResolver {
	return &HTTPResolver{
		client:   http.DefaultClient,
		priority: priority,
	}
}

// NewHTTPResolverWithClient creates an HTTPResolver with a custom HTTP client.
func NewHTTPResolverWithClient(client *http.Client, priority int) *HTTPResolver {
	return &HTTPResolver{
		client:   client,
		priority: priority,
	}
}

func (r *HTTPResolver) Name() string  { return "http" }
func (r *HTTPResolver) Priority() int { return r.priority }

// CanResolve returns true if the source hint is an HTTP(S) URL.
func (r *HTTPResolver) CanResolve(_ context.Context, req *ResolveRequest) bool {
	return strings.HasPrefix(req.SourceHint, "http://") ||
		strings.HasPrefix(req.SourceHint, "https://")
}

// Resolve fetches the asset content from the source hint URL.
func (r *HTTPResolver) Resolve(ctx context.Context, req *ResolveRequest) (*ResolveResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.SourceHint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", req.SourceHint, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch %s: HTTP %d", req.SourceHint, resp.StatusCode)
	}

	return &ResolveResult{
		Content:     resp.Body,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        resp.ContentLength,
	}, nil
}
