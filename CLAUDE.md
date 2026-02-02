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

# Docker & full system
docker-compose -f docker-compose.dev.yml up   # dev env
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

## Conventions

- **Go**: `NewService` constructor injection, error wrapping, table-driven tests, `*_test.go` beside source
- **TypeScript**: PascalCase components, camelCase functions, Zod validation, React Hook Form
- **Kotlin**: MVVM, Result sealed classes, Room for offline
- **Config**: env vars > `.env` > `config.json` > defaults
