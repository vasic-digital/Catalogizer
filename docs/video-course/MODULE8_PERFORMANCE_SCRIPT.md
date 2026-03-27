# Module 8: HTTP/3 and Performance - Script

**Duration**: 45 minutes
**Module**: 8 - HTTP/3 and Performance

---

## Scene 1: HTTP/3 with QUIC (0:00 - 20:00)

**[Visual: Protocol comparison diagram: HTTP/1.1 vs HTTP/2 vs HTTP/3 (QUIC)]**

**Narrator**: Welcome to Module 8. Catalogizer mandates HTTP/3 with QUIC for all network communication. HTTP/3 eliminates TCP head-of-line blocking, reduces connection establishment latency, and handles network changes (like WiFi to cellular) gracefully. Let us see how it is implemented.

**[Visual: Open `catalog-api/main.go` showing HTTP/3 imports]**

**Narrator**: The backend uses `quic-go/http3` for HTTP/3 support. On startup, it generates a self-signed TLS certificate (for development) and starts both HTTP/2 and HTTP/3 servers. The HTTP/3 server uses QUIC as its transport layer.

```go
// catalog-api/main.go
import (
    "github.com/quic-go/quic-go/http3"
    // ...
)
```

**[Visual: Show `generateSelfSignedCert` function]**

**Narrator**: For development, TLS certificates are generated at startup using Go's crypto libraries. This function creates a 2048-bit RSA key, generates a self-signed X.509 certificate, and returns a `tls.Certificate` ready for the server.

```go
// catalog-api/main.go
func generateSelfSignedCert() (tls.Certificate, error) {
    priv, err := rsa.GenerateKey(rand.Reader, 2048)
    // ... create X.509 certificate template
    // ... self-sign the certificate
    return tls.Certificate{...}, nil
}
```

**[Visual: Show dual server startup]**

**Narrator**: The server starts two listeners: a standard `net/http` server for HTTP/2 and an `http3.Server` for HTTP/3. Both serve the same Gin router. Clients that support HTTP/3 use QUIC; others fall back to HTTP/2. The `Alt-Svc` header advertises HTTP/3 availability.

**[Visual: Show ALPN negotiation]**

**Narrator**: Application-Layer Protocol Negotiation (ALPN) happens during the TLS handshake. The server advertises support for "h3" (HTTP/3) and "h2" (HTTP/2). The client selects the best protocol it supports. This negotiation is transparent to application code.

**[Visual: Show the TLS configuration]**

**Narrator**: The TLS configuration sets minimum version to TLS 1.3 (required by HTTP/3), configures the ALPN protocols, and applies the generated certificate. In production, you would replace the self-signed cert with a proper certificate from your CA.

**[Visual: Show version injection via ldflags]**

**Narrator**: Performance-related metadata is also injected at build time. Version, build number, and build date are set via Go's `-ldflags` mechanism, allowing the server to report its exact build in health check responses and metrics.

```go
// Build command
go build -ldflags "-X main.Version=1.0.0 -X main.BuildNumber=42 -X main.BuildDate=2026-03-05" -o catalog-api
```

---

## Scene 2: Brotli Compression (20:00 - 30:00)

**[Visual: Compression comparison chart: Brotli vs Gzip vs No compression]**

**Narrator**: Brotli compression provides 20-26% better compression ratios than gzip for web content. Catalogizer uses Brotli as the primary compression algorithm, with gzip as fallback for older clients.

**[Visual: Show Brotli middleware in the server]**

**Narrator**: The backend uses `andybalholm/brotli` for compression middleware. Every response goes through content negotiation: if the client sends `Accept-Encoding: br`, the response is Brotli-compressed. If only `gzip` is accepted, gzip is used. If neither, the response is sent uncompressed.

**[Visual: Show content negotiation flow]**

**Narrator**: Content negotiation reads the `Accept-Encoding` header, checks for `br` (Brotli), `gzip`, and `deflate`, and selects the best match. The `Content-Encoding` response header tells the client which encoding was used.

**[Visual: Show compression levels]**

**Narrator**: Brotli supports 11 compression levels (0-11). Higher levels produce smaller output but take longer. For dynamic API responses, Catalogizer uses level 4 -- a good balance between speed and size. For pre-compressed static assets, level 11 is used during build time.

**[Visual: Show static asset optimization]**

**Narrator**: The frontend build produces pre-compressed `.br` files for all static assets. When the web server detects a pre-compressed file, it serves it directly without real-time compression. This eliminates compression CPU overhead for the most frequently requested resources.

---

## Scene 3: Caching Strategies (30:00 - 45:00)

**[Visual: Caching architecture diagram: Redis -> Database Cache -> In-Memory -> Disk]**

**Narrator**: Catalogizer implements a multi-tier caching strategy. The first tier is Redis (optional), the second is database-backed caching, and the third is the HTTP cache headers strategy.

**[Visual: Show Redis integration in `main.go`]**

**Narrator**: Redis integration uses `go-redis/v9`. Redis is optional -- if unavailable, the system falls back to database-backed caching. The connection is established at startup with configurable address, password, and database number.

```go
// catalog-api/main.go
import (
    "github.com/redis/go-redis/v9"
)

// Redis client setup
rdb := redis.NewClient(&redis.Options{
    Addr:     redisAddr,
    Password: redisPassword,
    DB:       redisDB,
})
```

**[Visual: Open `catalog-api/internal/services/cache_service.go`]**

**Narrator**: The `CacheService` manages multiple cache types. The general cache stores arbitrary key-value pairs with TTL. The media metadata cache stores provider responses per media item. The API cache stores external API responses to reduce rate limiting pressure. The thumbnail cache stores generated thumbnails.

```go
// catalog-api/internal/services/cache_service.go
type CacheEntry struct {
    ID        int64     `json:"id"`
    CacheKey  string    `json:"cache_key"`
    Value     string    `json:"value"`
    ExpiresAt time.Time `json:"expires_at"`
}

type MediaMetadataCache struct {
    ID           int64     `json:"id"`
    MediaItemID  int64     `json:"media_item_id"`
    MetadataType string    `json:"metadata_type"`
    Provider     string    `json:"provider"`
    Data         string    `json:"data"`
    Quality      float64   `json:"quality"`
    ExpiresAt    time.Time `json:"expires_at"`
}

type APICache struct {
    ID          int64     `json:"id"`
    Provider    string    `json:"provider"`
    Endpoint    string    `json:"endpoint"`
    RequestHash string    `json:"request_hash"`
    Response    string    `json:"response"`
    StatusCode  int       `json:"status_code"`
    ExpiresAt   time.Time `json:"expires_at"`
}
```

**[Visual: Show cache invalidation strategy]**

**Narrator**: Cache invalidation follows a clear policy. When a scan completes, file-related caches are invalidated. When an entity is updated, its metadata and thumbnail caches are cleared. The API cache uses TTL-based expiration -- typically 24 hours for metadata providers and 1 hour for search results.

**[Visual: Open `catalog-api/middleware/cache_headers.go`]**

**Narrator**: HTTP cache headers are managed by middleware. The `cache_headers.go` middleware sets `Cache-Control`, `ETag`, and `Last-Modified` headers based on the response type. Static assets get long max-age values. API responses use short TTLs with `stale-while-revalidate`.

```go
// catalog-api/middleware/cache_headers.go
// Applies Cache-Control headers based on endpoint patterns
// Static assets: max-age=31536000, immutable
// API data: max-age=60, stale-while-revalidate=300
// User-specific: no-cache, no-store
```

**[Visual: Show Redis rate limiter]**

**Narrator**: Redis also powers the rate limiter. The `redis_rate_limiter.go` middleware uses Redis sorted sets to implement sliding window rate limiting. Each client IP gets a window of allowed requests. This is more accurate than fixed-window approaches and handles burst traffic gracefully.

**[Visual: Show Prometheus metrics for cache performance]**

**Narrator**: Cache performance is observable through Prometheus metrics. Hit rates, miss rates, eviction counts, and cache sizes are all tracked. Degrading hit rates trigger alerts before they impact user experience.

**[Visual: Course title card]**

**Narrator**: Performance optimization is not an afterthought in Catalogizer. HTTP/3 reduces latency, Brotli shrinks payloads, multi-tier caching eliminates redundant work, and Redis provides high-performance distributed caching. In Module 9, we extend the platform to desktop and mobile applications.

---

## Key Code Examples

### HTTP/3 Server Configuration
```go
// main.go - Dual server startup
// HTTP/2 server
httpServer := &http.Server{
    Addr:      addr,
    Handler:   router,
    TLSConfig: tlsConfig,
}
go httpServer.ListenAndServeTLS("", "")

// HTTP/3 server
http3Server := &http3.Server{
    Addr:      addr,
    Handler:   router,
    TLSConfig: tlsConfig,
}
go http3Server.ListenAndServe()
```

### Resource Limits (Mandatory)
```bash
# Go tests must respect resource limits
GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# Container resource limits
# PostgreSQL: --cpus=1 --memory=2g
# API: --cpus=2 --memory=4g
# Web: --cpus=1 --memory=2g
# Builder: --cpus=3 --memory=8g
```

### Redis Configuration
```bash
# config/redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
save ""
appendonly no
```

---

## Quiz Questions

1. Why does Catalogizer mandate HTTP/3 (QUIC) instead of HTTP/2?
   **Answer**: HTTP/3 uses QUIC (UDP-based) instead of TCP. This eliminates TCP head-of-line blocking (where a lost packet blocks all streams), reduces connection establishment from 2-3 RTTs (TCP+TLS) to 1 RTT (or 0-RTT for resumed connections), and supports connection migration (switching networks without dropping the connection). These benefits are significant for media browsing applications with many concurrent requests.

2. What are the different cache tiers in Catalogizer, and when is each used?
   **Answer**: (1) Redis: high-performance, distributed cache for rate limiting, session data, and frequently accessed values. Optional. (2) Database cache: persistent cache with TTL for metadata, API responses, and thumbnails. Always available. (3) HTTP cache headers: browser-side caching with Cache-Control, ETag, and stale-while-revalidate. Reduces server load. Each tier has different performance characteristics and durability guarantees.

3. How does Brotli compression differ between static assets and dynamic API responses?
   **Answer**: Static assets are pre-compressed at build time with Brotli level 11 (maximum compression). The server serves the `.br` file directly with zero runtime CPU cost. Dynamic API responses are compressed at level 4 (balanced speed/size) by the middleware in real time. Content negotiation via Accept-Encoding determines whether Brotli, gzip, or no compression is used.

4. How does the Redis-based sliding window rate limiter work?
   **Answer**: Each client IP gets a Redis sorted set where member scores are timestamps. On each request, expired entries (outside the window) are removed, the new request is added, and the set size is checked against the limit. If the count exceeds the limit, the request is rejected with 429 Too Many Requests. This provides accurate, distributed rate limiting across multiple API instances.

---

## Addendum: BoundedSemaphore, Concurrent SearchAll, and Non-Blocking Health

**[Visual: Diagram comparing unbounded goroutine fan-out vs semaphore-bounded fan-out]**

**Narrator**: As Catalogizer added more metadata providers -- TMDB, OMDB, OpenLibrary, MusicBrainz, IGDB, and others -- the sequential search approach became a bottleneck. Querying six providers one after another took 4-5 seconds. But spawning unbounded goroutines risked overwhelming external APIs and exhausting rate limits. The `BoundedSemaphore` pattern solves both problems.

### BoundedSemaphore

**[Visual: Open `catalog-api/internal/concurrency/semaphore.go`]**

**Narrator**: The `BoundedSemaphore` is a context-aware concurrency limiter. Unlike a simple buffered channel semaphore, it supports cancellation -- if the context expires while waiting for a slot, the goroutine exits cleanly instead of leaking.

```go
// internal/concurrency/semaphore.go
type BoundedSemaphore struct {
    sem chan struct{}
}

func NewBoundedSemaphore(limit int) *BoundedSemaphore {
    return &BoundedSemaphore{sem: make(chan struct{}, limit)}
}

func (s *BoundedSemaphore) Acquire(ctx context.Context) error {
    select {
    case s.sem <- struct{}{}:
        return nil
    case <-ctx.Done():
        return ctx.Err() // No goroutine leak
    }
}

func (s *BoundedSemaphore) Release() {
    <-s.sem
}
```

**Narrator**: The `Acquire` method uses a `select` statement that races between getting a semaphore slot and context cancellation. This is critical during shutdown -- when the application context is cancelled, all waiting goroutines exit immediately rather than blocking forever.

### Concurrent SearchAll

**[Visual: Timeline diagram: sequential (5 providers x 1s = 5s) vs parallel-bounded (5 providers concurrent = 1s max)]**

**Narrator**: The `SearchAll` function in the entity search service uses `BoundedSemaphore` to query metadata providers concurrently with a configurable maximum of 5 simultaneous requests.

```go
func (s *EntityService) SearchAll(ctx context.Context, query string) ([]Result, error) {
    sem := concurrency.NewBoundedSemaphore(5)
    var wg sync.WaitGroup
    var mu sync.Mutex
    var results []Result

    for _, provider := range s.providers {
        wg.Add(1)
        go func(p MetadataProvider) {
            defer wg.Done()
            if err := sem.Acquire(ctx); err != nil {
                return // Context cancelled during shutdown
            }
            defer sem.Release()

            res, err := p.Search(ctx, query)
            if err != nil {
                return // Individual provider failure is non-fatal
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

**Narrator**: Three critical design decisions here. First, each provider failure is isolated -- a timeout from TMDB does not block MusicBrainz results. Second, the mutex protects only the results slice, not the provider call, minimizing lock contention. Third, the semaphore limit of 5 prevents rate limit exhaustion across providers that share API rate budgets.

**[Visual: Performance metrics before and after: p95 latency from 4.2s to 1.2s]**

**Narrator**: The impact is dramatic. Multi-provider search p95 latency dropped from 4.2 seconds to 1.2 seconds -- a 71% improvement. And because the semaphore prevents unlimited concurrency, external API rate limit violations dropped to zero.

### Non-Blocking Health Check

**[Visual: Load balancer health check diagram showing cascading failures from slow health endpoint]**

**Narrator**: The `/health/deep` endpoint demonstrates another application of bounded concurrency. Load balancers and orchestrators poll health endpoints every 5-10 seconds. If the health check blocks because the database or Redis is slow, the load balancer marks the instance as unhealthy, stops routing traffic, and causes a cascading failure.

```go
func (h *HealthHandler) DeepHealthCheck(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Millisecond)
    defer cancel()

    checks := map[string]func(context.Context) error{
        "database": h.pingDB,
        "redis":    h.pingRedis,
        "storage":  h.checkStorageAccess,
    }

    results := make(map[string]string)
    var mu sync.Mutex
    var wg sync.WaitGroup

    for name, check := range checks {
        wg.Add(1)
        go func(n string, fn func(context.Context) error) {
            defer wg.Done()
            if err := fn(ctx); err != nil {
                mu.Lock()
                results[n] = "degraded"
                mu.Unlock()
            } else {
                mu.Lock()
                results[n] = "healthy"
                mu.Unlock()
            }
        }(name, check)
    }
    wg.Wait()

    overall := "healthy"
    for _, status := range results {
        if status != "healthy" {
            overall = "degraded"
        }
    }
    c.JSON(http.StatusOK, gin.H{"status": overall, "checks": results})
}
```

**Narrator**: All subsystem checks run concurrently with a shared 100ms deadline. If Redis takes 2 seconds to respond, it is marked "degraded" after 100ms and the endpoint returns immediately. The overall status is "degraded" rather than "unhealthy" -- the load balancer continues routing traffic, and operators get a clear signal about which subsystem needs attention.

**[Visual: Health endpoint latency chart: p99 from 2.1s to 105ms]**

**Narrator**: Health endpoint p99 latency dropped from 2.1 seconds to 105 milliseconds. More importantly, a slow Redis or database no longer triggers false-positive instance removal from the load balancer pool.

**Key takeaways:**
- `BoundedSemaphore` with context-aware `Acquire()` prevents goroutine leaks during shutdown.
- Concurrent provider search reduced latency by 71% while preventing rate limit exhaustion.
- Non-blocking health checks prevent cascading failures from slow subsystems.
- The 100ms health check timeout is deliberately aggressive -- it is better to report "degraded" than to block.
