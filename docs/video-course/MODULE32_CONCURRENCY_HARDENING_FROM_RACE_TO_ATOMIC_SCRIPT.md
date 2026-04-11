# Module 32 — Concurrency Hardening: From Race to Atomic

**Duration:** 22 minutes
**Prerequisites:** Module 15 (Concurrency), Module 25 (Concurrency Hardening)

## Learning objectives

1. Read a `go test -race` report and identify the offending memory access.
2. Choose between `sync.Mutex`, `sync.RWMutex`, `atomic.Int64`, and channels.
3. Understand the `sync.Once` + `WaitGroup` + `lifecycleMu` pattern for goroutine lifecycle.
4. Spot common anti-patterns: reading a closed channel and assuming the producer has exited, unbounded `go func()` in loops, `wg.Add` inside the goroutine.

## Segment 1 — Reading a race report (0:00 – 4:00)

**Show on screen:** example from `catalog-api/tests/stress/responsiveness_test.go` (pre-fix):

```
WARNING: DATA RACE
Read at 0x... by goroutine 7:
  catalogizer/tests/stress.glob..func1()
    responsiveness_test.go:237 +0x...
Previous write at 0x... by goroutine 5:
  catalogizer/tests/stress.glob..func1()
    responsiveness_test.go:225 +0x...
```

Race reports have two parts:
- **Read site** and **write site** — the two accesses that raced.
- **Goroutine IDs** that did each access.

The offending lines were `successCount++` and `errorCount++` on plain `int64` vars modified from concurrent goroutines.

## Segment 2 — Fix: `atomic.Int64` (4:00 – 8:00)

```go
var successCount, errorCount atomic.Int64
// ...
successCount.Add(1)
errorCount.Add(1)
// ...
successRate := float64(successCount.Load()) / float64(totalRequests) * 100
```

**Why atomic** instead of a mutex: these are simple counters with no invariant across multiple fields. Atomics are faster (no lock contention) and the intent is clearer.

## Segment 3 — Real-world example: `SmbConnectionPool` (8:00 – 14:00)

The pre-fix `StopCleanup()` did:

```go
func (p *SmbConnectionPool) StopCleanup() {
    p.mu.Lock()
    // ...
    close(p.cleanupDone)
    p.mu.Unlock()
    <-p.cleanupDone // ← this returns INSTANTLY, doesn't wait!
}
```

Reading a closed channel returns immediately — the cleanup goroutine might still be inside `cleanupIdleConnections()` when `StopCleanup` returns. **Not a wait at all.**

Fix: `sync.WaitGroup` + a separate `lifecycleMu` so concurrent `Start`/`Stop` don't reuse the WaitGroup mid-wait:

```go
type SmbConnectionPool struct {
    mu            sync.RWMutex // data mutex
    lifecycleMu   sync.Mutex   // serializes Start/Stop
    cleanupDone   chan struct{}
    wg            sync.WaitGroup
    isRunning     bool
}

func (p *SmbConnectionPool) StopCleanup() {
    p.lifecycleMu.Lock()
    defer p.lifecycleMu.Unlock()
    // ...
    close(stopCh)
    p.wg.Wait() // ← actually waits
}
```

**Why two mutexes**: `lifecycleMu` serializes `Start`/`Stop` calls, `mu` protects the data. Holding `mu` during `wg.Wait` would deadlock with the cleanup loop that also needs `mu`.

## Segment 4 — `wg.Add` placement (14:00 – 17:00)

Anti-pattern:
```go
go func() {
    wg.Add(1) // ← race with wg.Wait
    defer wg.Done()
    // ...
}()
```

Correct:
```go
wg.Add(1)
go func() {
    defer wg.Done()
    // ...
}()
```

The `Add` must happen **before** the goroutine launch, from the parent goroutine. This is non-negotiable.

## Segment 5 — `sync.Once` for cleanup idempotence (17:00 – 20:00)

```go
type CacheService struct {
    closeOnce sync.Once
    shutdown  chan struct{}
}

func (s *CacheService) Close() {
    s.closeOnce.Do(func() {
        close(s.shutdown)
        // ...
    })
}
```

`Close()` is now safe to call multiple times — second call is a no-op. Essential for cleanup paths that may run from multiple defer statements or shutdown handlers.

## Segment 6 — Non-blocking channel sends (20:00 – 22:00)

Anti-pattern:
```go
for _, entry := range entries {
    channel <- entry // ← blocks forever if receiver stops reading
}
```

Correct (from `log_management_service.streamLogEntries`):
```go
for _, entry := range entries {
    select {
    case <-done:
        return
    case channel <- entry:
    case <-time.After(5 * time.Second):
        return
    }
}
```

Three cases: caller cancelled, send succeeded, receiver stalled. Any of them exits the goroutine cleanly.

## Exercise

Add a test to `catalog-api/smb/types_lifecycle_test.go` that:
1. Starts the pool.
2. Spawns 10 goroutines concurrently calling `StopCleanup` and `StartCleanup`.
3. Asserts no panic and no race under `-race`.

## Assessment

1. What does `-race` cost at runtime? Answer: ~5-10x memory, 2-20x slower. Never run in production.
2. When would you prefer `sync.RWMutex` over `sync.Mutex`? Answer: read-heavy workloads with occasional writes.
3. What happens if you close a channel twice? Answer: panic.
