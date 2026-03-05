# Module 11: Security and Monitoring - Script

**Duration**: 30 minutes
**Module**: 11 - Security and Monitoring

---

## Scene 1: Security Best Practices (0:00 - 15:00)

**[Visual: Security audit summary: govulncheck 0 vulnerabilities, npm audit 0 critical/production vulnerabilities]**

**Narrator**: Welcome to Module 11. Catalogizer takes security seriously. Zero known vulnerabilities in Go dependencies. Zero critical vulnerabilities in production npm packages. This is not accidental -- it is the result of layered security controls baked into the codebase.

**[Visual: Show security scanning tools]**

**Narrator**: Three security tools run as part of the build pipeline:

1. **govulncheck** -- Go vulnerability database scanner. Checks all dependencies against the official Go vulnerability database.
2. **gosec** -- Static analysis for Go security issues. Detects hardcoded credentials, SQL injection, insecure random, and more.
3. **npm audit** -- Node.js dependency vulnerability scanner for the frontend.

```bash
# Run security scans
cd catalog-api
govulncheck ./...
gosec ./...

cd catalog-web
npm audit --production
```

**[Visual: Open `catalog-api/middleware/input_validation.go`]**

**Narrator**: Input validation is the first line of defense. The `InputValidationConfig` enables SQL injection detection, XSS detection, and path traversal detection by default.

```go
// catalog-api/middleware/input_validation.go
type InputValidationConfig struct {
    MaxRequestBodySize           int64
    EnableSQLInjectionDetection  bool
    EnableXSSDetection           bool
    EnablePathTraversalDetection bool
    CustomRules                  map[string]string
}

func DefaultInputValidationConfig() InputValidationConfig {
    return InputValidationConfig{
        MaxRequestBodySize:           10 * 1024 * 1024, // 10MB
        EnableSQLInjectionDetection:  true,
        EnableXSSDetection:           true,
        EnablePathTraversalDetection: true,
    }
}
```

**[Visual: Show SQL injection patterns]**

**Narrator**: SQL injection detection uses pattern matching against known attack vectors. The middleware scans request bodies and query parameters for SQL keywords in suspicious contexts: `SELECT`, `INSERT`, `DROP`, `UNION`, comment sequences, and time-based blind injection patterns like `WAITFOR DELAY` and `SLEEP`.

```go
// catalog-api/middleware/input_validation.go
var sqlInjectionPatterns = []string{
    `(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|UNION|SCRIPT|OR)\b)`,
    `(?i)(['"]\s*;\s*(SELECT|INSERT|UPDATE|DELETE|DROP))`,
    `(?i)(\bWAITFOR\s+DELAY\b)`,
    `(?i)(\bBENCHMARK\b)`,
    `(?i)(\bSLEEP\b)`,
}
```

**[Visual: Show XSS patterns]**

**Narrator**: XSS detection identifies script tags, event handlers, and JavaScript URI schemes in user input. Any input matching these patterns is rejected before it reaches the handler layer.

```go
var xssPatterns = []string{
    `(?i)(<script[^>]*>.*?</script>)`,
    // ... event handlers, javascript: URIs, etc.
}
```

**[Visual: Show the database layer's parameterized queries]**

**Narrator**: Beyond pattern matching, the database layer uses parameterized queries exclusively. The dialect abstraction rewrites `?` placeholders -- never string concatenation. This structural defense makes SQL injection impossible even if the middleware is bypassed.

**[Visual: Show CSRF and rate limiting]**

**Narrator**: Rate limiting prevents brute-force attacks. The Redis rate limiter in `middleware/redis_rate_limiter.go` uses sliding window counters per client IP. The advanced rate limiter in `middleware/advanced_rate_limiter.go` adds endpoint-specific limits -- login endpoints are rate-limited more aggressively than read endpoints.

**[Visual: Show bcrypt password hashing]**

**Narrator**: Passwords use bcrypt with per-user salts, as we saw in Module 3. Bcrypt's configurable work factor makes offline cracking computationally expensive. The cost factor is set high enough to be secure but low enough to not slow down login.

**[Visual: Show JWT secret configuration]**

**Narrator**: The JWT secret comes from environment variables, which override `config.json`. The secret is never hardcoded, never committed to git, and never logged. In production, it should be at least 256 bits of cryptographic randomness.

---

## Scene 2: Observability (15:00 - 30:00)

**[Visual: Open `catalog-api/internal/metrics/prometheus.go`]**

**Narrator**: Observability tells you what your application is doing in production. Catalogizer exposes Prometheus metrics, structured logging with Zap, and health check endpoints.

**Narrator**: Prometheus metrics are registered using `promauto` for automatic collection. The codebase tracks database queries (by operation and table), connection pool state, media files scanned and analyzed, analysis duration, media counts by type, and SMB connection health.

```go
// catalog-api/internal/metrics/prometheus.go
var (
    DBQueryTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "catalogizer_db_queries_total",
            Help: "Total number of database queries",
        },
        []string{"operation", "table"},
    )

    MediaFilesScanned = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "catalogizer_media_files_scanned_total",
            Help: "Total number of media files scanned",
        },
    )

    MediaAnalysisDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "catalogizer_media_analysis_duration_seconds",
            Help:    "Media analysis duration in seconds",
            Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
        },
    )
)
```

**[Visual: Show metrics middleware]**

**Narrator**: The metrics middleware in `internal/metrics/middleware.go` instruments every HTTP request. It tracks request count, duration, and status code by endpoint and method. These metrics feed dashboards and alerting rules.

**[Visual: Show the `/metrics` endpoint]**

**Narrator**: Metrics are exposed at the `/metrics` endpoint in Prometheus exposition format. Prometheus scrapes this endpoint at configurable intervals, stores the time series, and Grafana visualizes them.

**[Visual: Show structured logging with Zap]**

**Narrator**: All logging uses `go.uber.org/zap` for structured, leveled output. Each log entry is a JSON object with fields for timestamp, level, message, and context. Structured logs are machine-parseable, enabling log aggregation and search.

```go
// Usage throughout the codebase
logger.Info("Starting post-scan aggregation",
    zap.Int64("storage_root_id", storageRootID))

logger.Warn("Failed to process directory",
    zap.String("path", dir.path),
    zap.Error(err))
```

**[Visual: Open `catalog-api/internal/metrics/health.go`]**

**Narrator**: Health checks verify that critical components are working. The `HealthChecker` registers component-specific checks and aggregates them into an overall health status.

```go
// catalog-api/internal/metrics/health.go
type HealthCheckResponse struct {
    Status     HealthStatus               `json:"status"`
    Timestamp  time.Time                  `json:"timestamp"`
    Version    string                     `json:"version"`
    Uptime     string                     `json:"uptime"`
    Components map[string]ComponentHealth `json:"components"`
}

func NewHealthChecker(db *database.DB, version string) *HealthChecker {
    hc := &HealthChecker{
        db:        db,
        startTime: time.Now(),
        version:   version,
        checks:    make(map[string]func(context.Context) ComponentHealth),
    }
    hc.RegisterCheck("database", hc.checkDatabase)
    return hc
}
```

**[Visual: Show health status enum]**

**Narrator**: Each component reports one of three statuses: Healthy, Degraded, or Unhealthy. The overall status is the worst of all component statuses. A degraded database connection degrades the entire system.

```go
const (
    HealthStatusHealthy   HealthStatus = "healthy"
    HealthStatusDegraded  HealthStatus = "degraded"
    HealthStatusUnhealthy HealthStatus = "unhealthy"
)
```

**[Visual: Show the error reporting service]**

**Narrator**: The error reporting service in `services/error_reporting_service.go` captures, stores, and aggregates runtime errors. It tracks error frequency, categorizes errors by type, and provides an API for querying recent errors. This works alongside Prometheus for a complete observability picture.

**[Visual: Show SMB health metrics integration]**

**Narrator**: SMB connection states map directly to Prometheus gauge values. A connected state reports healthy (1.0), reconnecting reports degraded (0.5), and offline reports 0. These metrics feed into alerting rules that notify administrators when a NAS becomes unreachable.

**[Visual: Course title card]**

**Narrator**: Security and monitoring are not features you add later -- they are integral to the architecture. Input validation, parameterized queries, rate limiting, and dependency scanning protect the system. Prometheus metrics, structured logging, and health checks make it observable. In Module 12, we deploy everything to production.

---

## Key Code Examples

### Security Scanning Commands
```bash
# Go vulnerability check
cd catalog-api && govulncheck ./...

# Go static security analysis
cd catalog-api && gosec ./...

# Frontend dependency audit
cd catalog-web && npm audit --production

# Full security suite
./scripts/run-all-tests.sh  # includes security scans
```

### Health Check Endpoint
```bash
curl http://localhost:8080/api/v1/health
```
```json
{
  "status": "healthy",
  "timestamp": "2026-03-05T12:00:00Z",
  "version": "1.0.0",
  "uptime": "48h12m33s",
  "components": {
    "database": { "status": "healthy", "latency": "1ms" },
    "redis": { "status": "healthy", "latency": "0.5ms" },
    "smb_nas": { "status": "degraded", "message": "reconnecting" }
  }
}
```

### Prometheus Metrics Endpoint
```bash
curl http://localhost:8080/metrics
# Returns:
# catalogizer_db_queries_total{operation="select",table="files"} 45231
# catalogizer_media_files_scanned_total 85000
# catalogizer_media_analysis_duration_seconds_bucket{le="1"} 72000
# ...
```

---

## Quiz Questions

1. What three security scanning tools does Catalogizer use, and what does each detect?
   **Answer**: (1) **govulncheck** scans Go dependencies against the official Go vulnerability database for known CVEs. (2) **gosec** performs static analysis on Go source code, detecting hardcoded credentials, SQL injection patterns, insecure random number generation, and other coding issues. (3) **npm audit** scans JavaScript dependencies for known vulnerabilities with severity ratings.

2. How does the input validation middleware protect against SQL injection at two levels?
   **Answer**: First level: pattern-matching middleware scans all request bodies and query parameters for SQL injection signatures (SQL keywords in suspicious contexts, comment sequences, time-based blind injection). Second level: the database layer uses parameterized queries exclusively -- `?` placeholders rewritten by the dialect layer. Even if the middleware is bypassed, SQL injection is structurally impossible through parameterized queries.

3. What information does the health check endpoint expose?
   **Answer**: The health check returns an overall status (healthy/degraded/unhealthy), server timestamp, application version, uptime duration, and per-component health. Each component (database, Redis, SMB connections) reports its own status, optional message, and latency. The overall status is the worst of all components.

4. Why does Catalogizer use structured logging (Zap) instead of `fmt.Printf`?
   **Answer**: Structured logs produce JSON with consistent fields (timestamp, level, message, context), making them machine-parseable for log aggregation tools (ELK, Loki, CloudWatch). They support log levels (debug, info, warn, error) for filtering. Zap is high-performance with zero-allocation encoding. Context fields (like `storage_root_id`, `error`) enable correlation across log entries for debugging distributed issues.
