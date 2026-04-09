# AGENTS.md - Catalogizer Development Guide

Essential commands and style guidelines for AI agents working in the Catalogizer codebase.

## Project Overview

**Catalogizer** is a comprehensive multi-platform media collection management system that automatically detects, categorizes, and organizes media files across multiple storage protocols (SMB, FTP, NFS, WebDAV, local filesystem). It provides real-time monitoring, advanced analytics, and modern client applications.

### System Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| **catalog-api** | Go 1.25 / Gin | REST API backend with HTTP/3 (QUIC) support |
| **catalog-web** | React 18 / TypeScript / Vite | Modern responsive web application |
| **catalogizer-desktop** | Tauri (Rust + React) | Cross-platform desktop application |
| **installer-wizard** | Tauri (Rust + React) | SMB configuration and installation wizard |
| **catalogizer-android** | Kotlin / Jetpack Compose / Hilt | Android mobile application |
| **catalogizer-androidtv** | Kotlin / Jetpack Compose / Hilt | Android TV application with channel integration |
| **catalogizer-api-client** | TypeScript | Reusable API client library |

### Version Information
- Current Version: 2.2.0 (build 18)
- Go Version: 1.25.7
- Node.js: 18+
- Kotlin: 1.9.22
- Android Gradle Plugin: 8.2.2

## ⚠️ CRITICAL CONSTRAINTS

### MANDATORY: .devignore Devices - NEVER USE FOR TESTING

**Devices listed in `.devignore` MUST NEVER be used for any testing, QA, or app deployment.**

- **ATMOSphere devices are explicitly excluded** - never install or test on them
- **Always check `.devignore` before any ADB device operation**
- **Only use devices NOT matching any pattern in `.devignore`**
- **If no valid devices are connected, abort testing** - do not proceed with excluded devices
- **This constraint applies to HelixQA, manual testing, and all QA workflows**

Pre-flight check:
```bash
DEVICE_MODEL=$(adb -s $DEVICE shell getprop ro.product.model)
if grep -qi "$DEVICE_MODEL" .devignore; then
  echo "❌ Device $DEVICE_MODEL is in .devignore - CANNOT USE"
  exit 1
fi
```

### MANDATORY: HelixQA ONLY for UI/UX Automated Testing

**ALL automated UI/UX testing MUST be performed exclusively by HelixQA.**

- **NEVER** write custom scripts for UI testing (no ADB tap sequences, no coordinate-based scripts)
- **NEVER** use third-party UI testing tools outside of HelixQA
- **HelixQA is the SOLE authorized tool** for all automated UI/UX testing across all platforms
- **Every UI interaction** MUST be LLM vision-driven: screenshot → LLM analysis → action decision
- **No workarounds**: If HelixQA finds an issue, fix it in the app code, not with testing scripts
- **This applies to**: Android TV, Android mobile, Web, Desktop, and any future platforms
- **Manual testing** is allowed for debugging, but **automated QA = HelixQA only**

### MANDATORY: Real-Time Log Monitoring During ALL QA Sessions

**Real-time log monitoring is MANDATORY for ALL QA sessions across ALL platforms.**

- **Android/Android TV**: Active `adb logcat` monitoring during HelixQA execution
- **Web**: Browser console logs must be captured and streamed in real-time
- **Desktop**: Application logs and system logs must be actively monitored
- **Backend Services**: Service logs, error logs, and metrics must be watched
- **ANR Detection**: Application Not Responding errors must be detected IMMEDIATELY
- **Crash Detection**: Fatal exceptions and crashes must pause the QA session
- **Log Analysis**: Must happen in real-time, not post-session analysis
- **Session Validity**: NO QA session is considered valid without real-time log monitoring

**Crash/ANR Protocol:**
1. Real-time log monitor detects ANR/crash
2. QA session is immediately paused
3. Full logs and stack traces are captured
4. Issue is analyzed and root cause identified
5. Application code is fixed
6. Regression test verifies the fix
7. Full QA session resumes only after verification

### MANDATORY: Universal Solution Principle - Works With ANY Application

**ALL fixes, workarounds, and testing infrastructure MUST be UNIVERSAL and work with ANY application.**

- **NEVER add test-only code to the application under test** (no `QAInputReceiver`, no test endpoints, no bypasses)
- **NEVER modify the target application** to make it "testable" or facilitate testing
- **ALWAYS implement fixes in the testing tool/infrastructure** (HelixQA, test frameworks, automation tools)
- **HelixQA must handle text input** via on-screen keyboard navigation, not app modifications
- **If detection/monitoring fails** → Improve the testing framework, not the app
- **Target applications require ZERO modifications** for testing
- **Universal solutions ensure**: portability, maintainability, valid test results, reusability
- **App-specific testing code is PROHIBITED** and must be removed/reimplemented universally

### MANDATORY: Device Auto-Connect via .devconnect (Before HelixQA)

**Ensure Android TV devices are connected BEFORE running HelixQA.**

- **`.devconnect`** file: List of IP addresses to auto-connect via `adb connect`
- **Opposite of `.devignore`**: Ensures devices ARE connected (not excluded)
- **Git ignored**: Never commit `.devconnect` (contains local network IPs)
- **Pre-flight requirement**: Run `./scripts/devconnect.sh` before every HelixQA session
- **Validation**: Script pings devices first, only connects reachable devices
- **Idempotent**: Safe to run multiple times

**Pre-flight checklist:**
```bash
# 1. Check .devignore - ensure devices are NOT excluded
# 2. Check .devconnect - ensure devices ARE listed
grep -v "^#" .devconnect | grep -v "^$"

# 3. Auto-connect devices
./scripts/devconnect.sh

# 4. Verify connection
adb devices

# 5. Run HelixQA
./HelixQA/bin/helixqa autonomous -platforms android ...
```

### ⚠️⚠️⚠️ ABSOLUTELY MANDATORY: ZERO UNFINISHED WORK POLICY

**NO unfinished work, TODOs, or known issues may remain in the codebase. EVER.**

This is a **ZERO TOLERANCE** policy for all code, tests, scripts, and documentation.

**PROHIBITED:**
- ❌ **TODO/FIXME comments** in committed code
- ❌ **Empty implementations** with "// Implement later"
- ❌ **Silent error ignoring** (`_ = err` patterns in production code)
- ❌ **Hardcoded fake data** or fabricated metrics
- ❌ **Coverage fraud** - tests that inflate coverage without testing logic
- ❌ **unwrap() calls** in Rust that can panic (use proper error handling)
- ❌ **Empty catch blocks** in TypeScript/JavaScript
- ❌ **Partial implementations** left for "future completion"
- ❌ **Known bugs** documented but not fixed
- ❌ **"Temporary" workarounds** that become permanent

**REQUIRED:**
- ✅ **Fix ALL discovered issues immediately** - no deferrals
- ✅ **When fixing, fix ALL instances** - not just the reported one
- ✅ **Complete implementations** before committing
- ✅ **Proper error handling** in ALL code paths
- ✅ **Real test assertions** - no fake coverage
- ✅ **Code compiles without warnings**
- ✅ **Zero outstanding issues** at commit time

**Definition of "Done":**
1. Feature/bug fix is fully implemented
2. All TODOs are resolved (implemented or removed)
3. All error cases handled properly
4. All tests pass with real assertions
5. No fake/hardcoded data remains
6. Code review passes with ZERO outstanding issues
7. Documentation is updated
8. No compiler warnings or linter errors

**Enforcement:**
- Pre-commit hooks block commits with TODO/FIXME patterns
- CI/CD fails builds with unresolved issues
- Code reviews reject PRs with known unfinished work
- **This policy applies to ALL repositories, submodules, and branches**

**Quality Principle:**
> "If it's not finished, it doesn't ship. If it ships, it's finished."

## Submodule Architecture

The project uses 41 independent git submodules under the `digital.vasic.*` and `@vasic-digital/*` namespace for generic, reusable functionality. Each submodule has its own repository, tests, and documentation.

### Go Modules (23 modules)

Wired via `replace` directives in `catalog-api/go.mod`:

| Module | Path | Package | Purpose |
|--------|------|---------|---------|
| Assets | `Assets/` | `digital.vasic.assets` | Asset management (lazy loading, serving, defaults) |
| Auth | `Auth/` | `digital.vasic.auth` | JWT authentication, bcrypt password helpers |
| Cache | `Cache/` | `digital.vasic.cache` | Redis-backed caching with TTL management |
| Challenges | `Challenges/` | `digital.vasic.challenges` | Structured test scenario framework |
| Concurrency | `Concurrency/` | `digital.vasic.concurrency` | Retry with backoff, offline cache patterns |
| Config | `Config/` | `digital.vasic.config` | Configuration management (env, file, validation) |
| Containers | `Containers/` | `digital.vasic.containers` | Container discovery and service port detection |
| Database | `Database/` | `digital.vasic.database` | Migration patterns, dual SQLite/PostgreSQL support |
| Discovery | `Discovery/` | `digital.vasic.discovery` | Network/service discovery (SMB, mDNS) |
| Entities | `Entities/` | `digital.vasic.entities` | Entity model definitions |
| EventBus | `EventBus/` | `digital.vasic.eventbus` | Typed event channels and pub/sub |
| Filesystem | `Filesystem/` | `digital.vasic.filesystem` | Unified multi-protocol client (SMB, FTP, NFS, WebDAV, local) |
| Lazy | `Lazy/` | `digital.vasic.lazy` | Generic lazy loading utilities |
| Media | `Media/` | `digital.vasic.media` | Media detection, analysis, and metadata extraction |
| Memory | `Memory/` | `digital.vasic.memory` | Memory leak detection |
| Middleware | `Middleware/` | `digital.vasic.middleware` | HTTP middleware (CORS, logging, recovery, request ID) |
| Observability | `Observability/` | `digital.vasic.observability` | Prometheus metrics and OpenTelemetry integration |
| RateLimiter | `RateLimiter/` | `digital.vasic.ratelimiter` | Pluggable rate limiting (memory, Redis, sliding window) |
| Recovery | `Recovery/` | `digital.vasic.recovery` | Circuit breaker and recovery patterns |
| Security | `Security/` | `digital.vasic.security` | CORS config, CSP headers, request sanitization |
| Storage | `Storage/` | `digital.vasic.storage` | Object storage abstraction (MinIO/S3-compatible) |
| Streaming | `Streaming/` | `digital.vasic.streaming` | WebSocket hub with room/topic support |
| Watcher | `Watcher/` | `digital.vasic.watcher` | Filesystem watcher with debouncing and filtering |

### TypeScript/React Modules (9 modules)

Linked via `file:../` in `catalog-web/package.json`:

| Module | Path | Package | Purpose |
|--------|------|---------|---------|
| Auth Context | `Auth-Context-React/` | `@vasic-digital/auth-context` | React authentication context provider |
| API Client TS | `Catalogizer-API-Client-TS/` | `@vasic-digital/catalogizer-api-client` | TypeScript API client |
| Collection Manager | `Collection-Manager-React/` | `@vasic-digital/collection-manager` | Collection management UI components |
| Dashboard Analytics | `Dashboard-Analytics-React/` | `@vasic-digital/dashboard-analytics` | Dashboard and analytics components |
| Media Browser | `Media-Browser-React/` | `@vasic-digital/media-browser` | Media browsing components |
| Media Player | `Media-Player-React/` | `@vasic-digital/media-player` | Media playback components |
| Media Types | `Media-Types-TS/` | `@vasic-digital/media-types` | Shared media type definitions |
| UI Components | `UI-Components-React/` | `@vasic-digital/ui-components` | Reusable React UI component library |
| WebSocket Client | `WebSocket-Client-TS/` | `@vasic-digital/websocket-client` | WebSocket client with React hooks |

### HelixQA / AI Submodules (9 modules)

| Module | Path | Purpose |
|--------|------|---------|
| HelixQA | `HelixQA/` | LLM-driven autonomous QA testing pipeline |
| DocProcessor | `DocProcessor/` | Document processing |
| LLMOrchestrator | `LLMOrchestrator/` | LLM orchestration layer |
| LLMProvider | `LLMProvider/` | LLM provider abstraction |
| VisionEngine | `VisionEngine/` | Vision model integration |
| ReplayBuffer | `ReplayBuffer/` | Experience replay for QA sessions |
| ScreenDiff | `ScreenDiff/` | Screenshot diff analysis |
| TrainingCollector | `TrainingCollector/` | Training data collection |
| VisualRegression | `VisualRegression/` | Visual regression testing |

### Submodule Commands

```bash
# Initialize submodules after cloning
git submodule init && git submodule update --recursive

# Set up a new submodule with upstream remotes
./scripts/setup-submodule.sh ModuleName [--create-repos] [--go|--ts|--kotlin]

# Push to all upstreams (from within submodule)
cd SubmoduleName && commit "message"

# Install upstream remotes (from within submodule)
cd SubmoduleName && install_upstreams
```

## Build / Lint / Test Commands

### Backend (catalog-api)

```bash
cd catalog-api
go run main.go                                          # dev server (writes .service-port)
go build -o catalog-api                                 # build binary
GOMAXPROCS=3 go test ./... -p 2 -parallel 2            # all tests (resource-limited)
go test -v -run ^TestName$ ./path/to/pkg/               # single test (regex match)
go test -v -run ^TestSuiteName/TestSubtest$ ./path/     # single subtest in suite
go test -cover ./...                                    # coverage
go fmt ./... && go vet ./...                            # format + lint
```

### Frontend (catalog-web)

```bash
cd catalog-web
npm run dev                                             # dev server (port 3000, proxies /api)
npm run build                                           # production build (tsc + vite)
npm run lint                                            # ESLint (--max-warnings 0)
npm run lint:fix                                        # auto-fix lint issues
npm run type-check                                      # tsc --noEmit
npm run test                                            # Vitest (single run)
npm run test -- -t "test name pattern"                  # single test by name
npm run test:watch                                      # watch mode
npm run test:coverage                                   # coverage (v8)
npm run test:e2e                                        # Playwright E2E
npm run test:e2e -- --grep "test title"                 # single E2E test
```

### Desktop (catalogizer-desktop / installer-wizard)

```bash
cd catalogizer-desktop  # or installer-wizard
npm run tauri:dev       # dev with hot reload
npm run tauri:build     # build for platform
npm run test            # unit tests
```

### Android (catalogizer-android / catalogizer-androidtv)

```bash
cd catalogizer-android  # or catalogizer-androidtv
./gradlew test                                          # all unit tests
./gradlew test --tests "*TestClassName"                 # single test class
./gradlew test --tests "*TestClassName.testMethod"      # single test method
./gradlew assembleDebug                                 # debug APK
./gradlew lintKotlin                                    # lint
```

### Full System

```bash
podman-compose -f docker-compose.dev.yml up             # dev environment
./scripts/services-up.sh                                # start all services
./scripts/services-down.sh                              # stop all services
./scripts/run-all-tests.sh                              # all tests + security scans
./scripts/release-build.sh --container --force --skip-tests  # release build
```

## Code Style - Go Backend

- **Naming**: PascalCase exported, camelCase unexported. Interfaces: `Reader`, `Writer`, `Service` suffixes.
- **Receivers**: Single-letter (`s *Service`, `h *Handler`, `r *Repository`).
- **Imports**: Three groups separated by blank lines — stdlib, third-party, local:
  ```go
  import (
      "encoding/json"
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/stretchr/testify/assert"

      "catalogizer/database"
      "catalogizer/models"
  )
  ```
- **Constructors**: `NewService(dep Dependency) *Service` with dependency injection.
- **Error handling**: Wrap with `fmt.Errorf("context: %w", err)`. Use `errors.New` for static errors. Never expose internal details to clients.
- **Formatting**: `go fmt` (or `gofumpt`). All exported functions/types need doc comments.
- **Testing**: Table-driven tests with `t.Run`. Use `testify/suite` for complex suites, `testify/mock` for mocks. Files: `*_test.go` beside source. Use `database.WrapDB()` for in-memory SQLite test DB.
- **Concurrency**: Services spawning goroutines (`CacheService`, `WebSocketHandler`) use `sync.Once` for cleanup. Tests MUST `defer service.Close()` / `handler.Stop()`.
- **Database**: Use `?` placeholders (auto-converted to `$1, $2...` for Postgres). Use `InsertReturningID()` instead of `LastInsertId()`.

## Code Style - TypeScript/React Frontend

- **Naming**: PascalCase components/interfaces, camelCase functions/variables, SCREAMING_SNAKE_CASE constants.
- **Components**: Functional components with explicit TypeScript interfaces:
  ```tsx
  interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    loading?: boolean
  }
  const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, loading, children, ...props }, ref) => { /* ... */ }
  )
  ```
- **Imports**: Three groups — React, third-party, local path aliases:
  ```tsx
  import React from 'react'
  import { cva, type VariantProps } from 'class-variance-authority'
  import { cn } from '@/lib/utils'
  ```
- **Path aliases**: `@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, `@/assets`.
- **Formatting**: Prettier. Tailwind classes composed via `cn()` from `@/lib/utils`.
- **Linting**: ESLint with `@typescript-eslint`, `react`, `react-hooks`, `security`. Unused vars prefixed with `_`. `--max-warnings 0` enforced.
- **State**: React Query for server state, Zustand for client state.
- **Forms**: React Hook Form + Zod validation (`@hookform/resolvers`).
- **Testing**: Vitest + React Testing Library. Files: `__tests__/` or `.test.tsx` beside source. Playwright for E2E.

## Code Style - Kotlin/Android

- **Naming**: PascalCase classes, camelCase functions/variables.
- **Architecture**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit.
- **DI**: Hilt for dependency injection.
- **Async**: `suspend` functions, `Flow`/`StateFlow`, Paging 3.
- **Error handling**: Sealed `Result` classes for operation outcomes.
- **Testing**: JUnit 4 + MockK/Mockito. Coroutines via `kotlinx-coroutines-test`.
- **Build**: JDK 21 with `--add-opens` JVM args for kapt compatibility.

## Database (Dual Dialect)

SQLite (dev) and PostgreSQL (prod) via `database.DB` wrapper:

```go
db.Query("SELECT * FROM table WHERE created_at > ?", cutoff)
if db.Dialect().IsPostgres() {
    expr = "EXTRACT(EPOCH FROM (MAX(t) - MIN(t)))"
} else {
    expr = "(julianday(MAX(t)) - julianday(MIN(t))) * 86400"
}
```

- `RewritePlaceholders()` — `?` → `$1, $2, ...` for PostgreSQL
- `RewriteInsertOrIgnore()` — `INSERT OR IGNORE` → `ON CONFLICT DO NOTHING`
- `BooleanLiterals()` — `= 0/1` → `= FALSE/TRUE` for known boolean columns
- `InsertReturningID()` and `TxInsertReturningID()` replace `LastInsertId()` (PostgreSQL uses `RETURNING id`)
- **SQLite WAL mode**: Explicit `PRAGMA journal_mode=WAL` after connection
- Migrations in `database/migrations/` — separate SQLite and PostgreSQL variants

## Architecture Overview

### catalog-api (Go/Gin)

Handler → Service → Repository → SQLite/PostgreSQL. Routes under `/api/v1` in `main.go`.

- **Dual package layout**: top-level `handlers/`, `repository/`, `services/`, `middleware/` for domain logic; `internal/handlers/`, `internal/services/`, `internal/middleware/` for infrastructure concerns.
- `filesystem/interface.go` defines `UnifiedClient`; `filesystem/factory.go` creates per-protocol clients. New protocols: implement the interface.
- `internal/smb/`: circuit breaker + offline cache + exponential backoff retry.
- `internal/media/detector/` → `analyzer/` → `providers/` (TMDB, IMDB, etc.): detection pipeline.
- `internal/media/realtime/`: event bus → WebSocket → clients.
- `internal/auth/` + `middleware/`: JWT auth with role-based access.
- `internal/metrics/`: Prometheus metrics (exposed via `/metrics`).
- **Dynamic port binding**: On startup, writes chosen port to `.service-port` file. Frontend reads this for API proxy target.
- **HTTP/3 (QUIC)**: Uses `quic-go/http3` with self-signed TLS certs generated at startup.
- **Redis**: Optional caching layer via `go-redis/v9`.
- **Version injection**: `Version`, `BuildNumber`, `BuildDate` via `-ldflags` at build time.
- `internal/lifecycle/`: `LazyServiceRegistry` for deferred service initialization with dependency ordering.
- `internal/concurrency/`: Semaphore-based concurrency control for limiting parallel operations.
- `internal/httpclient/`: Pooled HTTP client with connection reuse, timeouts, and retry logic.

### Media Entity System

Scanned files are transformed into structured media entities via a post-scan aggregation pipeline:

```
UniversalScanner (scan completes)
       ↓ (post-scan hook)
AggregationService.AggregateAfterScan()
  ├── Title parser (regex: movie, TV, music, game, software)
  ├── MediaItem creation/update (media_items table)
  ├── MediaFile linking (media_files junction table)
  ├── Hierarchy builder (TV: show→season→episode, Music: artist→album→song)
  └── Duplicate detection (same title + type + year)
       ↓
Entity API (/api/v1/entities)
       ↓
Entity Browser UI (/browse, /entity/:id)
```

**11 media types** (seeded in `media_types` table): movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic.

### catalog-web (React/TypeScript/Vite)

AuthProvider → WebSocketProvider → Router.

Key tech: React Query (`@tanstack/react-query`) for server state, Zustand for client state, Tailwind CSS for styling, React Hook Form + Zod for forms, framer-motion for animations, Vitest for unit tests, Playwright for E2E tests.

- Auth-gated routes via `ProtectedRoute`.
- Path aliases configured in `vite.config.ts`.
- API proxy: reads `../catalog-api/.service-port` at dev server startup to resolve backend port (falls back to 8080).
- Build output split into vendor chunks: `vendor` (react), `router`, `ui`, `charts`, `utils`.

### Android TV Home Screen Channels (v2.3.0)

Full integration with Android TV's channel API (`androidx.tvprovider`):
- Default "Catalogizer Picks" channel auto-created on launch
- Dynamic per-category channels (one per media type with content)
- System Watch Next row for partially-watched items + auto-next-episode
- Deep linking via `catalogizer://media/{id}?type={type}`
- `WorkManager` periodic sync (6h) + app-launch + SyncService triggers
- Full cleanup on logout

Key files: `data/tv/TvChannelRepository.kt`, `data/tv/ChannelProgramMapper.kt`, `data/tv/WatchNextManager.kt`, `data/tv/TvChannelSyncWorker.kt`, `ui/ChannelDeepLinkActivity.kt`.

#### Automated Channels Testing via HelixQA

HelixQA includes a **generic, decoupled Android TV Channels testing framework** that automatically detects and tests all Channels functionality:

**Detection:** HelixQA scans the Android TV codebase for `androidx.tvprovider` API usage and automatically identifies:
- TvContractCompat integration patterns
- WatchNextManager implementations
- ChannelDeepLinkActivity handlers
- URI scheme configurations

**Generated Test Cases (30 comprehensive tests):**

| Test ID | Category | Description |
|---------|----------|-------------|
| ATV-CH-001 | Default Channel | Auto-creation on first launch |
| ATV-CH-002 | Default Channel | Content population (continue watching, recent, trending) |
| ATV-CH-003 | Default Channel | Browsable state verification |
| ATV-CH-004 | Default Channel | Duplicate prevention |
| ATV-CH-005 | Category Channels | Dynamic creation based on content |
| ATV-CH-006 | Category Channels | Display name localization |
| ATV-CH-007 | Category Channels | Empty category suppression |
| ATV-CH-008 | Category Channels | Stale channel removal |
| ATV-CH-009 | Watch Next | Continue watching entries (5-90% progress) |
| ATV-CH-010 | Watch Next | Completed item removal (>90%) |
| ATV-CH-011 | Watch Next | Minimum threshold exclusion |
| ATV-CH-012 | Watch Next | Auto next-episode surfacing |
| ATV-CH-013 | Watch Next | Stale entry cleanup (30+ days) |
| ATV-CH-014 | Sync | Content refresh on sync |
| ATV-CH-015 | Sync | WorkManager periodic sync (6h) |
| ATV-CH-016 | Sync | Launch-triggered sync |
| ATV-CH-017 | Sync | Manual sync trigger |
| ATV-CH-018 | Deep Links | Detail navigation from channels |
| ATV-CH-019 | Deep Links | Resume playback from Watch Next |
| ATV-CH-020 | Deep Links | URI format validation |
| ATV-CH-021 | Deep Links | App link intent URIs |
| ATV-CH-022 | Deep Links | Unauthenticated redirect handling |
| ATV-CH-023 | Security | Channel cleanup on logout |
| ATV-CH-024 | Security | Watch Next cleanup on logout |
| ATV-CH-025 | Security | Re-authentication restoration |
| ATV-CH-026 | Edge Cases | Invalid media ID handling |
| ATV-CH-027 | Edge Cases | No server connection graceful handling |
| ATV-CH-028 | Edge Cases | Program limit enforcement (max 30) |
| ATV-CH-029 | Edge Cases | Internal provider ID validation |
| ATV-CH-030 | Functional | Program metadata completeness |

**Framework Location:** `HelixQA/pkg/planning/androidtv_channels_framework.go`

**Usage:** Simply run HelixQA with androidtv platform - Channels tests are auto-generated:
```bash
./HelixQA/bin/helixqa autonomous -platforms androidtv
```

**Generic Framework:** The testing framework is app-agnostic. Any Android TV app using `androidx.tvprovider` will automatically receive comprehensive Channels testing by configuring a `ChannelFeatureSpec`.

## Challenge System

`digital.vasic.challenges` framework integrated via `Challenges/` submodule. Challenges are Go structs embedding `challenge.BaseChallenge` with custom `Execute()`. Registered in `catalog-api/challenges/register.go` via `RegisterAll()`, exposed via `/api/v1/challenges` REST endpoints. Challenge bank definitions loaded from `challenges/config/`.

**All challenge operations MUST be executed exclusively by system deliverables (compiled binaries) — the catalog-api service and other Catalogizer applications. Never use custom scripts, curl commands, or third-party tools to trigger API endpoints within challenge execution.**

Key constraints:
- `RunAll` is synchronous/blocking — no other challenge can run until it finishes.
- Progress-based liveness detection: 5-minute stale threshold kills stuck challenges.
- `challenge.NewConfig()` sets Timeout=5min by default — zero it to use runner's timeout.
- `config.json` `write_timeout` must be 900 (not 30) for long-running challenge RunAll.

### User Flow Automation

Multi-platform user flow automation via `Challenges/pkg/userflow/`. 174 Catalogizer-specific challenges across 4 platform groups:

| File | Platform | Challenges |
|------|----------|-----------|
| `userflow_api.go` | Go API (HTTP) | 49 |
| `userflow_web.go` | React web (Playwright) | 59 |
| `userflow_desktop.go` | Tauri desktop + wizard | 28 |
| `userflow_mobile.go` | Android + Android TV | 38 |

## Docker Compose Files

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Production stack |
| `docker-compose.dev.yml` | Development environment |
| `docker-compose.build.yml` | Containerized build pipeline |
| `docker-compose.test.yml` | Test stack (API, web, Playwright; `network_mode: host`) |
| `docker-compose.test-infra.yml` | Test infrastructure services |
| `docker-compose.security.yml` | Security scanning tools |
| `docker-compose.qa.yml` | QA environment |
| `docker-compose.qa-robot.yml` | QA robot configuration |

## Container Runtime

**Always use Podman** — this project uses Podman exclusively (no Docker). All container commands use `podman`/`podman-compose`.

```bash
podman-compose -f docker-compose.dev.yml up       # dev env
podman-compose -f docker-compose.yml config --quiet  # validate
podman run / podman build / podman ps              # single container commands
```

Critical container notes:
- Must use `podman build --network host` — default container networking has SSL issues.
- Must use `podman run --network host` for builds.
- Set `GOTOOLCHAIN=local` to prevent Go auto-downloading newer toolchain versions.
- Use fully qualified image names (`docker.io/library/...`) — short names fail without TTY.
- Set `APPIMAGE_EXTRACT_AND_RUN=1` in containers for Tauri AppImage bundling (no FUSE).
- API container needs `--add-host=synology.local:192.168.0.241` for NAS access.

## Constraints

### CRITICAL: API Keys and Secrets — NEVER Commit to Git

- **Never** commit `.env` files containing real API keys, tokens, or secrets
- **Never** hardcode API keys in source code, CLAUDE.md, AGENTS.md, or any tracked file
- Use `.env.example` with placeholder values only (e.g., `YOUR_API_KEY_HERE`)
- Verify `.gitignore` covers all `.env` files before every commit
- If an API key is accidentally committed, **rotate it immediately**
- All submodules MUST have `.env` in their `.gitignore`
- Pre-commit hooks scan for secrets via detect-secrets

### CRITICAL: Git Access via SSH Only — NEVER Use HTTPS

- **Always** use SSH (`git@github.com:user/repo.git`) for all Git operations
- **Never** use HTTPS (`https://github.com/user/repo.git`) for Git access
- Configure remotes to use SSH: `git remote set-url origin git@github.com:user/repo.git`
- For new clones: `git clone git@github.com:user/repo.git` — NOT `git clone https://...`
- For GitLab, GitFlic, GitVerse, and all other Git hosts: use SSH protocol exclusively
- HTTPS bypasses SSH key-based authentication and is less secure
- Submodules MUST be configured with SSH URLs in `.gitmodules`
- CI/CD scripts and automation MUST use SSH with deploy keys, never HTTPS with passwords/tokens

### CRITICAL: HelixQA — FULLY LLM-DRIVEN Autonomous Testing

HelixQA is a generic, universal QA tool driven entirely by LLM vision models. Pipeline: Learn → Plan → Execute → Curiosity → Analyze.

**Non-negotiable rules:**
- ALL navigation MUST be performed by real LLM vision models — the LLM sees a screenshot, decides the next action. Every single step.
- NEVER write hardcoded tap coordinates, sleep timers, keystroke sequences, or fallback scripts.
- Fix issues in HelixQA Go code (parsing, retry logic, prompts) — never work around with scripts.
- Every connected ADB device MUST be tested (except `.devignore` entries).
- QA priority: (1) Happy paths, (2) Standard flows, (3) Edge cases, (4) Adversarial.

### GitHub Actions are PERMANENTLY DISABLED

Do NOT create any GitHub Actions workflow files in `.github/workflows/`. CI/CD must be run locally.

### All Builds, Services, and QA Testing MUST Use Containers (Podman)

- **Builds**: Use `./scripts/release-build.sh --container` or `podman-compose -f docker-compose.build.yml`
- **Services**: Use `podman-compose` to run catalog-api, catalog-web, and all supporting services
- **QA Testing**: Use `./scripts/run-helixqa.sh` with containerized services
- **Android Emulators**: Run in containers via `docker-compose.test.yml --profile android`
- **Never build or run apps/services directly on bare metal** in production or QA contexts

### CRITICAL: Host Resource Limits (30-40% Maximum)

The host machine runs other mission-critical processes. All workloads MUST be limited to 30-40% of total host resources.

- **Go tests**: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- **Container CPU/memory limits** (mandatory):
  - PostgreSQL: `--cpus=1 --memory=2g`
  - API: `--cpus=2 --memory=4g`
  - Web: `--cpus=1 --memory=2g`
  - Builder: `--cpus=3 --memory=8g`
- **Total container budget**: max 4 CPUs, 8 GB RAM across all running containers
- **Challenges**: Run sequentially via the API, never in parallel
- **Monitor**: `podman stats --no-stream` and `cat /proc/loadavg`

### CRITICAL: HTTP/3 (QUIC) with Brotli Compression (Mandatory)

All network communication MUST use **HTTP/3 (QUIC)** with **Brotli compression**. Fallback: HTTP/2 + gzip. Never HTTP/1.1 in production.

- **catalog-api**: `quic-go/http3` server + Brotli middleware (`andybalholm/brotli`)
- **catalog-web**: Served via HTTP/3-capable reverse proxy, Brotli-compressed static assets
- **Tauri apps**: HTTP/3 client for API communication
- **Android apps**: OkHttp with HTTP/3 (Cronet) + Brotli
- **API client**: HTTP/3-capable fetch with Brotli Accept-Encoding

### Zero Warning / Zero Error Policy

All components must run with zero console warnings, zero console errors, and zero failed network requests in every environment.

- No browser console errors or warnings. Every failed network request is a defect.
- Every API endpoint the frontend calls must exist, return valid 2xx responses, and match expected shape.
- No framework deprecation warnings. No WebSocket connection failures.
- If a feature is not yet implemented, provide a stub endpoint that returns a valid empty response.

## Environment Configuration

### Config Precedence

`env vars > .env > config.json > defaults`

### Backend Environment Variables (.env)

```env
# Application
APP_ENV=development
LOG_LEVEL=debug
API_PORT=8080

# Database
DATABASE_TYPE=postgres  # or sqlite
POSTGRES_USER=catalogizer
POSTGRES_PASSWORD=change_me_in_production
POSTGRES_DB=catalogizer_dev
POSTGRES_PORT=5432

# For SQLite (alternative)
# SQLITE_PATH=./catalogizer.db

# Redis
REDIS_PORT=6379
REDIS_PASSWORD=

# Security
JWT_SECRET=your-super-secret-jwt-key-change-me-in-production

# CORS
CORS_ENABLED=true
CORS_ORIGINS=http://localhost:3000,http://localhost:19006

# SMB/File System
SMB_ENABLED=true
MEDIA_ROOT_PATH=./media

# Metadata Providers (optional)
TMDB_API_KEY=YOUR_TMDB_KEY_HERE
OMDB_API_KEY=YOUR_OMDB_KEY_HERE

# HelixQA / Vision Configuration
HELIX_VISION_HOSTS=thinker.local,amber.local
HELIX_VISION_MULTI_USER=milosvasic
ASTICA_API_KEY=YOUR_ASTICA_API_KEY_HERE
OPENAI_API_KEY=YOUR_OPENAI_KEY_HERE
ANTHROPIC_API_KEY=YOUR_ANTHROPIC_KEY_HERE
GEMINI_API_KEY=YOUR_GEMINI_API_KEY_HERE
```

## Testing Strategy

### Test Organization

- **Unit tests**: `*_test.go` files beside source code
- **Integration tests**: `catalog-api/internal/tests/`, `catalog-api/tests/integration/`
- **E2E tests**: Playwright in `catalog-web/e2e/`
- **Challenge tests**: `catalog-api/challenges/` — structured test scenarios
- **Stress tests**: `catalog-api/tests/stress/`
- **Security tests**: `catalog-api/tests/security/`

### Security Scanning

```bash
./scripts/security-scan.sh              # automated scanning
./scripts/run-sonarqube-scan.sh         # SonarQube code quality
./scripts/snyk-scan.sh                  # Snyk dependency scanning
```

Available tools:
- `govulncheck` — Go stdlib/dependency vulnerabilities
- `npm audit` — Frontend dependency vulnerabilities
- Semgrep — Static analysis for security anti-patterns
- Snyk, Trivy, Gosec via `docker-compose.security.yml`

### QA Campaign Protocol (Mandatory)

All QA campaigns MUST follow: **Rebuild → Execute all tests → Analyze results → Create tickets → Fix root causes → Create validation tests → Repeat**.

Loop stops only on: all pass, fatal blocker, or nothing left.

### Live Monitoring (Mandatory)

All test execution requires real-time status: platform, app/service, test case ID, description, progress, result. All output logged to `docs/reports/qa-sessions/qa-session-<date>/logs/`.

### Video Recording & Analysis (Mandatory)

All device/emulator QA sessions MUST record video. All recordings MUST be analyzed for: visual glitches, UI/UX issues, content gaps, brand compliance, performance, crashes.

### Fixes Validation Suite

Every bug fix MUST include a bank test entry in `fixes-validation.yaml` to prevent regression. Tests are permanent.

## Key Files

- `catalog-api/main.go` — API entry point, route registration
- `catalog-api/database/dialect.go` — dual-dialect SQL rewriting
- `catalog-api/filesystem/interface.go` — `UnifiedClient` protocol abstraction
- `catalog-api/challenges/register.go` — challenge registration
- `catalog-web/src/App.tsx` — React root (AuthProvider → WebSocketProvider → Router)
- `catalog-web/vite.config.ts` — path aliases, API proxy config
- `versions.json` — version tracking for all components
- `.env.example` — environment variable template

## Git

6 push targets configured on `origin` remote (2x GitHub, 2x GitLab, GitFlic, GitVerse). GitVerse uses port 2222.

```bash
# Push to all remotes
GIT_SSH_COMMAND="ssh -o BatchMode=yes" git push origin main

# Add hosts to known_hosts first
ssh-keyscan github.com gitlab.com gitflic.ru >> ~/.ssh/known_hosts
ssh-keyscan -p 2222 gitverse.ru >> ~/.ssh/known_hosts
```

`releases/` and `reports/` are gitignored — build artifacts are not version-controlled.

## Pre-Commit Checklist

- Go: `cd catalog-api && go fmt ./... && go vet ./...`
- TypeScript: `cd catalog-web && npm run lint && npm run type-check`
- Ensure zero console warnings/errors in browser
- Verify `.gitignore` covers `.env` — never commit secrets
- Run pre-commit hooks: `pre-commit run --all-files`

## Root Directory Structure (Mandatory Locations)

New files MUST be placed in the correct directory. Do NOT add files to the project root unless they are conventional root files (README, LICENSE, .gitignore, docker-compose, etc.).

| Directory | Purpose |
|-----------|---------|
| `catalog-api/` | Go backend API service |
| `catalog-web/` | React TypeScript web frontend |
| `catalogizer-android/` | Android mobile app (Kotlin) |
| `catalogizer-androidtv/` | Android TV app (Kotlin) |
| `catalogizer-desktop/` | Tauri desktop application |
| `catalogizer-api-client/` | TypeScript API client library |
| `installer-wizard/` | Tauri installation wizard |
| `challenges/` | Challenge bank definitions and runtime results |
| `config/` | Infrastructure config files (nginx.conf, redis.conf) |
| `scripts/` | Shell scripts (install, setup, CI/CD, testing runners) |
| `scripts/lib/` | Per-component build scripts (`build-*.sh`) |
| `tests/` | Standalone/integration test files |
| `docs/` | All documentation markdown files |
| `Assets/` | Static assets (images, HTML tutorials) — also a Go submodule |
| `Build/` | Generic build framework submodule |
| `build/` | Build output and container build context |
| `deployment/` | Deployment configurations |
| `monitoring/` | Monitoring and observability configs |
| `tools/` | Development tooling |
| `Upstreams/` | Git upstream remote configurations for submodules |
| `HelixQA/` | LLM-driven autonomous QA testing |
| `qa-results/` | QA session outputs and reports |

## Documentation

Comprehensive documentation is available in `docs/`:

- `docs/INSTALLATION_GUIDE.md` — Installation instructions
- `docs/DEVELOPER_GUIDE.md` — Developer setup and workflows
- `docs/DEPLOYMENT_GUIDE.md` — Production deployment
- `docs/SECURITY_TESTING_GUIDE.md` — Security best practices
- `docs/architecture/ARCHITECTURE.md` — System design
- `docs/api/API_DOCUMENTATION.md` — REST API reference
- `docs/guides/TROUBLESHOOTING.md` — Common issues and solutions

## CRITICAL CONSTRAINTS FOR APK BUILDING

### Android APK Build Requirements

1. **MUST Use Containers Builder Environment via Containers Submodule**
   - All Android APK builds MUST use the `catalogizer-builder` container
   - Use the `Containers` submodule for container orchestration
   - Builder defined in `docker-compose.build.yml`
   - NEVER build APKs directly on host without container

2. **Builder Container Definition (docker-compose.build.yml)**
   - Service: `catalogizer-builder`
   - Image: `localhost/catalogizer-builder:latest`
   - Android SDK at `/opt/android-sdk`
   - Java at `/usr/lib/jvm/java-21-openjdk-amd64`
   - Depends on: PostgreSQL and Redis (for integration tests)
   - Mounts project to `/project`

3. **Build Using Containers Submodule Boot**
   ```bash
   # Start builder infrastructure (PostgreSQL, Redis, builder)
   cd Containers && ./bin/boot --project /path/to/catalogizer
   
   # Or use docker-compose.build.yml directly:
   podman-compose -f docker-compose.build.yml up --build --abort-on-container-exit
   ```

4. **Direct Builder Container Usage (if boot not available)**
   ```bash
   podman run --rm --entrypoint="" \
     -v /path/to/project:/project \
     -w /project/catalogizer-androidtv \
     -e ANDROID_HOME=/opt/android-sdk \
     -e JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64 \
     localhost/catalogizer-builder:latest \
     ./gradlew assembleDebug --no-daemon
   ```

5. **Build Prerequisites**
   - All git submodules must be initialized
   - Builder container image must exist: `localhost/catalogizer-builder:latest`
   - If missing, build it: `podman build -f docker/Dockerfile.builder -t catalogizer-builder:latest .`

### QA Testing Requirements

1. **Catalog Population**
   - QA tests require populated catalog database
   - Run populate challenge or configure SMB storage before QA
   - Empty catalog shows "Your Library is Empty" (correct behavior)

2. **Android TV QA**
   - Requires ADB connected device
   - App must be installed: `com.catalogizer.androidtv`
   - API must be accessible from device network
   - Video recording available via `adb shell screenrecord`

3. **HelixQA Bank Format**
   - HelixQA requires JSON format for test banks
   - Convert YAML to JSON: `python3 -c "import yaml,json; json.dump(yaml.safe_load(open('bank.yaml')), open('bank.json','w'))"`

## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**


---

### 🔴 CRITICAL: HelixQA UI Testing Exclusivity

**ALL automated UI/UX testing MUST be performed exclusively by HelixQA.**

#### What is FORBIDDEN

- Writing custom ADB tap/click scripts
- Using coordinate-based automation
- Creating manual testing workflows
- Using third-party UI testing tools outside HelixQA
- Any form of non-LLM-driven UI automation

#### What is MANDATORY

- HelixQA for ALL automated UI/UX testing
- LLM vision-driven navigation (screenshot → analysis → action)
- Video recording with frame extraction for analysis
- High-quality video (16Mbps, 1920x1080 minimum)
- Frame-by-frame analysis of recorded video material

#### Rationale

Video-based analysis with frame extraction:
1. Captures ALL UI states including transitions
2. No timing issues - continuous recording
3. Frame-accurate analysis capability
4. Better debugging with video evidence
5. Standardized across all platforms

#### Compliance

Any agent not following this mandate will have their changes REJECTED.
This is enforced at the project level and is non-negotiable.


---

### 🔴 CRITICAL: QA Must Detect Screen Stagnation

**NEVER accept "tests passing" when app is stuck on one screen.**

Requirements:
1. Track actual screen state changes
2. Use EXECUTABLE actions (ADB commands), not text descriptions
3. Verify each action produces a visible change
4. Report stagnation immediately
5. Require frame-by-frame video analysis

**A QA system that doesn't recognize stuck screens is WORTHLESS.**

