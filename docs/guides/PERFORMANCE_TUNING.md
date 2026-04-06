# Performance Tuning Guide

**Document Version:** 1.0  
**Last Updated:** April 6, 2026  
**Applies to:** Catalogizer v2.2.0+  

---

## Table of Contents

1. [Overview](#1-overview)
2. [Database Optimization](#2-database-optimization)
3. [API Performance](#3-api-performance)
4. [Caching Strategies](#4-caching-strategies)
5. [Connection Pooling](#5-connection-pooling)
6. [Memory Management](#6-memory-management)
7. [Monitoring & Profiling](#7-monitoring--profiling)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. Overview

### 1.1 Performance Targets

| Metric | Target | Critical Threshold |
|--------|--------|-------------------|
| API Response (P95) | < 200ms | > 500ms |
| Database Query (P95) | < 50ms | > 200ms |
| File Scan Speed | > 1000 files/min | < 500 files/min |
| Memory Usage | < 4GB | > 8GB |
| CPU Usage | < 70% | > 90% |
| Concurrent Users | 1000+ | < 500 |

### 1.2 Performance Bottlenecks

Common bottlenecks in Catalogizer:
1. **Database queries** - Unoptimized SQL, missing indexes
2. **File scanning** - Sequential processing, blocking I/O
3. **Memory leaks** - Unclosed resources, goroutine leaks
4. **Network latency** - Protocol overhead, large payloads
5. **Lock contention** - Database locks, mutex contention

---

## 2. Database Optimization

### 2.1 Index Optimization

```sql
-- Essential indexes for performance

-- Media item lookups
CREATE INDEX idx_media_items_title ON media_items(title);
CREATE INDEX idx_media_items_type ON media_items(media_type);
CREATE INDEX idx_media_items_created ON media_items(created_at);
CREATE INDEX idx_media_items_title_type ON media_items(title, media_type);

-- Full-text search
CREATE INDEX idx_media_items_fts ON media_items USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- File lookups
CREATE INDEX idx_files_path ON files(storage_root_id, path);
CREATE INDEX idx_files_media_item ON files(media_item_id);

-- User queries
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- Collection queries
CREATE INDEX idx_collection_items_collection ON collection_items(collection_id);
CREATE INDEX idx_collection_items_media ON collection_items(media_item_id);

-- Scan performance
CREATE INDEX idx_files_scanned_at ON files(scanned_at);
CREATE INDEX idx_files_storage_root ON files(storage_root_id);
```

### 2.2 Query Optimization

```go
// Before: N+1 query problem
collections, _ := repo.GetCollections()
for _, c := range collections {
    items, _ := repo.GetCollectionItems(c.ID) // N queries
}

// After: Single query with JOIN
func (r *Repository) GetCollectionsWithItems() ([]CollectionWithItems, error) {
    query := `
        SELECT c.*, ci.media_item_id
        FROM collections c
        LEFT JOIN collection_items ci ON c.id = ci.collection_id
        WHERE c.deleted_at IS NULL
    `
    // Process results with proper grouping
}
```

### 2.3 Batch Operations

```go
// Batch insert for better performance
func (r *Repository) BatchInsertFiles(files []File) error {
    const batchSize = 1000
    
    for i := 0; i < len(files); i += batchSize {
        end := i + batchSize
        if end > len(files) {
            end = len(files)
        }
        
        batch := files[i:end]
        if err := r.insertBatch(batch); err != nil {
            return fmt.Errorf("batch insert failed at offset %d: %w", i, err)
        }
    }
    return nil
}

func (r *Repository) insertBatch(files []File) error {
    query := "INSERT INTO files (path, size, modified_at) VALUES "
    placeholders := make([]string, len(files))
    args := make([]interface{}, 0, len(files)*3)
    
    for i, f := range files {
        placeholders[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
        args = append(args, f.Path, f.Size, f.ModifiedAt)
    }
    
    query += strings.Join(placeholders, ", ")
    _, err := r.db.Exec(query, args...)
    return err
}
```

### 2.4 Connection Pool Tuning

```go
// database/connection.go
import (
    "database/sql"
    "time"
)

type Config struct {
    MaxOpenConns    int           // Maximum open connections
    MaxIdleConns    int           // Maximum idle connections
    ConnMaxLifetime time.Duration // Maximum lifetime of a connection
    ConnMaxIdleTime time.Duration // Maximum idle time
}

func NewPool(connStr string, cfg Config) (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    
    // Connection pool settings
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
    
    return db, nil
}

// Recommended settings
var ProductionConfig = Config{
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 3 * time.Minute,
}

var HighLoadConfig = Config{
    MaxOpenConns:    50,
    MaxIdleConns:    20,
    ConnMaxLifetime: 10 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
}
```

---

## 3. API Performance

### 3.1 Request Optimization

```go
// Enable HTTP/2 and connection reuse
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    EnableCompression:   true,
}

client := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

### 3.2 Response Compression

```go
// middleware/compression.go
import "github.com/gin-gonic/gin"

func CompressionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip compression for small responses
        c.Next()
        
        if c.Writer.Size() > 1024 { // 1KB threshold
            c.Header("Content-Encoding", "gzip")
        }
    }
}

// Use Brotli for better compression
import "github.com/andybalholm/brotli"
```

### 3.3 Pagination

```go
// handlers/pagination.go
type PaginationParams struct {
    Page  int `form:"page,default=1"`
    Limit int `form:"limit,default=20"`
}

func (p *PaginationParams) Validate() error {
    if p.Page < 1 {
        p.Page = 1
    }
    if p.Limit < 1 {
        p.Limit = 20
    }
    if p.Limit > 100 {
        p.Limit = 100 // Max limit
    }
    return nil
}

func (p *PaginationParams) Offset() int {
    return (p.Page - 1) * p.Limit
}

type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    TotalPages int         `json:"total_pages"`
}
```

### 3.4 Rate Limiting

```go
// middleware/rate_limit.go
import (
    "time"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
    }
}

func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    if limiter, ok := rl.limiters[key]; ok {
        return limiter
    }
    
    // 100 requests per minute per key
    limiter := rate.NewLimiter(rate.Every(time.Minute/100), 10)
    rl.limiters[key] = limiter
    return limiter
}

func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.ClientIP() // Or user ID if authenticated
        
        lim := limiter.GetLimiter(key)
        if !lim.Allow() {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## 4. Caching Strategies

### 4.1 Redis Caching

```go
// cache/redis.go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

type Cache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewCache(addr string, ttl time.Duration) *Cache {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        PoolSize: 10,
    })
    return &Cache{client: client, ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return ErrCacheMiss
    }
    if err != nil {
        return err
    }
    return json.Unmarshal(data, dest)
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, pattern string) error {
    iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        c.client.Del(ctx, iter.Val())
    }
    return iter.Err()
}
```

### 4.2 Cache Strategies

```go
// Cache-aside pattern
func (s *MediaService) GetMediaItem(ctx context.Context, id string) (*MediaItem, error) {
    // Try cache first
    var item MediaItem
    if err := s.cache.Get(ctx, "media:"+id, &item); err == nil {
        return &item, nil
    }
    
    // Cache miss - fetch from database
    item, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.Set(ctx, "media:"+id, item)
    return &item, nil
}

// Write-through pattern
func (s *MediaService) UpdateMediaItem(ctx context.Context, item *MediaItem) error {
    // Update database first
    if err := s.repo.Update(ctx, item); err != nil {
        return err
    }
    
    // Update cache
    s.cache.Set(ctx, "media:"+item.ID, item)
    return nil
}
```

### 4.3 Cache Invalidation

```go
// Invalidate related caches
func (s *MediaService) InvalidateMediaCaches(mediaID string) {
    ctx := context.Background()
    
    // Invalidate specific item
    s.cache.Delete(ctx, "media:"+mediaID)
    
    // Invalidate list caches
    s.cache.Delete(ctx, "media:list:*")
    
    // Invalidate search caches
    s.cache.Delete(ctx, "media:search:*")
}
```

---

## 5. Memory Management

### 5.1 Object Pooling

```go
// pool/buffer_pool.go
import (
    "bytes"
    "sync"
)

var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func GetBuffer() *bytes.Buffer {
    buf := bufferPool.Get().(*bytes.Buffer)
    buf.Reset()
    return buf
}

func PutBuffer(buf *bytes.Buffer) {
    bufferPool.Put(buf)
}

// Usage
func ProcessData(data []byte) []byte {
    buf := GetBuffer()
    defer PutBuffer(buf)
    
    // Process data into buffer
    buf.Write(data)
    // ... processing ...
    
    return buf.Bytes()
}
```

### 5.2 Streaming Responses

```go
// handlers/stream.go
func (h *Handler) StreamLargeFile(c *gin.Context) {
    filePath := c.Param("path")
    
    file, err := os.Open(filePath)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer file.Close()
    
    // Stream file in chunks
    c.Stream(func(w io.Writer) bool {
        buf := make([]byte, 64*1024) // 64KB chunks
        n, err := file.Read(buf)
        if n > 0 {
            w.Write(buf[:n])
        }
        return err == nil
    })
}
```

### 5.3 Memory Profiling

```go
// profiling/memory.go
import (
    "runtime"
    "runtime/pprof"
)

func WriteHeapProfile(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    runtime.GC() // Force garbage collection
    return pprof.WriteHeapProfile(f)
}

func PrintMemoryStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
    log.Printf("TotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
    log.Printf("Sys = %v MiB", m.Sys/1024/1024)
    log.Printf("NumGC = %v", m.NumGC)
}
```

---

## 6. File Scanning Optimization

### 6.1 Parallel Scanning

```go
// scanner/parallel.go
import (
    "context"
    "path/filepath"
    "sync"
)

type ParallelScanner struct {
    workers   int
    semaphore chan struct{}
}

func NewParallelScanner(workers int) *ParallelScanner {
    return &ParallelScanner{
        workers:   workers,
        semaphore: make(chan struct{}, workers),
    }
}

func (s *ParallelScanner) Scan(ctx context.Context, roots []string) ([]FileInfo, error) {
    var wg sync.WaitGroup
    filesChan := make(chan FileInfo, 1000)
    errChan := make(chan error, 1)
    
    // Start result collector
    var files []FileInfo
    go func() {
        for f := range filesChan {
            files = append(files, f)
        }
    }()
    
    // Scan each root in parallel
    for _, root := range roots {
        wg.Add(1)
        go func(r string) {
            defer wg.Done()
            s.scanDirectory(ctx, r, filesChan, errChan)
        }(root)
    }
    
    wg.Wait()
    close(filesChan)
    
    return files, nil
}

func (s *ParallelScanner) scanDirectory(ctx context.Context, root string, out chan<- FileInfo, errChan chan<- error) {
    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // Skip errors
        }
        
        if info.IsDir() {
            return nil
        }
        
        // Acquire semaphore
        s.semaphore <- struct{}{}
        defer func() { <-s.semaphore }()
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case out <- FileInfo{
            Path: path,
            Size: info.Size(),
            ModTime: info.ModTime(),
        }:
        }
        
        return nil
    })
}
```

### 6.2 Incremental Scanning

```go
// scanner/incremental.go
func (s *Scanner) ScanIncremental(ctx context.Context, root string, since time.Time) ([]FileInfo, error) {
    var newFiles []FileInfo
    
    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        
        // Skip unchanged files
        if info.ModTime().Before(since) {
            if info.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }
        
        if !info.IsDir() {
            newFiles = append(newFiles, FileInfo{
                Path:    path,
                Size:    info.Size(),
                ModTime: info.ModTime(),
            })
        }
        
        return nil
    })
    
    return newFiles, err
}
```

---

## 7. Monitoring & Profiling

### 7.1 Application Metrics

```go
// metrics/performance.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )
    
    databaseQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_query_duration_seconds",
            Help: "Database query duration",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"query_type"},
    )
    
    cacheHitRate = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total cache hits",
        },
        []string{"cache_name"},
    )
)

// Middleware to track request duration
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        duration := time.Since(start).Seconds()
        requestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}
```

### 7.2 Profiling Endpoints

```go
// handlers/debug.go
import (
    "net/http/pprof"
    "github.com/gin-gonic/gin"
)

func RegisterDebugHandlers(r *gin.Engine) {
    debug := r.Group("/debug/pprof")
    {
        debug.GET("/", gin.WrapF(pprof.Index))
        debug.GET("/cmdline", gin.WrapF(pprof.Cmdline))
        debug.GET("/profile", gin.WrapF(pprof.Profile))
        debug.GET("/symbol", gin.WrapF(pprof.Symbol))
        debug.GET("/trace", gin.WrapF(pprof.Trace))
        debug.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
        debug.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
        debug.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
    }
}
```

### 7.3 Performance Testing

```bash
# Load testing with k6
cat > load_test.js << 'EOF'
import http from 'k6/http';
import { check } from 'k6';

export const options = {
    stages: [
        { duration: '2m', target: 100 },
        { duration: '5m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '5m', target: 200 },
        { duration: '2m', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<200'],
        http_req_failed: ['rate<0.01'],
    },
};

export default function () {
    const res = http.get('http://localhost:8080/api/v1/media');
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 200ms': (r) => r.timings.duration < 200,
    });
}
EOF

k6 run load_test.js
```

---

## 8. Troubleshooting

### 8.1 Slow Query Analysis

```sql
-- Find slow queries in PostgreSQL
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check for missing indexes
SELECT 
    schemaname,
    tablename,
    attname AS column,
    n_tup_read,
    n_tup_fetch
FROM pg_stats
WHERE schemaname = 'public'
ORDER BY n_tup_read DESC;
```

### 8.2 Memory Leak Detection

```bash
# Check for goroutine leaks
curl http://localhost:8080/debug/pprof/goroutine?debug=1

# Check heap allocation
go tool pprof -top http://localhost:8080/debug/pprof/heap

# Generate memory profile over time
go tool pprof --pdf http://localhost:8080/debug/pprof/heap > heap.pdf
```

### 8.3 Common Issues

| Issue | Symptoms | Solution |
|-------|----------|----------|
| High CPU | Slow responses, timeouts | Profile with pprof, optimize hot paths |
| Memory leak | OOM crashes | Check for unclosed resources, goroutine leaks |
| Slow database | Query timeouts | Add indexes, optimize queries, scale DB |
| Cache stampede | Spike in DB load | Implement cache warming, stagger TTLs |
| Network latency | Slow external calls | Enable compression, use CDN, connection pooling |

### 8.4 Performance Checklist

- [ ] Database indexes created for common queries
- [ ] Connection pool configured appropriately
- [ ] Response compression enabled
- [ ] Caching strategy implemented
- [ ] Rate limiting configured
- [ ] Pagination on list endpoints
- [ ] Profiling endpoints enabled (dev/staging)
- [ ] Monitoring dashboards created
- [ ] Load tests run regularly
- [ ] Performance benchmarks documented

---

**Document Control:**
- Version: 1.0
- Approved by: [Engineering Lead]
- Date approved: April 6, 2026
- Next review: July 6, 2026

