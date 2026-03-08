# Lazy Loading Architecture

The `digital.vasic.lazy` module (`Lazy/` submodule) provides generic lazy-loading primitives used across the Catalogizer backend for deferred initialization of expensive resources.

## Core Types

### Value[T]

Lazily loads a value of type `T` on first `Get()` call. The loader function runs at most once, even under concurrent access.

```go
type Value[T any] struct {
    mu     sync.Mutex
    once   sync.Once
    value  T
    err    error
    loader func() (T, error)
}
```

**API:**

| Method | Description |
|--------|-------------|
| `NewValue[T](loader)` | Create with a loader function |
| `Get() (T, error)` | Returns the lazily-loaded value |
| `MustGet() T` | Returns value, panics on error |
| `Reset()` | Clears cached value; loader runs again on next `Get()` |

### Service[T]

Lazily initializes a service of type `T` exactly once. Identical to `Value[T]` but semantically represents a long-lived service rather than a data value.

```go
type Service[T any] struct {
    mu      sync.Mutex
    once    sync.Once
    service T
    initErr error
    init    func() (T, error)
}
```

**API:**

| Method | Description |
|--------|-------------|
| `NewService[T](init)` | Create with an init function |
| `Get() (T, error)` | Returns the lazily-initialized service |
| `Initialized() bool` | Returns true if init succeeded |

## Synchronization Strategy

Both types use a **dual-lock pattern**: `sync.Mutex` + `sync.Once`.

1. `sync.Once` guarantees the loader/init function executes at most once
2. `sync.Mutex` serializes access to protect the `Reset()` operation, which replaces the `sync.Once` instance

This is stronger than `sync.Once` alone because `Reset()` must atomically clear the cached value and replace the `sync.Once`. Without the mutex, a concurrent `Get()` could see a partially-reset state.

```
Get() flow:
    Lock mutex
    once.Do(loader)   <-- runs loader at most once
    read value, err
    Unlock mutex
    return value, err

Reset() flow:
    Lock mutex
    once = sync.Once{}   <-- new Once instance
    value = zero
    err = nil
    Unlock mutex
```

## Design Patterns

| Pattern | Application |
|---------|-------------|
| **Proxy** | `Value[T]` and `Service[T]` defer initialization until first access |
| **Singleton** | `sync.Once` guarantees loader runs exactly once per instance |
| **Factory** | `NewValue()` and `NewService()` constructors |

## Usage in Catalogizer

Lazy loading is applied to resources that are expensive to initialize and may not be needed on every request:

- **Database connections** -- deferred until first query
- **External service clients** -- TMDB, OMDB API clients initialized on first metadata request
- **File system clients** -- SMB/FTP/NFS connections created on first browse or scan
- **Asset loading** -- static assets loaded on first serve via the `Assets/` module

The `Concurrency/pkg/lazyloader` package provides an additional lazy loader implementation within the concurrency module, offering integration with the worker pool system.

## Cache Invalidation

`Value[T].Reset()` enables cache invalidation for scenarios like:

- Configuration reload (re-read config file)
- Connection recovery (reconnect after network failure)
- Credential rotation (re-authenticate with new credentials)

After `Reset()`, the next `Get()` call re-executes the loader function. All concurrent `Get()` calls block on the mutex until the new value is loaded.

## Testing

Tests are in `Lazy/pkg/lazy/lazy_test.go` and `lazy_coverage_test.go`. Run with:

```bash
cd Lazy && go test ./... -count=1 -race
```

The `-race` flag is important because the types are designed for concurrent use.

## Source

- Module: `Lazy/pkg/lazy/lazy.go`
- Tests: `Lazy/pkg/lazy/lazy_test.go`
- Go module: `digital.vasic.lazy` (Go 1.24+)
- Wired in `catalog-api/go.mod` via `replace digital.vasic.lazy => ../Lazy`
