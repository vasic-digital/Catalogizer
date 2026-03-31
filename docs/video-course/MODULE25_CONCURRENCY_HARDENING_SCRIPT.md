# Module 25: Concurrency Hardening

## Video Script — Concurrency Hardening & Memory Safety

### Duration: ~25 minutes

---

### Scene 1: Introduction (2 min)

"In this module, we'll deep-dive into the concurrency hardening patterns used in Catalogizer to prevent goroutine leaks, memory leaks, race conditions, and deadlocks. These patterns ensure the system remains stable under sustained load — even when handling thousands of concurrent file system operations."

**Key Topics:**
- Debounce map bounding with TTL and max-size eviction
- Active scan session cleanup with deferred goroutine + timer
- IP rate limiter bucket bounding with LRU eviction
- Lock ordering documentation for multi-mutex structs
- WaitGroup + context cancellation patterns for goroutine lifecycle

---

### Scene 2: Debounce Map Bounding (5 min)

**File:** `internal/media/realtime/watcher.go`

"The SMBChangeWatcher uses debounce entries to coalesce rapid filesystem events. Without bounding, this map grows indefinitely under high file churn."

**Show code:**
- `debounceEntry` struct with timer + generation counter
- Max size check at 10,000 entries
- LRU eviction to 5,000 when threshold exceeded
- Generation counter prevents stale timer callbacks from deleting newer entries

**Diagram:** Debounce flow — event arrives → check map → cancel old timer → set new timer → on fire: delete entry + enqueue

---

### Scene 3: Active Scan Cleanup (5 min)

**File:** `internal/services/universal_scanner.go`

"Each scan job creates a ScanStatus entry in the activeScans map. Without cleanup, completed scans accumulate forever."

**Show code:**
- `processScanJob()` deferred goroutine that waits 60 seconds then deletes the entry
- Goroutine tracked by WaitGroup for clean shutdown
- `select` on timer or stopCh for graceful termination

**Key pattern:** Deferred cleanup goroutine with WaitGroup tracking

---

### Scene 4: IP Bucket Rate Limiting (5 min)

**File:** `middleware/request.go`

"Token-bucket rate limiting per client IP. Under DDoS, the map can grow to millions of IPs."

**Show code:**
- `maxBuckets = 10000` constant
- Periodic cleanup goroutine (every 5 minutes)
- TTL eviction: delete entries older than 10 minutes
- Size eviction: sort by lastCheck, evict oldest half

**Pattern:** Bounded map with TTL + LRU eviction

---

### Scene 5: Goroutine Lifecycle Management (5 min)

**Files:** `services/sync_service.go`, `services/error_reporting_service.go`, `services/log_management_service.go`

"Every goroutine spawned by a service must be tracked and terminated cleanly."

**Show pattern:**
```go
type Service struct {
    wg     sync.WaitGroup
    ctx    context.Context
    cancel context.CancelFunc
}

func (s *Service) Close() {
    s.cancel()
    s.wg.Wait()
}

func (s *Service) doWork() {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        // work with s.ctx
    }()
}
```

**Key principles:**
1. Every `go func()` must have a corresponding `wg.Add(1)` / `defer wg.Done()`
2. Every service must have a `Close()` or `Stop()` that cancels context and waits
3. Return copies from goroutines, not shared pointers

---

### Scene 6: Lock Ordering & Race Prevention (3 min)

**File:** `internal/services/universal_scanner.go`

"When a struct has multiple mutexes, acquire them in documented order to prevent deadlocks."

**Show:** UniversalScanner with 3 mutexes:
1. `mu` (general state) — acquired first
2. `protocolScannersMu` (scanner registry) — acquired second
3. `activeScansMu` (active scans) — acquired third

**Rule:** Always document lock ordering in comments. Use `-race` flag in all tests.

---

### Scene 7: Testing Concurrency Safety (3 min)

"Run all tests with the `-race` flag to detect data races at runtime."

```bash
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -race -count=1
```

**Stress tests:** `tests/stress/` — 54+ tests that hammer concurrent paths.

**Goroutine leak detection:** `CH-092` challenge verifies goroutine count returns to baseline.

---

### Summary

- Bound all maps: max size + TTL eviction
- Track all goroutines: WaitGroup + context cancellation
- Clean up completed work: deferred cleanup goroutines
- Document lock ordering for multi-mutex structs
- Test with `-race` flag always
