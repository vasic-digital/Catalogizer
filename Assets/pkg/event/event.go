package event

import (
	"digital.vasic.assets/pkg/asset"
)

// Type represents the kind of asset lifecycle event.
type Type string

const (
	AssetRequested   Type = "asset_requested"
	AssetResolving   Type = "asset_resolving"
	AssetReady       Type = "asset_ready"
	AssetFailed      Type = "asset_failed"
	AssetInvalidated Type = "asset_invalidated"
)

// Event represents an asset lifecycle event.
type Event struct {
	Type      Type
	AssetID   asset.ID
	AssetType asset.Type
	Metadata  map[string]string
}

// EventHandler is a callback for asset events.
type EventHandler func(Event)

// EventBus is the interface for publishing and subscribing to asset events.
type EventBus interface {
	Publish(event Event)
	Subscribe(handler EventHandler) (unsubscribe func())
}
