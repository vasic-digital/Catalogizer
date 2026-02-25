# AGENTS.md - Catalogizer Development Guide

Essential commands and style guidelines for AI agents working in the Catalogizer codebase.

## Project Overview

Catalogizer is a multi-platform media collection manager with Go backend (catalog-api), React web frontend (catalog-web), Tauri desktop apps, Android apps, and TypeScript API client. Uses submodules for shared libraries.

## Essential Commands

### Backend (catalog-api)
```bash
cd catalog-api
go run main.go                    # dev server (dynamic port, writes .service-port)
go build -o catalog-api           # build binary
go test ./...                     # all tests
go test -v -run TestName ./pkg    # single test
go test -cover ./...              # coverage
go fmt ./...                      # format code
go vet ./...                      # static analysis
go mod tidy                       # dependencies
```

### Frontend (catalog-web)
```bash
cd catalog-web
npm run dev                       # dev server (port 3000)
npm run build                     # production build
npm run lint                      # ESLint
npm run lint:fix                  # auto-fix lint issues
npm run format                    # Prettier formatting
npm run type-check                # TypeScript type checking
npm run test                      # Vitest unit tests
npm run test:watch                # watch mode
npm run test:coverage             # coverage report
npm run test:e2e                  # Playwright E2E tests
# Single test: `npm run test -- -t "test name"`
```

### Desktop App (catalogizer-desktop)
```bash
cd catalogizer-desktop
npm run tauri:dev                 # dev with hot reload
npm run tauri:build               # build for current platform
npm run test                      # unit tests
```

### Installer Wizard (installer-wizard)
```bash
cd installer-wizard
npm run tauri:dev
npm run tauri:build
npm run test
npm run test:coverage
npm run health:check
```

### Android Apps
```bash
cd catalogizer-android   # or catalogizer-androidtv
./gradlew assembleDebug   # debug APK
./gradlew test            # unit tests
./gradlew lintKotlin      # linting
./gradlew installDebug    # build and install on emulator
```

### API Client Library (catalogizer-api-client)
```bash
cd catalogizer-api-client
npm run build
npm run test
npm run lint
```

### Docker/Podman Operations
```bash
podman-compose -f docker-compose.dev.yml up   # dev environment
podman-compose down                           # stop services
./scripts/services-up.sh                      # start all services
./scripts/services-down.sh                    # stop all services
```

### Running a Single Test
- **Go**: `go test -v -run TestFunctionName ./path/to/package`
- **TypeScript (Vitest)**: `npm run test -- -t "test name"`
- **Android (JUnit)**: `./gradlew test --tests "*TestClassName"`
- **Playwright (E2E)**: `npm run test:e2e -- --grep "test title"`

## Code Style Guidelines

### Go Backend
- **Naming**: PascalCase for exported identifiers, camelCase for unexported.
- **Interfaces**: `Reader`, `Writer`, `Service` suffixes.
- **Receivers**: Single-letter (e.g., `s *Service`).
- **Error handling**: Wrap errors with `fmt.Errorf` and `%w`. Use `errors.New` for simple errors.
- **Imports**: Group standard library, third-party, local imports separated by blank line.
- **Formatting**: `go fmt` standard. Use `gofumpt` if available.
- **Testing**: Table-driven tests with `t.Run` subtests. Use `*_test.go` files beside source.
- **Documentation**: Export all public functions with doc comments.

### TypeScript/React Frontend
- **Naming**: PascalCase for components/interfaces, camelCase for functions/variables, SCREAMING_SNAKE_CASE for constants.
- **Components**: Functional components with TypeScript interfaces for props.
- **Imports**: Group React, third-party, local imports. Use path aliases (`@/components`, `@/hooks`).
- **Formatting**: Prettier with Tailwind plugin. Line length 100.
- **Linting**: ESLint with React/TypeScript plugins. Rules: no-explicit-any warning, unused vars allowed with underscore prefix.
- **Error handling**: Use try/catch with proper error types. React Query for API errors.
- **State management**: React Query for server state, Zustand for client state.

### Kotlin/Android
- **Naming**: PascalCase for classes, camelCase for functions/variables.
- **Architecture**: MVVM with ViewModel, Repository pattern.
- **Dependency injection**: Hilt.
- **Coroutines**: Use `suspend` functions and `Flow` for asynchronous operations.
- **Testing**: JUnit for unit tests, Mockito for mocking.

## Constraints

**All builds and services MUST use containers.** Use Podman (not Docker). Run builds with `podman run --network host`. Use `podman-compose` for multi-service environments.

**GitHub Actions are PERMANENTLY DISABLED.** No CI/CD workflows.

**Host Resource Limits (30‑40% Maximum):**
- Go tests: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- Container limits: PostgreSQL `--cpus=1 --memory=2g`, API `--cpus=2 --memory=4g`, Web `--cpus=1 --memory=2g`.
- Total container budget: max 4 CPUs, 8 GB RAM across all containers.

**HTTP/3 (QUIC) with Brotli Compression (Mandatory):**
- All network communication must use HTTP/3 (QUIC) with Brotli compression.
- HTTP/2 with gzip is acceptable fallback. Never HTTP/1.1 in production.
- Implemented via `quic-go` (Go), HTTP/3-capable reverse proxy (web), Cronet (Android), and Brotli middleware.

## Challenge System

- **Challenge Execution Policy**: All challenge operations MUST be executed exclusively by system deliverables (compiled binaries) — the catalog-api service and other Catalogizer applications. Never use custom scripts, curl commands, or third-party tools to trigger API endpoints within challenge execution.
- **Running Challenges**: Use the API client built into the catalog-api binary. Challenges are registered in `catalog-api/challenges/register.go` and exposed via `/api/v1/challenges`.

## Quick Development Setup

1. Clone repository (with `--recursive` or run `git submodule init && git submodule update --recursive`) and run `./scripts/install.sh --mode=development`
2. Backend: `cd catalog-api && go run main.go`
3. Frontend: `cd catalog-web && npm run dev`
4. Access: Web UI at http://localhost:3000, API at http://localhost:8080

## Key Files for Reference
- `catalog-api/main.go` – API server entry point
- `catalog-api/filesystem/interface.go` – Unified filesystem interface
- `catalog-web/src/App.tsx` – React root component
- `catalog-web/vite.config.ts` – Vite configuration with path aliases
- `catalogizer-android/app/src/main/java/com/catalogizer/android/CatalogizerApplication.kt` – Android entry

**Note**: Always run linting and type checking before committing. Ensure zero console warnings/errors.