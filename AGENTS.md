# AGENTS.md - Catalogizer Development Guide

Essential commands and style guidelines for AI agents working in the Catalogizer codebase.

## Project Overview

Multi-platform media collection manager: **catalog-api** (Go/Gin backend), **catalog-web** (React/TS/Vite), **catalogizer-desktop** & **installer-wizard** (Tauri), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

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

### Container Operations
```bash
podman-compose -f docker-compose.dev.yml up             # dev environment
podman-compose down                                     # stop services
./scripts/services-up.sh                                # start all
./scripts/services-down.sh                              # stop all
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
- **DI**: Manual `DependencyContainer`. Async: `suspend` functions, `Flow`/`StateFlow`, Paging 3.
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

## Constraints

- **NEVER commit API keys/secrets** to git. Use `.env.example` with placeholders. Rotate immediately if leaked.
- **Container runtime**: Podman exclusively (not Docker). Production builds and QA must run in containers.
- **GitHub Actions**: PERMANENTLY DISABLED. No `.github/workflows/` files.
- **Host resource limits (30-40% max)**: Go tests use `GOMAXPROCS=3 -p 2 -parallel 2`. Containers: PostgreSQL `--cpus=1 --memory=2g`, API `--cpus=2 --memory=4g`, Web `--cpus=1 --memory=2g`. Total: max 4 CPUs, 8 GB RAM.
- **HTTP/3 (QUIC) with Brotli**: Mandatory. Fallback: HTTP/2 + gzip.
- **Zero-warning policy**: No console errors/warnings, no failed network requests. Unimplemented features need stub endpoints returning valid empty responses.
- **Config precedence**: env vars > `.env` > `config.json` > defaults.
- **PostCSS**: Must use `module.exports` (CommonJS) for Node 18 compat.

## Challenge System

All challenge operations executed by compiled binaries only (catalog-api service). Never use curl/scripts for API endpoints. Registered in `catalog-api/challenges/register.go`.

## Quick Setup

1. `git submodule init && git submodule update --recursive`
2. Backend: `cd catalog-api && go run main.go`
3. Frontend: `cd catalog-web && npm run dev`
4. Access: http://localhost:3000 (web), http://localhost:8080 (API)

## Key Files

- `catalog-api/main.go` — API entry point, route registration
- `catalog-api/database/dialect.go` — dual-dialect SQL rewriting
- `catalog-api/filesystem/interface.go` — `UnifiedClient` protocol abstraction
- `catalog-web/src/App.tsx` — React root (AuthProvider → WebSocketProvider → Router)
- `catalog-web/vite.config.ts` — path aliases, API proxy config

## Pre-Commit Checklist

- Go: `cd catalog-api && go fmt ./... && go vet ./...`
- TypeScript: `cd catalog-web && npm run lint && npm run type-check`
- Ensure zero console warnings/errors in browser
- Verify `.gitignore` covers `.env` — never commit secrets
