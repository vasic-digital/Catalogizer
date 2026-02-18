# User Guide

## Integration

### 1. Add as submodule

```bash
git submodule add git@github.com:vasic-digital/Assets.git Assets
```

### 2. Add replace directive

In your `go.mod`:
```go
replace digital.vasic.assets => ../Assets
```

### 3. Initialize manager

```go
store, _ := store.NewFileStore("./cache/assets")
bus := event.NewInMemoryBus()
chain := resolver.NewChain(
    resolver.NewHTTPResolver(1),
    resolver.NewLocalFileResolver(2),
)

mgr := manager.New(
    manager.WithStore(store),
    manager.WithResolver(chain),
    manager.WithEventBus(bus),
    manager.WithWorkers(4),
)
defer mgr.Stop()
```

### 4. Serve assets

```go
// In your HTTP handler:
content, info, isDefault, err := mgr.Get(ctx, assetID)
if err != nil {
    http.Error(w, "not found", 404)
    return
}
defer content.Close()

w.Header().Set("Content-Type", info.ContentType)
if isDefault {
    w.Header().Set("X-Asset-Status", "pending")
} else {
    w.Header().Set("X-Asset-Status", "ready")
}
io.Copy(w, content)
```

## Custom Resolvers

Implement the `resolver.Resolver` interface:

```go
type MyResolver struct{}

func (r *MyResolver) Name() string     { return "my_resolver" }
func (r *MyResolver) Priority() int    { return 5 }
func (r *MyResolver) CanResolve(ctx context.Context, req *resolver.ResolveRequest) bool {
    return req.Metadata["source"] == "my_service"
}
func (r *MyResolver) Resolve(ctx context.Context, req *resolver.ResolveRequest) (*resolver.ResolveResult, error) {
    // Your resolution logic here
}
```

## Event Handling

```go
bus.Subscribe(func(evt event.Event) {
    switch evt.Type {
    case event.AssetReady:
        log.Printf("Asset %s resolved", evt.AssetID)
    case event.AssetFailed:
        log.Printf("Asset %s failed", evt.AssetID)
    }
})
```
