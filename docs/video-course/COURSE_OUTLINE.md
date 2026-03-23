# Catalogizer Video Course - Complete Outline

## Course Overview

**Title**: Mastering Catalogizer - Building a Multi-Platform Media Collection Manager

**Duration**: 12-14 hours (across 18 modules)

**Audience**: Intermediate to advanced Go and TypeScript developers

**Prerequisites**:
- Go 1.24+ experience
- React 18+ and TypeScript knowledge
- Basic understanding of media processing concepts
- Familiarity with containerization (Podman/Docker)

---

## Module 1: Introduction and Architecture (45 min)

### Video 1.1: Course Overview (10 min)
- What is Catalogizer
- Course objectives and learning outcomes
- Repository structure overview
- Development environment setup

### Video 1.2: System Architecture (20 min)
- High-level architecture diagram
- Microservices vs monolithic decisions
- Protocol abstraction layer (SMB, FTP, NFS, WebDAV, Local)
- Database dual-dialect pattern (SQLite/PostgreSQL)

### Video 1.3: Submodule Architecture (15 min)
- 29 independent git submodules
- Go module replace directives
- TypeScript package linking
- Multi-remote push strategy

---

## Module 2: Backend Development - Core Services (60 min)

### Video 2.1: Project Structure and Main Entry Point (15 min)
- Directory layout conventions
- `main.go` structure and initialization
- Dependency injection patterns
- Graceful shutdown handling

### Video 2.2: Database Layer (20 min)
- Dialect abstraction (`database.DB` wrapper)
- Migration system
- Repository pattern implementation
- Connection pooling and resource limits

### Video 2.3: Service Layer Design (15 min)
- Service interfaces and implementations
- Constructor injection with `NewService` pattern
- Error handling and wrapping
- Context propagation

### Video 2.4: Handler Layer with Gin (10 min)
- Gin router setup
- Handler struct pattern
- Request validation
- Response formatting

---

## Module 3: Authentication and Authorization (45 min)

### Video 3.1: JWT Authentication (20 min)
- Token generation and validation
- Session management
- Refresh token rotation
- Security best practices

### Video 3.2: Role-Based Access Control (15 min)
- Permission system design
- Role definitions
- Middleware implementation
- Resource ownership checks

### Video 3.3: Multi-Device Sessions (10 min)
- Active session tracking
- Session deactivation
- Logout all devices

---

## Module 4: Media Detection and Processing (75 min)

### Video 4.1: Universal Scanner Design (20 min)
- Protocol abstraction interface
- Concurrent scanning strategies
- Progress reporting via WebSocket
- Error recovery and retry logic

### Video 4.2: Media Detection Pipeline (25 min)
- File type detection
- Metadata extraction
- Content analysis
- Duplicate detection algorithms

### Video 4.3: Entity Aggregation (20 min)
- Media entity model (11 types)
- Hierarchy building (TV shows, music)
- Title parsing with regex
- External metadata integration (TMDB, IMDB)

### Video 4.4: Thumbnail and Preview Generation (10 min)
- Image processing
- Video frame extraction
- Caching strategies

---

## Module 5: Frontend Development (60 min)

### Video 5.1: React + TypeScript Setup (15 min)
- Vite configuration
- Path aliases
- Environment variables
- API proxy setup

### Video 5.2: State Management (15 min)
- React Query for server state
- Zustand for client state
- Context providers pattern
- WebSocket integration

### Video 5.3: Component Architecture (20 min)
- Atomic design principles
- Shared UI components (`@vasic-digital/ui-components`)
- Form handling with React Hook Form + Zod
- Animation with Framer Motion

### Video 5.4: Testing Frontend (10 min)
- Vitest unit tests
- Playwright E2E tests
- MSW for API mocking

---

## Module 6: Real-Time Features (30 min)

### Video 6.1: WebSocket Server (15 min)
- Connection management
- Message protocol design
- Broadcast patterns
- Reconnection handling

### Video 6.2: Live Updates (15 min)
- Scan progress streaming
- Real-time notifications
- Presence indicators
- Event bus integration

---

## Module 7: Protocol Implementations (60 min)

### Video 7.1: SMB/CIFS Client (15 min)
- Connection handling
- Authentication
- File operations
- Offline caching

### Video 7.2: WebDAV Client (15 min)
- HTTP methods mapping
- PROPFIND parsing
- Lock management
- Batch operations

### Video 7.3: FTP Client (10 min)
- Active vs passive mode
- Binary vs ASCII transfer
- Directory listing

### Video 7.4: NFS and Local Filesystems (10 min)
- Mount point detection
- Permission handling
- Symbolic link resolution

### Video 7.5: Protocol Factory Pattern (10 min)
- Unified client interface
- Protocol auto-detection
- Configuration-driven creation

---

## Module 8: HTTP/3 and Performance (45 min)

### Video 8.1: HTTP/3 with QUIC (20 min)
- quic-go integration
- TLS certificate handling
- ALPN negotiation
- Fallback to HTTP/2

### Video 8.2: Brotli Compression (10 min)
- Content negotiation
- Compression levels
- Static asset optimization

### Video 8.3: Caching Strategies (15 min)
- Redis integration
- Cache invalidation
- Stale-while-revalidate
- Memory-mapped files

---

## Module 9: Desktop and Mobile Apps (60 min)

### Video 9.1: Tauri Desktop App (25 min)
- Rust backend integration
- IPC commands and events
- Native file dialogs
- Auto-updates

### Video 9.2: Android Development (20 min)
- Kotlin + Jetpack Compose
- MVVM architecture
- Room database
- Retrofit API client

### Video 9.3: Android TV (15 min)
- Leanback UI framework
- Remote control navigation
- Media playback
- Live channels integration

---

## Module 10: Testing and Quality (45 min)

### Video 10.1: Unit Testing Patterns (15 min)
- Table-driven tests
- Mock interfaces vs sqlmock
- Test helpers and fixtures
- Coverage measurement

### Video 10.2: Integration Testing (15 min)
- Container-based tests
- Database fixtures
- API contract testing
- End-to-end flows

### Video 10.3: Challenge System (15 min)
- Challenge framework architecture
- Writing custom challenges
- User flow automation
- CI/CD integration

---

## Module 11: Security and Monitoring (30 min)

### Video 11.1: Security Best Practices (15 min)
- gosec and staticcheck
- SQL injection prevention
- XSS protection
- CSRF tokens

### Video 11.2: Observability (15 min)
- Prometheus metrics
- Structured logging with Zap
- Health check endpoints
- Error reporting service

---

## Module 12: Deployment and DevOps (45 min)

### Video 12.1: Container Strategy (20 min)
- Multi-stage builds
- Podman compose configuration
- Resource limits
- Network isolation

### Video 12.2: Installation Wizard (15 min)
- Guided setup flow
- Environment detection
- Configuration generation
- First-time user experience

### Video 12.3: Production Considerations (10 min)
- PostgreSQL migration
- Horizontal scaling
- Backup strategies
- Monitoring alerts

---

## Module 13: Search, Browse & Cloud Sync (45 min)

### Video 13.1: Search API (15 min)
- Full-text search with query parameters
- Advanced search with JSON body
- Search filters: extension, type, size, date range
- Media entity search and statistics
- Pagination and sorting

### Video 13.2: Browse API (10 min)
- Storage root listing
- Directory navigation and file info
- Directory size aggregation
- Duplicate detection within directory trees
- Entity browsing by media type

### Video 13.3: Cloud Sync (20 min)
- Sync endpoint CRUD (S3, GCS, WebDAV, local)
- Sync directions: push, pull, bidirectional
- Sync sessions and progress tracking
- Scheduling recurring syncs
- Sync statistics and cleanup

---

## Module 14: Challenge System Deep Dive (45 min)

### Video 14.1: Challenge Framework Architecture (15 min)
- Challenge interface lifecycle (Configure, Validate, Execute, Cleanup)
- BaseChallenge template method pattern
- Result struct: status, assertions, metrics, duration
- Assertion engine with 16 built-in evaluators

### Video 14.2: Writing Custom Challenges (10 min)
- Embedding BaseChallenge
- Implementing Execute with assertions
- Challenge bank JSON configuration
- RegisterAll and challenge registration

### Video 14.3: Running Challenges (10 min)
- REST API endpoints for challenge execution
- RunAll blocking constraint and write_timeout
- Progress-based liveness detection
- StatusStuck vs StatusTimedOut
- CLI runner with platform and report flags

### Video 14.4: User Flow Automation and Module Verification (10 min)
- pkg/userflow adapter-per-platform pattern
- 174 user flow challenges across 4 platforms
- Module verification challenges (MOD-001 to MOD-015)
- Container test stack with docker-compose.test.yml
- Report generation: Markdown, JSON, HTML

---

## Module 15: Concurrency Patterns in Go (45 min)

### Video 15.1: Goroutine Lifecycle Management (15 min)
- Context-based cancellation and propagation
- WaitGroup patterns for coordinating goroutines
- UniversalScanner worker pool design
- Deferred cleanup in main.go

### Video 15.2: Mutex Patterns and sync.Once (15 min)
- sync.RWMutex for read-heavy workloads (CacheService)
- The CacheService Close() idempotent shutdown pattern
- sync.Once for lazy initialization
- The Lazy module (digital.vasic.lazy) generic pattern
- Semaphore pattern for bounded parallelism (digital.vasic.concurrency)

### Video 15.3: Race Detection and Prevention (15 min)
- Running the Go race detector with resource limits
- Common race conditions: concurrent map access, closure captures, TOCTOU
- Testing concurrent code with table-driven tests
- Known flaky test: TestChaos_ConcurrentDatabaseAccess (SQLite WAL contention)

---

## Module 16: Security Scanning (40 min)

### Video 16.1: Security Scanning Overview (10 min)
- The six-tool security scanning stack
- Defense in depth: code + dependencies + containers
- Running all scans via scripts

### Video 16.2: govulncheck for Go (10 min)
- Call graph analysis vs simple dependency matching
- Running and interpreting govulncheck results
- Fixing vulnerabilities and verifying fixes

### Video 16.3: Semgrep Static Analysis (10 min)
- Pattern-based code scanning for security issues
- SQL injection prevention via the dialect abstraction layer
- Running Semgrep with OWASP and language-specific rulesets

### Video 16.4: npm audit, Snyk, and Trivy (10 min)
- npm audit for frontend dependency scanning
- Snyk for comprehensive dependency and container scanning
- Trivy for container image vulnerability detection
- SonarQube for code quality and security hotspots
- Integrating scans into the local release pipeline

---

## Module 17: Load Testing with k6 (45 min)

### Video 17.1: k6 Setup and Configuration (10 min)
- Installing k6 and project test structure
- Authentication helper for JWT-protected endpoints
- Threshold configuration for pass/fail criteria

### Video 17.2: Load Test Scenarios (15 min)
- Standard load test: gradual ramp to normal capacity
- Stress test: finding the breaking point
- Soak test: detecting memory leaks under sustained load
- Running tests within host resource limits (30-40%)

### Video 17.3: Interpreting Results (10 min)
- Key metrics: p95 latency, error rate, throughput
- Sending results to Prometheus/Grafana
- Comparing results across runs for regression detection

### Video 17.4: Grafana Dashboard for Load Testing (10 min)
- Correlating k6 results with application metrics
- Identifying capacity thresholds from dashboard data
- Setting up automated performance baselines

---

## Module 18: Monitoring and Observability (45 min)

### Video 18.1: Prometheus Metrics (15 min)
- Metrics architecture: GinMiddleware to Prometheus to Grafana
- Built-in HTTP metrics (requests, duration, size)
- Custom application metrics (scans, entities, WebSocket)
- Metric types: Counter, Gauge, Histogram

### Video 18.2: Runtime Metrics Collector (10 min)
- StartRuntimeCollector(15s) for goroutine and memory sampling
- Detecting goroutine leaks via go_goroutines trends
- GC pause monitoring and heap allocation patterns

### Video 18.3: Grafana Dashboards (10 min)
- Pre-built dashboard panels: request rate, latency, Go runtime, application
- PromQL queries for endpoint-specific monitoring
- Prometheus scrape configuration

### Video 18.4: Alerting and Log Aggregation (10 min)
- Prometheus alerting rules: error rate, latency, goroutine leak, service down
- Alertmanager routing to email and webhooks
- Structured logging with Zap (JSON output)
- Built-in log management API for collection, analysis, and sharing

---

## Bonus Content

### B1: Troubleshooting Guide (20 min)
- Common issues and solutions
- Debug logging techniques
- Performance profiling
- Memory leak detection

### B2: Contributing Guide (15 min)
- Code style conventions
- PR requirements
- Commit message format
- Documentation standards

### B3: API Reference Deep Dive (30 min)
- OpenAPI/Swagger documentation
- Authentication flows
- Rate limiting
- Versioning strategy

---

## Course Materials

- **Source Code**: Complete repository with tagged versions per module
- **Slides**: PDF and web-based presentations
- **Exercises**: Hands-on coding challenges
- **Cheat Sheets**: Quick reference guides
- **Sample Media**: Test files for scanning exercises

## Certification

- Module quizzes (passing score: 80%)
- Final project: Build a custom media scanner plugin
- Certificate of completion
