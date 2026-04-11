# Module 35 — Lazy Loading & Semaphore Patterns

**Duration:** 18 minutes
**Prerequisites:** Module 15 (Concurrency), Module 29 (Module Architecture)

## Learning objectives

1. Recognize when lazy initialization helps and when it's overkill.
2. Use `digital.vasic.lazy` / `internal/lifecycle.LazyServiceRegistry` for deferred construction.
3. Bound parallelism with `internal/concurrency/semaphore.go` instead of unbounded `go func()`.
4. Avoid the "eager at boot, fail at boot" pitfall that blocks server startup.

## Segment 1 — Lazy vs eager (0:00 – 4:00)

**Eager**: service is fully constructed at startup, including DB connections, HTTP clients, background goroutines. Pro: every error surfaces at boot. Con: boot time scales with the number of services, and transient failures (slow DB, TMDB rate limit) block the whole server.

**Lazy**: service is created with its dependencies but expensive work is deferred until first use. Pro: fast boot, resilient to transient failures. Con: first-use latency spike, errors surface late.

**Rule of thumb**: services that depend on external APIs (TMDB, MusicBrainz) should be lazy. Services that are on every request path (auth, DB) should be eager so you find out at boot if they're broken.

## Segment 2 — `digital.vasic.lazy` primitives (4:00 – 8:00)

```go
import "digital.vasic.lazy/pkg/lazy"

type MetadataClient struct {
    inner lazy.Value[*http.Client]
}

func NewMetadataClient() *MetadataClient {
    return &MetadataClient{
        inner: lazy.NewValue(func() (*http.Client, error) {
            return &http.Client{Timeout: 30 * time.Second}, nil
        }),
    }
}

func (c *MetadataClient) Do(req *http.Request) (*http.Response, error) {
    client, err := c.inner.Get()
    if err != nil {
        return nil, err
    }
    return client.Do(req)
}
```

`lazy.Value[T]` guarantees the loader runs exactly once (via `sync.Once`) even under concurrent `Get()` calls. `Reset()` invalidates the cached value for deliberate re-initialization.

## Segment 3 — `LazyServiceRegistry` (8:00 – 12:00)

**Show on screen:** `catalog-api/internal/lifecycle/registry.go`.

```go
reg := lifecycle.NewLazyServiceRegistry()
reg.Register("cache", func() (interface{}, error) {
    return services.NewCacheService(db, logger), nil
})
reg.Register("reporter", func() (interface{}, error) {
    return services.NewReportingService(...), nil
})
```

Services are constructed on first `Get("cache")`. Dependencies are resolved in order. `Close()` tears down in reverse order.

**Use case**: providers in `internal/media/providers/providers.go` — 14 metadata sources (TMDB, MusicBrainz, IGDB, …) are each wrapped in a `LazyProvider` so missing API keys don't block construction of the `ProviderManager`.

## Segment 4 — Semaphore-bounded parallelism (12:00 – 16:00)

**Show on screen:** `catalog-api/internal/concurrency/semaphore.go`.

Anti-pattern:
```go
for _, item := range items {
    go processItem(item) // ← unbounded concurrency, resource exhaustion
}
```

Correct:
```go
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

The semaphore caps concurrency at 8 regardless of how many items are queued. Context cancellation propagates: `sem.Acquire(ctx)` returns early if the context is done.

## Segment 5 — Common pitfalls (16:00 – 18:00)

1. **Eager HTTP client at boot** — ties boot health to TMDB's availability. Use lazy.
2. **Unbounded goroutine fan-out over a scan result** — easily spawns 100k goroutines on a large NAS scan. Always bound with a semaphore.
3. **Lazy service that never errors** — if `lazy.Value.Get` always returns nil, you've hidden the real failure. Return `error` honestly.
4. **Forgetting to call `Release()`** — leaks the semaphore slot. Always pair `Acquire` with `defer Release`.

## Exercise

Wrap `internal/media/providers/tmdb.go` in a `lazy.Value[*TMDBProvider]` so construction is deferred until the first call, and fan out metadata fetches with a semaphore capping concurrency at 4.

## Assessment

1. What does `lazy.Value[T].Reset()` do? Answer: invalidates the cached value so the next `Get` re-runs the loader.
2. When would you prefer a buffered channel over a semaphore? Answer: when you also need queueing semantics, not just a concurrency cap.
