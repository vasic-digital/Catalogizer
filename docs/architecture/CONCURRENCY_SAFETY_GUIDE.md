# Concurrency Safety Guide

Patterns for safe concurrent access in the Catalogizer Go backend. This document covers the mutex, semaphore, channel, and circuit breaker patterns used throughout `catalog-api`.

## Mutex Patterns

### sync.RWMutex for Shared State

The most common pattern is `sync.RWMutex` protecting a map or struct fields. Readers acquire `RLock`, writers acquire `Lock`.

**CircuitBreakerManager** (`internal/recovery/circuit_breaker.go`):

```go
type CircuitBreakerManager struct {
    breakers map[string]*CircuitBreaker
    mutex    sync.RWMutex
    logger   *zap.Logger
}

func (m *CircuitBreakerManager) Get(name string) *CircuitBreaker {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.breakers[name]
}

func (m *CircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    if cb, exists := m.breakers[name]; exists {
        return cb
    }
    cb := NewCircuitBreaker(config)
    m.breakers[name] = cb
    return cb
}
```

**ChallengeService** (`services/challenge_service.go`):

```go
type ChallengeService struct {
    mu      sync.RWMutex
    results []*challenge.Result
}

func (s *ChallengeService) GetResults() []*challenge.Result {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]*challenge.Result, len(s.results))
    copy(out, s.results)
    return out
}
```

Note the defensive copy in `GetResults()` -- returning a copy prevents callers from modifying the internal slice.

### sync.Once for Lazy Initialization

The `pkg/lazy/` package wraps `sync.Once` for thread-safe lazy initialization:

```go
type Value[T any] struct {
    once   sync.Once
    value  T
    err    error
    loader func() (T, error)
}

func (v *Value[T]) Get() (T, error) {
    v.once.Do(func() {
        v.value, v.err = v.loader()
    })
    return v.value, v.err
}
```

Usage: database connections, configuration loading, and service initialization that should happen exactly once.

## Semaphore Pattern

### Channel-Based Semaphore

`pkg/semaphore/semaphore.go` implements a counting semaphore using a buffered channel:

```go
type Semaphore struct {
    ch     chan struct{}
    mu     sync.RWMutex
    closed bool
}

func New(maxConcurrent int) *Semaphore {
    return &Semaphore{
        ch: make(chan struct{}, maxConcurrent),
    }
}

func (s *Semaphore) Acquire(ctx context.Context) error {
    s.mu.RLock()
    if s.closed {
        s.mu.RUnlock()
        return ErrSemaphoreClosed
    }
    s.mu.RUnlock()

    select {
    case <-ctx.Done():
        return ctx.Err()
    case s.ch <- struct{}{}:
        return nil
    }
}

func (s *Semaphore) Release() {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.closed { return }
    select {
    case <-s.ch:
    default:
    }
}
```

Key design choices:
- `Acquire` blocks until a slot is available or the context is cancelled
- `TryAcquire` returns immediately with `true`/`false` (non-blocking)
- `Close` prevents new acquisitions and closes the channel
- The `RWMutex` protects the `closed` flag, not the channel operations

### golang.org/x/sync/semaphore.Weighted

The `UniversalScanner` (`internal/services/universal_scanner.go`) uses a weighted semaphore from the `x/sync` package to limit concurrent scans:

```go
type UniversalScanner struct {
    scanSem            *semaphore.Weighted
    maxConcurrentScans int
    // ...
}
```

## Channel Patterns

### Worker Pool with Channels

The `EnhancedChangeWatcher` (`internal/media/realtime/enhanced_watcher.go`) uses channels for a worker pool pattern:

```go
type EnhancedChangeWatcher struct {
    changeQueue chan EnhancedChangeEvent  // buffered work queue
    workers     int
    stopCh      chan struct{}             // signal channel for shutdown
    wg          sync.WaitGroup           // tracks active workers
}
```

- `changeQueue`: buffered channel acts as a work queue
- `stopCh`: unbuffered channel used as a signal (close to broadcast stop to all workers)
- `wg`: tracks goroutine completion for clean shutdown

### Scan Queue

The `UniversalScanner` uses the same pattern:

```go
type UniversalScanner struct {
    scanQueue chan ScanJob
    stopCh    chan struct{}
    wg        sync.WaitGroup
}
```

### Context-Based Cancellation with select

Throughout the codebase, `select` statements combine channel operations with context cancellation:

```go
select {
case <-ctx.Done():
    return ctx.Err()
case <-time.After(delay):
    // retry
}
```

This pattern appears in retry logic (`internal/recovery/retry.go`), semaphore acquisition, and bulkhead execution.

## Circuit Breaker Pattern

`internal/recovery/circuit_breaker.go` wraps `digital.vasic.concurrency/pkg/breaker` with logging and state change callbacks.

Three states:
- **Closed** (normal): requests pass through
- **Open** (failing): requests rejected immediately after `MaxFailures` consecutive errors
- **Half-Open** (probing): one request allowed through after `ResetTimeout`

```go
cb := NewCircuitBreaker(CircuitBreakerConfig{
    Name:         "smb-nas1",
    MaxFailures:  5,
    ResetTimeout: 60 * time.Second,
})

err := cb.Execute(func() error {
    return connectToSMB()
})
```

Used by SMB connections to prevent cascading failures when a NAS is unreachable.

## Bulkhead Pattern

`internal/recovery/retry.go` implements the bulkhead pattern for resource isolation:

```go
type Bulkhead struct {
    semaphore chan struct{}
    config    BulkheadConfig
}

func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
    select {
    case <-b.semaphore:
        defer func() { b.semaphore <- struct{}{} }()
        return fn()
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(b.config.Timeout):
        return NewRetryableError(context.DeadlineExceeded, true)
    }
}
```

The bulkhead pre-fills its semaphore channel and consumes a token for each concurrent execution. This isolates resource pools so that one failing subsystem cannot exhaust resources needed by others.

## Retry with Exponential Backoff

`internal/recovery/retry.go` provides retry with jitter:

```go
config := RetryConfig{
    MaxAttempts:   3,
    InitialDelay:  1 * time.Second,
    MaxDelay:      30 * time.Second,
    BackoffFactor: 2.0,
    Jitter:        true,
}

err := Retry(ctx, config, func() error {
    return someUnreliableOperation()
})
```

- `RetryableError` wraps errors with a `Retryable` flag -- non-retryable errors break the loop immediately
- 10% jitter prevents thundering herd when multiple clients retry simultaneously
- Context cancellation is checked between attempts

## SMB Connection State Machine

`internal/smb/resilience.go` manages connection states with mutex-protected transitions:

```go
type SMBSource struct {
    State     ConnectionState  // connected, disconnected, reconnecting, offline
    mu        sync.RWMutex
    // ...
}
```

Four states: `StateConnected`, `StateDisconnected`, `StateReconnecting`, `StateOffline`. State transitions are protected by the mutex and reported via Prometheus metrics.

## Common Rules

1. **Always use `defer` for unlock.** Prevents deadlocks on early returns or panics.
2. **Prefer `RWMutex` over `Mutex`** when reads significantly outnumber writes.
3. **Defensive copies** when returning slices or maps from mutex-protected methods.
4. **Context-aware blocking.** Every blocking operation should accept and honor `context.Context`.
5. **Close channels to broadcast.** Use `close(stopCh)` to signal all goroutines, not sending values.
6. **WaitGroup for cleanup.** Track goroutines with `sync.WaitGroup` and `wg.Wait()` during shutdown.

## Running Concurrency Tests

```bash
cd catalog-api

# Race detector catches data races at runtime
go test -race ./...

# Chaos tests specifically test concurrent scenarios
go test -race -v -run TestChaos_ConcurrentDatabaseAccess ./tests/integration/
go test -race -v -run TestChaos_ConnectionPoolExhaustion ./tests/integration/

# Semaphore and lazy value tests
go test -race -v ./pkg/semaphore/
go test -race -v ./pkg/lazy/

# Resource-limited execution
GOMAXPROCS=3 go test -race ./... -p 2 -parallel 2
```
