package resolver

import (
	"context"
	"io"

	"digital.vasic.assets/pkg/asset"
)

// ResolveRequest contains the information needed to resolve an asset.
type ResolveRequest struct {
	AssetID    asset.ID
	AssetType  asset.Type
	SourceHint string
	EntityType string
	EntityID   string
	Metadata   map[string]string
}

// ResolveResult contains the resolved asset content.
type ResolveResult struct {
	Content     io.ReadCloser
	ContentType string
	Size        int64
}

// Resolver is the strategy interface for asset resolution.
type Resolver interface {
	Name() string
	CanResolve(ctx context.Context, req *ResolveRequest) bool
	Resolve(ctx context.Context, req *ResolveRequest) (*ResolveResult, error)
	Priority() int
}
