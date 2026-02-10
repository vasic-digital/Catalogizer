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

## References

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [The Go Memory Model](https://go.dev/ref/mem)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
