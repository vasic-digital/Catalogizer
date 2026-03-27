# Catalogizer Project Completion: Full 8-Phase Implementation Plan

**Date:** 2026-03-27
**Status:** Approved
**Approach:** Bottom-Up Safety First (Approach A)
**Scope:** All 7 platforms (API, Web, Desktop, Android, AndroidTV, Installer, HelixQA) + all 35 submodules

---

## Table of Contents

1. [Unfinished Work Report](#unfinished-work-report)
2. [Phase 1: Infrastructure & Security Scanning](#phase-1-infrastructure--security-scanning)
3. [Phase 2: Safety Hardening](#phase-2-safety-hardening)
4. [Phase 3: Dead Code & Completion](#phase-3-dead-code--completion)
5. [Phase 4: Test Coverage Maximum](#phase-4-test-coverage-maximum)
6. [Phase 5: Optimization](#phase-5-optimization)
7. [Phase 6: Challenge Coverage](#phase-6-challenge-coverage)
8. [Phase 7: Documentation & Courses](#phase-7-documentation--courses)
9. [Phase 8: Website Content](#phase-8-website-content)
10. [Cross-Phase Constraints](#cross-phase-constraints)
11. [Success Criteria](#success-criteria)

---

## Unfinished Work Report

### catalog-api (Go/Gin Backend)

#### Unimplemented Media Providers (~20 methods)
**Location:** `catalog-api/internal/media/providers/providers.go`
- `IMDBProvider.Search()` / `GetDetails()` — logs "not yet implemented"
- `TVDBProvider.GetDetails()` — stubbed
- `MusicBrainzProvider.GetDetails()` — stubbed
- Spotify, LastFM, AniDB, MyAnimeList, and other providers — multiple Search/GetDetails methods return nil gracefully

#### Dead Middleware (8 functions, ~800 lines unwired)
**Not wired in `main.go`:**

| Function | File |
|----------|------|
| `RedisRateLimit()` | `middleware/redis_rate_limiter.go` |
| `SlidingWindowRedisRateLimit()` | `middleware/redis_rate_limiter.go` |
| `TokenBucketRedisRateLimit()` | `middleware/redis_rate_limiter.go` |
| `AdvancedRateLimit()` | `middleware/advanced_rate_limiter.go` |
| `UserBasedRateLimit()` | `middleware/advanced_rate_limiter.go` |
| `IPRateLimit()` | `middleware/advanced_rate_limiter.go` |
| `CacheHeaders()` | `middleware/cache_headers.go` |
| `StaticCacheHeaders()` | `middleware/cache_headers.go` |

#### Untested Packages (3)
- `cmd/boot` — main entry point, no tests
- `internal/monitoring` — no tests
- `tests/mocks` — mock files, no tests by design

#### Conditionally Skipped Tests (~80+)
All legitimate: infrastructure-dependent (SMB/FTP/NFS/WebDAV), stress tests in `-short` mode, challenge tests when endpoints unreachable.

### catalog-web (React/TypeScript Frontend)

#### Mock-Only APIs (CRITICAL)
- **`adminApi.ts`** — ENTIRELY MOCK, returns hardcoded data, no real API calls
- **`conversionApi.ts`** — ENTIRELY MOCK, returns hardcoded conversion jobs

#### API Path Inconsistency
- `collectionsApi.ts` uses `/api/collections` instead of `/api/v1/collections`

#### Missing Granular Error Boundaries
- Only root-level `ErrorBoundary` — no per-page isolation

### HelixQA (Go QA Framework)
- **Production-ready.** 24 packages, all complete, 300+ tests, 1.08:1 test/code ratio
- Zero TODOs, zero dead code, zero compilation errors
- 297 test bank cases across 8 YAML files
- All 6 CLI subcommands implemented

### Submodules (35 total)
- All have content and compile
- `Build/` submodule is minimal (docs + utilities skeleton)
- All Go submodules have tests (2-116 test files each)
- All TS/React submodules configured for local linking

### Documentation Gaps
- Video courses lack modules on: HelixQA autonomous QA, entity system, collection management, AI features, subtitle management
- VitePress website functional but minimal — needs expanded content
- No user manual for mobile apps (Android/AndroidTV)
- No user manual for desktop app (Tauri)
- No user manual for installer wizard

### Security Scanning
- Infrastructure ready (docker-compose.security.yml) but not executed
- SonarQube, Snyk, Trivy, Semgrep, Dependency-Check all configured
- No scan reports exist — need execution and resolution

### Performance & Optimization Gaps
- K6 load tests defined (4 scripts) but not run for optimization feedback
- Lazy loading partially implemented but not comprehensive
- Semaphore mechanisms exist in some areas but not universally applied
- No monitoring/metrics collection tests

---

## Phase 1: Infrastructure & Security Scanning

### 1.1 Fix Scanning Infrastructure

**Step 1.1.1:** Verify `docker-compose.security.yml` starts all services with Podman
```bash
podman-compose -f docker-compose.security.yml config --quiet
```

**Step 1.1.2:** Fix any broken volume mounts, image references, or network configs
- Ensure all images use fully qualified names (`docker.io/library/...`)
- Ensure `--network host` where required
- Verify PostgreSQL backend for SonarQube reaches healthy state

**Step 1.1.3:** Verify scanning scripts are functional
- `scripts/run-sonarqube-scan.sh` — test execution, fix paths
- `scripts/security-scan.sh` — test execution, fix missing tools
- `scripts/snyk-scan.sh` — test Snyk CLI availability
- `scripts/gosec-scan.sh` — test gosec availability
- `scripts/nancy-scan.sh` — test nancy availability

**Step 1.1.4:** Install missing tools if not containerized
- Ensure `govulncheck` installed: `go install golang.org/x/vuln/cmd/govulncheck@latest`
- Ensure `npm audit` available via Node.js

### 1.2 Execute Scans (All Platforms)

| Scanner | Target | Command | Output |
|---------|--------|---------|--------|
| SonarQube | catalog-api (Go), catalog-web (TS) | `scripts/run-sonarqube-scan.sh` | Quality gate report |
| Snyk | All go.mod, package.json, containers | `podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli` | Dependency vulns |
| Semgrep | All Go, TS, Kotlin, Rust source | `podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner` | SAST findings |
| Trivy | Container images, filesystem | `podman-compose -f docker-compose.security.yml --profile trivy-scan run --rm trivy-scanner` | CVE + secrets |
| Dependency-Check | All dependencies | `podman-compose -f docker-compose.security.yml --profile dependency-check run --rm dependency-check` | OWASP CVEs |
| govulncheck | catalog-api Go stdlib/deps | `cd catalog-api && govulncheck ./...` | Go vulns |
| npm audit | catalog-web, all TS submodules | `cd catalog-web && npm audit --audit-level=moderate` | NPM vulns |

### 1.3 Triage & Resolve Findings

**Priority order:**
1. **Critical** — Fix immediately (remote code execution, SQL injection, auth bypass)
2. **High** — Fix in this phase (XSS, SSRF, insecure deserialization)
3. **Medium** — Fix in this phase (missing security headers, weak crypto, info disclosure)
4. **Low/Info** — Document, fix in Phase 2 if code-adjacent

**Resolution approach:**
- Dependency vulnerabilities: Update to patched versions, test for breakage
- Code vulnerabilities: Apply fix, add regression test
- Configuration vulnerabilities: Fix config, add validation test
- Container vulnerabilities: Update base images, rebuild

### 1.4 Deliverables

- [ ] All 7 scanners operational via Podman compose
- [ ] Scan reports saved to `reports/security/`
- [ ] All Critical findings resolved
- [ ] All High findings resolved
- [ ] All Medium findings resolved
- [ ] Regression scan confirms zero new issues
- [ ] Security scan documentation updated

---

## Phase 2: Safety Hardening

### 2.1 catalog-api (Go)

#### Race Condition Audit
**Step 2.1.1:** Run comprehensive race detection
```bash
cd catalog-api && GOMAXPROCS=3 go test -race ./... -p 2 -parallel 2 -count=1
```

**Step 2.1.2:** Audit all goroutine spawns
- Search for `go func()` in all non-test Go files
- Verify each has: proper context cancellation, channel closure, WaitGroup, or sync.Once
- Verify shared state access is mutex-protected

**Step 2.1.3:** Specific audit targets
- `internal/services/cache_service.go` — cleanup goroutine lifecycle
- `handlers/websocket_handler.go` — readPump/writePump goroutines
- `internal/smb/resilience.go` — health checker goroutines
- `internal/media/realtime/` — event bus goroutines
- `internal/services/sync_service.go` — sync operation copies vs shared pointers
- `internal/services/log_management_service.go` — log collection copies vs shared pointers

**Step 2.1.4:** Fix any races found
- Add mutex protection where missing
- Add proper channel closure sequences
- Add context cancellation propagation
- Add timeout contexts to all blocking operations

#### Memory Leak Audit
**Step 2.1.5:** Audit goroutine lifecycle
- Every `go func()` must have a clear termination path
- Every channel must have a clear close path
- Every context must have a cancel call

**Step 2.1.6:** Add leak detection tests
```go
// Pattern: verify goroutine count returns to baseline
func TestNoGoroutineLeaks(t *testing.T) {
    before := runtime.NumGoroutine()
    // ... create and destroy service ...
    runtime.GC()
    time.Sleep(100 * time.Millisecond)
    after := runtime.NumGoroutine()
    assert.InDelta(t, before, after, 2, "goroutine leak detected")
}
```

**Step 2.1.7:** Audit resource cleanup
- All `defer file.Close()` patterns correct
- All database connections returned to pool
- All HTTP response bodies closed
- All temporary files cleaned up

#### Deadlock Prevention
**Step 2.1.8:** Audit mutex acquisition ordering
- No nested lock acquisitions (mutex A then mutex B)
- All blocking operations have timeout contexts
- All channel operations have `select` with `default` or timeout

### 2.2 catalog-web (React/TypeScript)

**Step 2.2.1:** Audit useEffect cleanup
- Every `useEffect` with subscriptions must return cleanup function
- Every `useEffect` with timers (setTimeout, setInterval) must clear on unmount
- Every `useEffect` with event listeners must remove on unmount

**Step 2.2.2:** Audit fetch/API calls
- Every fetch call should use AbortController
- Abort on component unmount via useEffect cleanup
- React Query handles this automatically — verify all API calls use React Query

**Step 2.2.3:** Audit WebSocket connections
- Verify WebSocketContext handles reconnection without leaking connections
- Verify cleanup on unmount
- Verify message handlers don't accumulate

### 2.3 HelixQA (Go)

**Step 2.3.1:** Run race detection
```bash
cd HelixQA && go test -race ./... -count=1
```

**Step 2.3.2:** Verify SQLite session memory store concurrent access
- Check mutex protection on read/write operations
- Add concurrent access stress test if not present

### 2.4 Android/AndroidTV (Kotlin)

**Step 2.4.1:** Audit coroutine scopes
- All coroutines launched in `viewModelScope` (auto-cancelled on ViewModel clear)
- No `GlobalScope` usage (leaks)
- All `withContext(Dispatchers.IO)` for blocking operations

**Step 2.4.2:** Audit Room database
- All transactions use `@Transaction` annotation
- No raw queries without proper threading
- WAL mode enabled for concurrent reads

**Step 2.4.3:** Audit Retrofit/OkHttp
- Call cancellation on lifecycle destroy
- Proper timeout configuration
- Connection pool limits

### 2.5 Desktop/Installer (Tauri/Rust)

**Step 2.5.1:** Audit Rust async tasks
- All `tokio::spawn` tasks have proper cancellation via `JoinHandle`
- No blocking operations on main thread
- IPC command handlers return promptly

**Step 2.5.2:** Audit React frontend
- Same patterns as catalog-web (useEffect cleanup, fetch abort)

### 2.6 Go Submodules (All 22)

**Step 2.6.1:** Run race detection on ALL submodules
```bash
for dir in Auth Cache Config Concurrency Containers Database Discovery \
    Entities EventBus Filesystem Lazy Memory Middleware Observability \
    RateLimiter Recovery Security Storage Streaming Watcher; do
    echo "=== $dir ===" && cd "$dir" && go test -race ./... -count=1 && cd ..
done
```

**Step 2.6.2:** Fix any races found in submodules

**Step 2.6.3:** Add safety tests for concurrent access patterns in each submodule

### 2.7 Deliverables

- [ ] Zero race conditions (all `-race` tests pass across all Go modules)
- [ ] Zero memory leaks (leak detection tests pass)
- [ ] Zero deadlock risks (timeout contexts on all blocking ops)
- [ ] Safety test suite added to catalog-api
- [ ] Safety audit for catalog-web (useEffect cleanup, fetch abort)
- [ ] Safety audit for Android/AndroidTV (coroutine scopes, Room threading)
- [ ] Safety audit for Desktop/Installer (Rust async, React cleanup)
- [ ] All 22 Go submodules pass `-race` tests

---

## Phase 3: Dead Code & Completion

### 3.1 Wire Unused Middleware (catalog-api)

**Step 3.1.1:** Wire Redis rate limiting (behind config toggle)
- Add `rate_limiter.type` config option: `memory` (default) | `redis`
- When `redis`: use `RedisRateLimit()` / `SlidingWindowRedisRateLimit()` / `TokenBucketRedisRateLimit()`
- When `memory`: use current in-memory rate limiting
- Add integration tests for Redis rate limiting variants
- Graceful fallback to memory if Redis unavailable

**Step 3.1.2:** Wire advanced rate limiting
- Add `rate_limiter.strategy` config option: `basic` (default) | `advanced` | `user` | `ip`
- Wire `AdvancedRateLimit()`, `UserBasedRateLimit()`, `IPRateLimit()` based on config
- Add integration tests for each strategy

**Step 3.1.3:** Wire cache headers
- Wire `CacheHeaders()` for API responses with appropriate max-age
- Wire `StaticCacheHeaders()` for static asset responses (long cache)
- Add tests verifying correct header values

### 3.2 Implement Remaining Providers (catalog-api)

**Step 3.2.1:** Implement IMDBProvider
- `Search()` — OMDB API integration (shares API key, already available)
- `GetDetails()` — OMDB API integration for movie/TV details
- Add unit tests with mocked HTTP responses

**Step 3.2.2:** Implement TVDBProvider
- `GetDetails()` — TheTVDB API v4 (free tier available)
- Add unit tests with mocked HTTP responses
- Graceful degradation if API key not configured

**Step 3.2.3:** Implement MusicBrainzProvider
- `GetDetails()` — MusicBrainz API (free, no key needed)
- Add unit tests with mocked HTTP responses

**Step 3.2.4:** Implement remaining providers where free APIs exist
- Spotify (Web API, requires OAuth — implement with graceful degradation)
- LastFM (free API key available)
- For providers without free APIs (AniDB, MyAnimeList, IGDB, GiantBomb):
  - Implement with clear "API key required" documentation
  - Graceful degradation returns nil with info log
  - Add unit tests for both key-present and key-absent scenarios

### 3.3 Replace Mock APIs (catalog-web)

**Step 3.3.1:** Implement real adminApi.ts
- Create backend endpoints if missing:
  - `GET /api/v1/admin/system-info` — system information
  - `GET /api/v1/admin/users` — user management list
  - `POST /api/v1/admin/users` — create user
  - `PUT /api/v1/admin/users/:id` — update user
  - `DELETE /api/v1/admin/users/:id` — delete user
  - `GET /api/v1/admin/storage` — storage statistics
  - `GET /api/v1/admin/backups` — backup list
  - `POST /api/v1/admin/backups` — create backup
- Replace mock functions with real axios/fetch calls
- Add proper error handling
- Add loading states
- Add unit tests for API layer

**Step 3.3.2:** Implement real conversionApi.ts
- Create backend endpoints:
  - `GET /api/v1/conversion/formats` — supported format list
  - `POST /api/v1/conversion/jobs` — create conversion job
  - `GET /api/v1/conversion/jobs` — list conversion jobs
  - `GET /api/v1/conversion/jobs/:id` — job status/progress
  - `DELETE /api/v1/conversion/jobs/:id` — cancel job
  - `GET /api/v1/conversion/jobs/:id/download` — download result
- Replace mock functions with real API calls
- Add progress tracking via WebSocket or polling

**Step 3.3.3:** Fix API path inconsistency
- Change `collectionsApi.ts` from `/api/collections` to `/api/v1/collections`
- Verify all other API paths use `/api/v1/` prefix consistently
- Add API path validation test

### 3.4 Add Granular Error Boundaries (catalog-web)

**Step 3.4.1:** Create page-level error boundaries
- Wrap each page component in its own ErrorBoundary
- Each boundary has page-specific fallback UI
- Retry button that resets the boundary state
- Error reporting to console (or future error tracking service)

**Step 3.4.2:** Add critical section boundaries
- Media player — isolate playback failures from page
- Collection manager — isolate collection operations
- Entity browser — isolate entity loading

### 3.5 Add Missing Package Tests (catalog-api)

**Step 3.5.1:** `cmd/boot` tests
- Test startup sequence initialization
- Test configuration loading
- Test graceful shutdown signal handling

**Step 3.5.2:** `internal/monitoring` tests
- Test metrics collection initialization
- Test Prometheus endpoint registration
- Test metric increment/observe operations

### 3.6 Deliverables

- [ ] All 8 middleware functions wired with config toggles
- [ ] Middleware integration tests pass
- [ ] IMDBProvider Search/GetDetails implemented and tested
- [ ] TVDBProvider GetDetails implemented and tested
- [ ] MusicBrainzProvider GetDetails implemented and tested
- [ ] Remaining providers implemented or documented with graceful degradation
- [ ] adminApi.ts real implementation with backend endpoints
- [ ] conversionApi.ts real implementation with backend endpoints
- [ ] collectionsApi.ts path fixed to `/api/v1/collections`
- [ ] Granular error boundaries on all pages
- [ ] cmd/boot tests added
- [ ] internal/monitoring tests added
- [ ] Zero dead/unwired code remaining

---

## Phase 4: Test Coverage Maximum

### 4.1 Unit Test Coverage Targets

#### catalog-api: Target 90%+ line coverage
**Step 4.1.1:** Generate current coverage baseline
```bash
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -coverprofile=coverage.out -count=1
go tool cover -func=coverage.out | tail -1
```

**Step 4.1.2:** Identify uncovered code paths
```bash
go tool cover -html=coverage.out -o coverage.html
```

**Step 4.1.3:** Add tests for uncovered paths
- Every handler: test all HTTP methods, status codes, error paths
- Every service: test all methods including edge cases
- Every repository: test CRUD, pagination, search, error paths
- Every middleware: test allow/deny, edge cases
- Table-driven tests for functions with multiple inputs

#### catalog-web: Target 95%+ line coverage
**Step 4.1.4:** Generate current coverage baseline
```bash
cd catalog-web && npm run test:coverage
```

**Step 4.1.5:** Fill coverage gaps
- Test all component props and states
- Test all hook edge cases (loading, error, empty)
- Test all utility functions
- Test all API layer functions (success, error, timeout)

#### HelixQA: Maintain 1.08:1 ratio
**Step 4.1.6:** Add basic tests for `types` package (validation)

#### Go Submodules: Target 85%+ per submodule
**Step 4.1.7:** Run coverage on each submodule and add tests where below target

#### Android/AndroidTV: Add ViewModel and Repository tests
**Step 4.1.8:** Add tests for all ViewModels (StateFlow emissions, error handling)
**Step 4.1.9:** Add tests for all Repositories (mock Room, mock Retrofit)

#### Desktop/Installer: Add Rust and React tests
**Step 4.1.10:** Add Rust unit tests for all IPC command handlers
**Step 4.1.11:** Add React component tests matching catalog-web patterns

### 4.2 Integration Tests (Cross-Platform)

| Test | Components | What It Validates |
|------|-----------|------------------|
| API-Database | catalog-api + SQLite/PostgreSQL | All CRUD operations, transactions, migrations, dialect rewriting |
| API-Filesystem | catalog-api + SMB/FTP/NFS/WebDAV | Protocol operations, circuit breaker, offline cache, retry |
| API-WebSocket | catalog-api + WebSocket clients | Event propagation, broadcast, connection lifecycle |
| API-Redis | catalog-api + Redis | Cache operations, TTL, eviction, graceful fallback |
| Web-API | catalog-web + catalog-api | All 36+ endpoints, auth flow, error handling, WebSocket |
| Desktop-API | catalogizer-desktop + catalog-api | Tauri IPC to HTTP chain, offline mode |
| Android-API | catalogizer-android + catalog-api | Retrofit to HTTP/3 chain, error handling |
| HelixQA-API | HelixQA + catalog-api | Challenge execution against live API |
| Entity-Pipeline | Scanner + Aggregation + Entity API | End-to-end media detection and entity creation |
| Auth-Flow | Login + JWT + Protected Routes | Full authentication lifecycle across platforms |

### 4.3 Stress Tests (Responsiveness Validation)

| Test | Parameters | Success Criteria | Implementation |
|------|-----------|-----------------|----------------|
| API concurrent requests | 500 concurrent, 60s | p95 < 500ms, 0 errors | k6 `stress_test.js` enhanced |
| WebSocket connections | 200 concurrent | Broadcast < 100ms | Custom Go stress test |
| Database under load | 1000 queries/sec | p99 < 200ms | Custom Go stress test |
| File scanning storm | 10 concurrent scans, 10K files | No OOM, all complete | Custom Go stress test |
| Frontend rendering | 1000 media items | 60fps, < 100ms interaction | Playwright performance test |
| Memory soak | 30min sustained load | Memory growth < 10% | k6 `soak_test.js` enhanced |
| Challenge marathon | All challenges sequential | All pass, no timeout | HelixQA orchestrator |
| Spike traffic | 0→300 users in 10s | Recovery < 30s | k6 `spike_test.js` |
| Provider storm | 50 concurrent metadata lookups | All resolve, no deadlock | Custom Go stress test |
| Rate limiter saturation | 10x limit requests | Correct 429 responses, no leak | Custom Go stress test |

### 4.4 Monitoring & Metrics Tests

**Step 4.4.1:** Test Prometheus metrics endpoint
```go
func TestPrometheusMetricsEndpoint(t *testing.T) {
    resp := httptest.NewRecorder()
    // GET /metrics
    // Assert: response contains expected metric families
    // Assert: http_requests_total, http_request_duration_seconds exist
}
```

**Step 4.4.2:** Test health check endpoints under load
- `/health` returns 200 within 100ms under 500 concurrent requests
- `/ready` returns 200 only when all dependencies healthy

**Step 4.4.3:** Test alerting rules
- Simulate high error rate → verify alert would fire
- Simulate high latency → verify alert would fire
- Simulate disk full → verify alert would fire

**Step 4.4.4:** Add metrics for Phase 5 optimization baseline
- Request latency histograms per endpoint
- Error rate counters per endpoint
- Goroutine count gauge
- Memory usage gauge
- Database connection pool stats (active, idle, waiting)
- Cache hit/miss ratio

### 4.5 Test Bank Expansion (HelixQA)

**New YAML test banks:**

| Bank File | Cases | Purpose |
|-----------|-------|---------|
| `entity-management.yaml` | 25 | Entity CRUD, hierarchy, search, duplicates |
| `collection-advanced.yaml` | 25 | Smart collections, templates, real-time sync |
| `ai-features.yaml` | 20 | AI metadata, analysis, recommendations |
| `subtitle-operations.yaml` | 20 | Subtitle search, download, sync, translate |
| `admin-operations.yaml` | 20 | Admin panel, user management, system health |
| `conversion-pipeline.yaml` | 20 | Format conversion, progress tracking |
| `performance-validation.yaml` | 25 | Response time validation, load handling |
| `security-validation.yaml` | 20 | Auth bypass attempts, injection, CSRF |
| `cross-platform-sync.yaml` | 20 | Desktop-mobile-web sync scenarios |
| `accessibility.yaml` | 15 | WCAG compliance, screen reader, keyboard nav |

**Target: 507 new test cases, 804 total (from 297)**

### 4.6 Deliverables

- [ ] catalog-api coverage: 90%+ (coverage report in `reports/coverage/`)
- [ ] catalog-web coverage: 95%+ (coverage report)
- [ ] HelixQA coverage maintained at 1.08:1 ratio
- [ ] Go submodules: 85%+ per submodule
- [ ] Android/AndroidTV: ViewModel and Repository tests added
- [ ] Desktop/Installer: Rust and React tests added
- [ ] 10 integration test suites passing
- [ ] 10 stress tests passing with defined criteria
- [ ] Monitoring tests validating observability stack
- [ ] 10 new test bank YAML files (507 new cases)
- [ ] All tests documented in test reports

---

## Phase 5: Optimization

### 5.1 Lazy Loading & Lazy Initialization

#### catalog-api (Go)

**Step 5.1.1:** Extend LazyServiceRegistry
- Register ALL services in LazyServiceRegistry (not just current subset)
- Services initialize on first access, not at startup
- Dependency ordering preserved via registry

**Step 5.1.2:** Lazy media providers
- Provider instances created on first Search/GetDetails call
- Provider configuration validated at registration, not instantiation
- Pool of initialized providers cached after first use

**Step 5.1.3:** Lazy filesystem clients
- SMB/FTP/NFS/WebDAV clients created on first access per storage root
- Connection pooling after initialization
- Idle client cleanup after configurable timeout

**Step 5.1.4:** Lazy Redis connection
- Connect to Redis on first cache operation
- Graceful fallback to in-memory cache if Redis unavailable
- Periodic reconnection attempts if initial connection fails

#### catalog-web (React)

**Step 5.1.5:** Comprehensive React.lazy + Suspense
- Audit all route-level components — ensure ALL use React.lazy
- Lazy-load chart libraries (recharts) only on Analytics/Dashboard
- Lazy-load media player only when user clicks play
- Lazy-load AI components only on AIDashboard page
- Lazy-load admin panel components only for admin users

**Step 5.1.6:** Image and data lazy loading
- IntersectionObserver-based image loading for media grids
- Virtual scrolling for lists > 100 items (media browser, collections, playlists)
- Pagination with infinite scroll for large datasets

#### Android/AndroidTV (Kotlin)

**Step 5.1.7:** Lazy Hilt module initialization
- Non-critical Hilt modules use `@InstallIn(SingletonComponent)` with lazy providers
- Room database created on first access
- Image loading with Coil lazy configuration

#### Desktop/Installer (Tauri)

**Step 5.1.8:** Lazy Rust service initialization
- Defer heavy operations (filesystem scanning, network discovery) to first use
- Frontend: same React lazy patterns as catalog-web

### 5.2 Semaphore & Concurrency Control

#### catalog-api

**Step 5.2.1:** Apply semaphore to ALL parallel operations
- File scanning: limit to `config.catalog.max_concurrent_scans` (default 3)
- Media detection pipeline: limit concurrent provider API calls (default 5)
- Thumbnail generation: limit concurrent image processing (default 3)
- Archive extraction: limit concurrent decompression (default 2)
- Background tasks: configurable worker pool size

**Step 5.2.2:** Implement backpressure on WebSocket broadcasts
- Drop messages if consumer not keeping up (verify existing implementation)
- Add per-client message buffer with configurable size
- Slow consumer detection and graceful disconnect

#### catalog-web

**Step 5.2.3:** Request deduplication
- Deduplicate concurrent requests for same resource
- Debounce all search inputs (300ms default)
- Throttle WebSocket message processing

#### Go Submodules

**Step 5.2.4:** Add semaphores where applicable
- `Streaming/` — semaphore on concurrent stream operations
- `Watcher/` — semaphore on concurrent file watch events
- `Storage/` — semaphore on concurrent storage operations
- `EventBus/` — backpressure on event processing

### 5.3 Non-Blocking Mechanisms

#### catalog-api

**Step 5.3.1:** Async scan operations
- Convert synchronous scan to async with progress channels
- Return scan ID immediately, stream progress via WebSocket
- Non-blocking scan status queries

**Step 5.3.2:** Non-blocking infrastructure
- Health checks with timeout (100ms max, fail-open)
- Metrics collection via buffered channel + periodic flush
- Async log writes via buffered writer

#### catalog-web

**Step 5.3.3:** Optimistic UI updates
- Favorites toggle: update UI immediately, reconcile on server response
- Playlist reorder: update UI immediately, reconcile on server response
- Collection operations: show pending state, confirm on response

**Step 5.3.4:** Navigation preloading
- Preload next-page data on hover/focus (React Query prefetch)
- Background sync for data that may have changed

#### Android

**Step 5.3.5:** Non-blocking UI
- All network/DB operations on IO dispatcher (verify)
- Non-blocking image decoding with Coil async pipeline
- Skeleton screens during loading

### 5.4 Performance Tuning Based on Metrics

**Step 5.4.1:** Run k6 load tests
```bash
podman run --rm --network host -v $(pwd)/tests/k6:/scripts \
    docker.io/grafana/k6:latest run /scripts/load_test.js
```

**Step 5.4.2:** Analyze Prometheus metrics under load
- Identify slow endpoints (p99 > 500ms)
- Identify high-error-rate endpoints
- Identify memory pressure points
- Identify database query bottlenecks

**Step 5.4.3:** Apply targeted optimizations
- Add database indexes for slow queries identified by profiling
- Add HTTP response caching for static/rarely-changing data
- Tune connection pool sizes based on observed concurrency
- Tune GOMAXPROCS and GC settings based on observed behavior
- Optimize SQL queries identified as bottlenecks

### 5.5 Deliverables

- [ ] All services lazy-initialized (API startup time < 2s)
- [ ] Web app initial load < 1s (lazy-loaded routes)
- [ ] Semaphores on all parallel operations (configurable limits)
- [ ] Non-blocking patterns for all user-facing operations
- [ ] k6 results: p95 < 500ms under 50 concurrent users
- [ ] Soak test: zero memory growth over 30 minutes
- [ ] Performance tuning applied based on metrics analysis
- [ ] Configuration options documented for all new tuning knobs

---

## Phase 6: Challenge Coverage

### 6.1 New API Challenges

| ID Range | Category | Count | What It Tests |
|----------|----------|-------|--------------|
| CH-100–110 | Admin API | 11 | System info, user management, backups, storage config |
| CH-111–120 | Conversion API | 10 | Format conversion jobs, progress, download |
| CH-121–130 | Entity System | 10 | Entity CRUD, hierarchy, duplicate detection, search |
| CH-131–140 | Collection Management | 10 | Smart collections, templates, sharing, analytics |
| CH-141–150 | Subtitle System | 10 | Search, download, upload, translate, sync |
| CH-151–160 | AI Features | 10 | AI metadata, analysis, recommendations |
| CH-161–170 | Middleware Stack | 10 | Rate limiting (all variants), cache headers, security headers |
| CH-171–180 | Performance | 10 | Response times, throughput, resource limits, backpressure |

**Total: 81 new API challenges**

### 6.2 New Web User Flow Challenges

| ID Range | Category | Count |
|----------|----------|-------|
| UF-W-060–070 | Entity Browser & Detail | 11 |
| UF-W-071–080 | Collection Manager Advanced | 10 |
| UF-W-081–090 | Subtitle Manager | 10 |
| UF-W-091–100 | AI Dashboard | 10 |
| UF-W-101–110 | Conversion Tools | 10 |
| UF-W-111–120 | Admin Panel | 10 |

**Total: 61 new web challenges**

### 6.3 New Desktop Challenges

| ID Range | Category | Count |
|----------|----------|-------|
| UF-D-029–038 | Offline Mode & Sync | 10 |
| UF-D-039–048 | File Management Advanced | 10 |

**Total: 20 new desktop challenges**

### 6.4 New Mobile Challenges

| ID Range | Category | Count |
|----------|----------|-------|
| UF-M-039–048 | Android Compose UI Flows | 10 |
| UF-M-049–058 | AndroidTV Navigation | 10 |
| UF-M-059–068 | Offline & Sync | 10 |

**Total: 30 new mobile challenges**

### 6.5 HelixQA Test Bank Expansion

10 new YAML test banks with 507 new test cases (see Phase 4.5 for details).

**Grand total: 804 test bank cases (from 297)**

### 6.6 Challenge Registration

- All new challenges registered in `catalog-api/challenges/register.go`
- API challenges via `RegisterAdminChallenges()`, `RegisterConversionChallenges()`, etc.
- Web challenges via `RegisterUserFlowWebChallengesExtended()`
- Desktop challenges via `RegisterUserFlowDesktopChallengesExtended()`
- Mobile challenges via `RegisterUserFlowMobileChallengesExtended()`

### 6.7 Deliverables

- [ ] 81 new API challenges implemented and registered
- [ ] 61 new Web user flow challenges implemented
- [ ] 20 new Desktop challenges implemented
- [ ] 30 new Mobile challenges implemented
- [ ] 10 new test bank YAML files (507 new cases)
- [ ] All 192 new challenges pass against running system
- [ ] Challenge documentation updated (IDs, descriptions, categories)

---

## Phase 7: Documentation & Courses

### 7.1 Extend Existing Architecture Documentation

| Document | Updates |
|----------|---------|
| `docs/architecture/DATABASE_SCHEMA.md` | Add admin, conversion, subtitle tables from Phase 3 |
| `docs/architecture/ARCHITECTURE.md` | Add lazy loading, semaphore, non-blocking patterns from Phase 5 |
| `docs/architecture/CONCURRENCY_PATTERNS.md` | Add new semaphore patterns, backpressure from Phase 5 |
| `docs/architecture/OPTIMIZATION_GUIDE.md` | Add Phase 5 optimization results and metrics |
| `docs/architecture/LAZY_LOADING.md` | Comprehensive update with all lazy patterns |

### 7.2 New Architecture Decision Records

| ADR | Title | Decision |
|-----|-------|----------|
| ADR-007 | Lazy Initialization Strategy | All services lazy-init via LazyServiceRegistry |
| ADR-008 | Semaphore-Based Concurrency Control | Configurable semaphores on all parallel operations |
| ADR-009 | Non-Blocking UI Patterns | Optimistic updates, preloading, virtual scrolling |

### 7.3 Extend API Documentation

| Document | Updates |
|----------|---------|
| `docs/api/openapi.yaml` | Add admin, conversion, subtitle, entity endpoints |
| `docs/api/API_DOCUMENTATION.md` | Add all new endpoint documentation |
| New: `docs/api/ADMIN_API.md` | Admin endpoint reference with examples |
| New: `docs/api/CONVERSION_API.md` | Conversion endpoint reference with examples |
| New: `docs/api/ENTITY_API.md` | Entity system endpoint reference (extend existing) |

### 7.4 Extend Testing Documentation

| Document | Updates |
|----------|---------|
| `docs/testing/TEST_STRATEGY.md` | Add stress, monitoring, safety test sections |
| New: `docs/testing/STRESS_TEST_REPORT.md` | Phase 4 stress test results and analysis |
| New: `docs/testing/SECURITY_SCAN_REPORT.md` | Phase 1 scan results and resolutions |
| `docs/testing/FINAL_TEST_REPORT.md` | Updated coverage numbers from all phases |

### 7.5 Extend Deployment Documentation

| Document | Updates |
|----------|---------|
| `docs/deployment/MONITORING_GUIDE.md` | New metrics from Phase 5 |
| `docs/deployment/PRODUCTION_RUNBOOK.md` | New admin/conversion operations |

### 7.6 New User Manuals

| Document | Platform | Content |
|----------|----------|---------|
| `docs/manuals/ANDROID_USER_MANUAL.md` | Android | Complete step-by-step user manual |
| `docs/manuals/ANDROIDTV_USER_MANUAL.md` | AndroidTV | Remote-control-focused user manual |
| `docs/manuals/DESKTOP_USER_MANUAL.md` | Desktop | Tauri desktop app user manual |
| `docs/manuals/INSTALLER_WIZARD_MANUAL.md` | Installer | Setup wizard step-by-step guide |
| `docs/manuals/ENTITY_USER_GUIDE.md` | All | Entity browser and media management |
| `docs/manuals/COLLECTION_USER_GUIDE.md` | All | Collection management complete guide |
| `docs/manuals/AI_FEATURES_GUIDE.md` | All | AI-powered features walkthrough |
| `docs/manuals/SUBTITLE_USER_GUIDE.md` | All | Subtitle management guide |
| `docs/manuals/ADMIN_USER_GUIDE.md` | Web | Admin panel operations guide |

### 7.7 Extend Existing User Guide
- `docs/USER_GUIDE.md` — Add sections for entity browser, collections, AI, subtitles, conversion
- `docs/INSTALLATION_GUIDE.md` — Add Android/Desktop/AndroidTV installation steps
- `docs/CONFIGURATION_GUIDE.md` — Add new config options from Phase 5

### 7.8 SQL Schema Documentation Updates
- Update `docs/architecture/SQL_COMPLETE_SCHEMA.md` with all tables
- Update `docs/architecture/SQL_MIGRATIONS.md` with new migrations
- Add index documentation (new indexes from Phase 5)
- Add query performance documentation

### 7.9 Diagram Updates
- Update ER diagrams with admin, conversion, subtitle tables
- Update sequence diagrams with lazy initialization flows
- New: Semaphore/concurrency control diagrams
- New: Non-blocking operation flow diagrams
- Re-render all SVGs from updated Mermaid sources

### 7.10 Video Course Extension

#### New Modules (extend from 18 to 24)

| Module | Title | Content (key sections) |
|--------|-------|----------------------|
| MODULE19 | Entity System Deep Dive | Entity creation, hierarchy building, metadata enrichment, entity browsing UI, duplicate detection |
| MODULE20 | Collection Management | Smart collections with rules, collection templates, sharing and permissions, real-time sync, collection analytics |
| MODULE21 | AI-Powered Features | AI metadata extraction, content analysis pipeline, recommendation engine, AI dashboard walkthrough |
| MODULE22 | HelixQA Autonomous Testing | Autonomous QA sessions, test bank authoring, evidence collection, LLM-driven exploration, curiosity mode |
| MODULE23 | Subtitle & Conversion Tools | Subtitle search/download/upload, translation pipeline, format conversion jobs, progress tracking |
| MODULE24 | Advanced Optimization | Lazy loading patterns (Go + React), semaphore-based concurrency, non-blocking UI, performance tuning methodology |

#### Update Existing Modules

| Module | Updates |
|--------|---------|
| MODULE2 (Backend Architecture) | Add lazy initialization, LazyServiceRegistry, semaphore sections |
| MODULE5 (React Frontend) | Add lazy components, virtual scrolling, optimistic UI sections |
| MODULE8 (Performance) | Update with Phase 5 optimization results and metrics |
| MODULE10 (Testing) | Add stress test, monitoring test, safety test sections |
| MODULE14 (Challenges) | Add new challenge categories from Phase 6 |
| MODULE16 (Security Scanning) | Update with Phase 1 scanning methodology and results |

### 7.11 HelixQA-Specific Documentation Updates

| Document | Updates |
|----------|---------|
| `HelixQA/USER_GUIDE.md` | Add new test banks from Phase 6 |
| `HelixQA/DEVELOPER_GUIDE.md` | Add new challenge patterns |
| `HelixQA/API_REFERENCE.md` | Add any new endpoints |
| `HelixQA/CONFIGURATION_GUIDE.md` | Add new config options |
| `HelixQA/DATA_DICTIONARY.md` | Add new data structures |
| New: `HelixQA/CHALLENGE_CATALOG.md` | Complete catalog of all challenges with descriptions |

### 7.12 Deliverables

- [ ] All existing architecture docs updated
- [ ] 3 new ADRs written
- [ ] API documentation updated with all new endpoints
- [ ] Testing documentation updated with all phase results
- [ ] Deployment documentation updated
- [ ] 9 new user manuals/guides written
- [ ] SQL schema documentation updated
- [ ] All diagrams updated and re-rendered
- [ ] 6 new video course module scripts written
- [ ] 6 existing video course modules updated
- [ ] HelixQA documentation updated
- [ ] All documentation reviewed for accuracy

---

## Phase 8: Website Content

### 8.1 HelixQA VitePress Website — Expand Existing Pages

| Page | Current State | Updates |
|------|--------------|---------|
| `index.md` | Basic hero | Complete hero with stats, feature highlights, quick start code block, platform badges |
| `features.md` | Feature list | Add autonomous QA, LLM integration (40+ providers), vision analysis, performance testing, evidence collection, session recording |
| `getting-started.md` | Basic guide | Step-by-step with code examples for each platform (API, Web, Android, Desktop) |
| `documentation.md` | Doc links | Link to all new guides from Phase 7, organized by audience (user, developer, admin) |
| `changelog.md` | Release notes | Update with all changes from Phases 1-7 |
| `download.md` | Download links | Add all platform download links, container images, package manager instructions |
| `faq.md` | FAQ list | Add questions about new features, optimization, challenge authoring |
| `support.md` | Support info | Add troubleshooting for common issues from all phases |

### 8.2 HelixQA VitePress Website — New Pages

| Page | Content |
|------|---------|
| `guides/autonomous-qa.md` | Autonomous QA session guide with configuration, examples, output analysis |
| `guides/test-banks.md` | Test bank YAML authoring guide with schema reference and examples |
| `guides/challenges.md` | Challenge development guide: creating, registering, running, debugging |
| `guides/evidence-collection.md` | Evidence and reporting guide: screenshots, video, logs, reports |
| `guides/ci-integration.md` | Integrating HelixQA into local CI pipelines (Podman-based) |
| `developer/architecture.md` | Full architecture reference with diagrams |
| `developer/extending.md` | Plugin/extension development: detectors, validators, reporters |
| `developer/llm-providers.md` | LLM provider configuration: registry, custom providers, API keys |
| `reference/cli.md` | Complete CLI reference: all commands, all flags, all options, examples |
| `reference/config.md` | Configuration file reference: all options, defaults, environment variables |
| `reference/test-bank-schema.md` | YAML test bank schema reference with field definitions |
| `reference/api.md` | REST API reference: all endpoints, request/response schemas |

### 8.3 VitePress Configuration Updates

**Step 8.3.1:** Update sidebar navigation
- Organize pages into sections: Getting Started, Guides, Developer, Reference
- Add collapsible sections for each group
- Add page ordering within sections

**Step 8.3.2:** Add search functionality
- Enable VitePress built-in local search
- Configure search to index all content pages

**Step 8.3.3:** Add metadata
- Version badge in header
- "Edit this page" links pointing to repository
- Last updated timestamps

**Step 8.3.4:** Ensure responsive design
- Test all pages on mobile viewports
- Verify sidebar navigation works on mobile
- Verify code blocks are scrollable on mobile

**Step 8.3.5:** Dark mode support
- Verify VitePress dark mode toggle works
- Test all content renders correctly in dark mode
- Custom theme adjustments if needed

### 8.4 Main Catalogizer Website Updates

If root `Website/` VitePress site needs updates:
- Update feature pages with entity system, collections, AI, subtitles
- Update platform pages for Android, AndroidTV, Desktop
- Update developer docs with new API endpoints
- Update changelog with all phase changes

### 8.5 Deliverables

- [ ] 8 existing HelixQA website pages expanded
- [ ] 12 new HelixQA website pages created
- [ ] Sidebar navigation updated with all pages
- [ ] Search functionality enabled
- [ ] Responsive design verified
- [ ] Dark mode support verified
- [ ] All content reviewed for accuracy against final codebase
- [ ] Website builds and serves correctly: `cd website && npm run dev`
- [ ] Main Catalogizer website updated (if applicable)

---

## Cross-Phase Constraints

| Constraint | How Enforced |
|-----------|-------------|
| Podman only (no Docker) | All compose commands use `podman-compose` |
| No GitHub Actions | Zero `.github/workflows/` files created |
| Container builds mandatory | All scanning, testing, QA via containers |
| Host resource limits (30-40%) | CPU/memory limits: `--cpus` and `--memory` on all containers |
| HTTP/3 (QUIC) + Brotli | Maintained throughout, tested in Phase 4 stress tests |
| Zero warning/error policy | Validated in Phase 4 stress and integration tests |
| Conventional Commits | All commits: `type(scope): message` format |
| SPDX headers | All new `.go` files include Apache 2.0 headers |
| No broken existing functionality | Phase 4 integration + regression tests verify |
| No interactive processes | No sudo, no root password prompts in any automation |
| Challenge execution via binaries only | Challenges use compiled catalog-api, never curl/scripts |

---

## Success Criteria

### Phase 1 Exit Criteria
- All 7 scanners operational
- Zero Critical/High/Medium security findings
- Reports saved and documented

### Phase 2 Exit Criteria
- `go test -race` passes on ALL Go modules (catalog-api + HelixQA + 22 submodules)
- Leak detection tests pass
- Safety audit checklists complete for all platforms

### Phase 3 Exit Criteria
- Zero dead/unwired code
- Zero mock-only APIs
- All providers implemented or documented with graceful degradation
- All packages have tests

### Phase 4 Exit Criteria
- catalog-api: 90%+ coverage
- catalog-web: 95%+ coverage
- Go submodules: 85%+ per module
- All 10 integration test suites pass
- All 10 stress tests pass with defined criteria
- 804 test bank cases available

### Phase 5 Exit Criteria
- API startup < 2s
- Web initial load < 1s
- p95 < 500ms under 50 concurrent users
- Zero memory growth in 30-minute soak test

### Phase 6 Exit Criteria
- 192 new challenges all pass
- 804 total test bank cases
- All challenges registered and documented

### Phase 7 Exit Criteria
- 9 new user manuals
- 3 new ADRs
- 6 new + 6 updated video course modules
- All documentation reviewed for accuracy

### Phase 8 Exit Criteria
- 20 website pages (8 expanded + 12 new)
- Search, responsive, dark mode all working
- Website builds cleanly

### Overall Project Exit Criteria
- **Zero** broken modules, applications, or libraries
- **Zero** disabled tests
- **Zero** dead code
- **Zero** mock APIs
- **Zero** security findings (Critical/High/Medium)
- **Zero** race conditions
- **Zero** memory leaks
- **100%** documentation coverage
- **100%** challenge coverage for all features
- **All** test types represented (unit, integration, stress, E2E, security, performance, monitoring)
- **All** platforms covered in every phase
