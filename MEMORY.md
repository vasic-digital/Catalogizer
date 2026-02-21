# Catalogizer Decoupling Refactoring — Progress Tracker

## Overview
Systematically extracting all reusable functionality into standalone Git submodules under vasic-digital.
**Goal**: 26 modules total (16 existing Go + 2 existing TS + 1 new Go + 7 new TS), all integrated as submodules.

## Current Phase: PHASE 5 — TypeScript Foundation Modules

## Existing Submodules (Before This Work)
- [x] `WebSocket-Client-TS` → `@vasic-digital/websocket-client`
- [x] `UI-Components-React` → `@vasic-digital/ui-components`
- [x] `Challenges` → `digital.vasic.challenges`
- [x] `Assets` → `digital.vasic.assets`

## Phase 1 — Foundation Go Modules
Status: COMPLETE ✅

| Module | Repo | Status | Notes |
|--------|------|--------|-------|
| digital.vasic.config | Config | ✅ Done | LoadFile() replaces manual JSON decode in internal/config/ |
| digital.vasic.database | Database | ✅ Submodule only | API differs (modernc vs mattn sqlite); deferred code integration |
| digital.vasic.filesystem | Filesystem | ✅ Done | Type aliases in filesystem/interface.go; all tests pass |
| digital.vasic.concurrency | Concurrency | ✅ Done | Circuit breaker wrapper in internal/recovery/; all tests pass |

## Phase 2 — Infrastructure Go Modules
Status: COMPLETE ✅

| Module | Repo | Status | Notes |
|--------|------|--------|-------|
| digital.vasic.auth | Auth | ✅ Done | JWT Manager integrated in internal/auth/service.go |
| digital.vasic.middleware | Middleware | ✅ Submodule only | net/http vs Gin incompatibility; deferred |
| digital.vasic.ratelimiter | RateLimiter | ✅ Submodule only | Needs Gin adapter; deferred |
| digital.vasic.observability | Observability | ✅ Submodule only | ClickHouse deps; deferred |

## Phase 3 — Advanced Go Modules
Status: COMPLETE ✅

| Module | Repo | Status | Notes |
|--------|------|--------|-------|
| digital.vasic.media | Media | ✅ Submodule only | Detector uses DB-rules system vs vasic filename-only engine |
| digital.vasic.watcher | Watcher | ✅ Submodule only | Watcher has DB/analyzer deps; can't use type aliases |
| digital.vasic.eventbus | EventBus | ✅ Done | Type aliases in internal/eventbus/; Catalogizer event constants defined |
| digital.vasic.cache | Cache | ✅ Done | Type aliases in internal/cache/; Cache interface + DefaultConfig + NewTypedCache |
| digital.vasic.security | Security | ✅ Submodule only | AI guardrails vs HTTP input validation — different domain |
| digital.vasic.storage | Storage | ✅ Submodule only | MinIO client vs AWS SDK v2; different client libraries |
| digital.vasic.streaming | Streaming | ✅ Submodule only | Generic net/http vs Gin-specific streaming |
| digital.vasic.discovery | Discovery | ✅ Submodule only | Simple TCP scanner vs Catalogizer's complex SMB scanner |

## Phase 4 — New Go Module: Entities
Status: COMPLETE ✅

| Module | Repo | Status | Notes |
|--------|------|--------|-------|
| digital.vasic.entities | Entities | ✅ Done | Created vasic-digital/Entities; pkg/parser (title parsers) + pkg/models; title_parser.go now delegates to vasicparser; all tests pass |

## Phase 5 — TypeScript Foundation Modules
Status: PENDING

| Module | Repo | Status | Notes |
|--------|------|--------|-------|
| @vasic-digital/media-types | Media-Types-TS | ⬜ Pending | NEW — Shared TS type definitions |
| @vasic-digital/catalogizer-api-client | Catalogizer-API-Client-TS | ⬜ Pending | NEW — Type-safe TS API client |
| @vasic-digital/auth-context | Auth-Context-React | ⬜ Pending | NEW — React AuthProvider + useAuth |

## Phase 6 — React UI Modules
Status: PENDING

| Module | Repo | Status | Notes |
|--------|------|--------|-------|
| @vasic-digital/media-browser | Media-Browser-React | ⬜ Pending | NEW — Entity browser + grid + pagination |
| @vasic-digital/media-player | Media-Player-React | ⬜ Pending | NEW — HTML5 media player |
| @vasic-digital/collection-manager | Collection-Manager-React | ⬜ Pending | NEW — Collection CRUD |
| @vasic-digital/dashboard-analytics | Dashboard-Analytics-React | ⬜ Pending | NEW — Stats + charts |

## Phase 7 — Documentation & GitHub Pages
Status: PENDING

## Phase 8 — Challenges per Module
Status: PENDING

## Phase 9 — Final Integration & Verification
Status: PENDING

---

## Test Baseline (Before Changes)
- Go tests: Run `cd catalog-api && GOMAXPROCS=3 go test ./... -p 2 -parallel 2`
- Web tests: Run `cd catalog-web && npm run test -- --run`
- Challenges: 20 total (CH-001 to CH-020)

## Key Decisions
- Use `replace` directives in go.mod for submodule paths (local filesystem)
- Run tests after each module incorporation before proceeding
- Never exceed 30-40% host resource usage during tests

## Last Updated
2026-02-21 — Initial MEMORY.md created. Starting Phase 1.
