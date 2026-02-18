# API Reference

## Package `asset`

### Types
- `ID` — `string` type for asset identifiers. Generate with `NewID()`.
- `Type` — asset kind: `TypeImage`, `TypeVideoThumbnail`, `TypeAudioCover`, `TypeDocumentThumbnail`
- `Status` — lifecycle state: `StatusPending`, `StatusResolving`, `StatusReady`, `StatusFailed`, `StatusExpired`

### `Asset` struct
Core data model with ID, Type, Status, ContentType, Size, SourceHint, EntityType, EntityID, Metadata, timestamps.

### Functions
- `New(assetType, entityType, entityID) *Asset` — create pending asset
- `NewID() ID` — generate UUID-based ID
- `(a *Asset) MarkResolving()` / `MarkReady(ct, size)` / `MarkFailed()`
- `(a *Asset) IsTerminal() bool` / `IsExpired() bool`

## Package `store`

### Interface `Store`
- `Get(ctx, id) (io.ReadCloser, *Info, error)`
- `Put(ctx, id, content, info) error`
- `Delete(ctx, id) error`
- `Exists(ctx, id) (bool, error)`

### Implementations
- `NewFileStore(baseDir) (*FileStore, error)` — filesystem-backed
- `NewMemoryStore() *MemoryStore` — in-memory for testing

## Package `resolver`

### Interface `Resolver`
- `Name() string` / `Priority() int`
- `CanResolve(ctx, req) bool`
- `Resolve(ctx, req) (*ResolveResult, error)`

### Implementations
- `NewChain(resolvers...) *ChainResolver` — tries in priority order
- `NewHTTPResolver(priority) *HTTPResolver` — fetches from HTTP URLs
- `NewLocalFileResolver(priority) *LocalFileResolver` — reads local files

## Package `event`

### Interface `EventBus`
- `Publish(event)`
- `Subscribe(handler) func()` — returns unsubscribe function

### Event types
`AssetRequested`, `AssetResolving`, `AssetReady`, `AssetFailed`, `AssetInvalidated`

## Package `defaults`

### Interface `Provider`
- `GetDefault(assetType) (io.ReadCloser, *store.Info, error)`
- `Register(assetType, content, contentType)`

### Implementation
`NewEmbeddedProvider()` — serves embedded placeholder PNGs

## Package `manager`

### `Manager`
- `New(opts...) *Manager`
- `Get(ctx, id) (io.ReadCloser, *store.Info, bool, error)` — bool is isDefault
- `GetTyped(ctx, id, assetType) (io.ReadCloser, *store.Info, bool, error)`
- `Request(ctx, req) (asset.ID, error)`
- `Invalidate(ctx, id, req) error`
- `Stop()`

### Options
`WithStore(s)`, `WithResolver(r)`, `WithEventBus(bus)`, `WithDefaults(p)`, `WithWorkers(n)`, `WithLogger(w)`
