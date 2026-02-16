# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go/Gin backend), **catalog-web** (React/TS frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## Submodule Architecture

Reusable functionality is extracted into independent git submodules under the vasic-digital organization. Each module has its own repo (GitHub + GitLab), tests, docs, and Upstreams for multi-remote push.

### Go Modules (Existing vasic-digital repos)

| Module | Path | Description |
|--------|------|-------------|
| `digital.vasic.auth` | `Auth/` | JWT, API key, OAuth2, HTTP auth middleware |
| `digital.vasic.cache` | `Cache/` | Redis/memory cache with TTL/eviction policies |
| `digital.vasic.database` | `Database/` | PostgreSQL/SQLite adapters, migrations, repository pattern |
| `digital.vasic.concurrency` | `Concurrency/` | Worker pools, circuit breaker, rate limiter, semaphore |
| `digital.vasic.storage` | `Storage/` | S3/local object storage with provider abstraction |
| `digital.vasic.eventbus` | `EventBus/` | Pub/sub event bus with middleware and filtering |
| `digital.vasic.streaming` | `Streaming/` | SSE, WebSocket, gRPC streaming, webhooks |
| `digital.vasic.security` | `Security/` | PII detection, content filtering, policy enforcement |
| `digital.vasic.observability` | `Observability/` | Tracing, Prometheus metrics, structured logging |
| `digital.vasic.formatters` | `Formatters/` | Code formatting framework with registry |
| `digital.vasic.plugins` | `Plugins/` | Plugin lifecycle, dynamic loading, sandboxing |
| `digital.vasic.challenges` | `Challenges/` | Challenge/test scenario framework |

### Go Modules (New, extracted from Catalogizer)

| Module | Path | Description |
|--------|------|-------------|
| `digital.vasic.filesystem` | `Filesystem/` | Multi-protocol filesystem (SMB, FTP, NFS, WebDAV, Local) |
| `digital.vasic.ratelimiter` | `RateLimiter/` | Sliding window rate limiter (memory + Redis) |
| `digital.vasic.config` | `Config/` | Config file/env loading with validation |
| `digital.vasic.discovery` | `Discovery/` | Network service/SMB share discovery |
| `digital.vasic.media` | `Media/` | Media type detection, metadata analysis, provider registry |
| `digital.vasic.middleware` | `Middleware/` | HTTP middleware (CORS, logging, recovery, request ID) |
| `digital.vasic.watcher` | `Watcher/` | Filesystem watcher with debounce and filtering |

### TypeScript/React Modules

| Module | Path | Description |
|--------|------|-------------|
| `@vasic-digital/websocket-client` | `WebSocket-Client-TS/` | WebSocket client with reconnection + React hooks |
| `@vasic-digital/ui-components` | `UI-Components-React/` | React UI component library (Button, Card, Input, etc.) |

### Android/Kotlin Module

| Module | Path | Description |
|--------|------|-------------|
| Android-Toolkit | `Android-Toolkit/` | Android utilities, UI components, Compose helpers |

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

## Root Directory Structure (Mandatory Locations)

New files MUST be placed in the correct directory. Do NOT add files to the project root unless they are conventional root files (README, LICENSE, .gitignore, docker-compose, etc.).

| Directory | Purpose |
|---|---|
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

## Conventions

- **Go**: `NewService` constructor injection, error wrapping, table-driven tests, `*_test.go` beside source
- **TypeScript**: PascalCase components, camelCase functions, Zod validation, React Hook Form
- **Kotlin**: MVVM, Result sealed classes, Room for offline
- **Config**: env vars > `.env` > `config.json` > defaults
