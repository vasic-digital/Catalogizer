# Comprehensive Decoupling Refactoring Design

**Date:** 2026-03-06
**Status:** Approved
**Scope:** Extract reusable Go modules from catalog-api into independent vasic-digital repositories

## 1. Goals

- Extract all reusable functionality from catalog-api into independent Go modules
- Maximize use of existing vasic-digital repositories (populate "dead" submodules with real code)
- Create new repos for functionality that has no existing module
- Each module: fully independent, tested, documented, with challenges
- Apply KISS, DRY, SOLID, YAGNI, Separation of Concerns, Composition over Inheritance
- Apply design patterns: Proxy, Facade, Factory, Abstract Factory, Observer, Mediator, Strategy, Chain of Responsibility, Decorator, State, Template Method, Singleton
- Each module gets: CLAUDE.md, AGENTS.md, README.md, docs/, website content, video course outlines
- Standard Go tests per module + thin challenge adapters in catalog-api

## 2. Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Approach | Top-down (highest-impact first) | Populates dead repos immediately, visible progress |
| Granularity | Fine-grained (one repo per concept) | Maximum reusability |
| Scope | Go modules first (Phase 1-5), TS/React deferred (Phase 6) | Go has most extraction candidates |
| Documentation | Markdown in docs/website/ and docs/courses/ | Deployable to any static site later |
| Challenges | Standard Go tests in modules + adapters in catalog-api | Keeps modules independent of Challenges framework |

## 3. Module Inventory

### Phase 1: High-Impact Existing Modules (8 modules)

| # | Module | Repo | Extract From | Key Packages to Add |
|---|--------|------|-------------|---------------------|
| 1 | Database | vasic-digital/Database | `database/` | `pkg/dialect`, `pkg/connection`, `pkg/helpers`, extend `pkg/migration` |
| 2 | Concurrency | vasic-digital/Concurrency | `internal/recovery/` | `pkg/retry`, `pkg/bulkhead` |
| 3 | Observability | vasic-digital/Observability | `internal/metrics/` | Extend `pkg/metrics`, `pkg/health`, add `pkg/middleware` |
| 4 | Security | vasic-digital/Security | `internal/middleware/` + `tests/security/` | `pkg/headers`, `pkg/scanning` |
| 5 | Middleware | vasic-digital/Middleware | `middleware/` + `internal/middleware/` | `pkg/auth`, `pkg/cache`, `pkg/ratelimit`, `pkg/validation`, `pkg/compression`, `pkg/security` |
| 6 | Media | vasic-digital/Media | `internal/media/` | Extend `pkg/detector`, `pkg/analyzer`, `pkg/provider`, add `pkg/models`, `pkg/manager` |
| 7 | Discovery | vasic-digital/Discovery | `internal/smb/` + `internal/services/smb*` | Extend `pkg/smb`, add `pkg/resilience` |
| 8 | Streaming | vasic-digital/Streaming | `internal/media/realtime/` | `pkg/realtime` |

### Phase 2: New Standalone Modules (3 modules)

| # | Module | New Repo | Extract From | Lines |
|---|--------|----------|-------------|-------|
| 9 | Lazy | vasic-digital/Lazy | `pkg/lazy/` | 240 |
| 10 | Memory | vasic-digital/Memory | `pkg/memory/` | 500 |
| 11 | Recovery | vasic-digital/Recovery | `internal/recovery/` | 1996 |

### Phase 3: Low-Priority Modules (4 modules)

| # | Module | Action |
|---|--------|--------|
| 12 | Storage | Add asset resolver package |
| 13 | Cache | Add service integration wrapper |
| 14 | Watcher | Wire into catalog-api |
| 15 | RateLimiter | Wire into catalog-api |

### Phase 4: Catalog-API Cleanup

- Delete replaced code (`pkg/lazy/`, `pkg/semaphore/`, `pkg/memory/`)
- Replace remaining inline code with module imports
- Update all internal imports
- Fix Upstreams for Containers + Entities
- Full test suite verification

### Phase 5: Documentation & Polish

- All modules: docs/architecture.md with diagrams
- All modules: docs/website/ content
- All modules: docs/courses/ outlines
- All modules: docs/sql-definitions.md where applicable
- Challenge adapters in catalog-api

### Phase 6: TS/React Modules (deferred)

## 4. Per-Module Architecture

### Standard directory structure

```
ModuleName/
  CLAUDE.md
  AGENTS.md
  README.md
  LICENSE
  go.mod / go.sum
  env.properties
  commit
  Upstreams/
    GitHub.sh
    GitLab.sh
  pkg/
    feature1/
      feature1.go
      feature1_test.go
    feature2/
      feature2.go
      feature2_test.go
  docs/
    architecture.md
    api-reference.md
    user-guide.md
    sql-definitions.md
    design-patterns.md
    website/
      index.md
      getting-started.md
      examples.md
      faq.md
    courses/
      outline.md
      lesson-01.md
      ...
```

### Design principles per module

| Principle | Application |
|-----------|------------|
| KISS | Each module does one thing. Minimal public API. |
| DRY | Code extracted once, imported via replace directives. |
| SOLID/SRP | Each package has single responsibility. |
| SOLID/OCP | Interfaces for extension. |
| SOLID/LSP | All implementations substitutable via interfaces. |
| SOLID/ISP | Small, focused interfaces. |
| SOLID/DIP | Depend on abstractions, not concrete types. |
| YAGNI | Only extract code that exists today. |
| Separation of Concerns | Infrastructure separate from domain. |
| Composition over Inheritance | Go interfaces + struct embedding. |

### Design patterns per module

| Module | Patterns |
|--------|----------|
| Database | Abstract Factory (dialect), Proxy (SQL rewriting wrapper), Strategy (SQLite vs PostgreSQL) |
| Middleware | Chain of Responsibility, Decorator, Strategy (memory vs Redis) |
| Media | Strategy (detection), Factory (providers), Observer (events), Template Method (pipeline) |
| Discovery | Facade (unified API), Circuit Breaker, Observer (state changes) |
| Observability | Observer (collectors), Facade (health aggregation), Singleton (registry) |
| Streaming | Observer (subscriptions), Mediator (event bus), Strategy (transports) |
| Security | Decorator (header middleware), Strategy (scanning backends) |
| Concurrency | Strategy (backoff), State (circuit breaker), Template Method (retry callbacks) |
| Lazy | Proxy (lazy value), Singleton (sync.Once) |
| Memory | Observer (leak alerts), Strategy (detection algorithms) |
| Recovery | Strategy (backoff), State (circuit breaker), Facade (resilience API) |

## 5. Integration Strategy

### go.mod after completion (23 replace directives)

```
// Existing (10)
replace digital.vasic.challenges => ../Challenges
replace digital.vasic.assets => ../Assets
replace digital.vasic.containers => ../Containers
replace digital.vasic.concurrency => ../Concurrency
replace digital.vasic.config => ../Config
replace digital.vasic.filesystem => ../Filesystem
replace digital.vasic.auth => ../Auth
replace digital.vasic.cache => ../Cache
replace digital.vasic.entities => ../Entities
replace digital.vasic.eventbus => ../EventBus

// New (13)
replace digital.vasic.database => ../Database
replace digital.vasic.middleware => ../Middleware
replace digital.vasic.media => ../Media
replace digital.vasic.discovery => ../Discovery
replace digital.vasic.observability => ../Observability
replace digital.vasic.streaming => ../Streaming
replace digital.vasic.security => ../Security
replace digital.vasic.ratelimiter => ../RateLimiter
replace digital.vasic.watcher => ../Watcher
replace digital.vasic.storage => ../Storage
replace digital.vasic.lazy => ../Lazy
replace digital.vasic.memory => ../Memory
replace digital.vasic.recovery => ../Recovery
```

### Catalog-API replacement map

| catalog-api package | Replaced by | Adapter |
|---|---|---|
| `database/dialect.go` + `connection.go` + `tx_helpers.go` + `migrations*.go` | `digital.vasic.database/pkg/*` | Thin facade re-exporting types + Catalogizer migration SQL |
| `middleware/auth.go`, `cache_headers.go`, `rate_limiter*.go`, `input_validation.go` | `digital.vasic.middleware/pkg/*` | Wiring layer composing middleware with Catalogizer config |
| `internal/media/detector/`, `analyzer/`, `providers/`, `models/`, `manager.go` | `digital.vasic.media/pkg/*` | Adapter mapping Catalogizer models to/from generic models |
| `internal/metrics/` | `digital.vasic.observability/pkg/*` | Register Catalogizer-specific metric names |
| `internal/smb/resilience.go` | `digital.vasic.discovery/pkg/resilience` | Direct import |
| `internal/media/realtime/` | `digital.vasic.streaming/pkg/realtime` | EventBus wiring adapter |
| `internal/recovery/` | `digital.vasic.recovery/pkg/*` | Direct import |
| `pkg/lazy/` | `digital.vasic.lazy` | Delete catalog-api copy |
| `pkg/semaphore/` | `digital.vasic.concurrency/pkg/semaphore` | Delete catalog-api copy |
| `pkg/memory/` | `digital.vasic.memory` | Delete catalog-api copy |

### Safety sequence per module

1. Populate module with generic code + tests
2. `cd ModuleName && go test ./...` passes
3. Add replace directive to catalog-api go.mod
4. Update catalog-api imports one package at a time
5. `GOMAXPROCS=3 go test ./... -p 2 -parallel 2` passes
6. Commit module + catalog-api together
7. Push module to upstreams, push catalog-api to 6 remotes

## 6. Repository Management

### New repos to create

```bash
# GitHub
gh repo create vasic-digital/Lazy --public --description "Generic reusable Go module: digital.vasic.lazy"
gh repo create vasic-digital/Memory --public --description "Generic reusable Go module: digital.vasic.memory"
gh repo create vasic-digital/Recovery --public --description "Generic reusable Go module: digital.vasic.recovery"

# GitLab
glab project create --group vasic-digital --name lazy --description "digital.vasic.lazy - Reusable Go module"
glab project create --group vasic-digital --name memory --description "digital.vasic.memory - Reusable Go module"
glab project create --group vasic-digital --name recovery --description "digital.vasic.recovery - Reusable Go module"
```

### Upstream setup per module

Each module: `Upstreams/GitHub.sh` + `Upstreams/GitLab.sh`, then `install_upstreams`.

### Push strategy

- GitHub + GitLab: All modules via Upstreams
- GitFlic + GitVerse: Only existing submodules that already have them
- Main repo: All 6 remotes

## 7. Task Breakdown

### Per-module tasks (6 per module)

1. Populate: Extract generic code, make framework-independent
2. Test: Module tests pass standalone
3. Wire: Add replace directive, update imports, create adapter
4. Verify: Full catalog-api test suite passes
5. Document: CLAUDE.md, AGENTS.md, README.md, docs/ tree
6. Push: Commit and push to all upstreams

### Phase totals

| Phase | Tasks | Description |
|-------|-------|-------------|
| Phase 1 | 48 | 8 high-impact modules x 6 tasks |
| Phase 2 | 18 | 3 new modules x 6 tasks |
| Phase 3 | 16 | 4 low-priority modules x 4 tasks |
| Phase 4 | 6 | Catalog-API cleanup |
| Phase 5 | 5 | Documentation and polish |
| **Total** | **93** | |

### Progress tracking

TaskCreate entries for each task. TaskList to check status at any time. Work is resumable at any task boundary.

## 8. Decoupling Techniques

### Removing Catalogizer-specific dependencies

| Dependency | Technique |
|-----------|-----------|
| `catalogizer/models` | Define generic model interfaces in module; adapter in catalog-api maps to/from |
| `catalogizer/config` | Accept generic config structs (e.g., `ConnectionConfig` instead of `config.DatabaseConfig`) |
| `catalogizer/database.DB` | Define `QueryExecutor` interface in module |
| `catalogizer/utils.SendErrorResponse` | Define `ErrorResponder` interface in module |
| `go.uber.org/zap` (hard dep) | Define `Logger` interface; accept via constructor injection |
| `github.com/gin-gonic/gin` | Use `net/http` interfaces; Gin adapter in catalog-api |
| `catalogizer/internal/metrics` | Define `MetricsReporter` interface |

### Interface pattern (applied everywhere)

```go
// In module: define minimal interface
type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}

// In catalog-api: adapt zap to interface
type zapAdapter struct{ logger *zap.SugaredLogger }
func (z *zapAdapter) Info(msg string, kv ...interface{}) { z.logger.Infow(msg, kv...) }
func (z *zapAdapter) Error(msg string, kv ...interface{}) { z.logger.Errorw(msg, kv...) }
```
