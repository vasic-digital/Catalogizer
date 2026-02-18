package resolver

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// LocalFileResolver reads assets from the local filesystem.
type LocalFileResolver struct {
	priority int
}

// NewLocalFileResolver creates a LocalFileResolver with the given priority.
func NewLocalFileResolver(priority int) *LocalFileResolver {
	return &LocalFileResolver{priority: priority}
}

func (r *LocalFileResolver) Name() string  { return "local" }
func (r *LocalFileResolver) Priority() int { return r.priority }

// CanResolve returns true if the source hint is a local file path
// starting with "/" and the file exists.
func (r *LocalFileResolver) CanResolve(_ context.Context, req *ResolveRequest) bool {
	if !strings.HasPrefix(req.SourceHint, "/") {
		return false
	}
	info, err := os.Stat(req.SourceHint)
	return err == nil && !info.IsDir()
}

// Resolve reads the asset content from the local filesystem.
func (r *LocalFileResolver) Resolve(_ context.Context, req *ResolveRequest) (*ResolveResult, error) {
	info, err := os.Stat(req.SourceHint)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", req.SourceHint, err)
	}

	f, err := os.Open(req.SourceHint)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", req.SourceHint, err)
	}

	ct := mime.TypeByExtension(filepath.Ext(req.SourceHint))
	if ct == "" {
		ct = "application/octet-stream"
	}

	return &ResolveResult{
		Content:     f,
		ContentType: ct,
		Size:        info.Size(),
	}, nil
}
