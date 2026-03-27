# ADR-007: Lazy Initialization via LazyServiceRegistry

## Status
Accepted (2026-03-15)

## Context

Catalogizer's backend (`catalog-api`) registers a large number of services at startup: media detection pipelines, metadata providers (TMDB, OMDB, OpenLibrary, MusicBrainz), subtitle services, format conversion workers, cache layers, WebSocket handlers, and more. In production, many of these services are used infrequently or only for specific user actions -- for example, subtitle translation is rarely invoked, and format conversion may go days without a request on a personal media server.

With eager initialization, every service is fully constructed, its dependencies resolved, and its background goroutines started during the application boot sequence. This creates several problems:

1. **Slow startup time**: Constructing all services sequentially takes 3-5 seconds on modest hardware (Raspberry Pi, NAS appliances), which is noticeable during development iteration and container restarts.

2. **Wasted resources**: Services that are never used during a session still hold open connections, allocate memory for internal caches, and run background cleanup goroutines. On resource-constrained deployments (the 30-40% host resource limit is a hard constraint), this idle overhead is significant.

3. **Initialization ordering complexity**: Services have interdependencies (e.g., `AggregationService` depends on `MediaItemRepository`, which depends on `database.DB`). Eager initialization requires careful ordering in `main.go`, and adding a new service means manually inserting it at the correct position in the boot sequence.

4. **Startup failures from optional dependencies**: If an optional external service (Redis, a metadata provider API) is unavailable at startup, eager initialization either fails the entire application or requires complex fallback logic in every constructor.

## Decision

We adopt a two-part lazy initialization strategy:

### 1. LazyServiceRegistry (`internal/lifecycle/`)

A centralized registry where services are registered with their factory functions and dependency declarations. Services are not instantiated until first access. The registry resolves dependencies automatically using topological ordering when a service is requested.

```
Registration Phase (startup):
    registry.Register("aggregation", factory, deps: ["media-item-repo", "title-parser"])
    registry.Register("media-item-repo", factory, deps: ["database"])
    registry.Register("title-parser", factory, deps: [])
    ...

Resolution Phase (first access):
    registry.Get("aggregation")
        -> resolves "title-parser" (no deps, instantiate)
        -> resolves "media-item-repo" (depends on "database", already resolved)
        -> instantiates "aggregation" with resolved deps
        -> caches instance for subsequent calls
```

### 2. sync.Once-based LazyProvider

For simpler cases where a full registry is unnecessary, individual services use `sync.Once` to defer expensive initialization until first use. This pattern is used for:

- **HTTP client pools** (`internal/httpclient/`): Connection pools are created on the first outbound request, not at startup.
- **Cache warming** (`internal/cache/`): Cache entries are populated on first read, not preloaded.
- **Metadata provider clients**: API clients for TMDB, OMDB, etc. are constructed when the first metadata lookup is requested.

```go
type LazyProvider[T any] struct {
    once     sync.Once
    factory  func() (T, error)
    instance T
    err      error
}

func (lp *LazyProvider[T]) Get() (T, error) {
    lp.once.Do(func() {
        lp.instance, lp.err = lp.factory()
    })
    return lp.instance, lp.err
}
```

### Key Implementation Details

- **Thread safety**: Both `LazyServiceRegistry` and `LazyProvider` are safe for concurrent access. The registry uses `sync.RWMutex` for the instance cache; `LazyProvider` relies on `sync.Once` guarantees.
- **Circular dependency detection**: The registry performs cycle detection during resolution and returns a clear error identifying the cycle path.
- **Error propagation**: If a factory function returns an error, the error is cached and returned on every subsequent `Get()` call. The service is not retried automatically -- this prevents repeated expensive failures.
- **Shutdown ordering**: The registry tracks initialization order and provides a `Shutdown()` method that tears down services in reverse order, ensuring dependencies outlive their dependents.

## Consequences

### Positive

- **Faster startup**: Application boot time reduced from ~4 seconds to under 1 second on reference hardware. Only the HTTP server, router, and registry scaffolding are initialized eagerly. Services are instantiated on demand.
- **Lower idle resource usage**: Unused services consume zero resources. A deployment that only uses media browsing (no subtitles, no conversion) never instantiates those services or their background goroutines.
- **Simplified main.go**: Service registration is declarative (name, factory, dependencies) rather than imperative (ordered constructor calls). Adding a new service requires one `Register()` call, and the registry handles ordering.
- **Graceful degradation for optional services**: If Redis is unavailable, the cache service fails on first access rather than at startup. The application can still serve requests that do not require caching.
- **Dependency ordering is automatic**: The topological sort in the registry eliminates manual ordering bugs.

### Negative

- **First-request latency**: The first request that triggers a lazy service incurs the initialization cost. For metadata providers that require HTTP connection setup, this can add 100-500ms to the first request. Subsequent requests are unaffected.
- **More complex initialization code**: Factory functions must be self-contained closures that capture their configuration. Debugging initialization issues requires understanding the registry's resolution flow rather than reading sequential code in `main.go`.
- **Deferred error discovery**: Errors in service construction (misconfiguration, missing credentials) are not surfaced until the service is first used, which may be minutes or hours after startup. This can make operational debugging harder compared to fail-fast eager initialization.
- **Testing considerations**: Unit tests must account for lazy initialization by either pre-warming services or using direct construction. The `LazyProvider` pattern is testable via factory injection, but the `LazyServiceRegistry` requires a test registry setup.
