# Security Headers

catalog-api sets security headers on every HTTP response via the `SecurityHeaders()` middleware in `catalog-api/middleware/security_headers.go`.

## Headers Applied

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Content-Type-Options` | `nosniff` | Prevents MIME-type sniffing -- browser must respect declared Content-Type |
| `X-Frame-Options` | `DENY` | Blocks all framing -- prevents clickjacking attacks |
| `X-XSS-Protection` | `1; mode=block` | Enables browser XSS filter and blocks the page on detection |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Sends full URL for same-origin, only origin for cross-origin |
| `Permissions-Policy` | `camera=(), microphone=(), geolocation=()` | Denies access to camera, microphone, and geolocation APIs |

## Conditional Headers

| Header | Value | Condition |
|--------|-------|-----------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Only set when `c.Request.TLS != nil` (HTTPS/QUIC connections) |

HSTS tells browsers to only connect via HTTPS for 1 year, including subdomains. It is not set on plain HTTP to avoid breaking development setups.

## CORS Headers

CORS is handled by a separate middleware (`CORS()` in both `middleware/request.go` and `internal/middleware/middleware.go`). See `docs/security/CORS_CONFIGURATION.md` for details.

## Cache Headers

Cache headers are managed by dedicated middleware:

| Middleware | Behavior |
|------------|----------|
| `CacheHeaders(maxAge)` | Adds `Cache-Control: public, max-age=N` and computes SHA-256 `ETag` for GET responses. Returns 304 when `If-None-Match` matches. |
| `StaticCacheHeaders()` | Sets `Cache-Control: public, max-age=31536000, immutable` for fingerprinted static assets. |

## Compression Headers

The compression middleware (`internal/middleware/compression.go`) adds:

| Header | Value | When |
|--------|-------|------|
| `Content-Encoding` | `br` or `gzip` | Response body exceeds `MinSize` (default 1024 bytes) |
| `Vary` | `Accept-Encoding` | Always when compression is applied |

Brotli is preferred over gzip. Writer pools (`sync.Pool`) are used to reduce allocations.

## Request Tracking

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Request-ID` | UUID v4 or forwarded value | Set by `RequestID()` middleware. Preserves incoming `X-Request-ID` if present, otherwise generates a new UUID. Stored in Gin context as `request_id`. |

## Middleware Stack Order

The middleware chain executes in this order on every request:

1. `RequestID()` -- assign/forward request ID
2. `SecurityHeaders()` -- set security response headers
3. `CORS()` -- handle cross-origin requests and preflight
4. `RequestTimeout(30s)` -- apply request deadline to context
5. `ConcurrencyLimiter(N)` -- semaphore-based admission control
6. `CompressionMiddleware()` -- Brotli/gzip response compression
7. `CacheHeaders()` -- ETag and Cache-Control for GET requests
8. Route handler

## Stress Test Validation

The stress test suite (`catalog-api/tests/stress/middleware_chain_stress_test.go`) validates that security headers are present on every response under concurrent load (30 workers, 50 requests each = 1,500 concurrent checks). The test asserts zero missing headers.

## Source

- `catalog-api/middleware/security_headers.go`
- `catalog-api/middleware/request.go` (RequestID, CORS, RateLimiter)
- `catalog-api/middleware/cache_headers.go` (CacheHeaders, StaticCacheHeaders)
- `catalog-api/internal/middleware/middleware.go` (CORS, RequestID, RateLimiter)
- `catalog-api/internal/middleware/compression.go` (CompressionMiddleware)
