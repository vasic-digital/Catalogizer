# Concurrency Patterns in Catalogizer

## Overview

This document describes the concurrency patterns, synchronization primitives, and best practices used throughout the Catalogizer project to ensure thread-safe, race-free, and deadlock-free operation.

## Table of Contents

1. [Goroutine Management](#goroutine-management)
2. [Channel Patterns](#channel-patterns)
3. [Mutex Usage](#mutex-usage)
4. [Context Cancellation](#context-cancellation)
5. [Graceful Shutdown](#graceful-shutdown)
6. [Race Prevention](#race-prevention)
7. [Deadlock Avoidance](#deadlock-avoidance)
8. [Testing Concurrent Code](#testing-concurrent-code)

## Goroutine Management

### Principle: Always Clean Up Goroutines

Every goroutine must have a clear termination path. Never launch "fire-and-forget" goroutines without a way to shut them down.

**✅ Good Example:**

```go
func (s *Service) Start(ctx context.Context) error {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()

        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return // Clean termination
            case <-ticker.C:
                s.performTask()
            }
        }
    }()

    return nil
}

func (s *Service) Stop() {
    s.cancel() // Signal all goroutines to stop
    s.wg.Wait() // Wait for all goroutines to finish
}
```

**❌ Bad Example:**

```go
func (s *Service) Start() {
    go func() {
        for {
            time.Sleep(1 * time.Minute)
            s.performTask() // No way to stop this goroutine!
        }
    }()
}
```

### Pattern: Worker Pool

For CPU-bound tasks, use a worker pool with bounded concurrency:

```go
type WorkerPool struct {
    workers   int
    tasks     chan func()
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewWorkerPool(workers int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())

    wp := &WorkerPool{
        workers: workers,
        tasks:   make(chan func(), 100), // Buffered channel
        ctx:     ctx,
        cancel:  cancel,
    }

    // Start workers
    for i := 0; i < workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }

    return wp
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()

    for {
        select {
        case <-wp.ctx.Done():
            return
        case task := <-wp.tasks:
            task()
        }
    }
}

func (wp *WorkerPool) Submit(task func()) error {
    select {
    case <-wp.ctx.Done():
        return errors.New("worker pool is stopped")
    case wp.tasks <- task:
        return nil
    }
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
    wp.wg.Wait()
}
```

**Usage in Catalogizer:**
- File scanning: `internal/media/scanner/pool.go`
- Media analysis: `internal/media/analyzer/worker_pool.go`

### Pattern: Panic Recovery in Goroutines

Always recover from panics in goroutines to prevent crashes:

```go
func (s *Service) safeGoroutine(fn func()) {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        defer func() {
            if r := recover(); r != nil {
                s.logger.Error("Goroutine panic",
                    zap.Any("panic", r),
                    zap.Stack("stacktrace"),
                )
            }
        }()

        fn()
    }()
}
```

## Channel Patterns

### Pattern: Fan-Out, Fan-In

Distribute work across multiple goroutines and collect results:

```go
func processFiles(files []string) []Result {
    // Fan-out: distribute work
    results := make(chan Result, len(files))
    var wg sync.WaitGroup

    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            results <- processFile(f)
        }(file)
    }

    // Close results channel when all workers done
    go func() {
        wg.Wait()
        close(results)
    }()

    // Fan-in: collect results
    var collected []Result
    for result := range results {
        collected = append(collected, result)
    }

    return collected
}
```

### Pattern: Pipeline

Chain multiple processing stages:

```go
func pipeline(ctx context.Context, input <-chan File) <-chan ProcessedFile {
    // Stage 1: Validate files
    validated := make(chan File)
    go func() {
        defer close(validated)
        for file := range input {
            if validate(file) {
                select {
                case <-ctx.Done():
                    return
                case validated <- file:
                }
            }
        }
    }()

    // Stage 2: Process files
    processed := make(chan ProcessedFile)
    go func() {
        defer close(processed)
        for file := range validated {
            select {
            case <-ctx.Done():
                return
            case processed <- process(file):
            }
        }
    }()

    return processed
}
```

### Pattern: Bounded Concurrency with Semaphore

Limit concurrent operations:

```go
type Semaphore struct {
    sem chan struct{}
}

func NewSemaphore(maxConcurrency int) *Semaphore {
    return &Semaphore{
        sem: make(chan struct{}, maxConcurrency),
    }
}

func (s *Semaphore) Acquire() {
    s.sem <- struct{}{}
}

func (s *Semaphore) Release() {
    <-s.sem
}

// Usage
func processWithLimit(items []Item) {
    sem := NewSemaphore(10) // Max 10 concurrent operations
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()

            sem.Acquire()
            defer sem.Release()

            process(i)
        }(item)
    }

    wg.Wait()
}
```

## Mutex Usage

### Golden Rule: Always Use Defer for Unlock

This prevents deadlocks from panics or early returns:

```go
// ✅ Correct
func (s *Service) UpdateState() {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.state = "updated"
    // Even if panic or return here, unlock will happen
}

// ❌ Wrong
func (s *Service) UpdateState() {
    s.mu.Lock()
    s.state = "updated"
    s.mu.Unlock() // Won't execute if panic occurs above
}
```

### Pattern: Read-Write Locks

Use `sync.RWMutex` when reads are frequent:

```go
type Cache struct {
    mu    sync.RWMutex
    data  map[string]interface{}
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock() // Multiple readers can hold this simultaneously
    defer c.mu.RUnlock()

    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key string, val interface{}) {
    c.mu.Lock() // Exclusive lock for writes
    defer c.mu.Unlock()

    c.data[key] = val
}
```

### Pattern: Avoiding Lock Contention

Minimize critical sections:

```go
// ✅ Good: Lock only when accessing shared state
func (s *Service) ProcessItem(item Item) {
    // Do expensive work outside lock
    result := expensiveComputation(item)

    // Only lock when updating shared state
    s.mu.Lock()
    s.results[item.ID] = result
    s.mu.Unlock()
}

// ❌ Bad: Holding lock during expensive work
func (s *Service) ProcessItem(item Item) {
    s.mu.Lock()
    defer s.mu.Unlock()

    result := expensiveComputation(item) // Lock held too long!
    s.results[item.ID] = result
}
```

## Context Cancellation

### Pattern: Context-Aware Operations

All long-running operations should respect context cancellation:

```go
func (s *Service) ProcessFiles(ctx context.Context, files []string) error {
    for _, file := range files {
        // Check context before each iteration
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        if err := s.processFile(ctx, file); err != nil {
            return err
        }
    }

    return nil
}
```

### Pattern: Timeout with Context

```go
func fetchDataWithTimeout() (Data, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resultCh := make(chan Data, 1)
    errCh := make(chan error, 1)

    go func() {
        data, err := slowFetch()
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- data
    }()

    select {
    case <-ctx.Done():
        return Data{}, ctx.Err()
    case err := <-errCh:
        return Data{}, err
    case data := <-resultCh:
        return data, nil
    }
}
```

## Graceful Shutdown

### Pattern: Coordinated Service Shutdown

```go
type Service struct {
    wg     sync.WaitGroup
    ctx    context.Context
    cancel context.CancelFunc
    logger *zap.Logger
}

func NewService() *Service {
    ctx, cancel := context.WithCancel(context.Background())
    return &Service{
        ctx:    ctx,
        cancel: cancel,
        logger: zap.L(),
    }
}

func (s *Service) Start() error {
    // Start background workers
    s.wg.Add(3)
    go s.worker1()
    go s.worker2()
    go s.worker3()

    return nil
}

func (s *Service) Stop() error {
    s.logger.Info("Stopping service")

    // Signal all goroutines to stop
    s.cancel()

    // Wait for graceful shutdown with timeout
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        s.logger.Info("Service stopped gracefully")
        return nil
    case <-time.After(30 * time.Second):
        s.logger.Error("Service shutdown timeout")
        return errors.New("shutdown timeout")
    }
}

func (s *Service) worker1() {
    defer s.wg.Done()

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-s.ctx.Done():
            s.logger.Info("Worker1 stopping")
            return
        case <-ticker.C:
            s.doWork()
        }
    }
}
```

## Race Prevention

### Tools and Techniques

1. **Always run tests with race detector:**

```bash
go test -race ./...
```

2. **Use atomic operations for counters:**

```go
import "sync/atomic"

type Counter struct {
    value int64
}

func (c *Counter) Increment() {
    atomic.AddInt64(&c.value, 1)
}

func (c *Counter) Get() int64 {
    return atomic.LoadInt64(&c.value)
}
```

3. **Use channels instead of shared memory when possible:**

```go
// ✅ Good: Pass data through channels
func producer(out chan<- Data) {
    for {
        out <- generateData()
    }
}

func consumer(in <-chan Data) {
    for data := range in {
        process(data)
    }
}

// ❌ Bad: Share memory with mutex
type SharedData struct {
    mu   sync.Mutex
    data []Data
}
```

### Fixed Race Condition: Debounce Map

**Before (Race Condition):**

```go
func (w *Watcher) handleFileChange(path string) {
    timer := time.AfterFunc(500*time.Millisecond, func() {
        // Race: debounceMap accessed without lock!
        delete(w.debounceMap, path)
        w.processFile(path)
    })

    w.mu.Lock()
    w.debounceMap[path] = timer
    w.mu.Unlock()
}
```

**After (Fixed):**

```go
type debounceEntry struct {
    timer      *time.Timer
    generation uint64
}

func (w *Watcher) handleFileChange(path string) {
    w.mu.Lock()

    // Cancel existing timer
    if entry, exists := w.debounceMap[path]; exists {
        entry.timer.Stop()
    }

    // Create new generation
    generation := w.generation
    w.generation++

    timer := time.AfterFunc(500*time.Millisecond, func() {
        w.mu.Lock()
        defer w.mu.Unlock()

        // Check if this generation is still valid
        if entry, exists := w.debounceMap[path]; exists && entry.generation == generation {
            delete(w.debounceMap, path)
            w.mu.Unlock() // Unlock before processing
            w.processFile(path)
            w.mu.Lock()
        }
    })

    w.debounceMap[path] = debounceEntry{
        timer:      timer,
        generation: generation,
    }

    w.mu.Unlock()
}
```

## Deadlock Avoidance

### Rules to Prevent Deadlocks

1. **Always acquire locks in the same order:**

```go
// ✅ Good: Consistent lock ordering
func transfer(from, to *Account, amount int) {
    // Always lock lower ID first
    first, second := from, to
    if from.ID > to.ID {
        first, second = to, from
    }

    first.mu.Lock()
    defer first.mu.Unlock()

    second.mu.Lock()
    defer second.mu.Unlock()

    from.balance -= amount
    to.balance += amount
}
```

2. **Never call unknown functions while holding a lock:**

```go
// ❌ Bad: Callback might try to acquire the same lock
func (s *Service) ProcessWithCallback(callback func()) {
    s.mu.Lock()
    defer s.mu.Unlock()

    callback() // Dangerous! Callback might call back into Service
}

// ✅ Good: Release lock before callback
func (s *Service) ProcessWithCallback(callback func()) {
    s.mu.Lock()
    data := s.prepareData()
    s.mu.Unlock()

    callback() // Safe: lock not held
}
```

3. **Use timeouts for lock acquisition when possible:**

```go
func (s *Service) TryLock(timeout time.Duration) bool {
    lockCh := make(chan struct{})

    go func() {
        s.mu.Lock()
        close(lockCh)
    }()

    select {
    case <-lockCh:
        return true
    case <-time.After(timeout):
        return false
    }
}
```

## Testing Concurrent Code

### Pattern: Concurrent Test Runner

```go
func TestConcurrentOperations(t *testing.T) {
    const concurrency = 100
    const iterations = 1000

    service := NewService()
    var wg sync.WaitGroup
    errors := make(chan error, concurrency)

    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            for j := 0; j < iterations; j++ {
                if err := service.Operation(id, j); err != nil {
                    errors <- err
                    return
                }
            }
        }(i)
    }

    wg.Wait()
    close(errors)

    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

### Pattern: Race Detector in Tests

```go
func TestRaceCondition(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping race condition test in short mode")
    }

    counter := &Counter{}
    done := make(chan bool)

    // Start multiple goroutines incrementing counter
    for i := 0; i < 100; i++ {
        go func() {
            for j := 0; j < 1000; j++ {
                counter.Increment()
            }
            done <- true
        }()
    }

    // Wait for all goroutines
    for i := 0; i < 100; i++ {
        <-done
    }

    expected := int64(100 * 1000)
    if counter.Get() != expected {
        t.Errorf("Expected %d, got %d", expected, counter.Get())
    }
}

// Run with: go test -race
```

## Best Practices Summary

### ✅ DO

- Always use `defer` for mutex unlocks
- Use contexts for cancellation and timeouts
- Clean up all goroutines on shutdown
- Use `sync.WaitGroup` to wait for goroutines
- Run tests with `-race` flag
- Use channels for communication between goroutines
- Recover from panics in goroutines
- Document lock ordering requirements

### ❌ DON'T

- Launch goroutines without cleanup mechanism
- Hold locks during I/O or network calls
- Access shared state without synchronization
- Ignore context cancellation
- Use `time.Sleep()` for synchronization in tests
- Call external functions while holding locks
- Share memory when channels would work better
- Mix channel and mutex patterns for the same data

## Real-World Examples in Catalogizer

### Media Watcher (Fixed Race Condition)

**Location:** `internal/media/realtime/watcher.go`

Uses generation counters to prevent race conditions in debounced file events.

### File Scanner Worker Pool

**Location:** `internal/media/scanner/scanner.go`

Implements bounded concurrency for file system scanning using semaphore pattern.

### WebSocket Event Bus

**Location:** `internal/media/realtime/event_bus.go`

Uses channels and fan-out pattern to broadcast events to multiple WebSocket clients.

### Circuit Breaker (SMB Client)

**Location:** `internal/smb/circuit_breaker.go`

Implements concurrent state management for circuit breaker pattern with atomic operations.

## BoundedSemaphore Pattern

### Overview

The `BoundedSemaphore` in `internal/concurrency/semaphore.go` extends the basic channel-based semaphore pattern (shown above in [Bounded Concurrency with Semaphore](#pattern-bounded-concurrency-with-semaphore)) with context-aware acquisition. This prevents goroutine leaks when the system is under heavy load or shutting down.

### Implementation

```go
// internal/concurrency/semaphore.go
type BoundedSemaphore struct {
    sem chan struct{}
}

func NewBoundedSemaphore(limit int) *BoundedSemaphore {
    return &BoundedSemaphore{sem: make(chan struct{}, limit)}
}

// Acquire blocks until a slot is available or the context is cancelled.
// Returns ctx.Err() if the context expires before a slot opens.
func (s *BoundedSemaphore) Acquire(ctx context.Context) error {
    select {
    case s.sem <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// TryAcquire attempts to acquire without blocking.
// Returns true if a slot was available, false otherwise.
func (s *BoundedSemaphore) TryAcquire() bool {
    select {
    case s.sem <- struct{}{}:
        return true
    default:
        return false
    }
}

func (s *BoundedSemaphore) Release() {
    <-s.sem
}
```

### Usage in SearchAll

The media entity service's `SearchAll` function uses `BoundedSemaphore` to limit concurrent metadata provider queries:

```go
func (s *EntityService) SearchAll(ctx context.Context, query string) ([]Result, error) {
    sem := concurrency.NewBoundedSemaphore(5) // Max 5 concurrent provider calls
    var mu sync.Mutex
    var results []Result
    var wg sync.WaitGroup

    for _, provider := range s.providers {
        wg.Add(1)
        go func(p MetadataProvider) {
            defer wg.Done()
            if err := sem.Acquire(ctx); err != nil {
                return // Context cancelled, stop gracefully
            }
            defer sem.Release()

            res, err := p.Search(ctx, query)
            if err != nil {
                return // Provider failure does not block others
            }
            mu.Lock()
            results = append(results, res...)
            mu.Unlock()
        }(provider)
    }

    wg.Wait()
    return results, nil
}
```

**Key benefits over the basic semaphore:**
- Context cancellation prevents goroutine pile-up during shutdown.
- `TryAcquire()` enables non-blocking "best-effort" patterns for optional work.
- The pattern prevents API rate limit exhaustion by capping concurrent outbound requests.

### Where Used

| Location | Limit | Purpose |
|----------|-------|---------|
| `SearchAll` (entity service) | 5 | Concurrent metadata provider searches |
| File scanning worker pool | 10 | Concurrent filesystem operations |
| Thumbnail generation | 3 | Concurrent image processing |

## Default Query Timeout on Database Wrappers

### Problem

Without a default timeout, a runaway SQL query (e.g., a full table scan on a large database or a query blocked by SQLite's single-writer lock) can hold a database connection indefinitely. With a connection pool of 25, just 25 stuck queries exhaust the pool and freeze the entire application.

### Solution

The `database.DB` wrapper applies a default 30-second timeout to every query if the caller has not already set a deadline on the context:

```go
// database/connection.go
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    if _, hasDeadline := ctx.Deadline(); !hasDeadline {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }
    rewritten := db.dialect.Rewrite(query)
    return db.DB.QueryContext(ctx, rewritten, args...)
}
```

The same pattern is applied to `ExecContext` and `QueryRowContext`.

**Design decisions:**
- 30 seconds is generous enough for complex joins across large tables but short enough to prevent connection pool starvation.
- Callers can override by passing a context with a tighter deadline (e.g., health checks use 100ms).
- The timeout produces a clear `context.DeadlineExceeded` error that is logged and returned to the client as a 504 Gateway Timeout.

## Redis Middleware Timeout Pattern

### Problem

The Redis-based rate limiter middleware runs on every authenticated request. If Redis becomes slow or unresponsive, every request blocks waiting for Redis, effectively taking down the entire API.

### Solution: Fail-Open with 500ms Timeout

```go
// middleware/redis_rate_limiter.go
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
        defer cancel()

        key := fmt.Sprintf("ratelimit:%s", c.ClientIP())
        count, err := rdb.Incr(ctx, key).Result()
        if err != nil {
            // Redis unavailable: fail open, allow the request
            c.Next()
            return
        }

        if count == 1 {
            rdb.Expire(ctx, key, window)
        }

        if count > int64(limit) {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Key properties:**
- The 500ms timeout prevents Redis latency from blocking requests.
- Fail-open means Redis outages do not cause a service-wide outage (availability over strictness).
- When Redis recovers, rate limiting resumes automatically with no manual intervention.
- The pattern is also used in the Redis cache middleware for similar reasons.

## Goroutine Lifecycle Management Fixes

Several goroutine lifecycle issues were identified and fixed during the safety improvement pass:

### CacheService Cleanup Goroutine

**Before:** The `CacheService` spawned a background goroutine in `NewCacheService()` for expired entry cleanup but had no way to stop it.

**After:** Added `sync.Once`-guarded `Close()` method with a `done` channel:

```go
type CacheService struct {
    done     chan struct{}
    closeOnce sync.Once
}

func (s *CacheService) Close() {
    s.closeOnce.Do(func() {
        close(s.done)
    })
}

// In the cleanup goroutine:
select {
case <-s.done:
    return
case <-ticker.C:
    s.cleanupExpired()
}
```

### WebSocketHandler Stop

**Before:** The `WebSocketHandler` cleanup goroutine had no shutdown path. Tests calling `server.Close()` before `handler.Stop()` caused `readPump` to block indefinitely.

**After:** Added `sync.Once`-guarded `Stop()` method. Production shutdown in `main.go` now calls `wsHandler.Stop()` before `httpServer.Shutdown()` to unblock `readPump`.

### SyncService and LogManagementService

**Before:** `StartSync()` and `CollectLogs()` returned pointers to internal state, creating shared-pointer races when the caller read the result while the service continued modifying it.

**After:** Both methods now return deep copies of their results, eliminating the shared-pointer race.

## Backup Semaphore Pattern

### Problem

Database backup and restore are expensive, exclusive operations. Running two backups concurrently can corrupt the output file, and restoring while a backup is in progress produces an inconsistent snapshot.

### Solution: Weighted Semaphore with Weight 1

The admin handler uses `semaphore.NewWeighted(1)` to enforce mutual exclusion across all backup/restore operations without holding a mutex for the entire (potentially multi-second) I/O operation.

```go
// internal/handlers/admin_handler.go
type AdminHandler struct {
    backupSem *semaphore.Weighted
    // ...
}

func NewAdminHandler(db *database.DB) *AdminHandler {
    return &AdminHandler{
        backupSem: semaphore.NewWeighted(1),
        // ...
    }
}

func (h *AdminHandler) CreateBackup(c *gin.Context) {
    ctx := c.Request.Context()

    // TryAcquire returns immediately if another backup/restore is running
    if !h.backupSem.TryAcquire(1) {
        c.JSON(http.StatusConflict, gin.H{"error": "another backup operation is in progress"})
        return
    }
    defer h.backupSem.Release(1)

    // Perform VACUUM INTO (may take seconds for large databases)
    // ...
}
```

**Key properties:**
- `TryAcquire` is non-blocking -- the caller gets an immediate 409 Conflict rather than waiting.
- The semaphore is shared between `CreateBackup` and `RestoreBackup`, so they are mutually exclusive.
- Unlike a mutex, `semaphore.Weighted` supports context-aware `Acquire()` for callers that prefer to wait with a timeout rather than fail immediately.

### Where Used

| Operation | Behavior |
|-----------|----------|
| `CreateBackup` | `TryAcquire(1)` -- fail fast with 409 if busy |
| `RestoreBackup` | `TryAcquire(1)` -- fail fast with 409 if busy |

## Service Close() Pattern

### Problem

Services that spawn background goroutines (cache cleanup, file watchers, health check loops) must shut down those goroutines cleanly before the process exits. Without coordinated shutdown, goroutines can leak, hold database connections, or write to closed channels.

### Solution: WaitGroup + Context Cancellation + sync.Once

```go
type Service struct {
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    closeOnce sync.Once
}

func NewService() *Service {
    ctx, cancel := context.WithCancel(context.Background())
    s := &Service{ctx: ctx, cancel: cancel}

    s.wg.Add(1)
    go s.backgroundLoop()

    return s
}

func (s *Service) backgroundLoop() {
    defer s.wg.Done()
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-s.ctx.Done():
            return
        case <-ticker.C:
            s.doWork()
        }
    }
}

func (s *Service) Close() {
    s.closeOnce.Do(func() {
        s.cancel()  // Signal all goroutines
        s.wg.Wait() // Wait for all to finish
    })
}
```

**Key properties:**
- `sync.Once` makes `Close()` idempotent -- safe to call from both `defer` in tests and the shutdown sequence in `main.go`.
- `context.WithCancel` propagates the stop signal to all goroutines and any context-aware operations they call (database queries, HTTP requests).
- `sync.WaitGroup` ensures `Close()` blocks until all goroutines have exited, preventing use-after-close races.
- No `time.Sleep` polling -- goroutines wake immediately when the context is cancelled.

**Services using this pattern:**
- `CacheService` (cleanup goroutine)
- `WebSocketHandler` (cleanup goroutine)
- `ScanService` (progress reporting goroutine)
- `WatcherService` (filesystem event loop)

## Rate Limiter Bucket Cap Pattern

### Problem

The in-memory rate limiter creates a bucket (token counter) for each unique client IP. Without a cap, an attacker sending requests from thousands of spoofed IPs can exhaust server memory by creating unbounded map entries.

### Solution: Maximum 10K Entries with LRU Eviction

```go
type RateLimiter struct {
    mu      sync.Mutex
    buckets map[string]*bucket
    maxSize int // Hard cap: 10,000
}

func (rl *RateLimiter) getBucket(key string) *bucket {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    if b, ok := rl.buckets[key]; ok {
        b.lastAccess = time.Now()
        return b
    }

    // Evict oldest entry if at capacity
    if len(rl.buckets) >= rl.maxSize {
        rl.evictOldest()
    }

    b := &bucket{
        tokens:     rl.limit,
        lastAccess: time.Now(),
        lastRefill: time.Now(),
    }
    rl.buckets[key] = b
    return b
}

func (rl *RateLimiter) evictOldest() {
    var oldestKey string
    var oldestTime time.Time

    for key, b := range rl.buckets {
        if oldestKey == "" || b.lastAccess.Before(oldestTime) {
            oldestKey = key
            oldestTime = b.lastAccess
        }
    }

    if oldestKey != "" {
        delete(rl.buckets, oldestKey)
    }
}
```

**Key properties:**
- The 10K cap bounds memory to approximately 2MB regardless of traffic patterns.
- LRU eviction removes the least recently seen IP, preserving buckets for active legitimate clients.
- The eviction scan is O(n) but runs only when the map is full, which is rare under normal traffic.
- When Redis is available, the Redis-based rate limiter is preferred and this in-memory limiter serves as fallback.

**Configuration:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `maxSize` | 10,000 | Maximum number of tracked client IPs |
| `limit` | 100 | Requests per window per client |
| `window` | 1 minute | Rate limit window duration |

## References

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [The Go Memory Model](https://go.dev/ref/mem)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
