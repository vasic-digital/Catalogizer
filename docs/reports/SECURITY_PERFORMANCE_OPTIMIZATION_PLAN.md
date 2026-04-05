# SECURITY & PERFORMANCE OPTIMIZATION PLAN
## Memory Safety, Concurrency, Lazy Loading, and Stress Testing

---

## 1. SECURITY OPTIMIZATION

### 1.1 Current Security Status

| Scanner | Status | Findings |
|---------|--------|----------|
| go vet | ✅ PASS | 0 issues |
| govulncheck | ✅ PASS | 0 vulnerabilities |
| npm audit | ✅ PASS | 0 vulnerabilities |
| Trivy | ✅ PASS | 0 HIGH/CRITICAL |
| SonarQube | ⚠️ WARN | 38 bugs, 105 hotspots |

### 1.2 SonarQube Issues Resolution

#### Bug Category: Accessibility (30 issues)
**Pattern:** Non-interactive elements with click handlers

```typescript
// BEFORE (Problematic)
<div onClick={handleClick}>Some content</div>

// AFTER (Fixed)
<div 
  onClick={handleClick}
  onKeyDown={(e) => e.key === 'Enter' && handleClick()}
  tabIndex={0}
  role="button"
>
  Some content
</div>
```

**Automated Fix Script:**
```bash
#!/bin/bash
# scripts/fix-accessibility.sh

# Find all non-interactive elements with onClick
grep -rn "onClick" catalog-web/src --include="*.tsx" | \
  grep -v "onKeyDown\|tabIndex\|role=" | \
  while read line; do
    echo "Accessibility issue: $line"
  done
```

#### Security Hotspots: False Positives (105 issues)
**Categories:**
1. Hard-coded passwords in test files (60) - Acceptable
2. Registration form validators (30) - False positive
3. Mock data (15) - Acceptable

**Action:** Mark as "Won't Fix" in SonarQube with justification.

#### Conditional Logic Issue (1 issue)
**File:** `services/configuration_wizard_service.go:1013`

```go
// BEFORE
if storageType == "smb" {
    // SMB test logic
} else if storageType == "ftp" {
    // FTP test logic
} else {
    log.Printf("Storage type %s test not implemented", storageType)
}

// AFTER
func (s *ConfigurationWizardService) testStorageConnection(storageType string, config StorageConfig) error {
    switch storageType {
    case "smb":
        return s.testSMBConnection(config)
    case "ftp":
        return s.testFTPConnection(config)
    case "webdav":
        return s.testWebDAVConnection(config)
    case "nfs":
        return s.testNFSConnection(config)
    case "local":
        return s.testLocalConnection(config)
    default:
        return fmt.Errorf("unsupported storage type: %s", storageType)
    }
}
```

### 1.3 Security Enhancements

#### Input Validation Hardening
```go
// middleware/input_validation.go

func ValidateFilename(filename string) error {
    // Prevent path traversal
    if strings.Contains(filename, "..") {
        return errors.New("invalid filename: path traversal detected")
    }
    
    // Prevent null bytes
    if strings.ContainsRune(filename, '\x00') {
        return errors.New("invalid filename: null byte detected")
    }
    
    // Validate length
    if len(filename) > 255 {
        return errors.New("filename too long")
    }
    
    // Validate characters
    validFilename := regexp.MustCompile(`^[a-zA-Z0-9._\-\s]+$`)
    if !validFilename.MatchString(filename) {
        return errors.New("filename contains invalid characters")
    }
    
    return nil
}
```

#### Rate Limiting Enhancement
```go
// internal/middleware/rate_limiter.go

type AdaptiveRateLimiter struct {
    baseLimit    int
    currentLimit int
    windowSize   time.Duration
    mu           sync.RWMutex
}

func (rl *AdaptiveRateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        
        // Check current load
        rl.mu.RLock()
        limit := rl.currentLimit
        rl.mu.RUnlock()
        
        // Check request count
        count := getRequestCount(ip, rl.windowSize)
        if count >= limit {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
                "retry_after": rl.windowSize.Seconds(),
            })
            c.Abort()
            return
        }
        
        incrementRequestCount(ip)
        c.Next()
    }
}

func (rl *AdaptiveRateLimiter) AdjustBasedOnLoad() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        load := getSystemLoad()
        rl.mu.Lock()
        if load > 0.8 {
            rl.currentLimit = int(float64(rl.baseLimit) * 0.5)
        } else if load < 0.3 {
            rl.currentLimit = int(float64(rl.baseLimit) * 1.5)
        } else {
            rl.currentLimit = rl.baseLimit
        }
        rl.mu.Unlock()
    }
}
```

---

## 2. MEMORY SAFETY

### 2.1 Memory Leak Detection

#### Current Test Infrastructure
**File:** `internal/services/leak_test.go`

```go
func TestNoMemoryLeaks(t *testing.T) {
    // Get initial memory stats
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Run operations
    for i := 0; i < 1000; i++ {
        svc := NewService(testConfig)
        svc.Process(context.Background(), testItem)
        svc.Close()
    }
    
    // Force garbage collection
    runtime.GC()
    
    // Get final memory stats
    var m2 runtime.MemStats
    runtime.ReadMemStats(&m2)
    
    // Allow 10% overhead
    maxAllowed := uint64(float64(m1.Alloc) * 1.1)
    assert.LessOrEqual(t, m2.Alloc, maxAllowed, 
        "Potential memory leak detected")
}
```

#### Enhanced Memory Testing
```go
// internal/services/memory_test.go

func TestMemoryUnderLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory test in short mode")
    }
    
    // Start memory profiling
    f, err := os.Create("memprofile.out")
    require.NoError(t, err)
    defer f.Close()
    
    // Run sustained load
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    // Simulate workload
                    processRandomMedia()
                }
            }
        }()
    }
    wg.Wait()
    
    // Write heap profile
    pprof.WriteHeapProfile(f)
}
```

### 2.2 Goroutine Leak Prevention

```go
// internal/recovery/goroutine_leak_test.go

func TestNoGoroutineLeaks(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // Run operations that spawn goroutines
    for i := 0; i < 100; i++ {
        ctx, cancel := context.WithCancel(context.Background())
        go func() {
            <-ctx.Done()
        }()
        cancel()
    }
    
    // Wait for cleanup
    time.Sleep(100 * time.Millisecond)
    
    after := runtime.NumGoroutine()
    
    // Allow small variance
    assert.LessOrEqual(t, after, before+2, 
        "Goroutine leak detected: before=%d, after=%d", before, after)
}
```

### 2.3 Proper Resource Cleanup

```go
// services/universal_scanner.go

type UniversalScanner struct {
    workerPool chan struct{}
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
    closed     bool
    mu         sync.Mutex
}

func (s *UniversalScanner) Close() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.closed {
        return nil
    }
    s.closed = true
    
    // Cancel all operations
    s.cancel()
    
    // Wait for all workers to finish
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(30 * time.Second):
        return errors.New("timeout waiting for workers to finish")
    }
}

func (s *UniversalScanner) scan(path string) error {
    s.mu.Lock()
    if s.closed {
        s.mu.Unlock()
        return errors.New("scanner is closed")
    }
    s.wg.Add(1)
    s.mu.Unlock()
    
    defer s.wg.Done()
    
    // Acquire worker slot
    select {
    case s.workerPool <- struct{}{}:
        defer func() { <-s.workerPool }()
    case <-s.ctx.Done():
        return s.ctx.Err()
    }
    
    // Do scanning work
    return s.doScan(path)
}
```

---

## 3. CONCURRENCY OPTIMIZATION

### 3.1 Race Condition Prevention

#### Known Issue: SMB Timing
**File:** `internal/smb/resilience.go`

```go
// BEFORE (Race condition)
type SMBClient struct {
    conn     *smb2.Share
    connMu   sync.RWMutex
    lastUsed time.Time
}

func (c *SMBClient) Ping() error {
    c.connMu.RLock()
    conn := c.conn
    c.connMu.RUnlock()
    
    // Race: conn could be closed between unlock and use
    _, err := conn.Stat(".")
    return err
}

// AFTER (Fixed)
func (c *SMBClient) Ping() error {
    c.connMu.Lock()
    defer c.connMu.Unlock()
    
    if c.conn == nil {
        return errors.New("connection closed")
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    done := make(chan error, 1)
    go func() {
        _, err := c.conn.Stat(".")
        done <- err
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 3.2 Semaphore Implementation

```go
// pkg/semaphore/semaphore.go

package semaphore

import (
    "context"
    "sync"
)

type Semaphore struct {
    ch chan struct{}
}

func New(maxConcurrent int) *Semaphore {
    return &Semaphore{
        ch: make(chan struct{}, maxConcurrent),
    }
}

func (s *Semaphore) Acquire(ctx context.Context) error {
    select {
    case s.ch <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (s *Semaphore) Release() {
    <-s.ch
}

func (s *Semaphore) TryAcquire() bool {
    select {
    case s.ch <- struct{}{}:
        return true
    default:
        return false
    }
}

// Usage example
func (s *UniversalScanner) ScanAll(paths []string) error {
    sem := semaphore.New(runtime.NumCPU() * 2)
    
    var wg sync.WaitGroup
    errCh := make(chan error, len(paths))
    
    for _, path := range paths {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            
            if err := sem.Acquire(context.Background()); err != nil {
                errCh <- err
                return
            }
            defer sem.Release()
            
            if err := s.Scan(p); err != nil {
                errCh <- err
            }
        }(path)
    }
    
    wg.Wait()
    close(errCh)
    
    // Return first error
    return <-errCh
}
```

### 3.3 Non-Blocking Operations

```go
// services/nonblocking.go

type NonBlockingScanner struct {
    scanner   *UniversalScanner
    queue     chan scanRequest
    results   chan scanResult
    workers   int
    ctx       context.Context
    cancel    context.CancelFunc
}

type scanRequest struct {
    path    string
    respond chan scanResult
}

type scanResult struct {
    path string
    err  error
}

func NewNonBlockingScanner(scanner *UniversalScanner, workers int) *NonBlockingScanner {
    ctx, cancel := context.WithCancel(context.Background())
    s := &NonBlockingScanner{
        scanner: scanner,
        queue:   make(chan scanRequest, workers*10),
        results: make(chan scanResult, workers*10),
        workers: workers,
        ctx:     ctx,
        cancel:  cancel,
    }
    
    for i := 0; i < workers; i++ {
        go s.worker()
    }
    
    return s
}

func (s *NonBlockingScanner) worker() {
    for {
        select {
        case <-s.ctx.Done():
            return
        case req := <-s.queue:
            err := s.scanner.Scan(req.path)
            req.respond <- scanResult{path: req.path, err: err}
        }
    }
}

// Non-blocking scan - returns channel for result
func (s *NonBlockingScanner) ScanAsync(path string) <-chan scanResult {
    respond := make(chan scanResult, 1)
    
    select {
    case s.queue <- scanRequest{path: path, respond: respond}:
    default:
        // Queue full, return error immediately
        respond <- scanResult{path: path, err: errors.New("queue full")}
    }
    
    return respond
}

func (s *NonBlockingScanner) Close() {
    s.cancel()
}
```

---

## 4. LAZY LOADING IMPLEMENTATION

### 4.1 Service Lazy Initialization

```go
// services/lazy_service.go

type LazyService struct {
    once     sync.Once
    service  *ExpensiveService
    initErr  error
    initFunc func() (*ExpensiveService, error)
}

func NewLazyService(initFunc func() (*ExpensiveService, error)) *LazyService {
    return &LazyService{
        initFunc: initFunc,
    }
}

func (l *LazyService) Get() (*ExpensiveService, error) {
    l.once.Do(func() {
        l.service, l.initErr = l.initFunc()
    })
    return l.service, l.initErr
}

// Usage
var analyticsService = NewLazyService(func() (*AnalyticsService, error) {
    config, err := loadAnalyticsConfig()
    if err != nil {
        return nil, err
    }
    return NewAnalyticsService(config), nil
})
```

### 4.2 Asset Lazy Loading (Already Implemented)
**File:** `challenges/asset_lazy_loading.go`

The asset lazy loading system is already implemented and tested. Key features:
- Assets loaded on first request
- Caching after initial load
- Background preloading for predicted assets

### 4.3 Database Connection Lazy Loading

```go
// database/lazy_pool.go

type LazyConnectionPool struct {
    once     sync.Once
    pool     *sql.DB
    config   DatabaseConfig
    initErr  error
}

func NewLazyConnectionPool(config DatabaseConfig) *LazyConnectionPool {
    return &LazyConnectionPool{config: config}
}

func (l *LazyConnectionPool) Get() (*sql.DB, error) {
    l.once.Do(func() {
        dsn := buildDSN(l.config)
        l.pool, l.initErr = sql.Open("sqlite3", dsn)
        if l.initErr != nil {
            return
        }
        l.pool.SetMaxOpenConns(l.config.MaxConnections)
        l.pool.SetMaxIdleConns(l.config.MaxIdle)
        l.pool.SetConnMaxLifetime(l.config.ConnLifetime)
    })
    return l.pool, l.initErr
}

func (l *LazyConnectionPool) Close() error {
    if l.pool != nil {
        return l.pool.Close()
    }
    return nil
}
```

---

## 5. STRESS TESTING FRAMEWORK

### 5.1 API Load Tests

```go
// tests/stress/api_load_test.go

package stress

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

type LoadTestConfig struct {
    ConcurrentUsers    int
    RequestsPerUser    int
    RampUpDuration     time.Duration
    Timeout            time.Duration
    TargetErrorRate    float64
    TargetResponseTime time.Duration
}

func TestAPILoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    config := LoadTestConfig{
        ConcurrentUsers:    100,
        RequestsPerUser:    50,
        RampUpDuration:     10 * time.Second,
        Timeout:            30 * time.Second,
        TargetErrorRate:    0.01,  // 1%
        TargetResponseTime: 200 * time.Millisecond,
    }
    
    results := runLoadTest(t, config)
    
    t.Logf("Load Test Results:")
    t.Logf("  Total Requests: %d", results.TotalRequests)
    t.Logf("  Successful: %d", results.Successful)
    t.Logf("  Failed: %d", results.Failed)
    t.Logf("  Error Rate: %.2f%%", results.ErrorRate*100)
    t.Logf("  Avg Response Time: %v", results.AvgResponseTime)
    t.Logf("  P95 Response Time: %v", results.P95ResponseTime)
    t.Logf("  P99 Response Time: %v", results.P99ResponseTime)
    
    if results.ErrorRate > config.TargetErrorRate {
        t.Errorf("Error rate %.2f%% exceeds target %.2f%%",
            results.ErrorRate*100, config.TargetErrorRate*100)
    }
    
    if results.P95ResponseTime > config.TargetResponseTime {
        t.Errorf("P95 response time %v exceeds target %v",
            results.P95ResponseTime, config.TargetResponseTime)
    }
}

func runLoadTest(t *testing.T, config LoadTestConfig) *LoadTestResults {
    client := &http.Client{
        Timeout: config.Timeout,
        Transport: &http.Transport{
            MaxIdleConns:        config.ConcurrentUsers,
            MaxIdleConnsPerHost: config.ConcurrentUsers,
        },
    }
    
    results := &LoadTestResults{
        ResponseTimes: make([]time.Duration, 0, config.ConcurrentUsers*config.RequestsPerUser),
    }
    
    var wg sync.WaitGroup
    var successCount, failCount int64
    
    // Ramped user spawn
    userInterval := config.RampUpDuration / time.Duration(config.ConcurrentUsers)
    
    for i := 0; i < config.ConcurrentUsers; i++ {
        time.Sleep(userInterval)
        
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            
            for j := 0; j < config.RequestsPerUser; j++ {
                start := time.Now()
                
                req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/media", nil)
                req.Header.Set("Authorization", "Bearer "+testToken)
                
                resp, err := client.Do(req)
                duration := time.Since(start)
                
                results.mu.Lock()
                results.ResponseTimes = append(results.ResponseTimes, duration)
                results.mu.Unlock()
                
                if err != nil || resp.StatusCode >= 400 {
                    atomic.AddInt64(&failCount, 1)
                    if resp != nil {
                        resp.Body.Close()
                    }
                    continue
                }
                
                atomic.AddInt64(&successCount, 1)
                resp.Body.Close()
            }
        }(i)
    }
    
    wg.Wait()
    
    results.TotalRequests = successCount + failCount
    results.Successful = successCount
    results.Failed = failCount
    results.ErrorRate = float64(failCount) / float64(results.TotalRequests)
    results.calculateStats()
    
    return results
}
```

### 5.2 Database Stress Tests

```go
// tests/stress/database_stress_test.go

func TestDatabaseConcurrency(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    // Test concurrent writes
    t.Run("ConcurrentWrites", func(t *testing.T) {
        var wg sync.WaitGroup
        errors := make(chan error, 1000)
        
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func(id int) {
                defer wg.Done()
                
                for j := 0; j < 10; j++ {
                    _, err := db.Exec(
                        "INSERT INTO media_items (id, title, type) VALUES (?, ?, ?)",
                        fmt.Sprintf("item-%d-%d", id, j),
                        fmt.Sprintf("Title %d-%d", id, j),
                        "movie",
                    )
                    if err != nil {
                        errors <- err
                    }
                }
            }(i)
        }
        
        wg.Wait()
        close(errors)
        
        var errCount int
        for range errors {
            errCount++
        }
        
        assert.Less(t, errCount, 10, "Too many write errors")
    })
    
    // Test concurrent reads
    t.Run("ConcurrentReads", func(t *testing.T) {
        var wg sync.WaitGroup
        var totalLatency int64
        
        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                start := time.Now()
                rows, err := db.Query("SELECT * FROM media_items LIMIT 100")
                if err != nil {
                    return
                }
                defer rows.Close()
                
                for rows.Next() {
                    // Scan rows
                }
                
                atomic.AddInt64(&totalLatency, int64(time.Since(start)))
            }()
        }
        
        wg.Wait()
        
        avgLatency := time.Duration(totalLatency / 100)
        assert.Less(t, avgLatency, 100*time.Millisecond, "Read latency too high")
    })
}
```

### 5.3 Memory Stress Tests

```go
// tests/stress/memory_stress_test.go

func TestLargeFileHandling(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    // Create a large test file (100MB)
    largeFile := createLargeTestFile(t, 100*1024*1024)
    defer os.Remove(largeFile)
    
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Process the file
    processor := NewFileProcessor(testConfig)
    err := processor.Process(context.Background(), largeFile)
    require.NoError(t, err)
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Memory should not grow significantly
    memGrowth := m2.Alloc - m1.Alloc
    maxAllowed := uint64(50 * 1024 * 1024) // 50MB
    
    assert.Less(t, memGrowth, maxAllowed,
        "Memory growth %d bytes exceeds limit %d bytes", memGrowth, maxAllowed)
}

func TestSustainedLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Run for 5 minutes
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    // Simulate workload
                    processMediaItem(generateTestMedia())
                }
            }
        }()
    }
    wg.Wait()
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Check for memory leak (should not grow more than 2x)
    maxAllowed := m1.Alloc * 2
    assert.Less(t, m2.Alloc, maxAllowed,
        "Potential memory leak: memory grew from %d to %d bytes", m1.Alloc, m2.Alloc)
}
```

---

## 6. MONITORING INTEGRATION

### 6.1 Prometheus Metrics

```go
// internal/metrics/performance.go

var (
    // Request metrics
    httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "catalogizer_http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "path", "status"})
    
    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "catalogizer_http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    }, []string{"method", "path"})
    
    // Scanner metrics
    scanDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "catalogizer_scan_duration_seconds",
        Help:    "File scan duration in seconds",
        Buckets: []float64{.1, .5, 1, 5, 10, 30, 60, 120, 300},
    }, []string{"protocol", "storage_root"})
    
    concurrentScans = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "catalogizer_concurrent_scans",
        Help: "Number of concurrent scans",
    })
    
    // Memory metrics
    memoryUsage = promauto.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "catalogizer_memory_bytes",
        Help: "Current memory usage in bytes",
    }, func() float64 {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        return float64(m.Alloc)
    })
    
    goroutineCount = promauto.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "catalogizer_goroutines",
        Help: "Number of goroutines",
    }, func() float64 {
        return float64(runtime.NumGoroutine())
    })
)
```

### 6.2 Health Check Endpoints

```go
// internal/handlers/health.go

func (h *HealthHandler) Liveness(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "alive",
        "time":   time.Now().UTC(),
    })
}

func (h *HealthHandler) Readiness(c *gin.Context) {
    checks := map[string]bool{
        "database": h.checkDatabase(),
        "storage":  h.checkStorage(),
        "cache":    h.checkCache(),
    }
    
    allHealthy := true
    for _, ok := range checks {
        if !ok {
            allHealthy = false
            break
        }
    }
    
    status := http.StatusOK
    if !allHealthy {
        status = http.StatusServiceUnavailable
    }
    
    c.JSON(status, gin.H{
        "status": map[bool]string{true: "ready", false: "not_ready"}[allHealthy],
        "checks": checks,
        "time":   time.Now().UTC(),
    })
}

func (h *HealthHandler) checkDatabase() bool {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    return h.db.PingContext(ctx) == nil
}
```

---

## 7. IMPLEMENTATION TIMELINE

| Week | Task | Deliverable |
|------|------|-------------|
| 9-1 | Race condition fixes | Zero race conditions |
| 9-2 | Memory leak tests | All leak tests passing |
| 9-3 | Semaphore implementation | Concurrent operations |
| 9-4 | Lazy loading implementation | All services lazy |
| 10-1 | Non-blocking operations | Async API handlers |
| 10-2 | API stress tests | 100 concurrent users |
| 10-3 | Database stress tests | Concurrent read/write |
| 10-4 | Memory stress tests | Sustained load passing |

---

*Document Generated: 2026-02-27*
*Status: Implementation Ready*
