# AGENTS.md - Catalogizer Development Guide

Essential commands and style guidelines for AI agents working in the Catalogizer codebase.

## Project Overview

Multi-platform media collection manager: **catalog-api** (Go/Gin backend), **catalog-web** (React/TS/Vite), **catalogizer-desktop** & **installer-wizard** (Tauri), **catalogizer-android** & **catalogizer-androidtv** (Kotlin/Compose), **catalogizer-api-client** (TS library).

## Essential Commands

### Backend (catalog-api)
```bash
cd catalog-api
go run main.go                              # dev server (dynamic port, writes .service-port)
go build -o catalog-api                     # build binary
go test ./...                               # all tests
go test -v -run TestFunctionName ./path/    # single test
go test -cover ./...                        # coverage
go fmt ./... && go vet ./...                # format + lint
```

### Frontend (catalog-web)
```bash
cd catalog-web
npm run dev                                 # dev server (port 3000)
npm run build                               # production build (tsc + vite)
npm run lint                                # ESLint
npm run lint:fix                            # auto-fix lint issues
npm run type-check                          # TypeScript check
npm run test                                # Vitest (single run)
npm run test -- -t "test name"              # single test
npm run test:watch                          # watch mode
npm run test:coverage                       # coverage
npm run test:e2e                            # Playwright E2E
npm run test:e2e -- --grep "test title"     # single E2E test
```

### Desktop Apps (Tauri)
```bash
cd catalogizer-desktop  # or installer-wizard
npm run tauri:dev       # dev with hot reload
npm run tauri:build     # build for platform
npm run test            # unit tests
```

### Android
```bash
cd catalogizer-android  # or catalogizer-androidtv
./gradlew test          # unit tests
./gradlew test --tests "*TestClassName"   # single test
./gradlew assembleDebug                    # debug APK
./gradlew lintKotlin                      # lint
```

### Container Operations
```bash
podman-compose -f docker-compose.dev.yml up   # dev environment
podman-compose down                           # stop services
./scripts/services-up.sh                      # start all
./scripts/services-down.sh                    # stop all
```

## Code Style Guidelines

### Go Backend
- **Naming**: PascalCase exported, camelCase unexported. Interfaces: `Reader`, `Writer`, `Service` suffixes.
- **Receivers**: Single-letter (e.g., `s *Service`).
- **Imports**: Group stdlib, third-party, local with blank lines:
  ```go
  import (
      "encoding/json"
      "net/http"
      
      "github.com/gin-gonic/gin"
      
      "catalogizer/models"
  )
  ```
- **Error handling**: Wrap with `fmt.Errorf("context: %w", err)`. Use `errors.New` for simple errors.
- **Testing**: Table-driven tests with `t.Run`. Use `testify/suite` for test suites. Files: `*_test.go` beside source.
- **Constructors**: `NewService(dep Dependency) *Service` pattern with dependency injection.
- **Formatting**: `go fmt` (or `gofumpt`). All public functions need doc comments.

### TypeScript/React Frontend
- **Naming**: PascalCase components/interfaces, camelCase functions/variables, SCREAMING_SNAKE_CASE constants.
- **Components**: Functional components with TypeScript interfaces:
  ```tsx
  interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    loading?: boolean
  }
  const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, loading, children, ...props }, ref) => { ... }
  )
  ```
- **Imports**: Group React, third-party, local. Use path aliases:
  ```tsx
  import React from 'react'
  import { cva, type VariantProps } from 'class-variance-authority'
  import { cn } from '@/lib/utils'
  ```
- **Path aliases**: `@/components`, `@/hooks`, `@/lib`, `@/types`, `@/services`, `@/store`, `@/pages`, `@/assets`.
- **Formatting**: Prettier. Tailwind classes via `cn()` utility from `@/lib/utils`.
- **Linting**: ESLint with `@typescript-eslint`, `react`, `react-hooks`. Unused vars with `_` prefix allowed.
- **State**: React Query for server state, Zustand for client state.
- **Forms**: React Hook Form + Zod validation (`@hookform/resolvers`).
- **Testing**: Vitest + React Testing Library. Test files: `__tests__/` or `.test.tsx` beside source.

### Kotlin/Android
- **Naming**: PascalCase classes, camelCase functions/variables.
- **Architecture**: MVVM with ViewModel, Repository pattern, Hilt DI.
- **Async**: `suspend` functions, `Flow` for streams, Paging 3 for lists.
- **Testing**: JUnit, Mockito.
- **Error handling**: Sealed `Result` classes for operation outcomes.

## Constraints

**Container Runtime**: Use Podman exclusively (not Docker). Build with `podman run --network host`.

**GitHub Actions**: PERMANENTLY DISABLED. No CI/CD workflows.

**Host Resource Limits (30-40% max)**:
- Go tests: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- Containers: PostgreSQL `--cpus=1 --memory=2g`, API `--cpus=2 --memory=4g`, Web `--cpus=1 --memory=2g`.
- Total budget: max 4 CPUs, 8 GB RAM.

**HTTP/3 (QUIC) with Brotli**: Mandatory for all network communication. Fallback: HTTP/2 + gzip.

## Database Dialect

SQLite (dev) and PostgreSQL (prod). Use the `database.DB` wrapper:
```go
// Use ? placeholders - auto-converted to $1, $2... for Postgres
cutoff := time.Now().Add(-24 * time.Hour)
db.Query("SELECT * FROM table WHERE created_at > ?", cutoff)

// Dialect-specific expressions
if db.Dialect().IsPostgres() {
    expr = "EXTRACT(EPOCH FROM (MAX(t) - MIN(t)))"
} else {
    expr = "(julianday(MAX(t)) - julianday(MIN(t))) * 86400"
}
```

## Challenge System

All challenge operations executed by compiled binaries only (catalog-api service). Never use curl/scripts for API endpoints. Challenges registered in `catalog-api/challenges/register.go`.

## Quick Setup

1. `git submodule init && git submodule update --recursive`
2. Backend: `cd catalog-api && go run main.go`
3. Frontend: `cd catalog-web && npm run dev`
4. Access: http://localhost:3000 (web), http://localhost:8080 (API)

## Key Files
- `catalog-api/main.go` - API entry point
- `catalog-api/filesystem/interface.go` - Unified filesystem interface
- `catalog-web/src/App.tsx` - React root
- `catalog-web/vite.config.ts` - Path aliases, proxy config

## Pre-Commit Checklist

Run linting and type checking before committing:
- Go: `go fmt ./... && go vet ./...`
- TypeScript: `npm run lint && npm run type-check`
- Ensure zero console warnings/errors in browser
