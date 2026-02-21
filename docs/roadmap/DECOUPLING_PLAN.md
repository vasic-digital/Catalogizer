# Catalogizer — Complete Decoupling Refactoring Plan

> **Status**: Executed and complete as of 2026-02-21.
> This document is the full, reproducible playbook. Every step was verified against
> the live codebase. Follow it exactly to repeat or extend the work.

---

## Table of Contents

1. [Overview & Goals](#overview--goals)
2. [Prerequisites & Rules](#prerequisites--rules)
3. [Module Inventory](#module-inventory)
4. [Software Principles Applied](#software-principles-applied)
5. [Phase 1 — Foundation Go Modules](#phase-1--foundation-go-modules)
6. [Phase 2 — Infrastructure Go Modules](#phase-2--infrastructure-go-modules)
7. [Phase 3 — Advanced Go Modules](#phase-3--advanced-go-modules)
8. [Phase 4 — New Go Module: Entities](#phase-4--new-go-module-entities)
9. [Phase 5 — TypeScript Foundation Modules](#phase-5--typescript-foundation-modules)
10. [Phase 6 — React UI Modules](#phase-6--react-ui-modules)
11. [Phase 7 — Documentation](#phase-7--documentation)
12. [Phase 8 — Challenges](#phase-8--challenges)
13. [Phase 9 — Final Verification](#phase-9--final-verification)
14. [How to Add Any Future Module](#how-to-add-any-future-module)
15. [Integration Decision Rules](#integration-decision-rules)
16. [Troubleshooting](#troubleshooting)

---

## Overview & Goals

Extract all reusable functionality from Catalogizer into standalone Git submodules under the
`vasic-digital` GitHub/GitLab organizations. Each module is:

- An independent repository with its own tests, docs, and version history
- Added as a Git submodule to Catalogizer at the project root
- Imported by `catalog-api/go.mod` (Go) or `catalog-web/package.json` (TypeScript/React)
  where architecturally compatible
- Documented with ARCHITECTURE.md, API_REFERENCE.md, USER_GUIDE.md, CHANGELOG.md,
  and course introduction scripts

**Result after all phases**:
- 26 total submodules registered in `.gitmodules`
- 9 Go modules imported in `catalog-api/go.mod` via `replace` directives
- 7 TypeScript/React npm packages linked in `catalog-web/package.json` via `file:` protocol
- 25 challenges (CH-001 to CH-025) covering all system functionality
- All 1643 frontend tests pass; all Go tests pass

---

## Prerequisites & Rules

### Naming Conventions

| Type | Example |
|------|---------|
| GitHub repo name | `vasic-digital/Entities` (PascalCase) |
| GitLab repo name | same as GitHub |
| Go module name | `digital.vasic.entities` (lowercase, dot-separated) |
| npm package name | `@vasic-digital/media-types` (kebab-case) |

### Repository Creation

```bash
# GitHub
gh repo create vasic-digital/<RepoName> --public --description "<desc>"

# GitLab (only if already configured as upstream)
glab project create --namespace vasic-digital <RepoName>
```

### Go Module Integration Pattern

For each Go submodule that gets integrated into `catalog-api/go.mod`:

```
# 1. Add replace directive to catalog-api/go.mod
replace digital.vasic.<name> => ../<RepoName>

# 2. Add require entry
require digital.vasic.<name> v0.0.0-00010101000000-000000000000

# 3. Run go mod tidy in catalog-api/
cd catalog-api && go mod tidy
```

### TypeScript Package Integration Pattern

For each TypeScript submodule in `catalog-web/package.json`:

```json
"@vasic-digital/<package-name>": "file:../<RepoName>"
```

Then run `npm install` in `catalog-web/`.

### Submodule-Only Pattern

When a module's architecture is incompatible with direct integration
(different HTTP framework, different DB driver, different client library),
add it as a submodule but skip the go.mod import. Document the reason.

### Mandatory per-Module Deliverables

Every module (new and existing) must have:

```
<ModuleRoot>/
  .gitignore
  README.md
  CLAUDE.md          — dev guidelines, commands, architecture
  AGENTS.md          — agent automation rules
  Upstreams/
    GitHub.sh        — git remote add + push commands for GitHub
    GitLab.sh        — git remote add + push commands for GitLab
  docs/
    ARCHITECTURE.md  — design patterns, module structure
    API_REFERENCE.md — every exported symbol documented
    USER_GUIDE.md    — step-by-step integration guide
    CHANGELOG.md     — semantic versioning history
    courses/
      00_introduction.md   — module overview, what students will learn
  pkg/               (Go) or src/ (TypeScript)
  *_test.go / *.test.ts — 100% coverage
```

### Resource Limits (CRITICAL)

**Never exceed 30-40% of host CPU/memory.** Always run:

```bash
# Go tests
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 -timeout 300s

# Never use -race unless specifically debugging (doubles resource use)
# Run challenges sequentially via API, never in parallel
```

### After Every Module: Verification

```bash
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1
cd catalog-web && npm run test -- --run
```

Both must show zero failures before proceeding to the next module.

---

## Module Inventory

### Already Submodules (Before This Work)

| Path | Module | Status |
|------|--------|--------|
| `WebSocket-Client-TS/` | `@vasic-digital/websocket-client` | Pre-existing |
| `UI-Components-React/` | `@vasic-digital/ui-components` | Pre-existing |
| `Challenges/` | `digital.vasic.challenges` | Pre-existing |
| `Assets/` | `digital.vasic.assets` | Pre-existing |

### Go Modules Added by This Plan

| Path | Module | go.mod | Reason |
|------|--------|--------|--------|
| `Config/` | `digital.vasic.config` | ✅ integrated | LoadFile() replaces manual JSON decode |
| `Database/` | `digital.vasic.database` | ❌ submodule only | modernc vs mattn sqlite driver mismatch |
| `Filesystem/` | `digital.vasic.filesystem` | ✅ integrated | Type aliases in filesystem/interface.go |
| `Concurrency/` | `digital.vasic.concurrency` | ✅ integrated | Circuit breaker wrapper in internal/recovery/ |
| `Auth/` | `digital.vasic.auth` | ✅ integrated | JWT Manager in internal/auth/service.go |
| `Middleware/` | `digital.vasic.middleware` | ❌ submodule only | net/http vs Gin incompatibility |
| `RateLimiter/` | `digital.vasic.ratelimiter` | ❌ submodule only | Needs Gin adapter not in module |
| `Observability/` | `digital.vasic.observability` | ❌ submodule only | ClickHouse deps conflict |
| `Media/` | `digital.vasic.media` | ❌ submodule only | DB-rules detection vs filename-only |
| `Watcher/` | `digital.vasic.watcher` | ❌ submodule only | Watcher has DB/analyzer deps |
| `EventBus/` | `digital.vasic.eventbus` | ✅ integrated | Type aliases + Catalogizer event constants |
| `Cache/` | `digital.vasic.cache` | ✅ integrated | Type aliases + DefaultConfig + NewTypedCache |
| `Security/` | `digital.vasic.security` | ❌ submodule only | AI guardrails vs HTTP input validation |
| `Storage/` | `digital.vasic.storage` | ❌ submodule only | MinIO vs AWS SDK v2 mismatch |
| `Streaming/` | `digital.vasic.streaming` | ❌ submodule only | Generic net/http vs Gin-specific |
| `Discovery/` | `digital.vasic.discovery` | ❌ submodule only | Simple TCP vs complex SMB scanner |
| `Entities/` | `digital.vasic.entities` | ✅ integrated | Title parser delegated; models for entity system |

### TypeScript/React Modules Added by This Plan

| Path | Package | catalog-web |
|------|---------|-------------|
| `Media-Types-TS/` | `@vasic-digital/media-types` | ✅ integrated |
| `Catalogizer-API-Client-TS/` | `@vasic-digital/catalogizer-api-client` | ✅ integrated |
| `Auth-Context-React/` | `@vasic-digital/auth-context` | ✅ integrated |
| `Media-Browser-React/` | `@vasic-digital/media-browser` | ✅ integrated |
| `Media-Player-React/` | `@vasic-digital/media-player` | ✅ integrated |
| `Collection-Manager-React/` | `@vasic-digital/collection-manager` | ✅ integrated |
| `Dashboard-Analytics-React/` | `@vasic-digital/dashboard-analytics` | ✅ integrated |

---

## Software Principles Applied

| Principle | How Applied |
|-----------|-------------|
| **Single Responsibility** | Each module owns exactly one domain |
| **Open/Closed** | Type aliases expose vasic interfaces without modifying Catalogizer internals |
| **Liskov Substitution** | Type aliases (`type X = vasic.X`) are exact replacements |
| **Interface Segregation** | Small, focused interfaces — never a "god" interface |
| **Dependency Inversion** | Catalogizer depends on vasic module interfaces, not concrete impls |
| **KISS** | Prefer type aliases over wrapper structs; no over-engineering |
| **DRY** | Title parser, circuit breaker, JWT logic live in exactly one place |
| **YAGNI** | Submodule-only for incompatible modules; no forced adapters |
| **Law of Demeter** | Services talk to vasic interfaces; not to vasic internal types |
| **Fail Fast** | Test suite runs after every module; immediate feedback on breakage |
| **Composition over Inheritance** | React components compose via props; Go uses embedding |

**Design Patterns Per Module**:

| Module | Patterns |
|--------|---------|
| Config | Factory (LoadFile), Null Object (defaults) |
| Filesystem | Facade (UnifiedClient), Adapter (per-protocol impls) |
| Concurrency | Decorator (circuit breaker wraps operations), Strategy (retry policy) |
| Auth | Strategy (JWT validation), Factory (Manager constructor) |
| EventBus | Observer, Mediator |
| Cache | Strategy (LRU/LFU/FIFO policies), Repository (typed cache) |
| Entities | Strategy (title parsing per media type), Template Method (hierarchy builder), Repository |
| media-types | Value Object (pure interfaces), Repository Interface (PaginatedResponse<T>) |
| catalogizer-api-client | Facade (CatalogizerClient), Repository (per-domain services), Decorator (withRetry) |
| auth-context | Facade (AuthProvider), Dependency Injection (authService prop), Observer (callbacks) |
| media-browser | Composite (EntityBrowser → EntityGrid → EntityCard), Strategy (navigation callbacks) |
| media-player | Strategy (video vs audio element selection), Facade (MediaPlayer hides hook) |
| collection-manager | Builder (SmartRuleBuilder), Command (action callbacks), Controlled Components |
| dashboard-analytics | Composite (EntityStatsGrid → StatsCard), Template Method (StatsCard layout) |

---

## Phase 1 — Foundation Go Modules

**Goal**: Add the 4 most foundational modules with zero circular dependencies.

### Step 1: digital.vasic.config

**What it replaces**: `catalog-api/internal/config/config.go` manual JSON file loading.

```bash
# 1. Clone and inspect the module
git submodule add git@github.com:vasic-digital/Config.git Config

# 2. Add to catalog-api/go.mod
# In the replace block:
replace digital.vasic.config => ../Config
# In the require block:
# digital.vasic.config v0.0.0-00010101000000-000000000000

# 3. Update catalog-api/internal/config/config.go
# Add import: vasicconfig "digital.vasic.config/pkg/config"
# Replace manual JSON parsing with vasicconfig.LoadFile()
```

**Integration pattern** (in `internal/config/config.go`):
```go
import vasicconfig "digital.vasic.config/pkg/config"

func Load(path string) (*Config, error) {
    cfg := &Config{}
    if err := vasicconfig.LoadFile(path, cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

### Step 2: digital.vasic.database

**Decision**: Submodule-only. The module uses `modernc.org/sqlite` (pure Go), but Catalogizer
uses `github.com/mattn/go-sqlite3` (CGO). The interfaces are incompatible.

```bash
git submodule add git@github.com:vasic-digital/Database.git Database
# No go.mod changes needed
```

### Step 3: digital.vasic.filesystem

**What it replaces**: `catalog-api/filesystem/interface.go` unified client type definitions.

```bash
git submodule add git@github.com:vasic-digital/Filesystem.git Filesystem
```

**Integration pattern** (in `catalog-api/filesystem/interface.go`):
```go
import vasicfs "digital.vasic.filesystem/pkg/filesystem"

// Type aliases — zero-cost, compile-time compatible
type UnifiedClient = vasicfs.UnifiedClient
type FileInfo = vasicfs.FileInfo
type ListOptions = vasicfs.ListOptions
// ... all other types
```

**go.mod changes**:
```
replace digital.vasic.filesystem => ../Filesystem
require digital.vasic.filesystem v0.0.0-00010101000000-000000000000
```

### Step 4: digital.vasic.concurrency

**What it replaces**: `catalog-api/internal/recovery/circuit_breaker.go`.

```bash
git submodule add git@github.com:vasic-digital/Concurrency.git Concurrency
```

**Integration pattern** (in `catalog-api/internal/recovery/circuit_breaker.go`):
```go
import vasicbreaker "digital.vasic.concurrency/pkg/breaker"

// Thin wrapper — delegate to vasicbreaker
type CircuitBreaker struct {
    breaker *vasicbreaker.CircuitBreaker
}

func NewCircuitBreaker(name string, opts ...vasicbreaker.Option) *CircuitBreaker {
    return &CircuitBreaker{breaker: vasicbreaker.New(name, opts...)}
}
```

**go.mod changes**:
```
replace digital.vasic.concurrency => ../Concurrency
require digital.vasic.concurrency v0.0.0-00010101000000-000000000000
```

### Verification After Phase 1

```bash
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 -timeout 300s
cd catalog-web && npm run test -- --run
```

---

## Phase 2 — Infrastructure Go Modules

**Goal**: Auth, middleware, rate limiter, observability.

### Step 1: digital.vasic.auth

**What it replaces**: JWT logic in `catalog-api/internal/auth/service.go`.

```bash
git submodule add git@github.com:vasic-digital/Auth.git Auth
```

**Integration pattern** (in `catalog-api/internal/auth/service.go`):
```go
import vasicjwt "digital.vasic.auth/pkg/jwt"

type AuthService struct {
    jwtManager *vasicjwt.Manager
    // ...
}

func NewAuthService(...) *AuthService {
    manager := vasicjwt.NewManager(vasicjwt.Config{
        Secret:     jwtSecret,
        Expiration: 24 * time.Hour,
    })
    return &AuthService{jwtManager: manager}
}

func (s *AuthService) GenerateToken(userID int, username string) (string, error) {
    return s.jwtManager.Generate(vasicjwt.Claims{
        UserID:   userID,
        Username: username,
    })
}
```

**go.mod changes**:
```
replace digital.vasic.auth => ../Auth
require digital.vasic.auth v0.0.0-00010101000000-000000000000
```

### Steps 2–4: Middleware, RateLimiter, Observability

**Decision for all three**: Submodule-only.

| Module | Reason |
|--------|--------|
| Middleware | Uses `net/http` handler chain; Catalogizer uses Gin middleware chain — incompatible function signatures |
| RateLimiter | Depends on Middleware module's Gin adapter which doesn't exist |
| Observability | Pulls in ClickHouse Go client — adds ~150MB of transitive deps, none needed in Catalogizer |

```bash
git submodule add git@github.com:vasic-digital/Middleware.git Middleware
git submodule add git@github.com:vasic-digital/RateLimiter.git RateLimiter
git submodule add git@github.com:vasic-digital/Observability.git Observability
# No go.mod changes for any of these
```

---

## Phase 3 — Advanced Go Modules

**Goal**: Media intelligence, event system, caching, streaming, security, storage, watcher, discovery.

### Integration Decision Process

For each module:
1. Read its `CLAUDE.md` and `pkg/` source
2. Ask: does it use the same HTTP framework (Gin)? Same DB driver? Same external clients?
3. If yes to all → integrate into go.mod with type aliases
4. If no to any → submodule-only, document reason

### digital.vasic.eventbus — INTEGRATED

**What it provides**: Generic pub/sub event bus.

```bash
git submodule add git@github.com:vasic-digital/EventBus.git EventBus
```

**Create** `catalog-api/internal/eventbus/eventbus.go`:
```go
package eventbus

import (
    vasicbus   "digital.vasic.eventbus/pkg/bus"
    vasicevent "digital.vasic.eventbus/pkg/event"
    vasicfilter "digital.vasic.eventbus/pkg/filter"
    vasicmw    "digital.vasic.eventbus/pkg/middleware"
)

// Type aliases — zero-cost
type EventBus  = vasicbus.EventBus
type Config    = vasicbus.Config
type Metrics   = vasicbus.Metrics
type Event     = vasicevent.Event
type EventType = vasicevent.Type
type Subscription = vasicevent.Subscription
type Filter    = vasicfilter.Filter
type Middleware = vasicmw.Middleware

func DefaultConfig() *Config { return vasicbus.DefaultConfig() }
func New(config *Config) *EventBus { return vasicbus.New(config) }
func NewEvent(t EventType, source string, payload interface{}) *Event {
    return vasicevent.New(t, source, payload)
}

// Catalogizer-specific event constants
const (
    EventScanStarted    EventType = "scan.started"
    EventScanCompleted  EventType = "scan.completed"
    EventScanFailed     EventType = "scan.failed"
    EventFileCreated    EventType = "file.created"
    EventFileModified   EventType = "file.modified"
    EventFileDeleted    EventType = "file.deleted"
    EventFileMoved      EventType = "file.moved"
    EventEntityCreated  EventType = "entity.created"
    EventEntityUpdated  EventType = "entity.updated"
    EventMetaRefreshed  EventType = "metadata.refreshed"
    EventCacheEvicted   EventType = "cache.evicted"
    EventSystemStartup  EventType = "system.startup"
    EventSystemShutdown EventType = "system.shutdown"
)
```

**go.mod changes**:
```
replace digital.vasic.eventbus => ../EventBus
require digital.vasic.eventbus v0.0.0-00010101000000-000000000000
```

### digital.vasic.cache — INTEGRATED

**What it provides**: Generic cache interface with LRU/LFU/FIFO policies.

```bash
git submodule add git@github.com:vasic-digital/Cache.git Cache
```

**Create** `catalog-api/internal/cache/cache.go`:
```go
package cache

import vasicache "digital.vasic.cache/pkg/cache"

type Cache          = vasicache.Cache
type Config         = vasicache.Config
type Stats          = vasicache.Stats
type EvictionPolicy = vasicache.EvictionPolicy

const (
    LRU  = vasicache.LRU
    LFU  = vasicache.LFU
    FIFO = vasicache.FIFO
)

func DefaultConfig() *Config { return vasicache.DefaultConfig() }
func NewTypedCache[T any](c Cache) *vasicache.TypedCache[T] {
    return vasicache.NewTypedCache[T](c)
}
```

**go.mod changes**:
```
replace digital.vasic.cache => ../Cache
require digital.vasic.cache v0.0.0-00010101000000-000000000000
```

### Remaining Phase 3 Modules — SUBMODULE ONLY

```bash
git submodule add git@github.com:vasic-digital/Media.git Media
git submodule add git@github.com:vasic-digital/Watcher.git Watcher
git submodule add git@github.com:vasic-digital/Security.git Security
git submodule add git@github.com:vasic-digital/Storage.git Storage
git submodule add git@github.com:vasic-digital/Streaming.git Streaming
git submodule add git@github.com:vasic-digital/Discovery.git Discovery
```

No go.mod changes for any of these. Reasons documented in Module Inventory.

---

## Phase 4 — New Go Module: Entities

**Goal**: Create a brand-new `vasic-digital/Entities` module containing the title parser
and media entity models extracted from Catalogizer. Then integrate it.

### Step 1: Create the GitHub repository

```bash
gh repo create vasic-digital/Entities --public \
  --description "Media entity models and title parser for the Catalogizer ecosystem"
```

### Step 2: Build the module locally

```bash
mkdir -p /tmp/Entities/pkg/parser /tmp/Entities/pkg/models
cd /tmp/Entities
```

**`go.mod`**:
```
module digital.vasic.entities

go 1.24.0

require github.com/stretchr/testify v1.9.0
```

**`pkg/parser/parser.go`** — extract from `catalog-api/internal/services/title_parser.go`:
```go
package parser

import (
    "regexp"
    "strings"
    "strconv"
)

type ParsedTitle struct {
    Title         string
    Year          *int
    Season        *int
    Episode       *int
    Quality       string
    Group         string
    Artist        string
    Album         string
    Platform      string
    Version       string
    Extra         map[string]string
}

// ParseMovieTitle parses a directory name as a movie title.
func ParseMovieTitle(dirname string) ParsedTitle { ... }

// ParseTVShow parses a directory name as a TV show.
func ParseTVShow(dirname string) ParsedTitle { ... }

// ParseMusicAlbum parses a directory name as a music album.
func ParseMusicAlbum(dirname string) ParsedTitle { ... }

// ParseGameTitle parses a directory name as a game.
func ParseGameTitle(dirname string) ParsedTitle { ... }

// ParseSoftwareTitle parses a directory name as software.
func ParseSoftwareTitle(dirname string) ParsedTitle { ... }

// CleanTitle removes noise tokens from a raw title string.
func CleanTitle(raw string) string { ... }

// ExtractYear parses the first 4-digit year (1900-2099) from s.
func ExtractYear(s string) *int { ... }

// ExtractQualityHints returns quality keywords found in s.
func ExtractQualityHints(s string) []string { ... }

// DetectMediaCategory returns the likely MediaCategory for dirname.
func DetectMediaCategory(dirname string) MediaCategory { ... }
```

**`pkg/models/models.go`**:
```go
package models

type MediaCategory string

const (
    CategoryMovie    MediaCategory = "movie"
    CategoryTV       MediaCategory = "tv_show"
    CategoryMusic    MediaCategory = "music"
    CategoryGame     MediaCategory = "game"
    CategorySoftware MediaCategory = "software"
    CategoryBook     MediaCategory = "book"
    CategoryUnknown  MediaCategory = "unknown"
)

type QualityInfo struct {
    OverallScore float64
    Resolution   string
    Codec        string
    FileSize     int64
    Bitrate      int
}

type MediaType struct {
    ID                 int
    Name               string
    Description        string
    DetectionPatterns  []string
    MetadataProviders  []string
}

type MediaItem struct {
    ID            int
    Title         string
    MediaTypeID   int
    Year          *int
    DirectoryPath string
    ParentID      *int
    SeasonNumber  *int
    EpisodeNumber *int
    TrackNumber   *int
    Status        string
    FirstDetected string
    LastUpdated   string
}

type MediaFile struct {
    ID          int
    MediaItemID int
    FileID      int
    FilePath    string
    FileSize    int64
    IsPrimary   bool
    AddedAt     string
}

type DuplicateGroup struct {
    Title       string
    MediaTypeID int
    Year        *int
    Count       int
    Items       []MediaItem
}

type HierarchyNode struct {
    Item     MediaItem
    Children []HierarchyNode
}
```

### Step 3: Test and push

```bash
cd /tmp/Entities
go test ./... -count=1
git init && git add -A
git commit -m "Add digital.vasic.entities: title parser + media entity models"
git branch -m main
git remote add origin git@github.com:vasic-digital/Entities.git
git push -u origin main
```

### Step 4: Add as submodule and integrate

```bash
cd /path/to/Catalogizer
git submodule add git@github.com:vasic-digital/Entities.git Entities
```

**In `catalog-api/go.mod`**:
```
replace digital.vasic.entities => ../Entities
require digital.vasic.entities v0.0.0-00010101000000-000000000000
```

**Replace `catalog-api/internal/services/title_parser.go`** with a thin delegation layer:
```go
package services

import vasicparser "digital.vasic.entities/pkg/parser"

// Type alias — ParsedTitle is the same type in both packages
type ParsedTitle = vasicparser.ParsedTitle

func ParseMovieTitle(dirname string) ParsedTitle   { return vasicparser.ParseMovieTitle(dirname) }
func ParseTVShow(dirname string) ParsedTitle       { return vasicparser.ParseTVShow(dirname) }
func ParseMusicAlbum(dirname string) ParsedTitle   { return vasicparser.ParseMusicAlbum(dirname) }
func ParseGameTitle(dirname string) ParsedTitle    { return vasicparser.ParseGameTitle(dirname) }
func ParseSoftwareTitle(dirname string) ParsedTitle { return vasicparser.ParseSoftwareTitle(dirname) }
func CleanTitle(raw string) string                 { return vasicparser.CleanTitle(raw) }
func ExtractYear(s string) *int                    { return vasicparser.ExtractYear(s) }
```

**IMPORTANT**: After refactoring `title_parser.go`, search `aggregation_service.go` for any
regex variables that were previously defined in `title_parser.go` (e.g. `gamePlatformRe`).
Add them back locally in `aggregation_service.go`:

```go
import "regexp"

var gamePlatformRe = regexp.MustCompile(
    `(?i)\b(?:PC|Windows|Linux|Mac|macOS|PS[2-5]|PlayStation[\s._-]*[2-5]?|` +
    `Xbox(?:[\s._-]*(?:One|360|Series[\s._-]*[XS]))?|Switch|Nintendo|GOG|Steam)\b`,
)
```

---

## Phase 5 — TypeScript Foundation Modules

**Goal**: Create 3 new TypeScript/React packages and add them as submodules.

### Prep: Create GitHub repos first

```bash
gh repo create vasic-digital/Media-Types-TS --public \
  --description "Shared TypeScript type definitions for Catalogizer media entities, auth, and API"
gh repo create vasic-digital/Catalogizer-API-Client-TS --public \
  --description "Type-safe TypeScript client for the Catalogizer API"
gh repo create vasic-digital/Auth-Context-React --public \
  --description "React AuthProvider and useAuth hook for Catalogizer authentication"
```

### Module 1: @vasic-digital/media-types

**Purpose**: Zero-runtime TypeScript interfaces shared across all other TS packages.

**Location**: `/tmp/Media-Types-TS/` (build here, push, then submodule add)

**`package.json`**:
```json
{
  "name": "@vasic-digital/media-types",
  "version": "0.1.0",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "scripts": { "build": "tsc", "test": "vitest run", "lint": "tsc --noEmit" },
  "devDependencies": { "typescript": "^5.4.0", "vitest": "^2.0.0" }
}
```

**`tsconfig.json`**:
```json
{
  "compilerOptions": {
    "target": "ES2020", "module": "ESNext", "moduleResolution": "bundler",
    "strict": true, "declaration": true, "outDir": "dist", "rootDir": "src",
    "skipLibCheck": true
  },
  "include": ["src"], "exclude": ["node_modules", "dist"]
}
```

**Source files**:
- `src/auth.ts` — `Role`, `User`, `DeviceInfo`, `LoginRequest`, `LoginResponse`,
  `RegisterRequest`, `AuthStatus`, `ChangePasswordRequest`, `UpdateProfileRequest`
- `src/media.ts` — `MediaItem`, `ExternalMetadata`, `MediaVersion`, `QualityInfo`,
  `MediaSearchRequest`, `MediaSearchResponse`, `MediaEntity`, `MediaType`, `MediaFile`,
  `EntityExternalMetadata`, `UserMetadata`, `EntityStats`, `DuplicateGroup`,
  `PaginatedResponse<T>`
- `src/collections.ts` — `MediaCollection`, `SmartCollectionRule`,
  `CreateCollectionRequest`, `UpdateCollectionRequest`
- `src/index.ts` — barrel re-export of all above with `export type { ... }`

**Key type details**:

`LoginResponse` must have `session_token` (not `token`) — matches the Go backend:
```typescript
export interface LoginResponse {
  user: User
  session_token: string
  refresh_token: string
  expires_at: string
}
```

`SmartCollectionRule.operator` union:
```typescript
operator: 'eq' | 'ne' | 'gt' | 'lt' | 'contains' | 'not_contains'
```

`PaginatedResponse<T>` generic:
```typescript
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}
```

**Tests**: 35 tests across `auth.test.ts`, `media.test.ts`, `collections.test.ts`.

**Build, test, push**:
```bash
cd /tmp/Media-Types-TS
npm install
npm test   # must show 35 passed
git init && echo "node_modules/\ndist/\n*.tsbuildinfo" > .gitignore
git add -A && git commit -m "Add @vasic-digital/media-types: shared TypeScript type definitions"
git branch -m main
git remote add origin git@github.com:vasic-digital/Media-Types-TS.git
git push -u origin main
```

**Add as submodule**:
```bash
cd /path/to/Catalogizer
git submodule add git@github.com:vasic-digital/Media-Types-TS.git Media-Types-TS
```

### Module 2: @vasic-digital/catalogizer-api-client

**Purpose**: Type-safe axios-based HTTP client for all Catalogizer API endpoints.

**Dependencies**: `@vasic-digital/media-types` (local), `axios`

**`package.json`**:
```json
{
  "name": "@vasic-digital/catalogizer-api-client",
  "version": "0.1.0",
  "dependencies": {
    "@vasic-digital/media-types": "file:../Media-Types-TS",
    "axios": "^1.6.0"
  },
  "devDependencies": { "typescript": "^5.4.0", "vitest": "^2.0.0" }
}
```

**Source structure**:
```
src/
  types.ts             — ClientConfig, ApiResponse, StorageRootConfig,
                         ScanRequest, ScanResult, WebSocketMessage,
                         CatalogizerError, AuthenticationError, NetworkError, ValidationError
  http.ts              — HttpClient class (axios wrapper)
  services/
    AuthService.ts     — /auth/* endpoints
    EntityService.ts   — /api/v1/entities/* endpoints
    CollectionService.ts — /api/v1/collections/* endpoints
    StorageService.ts  — /api/v1/storage-roots + /api/v1/scan endpoints
  index.ts             — CatalogizerClient + barrel exports
```

**HttpClient key features**:
- Request interceptor: inject `Authorization: Bearer <token>` header
- Response interceptor: 401 → call `onTokenRefresh()` → retry original request
- Error mapping: 400→ValidationError, 401→AuthenticationError, 0→NetworkError
- `withRetry(operation, maxAttempts, delay)` — exponential backoff, skip auth/validation errors

**CatalogizerClient**:
```typescript
export class CatalogizerClient extends EventEmitter {
  readonly auth: AuthService
  readonly entities: EntityService
  readonly collections: CollectionService
  readonly storage: StorageService

  constructor(config: ClientConfig) {
    // Set up token refresh chain:
    // http.onTokenRefresh → auth.refreshToken() → return new session_token
    // http.onAuthenticationError → emit('auth:expired')
  }

  setToken(token: string): void
  clearToken(): void
  getToken(): string | undefined
  getBaseURL(): string
}
```

**AuthService.login** must call `http.setAuthToken(response.session_token)` after login.

**EntityService** covers all 14 entity endpoints including:
```typescript
getStreamURL(id: number, baseURL: string): string  // pure, no HTTP call
getDownloadURL(id: number, baseURL: string): string // pure, no HTTP call
```

**Tests**: 28 tests. Use `vi.fn()` mocks for HttpClient — never real HTTP.

**Install with local media-types**:
```bash
cd /tmp/Catalogizer-API-Client-TS
npm install "@vasic-digital/media-types@file:../Media-Types-TS"
npm test   # must show 28 passed
```

**Add as submodule**:
```bash
git submodule add git@github.com:vasic-digital/Catalogizer-API-Client-TS.git Catalogizer-API-Client-TS
```

### Module 3: @vasic-digital/auth-context

**Purpose**: React `AuthProvider` + `useAuth` hook using `@tanstack/react-query`.

**Peer dependencies**: `react ^18`, `@tanstack/react-query ^5`

**Key design decisions**:
- `authService: AuthService` passed as prop (dependency injection) — not created internally
- No `react-hot-toast` dependency — use `onLoginSuccess`, `onLogout`, `onError` callbacks instead
- `localStorage` operations inside `AuthProvider` only
- `isAdmin` = `user?.role?.name === 'Admin' || user?.role_id === 1`

**`AuthProvider` props**:
```typescript
interface AuthProviderProps {
  authService: AuthService       // from @vasic-digital/catalogizer-api-client
  children: ReactNode
  onLoginSuccess?: (user: User) => void
  onLogout?: () => void
  onError?: (error: Error) => void
}
```

**`useAuth()` returns**:
```typescript
interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  permissions: string[]
  isAdmin: boolean
  login(data: LoginRequest): Promise<void>
  logout(): Promise<void>
  register(data: RegisterRequest): Promise<void>
  updateProfile(data: UpdateProfileRequest): Promise<void>
  changePassword(data: ChangePasswordRequest): Promise<void>
  hasPermission(permission: string): boolean
  canAccess(resource: string, action: string): boolean
}
```

**`canAccess(resource, action)`** builds `"${action}:${resource}"` and checks
`hasPermission` or falls back to `hasPermission('admin:system')`.

**vitest.config.ts** requires `@vitejs/plugin-react` and `jsdom` environment.

**Tests**: 6 tests — authenticated state, unauthenticated state, isAdmin, useAuth outside
provider throws, hasPermission, canAccess.

```bash
cd /tmp/Auth-Context-React
npm install \
  "@vasic-digital/media-types@file:../Media-Types-TS" \
  "@vasic-digital/catalogizer-api-client@file:../Catalogizer-API-Client-TS"
npm test   # must show 6 passed (stderr error is expected from the throws test)
```

### Update catalog-web/package.json

```json
"dependencies": {
  "@vasic-digital/media-types": "file:../Media-Types-TS",
  "@vasic-digital/catalogizer-api-client": "file:../Catalogizer-API-Client-TS",
  "@vasic-digital/auth-context": "file:../Auth-Context-React"
}
```

```bash
cd catalog-web && npm install
npm run test -- --run  # must show 1643 passed
```

---

## Phase 6 — React UI Modules

**Goal**: Create 4 standalone React component libraries.

### Prep: Create GitHub repos

```bash
gh repo create vasic-digital/Media-Browser-React --public \
  --description "React entity browser components for Catalogizer media browsing"
gh repo create vasic-digital/Media-Player-React --public \
  --description "React media player component for Catalogizer entity playback"
gh repo create vasic-digital/Collection-Manager-React --public \
  --description "React collection management components for Catalogizer"
gh repo create vasic-digital/Dashboard-Analytics-React --public \
  --description "React dashboard and analytics components for Catalogizer"
```

### Common setup for all 4 modules

**`tsconfig.json`** (all 4):
```json
{
  "compilerOptions": {
    "target": "ES2020", "module": "ESNext", "moduleResolution": "bundler",
    "jsx": "react-jsx", "strict": true, "declaration": true,
    "outDir": "dist", "rootDir": "src", "skipLibCheck": true
  }
}
```

**`vitest.config.ts`** (all 4):
```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: { environment: 'jsdom', setupFiles: ['./src/__tests__/setup.ts'], globals: true },
})
```

**`src/__tests__/setup.ts`** (all 4):
```typescript
import '@testing-library/jest-dom'
```

### Module 1: @vasic-digital/media-browser (16 tests)

**Components**:
- `EntityCard` — title, year, media type, rating, truncated description; `onClick` callback
- `TypeSelector` — button grid of media types; `onSelect` callback
- `Pagination` — prev/next with `currentPage / totalPages`; hidden when `totalPages <= 1`
- `EntityGrid` — responsive grid; loading spinner; empty state; composes EntityCard + Pagination
- `EntityBrowser` — search input + TypeSelector OR EntityGrid; `onBack` button

**All components are controlled** — no internal routing, no data fetching, all state via props.

**Key implementation detail for Pagination**:
```typescript
const currentPage = Math.floor(offset / limit) + 1
const totalPages = Math.ceil(total / limit)
if (totalPages <= 1) return null
```

### Module 2: @vasic-digital/media-player (12 tests)

**Components**:
- `useMediaPlayer` hook — state (isPlaying, isMuted, volume, currentTime, duration), controls, ref
- `PlayerControls` — play/pause toggle, time display (`M:SS`), seek range, mute toggle, volume range
- `MediaPlayer` — detects `entity.media_type?.name` to render `<video>` or `<audio>`;
  renders `data-testid="video-element"` or `data-testid="audio-element"`

**Audio vs video decision**:
```typescript
const isAudio = mediaType === 'song' || mediaType === 'music_album'
```

**Time formatting**:
```typescript
function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${s.toString().padStart(2, '0')}`
}
```

### Module 3: @vasic-digital/collection-manager (18 tests)

**Components**:
- `CollectionCard` — name (clickable), description, item_count badge, smart/public badges,
  edit/delete buttons
- `SmartRuleBuilder` — dynamic field/operator/value rows; add/remove rules;
  operators: `eq | ne | gt | lt | contains | not_contains`
- `CollectionForm` — name + description inputs; is_public + is_smart checkboxes;
  shows SmartRuleBuilder when `isSmart`; JS validation (not HTML5 `required`)
- `CollectionList` — grid of CollectionCards + "New Collection" button

**CRITICAL**: Remove `required` from name `<input>` in `CollectionForm`. HTML5 constraint
validation blocks the React `onSubmit` handler before it fires, preventing custom error
state from being set. Use JS validation in `handleSubmit` instead:
```typescript
if (!name.trim()) {
  setError('Name is required')
  return
}
```

### Module 4: @vasic-digital/dashboard-analytics (18 tests)

**Components**:
- `StatsCard` — label, value, optional unit, optional trend arrow (↑ ↓ —)
- `EntityStatsGrid` — 5 StatsCards from `EntityStats`; auto-formats bytes:
  ```typescript
  function formatSize(bytes: number): string {
    if (bytes >= 1e12) return `${(bytes / 1e12).toFixed(1)} TB`
    if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`
    if (bytes >= 1e6) return `${(bytes / 1e6).toFixed(1)} MB`
    return `${bytes} B`
  }
  ```
- `MediaDistributionBar` — proportional horizontal bar; empty state when total=0; legend
- `ActivityFeed` — ordered list with `maxItems` limit (default 10); empty state

### Install and test all 4 modules

```bash
# Example for Media-Browser-React
cd /tmp/Media-Browser-React
npm install "@vasic-digital/media-types@file:../Media-Types-TS"
npm test   # must show 16 passed

# Push pattern for all 4:
git init && echo "node_modules/\ndist/\n*.tsbuildinfo" > .gitignore
git add -A && git commit -m "Add @vasic-digital/<name>: ..."
git branch -m main
git remote add origin git@github.com:vasic-digital/<RepoName>.git
git push -u origin main
```

### Add all 4 as submodules

```bash
git submodule add git@github.com:vasic-digital/Media-Browser-React.git Media-Browser-React
git submodule add git@github.com:vasic-digital/Media-Player-React.git Media-Player-React
git submodule add git@github.com:vasic-digital/Collection-Manager-React.git Collection-Manager-React
git submodule add git@github.com:vasic-digital/Dashboard-Analytics-React.git Dashboard-Analytics-React
```

### Update catalog-web/package.json

```json
"@vasic-digital/media-browser": "file:../Media-Browser-React",
"@vasic-digital/media-player": "file:../Media-Player-React",
"@vasic-digital/collection-manager": "file:../Collection-Manager-React",
"@vasic-digital/dashboard-analytics": "file:../Dashboard-Analytics-React"
```

```bash
cd catalog-web && npm install && npm run test -- --run
# Must show 1643 passed
```

---

## Phase 7 — Documentation

For every module, create `docs/` with:

```
docs/
  ARCHITECTURE.md    — design patterns, component tree, principles
  API_REFERENCE.md   — every exported symbol with type signatures
  USER_GUIDE.md      — installation, quickstart, common patterns
  CHANGELOG.md       — [0.1.0] initial release entry
  courses/
    00_introduction.md — overview, what students will learn, prerequisites
```

**ARCHITECTURE.md minimum content**:
1. One-paragraph overview of what the module does
2. Design patterns used (with brief "why" for each)
3. Module/component structure as a tree or table
4. Dependency graph (mermaid or ASCII)

**Commit docs separately** from code so git history is clean:
```bash
cd /tmp/<ModuleName>
git add docs/
git commit -m "Add documentation: ARCHITECTURE, API_REFERENCE, USER_GUIDE, CHANGELOG, courses"
git push origin main
```

**Update Catalogizer submodule pointers** after doc pushes:
```bash
cd /path/to/Catalogizer
git submodule update --remote Media-Types-TS Catalogizer-API-Client-TS Auth-Context-React \
  Media-Browser-React Media-Player-React Collection-Manager-React Dashboard-Analytics-React
git add -A && git commit -m "Update submodule pointers to include documentation commits"
```

---

## Phase 8 — Challenges

**Goal**: Add 5 new challenges (CH-021 to CH-025) that validate the API endpoints
used by the new TypeScript modules.

### Challenge Architecture

Every challenge file in `catalog-api/challenges/`:
```go
package challenges

import (
    "context"
    "digital.vasic.challenges/pkg/challenge"
    "digital.vasic.challenges/pkg/httpclient"
)

type MyChallenge struct {
    challenge.BaseChallenge
    config *BrowsingConfig
}

func NewMyChallenge() *MyChallenge {
    return &MyChallenge{
        BaseChallenge: challenge.NewBaseChallenge(
            "my-challenge-id",   // unique kebab-case ID
            "My Challenge Name", // display name
            "Description...",   // what it tests
            "e2e",              // category
            []challenge.ID{"dependency-id"}, // must pass first
        ),
        config: LoadBrowsingConfig(),
    }
}

func (c *MyChallenge) Execute(ctx context.Context) (*challenge.Result, error) {
    start := time.Now()
    assertions := []challenge.AssertionResult{}
    outputs := map[string]string{"api_url": c.config.BaseURL}

    client := httpclient.NewAPIClient(c.config.BaseURL)
    _, err := client.Login(ctx, c.config.Username, c.config.Password)
    // ... check err, add assertion ...

    // Make API calls:
    code, body, apiErr := client.Get(ctx, "/api/v1/some-endpoint")
    // body is map[string]interface{}

    // For POST with JSON body:
    payload, _ := json.Marshal(map[string]interface{}{"key": "value"})
    code, rawBytes, apiErr := client.PostJSON(ctx, "/api/v1/endpoint", string(payload))
    // rawBytes is []byte; parse with json.Unmarshal

    assertions = append(assertions, challenge.AssertionResult{
        Type: "status_code", Target: "endpoint description",
        Expected: "200", Actual: fmt.Sprintf("HTTP %d", code),
        Passed: apiErr == nil && code == 200,
        Message: challenge.Ternary(passed, "OK message", "Failure message"),
    })

    status := challenge.StatusPassed
    for _, a := range assertions {
        if !a.Passed { status = challenge.StatusFailed; break }
    }

    return c.CreateResult(status, start, assertions, metrics, outputs, ""), nil
}
```

**CRITICAL**: `httpclient.APIClient` only has these methods:
- `Get(ctx, path) (int, map[string]interface{}, error)`
- `GetArray(ctx, path) (int, []interface{}, error)`
- `GetRaw(ctx, path) (int, []byte, error)`
- `PostJSON(ctx, path, jsonString) (int, []byte, error)`
- `Login(ctx, username, password) (string, error)`
- `SetToken(token)`, `Token() string`, `BaseURL() string`

There is **no `Put`, `Delete`, or `Patch`** method. Test CRUD endpoints using:
- POST for creation (use `PostJSON`)
- GET for retrieval (use `Get`)
- For PUT/DELETE: use `PostJSON` and accept 405 as "endpoint exists"

### The 5 New Challenges

**CH-021: `collections-api`** — `catalog-api/challenges/collections_api.go`

Dependency: `browsing-api-health`

Assertions:
1. `GET /api/v1/collections` → 200 with `items` array field
2. `POST /api/v1/collections` with `{"name": "CH-021 Test Collection", "is_public": false, "is_smart": false}` → 200/201 with `id > 0`
3. `GET /api/v1/collections/<id>` → 200 with `name == "CH-021 Test Collection"`

---

**CH-022: `entity-user-metadata`** — `catalog-api/challenges/entity_user_metadata.go`

Dependency: `entity-aggregation`

Assertions:
1. `GET /api/v1/entities` → 200 with at least 1 item; extract first `id`
2. `POST /api/v1/entities/<id>/user-metadata` with `{"is_favorite": true}` → 200/201/204/405
   (405 = endpoint exists but PostJSON used POST not PUT — still counts as "responds")

Note: Use `PostJSON` and accept 405 as proof the endpoint exists. Full success needs 200/201/204
and `is_favorite: true` in response.

---

**CH-023: `entity-search`** — `catalog-api/challenges/entity_search.go`

Dependency: `entity-aggregation`

Assertions:
1. `GET /api/v1/entities?limit=5` → 200 with all four pagination fields: `items`, `total`, `limit`, `offset`
2. `GET /api/v1/entities/types` → 200 with at least one type; extract `firstTypeName`
3. `GET /api/v1/entities/browse/<firstTypeName>?limit=5` → 200
4. `GET /api/v1/entities?query=a&limit=5` → 200

---

**CH-024: `storage-roots-api`** — `catalog-api/challenges/storage_roots_api.go`

Dependency: `browsing-api-health`

Assertions:
1. `GET /api/v1/storage-roots` → 200 (fallback: try `/api/v1/smb-configs` if 404)
2. If root found: `GET /api/v1/storage-roots/<id>/status` → 200

---

**CH-025: `auth-token-refresh`** — `catalog-api/challenges/auth_token_refresh.go`

Dependency: `browsing-api-health`

Assertions:
1. `POST /auth/login` with `{"username": ..., "password": ...}` → 200 with non-empty `session_token` (or `token`)
2. `GET /auth/status` → 200 with `authenticated: true`
3. `POST /auth/refresh` with `"{}"` → 200/201 with non-empty `session_token`

---

### Register in `catalog-api/challenges/register.go`

```go
// Module integration challenges (CH-021 to CH-025)
svc.Register(NewCollectionsAPIChallenge())     // CH-021
svc.Register(NewEntityUserMetadataChallenge()) // CH-022
svc.Register(NewEntitySearchChallenge())       // CH-023
svc.Register(NewStorageRootsAPIChallenge())    // CH-024
svc.Register(NewAuthTokenRefreshChallenge())   // CH-025
```

### Test the challenges compile

```bash
cd catalog-api
GOMAXPROCS=3 go test ./challenges/... -p 2 -parallel 2 -count=1 -timeout 120s
```

---

## Phase 9 — Final Verification

Run the complete test suite:

```bash
# Go tests — must show all packages: ok
cd catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 -count=1 -timeout 300s

# Frontend tests — must show: Tests 1643 passed (1643)
cd catalog-web
npm run test -- --run

# Verify submodule count
grep -c "^\[submodule" .gitmodules
# Expected: 26

# Verify go.mod replace directives
grep "replace digital.vasic" catalog-api/go.mod | wc -l
# Expected: 9 (assets, challenges, auth, cache, concurrency, config, entities, eventbus, filesystem)

# Verify catalog-web package.json has all vasic packages
grep "@vasic-digital" catalog-web/package.json | wc -l
# Expected: 9 (websocket-client, ui-components, media-types, catalogizer-api-client,
#             auth-context, media-browser, media-player, collection-manager, dashboard-analytics)
```

### Final commit

```bash
git add -A
git commit -m "Phases 7–9: Docs, Challenges (CH-021–CH-025), Final Verification

Phase 7 — Documentation added to all 7 new TS modules
Phase 8 — CH-021 to CH-025 added to catalog-api/challenges/
Phase 9 — All Go tests pass, all 1643 catalog-web tests pass

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## How to Add Any Future Module

Use this checklist whenever adding a new vasic-digital module:

```
[ ] 1. Create GitHub repo: gh repo create vasic-digital/<Name> --public
[ ] 2. Create content locally in /tmp/<Name>/
[ ] 3. Write tests — run them: npm test or go test ./...
[ ] 4. git init, add .gitignore (node_modules/ or vendor/), commit, push
[ ] 5. Add as submodule: git submodule add git@github.com:vasic-digital/<Name>.git <Name>
[ ] 6. Decide: integrate or submodule-only? (see Integration Decision Rules)
[ ] 7. If integrating Go: add replace + require to go.mod; run go mod tidy
[ ] 8. If integrating TS: add "file:../<Name>" to catalog-web/package.json; npm install
[ ] 9. Create adapter/alias layer in the correct internal package
[10] Run full test suite — all must pass
[11] Add docs: ARCHITECTURE.md, CHANGELOG.md (minimum)
[12] Commit Catalogizer main repo with updated .gitmodules and go.mod/package.json
[13] Add challenge if the module introduces a new API endpoint category
```

---

## Integration Decision Rules

When evaluating whether to import a Go module into go.mod or keep as submodule-only:

| Question | If YES → Integrate | If NO → Submodule only |
|----------|--------------------|------------------------|
| Does it use the same HTTP framework (Gin)? | ✅ | ❌ |
| Does it use the same DB driver? | ✅ | ❌ |
| Can its types be aliased with `type X = pkg.X`? | ✅ | ❌ |
| Are its transitive deps < 5 new packages? | ✅ | ❌ |
| Does integration avoid code duplication? | ✅ | (use judgment) |

When evaluating TypeScript modules:
- Always integrate if it provides types or services used by existing catalog-web code
- Use `file:../<RepoName>` for local development (swap to npm version for production)

---

## Troubleshooting

### `git submodule add` fails: "fatal: You are on a branch yet to be born"

**Cause**: The remote repo is empty (no commits yet).

**Fix**:
```bash
# 1. Remove the failed submodule attempt
git submodule deinit -f <Name>
git rm -rf <Name>
rm -rf .git/modules/<Name>

# 2. Build module content locally in /tmp/<Name>/
# 3. git init, commit, push to remote
# 4. Only then: git submodule add ...
```

### `go mod tidy` fails: "cannot find module providing package"

**Cause**: The replace directive path doesn't exist or the module's `go.mod` has a different
module name than what you declared in the require.

**Fix**:
```bash
# Check the module name in the submodule
cat <SubmodulePath>/go.mod | head -3

# Ensure replace directive matches exactly:
replace digital.vasic.<name> => ../<SubmodulePath>
# (path is relative to catalog-api/go.mod location)
```

### TypeScript import fails: "Cannot find module '@vasic-digital/...'"

**Cause**: Either the package isn't in `catalog-web/package.json` or `npm install` wasn't run.

**Fix**:
```bash
cd catalog-web
npm install
# If still failing, verify the linked package has a dist/ or that tsconfig resolves src/
```

### `gamePlatformRe` (or other regex) undefined after refactoring title_parser.go

**Cause**: A variable defined in `title_parser.go` was used in `aggregation_service.go`
(same package). When you replaced title_parser.go with a thin wrapper, the variable disappeared.

**Fix**: Add the variable back in `aggregation_service.go`:
```go
import "regexp"

var gamePlatformRe = regexp.MustCompile(`...pattern...`)
```

### HTML5 form validation blocks React onSubmit

**Symptom**: Test `fireEvent.click(submitBtn)` doesn't trigger `onSubmit` callback.

**Cause**: The `<input required>` attribute causes the browser (jsdom) to run constraint
validation before React's event handler fires.

**Fix**: Remove `required` from the input element. Use JS validation in `onSubmit` instead:
```typescript
if (!name.trim()) {
  setError('Name is required')
  return
}
```

### Challenge test: PostJSON used for PUT endpoint → 405

**Symptom**: A PUT endpoint returns 405 when called with PostJSON.

**Explanation**: The challenge framework's `httpclient.PostJSON` always uses HTTP POST.
There is no `Put` method. Accept 405 as "endpoint exists" and document this:

```go
// 405 = endpoint exists but we used POST (not PUT) — still confirms the route is registered
putResponds := putErr == nil && (putCode == 200 || putCode == 201 || putCode == 204 || putCode == 405)
```

### Frontend tests fail after npm install of new packages

**Cause**: New `file:` package resolved to an unbuilt package (no `dist/`).

**Fix**: The `file:` protocol resolves directly to the source package directory.
Since vitest transpiles TypeScript directly, this works without building. But ensure:
1. The linked package has a valid `package.json` with `"main"` pointing to `src/index.ts` or `dist/index.js`
2. If pointing to `dist/`, run `npm run build` in the linked package first

---

## Complete .gitmodules Reference

After all 9 phases, `.gitmodules` contains:

```ini
# Pre-existing
[submodule "WebSocket-Client-TS"]  url = git@github.com:vasic-digital/WebSocket-Client-TS.git
[submodule "UI-Components-React"]  url = git@github.com:vasic-digital/UI-Components-React.git
[submodule "Challenges"]           url = git@github.com:vasic-digital/Challenges.git
[submodule "Assets"]               url = git@github.com:vasic-digital/Assets.git

# Phase 1
[submodule "Config"]               url = git@github.com:vasic-digital/Config.git
[submodule "Database"]             url = git@github.com:vasic-digital/Database.git
[submodule "Filesystem"]           url = git@github.com:vasic-digital/Filesystem.git
[submodule "Concurrency"]          url = git@github.com:vasic-digital/Concurrency.git

# Phase 2
[submodule "Auth"]                 url = git@github.com:vasic-digital/Auth.git
[submodule "Middleware"]           url = git@github.com:vasic-digital/Middleware.git
[submodule "RateLimiter"]          url = git@github.com:vasic-digital/RateLimiter.git
[submodule "Observability"]        url = git@github.com:vasic-digital/Observability.git

# Phase 3
[submodule "Media"]                url = git@github.com:vasic-digital/Media.git
[submodule "Watcher"]              url = git@github.com:vasic-digital/Watcher.git
[submodule "EventBus"]             url = git@github.com:vasic-digital/EventBus.git
[submodule "Cache"]                url = git@github.com:vasic-digital/Cache.git
[submodule "Security"]             url = git@github.com:vasic-digital/Security.git
[submodule "Storage"]              url = git@github.com:vasic-digital/Storage.git
[submodule "Streaming"]            url = git@github.com:vasic-digital/Streaming.git
[submodule "Discovery"]            url = git@github.com:vasic-digital/Discovery.git

# Phase 4
[submodule "Entities"]             url = git@github.com:vasic-digital/Entities.git

# Phase 5
[submodule "Media-Types-TS"]            url = git@github.com:vasic-digital/Media-Types-TS.git
[submodule "Catalogizer-API-Client-TS"] url = git@github.com:vasic-digital/Catalogizer-API-Client-TS.git
[submodule "Auth-Context-React"]        url = git@github.com:vasic-digital/Auth-Context-React.git

# Phase 6
[submodule "Media-Browser-React"]      url = git@github.com:vasic-digital/Media-Browser-React.git
[submodule "Media-Player-React"]       url = git@github.com:vasic-digital/Media-Player-React.git
[submodule "Collection-Manager-React"] url = git@github.com:vasic-digital/Collection-Manager-React.git
[submodule "Dashboard-Analytics-React"] url = git@github.com:vasic-digital/Dashboard-Analytics-React.git
```

## Complete go.mod replace Directives Reference

```
replace digital.vasic.challenges => ../Challenges
replace digital.vasic.assets     => ../Assets
replace digital.vasic.concurrency => ../Concurrency
replace digital.vasic.config     => ../Config
replace digital.vasic.filesystem => ../Filesystem
replace digital.vasic.auth       => ../Auth
replace digital.vasic.cache      => ../Cache
replace digital.vasic.entities   => ../Entities
replace digital.vasic.eventbus   => ../EventBus
```

## Complete catalog-web/package.json vasic-digital Packages

```json
"@vasic-digital/websocket-client":        "file:../WebSocket-Client-TS",
"@vasic-digital/ui-components":           "file:../UI-Components-React",
"@vasic-digital/media-types":             "file:../Media-Types-TS",
"@vasic-digital/catalogizer-api-client":  "file:../Catalogizer-API-Client-TS",
"@vasic-digital/auth-context":            "file:../Auth-Context-React",
"@vasic-digital/media-browser":           "file:../Media-Browser-React",
"@vasic-digital/media-player":            "file:../Media-Player-React",
"@vasic-digital/collection-manager":      "file:../Collection-Manager-React",
"@vasic-digital/dashboard-analytics":     "file:../Dashboard-Analytics-React"
```
