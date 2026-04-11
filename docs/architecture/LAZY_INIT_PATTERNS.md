# Lazy Initialization Patterns

This document is the canonical reference for when a catalog-api service should defer expensive construction work to first use, and when it should eagerly fail at boot. It was written alongside `scripts/audit-lazy-init.sh` (Phase 6) which flags constructors with eager signals (DB open, network dial, goroutine spawn, FS scan, timer start).

## Eager vs Lazy: the decision rule

**Eager** is correct when:
1. The dependency is on every request's hot path (auth, DB pool, metrics middleware). If it's broken, the server is broken — fail at boot so the operator sees the problem immediately.
2. The construction cost is small enough to not bloat boot time measurably.
3. Failure modes are deterministic (network config wrong vs transient).

**Lazy** is correct when:
1. The dependency is optional or per-feature (metadata providers TMDB/MusicBrainz/IGDB that may or may not be configured).
2. The dependency is flaky (external API with rate limits, transient timeouts).
3. The construction cost is non-trivial (parsing large files, scanning disk trees).
4. Most requests don't touch the service — deferring hides construction latency from the hot path.

## Audit triage (2026-04-11)

`scripts/audit-lazy-init.sh` flagged 4 eager candidates in catalog-api. After review all four are **intentional eager** and stay as-is:

| Constructor | File | Signal | Verdict | Reason |
|---|---|---|---|---|
| `NewConnection` | `database/connection.go` | db/net-open | **Stay eager** | The DB is the primary dependency. Every request touches it. Boot-time failure is what the operator needs. |
| `NewWebSocketHandlerWithConfig` | `handlers/websocket_handler.go` | timer-start | **Stay eager** | Cleanup ticker + single goroutine. The handler has `Stop()` + `sync.Once` + `wg.Wait()` already. Deferring would complicate shutdown ordering without a real benefit. |
| `NewSmbClient` | `smb/client.go` | db/net-open | **Stay eager** | Callers explicitly ask for a reachable SMB session. If they don't want one they shouldn't call this constructor. Bounded 5s `DialTimeout` already prevents hangs. |
| `NewThrottler` | `utils/concurrency.go` | timer-start | **Stay eager** | Throttlers are singletons created once at startup. |

For each case, the audit's heuristic is valuable but the human decision is "eager is right here".

## When you DO need lazy: the `digital.vasic.lazy` primitives

The project pulls in the `digital.vasic.lazy` submodule via `replace` in `go.mod`. It provides two primitives:

### `lazy.Value[T]` — deferred single value

```go
import "digital.vasic.lazy/pkg/lazy"

type MetadataClient struct {
    tmdb lazy.Value[*tmdbClient]
}

func NewMetadataClient(apiKey string) *MetadataClient {
    return &MetadataClient{
        tmdb: lazy.NewValue(func() (*tmdbClient, error) {
            if apiKey == "" {
                return nil, ErrTMDBNotConfigured
            }
            return newTMDBClient(apiKey)
        }),
    }
}

func (c *MetadataClient) Lookup(id string) (*Movie, error) {
    client, err := c.tmdb.Get()
    if err != nil {
        return nil, fmt.Errorf("tmdb client: %w", err)
    }
    return client.Lookup(id)
}
```

Guarantees:
- Loader runs exactly once, even under concurrent `Get()`.
- Subsequent `Get()` calls return the cached value (or the cached error — if loading failed, it stays failed until `Reset()`).
- `Reset()` invalidates the cached value for deliberate re-initialization.

### `lazy.Service[T]` — deferred service

Same semantics as `Value[T]` but exposes `Initialized() bool` so callers can check without triggering construction. Useful when tests want to assert a service has been touched.

## When you DO need bounded parallelism: `internal/concurrency.Semaphore`

Separately from lazy-init, the audit flagged unbounded `go func()` fan-out inside range loops. The standard answer is:

```go
import "catalogizer/internal/concurrency"

sem := concurrency.NewSemaphore(8) // 8 concurrent workers max
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(it Item) {
        defer wg.Done()
        if err := sem.Acquire(ctx); err != nil {
            return
        }
        defer sem.Release()
        processItem(it)
    }(item)
}
wg.Wait()
```

The semaphore caps concurrency regardless of how many items are queued. Context cancellation propagates: `sem.Acquire(ctx)` returns early if the context is done.

### Real-world example — `ProviderManager.Search`

`internal/media/providers/providers.go` already uses this pattern:

```go
resultsCh := make(chan searchResult, len(targets))
var wg sync.WaitGroup
for _, tgt := range targets {
    wg.Add(1)
    go func(t target) {
        defer wg.Done()
        if err := pm.searchSem.Acquire(ctx); err != nil {
            return
        }
        defer pm.searchSem.Release()
        providerResults, err := t.provider.Search(ctx, query, mediaType, year)
        resultsCh <- searchResult{name: t.name, results: providerResults, err: err}
    }(tgt)
}
```

The semaphore lives on the `ProviderManager` struct, so it's shared across all concurrent `Search()` calls — each one contributes to the same concurrency budget.

## Checklist for new service constructors

Before merging a new `New*()`:

1. **Is the dependency critical at boot?** Yes → eager.
2. **Does construction touch the network or disk?** Consider lazy.
3. **Can construction fail transiently?** Consider lazy + retry semantics.
4. **Does the constructor spawn a goroutine?** Ensure `Stop()` / `Close()` exists, uses `sync.Once`, and `main.go` calls it during shutdown.
5. **Does the service use bounded parallelism?** If it spawns goroutines in a loop, wrap them with `internal/concurrency.Semaphore`.
6. **Are there cleanup goroutines?** Register them with `middleware.StopAll()` (for middleware) or the service's own `wg.Wait()` path.

## References

- `digital.vasic.lazy/pkg/lazy` — primitive types + tests
- `catalog-api/internal/concurrency/semaphore.go` — bounded parallelism
- `catalog-api/internal/lifecycle/lazy_services.go` — `LazyServiceRegistry` for deferred service wiring
- `scripts/audit-lazy-init.sh` — heuristic audit
- `scripts/audit-semaphores.sh` — heuristic audit
- Module 35 of the video course — `docs/video-course/MODULE35_LAZY_LOADING_SEMAPHORE_PATTERNS_SCRIPT.md`
