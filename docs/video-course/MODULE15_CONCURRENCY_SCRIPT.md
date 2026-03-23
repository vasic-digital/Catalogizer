# Module 15: Concurrency Patterns in Go -- Video Script

**Duration**: 45 minutes
**Prerequisites**: Module 2 (Backend Development), familiarity with Go goroutines and channels

---

## Video 15.1: Goroutine Lifecycle Management (15 min)

### Opening

Welcome to Module 15 where we explore concurrency patterns used throughout Catalogizer. Go's concurrency primitives are powerful, but incorrect usage leads to goroutine leaks, race conditions, and deadlocks. This module shows how Catalogizer solves each of these challenges in production.

### Goroutine Lifecycle in Catalogizer

Every goroutine in Catalogizer follows a lifecycle: creation, work, and termination. The key principle is that every goroutine must have a clear owner responsible for stopping it.

Let us look at how the `UniversalScanner` manages goroutines. Open `catalog-api/internal/services/universal_scanner.go`:

```go
func (s *UniversalScanner) Start() error {
    s.ctx, s.cancel = context.WithCancel(context.Background())
    // Worker goroutines are bounded by scannerConcurrency
    for i := 0; i < s.concurrency; i++ {
        go s.worker(s.ctx)
    }
    return nil
}

func (s *UniversalScanner) Stop() {
    s.cancel() // Signal all workers to stop
    s.wg.Wait() // Wait for all workers to finish
}
```

The pattern here is:
1. `context.WithCancel` creates a cancellation signal
2. Workers check `ctx.Done()` in their loop
3. `Stop()` cancels the context and waits via `sync.WaitGroup`
4. The `defer universalScanner.Stop()` in `main.go` guarantees cleanup

### Context Propagation

Every database query, HTTP request, and scan operation carries a context. This allows cancellation to propagate through the entire call chain:

```go
func (s *UniversalScanner) worker(ctx context.Context) {
    defer s.wg.Done()
    for {
        select {
        case <-ctx.Done():
            return
        case job := <-s.jobQueue:
            s.processJob(ctx, job)
        }
    }
}
```

When `Stop()` is called, the context cancellation cascades: the worker exits its loop, in-flight database queries return with `context.Canceled`, and network connections are closed.

### WaitGroup Pattern

The `sync.WaitGroup` is used whenever we need to wait for multiple goroutines to complete. In Catalogizer, it appears in:

- `UniversalScanner`: waits for all scan workers
- `AggregationService`: waits for post-scan processing
- `AssetManager`: waits for resolution workers

The pattern is always the same:

```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(ctx, item)
    }(item)
}
wg.Wait()
```

The critical rule: call `wg.Add(1)` before launching the goroutine, and `defer wg.Done()` as the first line inside the goroutine.

---

## Video 15.2: Mutex Patterns and sync.Once (15 min)

### Mutex Usage in Catalogizer

The `sync.Mutex` protects shared state from concurrent access. Catalogizer uses mutexes in several services. Let us examine the `CacheService` in `catalog-api/internal/services/cache_service.go`:

```go
type CacheService struct {
    mu      sync.RWMutex
    cache   map[string]*cacheEntry
    stopCh  chan struct{}
    stopped bool
}
```

This uses a `sync.RWMutex` which allows multiple concurrent readers but exclusive writers:

```go
func (s *CacheService) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    entry, ok := s.cache[key]
    if !ok || entry.isExpired() {
        return nil, false
    }
    return entry.value, true
}

func (s *CacheService) Set(key string, value interface{}, ttl time.Duration) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.cache[key] = &cacheEntry{value: value, expiry: time.Now().Add(ttl)}
}
```

The rules for mutex usage:
1. Always use `defer` for `Unlock` to prevent deadlocks on panics
2. Use `RWMutex` when reads vastly outnumber writes
3. Hold the lock for the minimum necessary duration
4. Never call an external function while holding a lock (risk of deadlock)

### The CacheService Close() Pattern

The `CacheService` demonstrates a clean shutdown pattern:

```go
func (s *CacheService) Close() {
    s.mu.Lock()
    if s.stopped {
        s.mu.Unlock()
        return
    }
    s.stopped = true
    close(s.stopCh)
    s.mu.Unlock()
}
```

This pattern ensures:
- `Close()` is idempotent (safe to call multiple times)
- The `stopCh` channel signals the cleanup goroutine to exit
- The mutex prevents a race between `Close()` and the cleanup goroutine

### sync.Once for Initialization

`sync.Once` guarantees that a function runs exactly once, regardless of how many goroutines call it. Catalogizer uses this for lazy initialization:

```go
type lazyConnection struct {
    once sync.Once
    conn *Connection
    err  error
}

func (l *lazyConnection) Get() (*Connection, error) {
    l.once.Do(func() {
        l.conn, l.err = createConnection()
    })
    return l.conn, l.err
}
```

The `Lazy` module (`digital.vasic.lazy`) provides a generic version of this pattern:

```go
lazy := lazy.New(func() (*Database, error) {
    return database.NewConnection(config)
})

// First call initializes; subsequent calls return cached value
db, err := lazy.Get()
```

This is used throughout Catalogizer for:
- Database connections (lazy initialization avoids connecting at import time)
- Redis client initialization
- Asset resolver chain setup

### Semaphore Pattern

The `Concurrency` module (`digital.vasic.concurrency`) provides a semaphore for bounding parallelism:

```go
sem := concurrency.NewSemaphore(4) // max 4 concurrent operations

for _, file := range files {
    sem.Acquire()
    go func(f File) {
        defer sem.Release()
        processFile(ctx, f)
    }(file)
}
sem.Wait() // Wait for all to complete
```

This pattern is used in:
- The scanner to limit concurrent file operations per storage root
- The asset manager to limit concurrent resolution workers (configured as 4 in `main.go`)
- The middleware `ConcurrencyLimiter(100)` to cap in-flight HTTP requests

---

## Video 15.3: Race Detection and Prevention (15 min)

### Running the Race Detector

Go's built-in race detector catches data races at runtime:

```bash
# Run all tests with race detection
GOMAXPROCS=3 go test -race ./... -p 2 -parallel 2

# Run a specific test with race detection
go test -race -v -run TestCacheService ./internal/services/
```

The resource limits (`GOMAXPROCS=3 -p 2 -parallel 2`) are critical for Catalogizer's host machine which limits workloads to 30-40% of total resources.

### Common Race Conditions and Fixes

**Race 1: Concurrent map access**

Go maps are not safe for concurrent use. Always protect with a mutex:

```go
// WRONG - data race
type Registry struct {
    items map[string]Item
}

// CORRECT - protected by mutex
type Registry struct {
    mu    sync.RWMutex
    items map[string]Item
}
```

**Race 2: Shared variable in goroutine closure**

```go
// WRONG - loop variable captured by reference
for _, root := range roots {
    go func() {
        scan(root) // root changes on each iteration
    }()
}

// CORRECT - pass as parameter
for _, root := range roots {
    go func(r StorageRoot) {
        scan(r)
    }(root)
}
```

**Race 3: Check-then-act without lock**

```go
// WRONG - TOCTOU race
if !s.stopped {
    s.stopped = true
    close(s.stopCh)
}

// CORRECT - atomic check-and-set under lock
s.mu.Lock()
if !s.stopped {
    s.stopped = true
    close(s.stopCh)
}
s.mu.Unlock()
```

### Known Flaky Test

Catalogizer has one known pre-existing flaky test: `TestChaos_ConcurrentDatabaseAccess`. This test intentionally creates concurrent SQLite access patterns that can trigger WAL mode contention. The test documents the behavior rather than indicating a bug -- SQLite serializes writes by design.

### Testing Concurrency

Catalogizer uses table-driven tests for concurrent scenarios:

```go
func TestCacheService_ConcurrentAccess(t *testing.T) {
    cache := NewCacheService(db, logger)
    defer cache.Close()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(2)
        go func(i int) {
            defer wg.Done()
            cache.Set(fmt.Sprintf("key-%d", i), i, time.Minute)
        }(i)
        go func(i int) {
            defer wg.Done()
            cache.Get(fmt.Sprintf("key-%d", i))
        }(i)
    }
    wg.Wait()
}
```

### Summary

The concurrency patterns in Catalogizer follow three principles:
1. Every goroutine has an owner and a shutdown path
2. Shared state is protected by mutexes or channels
3. Parallelism is bounded by semaphores and the `ConcurrencyLimiter` middleware

These patterns are extracted into reusable modules: `digital.vasic.concurrency` (semaphores, WaitGroups), `digital.vasic.lazy` (lazy initialization), `digital.vasic.memory` (leak detection), and `digital.vasic.recovery` (circuit breaker).

---

## Exercises

1. Add a goroutine leak detector to the `CacheService` that logs a warning if the cleanup goroutine has not run in 5 minutes
2. Convert the `CacheService` mutex from `sync.Mutex` to `sync.RWMutex` and benchmark the difference with 90% reads / 10% writes
3. Write a table-driven test that verifies the `UniversalScanner` properly shuts down all workers within 5 seconds

---

## Key Files Referenced

- `catalog-api/internal/services/universal_scanner.go` -- Scanner goroutine management
- `catalog-api/internal/services/cache_service.go` -- CacheService Close() pattern
- `catalog-api/main.go` -- Service initialization and deferred cleanup
- `catalog-api/middleware/concurrency.go` -- ConcurrencyLimiter middleware
- `Concurrency/` -- Reusable concurrency utilities module
- `Lazy/` -- Generic lazy initialization module
- `Memory/` -- Memory leak detection module
- `Recovery/` -- Circuit breaker and recovery patterns module
