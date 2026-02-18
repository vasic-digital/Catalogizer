package manager

import (
	"context"
	"fmt"
	"io"

	"digital.vasic.assets/pkg/asset"
	"digital.vasic.assets/pkg/defaults"
	"digital.vasic.assets/pkg/event"
	"digital.vasic.assets/pkg/resolver"
	"digital.vasic.assets/pkg/store"
)

// Manager orchestrates asset resolution, storage, and event notification.
type Manager struct {
	store       store.Store
	resolver    resolver.Resolver
	eventBus    event.EventBus
	defaults    defaults.Provider
	pool        *workerPool
	workerCount int
	logger      io.Writer
}

// New creates a new Manager with the given options.
func New(opts ...Option) *Manager {
	m := &Manager{
		workerCount: 2,
	}
	for _, opt := range opts {
		opt(m)
	}

	if m.defaults == nil {
		m.defaults = defaults.NewEmbeddedProvider()
	}

	if m.store != nil && m.resolver != nil {
		m.pool = newWorkerPool(m.workerCount, m.resolver, m.store, m.eventBus, m.logger)
	}

	return m
}

// Get returns asset content. If the asset is not yet resolved, it returns
// the default placeholder and isDefault=true. The third return value
// indicates whether the returned content is a default placeholder.
func (m *Manager) Get(ctx context.Context, id asset.ID) (io.ReadCloser, *store.Info, bool, error) {
	if m.store != nil {
		exists, err := m.store.Exists(ctx, id)
		if err == nil && exists {
			content, info, err := m.store.Get(ctx, id)
			if err == nil {
				return content, info, false, nil
			}
		}
	}

	if m.defaults != nil {
		content, info, err := m.defaults.GetDefault(asset.TypeImage)
		if err == nil {
			return content, info, true, nil
		}
	}

	return nil, nil, false, fmt.Errorf("asset not found and no default available: %s", id)
}

// GetTyped returns asset content with type-aware default selection.
func (m *Manager) GetTyped(ctx context.Context, id asset.ID, assetType asset.Type) (io.ReadCloser, *store.Info, bool, error) {
	if m.store != nil {
		exists, err := m.store.Exists(ctx, id)
		if err == nil && exists {
			content, info, err := m.store.Get(ctx, id)
			if err == nil {
				return content, info, false, nil
			}
		}
	}

	if m.defaults != nil {
		content, info, err := m.defaults.GetDefault(assetType)
		if err == nil {
			return content, info, true, nil
		}
	}

	return nil, nil, false, fmt.Errorf("asset not found and no default available: %s", id)
}

// Request registers a new asset for background resolution and returns
// the assigned asset ID.
func (m *Manager) Request(ctx context.Context, req *resolver.ResolveRequest) (asset.ID, error) {
	if req.AssetID == "" {
		req.AssetID = asset.NewID()
	}

	if m.eventBus != nil {
		m.eventBus.Publish(event.Event{
			Type:      event.AssetRequested,
			AssetID:   req.AssetID,
			AssetType: req.AssetType,
			Metadata: map[string]string{
				"entity_type": req.EntityType,
				"entity_id":   req.EntityID,
				"source_hint": req.SourceHint,
			},
		})
	}

	if m.pool != nil {
		m.pool.submit(workItem{request: req})
	}

	return req.AssetID, nil
}

// Invalidate marks an asset for re-resolution by removing the stored
// content and queuing a new resolution request.
func (m *Manager) Invalidate(ctx context.Context, id asset.ID, req *resolver.ResolveRequest) error {
	if m.store != nil {
		if err := m.store.Delete(ctx, id); err != nil {
			return fmt.Errorf("delete stored asset: %w", err)
		}
	}

	if m.eventBus != nil {
		m.eventBus.Publish(event.Event{
			Type:    event.AssetInvalidated,
			AssetID: id,
		})
	}

	if req != nil && m.pool != nil {
		req.AssetID = id
		m.pool.submit(workItem{request: req})
	}

	return nil
}

// Stop shuts down the background worker pool.
func (m *Manager) Stop() {
	if m.pool != nil {
		m.pool.stop()
	}
}
