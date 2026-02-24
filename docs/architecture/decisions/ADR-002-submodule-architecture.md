# ADR-002: Submodule Architecture & Integration Decisions

## Status
Accepted (2026-02-23)

## Context

Catalogizer uses 29 independent git submodules under the vasic-digital organization. Each has its own repo, tests, and documentation. Of the 20 Go submodules, 10 are wired into `catalog-api` via `replace` directives in `go.mod`, and 10 are available but not imported. Of the 9 TypeScript/React submodules, all are declared in `catalog-web/package.json` but integration varies.

This ADR documents the rationale for which submodules are integrated and which are excluded.

## Decision

### Actively Integrated Go Submodules (10)

| Module | Path | Reason |
|--------|------|--------|
| `digital.vasic.challenges` | `Challenges/` | Core challenge framework for testing |
| `digital.vasic.assets` | `Assets/` | Asset management (lazy loading, serving) |
| `digital.vasic.containers` | `Containers/` | Container discovery and service port detection |
| `digital.vasic.concurrency` | `Concurrency/` | Concurrency utilities used in scanner |
| `digital.vasic.config` | `Config/` | Configuration management layer |
| `digital.vasic.filesystem` | `Filesystem/` | Unified protocol abstraction (SMB, FTP, NFS, WebDAV, local) |
| `digital.vasic.auth` | `Auth/` | Authentication primitives |
| `digital.vasic.cache` | `Cache/` | Caching layer |
| `digital.vasic.entities` | `Entities/` | Entity model definitions |
| `digital.vasic.eventbus` | `EventBus/` | Pub/sub event bus |

### Excluded Go Submodules (10) with Rationale

| Module | Path | Decision | Rationale |
|--------|------|----------|-----------|
| `digital.vasic.database` | `Database/` | EXCLUDED | **Redundant.** catalog-api has its own dual-dialect database layer (`database/dialect.go`) with PostgreSQL/SQLite support, connection pooling, and migrations. The submodule's implementation overlaps entirely. |
| `digital.vasic.discovery` | `Discovery/` | EXCLUDED | **Redundant.** catalog-api implements protocol discovery via `filesystem/` submodule and `internal/services/universal_scanner.go` with SMB, FTP, NFS, WebDAV, and local scanning. |
| `digital.vasic.media` | `Media/` | EXCLUDED | **Redundant.** catalog-api has `internal/media/detector/`, `internal/media/analyzer/`, and provider-based metadata extraction in `internal/media/providers/`. |
| `digital.vasic.middleware` | `Middleware/` | EXCLUDED | **Redundant.** catalog-api uses Gin's middleware system with custom implementations in `middleware/` and `internal/middleware/` (CORS, auth, rate limiting, compression). |
| `digital.vasic.observability` | `Observability/` | EXCLUDED | **Partially redundant.** catalog-api integrates Prometheus metrics directly (`internal/metrics/`) and uses `go.uber.org/zap` for structured logging. The submodule adds analytics features not currently needed. |
| `digital.vasic.ratelimiter` | `RateLimiter/` | DEFERRED | **Valuable for future.** Provides Redis-backed distributed rate limiting. catalog-api currently uses in-memory rate limiting (`middleware/advanced_rate_limiter.go`). Will integrate when horizontal scaling requires distributed state. |
| `digital.vasic.security` | `Security/` | DEFERRED | **Valuable for future.** Provides PII detection, content filtering, and guardrails. Not required for current media cataloging use case but useful for user-generated content features. |
| `digital.vasic.storage` | `Storage/` | EXCLUDED | **Not needed.** Provides S3/GCS object storage abstractions. Catalogizer accesses media on NAS/local filesystems via the `Filesystem` submodule, not cloud object stores. |
| `digital.vasic.streaming` | `Streaming/` | EXCLUDED | **Not needed.** Provides gRPC, HTTP streaming, and SSE abstractions. catalog-api uses WebSocket for real-time events (`internal/media/realtime/`) and HTTP/3 for API communication. |
| `digital.vasic.watcher` | `Watcher/` | EXCLUDED | **Redundant.** catalog-api has `internal/media/realtime/watcher.go` with file system watching, debouncing, and event filtering built specifically for the media scanning pipeline. |

### Actively Integrated TypeScript/React Submodules (9)

All 9 are declared in `catalog-web/package.json` via `file:../` dependencies:

| Module | Path | Integration Status |
|--------|------|-------------------|
| `@vasic-digital/media-types` | `Media-Types-TS/` | **Active** - shared type definitions re-exported from `src/types/auth.ts` and `src/types/media.ts` |
| `@vasic-digital/websocket-client` | `WebSocket-Client-TS/` | **Active** - imported in `src/lib/websocket.ts` for real-time events |
| `@vasic-digital/ui-components` | `UI-Components-React/` | **Available** - catalog-web has equivalent local components in `src/components/ui/` |
| `@vasic-digital/catalogizer-api-client` | `Catalogizer-API-Client-TS/` | **Available** - catalog-web uses local API modules in `src/lib/` |
| `@vasic-digital/auth-context` | `Auth-Context-React/` | **Available** - catalog-web has local `src/contexts/AuthContext.tsx` |
| `@vasic-digital/media-browser` | `Media-Browser-React/` | **Available** - catalog-web has local entity components in `src/components/entity/` |
| `@vasic-digital/media-player` | `Media-Player-React/` | **Available** - catalog-web has local player in `src/components/media/MediaPlayer.tsx` |
| `@vasic-digital/collection-manager` | `Collection-Manager-React/` | **Available** - catalog-web has local collection components in `src/components/collections/` |
| `@vasic-digital/dashboard-analytics` | `Dashboard-Analytics-React/` | **Available** - catalog-web has local dashboard in `src/pages/Dashboard.tsx` |

The TypeScript submodules with "Available" status are designed for reuse across multiple Catalogizer applications (web, desktop, Android). The catalog-web frontend evolved with its own implementations that are more tightly integrated with the app's specific needs (routing, state management, styling). The submodules serve as the shared foundation for other platforms.

## Consequences

- 10 Go submodules remain in the repository but are not imported, reducing catalog-api's dependency surface
- RateLimiter and Security submodules are candidates for future integration
- TypeScript submodules provide shared types and a component library available to all platforms
- Each submodule maintains independent versioning and can be updated without affecting catalog-api
