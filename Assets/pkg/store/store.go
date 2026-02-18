package store

import (
	"context"
	"io"

	"digital.vasic.assets/pkg/asset"
)

// Info holds metadata about stored asset content.
type Info struct {
	ContentType string
	Size        int64
}

// Store is the interface for asset content storage.
type Store interface {
	Get(ctx context.Context, id asset.ID) (io.ReadCloser, *Info, error)
	Put(ctx context.Context, id asset.ID, content io.Reader, info *Info) error
	Delete(ctx context.Context, id asset.ID) error
	Exists(ctx context.Context, id asset.ID) (bool, error)
}
