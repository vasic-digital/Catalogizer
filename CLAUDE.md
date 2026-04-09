# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go 1.25/Gin backend), **catalog-web** (React 18/TS/Vite frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## ⚠️ MANDATORY CONSTRAINTS

### CRITICAL: .devignore Devices - NEVER USE FOR TESTING

**Devices listed in `.devignore` MUST NEVER be used for any testing, QA, or app deployment.**

- **Before any ADB operation**, check `.devignore` for excluded device models
- **ATMOSphere devices are explicitly excluded** - never install or test on them
- **Only use devices NOT matching any pattern in `.devignore`**
- **If no valid devices are connected, abort testing** - do not proceed with excluded devices
- **This constraint is NON-NEGOTIABLE and applies to all QA, HelixQA, and manual testing**

### CRITICAL: HelixQA ONLY for UI/UX Automated Testing

**ALL automated UI/UX testing MUST be performed exclusively by HelixQA.**

- **NEVER** use custom scripts, manual ADB commands, or third-party tools for UI/UX testing
- **HelixQA is the SOLE authorized tool** for automated Android TV, Web, and Desktop UI testing
- **All UI interactions** (clicks, navigation, screenshots, validation) MUST be LLM-driven via HelixQA
- **No hardcoded coordinates, sleep timers, or tap sequences** - HelixQA vision models decide every action
- **If HelixQA reports an issue**, it must be fixed in the app code, not worked around with scripts
- **This constraint applies to all platforms**: Android TV, Android mobile, Web, and Desktop

### CRITICAL: Real-Time Log Monitoring (MANDATORY)

**ALL QA sessions MUST monitor application and system logs in REAL-TIME.**

- **Android/Android TV**: `adb logcat` must be actively monitored during testing
- **Web**: Browser console logs must be captured and monitored
- **Desktop**: Application and system logs must be monitored
- **Backend Services**: Service logs must be actively watched
- **Purpose**: Immediate detection of ANRs, crashes, fatal exceptions, memory issues
- **HelixQA must implement real-time log streaming and analysis** - not post-session
- **ANR/Crash detection must pause the session** and alert immediately
- **NO QA session is valid without real-time log monitoring**

### CRITICAL: Universal Solution Principle (MANDATORY)

**ALL fixes, workarounds, and testing infrastructure MUST be UNIVERSAL - working with ANY application, not just Catalogizer.**

- **NEVER add test-only code to the application under test** (no `QAInputReceiver`, no test endpoints, no bypasses in app code)
- **NEVER modify the target application to make it "testable"**
- **ALWAYS implement fixes in the testing tool/infrastructure** itself
- **HelixQA must handle text input via on-screen keyboard navigation** - not by modifying the app
- **If detection fails** → Improve detection algorithms in HelixQA, not in the app
- **The target application should require ZERO modifications for testing**
- **Universal solutions ensure portability, maintainability, and valid test results**
- **ANY solution that modifies the app under test is INVALID and must be reimplemented**

### CRITICAL: Device Auto-Connect via .devconnect

**Devices listed in `.devconnect` MUST be auto-connected before HelixQA executes.**

- **`.devconnect`** file contains IP addresses of Android TV devices that must be connected
- **Opposite of `.devignore`** - ensures devices ARE connected rather than excluded
- **Format**: One IP per line (e.g., `192.168.0.214` or `192.168.0.214:5555`)
- **Git ignored**: `.devconnect` is in `.gitignore` (local device IPs should not be committed)
- **Pre-flight check**: Run `./scripts/devconnect.sh` before HelixQA to validate and connect devices
- **Reachability validation**: Script pings devices before attempting ADB connect
- **Idempotent**: Safe to run multiple times - skips already connected devices

Usage:
```bash
# Create .devconnect with your Android TV devices
echo "192.168.0.214" > .devconnect

# Auto-connect all listed devices (validates reachability first)
./scripts/devconnect.sh

# Then run HelixQA
./HelixQA/bin/helixqa autonomous -platforms android ...
```

Example check:
```bash
# Get device model and check against .devignore
DEVICE_MODEL=$(adb -s $DEVICE shell getprop ro.product.model)
if grep -qi "$DEVICE_MODEL" .devignore; then
  echo "❌ Device $DEVICE_MODEL is in .devignore - CANNOT USE"
  exit 1
fi
```

## Submodule Architecture

41 independent git submodules under the vasic-digital organization. Each has its own repo (GitHub + GitLab), tests, docs, and Upstreams for multi-remote push.

### Go Modules (all wired via `replace` directives in `catalog-api/go.mod`)

| Module | Path | Description |
|--------|------|-------------|
| `digital.vasic.challenges` | `Challenges/` | Challenge framework: define, register, run, and report on structured test scenarios |
| `digital.vasic.assets` | `Assets/` | Asset management (lazy loading, serving, defaults) |
| `digital.vasic.containers` | `Containers/` | Container discovery and service port detection |
| `digital.vasic.concurrency` | `Concurrency/` | Concurrency utilities |
| `digital.vasic.config` | `Config/` | Configuration management |
| `digital.vasic.filesystem` | `Filesystem/` | Unified filesystem protocol abstraction |
| `digital.vasic.auth` | `Auth/` | Authentication primitives |
| `digital.vasic.cache` | `Cache/` | Caching layer |
| `digital.vasic.entities` | `Entities/` | Entity model definitions |
| `digital.vasic.eventbus` | `EventBus/` | Event bus for pub/sub |
| `digital.vasic.database` | `Database/` | Database connection, dialect abstraction, migrations |
| `digital.vasic.discovery` | `Discovery/` | Service discovery |
| `digital.vasic.lazy` | `Lazy/` | Generic lazy loading |
| `digital.vasic.media` | `Media/` | Media detection and analysis pipeline |
| `digital.vasic.memory` | `Memory/` | Memory leak detection |
| `digital.vasic.middleware` | `Middleware/` | HTTP middleware components |
| `digital.vasic.observability` | `Observability/` | Metrics, logging, tracing |
| `digital.vasic.ratelimiter` | `RateLimiter/` | Rate limiting |
| `digital.vasic.recovery` | `Recovery/` | Circuit breaker and recovery patterns |
| `digital.vasic.security` | `Security/` | Security utilities |
| `digital.vasic.storage` | `Storage/` | Storage abstraction layer |
| `digital.vasic.streaming` | `Streaming/` | Media streaming |
| `digital.vasic.watcher` | `Watcher/` | File system watcher |

### TypeScript/React Modules (linked via `file:../` in `catalog-web/package.json`)

| Module | Path | Description |
|--------|------|-------------|
| `@vasic-digital/websocket-client` | `WebSocket-Client-TS/` | WebSocket client with reconnection + React hooks |
| `@vasic-digital/ui-components` | `UI-Components-React/` | React UI component library |
| `@vasic-digital/media-types` | `Media-Types-TS/` | Shared media type definitions |
| `@vasic-digital/catalogizer-api-client` | `Catalogizer-API-Client-TS/` | TypeScript API client |
| `@vasic-digital/auth-context` | `Auth-Context-React/` | Auth context provider |
| `@vasic-digital/media-browser` | `Media-Browser-React/` | Media browsing components |
| `@vasic-digital/media-player` | `Media-Player-React/` | Media playback components |
| `@vasic-digital/collection-manager` | `Collection-Manager-React/` | Collection management UI |
| `@vasic-digital/dashboard-analytics` | `Dashboard-Analytics-React/` | Dashboard and analytics |

### HelixQA / AI Submodules

| Module | Path | Description |
|--------|------|-------------|
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
git submodule init && git submodule update --recursive   # after cloning
./scripts/setup-submodule.sh ModuleName [--create-repos] [--go|--ts|--kotlin]
cd SubmoduleName && commit "message"        # push to all upstreams
cd SubmoduleName && install_upstreams       # install upstream remotes
```

## Commands

```bash
# Backend (catalog-api)
cd catalog-api
go run main.go                              # dev server (dynamic port, writes .service-port)
go build -o catalog-api                     # build binary
GOMAXPROCS=3 go test ./... -p 2 -parallel 2 # all tests (resource-limited)
go test -v -run TestName ./path/to/pkg/     # single test

# Frontend (catalog-web) — port 3000, proxies /api to catalog-api
cd catalog-web
npm run dev                                 # dev server (:3000)
npm run test                                # tests (vitest, single run)
npm run test:watch                          # tests (watch mode)
npm run test:coverage                       # tests with coverage
npm run test:e2e                            # Playwright E2E tests
npm run build                               # production build (tsc + vite)
npm run lint && npm run type-check          # lint + typecheck

# Desktop / Installer Wizard
cd catalogizer-desktop   # or installer-wizard
npm run tauri:dev                           # dev
npm run tauri:build                         # build

# API Client
cd catalogizer-api-client
npm run build && npm run test

# Android
cd catalogizer-android   # or catalogizer-androidtv
./gradlew test                              # unit tests
./gradlew assembleDebug                     # debug build

# Full system
podman-compose -f docker-compose.dev.yml up # dev env
./scripts/services-up.sh                    # start all services
./scripts/services-down.sh                  # stop all services
./scripts/run-all-tests.sh                  # all tests + security

# Release build (containerized, all 7 components)
./scripts/release-build.sh --container --force --skip-tests

# SonarQube code quality scan
./scripts/run-sonarqube-scan.sh
```

Test helper in `catalog-api/internal/tests/test_helper.go` provides SQLite test database setup via `database.WrapDB()`.

## Architecture

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

### Database Layer

Dual-dialect abstraction supporting SQLite (dev) and PostgreSQL (production).

- `database/dialect.go`: `DialectType` enum (DialectSQLite | DialectPostgres) with query rewriting:
  - `RewritePlaceholders()` — `?` → `$1, $2, ...` for PostgreSQL
  - `RewriteInsertOrIgnore()` — `INSERT OR IGNORE` → `ON CONFLICT DO NOTHING`
  - `BooleanLiterals()` — `= 0/1` → `= FALSE/TRUE` for known boolean columns
- `database.DB` wraps `*sql.DB` with shadowed `Exec()`, `Query()`, `QueryRow()` that auto-rewrite SQL.
- `InsertReturningID()` and `TxInsertReturningID()` replace `LastInsertId()` (PostgreSQL uses `RETURNING id`).
- `database.WrapDB(sqlDB, DialectSQLite)` for unit tests (in-memory SQLite).
- Migrations in `database/migrations/` — separate SQLite and PostgreSQL variants per migration.
- SQLCipher support imported for encrypted SQLite.
- **SQLite WAL mode**: Explicit `PRAGMA journal_mode=WAL` after connection in `database/connection.go` — go-sqlcipher ignores connection string pragmas.

### catalog-web (React/TypeScript/Vite)

AuthProvider → WebSocketProvider → Router. Key tech: React Query (`@tanstack/react-query`) for server state, Zustand for client state, Tailwind CSS for styling, React Hook Form + Zod for forms, framer-motion for animations, Vitest for unit tests, Playwright for E2E tests.

- Auth-gated routes via `ProtectedRoute`.
- Path aliases configured in `vite.config.ts`: `@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, `@/assets`.
- API proxy: reads `../catalog-api/.service-port` at dev server startup to resolve backend port (falls back to 8080).
- Build output split into vendor chunks: `vendor` (react), `router`, `ui`, `charts`, `utils`.

### Other Components

**Android**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit. Hilt DI. Requires `jvmToolchain(17)` and `--add-opens` JVM args for kapt + JDK 21 compatibility.

**Android TV Home Screen Channels** (catalogizer-androidtv v2.3.0): Full integration with Android TV's channel API (`androidx.tvprovider`). Default "Catalogizer Picks" channel auto-created on launch. Dynamic per-category channels (one per media type with content). System Watch Next row for partially-watched items + auto-next-episode. Deep linking via `catalogizer://media/{id}?type={type}` with per-category launch behavior (detail screen vs. immediate play, configurable in Settings). `WorkManager` periodic sync (6h) + app-launch + SyncService triggers. Full cleanup on logout. Key files: `data/tv/TvChannelRepository.kt`, `data/tv/ChannelProgramMapper.kt`, `data/tv/WatchNextManager.kt`, `data/tv/TvChannelSyncWorker.kt`, `ui/ChannelDeepLinkActivity.kt`.

**Tauri apps**: React frontend ↔ Rust backend via IPC commands/events.

### Build Framework

`Build/` submodule provides a generic, reusable shell-based build framework. `scripts/release-build.sh` orchestrates all 7 components using per-component builders in `scripts/lib/build-*.sh`.

```bash
# Source the framework in build scripts
source Build/lib/common.sh      # logging, container runtime detection, git helpers
source Build/lib/version.sh     # semantic versioning via versions.json
source Build/lib/hash.sh        # SHA256 change detection (skip unchanged components)
source Build/lib/orchestrator.sh # CLI parsing, build loop
```

Projects must define `BUILD_COMPONENTS`, `BUILD_COMPONENT_PATTERNS`, and `build_single_component()`.

### Challenge System

`digital.vasic.challenges` framework integrated via `Challenges/` submodule. Challenges are Go structs embedding `challenge.BaseChallenge` with custom `Execute()`. Registered in `catalog-api/challenges/register.go` via `RegisterAll()`, exposed via `/api/v1/challenges` REST endpoints. Challenge bank definitions loaded from `challenges/config/`.

**All challenge operations MUST be executed exclusively by system deliverables (compiled binaries) — the catalog-api service and other Catalogizer applications. Never use custom scripts, curl commands, or third-party tools to trigger API endpoints within challenge execution. Scanning, storage root creation, and all other operations must go through the running services, exactly as an end user would.**

Key constraints:
- `RunAll` is synchronous/blocking — no other challenge can run until it finishes.
- Progress-based liveness detection: 5-minute stale threshold kills stuck challenges.
- `challenge.NewConfig()` sets Timeout=5min by default — zero it to use runner's timeout.
- `config.json` `write_timeout` must be 900 (not 30) for long-running challenge RunAll.

### User Flow Automation

Multi-platform user flow automation via `Challenges/pkg/userflow/`. 174 Catalogizer-specific challenges in `catalog-api/challenges/userflow_*.go` across 4 platform groups:

| File | Platform | Challenges |
|------|----------|-----------|
| `userflow_api.go` | Go API (HTTP) | 49 |
| `userflow_web.go` | React web (Playwright) | 59 |
| `userflow_desktop.go` | Tauri desktop + wizard | 28 |
| `userflow_mobile.go` | Android + Android TV | 38 |

Registered via `RegisterUserFlowAPIChallenges()`, `RegisterUserFlowWebChallenges()`, `RegisterUserFlowDesktopChallenges()`, `RegisterUserFlowMobileChallenges()` in `register.go`.

CLI runner: `Challenges/cmd/userflow-runner` — flags: `--platform`, `--report`, `--compose`, `--root`, `--timeout`, `--output`.

Container test stack: `docker-compose.test.yml` (catalog-api, catalog-web, playwright; all `network_mode: host`).

## Media Entity System

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

**Entity tables**: media_types, media_items (parent_id self-ref for hierarchy), media_files (junction to files), media_collections, media_collection_items, external_metadata, user_metadata, directory_analyses, detection_rules.

Entity API routes are defined in `handlers/media_entity_handler.go`. Key entity files: `repository/media_item_repository.go` (CRUD, search, hierarchy), `internal/services/aggregation_service.go` (post-scan creation), `internal/services/title_parser.go` (regex parsers).

**Metadata providers**: OpenLibrary (books) and MusicBrainz (music) are fully implemented. TMDB and OMDB provide movie/TV metadata. Other providers (IGDB, GiantBomb, etc.) have graceful degradation — missing API keys or unavailable services do not block the pipeline.

Entity constraints:
- All scanned files MUST be associated with a recognized media entity after aggregation.
- Entity hierarchy: parent_id self-reference (TV Show → seasons → episodes, Music Artist → albums → songs).

## Root Directory Structure (Mandatory Locations)

New files MUST be placed in the correct directory. Do NOT add files to the project root unless they are conventional root files (README, LICENSE, .gitignore, docker-compose, etc.).

| Directory | Purpose |
|---|---|
| `challenges/` | Challenge bank definitions and runtime results |
| `config/` | Infrastructure config files (nginx.conf, redis.conf) |
| `scripts/` | Shell scripts (install, setup, CI/CD, testing runners) |
| `scripts/lib/` | Per-component build scripts (`build-*.sh`) used by release-build.sh |
| `tests/` | Standalone/integration test files |
| `docs/` | All documentation markdown files, organized by subdirectory |
| `Assets/` | Static assets (images, HTML tutorials) — also a Go submodule |
| `Build/` | Generic build framework submodule (versioning, change detection, orchestration) |
| `build/` | Build output and container build context |
| `deployment/` | Deployment configurations |
| `monitoring/` | Monitoring and observability configs |
| `tools/` | Development tooling |
| `Upstreams/` | Git upstream remote configurations for submodules |

Docker Compose files reference `config/` for nginx and redis configs. Do NOT move these config files without updating the Compose volume mounts.

### Docker Compose Files

| File | Purpose |
|---|---|
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

**CRITICAL: API Keys and Secrets — NEVER Commit to Git.** This is a MANDATORY, NON-NEGOTIABLE security rule:
- **Never** commit `.env` files containing real API keys, tokens, or secrets
- **Never** hardcode API keys in source code, CLAUDE.md, AGENTS.md, or any tracked file
- Use `.env.example` with placeholder values only (e.g., `YOUR_API_KEY_HERE`)
- Verify `.gitignore` covers all `.env` files before every commit
- If an API key is accidentally committed, **rotate it immediately**
- All submodules MUST have `.env` in their `.gitignore`
- Pre-commit hooks should scan for secrets when available

**CRITICAL: Git Access via SSH Only — NEVER Use HTTPS.** This is a MANDATORY, NON-NEGOTIABLE security rule:
- **Always** use SSH (`git@github.com:user/repo.git`) for all Git operations — cloning, fetching, pushing
- **Never** use HTTPS (`https://github.com/user/repo.git`) for Git access
- Configure remotes to use SSH: `git remote set-url origin git@github.com:user/repo.git`
- For new clones: `git clone git@github.com:user/repo.git` — NOT `git clone https://...`
- For GitLab, GitFlic, GitVerse, and all other Git hosts: use SSH protocol exclusively
- HTTPS bypasses SSH key-based authentication and is less secure
- Submodules MUST be configured with SSH URLs in `.gitmodules`
- CI/CD scripts and automation MUST use SSH with deploy keys, never HTTPS with passwords/tokens

**CRITICAL: HelixQA — FULLY LLM-DRIVEN Autonomous Testing.**

HelixQA is a generic, universal QA tool driven entirely by LLM vision models. Pipeline: Learn → Plan → Execute → Curiosity → Analyze. Run via `helixqa autonomous --platforms androidtv`. See `HelixQA/.env.example` for configuration.

**Vision architecture**: Phase-specific model selection via LLMsVerifier strategies — **NavigationStrategy** (Execute/Curiosity) for JSON-action models, **AnalysisStrategy** (Analyze) for rich-description models, **PlanningStrategy** (Learn/Plan) for reasoning models. **llama.cpp RPC distributed inference** is the primary local backend; cloud providers (Astica.AI, Gemini, OpenAI) complement it. Models scored dynamically per-phase (no hardcoded preferences). Bridged CLI models discovered via `pkg/bridge/`.

**Non-negotiable rules:**
- ALL navigation MUST be performed by real LLM vision models — the LLM sees a screenshot, decides the next action. Every single step.
- NEVER write hardcoded tap coordinates, sleep timers, keystroke sequences, or fallback scripts. If vision providers are unavailable, the phase MUST skip — not fake results. Malformed LLM JSON → retry the vision call, never substitute hardcoded actions.
- Fix issues in HelixQA Go code (parsing, retry logic, prompts) — never work around with scripts.
- Every connected ADB device MUST be tested (except `.devignore` entries). ADB reverse proxy set up automatically.
- QA priority: (1) Happy paths (login, browse, play), (2) Standard flows, (3) Edge cases, (4) Adversarial.
- Never type credentials into non-login fields — LLM must understand which screen it's on.

**Evidence validation (mandatory):**
- Visually inspect every screenshot to verify expected screen state.
- Login verification: UI dump must NOT contain "Sign In" after login attempt — if it does, login FAILED.
- Data validation: compare API responses against screen content. Empty screens with backend data = BUG.
- Review all video recordings for visual glitches, frozen frames, wrong screens.
- Verify every phase transition before proceeding to the next phase.
- Cross-reference screen content against codebase logic and database state.

**GitHub Actions are PERMANENTLY DISABLED.** Do NOT create any GitHub Actions workflow files in `.github/workflows/`. CI/CD must be run locally.

**All builds, services, and QA testing MUST use containers (Podman).** This is a MANDATORY, NON-NEGOTIABLE rule:
- **Builds**: Use `./scripts/release-build.sh --container` or `podman-compose -f docker-compose.build.yml`. Single-container builds: `podman run --network host`.
- **Services**: Use `podman-compose` to run catalog-api, catalog-web, and all supporting services.
- **QA Testing**: Use `./scripts/run-helixqa.sh` with containerized services. HelixQA sessions run against containers, not bare-metal processes.
- **Android Emulators**: Run in containers via `docker-compose.test.yml --profile android`.
- **Never build or run apps/services directly on bare metal** in production or QA contexts. Local `go run` / `npm run dev` is acceptable only for rapid development iteration.

## Local Development Setup

### Database

**SQLite (Development):** No setup needed — catalog-api creates `catalogizer.db` automatically.

**PostgreSQL (Production):** Set env vars `DB_TYPE=postgres`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`. Container port mapping: 5432→5433.

### Environment Variables

Create `.env` file in `catalog-api/`. Env vars always override `config.json`.

```env
PORT=8080
GIN_MODE=debug
DB_TYPE=sqlite
JWT_SECRET=your-dev-secret-key
ADMIN_PASSWORD=admin123
TMDB_API_KEY=your_tmdb_key     # optional
OMDB_API_KEY=your_omdb_key     # optional
```

### Running the Full Stack

```bash
# Kill anything on port 3000 first (e.g., Bear Messenger)
ss -tlnp | grep :3000

# Terminal 1: Backend (writes .service-port for frontend discovery)
cd catalog-api && go run main.go

# Terminal 2: Frontend (reads .service-port, proxies /api to backend)
cd catalog-web && npm install && npm run dev

# Access: http://localhost:3000 (frontend) / http://localhost:8080 (API)
```

## Zero Warning / Zero Error Policy

All components must run with zero console warnings, zero console errors, and zero failed network requests in every environment.

- No browser console errors or warnings. Every failed network request is a defect.
- Every API endpoint the frontend calls must exist, return valid 2xx responses, and match expected shape.
- No framework deprecation warnings. No WebSocket connection failures.
- If a feature is not yet implemented, provide a stub endpoint that returns a valid empty response.
- The challenge suite (CH-001 to CH-020+) enforces this end-to-end.

## CRITICAL: Host Resource Limits (30-40% Maximum)

The host machine runs other mission-critical processes. All workloads MUST be limited to 30-40% of total host resources. Exceeding this can freeze the system.

- **Go tests**: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- **Container CPU/memory limits** (mandatory): PostgreSQL `--cpus=1 --memory=2g`, API `--cpus=2 --memory=4g`, Web `--cpus=1 --memory=2g`, Builder `--cpus=3 --memory=8g`
- **Total container budget**: max 4 CPUs, 8 GB RAM across all running containers
- **Challenges**: Run sequentially via the API, never in parallel
- **Monitor**: `podman stats --no-stream` and `cat /proc/loadavg`

## CRITICAL: HTTP/3 (QUIC) with Brotli Compression (Mandatory)

All network communication MUST use **HTTP/3 (QUIC)** with **Brotli compression**. Fallback: HTTP/2 + gzip. Never HTTP/1.1 in production.

- **catalog-api**: `quic-go/http3` server + Brotli middleware (`andybalholm/brotli`)
- **catalog-web**: Served via HTTP/3-capable reverse proxy, Brotli-compressed static assets
- **Tauri apps**: HTTP/3 client for API communication
- **Android apps**: OkHttp with HTTP/3 (Cronet) + Brotli
- **API client**: HTTP/3-capable fetch with Brotli Accept-Encoding

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

## Concurrency & Lifecycle Patterns

- **CacheService**: Spawns cleanup goroutine in `NewCacheService()`. Tests MUST call `defer service.Close()`. Uses `sync.Once` for safe double-close.
- **WebSocketHandler**: Spawns cleanup goroutine in constructor. Uses `sync.Once` for safe `Stop()`. Tests must call `handler.Stop()` before `server.Close()` to unblock `readPump`.
- **Production shutdown**: `main.go` shutdown calls `wsHandler.Stop()` and `cacheService.Close()` before HTTP server shutdown.
- **Database pool**: Connection pool defaults: MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=3m. Overridable via config.
- **Race safety**: `connCount` reads in WebSocketHandler protected by mutex. `SyncService.StartSync()` and `LogManagementService.CollectLogs()` return copies to prevent shared-pointer races.

## CRITICAL: Mandatory Video Recording for Device/Emulator QA

**All HelixQA sessions involving Android devices or emulators MUST include video recording.** This is a MANDATORY, NON-NEGOTIABLE rule for enterprise-grade quality assurance.

### Recording Methods
- **Android 9 and below** (Mi Box, emulators): `adb shell screenrecord --bit-rate 4000000 /sdcard/qa_session.mp4`
- **Android 10+** (phones): Use rapid screenshot capture (`adb shell screencap`) assembled into video via ffmpeg
- **Web sessions**: Playwright `--video on` or ffmpeg x11grab when X display available
- **Desktop (Tauri)**: ffmpeg x11grab or Xvfb-based recording

### Post-Recording Analysis (MANDATORY)
Every recorded video and screenshot sequence MUST be deeply analyzed for:
- **Visual inconsistencies**: misaligned elements, clipped text, wrong colors, missing assets
- **UI/UX issues**: unresponsive buttons, incorrect focus states, broken animations
- **Content issues**: empty screens that should have data, placeholder text in production
- **Brand compliance**: Vasic Digital logo must display in rounded square with red border
- **Performance issues**: visible jank, slow transitions, loading states lasting too long
- **Crash indicators**: black screens, frozen frames, unexpected app restarts

### Recording Output
- Videos stored in `qa-results/video-sessions-<timestamp>/`
- Screenshot sequences stored in `qa-results/video-sessions-<timestamp>/<device>-frames/`
- Analysis reports stored alongside videos as markdown files

### Device-Specific Notes
- Android 15 (SDK 35): `screenrecord` fails with `Encoder failed (err=-38)` from ADB — use screenshot-to-video approach
- Mi Box (Android 9): Native `screenrecord` works, use `--bit-rate 4000000 --time-limit 120`
- Always pull videos from device after recording: `adb pull /sdcard/qa_session.mp4`

## Load Testing

k6 test scripts in `tests/k6/`:
- `load_test.js` — Ramp to 50 users, verify p95 < 500ms
- `stress_test.js` — Ramp to 300 users, find breaking point
- `soak_test.js` — 20 users for 30 minutes, detect memory leaks

Run via: `podman run --rm --network host -v $(pwd)/tests/k6:/scripts docker.io/grafana/k6:latest run /scripts/load_test.js`

## Security Scanning

Run `./scripts/security-scan.sh` for automated scanning. Run `./scripts/run-sonarqube-scan.sh` for SonarQube code quality analysis. Available tools:
- `govulncheck` — Go stdlib/dependency vulnerabilities
- `npm audit` — Frontend dependency vulnerabilities
- Semgrep — Static analysis for security anti-patterns: `podman-compose -f docker-compose.security.yml --profile semgrep-scan run --rm semgrep-scanner`
- Snyk, Trivy, Gosec via `docker-compose.security.yml`

## Conventions

**Config precedence**: env vars > `.env` > `config.json` > defaults

**PostCSS**: `postcss.config.js` must use `module.exports` (CommonJS) for Node 18 compat

### Go Backend Style

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
- **Testing**: Table-driven tests with `t.Run`. Use `testify/suite` for complex suites, `testify/mock` for mocks. Files: `*_test.go` beside source. Use `database.WrapDB()` for in-memory SQLite test DB.
- **Concurrency**: Services spawning goroutines (`CacheService`, `WebSocketHandler`) use `sync.Once` for cleanup. Tests MUST `defer service.Close()` / `handler.Stop()`.
- **Database**: Use `?` placeholders (auto-converted to `$1, $2...` for Postgres). Use `InsertReturningID()` instead of `LastInsertId()`.

### TypeScript/React Frontend Style

- **Naming**: PascalCase components/interfaces, camelCase functions/variables, SCREAMING_SNAKE_CASE constants.
- **Imports**: Three groups — React, third-party, local path aliases (`@/components`, `@/hooks`, `@/lib`, etc.).
- **Formatting**: Prettier. Tailwind classes composed via `cn()` from `@/lib/utils`.
- **Linting**: ESLint with `@typescript-eslint`. `--max-warnings 0` enforced.
- **State**: React Query for server state, Zustand for client state.
- **Forms**: React Hook Form + Zod validation.
- **Testing**: Vitest + React Testing Library. Playwright for E2E.

### Kotlin/Android Style

- **Architecture**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit.
- **DI**: Hilt for dependency injection.
- **Async**: `suspend` functions, `Flow`/`StateFlow`, Paging 3.
- **Error handling**: Sealed `Result` classes for operation outcomes.
- **Testing**: JUnit 4 + MockK/Mockito. Coroutines via `kotlinx-coroutines-test`.
- **Build**: JDK 21 with `--add-opens` JVM args for kapt compatibility.

## Pre-Commit Checklist

```bash
cd catalog-api && go fmt ./... && go vet ./...           # Go format + lint
cd catalog-web && npm run lint && npm run type-check     # TS lint + typecheck
pre-commit run --all-files                               # run all hooks
```

Ensure zero console warnings/errors in browser. Verify `.gitignore` covers `.env`.

## Key Files

- `catalog-api/main.go` — API entry point, route registration
- `catalog-api/database/dialect.go` — dual-dialect SQL rewriting
- `catalog-api/filesystem/interface.go` — `UnifiedClient` protocol abstraction
- `catalog-api/challenges/register.go` — challenge registration
- `catalog-web/src/App.tsx` — React root (AuthProvider → WebSocketProvider → Router)
- `catalog-web/vite.config.ts` — path aliases, API proxy config
- `versions.json` — version tracking for all components
- `.env.example` — environment variable template

## CRITICAL: Iterative Test-Fix-Rebuild QA Loop (Mandatory)

**All QA campaigns MUST follow the iterative test-fix-rebuild loop.** This is a MANDATORY, NON-NEGOTIABLE process:

### Loop Protocol
1. **Rebuild** all affected binaries, containers, and deployments before each test iteration
2. **Execute** all tests (unit, challenges, HelixQA bank, autonomous QA)
3. **Analyze** all results, video recordings, screenshots, and logs
4. **Create tickets** for every defect with severity, evidence, reproduction steps
5. **Fix** root causes (never workarounds), create validation tests for each fix
6. **Repeat** from step 1 until: all pass, fatal blocker, or nothing left to fix

### Exit Conditions (only these stop the loop)
- All tests pass across all platforms and all test types
- FATAL BLOCKER encountered (system crash, hardware failure, unrecoverable state)
- Nothing left to test, fix, or polish

### Validation Tests (Fixes Validation Suite)
- Every bug fix MUST include a corresponding test in the "Fixes Validation" bank
- Tests must verify the specific fix prevents the issue from recurring
- Tests are permanent — they persist in the bank system across all future QA campaigns

## CRITICAL: Live Monitoring During All Test Execution (Mandatory)

**All test execution MUST be actively monitored with real-time status reporting.** This is NON-NEGOTIABLE:

- **Platform identification**: Which platform and app/service is being tested
- **Test case tracking**: Current test case ID, short description, progress percentage
- **Per-test results**: Running/pass/fail/skip status with duration
- **Aggregate stats**: Total tests, passed, failed, skipped, warnings — updated in real-time
- **Complete logging**: All output captured to `docs/reports/qa-sessions/qa-session-<date>/logs/`
- **Session archival**: Every QA session produces a complete archive including:
  - Test execution logs (stdout/stderr)
  - Video recordings and screenshots
  - Challenge results (JSON)
  - HelixQA pipeline reports
  - Ticket/issue files
  - Deep analysis and conclusions

### Reporting Directory Structure
```
docs/reports/qa-sessions/qa-session-YYYY-MM-DD/
├── FINAL-REPORT.md              # Executive summary with all results
├── logs/                        # Complete execution logs
│   ├── unit-tests-go.log
│   ├── unit-tests-frontend.log
│   ├── challenges.log
│   ├── helixqa-bank.log
│   └── helixqa-autonomous-<platform>.log
├── challenges/                  # Challenge results (JSON + summary)
├── helixqa/                     # HelixQA reports and evidence
│   ├── bank-results/
│   └── autonomous/
├── videos/                      # All video recordings
├── screenshots/                 # All screenshots
├── tickets/                     # Issue tickets with evidence
└── analysis/                    # Deep analysis documents
```

## CRITICAL: Comprehensive HelixQA Test Coverage (Mandatory)

**HelixQA test banks MUST cover ALL features, ALL screens, ALL use cases with varied data sets.** No feature or functionality may be left untested.

### Required Coverage per Platform
- **All happy paths** with known data from the system (real titles, music, books)
- **All screens/pages/views** with full UI element validation
- **All CRUD operations** for each entity type (media items, collections, playlists, users)
- **All media types** (movie, tv_show, music, book, comic, game, software — all 11 types)
- **Search** with real content terms, empty queries, long text, special characters, Cyrillic
- **Authentication** flows (login, logout, session expiry, invalid credentials)
- **Navigation** (forward, back, deep linking, TV channels, DPAD)
- **Media interaction** (play video, listen to music, read books, view images, browse comics)
- **Settings and configuration** (all toggleable options, theme, preferences)
- **Edge cases** (boundary values, rapid actions, network interruption)
- **Negative data** (SQL injection strings, XSS payloads, malformed input, wrong data types)

### Data Set Requirements
Test banks MUST include specific data sets drawn from:
- Actual catalog content (titles scanned from NAS)
- Known media metadata (TMDB, OpenLibrary, MusicBrainz entries)
- Invalid/malformed data for negative testing
- Boundary values (max-length strings, zero, negative, overflow)
- Internationalized content (Cyrillic paths, Unicode characters)

### Bank Organization
- `banks/full-qa-api.yaml` — Comprehensive API endpoint testing
- `banks/full-qa-web.yaml` — Comprehensive web UI testing
- `banks/full-qa-androidtv.yaml` — Comprehensive Android TV testing
- `banks/full-qa-android.yaml` — Comprehensive Android phone testing
- `banks/full-qa-cross-platform.yaml` — Cross-platform consistency
- `banks/fixes-validation.yaml` — Regression tests for all bug fixes

**HelixQA bank format**: JSON required. Convert YAML: `python3 -c "import yaml,json; json.dump(yaml.safe_load(open('bank.yaml')), open('bank.json','w'))"`

## CRITICAL: Android APK Build Requirements

All Android APK builds MUST use the `catalogizer-builder` container. Never build APKs directly on host without container.

```bash
# Start builder infrastructure (PostgreSQL, Redis, builder)
cd Containers && ./bin/boot --project /path/to/catalogizer

# Or use docker-compose.build.yml directly:
podman-compose -f docker-compose.build.yml up --build --abort-on-container-exit

# Direct builder container usage:
podman run --rm --entrypoint="" \
  -v /path/to/project:/project \
  -w /project/catalogizer-androidtv \
  -e ANDROID_HOME=/opt/android-sdk \
  -e JAVA_HOME=/usr/lib/jvm/java-21-openjdk-amd64 \
  localhost/catalogizer-builder:latest \
  ./gradlew assembleDebug --no-daemon
```

Builder image must exist: `localhost/catalogizer-builder:latest`. If missing: `podman build -f docker/Dockerfile.builder -t catalogizer-builder:latest .`

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


### One-Command HelixQA Execution

**Use the orchestrator script for full automation:**

```bash
# One command to rule them all - connects devices, installs APKs, runs HelixQA
./scripts/helixqa-orchestrator.sh [platforms]

# Examples:
./scripts/helixqa-orchestrator.sh           # Test all platforms
./scripts/helixqa-orchestrator.sh android   # Android TV only
./scripts/helixqa-orchestrator.sh web       # Web only
```

**Orchestrator Phases:**
1. ✅ Environment validation (API health check)
2. ✅ Device connection (.devconnect auto-connect)
3. ✅ APK installation (builds if needed)
4. ✅ Monitoring (background health monitoring)
5. ✅ HelixQA execution (autonomous testing)
6. ✅ Report generation (consolidated results)

**Output:** `qa-results/session-<timestamp>/`

---

## 🔴 CRITICAL: HelixQA EXCLUSIVITY MANDATE

### UI/UX Testing - HelixQA ONLY

**Effective Immediately - ALL automated UI/UX testing MUST be performed exclusively by HelixQA.**

#### Absolute Rules

1. **NO MANUAL UI TESTING SCRIPTS**
   - ❌ Never write custom ADB tap sequences
   - ❌ Never use coordinate-based automation
   - ❌ Never create shell scripts for UI interactions
   - ✅ ALWAYS use HelixQA for automated UI testing

2. **NO THIRD-PARTY UI TESTING TOOLS**
   - ❌ No Appium scripts outside HelixQA
   - ❌ No Playwright scripts outside HelixQA
   - ❌ No custom UI automation frameworks
   - ✅ HelixQA is the SOLE authorized UI testing tool

3. **VISION-DRIVEN TESTING MANDATORY**
   - ALL navigation MUST be LLM vision-driven
   - Screenshot → LLM analysis → Action decision
   - Every single step must use vision models
   - NO hardcoded coordinates or sleep timers

#### Implementation Requirements

1. **Video-Based Analysis**
   - All testing MUST record high-quality video
   - Frames extracted from video for analysis
   - 16Mbps bitrate minimum for Android recording
   - 1920x1080 resolution for frame extraction

2. **Screenshot Replacement Protocol**
   - Screenshots replaced with video frame extraction
   - Use `ExtractFrameAt()` for specific timestamps
   - Use `ExtractLatestFrame()` for current state
   - Higher quality and reliability than direct screenshots

3. **Timing Issue Resolution**
   - Video captures ALL frames continuously
   - No missed frames due to timing issues
   - Extract frames after UI has rendered
   - Frame-by-frame analysis capability

#### Consequences of Non-Compliance

- Code violating this mandate will be REJECTED
- Manual testing scripts will be DELETED
- Only HelixQA-based solutions are acceptable
- This is a ZERO TOLERANCE policy

#### Verification Checklist

Before submitting any UI testing solution:
- [ ] Is HelixQA the ONLY testing tool used?
- [ ] Are video recordings used for analysis?
- [ ] Is every action LLM vision-driven?
- [ ] Are there NO hardcoded coordinates?
- [ ] Are there NO manual adb tap sequences?

**This mandate is ABSOLUTE and NON-NEGOTIABLE.**


---

## 🔴 CRITICAL: QA System Requirements - SCREEN RECOGNITION & ACTION

### The Problem
A QA system that cannot recognize when the app is stuck on the same screen is **USELESS**.

Example of FAILURE:
- Tests "pass" because they check "does login screen exist?" 
- App is stuck on login for ALL tests
- Never actually logs in
- Never tests home screen, media browsing, playback
- QA reports "12/12 tests PASSED" but 0 actual functionality tested

### Mandate: QA Must Recognize and Report Screen Stagnation

#### 1. Screen State Tracking
HelixQA MUST:
- Track which screen is currently displayed
- Compare current screen to previous screen
- Detect if screen hasn't changed after an action
- Report STAGNATION as a critical issue

#### 2. Action Verification
HelixQA MUST:
- Actually EXECUTE actions (ADB commands, not just descriptions)
- Verify the action had an effect
- Confirm screen state changed as expected
- FAIL the test if no change occurred

#### 3. Test Bank Requirements
ALL test banks MUST use EXECUTABLE actions:

❌ WRONG (text description):
```json
{
  "action": "Enter admin/admin123 credentials",
  "expected": "Home screen loads"
}
```

✅ CORRECT (executable command):
```yaml
- name: Type username
  action: "adb_shell: input text admin"
  expected: "Username field populated"
  
- name: Click login
  action: "adb_shell: input keyevent KEYCODE_ENTER"
  expected: "Screen changes from login to home"
```

#### 4. Frame-by-Frame Video Analysis
For EVERY test:
1. Record video at 16Mbps, 1920x1080
2. Extract frames at 1-second intervals
3. Compare frame N to frame N+1
4. If frames are identical for >5 seconds after an action → REPORT STAGNATION
5. If app stays on login screen for entire test → REPORT BLOCKER ISSUE

#### 5. Critical Issues to Auto-Report
- App stuck on same screen for >10 seconds
- Login screen visible but login never attempted
- Home screen never reached after login action
- Blank/black screens
- ANR/Crash detection

### Consequence of Non-Compliance
QA results showing "100% pass" when app never progressed past login screen are **FRAUDULENT** and **UNACCEPTABLE**.

**All test banks must be re-written with EXECUTABLE actions.**
**All QA results must show ACTUAL screen progression, not just screen existence.**

