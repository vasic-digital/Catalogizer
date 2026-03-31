# Module 29: Module Architecture Deep Dive

## Video Script — 43 Submodules, Wiring Patterns, Lazy Init, Semaphores

### Duration: ~25 minutes

---

### Scene 1: Introduction (2 min)

"Catalogizer is built from 43 independent Git submodules — each with its own repository, tests, and documentation. This architecture enables code reuse across projects, independent versioning, and clean separation of concerns. In this module, we'll explore how they're organized, wired, and initialized."

---

### Scene 2: Submodule Categories (4 min)

**Go Modules (24):**
- Core: Assets, Auth, Cache, Config, Containers, Database, Entities, EventBus, Filesystem
- Infrastructure: Concurrency, Discovery, Lazy, Media, Memory, Middleware, Observability, RateLimiter, Recovery, Security, Storage, Streaming, Watcher
- Testing: Challenges, Build

**TypeScript Modules (9):**
- WebSocket-Client-TS, UI-Components-React, Media-Types-TS
- Catalogizer-API-Client-TS, Auth-Context-React
- Media-Browser-React, Media-Player-React, Collection-Manager-React, Dashboard-Analytics-React

**Specialized (10):**
- AI/QA: HelixQA, DocProcessor, LLMOrchestrator, LLMProvider, LLMsVerifier, VisionEngine
- Testing: ScreenDiff, ReplayBuffer, VisualRegression, TrainingCollector

---

### Scene 3: Go Module Wiring via replace Directives (5 min)

**File:** `catalog-api/go.mod`

```go
replace (
    digital.vasic.assets => ../Assets
    digital.vasic.auth => ../Auth
    digital.vasic.cache => ../Cache
    // ... 22 total replace directives
)
```

"Each module has its own `go.mod` with a vanity import path like `digital.vasic.database`. The `replace` directive in catalog-api points to the local submodule path."

**Module Registry:** `internal/modules/registry.go`
- Initializes all 12 infrastructure modules at startup
- Provides `StartBackgroundServices()` and `Stop()` lifecycle
- `GetModuleInfo()` returns status of all wired modules

---

### Scene 4: TypeScript Module Wiring via file:// (3 min)

**File:** `catalog-web/package.json`

```json
{
    "@vasic-digital/auth-context": "file:../Auth-Context-React",
    "@vasic-digital/websocket-client": "file:../WebSocket-Client-TS",
    "@vasic-digital/ui-components": "file:../UI-Components-React"
}
```

**Module Registry:** `src/lib/module-registry.ts`
- Re-exports all 6 packages with documentation
- Type-only exports for API client (avoids bundling server code)

---

### Scene 5: Lazy Initialization (4 min)

**Module:** `digital.vasic.lazy` (Lazy/)

"The Lazy module provides generic lazy loading primitives. Values are computed on first access, not at startup."

```go
import "digital.vasic.lazy"

// Create a lazy value
dbConn := lazy.NewValue(func() (*sql.DB, error) {
    return sql.Open("postgres", connStr)
})

// First call computes, subsequent calls return cached
db, err := dbConn.Get()
```

**Benefits:**
- Faster startup (services initialize on demand)
- Dependency ordering handled naturally
- Failed initialization can be retried

---

### Scene 6: Semaphore-Based Concurrency Control (4 min)

**Module:** `digital.vasic.concurrency` (Concurrency/)

"Semaphores limit concurrent operations to prevent resource exhaustion."

```go
import scancontrol "digital.vasic.concurrency"

sem := scancontrol.NewSemaphore(4) // max 4 concurrent

func processJob(ctx context.Context) {
    if err := sem.Acquire(ctx); err != nil {
        return // context cancelled
    }
    defer sem.Release()
    // do work
}
```

**Used in:**
- UniversalScanner: limits concurrent scan operations
- HTTP client pool: limits concurrent outbound requests
- Backup operations: ensures only 1 backup at a time

---

### Scene 7: Circuit Breaker Pattern (3 min)

**Module:** `digital.vasic.recovery` (Recovery/)

"Circuit breakers prevent cascading failures when external services are down."

```go
import "digital.vasic.recovery"

facade := recovery.NewResilienceFacade()
facade.AddCircuitBreaker("smb", recovery.CircuitBreakerConfig{
    MaxFailures: 5,
    Timeout:     30 * time.Second,
})

result, err := facade.Execute("smb", func() (interface{}, error) {
    return smbClient.ListFiles(path)
})
```

**States:** Closed (normal) → Open (failing, fast-fail) → Half-Open (testing recovery)

---

### Summary

- 43 submodules: Go (24), TypeScript (9), Specialized (10)
- Go wiring: `replace` directives + module registry
- TS wiring: `file://` references + module registry
- Lazy init: compute on first access, not startup
- Semaphores: bound concurrent operations
- Circuit breakers: prevent cascading failures
- Each module has: go.mod/package.json, tests, README, CLAUDE.md, ARCHITECTURE.md, Upstreams
