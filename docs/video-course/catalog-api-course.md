# Catalog API -- Go REST Backend Course

**Component**: catalog-api
**Language**: Go 1.25 / Gin framework
**Total Duration**: 7.5 hours (10 modules)
**Level**: Intermediate to Advanced

---

## Course Overview

This course covers the complete architecture of catalog-api, the Go backend that powers every Catalogizer client. You will learn the layered Handler-Service-Repository pattern, dual-dialect database abstraction, media detection pipeline, JWT authentication, WebSocket real-time events, multi-protocol filesystem access, the challenge quality-assurance system, Prometheus monitoring, security hardening, and performance optimization techniques used throughout the codebase.

---

### Module 1: Architecture Overview

**Duration**: 45 minutes
**Prerequisites**: Go fundamentals, basic HTTP/REST concepts

#### Learning Objectives
- Understand the Handler-Service-Repository layered architecture and how dependencies flow downward
- Trace the application boot sequence from `main.go` through service construction to route registration
- Distinguish between the dual package layout: top-level domain packages vs `internal/` infrastructure packages
- Explain how dynamic port binding and `.service-port` discovery work

#### Topics Covered
1. **Application entry point (`main.go`)**
   - Constructor injection: every service receives dependencies explicitly via `NewService(dep)` constructors
   - Version injection via `-ldflags`: `Version`, `BuildNumber`, `BuildDate` compiled into the binary
   - Graceful shutdown with OS signal handling (SIGINT, SIGTERM) and connection draining
2. **Layered architecture**
   - `handlers/` -- HTTP request parsing, response formatting, Gin context management
   - `services/` -- business logic, orchestration, validation
   - `repository/` -- data access, SQL queries, transaction management
   - Dependency direction: handlers depend on services, services depend on repositories, repositories depend on `database.DB`
3. **Dual package layout**
   - Top-level `handlers/`, `services/`, `repository/`, `middleware/` for domain logic (auth, media, collections, configuration)
   - `internal/handlers/`, `internal/services/`, `internal/middleware/` for infrastructure (metrics, health, caching, lifecycle)
   - Go's `internal/` visibility rule enforcing encapsulation
4. **Dynamic port binding**
   - `findAvailablePort()` probing with `digital.vasic.discovery` TCP discoverer
   - Writing bound port to `.service-port` file for frontend proxy auto-configuration
   - Fallback to port 8080 when no `.service-port` is found
5. **Route registration**
   - All routes mounted under `/api/v1` with Gin router groups
   - Handler method naming convention: `GetX`, `CreateX`, `UpdateX`, `DeleteX`
   - Middleware chain applied per group: auth, rate limiting, input validation, security headers

#### Hands-On Exercise
Clone the catalog-api repository and trace the boot sequence by reading `main.go`. Draw a dependency graph showing which services depend on which repositories. Start the server with `go run main.go` and verify that `.service-port` is created. Use `curl` to hit the `/api/v1/health` endpoint and examine the response structure.

#### Key Takeaways
- The entire application is wired through constructor injection with zero global state
- Domain logic and infrastructure concerns are separated by Go's `internal/` package boundary
- Dynamic port binding enables multiple instances and avoids port conflicts during development
- Every handler follows the same pattern: parse request, call service, format response

---

### Module 2: Database Layer

**Duration**: 60 minutes
**Prerequisites**: Module 1, SQL fundamentals, understanding of SQLite and PostgreSQL

#### Learning Objectives
- Explain the dual-dialect abstraction that lets the same Go code run against both SQLite and PostgreSQL
- Trace how `database.DB` shadows `*sql.DB` methods to auto-rewrite SQL syntax
- Write queries using the portable placeholder style and understand the rewriting pipeline
- Apply the migration system to evolve the schema across both dialects

#### Topics Covered
1. **Dialect abstraction (`database/dialect.go`)**
   - `DialectType` enum: `DialectSQLite` and `DialectPostgres`
   - `RewritePlaceholders()` -- converts `?` markers to `$1, $2, ...` for PostgreSQL
   - `RewriteInsertOrIgnore()` -- converts `INSERT OR IGNORE` to `INSERT ... ON CONFLICT DO NOTHING`
   - `BooleanLiterals()` -- rewrites `= 0`/`= 1` to `= FALSE`/`= TRUE` for known boolean columns
2. **Wrapped database operations (`database/connection.go`)**
   - `database.DB` struct wrapping `*sql.DB` with dialect field
   - Shadowed `Exec()`, `Query()`, `QueryRow()` methods that pass SQL through the rewrite pipeline before execution
   - `InsertReturningID()` and `TxInsertReturningID()` replacing `LastInsertId()` for PostgreSQL compatibility (uses `RETURNING id`)
   - `database.WrapDB(sqlDB, DialectSQLite)` factory for unit tests with in-memory SQLite
3. **Connection setup**
   - Explicit `PRAGMA journal_mode=WAL` after SQLite connection (go-sqlcipher ignores connection-string pragmas)
   - Connection pool defaults: MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=3m
   - PostgreSQL connection via environment variables: `DB_TYPE`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
4. **Migration system (`database/migrations/`)**
   - Separate migration files for SQLite (`migrations_sqlite.go`) and PostgreSQL (`migrations_postgres.go`)
   - 13+ migration versions: v1 (core tables) through v13 (playlist tables)
   - v8: media entity tables (media_types, media_items, media_files, media_collections, external_metadata)
   - v9: performance indexes with `media_files` deduplication before unique index creation
   - v10: sync tables, v11: service tables, v13: playlist tables
5. **SQLCipher encryption**
   - Encrypted SQLite via go-sqlcipher import
   - Transparent encryption at the database file level

#### Hands-On Exercise
Create an in-memory SQLite test database using `database.WrapDB()`. Write a query using `?` placeholders and examine the rewritten SQL for both dialects. Add a new migration version that creates a custom table with both SQLite and PostgreSQL variants. Run the migration and verify the schema.

#### Key Takeaways
- Write SQL once using SQLite syntax; the dialect layer rewrites it transparently for PostgreSQL
- Never use `LastInsertId()` -- always use `InsertReturningID()` for cross-dialect compatibility
- WAL mode must be set explicitly after connection, not in the connection string
- Migrations maintain separate SQLite and PostgreSQL variants to handle syntax differences that cannot be auto-rewritten

---

### Module 3: Media Detection Pipeline

**Duration**: 50 minutes
**Prerequisites**: Module 2, familiarity with media file formats

#### Learning Objectives
- Trace a scanned file through the full detection-to-entity pipeline
- Understand the title parser regex system for extracting structured metadata from filenames
- Explain how external metadata providers (TMDB, OMDB, OpenLibrary, MusicBrainz) enrich entities
- Describe the aggregation service that builds hierarchical relationships (show-season-episode, artist-album-song)

#### Topics Covered
1. **Scanner to entity flow**
   - `UniversalScanner` completes a filesystem scan and triggers the post-scan hook
   - `AggregationService.AggregateAfterScan()` processes scanned files into structured entities
   - Pipeline: title parsing, MediaItem creation/update, MediaFile junction linking, hierarchy building, duplicate detection
2. **Title parser (`internal/services/title_parser.go`)**
   - Regex-based parsers for each media type: movie, TV show, music, game, software, book, comic
   - Extraction of title, year, season number, episode number, quality indicators, codec information
   - Handling of edge cases: multi-episode files, special characters, internationalized titles (Cyrillic, CJK)
3. **11 media types**
   - Seeded in `media_types` table: movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic
   - Hierarchical relationships: tv_show contains tv_season contains tv_episode; music_artist contains music_album contains song
   - `parent_id` self-referencing in `media_items` for hierarchy
4. **Metadata providers (`internal/media/providers/`)**
   - TMDB and OMDB for movies and TV shows
   - OpenLibrary for books (fully implemented)
   - MusicBrainz for music (fully implemented)
   - Graceful degradation: missing API keys never block the pipeline; entities are created with available metadata
   - `lazy_provider.go`: deferred provider initialization to avoid startup overhead
5. **Detector and analyzer (`internal/media/detector/` and `analyzer/`)**
   - File extension and MIME type detection
   - Content-based analysis for ambiguous files
   - Directory structure analysis (`directory_analyses` table) for batch categorization
6. **Entity API (`handlers/media_entity_handler.go`)**
   - CRUD endpoints under `/api/v1/entities`
   - Search with filters by type, title, year, parent
   - Hierarchy traversal: fetch all episodes of a season, all albums of an artist

#### Hands-On Exercise
Add a new storage root pointing to a local directory containing sample media files. Trigger a scan via the API and trace the logs to observe each pipeline stage. Query the entity API to see the created `media_items` and their hierarchical relationships. Examine how the title parser extracted metadata from filenames.

#### Key Takeaways
- Every scanned file passes through detection, analysis, title parsing, and aggregation before becoming a structured entity
- The aggregation service automatically builds hierarchies (show-season-episode) from directory structure and parsed metadata
- Metadata providers enrich entities with external data but never block the pipeline if unavailable
- Duplicate detection prevents the same title+type+year from creating multiple entities

---

### Module 4: Authentication and Authorization

**Duration**: 45 minutes
**Prerequisites**: Module 1, JWT concepts

#### Learning Objectives
- Trace the complete authentication flow from login request through token generation to middleware validation
- Explain the dual-token system (access + refresh) and session management
- Implement role-based access control using the middleware chain
- Understand rate limiting strategy: strict on auth endpoints, default on status endpoints

#### Topics Covered
1. **JWT token system (`internal/auth/service.go`)**
   - Dual JWT managers: access tokens (24h TTL) and refresh tokens (7d TTL)
   - Token claims: UserID, Username, RoleID, SessionID for per-device logout capability
   - `digital.vasic.auth` submodule providing `jwtmod.Manager` with configurable expiration
   - Token generation on login, refresh on expiry, invalidation on logout
2. **Authentication middleware (`internal/auth/middleware.go` and `middleware/auth.go`)**
   - `Authorization: Bearer <token>` header extraction and validation
   - Claims parsing and injection into Gin context for downstream handlers
   - Protected vs public route separation in route registration
3. **User management (`services/auth_service.go`)**
   - Registration with password hashing (bcrypt)
   - Login with credential validation and session creation
   - Session tracking with `SessionID` embedded in JWT claims
   - Per-device logout by invalidating specific sessions
4. **Role-based access control**
   - Roles stored in database: admin, user, viewer
   - `ProtectedRoute` component on frontend matching backend role checks
   - Admin-only endpoints (backup, user management, system configuration)
5. **Rate limiting on auth endpoints**
   - Strict rate limit (5 requests/minute) on `/api/v1/auth/login` and `/api/v1/auth/register`
   - Default rate limit (100 requests/minute) on `/api/v1/auth/status`, `/api/v1/auth/me`, `/api/v1/auth/permissions`
   - Redis-backed rate limiter (`middleware/redis_rate_limiter.go`) with fallback to in-memory
6. **Configuration precedence for secrets**
   - Environment variables override `.env` which overrides `config.json` which overrides defaults
   - `JWT_SECRET`, `ADMIN_USERNAME`, `ADMIN_PASSWORD` always sourced from env vars in production

#### Hands-On Exercise
Register a new user via the API, login to receive access and refresh tokens, and inspect the JWT payload using a decoder. Write a test that attempts to access a protected endpoint without a token and verifies the 401 response. Implement a custom role check for a new admin-only endpoint.

#### Key Takeaways
- Every token carries a SessionID enabling surgical per-device logout
- Auth rate limiting is split: strict on login/register to prevent brute force, relaxed on status/info endpoints
- The `digital.vasic.auth` submodule handles cryptographic operations; the application layer handles policy
- Secrets are never hardcoded -- env vars always take precedence over config files

---

### Module 5: WebSocket Real-Time Events

**Duration**: 40 minutes
**Prerequisites**: Module 1, Module 4, WebSocket protocol basics

#### Learning Objectives
- Describe the event bus architecture that connects backend state changes to WebSocket clients
- Implement a new real-time event type end-to-end (backend emit to frontend receive)
- Explain the connection lifecycle: upgrade, readPump, writePump, cleanup
- Apply concurrency safety patterns for connection tracking

#### Topics Covered
1. **Event bus (`internal/media/realtime/`)**
   - Publish/subscribe model: services emit events, the WebSocket handler broadcasts to connected clients
   - Event types: scan progress, media detection, entity updates, sync status, error notifications
   - `enhanced_watcher.go`: filesystem change detection feeding into the event bus
2. **WebSocket handler lifecycle**
   - HTTP upgrade to WebSocket via Gin handler
   - `readPump` goroutine: reads client messages, handles ping/pong for keepalive
   - `writePump` goroutine: delivers queued messages to the client
   - `sync.Once` for safe `Stop()` to prevent double-close panics
   - Tests must call `handler.Stop()` before `server.Close()` to unblock `readPump`
3. **Connection management**
   - `connCount` tracking protected by mutex for race safety
   - Broadcast to all connected clients with per-client send buffers
   - Graceful disconnection handling with cleanup of per-client resources
   - Production shutdown sequence: `wsHandler.Stop()` called before HTTP server shutdown in `main.go`
4. **Frontend integration (`@vasic-digital/websocket-client`)**
   - React hooks for WebSocket connection management
   - Automatic reconnection with exponential backoff
   - Type-safe event handlers matching backend event types
5. **Testing WebSocket connections**
   - `defer handler.Stop()` pattern in every test to prevent goroutine leaks
   - In-memory server setup for unit tests
   - Verifying broadcast delivery with concurrent client connections

#### Hands-On Exercise
Connect to the WebSocket endpoint from a browser console and observe real-time scan events while triggering a filesystem scan. Write a Go test that creates two WebSocket clients, emits an event through the event bus, and verifies both clients receive it. Implement a new custom event type for collection updates.

#### Key Takeaways
- The event bus decouples event producers (services) from consumers (WebSocket clients)
- Every WebSocket handler must use `sync.Once` for cleanup to prevent double-close panics
- Connection count is mutex-protected to prevent race conditions under concurrent connect/disconnect
- Test cleanup order matters: stop the handler before closing the server

---

### Module 6: File System Protocols

**Duration**: 50 minutes
**Prerequisites**: Module 1, network protocol basics (SMB, FTP, NFS, WebDAV)

#### Learning Objectives
- Implement a new protocol by conforming to the `UnifiedClient` interface
- Trace how the factory pattern selects the correct client implementation based on protocol URI
- Explain the SMB circuit breaker with offline cache and exponential-backoff retry
- Understand platform-specific builds for NFS (Linux, macOS, Windows stubs)

#### Topics Covered
1. **UnifiedClient interface (`filesystem/interface.go`)**
   - Methods: `Connect()`, `List()`, `Read()`, `Write()`, `Stat()`, `Close()`
   - Protocol-agnostic contract that all filesystem clients implement
   - Error wrapping with protocol-specific context
2. **Factory pattern (`filesystem/factory.go`)**
   - `NewClient(protocol, config)` dispatching to the correct implementation
   - Protocol detection from storage root URIs: `smb://`, `ftp://`, `nfs://`, `webdav://`, local paths
   - Configuration validation per protocol before client construction
3. **Protocol implementations**
   - `smb_client.go`: SMB/CIFS shares with circuit breaker, offline cache, exponential-backoff retry (`internal/smb/`)
   - `ftp_client.go`: FTP/FTPS connections with passive mode support
   - `nfs_client.go` / `nfs_client_darwin.go` / `nfs_client_windows.go`: platform-specific NFS with build tags
   - `webdav_client.go`: WebDAV over HTTP/HTTPS
   - `local_client.go`: local filesystem access
4. **SMB resilience (`internal/smb/`)**
   - Circuit breaker preventing repeated connection attempts to unreachable NAS devices
   - Offline cache serving last-known directory listings when the share is temporarily unavailable
   - Exponential-backoff retry with configurable max attempts and backoff multiplier
5. **Testing protocols**
   - `comprehensive_test.go`: integration tests covering all protocol paths
   - `factory_fuzz_test.go`: fuzz testing the factory with random URIs
   - Per-client unit tests with mock servers

#### Hands-On Exercise
Implement a mock filesystem client that conforms to `UnifiedClient` for an in-memory filesystem. Register it in the factory for a custom protocol scheme (e.g., `mem://`). Write a test that creates, reads, and lists files through the interface. Examine the SMB circuit breaker behavior by simulating connection failures.

#### Key Takeaways
- Adding a new protocol requires implementing `UnifiedClient` and registering in the factory -- no changes to service or handler layers
- The SMB client includes production-grade resilience: circuit breaker, offline cache, and exponential backoff
- NFS uses build tags for cross-platform compilation with platform-specific implementations
- The factory validates protocol-specific configuration before constructing clients

---

### Module 7: Challenge System

**Duration**: 45 minutes
**Prerequisites**: Module 1, Module 2, Module 4

#### Learning Objectives
- Define a custom challenge by embedding `BaseChallenge` and implementing `Execute()`
- Register challenges and understand the dependency resolution via topological sort
- Configure the challenge runner with timeouts, stale thresholds, and progress reporting
- Write challenge bank definitions in JSON format

#### Topics Covered
1. **Challenge interface and BaseChallenge (`digital.vasic.challenges`)**
   - Lifecycle: `Configure()` -> `Validate()` -> `Execute()` -> `Cleanup()`
   - `BaseChallenge` provides default implementations; concrete challenges override `Execute()`
   - Dependencies between challenges expressed as ID references
   - Result types: `StatusPassed`, `StatusFailed`, `StatusSkipped`, `StatusStuck`
2. **Registration (`catalog-api/challenges/register.go`)**
   - `RegisterAll()` function wiring all challenges into the registry
   - Four registration groups: `RegisterUserFlowAPIChallenges`, `RegisterUserFlowWebChallenges`, `RegisterUserFlowDesktopChallenges`, `RegisterUserFlowMobileChallenges`
   - 507+ challenges across original (CH-001 to CH-050), userflow (UF-*), and module verification (MOD-*)
3. **Runner configuration**
   - Runner timeout: 72 hours (hard upper bound for massive NAS scans)
   - Stale threshold: 5 minutes -- kills challenges reporting no progress
   - `challenge.NewConfig()` sets default timeout of 5 minutes; must be zeroed to use runner's timeout
   - Progress-based liveness: challenges embedding `BaseChallenge` auto-get `ProgressReporter`
   - `RunAll` is synchronous and blocking -- no other challenge can execute until it completes
4. **REST API (`handlers/challenge.go`)**
   - Endpoints under `/api/v1/challenges`: list, run single, run all, get results
   - `config.json` `write_timeout` must be 900 (not 30) for long-running `RunAll`
   - Context management: uses `context.Background()` for runner (not Gin request context which expires)
5. **Challenge banks**
   - Bank definitions in `challenges/config/` as JSON files
   - 18 categories covering API, web, desktop, mobile, cross-platform
   - Assertions with real expected values, not placeholder text
6. **User flow automation (`Challenges/pkg/userflow/`)**
   - Generic framework with adapter interfaces: Browser, Mobile, Desktop, API, Build, Process
   - CLI adapter implementations: Playwright, ADB, Tauri, HTTP, Gradle, npm, Go, Cargo
   - 13 challenge templates: EnvSetup, Build, UnitTest, Lint, APIHealth, APIFlow, BrowserFlow, MobileLaunch, MobileFlow, InstrumentedTest, DesktopLaunch, DesktopFlow, DesktopIPC

#### Hands-On Exercise
Create a new challenge `CH-CUSTOM` that verifies a specific API endpoint returns expected data. Embed `BaseChallenge`, implement `Execute()` with real assertions, and register it in `register.go`. Run it via the REST API and examine the result including assertions, metrics, and duration.

#### Key Takeaways
- Challenges are the backbone of quality assurance: every feature has a corresponding challenge
- The runner uses progress-based liveness detection, not just timeouts -- stale challenges are killed after 5 minutes of silence
- `RunAll` blocks the entire system -- plan accordingly for long-running campaigns
- Challenge banks define reusable test scenarios in JSON with real expected values

---

### Module 8: Monitoring and Metrics

**Duration**: 40 minutes
**Prerequisites**: Module 1, Prometheus/Grafana basics

#### Learning Objectives
- Configure Prometheus metric collection via the `/metrics` endpoint
- Build custom metrics using histograms, counters, and gauges for application-specific monitoring
- Set up health check endpoints with dependency status reporting
- Design Grafana dashboards for request latency, throughput, and error rates

#### Topics Covered
1. **Prometheus integration (`internal/metrics/`)**
   - `prometheus.go`: metric registration and HTTP handler for `/metrics` endpoint
   - `metrics.go`: custom application metrics -- request counters, latency histograms, active connections gauge
   - `histogram_buckets`: calibrated bucket boundaries for API response time distribution
   - `middleware.go`: middleware that instruments every request with duration, status code, and path labels
2. **Health checks (`internal/metrics/health.go`)**
   - `/api/v1/health` endpoint returning aggregated health status
   - Dependency checks: database connectivity, Redis availability, filesystem accessibility
   - Structured response with per-dependency status and latency
3. **Metrics middleware (`internal/metrics/middleware.go`)**
   - Automatic instrumentation of all HTTP requests
   - Labels: method, path, status code
   - Histogram observation for p50/p95/p99 latency tracking
4. **Grafana dashboards**
   - 16 pre-configured panels: request rate, error rate, latency percentiles, active connections, database pool utilization, cache hit rate, scan throughput, WebSocket connections, memory usage
   - Dashboard JSON definitions in `monitoring/` directory
   - Alert rules for latency spikes and error rate thresholds
5. **k6 load testing integration (`tests/k6/`)**
   - `load_test.js`: ramp to 50 users, verify p95 < 500ms
   - `stress_test.js`: ramp to 300 users to find breaking points
   - `soak_test.js`: 20 users for 30 minutes for memory leak detection
   - 11 total k6 scripts covering various scenarios

#### Hands-On Exercise
Start catalog-api and access the `/metrics` endpoint to see raw Prometheus output. Set up a local Prometheus instance scraping the endpoint. Create a Grafana dashboard with panels for request rate, p95 latency, and error percentage. Run a k6 load test and observe how the dashboard reflects the traffic pattern.

#### Key Takeaways
- Every HTTP request is automatically instrumented with latency and status code metrics
- Health checks verify all dependencies (database, cache, filesystems) and report individual status
- Histogram buckets are calibrated for API response times, not generic defaults
- k6 scripts cover load, stress, and soak testing scenarios with defined performance thresholds

---

### Module 9: Security Hardening

**Duration**: 45 minutes
**Prerequisites**: Module 4, HTTP security concepts

#### Learning Objectives
- Configure the complete security middleware chain: rate limiting, input validation, CORS, security headers
- Explain how HTTP/3 (QUIC) with Brotli compression is implemented using `quic-go/http3`
- Apply input validation middleware to prevent injection attacks
- Run the security scanning pipeline: govulncheck, npm audit, Semgrep, Gosec, Trivy

#### Topics Covered
1. **Rate limiting (`middleware/redis_rate_limiter.go`)**
   - Redis-backed sliding window rate limiter with in-memory fallback
   - Per-endpoint configuration: strict (5/min) for auth, default (100/min) for general endpoints
   - `digital.vasic.ratelimiter` submodule providing the core algorithm
   - Rate limit headers in responses: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`
2. **Input validation (`middleware/input_validation.go`)**
   - Request body size limits
   - SQL injection pattern detection in query parameters
   - XSS payload filtering in text inputs
   - Path traversal prevention in file-related endpoints
3. **Security headers (`middleware/security_headers.go`)**
   - `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`
   - `Strict-Transport-Security` for HTTPS enforcement
   - CORS configuration with allowed origins, methods, and headers
4. **HTTP/3 with QUIC**
   - `quic-go/http3` server with self-signed TLS certificates generated at startup
   - `andybalholm/brotli` compression for response bodies
   - Fallback chain: HTTP/3 -> HTTP/2 + gzip (never HTTP/1.1 in production)
5. **Concurrency limiter (`middleware/concurrency_limiter.go`)**
   - Semaphore-based request concurrency control
   - Prevents resource exhaustion under burst traffic
   - Configurable maximum concurrent requests per endpoint group
6. **Security scanning pipeline (`scripts/security-scan.sh`)**
   - `govulncheck`: Go dependency vulnerability scanning
   - `npm audit`: frontend dependency audit
   - Semgrep: static analysis via containerized scanner (`docker-compose.security.yml`)
   - Gosec: Go-specific security linter
   - Trivy: container image vulnerability scanning
   - Snyk: comprehensive dependency analysis

#### Hands-On Exercise
Enable all security middleware and test each layer: send a request exceeding the rate limit and observe the 429 response. Attempt a SQL injection payload and verify it is blocked. Run `scripts/security-scan.sh` and review the report. Inspect the HTTP/3 negotiation using `curl --http3`.

#### Key Takeaways
- Security is layered: rate limiting, input validation, security headers, and concurrency limiting work together
- HTTP/3 with Brotli is the default transport; fallback to HTTP/2 + gzip, never HTTP/1.1
- The security scanning pipeline covers dependencies (govulncheck, npm audit, Snyk), code (Semgrep, Gosec), and containers (Trivy)
- Rate limiting uses Redis when available and falls back to in-memory without configuration changes

---

### Module 10: Performance Optimization

**Duration**: 45 minutes
**Prerequisites**: Modules 1-9

#### Learning Objectives
- Apply the `LazyServiceRegistry` pattern for deferred service initialization with dependency ordering
- Use semaphore-based parallelism control for resource-constrained operations
- Configure the pooled HTTP client for connection reuse, timeouts, and retries
- Implement caching strategies with the `CacheService` and `digital.vasic.cache` submodule

#### Topics Covered
1. **Lazy service initialization (`internal/lifecycle/lazy_services.go`)**
   - `LazyServiceRegistry` deferring service construction until first use
   - Dependency ordering: services declare what they depend on, the registry resolves initialization order
   - Thread-safe lazy loading via `digital.vasic.lazy` generic module
   - Avoiding startup overhead for services that may never be used in a given request
2. **Semaphore-based concurrency (`internal/concurrency/semaphore.go`)**
   - Weighted semaphore controlling parallel scan operations
   - Preventing NAS overload by limiting concurrent filesystem connections
   - `digital.vasic.concurrency` submodule with configurable semaphore capacity
   - Integration with `GOMAXPROCS` resource limits: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
3. **Pooled HTTP client (`internal/httpclient/pool.go`)**
   - Connection pooling with reuse for external metadata provider calls (TMDB, OMDB, OpenLibrary, MusicBrainz)
   - Configurable timeouts, retry counts, and backoff strategies
   - Per-host connection limits preventing connection exhaustion
4. **Caching layer (`internal/services/cache_service.go`)**
   - In-memory cache with TTL-based expiration
   - Cleanup goroutine spawned in `NewCacheService()` -- tests must `defer service.Close()`
   - `sync.Once` for safe double-close prevention
   - Redis integration via `go-redis/v9` for shared cache across instances
   - `digital.vasic.cache` submodule providing cache abstraction
5. **Database connection pooling**
   - MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=3m defaults
   - Overridable via configuration for different deployment sizes
   - Connection health checking and stale connection eviction
6. **Response optimization**
   - Cache headers middleware (`middleware/cache_headers.go`) for static asset caching
   - Brotli compression reducing response payload sizes
   - Vite build chunks: vendor (react), router, ui, charts, utils for optimal browser caching

#### Hands-On Exercise
Profile a scan operation with and without semaphore limits to observe throughput differences. Enable Redis caching and measure cache hit rates for repeated entity queries. Configure the HTTP client pool for a metadata provider and observe connection reuse in the logs. Benchmark the `LazyServiceRegistry` startup time versus eager initialization.

#### Key Takeaways
- Lazy initialization avoids constructing services that are never used, reducing startup time
- Semaphores prevent resource exhaustion on both the API server and remote NAS devices
- The HTTP client pool reuses connections to external providers, reducing latency and TCP overhead
- Cache cleanup goroutines require explicit `Close()` calls -- leaking them causes goroutine leaks in tests
