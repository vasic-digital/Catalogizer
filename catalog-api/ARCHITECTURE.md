# Architecture -- catalog-api

## Purpose

Go 1.24 REST API built with Gin serving as the backend for all Catalogizer clients (web, desktop, Android, Android TV). Provides media browsing, search, AI-powered recognition, recommendations, deep linking, reading experience, and multi-protocol storage access (SMB, FTP, NFS, WebDAV, local) with HTTP/3 (QUIC) and Brotli compression.

## Structure

```
handlers/              Domain HTTP handlers (catalog, media entities, search, downloads)
services/              Domain business logic
repository/            Data access layer (CRUD for files, media items, collections)
middleware/            Domain middleware (CORS, logging)
models/               Shared data structures
database/             Connection management, dialect abstraction (SQLite/PostgreSQL), migrations
filesystem/           Unified multi-protocol client (interface.go, factory.go)
challenges/           Challenge bank definitions, registration, and config
config/               Configuration loading (env vars > .env > config.json > defaults)
internal/
  auth/               JWT authentication with role-based access
  handlers/           Infrastructure HTTP handlers
  services/           Aggregation, title parsing, scanning, media detection pipeline
  middleware/         Infrastructure middleware (auth, rate limiting, metrics)
  media/
    detector/         Rule-based media type detection
    analyzer/         Metadata extraction
    providers/        External metadata providers (TMDB, IMDB, MusicBrainz, OpenLibrary, etc.)
    realtime/         Event bus -> WebSocket -> clients
  smb/                Circuit breaker, offline cache, exponential backoff retry
  metrics/            Prometheus metrics (exposed at /metrics)
  lifecycle/          LazyServiceRegistry for deferred service init with dependency ordering
  concurrency/        Semaphore-based concurrency control
  httpclient/         Pooled HTTP client with connection reuse and retry
```

## Key Components

- **Gin HTTP framework** -- Routes under `/api/v1` in main.go
- **Dual-dialect database** -- SQLite (dev) / PostgreSQL (production) with auto-rewriting SQL in database/dialect.go
- **Media entity pipeline** -- UniversalScanner -> AggregationService -> title parser -> entity creation -> hierarchy builder -> duplicate detection
- **Dynamic port binding** -- Writes `.service-port` for frontend discovery
- **HTTP/3 (QUIC)** -- quic-go/http3 with self-signed TLS certs generated at startup
- **Brotli compression** -- andybalholm/brotli middleware
- **22 submodule integrations** -- Via `replace` directives in go.mod

## Data Flow

```
HTTP Request -> Gin Router -> Middleware (auth, CORS, metrics, rate limit)
    |
    Handler -> Service -> Repository -> database.DB (auto-rewrites SQL for dialect)
    |
    Media detection: filesystem.Client.ListDirectory() -> detector.Engine -> analyzer -> provider.Registry
    |
    Real-time: internal/media/realtime -> EventBus -> WebSocket -> connected clients
```

## Dependencies

Key dependencies include Gin, quic-go/http3, go-redis/v9, go-sqlcipher, and 22 internal submodules (Assets, Auth, Cache, Challenges, Concurrency, Config, Containers, Database, Discovery, Entities, EventBus, Filesystem, Lazy, Media, Memory, Middleware, Observability, RateLimiter, Recovery, Security, Storage, Streaming, Watcher).

## Testing Strategy

Table-driven tests with `testify`. Test helper provides in-memory SQLite via `database.WrapDB(sqlDB, DialectSQLite)`. Resource-limited: `GOMAXPROCS=3 go test ./... -p 2 -parallel 2`. Constructor injection for all services enables mock-based testing.
