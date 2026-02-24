# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go 1.24/Gin backend), **catalog-web** (React 18/TS/Vite frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## Submodule Architecture

29 independent git submodules under the vasic-digital organization. Each has its own repo (GitHub + GitLab), tests, docs, and Upstreams for multi-remote push.

### Go Modules (used via `replace` directives in `catalog-api/go.mod`)

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

Additional Go submodules (Database, Discovery, Media, Middleware, Observability, RateLimiter, Security, Storage, Streaming, Watcher) exist but are not currently wired via `replace` directives.

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
go test ./...                               # all tests
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
```

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

### catalog-web (React/TypeScript/Vite)

AuthProvider → WebSocketProvider → Router. Key tech: React Query (`@tanstack/react-query`) for server state, Zustand for client state, Tailwind CSS for styling, React Hook Form + Zod for forms, framer-motion for animations, Vitest for unit tests, Playwright for E2E tests.

- Auth-gated routes via `ProtectedRoute`.
- Path aliases configured in `vite.config.ts`: `@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, `@/assets`.
- API proxy: reads `../catalog-api/.service-port` at dev server startup to resolve backend port (falls back to 8080).
- Build output split into vendor chunks: `vendor` (react), `router`, `ui`, `charts`, `utils`.

### Other Components

**Android**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit. Hilt DI.

**Tauri apps**: React frontend ↔ Rust backend via IPC commands/events.

### Challenge System

`digital.vasic.challenges` framework integrated via `Challenges/` submodule. Challenges are Go structs embedding `challenge.BaseChallenge` with custom `Execute()`. Registered in `catalog-api/challenges/register.go` via `RegisterAll()`, exposed via `/api/v1/challenges` REST endpoints. Challenge bank definitions loaded from `challenges/config/`.

**All challenge operations MUST be executed exclusively by system deliverables (compiled binaries) — the catalog-api service and other Catalogizer applications. Never use custom scripts, curl commands, or third-party tools to trigger API endpoints within challenge execution. Scanning, storage root creation, and all other operations must go through the running services, exactly as an end user would.**

Key constraints:
- `RunAll` is synchronous/blocking — no other challenge can run until it finishes.
- Progress-based liveness detection: 5-minute stale threshold kills stuck challenges.
- `challenge.NewConfig()` sets Timeout=5min by default — zero it to use runner's timeout.

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
| `tests/` | Standalone/integration test files |
| `docs/` | All documentation markdown files, organized by subdirectory |
| `Assets/` | Static assets (images, HTML tutorials) — also a Go submodule |

Docker Compose files reference `config/` for nginx and redis configs. Do NOT move these config files without updating the Compose volume mounts.

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

## Constraints

**GitHub Actions are PERMANENTLY DISABLED.** Do NOT create any GitHub Actions workflow files in `.github/workflows/`. CI/CD must be run locally.

**All builds and services MUST use containers.** Use the containerized build pipeline (`./scripts/container-build.sh` or `podman-compose -f docker-compose.build.yml`) for builds, and `podman-compose` / `podman run` for running services. Use `podman run --network host` for single-container builds.

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
# Terminal 1: Backend (writes .service-port for frontend discovery)
cd catalog-api && go run main.go

# Terminal 2: Frontend (reads .service-port, proxies /api to backend)
cd catalog-web && npm install && npm run dev

# Access: http://localhost:3000 (frontend) / http://localhost:8080 (API)
```

## Testing

```bash
# Go tests (resource-limited)
cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# Go single test
cd catalog-api && go test -v -run TestFunctionName ./path/to/package/

# Web unit tests (Vitest)
cd catalog-web && npm run test           # single run
cd catalog-web && npm run test:watch     # watch mode
cd catalog-web && npm run test:coverage  # with coverage

# Web E2E tests (Playwright)
cd catalog-web && npm run test:e2e

# All tests
./scripts/run-all-tests.sh
```

Test helper in `catalog-api/internal/tests/test_helper.go` provides SQLite test database setup via `database.WrapDB()`.

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

## Conventions

- **Go**: `NewService` constructor injection, error wrapping, table-driven tests, `*_test.go` beside source
- **TypeScript**: PascalCase components, camelCase functions, Zod validation, React Hook Form
- **Kotlin**: MVVM, Result sealed classes, Room for offline
- **Config precedence**: env vars > `.env` > `config.json` > defaults
- **PostCSS**: `postcss.config.js` must use `module.exports` (CommonJS) for Node 18 compat
