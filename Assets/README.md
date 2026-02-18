# Assets

Universal lazy asset loading framework for Go. Provides strategy-based resolution, filesystem storage, event-driven lifecycle notifications, and embedded default placeholders.

## Quick Start

```go
import (
    "digital.vasic.assets/pkg/asset"
    "digital.vasic.assets/pkg/defaults"
    "digital.vasic.assets/pkg/event"
    "digital.vasic.assets/pkg/manager"
    "digital.vasic.assets/pkg/resolver"
    "digital.vasic.assets/pkg/store"
)

// Create components
fileStore, _ := store.NewFileStore("./cache/assets")
eventBus := event.NewInMemoryBus()
httpResolver := resolver.NewHTTPResolver(1)
chain := resolver.NewChain(httpResolver)

// Create manager
mgr := manager.New(
    manager.WithStore(fileStore),
    manager.WithResolver(chain),
    manager.WithEventBus(eventBus),
    manager.WithDefaults(defaults.NewEmbeddedProvider()),
    manager.WithWorkers(4),
)
defer mgr.Stop()

// Request asset resolution
id, _ := mgr.Request(ctx, &resolver.ResolveRequest{
    AssetType:  asset.TypeImage,
    SourceHint: "https://example.com/cover.jpg",
    EntityType: "media_item",
    EntityID:   "42",
})

// Get asset (returns default placeholder if not yet resolved)
content, info, isDefault, _ := mgr.Get(ctx, id)
```

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
```

## Packages

| Package | Purpose |
|---------|---------|
| `pkg/asset` | Core types: Asset, ID, Type, Status |
| `pkg/store` | Storage interface + FileStore + MemoryStore |
| `pkg/resolver` | Strategy pattern: Resolver interface, ChainResolver, HTTPResolver, LocalFileResolver |
| `pkg/event` | Event bus for asset lifecycle notifications |
| `pkg/defaults` | Default/fallback content with embedded placeholder images |
| `pkg/manager` | Orchestrator: resolve + store + events + worker pool |
