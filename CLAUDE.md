# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Multi-platform media collection manager. Detects, categorizes, and organizes media across SMB, FTP, NFS, WebDAV, and local filesystems. Components: **catalog-api** (Go/Gin backend), **catalog-web** (React/TS frontend), **catalogizer-desktop** & **installer-wizard** (Tauri/Rust+React), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library), **Catalogizer/** (legacy Kotlin).

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

## Conventions

- **Go**: `NewService` constructor injection, error wrapping, table-driven tests, `*_test.go` beside source
- **TypeScript**: PascalCase components, camelCase functions, Zod validation, React Hook Form
- **Kotlin**: MVVM, Result sealed classes, Room for offline
- **Config**: env vars > `.env` > `config.json` > defaults
