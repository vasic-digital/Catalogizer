# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go/Gin backend), **catalog-web** (React/TS frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## Submodule Architecture

Reusable functionality is extracted into independent git submodules under the vasic-digital organization. Each module has its own repo (GitHub + GitLab), tests, docs, and Upstreams for multi-remote push.

### Go Modules

| Module | Path | Description |
|--------|------|-------------|
| `digital.vasic.challenges` | `Challenges/` | Challenge framework: define, register, run, and report on structured test scenarios. Includes progress-based liveness detection (stale threshold) for long-running challenges |

### TypeScript/React Modules

| Module | Path | Description |
|--------|------|-------------|
| `@vasic-digital/websocket-client` | `WebSocket-Client-TS/` | WebSocket client with reconnection + React hooks |
| `@vasic-digital/ui-components` | `UI-Components-React/` | React UI component library (Button, Card, Input, etc.) |

### Submodule Commands

```bash
# Initialize all submodules after cloning
git submodule init && git submodule update --recursive

# Add a new submodule
./scripts/setup-submodule.sh ModuleName [--create-repos] [--go|--ts|--kotlin]

# Push a submodule to all upstreams
cd SubmoduleName && commit "message"

# Install upstream remotes for a submodule
cd SubmoduleName && install_upstreams
```

## Commands

```bash
# Backend (catalog-api)
cd catalog-api
go test ./...                        # all tests
go test -v -run TestName ./pkg/      # single test
go run main.go                       # dev server
go build -o catalog-api              # build

# Frontend (catalog-web)
cd catalog-web
npm run dev                          # dev server (:5173)
npm run test                         # tests
npm run build                        # production build
npm run lint && npm run type-check   # lint + typecheck

# Desktop / Installer Wizard
cd catalogizer-desktop   # or installer-wizard
npm run tauri:dev                    # dev
npm run tauri:build                  # build

# API Client
cd catalogizer-api-client
npm run build && npm run test

# Android
cd catalogizer-android   # or catalogizer-androidtv
./gradlew test                       # unit tests
./gradlew assembleDebug              # debug build

# Containers & full system (use podman if docker is unavailable)
podman-compose -f docker-compose.dev.yml up   # dev env
./scripts/run-all-tests.sh                    # all tests + security
```

## Architecture

**catalog-api**: Handler → Service → Repository → SQLite. Routes under `/api/v1` in `main.go`.
- `filesystem/interface.go` defines `UnifiedClient`; `filesystem/factory.go` creates per-protocol clients. New protocols: implement the interface.
- `internal/smb/`: circuit breaker + offline cache + exponential backoff retry.
- `internal/media/detector/` → `analyzer/` → `providers/` (TMDB, IMDB, etc.): detection pipeline.
- `internal/media/realtime/`: event bus → WebSocket → clients.
- `internal/auth/` + `middleware/`: JWT auth.

**catalog-web**: AuthProvider → WebSocketProvider → Router. Server state via React Query; auth-gated routes via `ProtectedRoute`.

**Android**: MVVM — Compose UI → ViewModel (StateFlow) → Repository → Room + Retrofit. Hilt DI.

**Tauri apps**: React frontend ↔ Rust backend via IPC commands/events.

**Challenges**: `digital.vasic.challenges` framework integrated via `Challenges/` submodule. Challenges are Go structs embedding `challenge.BaseChallenge` with custom `Execute()`. Registered in `catalog-api/challenges/register.go`, exposed via `/api/v1/challenges` REST endpoints. Challenge bank definitions in `challenges/data/challenges_bank.json`. **All challenge operations MUST be executed exclusively by system deliverables (compiled binaries) — the catalog-api service and other Catalogizer applications. Never use custom scripts, curl commands, or third-party tools to trigger API endpoints within challenge execution. Scanning, storage root creation, and all other operations must go through the running services, exactly as an end user would.**

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

**11 media types** (seeded in `media_types` table, migration v8): movie, tv_show, tv_season, tv_episode, music_artist, music_album, song, game, software, book, comic.

**Entity tables** (migration v8): media_types, media_items (parent_id self-ref for hierarchy), media_files (junction to files), media_collections, media_collection_items, external_metadata, user_metadata, directory_analyses, detection_rules.

### Entity API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/entities` | List entities with type filter, search, pagination |
| GET | `/api/v1/entities/:id` | Entity detail with files, metadata, children |
| GET | `/api/v1/entities/:id/children` | Child entities (episodes, songs) |
| GET | `/api/v1/entities/:id/files` | File versions for entity |
| GET | `/api/v1/entities/:id/metadata` | External metadata |
| GET | `/api/v1/entities/:id/duplicates` | Duplicate entities |
| GET | `/api/v1/entities/types` | Media types with counts |
| GET | `/api/v1/entities/browse/:type` | Browse by type with pagination |
| GET | `/api/v1/entities/stats` | Entity-level stats |
| GET | `/api/v1/entities/duplicates` | Global duplicate groups |
| GET | `/api/v1/entities/:id/stream` | Stream primary file |
| GET | `/api/v1/entities/:id/download` | Download file |
| POST | `/api/v1/entities/:id/metadata/refresh` | Trigger metadata refresh |
| PUT | `/api/v1/entities/:id/user-metadata` | Update user rating/watched/favorite |

### Key Entity Files

| File | Role |
|------|------|
| `repository/media_item_repository.go` | Entity CRUD, search, hierarchy, duplicates |
| `repository/media_file_repository.go` | File-entity linking |
| `repository/duplicate_entity_repository.go` | Duplicate group queries |
| `repository/external_metadata_repository.go` | External metadata CRUD |
| `repository/user_metadata_repository.go` | User ratings, favorites |
| `repository/directory_analysis_repository.go` | Detection result storage |
| `internal/services/aggregation_service.go` | Post-scan entity creation |
| `internal/services/title_parser.go` | Regex parsers for all media types |
| `handlers/media_entity_handler.go` | Entity browsing API |
| `internal/media/models/media.go` | All model structs |
| `catalog-web/src/pages/EntityBrowser.tsx` | Frontend type selector + entity grid |
| `catalog-web/src/pages/EntityDetail.tsx` | Entity detail with hierarchy navigation |

### Entity Challenges

| ID | Name | Validates |
|----|------|-----------|
| CH-016 | Entity Aggregation | Stats, types, listing after scan |
| CH-017 | Entity Browsing | Filtered listing, detail, pagination |
| CH-018 | Entity Metadata | External metadata, refresh, user metadata |
| CH-019 | Entity Duplicates | Duplicate groups, per-entity duplicates |
| CH-020 | Entity Hierarchy | TV show→seasons→episodes, album→songs |

### Entity Constraints

- All scanned files MUST be associated with a recognized media entity after aggregation
- Entity API endpoints MUST return real data from the database
- All entity types MUST support browsing, search, metadata, and playback/download
- Entity hierarchy: parent_id self-reference (TV Show → seasons → episodes, Music Artist → albums → songs)

## Root Directory Structure (Mandatory Locations)

New files MUST be placed in the correct directory. Do NOT add files to the project root unless they are conventional root files (README, LICENSE, .gitignore, docker-compose, etc.).

| Directory | Purpose |
|---|---|
| `challenges/` | Challenge bank definitions and runtime results |
| `config/` | Infrastructure config files (nginx.conf, redis.conf) |
| `scripts/` | Shell scripts (install, setup, CI/CD, testing runners) |
| `tests/` | Standalone/integration test files (test_*.js, test_*.go, test_*.sh) |
| `docs/` | All documentation markdown files, organized by subdirectory |
| `docs/architecture/` | Architecture and design docs |
| `docs/deployment/` | Deployment and Docker setup docs |
| `docs/testing/` | Test reports and testing docs |
| `docs/qa/` | QA guides and checklists |
| `docs/guides/` | User-facing guides and troubleshooting |
| `docs/status/` | Status reports, dashboards, completion summaries |
| `docs/phases/` | Phase-specific progress and completion reports |
| `docs/roadmap/` | Roadmap and planning docs |
| `Assets/` | Static assets (images, HTML tutorials) |

Docker Compose files (`docker-compose.yml`, `docker-compose.dev.yml`) reference `config/` for nginx and redis configs. Do NOT move these config files without updating the Compose volume mounts.

## Container Runtime

**Always use Podman when Docker is not available.** This project supports both Docker and Podman as container runtimes. Before running any container command, check which is available and prefer `podman`/`podman-compose` over `docker`/`docker-compose`. All `docker-compose.yml` files are compatible with both runtimes.

```bash
# Check available runtime
which podman && podman --version || which docker && docker --version

# Use podman-compose (preferred) or docker-compose
podman-compose -f docker-compose.dev.yml up
podman-compose -f docker-compose.yml config --quiet  # validate

# Single container commands
podman run / podman build / podman ps  # instead of docker equivalents
```

## Constraints

**GitHub Actions are PERMANENTLY DISABLED.** All workflow files have been deleted from `.github/workflows/`. Do NOT create any GitHub Actions workflow files (*.yml, *.yaml) in this directory. CI/CD, security scanning, and automated builds must be run locally using the commands documented below.

**All builds and services MUST use containers.** Never build or run services directly on the host machine. Always use the containerized build pipeline (`./scripts/container-build.sh` or `podman-compose -f docker-compose.build.yml`) for builds, and `podman-compose` / `podman run` for running services. The builder container has all required toolchains (Go, Node, Rust, JDK, Android SDK). Nothing — builds, tests, service execution — should be executed directly on the host. Use `podman run --network host` for single-container builds and `podman-compose` for multi-service environments.

## Local Development Setup

### Prerequisites
- **Go** 1.21+ (for catalog-api)
- **Node.js** 18+ and npm (for catalog-web, installer-wizard, catalogizer-desktop)
- **Rust** and Cargo (for Tauri apps)
- **Android Studio** with Kotlin (for Android apps)
- **SQLite3** or **PostgreSQL** (database)
- **Podman** or **Docker** (optional, for containerized development)

### Database Setup

**SQLite (Development - Recommended):**
```bash
# No setup needed - catalog-api creates catalogizer.db automatically
cd catalog-api && go run main.go
```

**PostgreSQL (Production):**
```bash
# Set environment variables
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=catalogizer
export DB_USER=catalogizer
export DB_PASSWORD=your_password
```

### Environment Variables

Create `.env` file in `catalog-api/`:
```env
# Server
PORT=8080
GIN_MODE=debug

# Database (SQLite default, or set for PostgreSQL)
DB_TYPE=sqlite
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME=catalogizer
# DB_USER=catalogizer
# DB_PASSWORD=secret

# Authentication
JWT_SECRET=your-dev-secret-key
ADMIN_PASSWORD=admin123

# External APIs (optional)
TMDB_API_KEY=your_tmdb_key
OMDB_API_KEY=your_omdb_key
```

### Running the Full Stack

```bash
# Terminal 1: Backend
cd catalog-api && go run main.go

# Terminal 2: Frontend
cd catalog-web && npm install && npm run dev

# Access: http://localhost:5173 (frontend) / http://localhost:8080 (API)
```

## Testing

### Running Tests by Component

```bash
# Go tests (all packages)
cd catalog-api && go test ./...

# Go tests (single test with verbose output)
cd catalog-api && go test -v -run TestFunctionName ./path/to/package/

# Web tests (Vitest)
cd catalog-web && npm run test         # watch mode
cd catalog-web && npm run test -- --run  # single run

# Installer wizard tests
cd installer-wizard && npm run test -- --run

# All tests
./scripts/run-all-tests.sh
```

### Test File Conventions
- Go: `*_test.go` alongside source files
- React/TS: `__tests__/*.test.tsx` or `*.test.ts` in same directory
- Test helper in `catalog-api/internal/tests/test_helper.go` provides SQLite test database setup

## Zero Warning / Zero Error Policy

**All components must run with zero console warnings, zero console errors, and zero failed network requests.** This policy applies to every environment: development, container, and production.

### Rules
- **No browser console errors or warnings.** Every `console.error`, `console.warn`, and failed network request visible in browser DevTools is a defect that must be fixed.
- **No unhandled API errors.** Every API endpoint the frontend calls must exist, return valid responses (2xx), and match the expected response shape. A 4xx or 5xx response from any endpoint used by the UI is a bug.
- **No framework deprecation warnings.** React Router future flags, React strict mode warnings, and similar deprecation notices must be addressed proactively.
- **No WebSocket connection failures.** The backend must expose all endpoints the frontend expects. If a feature is not yet implemented, provide a stub endpoint that returns a valid empty response.
- **Challenges enforce this policy.** The challenge suite validates end-to-end system behavior: CH-001 to CH-007 validate NAS connectivity and content discovery, CH-008 populates the catalog database via the scan pipeline, CH-009 to CH-011 validate API endpoints and web app, CH-012 to CH-015 validate assets and database, and CH-016 to CH-020 validate the entity system (aggregation, browsing, metadata, duplicates, hierarchy). Challenges must fail if any endpoint returns an error.

### How to Verify
```bash
# 1. Start services
./scripts/services-up.sh

# 2. Open http://localhost:3000 in browser
# 3. Open browser DevTools (F12) → Console tab
# 4. Login with admin / admin123
# 5. Navigate to Media, Dashboard, Analytics, Admin pages
# 6. Console must show ZERO errors and ZERO warnings
# 7. Network tab must show ZERO failed (red) requests

# 8. Run challenges to verify programmatically
curl -s -X POST http://localhost:8080/api/v1/challenges/run \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"challenge_id":"browsing-api-health"}'
# All 20 challenges (CH-001 to CH-020) must pass
```

## CRITICAL: Host Resource Limits (30-40% Maximum)

The host machine runs other mission-critical processes. All tests, challenges, builds, and container workloads MUST be strictly limited to 30-40% of total host resources. Exceeding this limit can freeze the entire system, requiring a hard reset. Apply these limits:

- **Go tests**: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2` (max 3 OS threads, 2 packages at a time, 2 parallel tests per package)
- **Container CPU/memory limits** (mandatory for all `podman run`):
  - PostgreSQL: `--cpus=1 --memory=2g`
  - API server: `--cpus=2 --memory=4g`
  - Web frontend: `--cpus=1 --memory=2g`
  - Builder: `--cpus=3 --memory=8g`
- **Challenges**: Run sequentially via the API, never in parallel
- **Total container budget**: max 4 CPUs, 8 GB RAM across all running containers
- **Monitor**: Use `podman stats --no-stream` and `cat /proc/loadavg` to verify resource usage stays within bounds

## CRITICAL: HTTP/3 (QUIC) with Brotli Compression (Mandatory)

All network communication in every component of the system MUST use **HTTP/3 (QUIC) as the primary protocol** with **Brotli compression** as the default content encoding. This applies to all API servers, web servers, reverse proxies, client libraries, and inter-service communication.

- **Primary**: HTTP/3 (QUIC) + Brotli compression
- **Fallback**: HTTP/2 + gzip compression (only when HTTP/3 is unavailable)
- **Never**: HTTP/1.1 in production (development/debugging only)

Implementation requirements:
- **catalog-api (Go)**: Use QUIC-enabled server (e.g., `quic-go`), enable Brotli middleware
- **catalog-web (React)**: Serve via HTTP/3-capable reverse proxy (nginx with QUIC or Caddy), Brotli-compressed static assets
- **catalogizer-desktop / installer-wizard (Tauri)**: HTTP/3 client for API communication
- **catalogizer-android / catalogizer-androidtv**: OkHttp with HTTP/3 (Cronet) + Brotli
- **catalogizer-api-client (TS)**: HTTP/3-capable fetch/axios with Brotli Accept-Encoding
- **Reverse proxy / Load balancer**: Must terminate HTTP/3 and negotiate Brotli, with HTTP/2+gzip fallback

## Conventions

- **Go**: `NewService` constructor injection, error wrapping, table-driven tests, `*_test.go` beside source
- **TypeScript**: PascalCase components, camelCase functions, Zod validation, React Hook Form
- **Kotlin**: MVVM, Result sealed classes, Room for offline
- **Config**: env vars > `.env` > `config.json` > defaults
