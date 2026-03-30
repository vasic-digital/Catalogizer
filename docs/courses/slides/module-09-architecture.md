# Module 9: Architecture Deep Dive - Slide Deck Outline

**Total Slides**: 10
**Estimated Duration**: 50 minutes

---

## Slide 1: Title Slide (2 min)

**Title**: Architecture Deep Dive

- Module registry, concurrency patterns, cross-module design decisions
- Prerequisites: Module 6 completed
- By the end: explain how modules integrate and identify design patterns

---

## Slide 2: Module Registry Overview (5 min)

**Title**: 22 Go Modules Wired via Replace Directives

- catalog-api/go.mod maps module paths to local submodule directories
- 10 original modules: Auth, Cache, Config, Concurrency, Entities, EventBus, Filesystem, Assets, Containers, Challenges
- 12 additional: Database, Lazy, Media, Memory, Middleware, Observability, RateLimiter, Recovery, Security, Storage, Streaming, Watcher
- Each module: independent repo, tests, ARCHITECTURE.md, Upstreams
- Demo: read go.mod replace directives and trace one module

---

## Slide 3: Module Dependency Graph (5 min)

**Title**: How Modules Connect

- NewService constructors inject module dependencies
- Example: CatalogService depends on Database, Cache, Filesystem, Media
- No circular dependencies allowed between modules
- LazyServiceRegistry for deferred initialization with dependency ordering
- Exercise reference: Exercise 9.1 -- draw the dependency graph for a service

---

## Slide 4: Semaphore and Concurrency Control (5 min)

**Title**: Bounding Parallel Operations

- internal/concurrency/ provides semaphore-based concurrency control
- Limits parallel scan operations to prevent resource exhaustion
- Backup semaphore prevents concurrent database backups
- GOMAXPROCS=3 for tests to stay within host resource limits
- Demo: trace semaphore acquire/release during a scan

---

## Slide 5: Goroutine Lifecycle Management (5 min)

**Title**: Safe Startup and Shutdown

- CacheService spawns cleanup goroutine, tests must call defer service.Close()
- WebSocketHandler spawns cleanup goroutine, uses sync.Once for safe Stop()
- Production shutdown: wsHandler.Stop() then cacheService.Close() before HTTP server
- WaitGroup and Close() methods for coordinated shutdown
- Exercise reference: Exercise 9.2 -- identify all goroutine-spawning constructors

---

## Slide 6: Strategy Pattern -- Database Dialect (5 min)

**Title**: One Codebase, Two Databases

- DialectType enum: DialectSQLite and DialectPostgres
- RewritePlaceholders: ? -> $1, $2, ... for PostgreSQL
- RewriteInsertOrIgnore: INSERT OR IGNORE -> ON CONFLICT DO NOTHING
- BooleanLiterals: = 0/1 -> = FALSE/TRUE for known boolean columns
- database.DB wraps *sql.DB with shadowed Exec/Query/QueryRow

---

## Slide 7: Decorator Pattern -- Middleware (5 min)

**Title**: Composable Request Processing

- Auth middleware verifies JWT tokens and injects user context
- Rate limiter middleware: strict on login/register, default elsewhere
- Metrics middleware records request count, latency, status codes
- Brotli compression middleware for response encoding
- Middleware chain order matters: auth before rate limiting before metrics

---

## Slide 8: Observer and Facade Patterns (5 min)

**Title**: Event Propagation and Pipeline Orchestration

- Observer: EventBus publishes, WebSocket relays, Watcher triggers on filesystem changes
- Facade: media detection pipeline hides detector, analyzer, providers complexity
- Composite: filter chains combine multiple criteria
- Chain of Responsibility: handler chains for request processing
- Demo: trace an event from file change to WebSocket client notification

---

## Slide 9: Lazy Initialization and Recovery (5 min)

**Title**: Deferred Startup and Fault Tolerance

- LazyServiceRegistry defers service init until first use
- Dependency ordering ensures prerequisites are initialized first
- Recovery module: circuit breaker pattern for external services
- Memory module: heap tracking, goroutine monitoring, leak detection
- Exercise reference: Exercise 9.3 -- add a new service to LazyServiceRegistry

---

## Slide 10: Module Summary (3 min)

**Title**: What We Covered

- 22 Go modules wired via replace directives in go.mod
- Semaphore-based concurrency control and goroutine lifecycle
- Strategy (dialect), Decorator (middleware), Observer (eventbus), Facade (pipeline)
- Lazy initialization and circuit breaker recovery
- These patterns are the foundation for extending Catalogizer safely
- Next module: Advanced Features (challenges, user flows, entity pipeline)
