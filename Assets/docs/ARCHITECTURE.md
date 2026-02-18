# Architecture

## Data Flow

```
Client Request
    │
    ▼
Manager.Get(id)
    │
    ├─ Store.Exists(id)?
    │   ├─ Yes → Store.Get(id) → return content (isDefault=false)
    │   └─ No  → Defaults.GetDefault(type) → return placeholder (isDefault=true)
    │
Manager.Request(req)
    │
    ▼
EventBus.Publish(AssetRequested)
    │
    ▼
WorkerPool.submit(req)
    │
    ▼
Worker.process()
    ├─ EventBus.Publish(AssetResolving)
    ├─ Resolver.Resolve(req)
    │   ├─ ChainResolver tries each in priority order
    │   ├─ First CanResolve() + Resolve() success wins
    │   └─ Returns content + content_type
    ├─ Store.Put(id, content)
    └─ EventBus.Publish(AssetReady or AssetFailed)
```

## Interface Contracts

### Store
- Thread-safe for concurrent access
- Get returns `io.ReadCloser` — caller must close
- Put accepts `io.Reader` — store reads to completion
- Delete is idempotent — no error for missing assets

### Resolver
- CanResolve is a fast check — no I/O if possible
- Resolve returns `io.ReadCloser` — caller must close
- Priority: lower number = tried first
- Must respect context cancellation

### EventBus
- Publish is synchronous — handlers run in the publisher's goroutine
- Subscribe returns unsubscribe function
- Thread-safe for concurrent publish/subscribe

## Design Decisions

1. **No database dependency** — the module stores only bytes. Asset metadata tracking is the consumer's responsibility.
2. **Minimal dependencies** — only `google/uuid` for ID generation. No logging framework (accepts `io.Writer`).
3. **Embedded defaults** — placeholder images are compiled into the binary via `//go:embed`. No runtime file dependencies.
4. **Worker pool** — configurable goroutine count prevents unbounded parallelism during bulk resolution.
