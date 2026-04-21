# Catalogizer API Contracts

> **Purpose.** Single source of truth for the operational contract between
> `catalog-api` and every client (catalog-web, catalogizer-android,
> catalogizer-androidtv, catalogizer-desktop, installer-wizard,
> catalogizer-api-client). Layered on top of
> [`docs/api/openapi.yaml`](api/openapi.yaml) (197 endpoint operations) —
> this document adds dimensions openapi does not express:
>
> 1. Authorization tier per route group
> 2. Rate-limit tier per route group
> 3. Which clients consume which route
> 4. WebSocket real-time event contract
> 5. Dynamic-port + `.service-port` protocol
>
> **Last refresh:** 2026-04-21.

## 1. Base URL + Port Binding Protocol

catalog-api binds a dynamic port at startup and writes the port number
to `catalog-api/.service-port`. Clients use this contract to reach it:

| Client | Source of truth |
|---|---|
| catalog-web (dev) | `vite.config.ts` reads `../catalog-api/.service-port` (fallback 8080) |
| catalog-web (prod) | Build-time `VITE_API_URL` env |
| catalogizer-android | `BASE_URL` in `ApiConfig.kt` (compile-time) |
| catalogizer-androidtv | `BASE_URL` in `ApiConfig.kt` (compile-time) |
| catalogizer-desktop | Tauri IPC → user-configured URL |
| catalogizer-api-client (TS) | Constructor `baseUrl` arg |

Production servers expose:
- HTTP/3 (QUIC) on port 8443
- HTTP/2 fallback on port 8080
- HTTP/1.1 **forbidden** in production (RULE-CONST-005)

## 2. Authorization Tiers

The entire API is organised into four tiers:

| Tier | Header | Enforcement |
|---|---|---|
| **PUBLIC** | (none) | Open — used for health, OpenAPI, login |
| **JWT** | `Authorization: Bearer <token>` | `internal/auth/middleware.go#RequireAuth` |
| **JWT-ADMIN** | JWT + `role == admin` | `internal/auth/middleware.go#RequireAdmin` |
| **API-KEY** | `X-API-Key: <key>` | `internal/auth/middleware.go#RequireAPIKey` |

### 2.1 Endpoint group → auth tier

| Path prefix | Tier | Notes |
|---|---|---|
| `/api/v1/health`, `/health`, `/metrics` | PUBLIC | Liveness + Prometheus |
| `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/api/v1/auth/register` | PUBLIC | Authentication entry points |
| `/api/v1/auth/logout`, `/api/v1/auth/me` | JWT | Session introspection |
| `/api/v1/entities/*`, `/api/v1/media/*`, `/api/v1/catalog/*` | JWT | Read-only browsing |
| `/api/v1/search/*`, `/api/v1/duplicates/*` | JWT | Search + duplicate detection |
| `/api/v1/collections/*`, `/api/v1/favorites/*` | JWT | User-owned state |
| `/api/v1/scan/*`, `/api/v1/realtime/*` | JWT | Scanning + live progress |
| `/api/v1/cover/*`, `/api/v1/image-proxy/*` | JWT | Image delivery |
| `/api/v1/challenges/*` | JWT-ADMIN | Challenge framework |
| `/api/v1/admin/*`, `/admin/*` (aliases) | JWT-ADMIN | Config, errors, logs |
| `/api/v1/automation/*`, `/api/v1/llm/*` | JWT-ADMIN | Operator tooling |
| `/ws` (WebSocket upgrade) | JWT (via `?token=` query param) | Real-time events |

Login flow:
1. `POST /api/v1/auth/login` with `{username, password}` → `200 {accessToken, refreshToken, expiresAt}` (JWT in body, not set-cookie)
2. Every subsequent request `Authorization: Bearer <accessToken>`
3. On 401 with `code=TOKEN_EXPIRED`, client calls `POST /api/v1/auth/refresh` with `{refreshToken}` → new `accessToken`

## 3. Rate-Limit Tiers

Redis-backed sliding window (RULE-GO-005). Enforced by
`internal/middleware/ratelimit.go`:

| Tier | Budget | Scope | Endpoints |
|---|---|---|---|
| **AUTH** | 10 req / IP / minute | per-IP | `POST /auth/login`, `POST /auth/refresh`, `POST /auth/register` |
| **READ** | 600 req / user / minute | per-user | All `GET /entities`, `/media`, `/search`, `/cover`, `/image-proxy` |
| **WRITE** | 60 req / user / minute | per-user | `POST|PUT|DELETE` mutations |
| **SCAN** | 5 req / user / minute | per-user | `POST /scan`, `POST /catalog/scan-dir` |
| **ADMIN** | 30 req / user / minute | per-user | `/admin/*`, `/challenges/*` |
| **NONE** | unlimited | — | `/health`, `/metrics`, WebSocket frames |

Exceeding the budget returns `HTTP 429` with headers:
```
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1745247900
Retry-After: 23
```

## 4. Response Envelope

All successful JSON responses follow:

```json
{
  "data": <payload>,
  "meta": {
    "requestId": "uuid",
    "timestamp": "2026-04-21T22:40:00Z"
  },
  "pagination": {        // only when applicable
    "page": 1,
    "pageSize": 20,
    "total": 189,
    "hasNext": true
  }
}
```

All error responses follow:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "media item with id=42 not found",
    "details": {},        // optional structured detail
    "requestId": "uuid"
  }
}
```

Standard error `code` values: `BAD_REQUEST`, `UNAUTHENTICATED`,
`FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`,
`RATE_LIMITED`, `INTERNAL_ERROR`, `TOKEN_EXPIRED`,
`SERVICE_UNAVAILABLE`.

## 5. Client Consumption Matrix

Which client calls which route group. Used by Phase 10 cross-platform
contract validation.

| Route group | web | android | androidtv | desktop | installer | api-client |
|---|:-:|:-:|:-:|:-:|:-:|:-:|
| `/auth/*` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `/health`, `/metrics` | ✓ | — | — | ✓ | ✓ | ✓ |
| `/entities/*` | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `/entities/browse/movie`, `/browse/tv_show`, `/browse/music_album`, `/browse/song`, `/browse/game`, `/browse/software`, `/browse/book`, `/browse/comic` | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `/collections/*`, `/favorites/*` | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `/search/*` | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `/duplicates/*` | ✓ | — | — | ✓ | — | ✓ |
| `/scan/*`, `/catalog/scan-dir` | ✓ | — | — | ✓ | ✓ | ✓ |
| `/cover/*`, `/image-proxy/*` | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `/realtime/*`, `/ws` | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `/challenges/*` | — | — | — | ✓ | — | ✓ |
| `/admin/*` | ✓ | — | — | ✓ | ✓ | ✓ |
| `/automation/*`, `/llm/*` | ✓ | — | — | ✓ | — | ✓ |
| `/userflow/*` | — | — | — | — | ✓ | — |

Contract test at `tests/contract/` must verify every ✓ cell parses the
canonical response successfully.

## 6. WebSocket Event Contract

**Endpoint:** `GET /ws?token=<JWT>` (upgrade to WebSocket).

Clients MUST auto-reconnect (RULE-WEB-002). On disconnect:
exponential backoff 500ms → 16s with jitter, max 20 retries.

### 6.1 Message framing

Every frame is a JSON text message of the shape:

```json
{
  "type": "<event-type>",
  "seq": 1234,
  "timestamp": "2026-04-21T22:40:00Z",
  "payload": { ... }
}
```

`seq` is per-connection monotonically increasing; clients detect gaps
and request re-sync via `POST /realtime/sync?since=<seq>`.

### 6.2 Event types

| `type` | When fired | Debounce | Payload shape |
|---|---|---|---|
| `scan.started` | User triggers scan | none | `{scanId, root, protocol}` |
| `scan.progress` | Every 200 files OR 2s, whichever first | 2s | `{scanId, filesSeen, filesTotal, etaSeconds}` |
| `scan.completed` | Scan finishes | none | `{scanId, duration, added, updated, removed}` |
| `scan.failed` | Scan aborts | none | `{scanId, error, stage}` |
| `media.added` | New media_item | batch 500ms | `{items: [...]}` |
| `media.updated` | Existing item changed | batch 500ms | `{items: [...]}` |
| `media.removed` | Item deleted | batch 500ms | `{ids: [...]}` |
| `collection.changed` | Collection CRUD | none | `{collectionId, action}` |
| `health.degraded` | Backend detected issue | 30s min | `{service, level, reason}` |
| `health.restored` | Issue cleared | none | `{service}` |

### 6.3 Heartbeat

Server sends `{"type": "ping", "seq": N}` every 30s; client responds with
`{"type": "pong", "seq": N}`. Two missed pongs → server closes the
connection with code `4001 heartbeat_timeout`.

## 7. Real-Time Update Triggers

These server-side events MUST emit the corresponding WebSocket events.
Breaking this mapping breaks live updates on every client.

| Backend event | WS event(s) | Source file |
|---|---|---|
| `UniversalScanner` starts | `scan.started` | `internal/scan/scanner.go` |
| `UniversalScanner` file processed | `scan.progress` (debounced) | `internal/scan/scanner.go` |
| `UniversalScanner` finishes | `scan.completed` or `scan.failed` | `internal/scan/scanner.go` |
| `AggregationService.AggregateAfterScan` inserts | `media.added` | `internal/services/aggregation_service.go` |
| `MediaItemRepo.Update` | `media.updated` | `repository/media_item_repository.go` |
| `MediaItemRepo.Delete` | `media.removed` | `repository/media_item_repository.go` |
| `CollectionService` CRUD | `collection.changed` | `services/collection_service.go` |
| Redis connection loss | `health.degraded` (service=redis) | `internal/cache/redis.go` |
| DB read timeout > 5s | `health.degraded` (service=database) | `database/connection.go` |

## 8. Contract Stability Rules

1. **Adding a field is safe.** All clients parse JSON laxly — extra
   fields are ignored.
2. **Removing a field is a breaking change.** Clients may rely on its
   presence. Requires deprecation period ≥ 90 days + version bump.
3. **Changing a field's type is a breaking change.** Always. No
   exceptions.
4. **Adding a required request field is a breaking change.** Make it
   optional with a safe default.
5. **Adding a new endpoint is safe** if clients don't need to call it.
6. **Renaming an endpoint path is a breaking change.** Use the alias
   pattern (`/health` vs `/api/v1/health`) and support both for ≥ 90 days.
7. **Error code additions are safe** if clients handle unknown codes as
   `INTERNAL_ERROR`. Removals are breaking.

## 9. OpenAPI Source of Truth

The full schema — every endpoint, every request/response body, every
field type — lives in [`docs/api/openapi.yaml`](api/openapi.yaml).
197 operations covered. Companion docs in `docs/api/`:

- `API_DOCUMENTATION.md` — narrative reference
- `BROWSE_API.md` — entity browsing semantics
- `SEARCH_API.md` — search endpoint shape
- `SYNC_API.md` — sync + delta endpoints
- `WEBSOCKET_EVENTS.md` — detailed WS event reference
- `CHANGELOG.md` — breaking-change log

## 10. Per-Endpoint Contract Summaries

These are the 35 endpoints all clients rely on most. Full schema lives
in `docs/api/openapi.yaml`; each entry below captures the operational
contract a client integrator must respect.

### GET /api/v1/health

Public liveness probe. Returns `{"data":{"status":"ok","version":"..."}}`
with HTTP 200. No auth, no rate limit, idempotent. Alias: `/health`.

### GET /metrics

Public Prometheus scrape endpoint. No auth. Returns `text/plain`
Prometheus exposition. Used by operator monitoring only.

### POST /api/v1/auth/login

Body `{username, password}`. 200 → `{accessToken, refreshToken, expiresAt,
user{id, username, role}}`. 401 on bad creds. Rate tier `AUTH` (10 req /
IP / min). No JWT required.

### POST /api/v1/auth/refresh

Body `{refreshToken}`. 200 → `{accessToken, expiresAt}`. 401 on expired
or revoked token. Rate tier `AUTH`. No access-token required.

### POST /api/v1/auth/logout

JWT required. Invalidates the caller's refresh tokens server-side.
Returns 204. Rate tier `WRITE`.

### GET /api/v1/auth/me

JWT required. Returns the authenticated user's profile record.
Rate tier `READ`.

### GET /api/v1/entities/browse/{type}

JWT required. `{type}` is one of 11 media types
(`movie`, `tv_show`, `tv_season`, `tv_episode`, `music_artist`,
`music_album`, `song`, `game`, `software`, `book`, `comic`). Query:
`page`, `pageSize` (≤100), `sort_by` (`created`, `title`, `year`,
`rating`), `sort_order` (`asc`/`desc`). 200 → paginated list with
`cover_url` populated. Rate tier `READ`.

### GET /api/v1/entities/{id}

JWT required. Full entity with external metadata + file list +
hierarchy (season→episodes, artist→albums→songs). 404 if unknown.
Rate tier `READ`.

### GET /api/v1/entities/{id}/files

JWT required. List of `media_files` rows joined to this entity.
Includes protocol, path, size, duration, resolution. Rate tier `READ`.

### GET /api/v1/entities/{id}/related

JWT required. Entities sharing metadata (same director, series, artist).
Paginated. Rate tier `READ`.

### PUT /api/v1/media/{id}/favorite

JWT required. Body `{favorite: true|false}`. 200 → updated record.
Requires migration v18 schema (is_favorite column). Rate tier `WRITE`.

### GET /api/v1/search

JWT required. Query `q` (free-text), `type` (media type filter),
`year_from`, `year_to`, `page`, `pageSize`. 200 → results with matched
fields highlighted. Rate tier `READ`.

### GET /api/v1/search/adversarial

Same as `/search` but explicitly tests Cyrillic input, SQL-injection
payloads, XSS payloads. Expected to behave safely (no error, no leak).
Rate tier `READ`.

### GET /api/v1/duplicates

JWT required. Paginated list of duplicate groups detected by hash +
title+year heuristic. Rate tier `READ`.

### POST /api/v1/duplicates/{groupId}/resolve

JWT required. Body `{action: "keep_largest"|"keep_newest"|"manual"}`
plus manual selection when `manual`. 200 → resolution record.
Rate tier `WRITE`.

### GET /api/v1/collections

JWT required. User's collections. Paginated. Rate tier `READ`.

### POST /api/v1/collections

JWT required. Body `{name, description, items[]}`. 201 → collection id.
Rate tier `WRITE`.

### PUT /api/v1/collections/{id}

JWT required. Partial update. Rate tier `WRITE`.

### DELETE /api/v1/collections/{id}

JWT required. 204 on success. Cascade-deletes collection_items.
Rate tier `WRITE`.

### POST /api/v1/collections/{id}/items

JWT required. Body `{media_item_id}`. 201 on add.
Rate tier `WRITE`.

### GET /api/v1/favorites

JWT required. Paginated list of favorited media items. Rate tier `READ`.

### POST /api/v1/scan

JWT required. Body `{protocol, root, credentials?}`. Triggers
`UniversalScanner`; emits `scan.started` WS event. Returns
`{scanId}`. Rate tier `SCAN`.

### GET /api/v1/scan/{scanId}/status

JWT required. Returns progress snapshot. Used by clients that lose WS
connection. Rate tier `READ`.

### POST /api/v1/scan/{scanId}/cancel

JWT required. Signals the scanner context; scanner cleans up and
emits `scan.failed` with stage=cancelled. Rate tier `WRITE`.

### GET /api/v1/cover/{id}

JWT required. Returns the cached cover image for a media item. Falls
through to image-proxy if cache miss. Response: binary image
(image/jpeg, image/png). On 404 returns a 1×1 placeholder, not 404,
to keep `<img>` tags from firing `onError`. Rate tier `READ`.

### GET /api/v1/image-proxy

JWT required. Query `url` (TMDB / OMDB / OpenLibrary CDN). Server
fetches + re-encodes + serves with cache headers. SSRF-guarded by
`Security` submodule. Rate tier `READ`.

### GET /api/v1/catalog/roots

JWT required. Configured storage roots. Rate tier `READ`.

### POST /api/v1/catalog/scan-dir

JWT required. Body `{path}`. Scans a specific directory under a
configured root. Rate tier `SCAN`.

### GET /api/v1/realtime/status

JWT required. Returns backend-side WebSocket fan-out state — number of
live clients, per-user connection count. Rate tier `READ`.

### POST /api/v1/realtime/sync

JWT required. Body `{since_seq: N}`. Server replays missed events.
Rate tier `READ`.

### GET /api/v1/challenges

JWT-ADMIN. List all registered challenges. Rate tier `ADMIN`.

### POST /api/v1/challenges/{id}/run

JWT-ADMIN. Triggers a single challenge. Progress-based liveness (5-min
stale → kill). Returns `{runId, status}`. Rate tier `ADMIN`.

### POST /api/v1/challenges/run-all

JWT-ADMIN. Synchronous blocking. No other challenge may run until this
completes. Requires config `write_timeout=900`. Rate tier `ADMIN`.

### GET /api/v1/challenges/results

JWT-ADMIN. Paginated historical results. Default `limit=100` (enforced
by handler, not client — added in FIX-QA-2026-04-20 to prevent
unbounded memory reads). Rate tier `READ`.

### GET /api/v1/admin/config

JWT-ADMIN. Returns current effective configuration (with secrets
redacted). Alias: `/admin/config`. Rate tier `ADMIN`.

### GET /api/v1/admin/errors

JWT-ADMIN. Recent server errors with stack traces. Rate tier `ADMIN`.

### GET /api/v1/admin/health

JWT-ADMIN. Deep health check (DB, Redis, filesystem backends,
external metadata providers). Rate tier `ADMIN`.

### GET /api/v1/admin/logs

JWT-ADMIN. Recent structured log lines. Filter by level, service,
since-timestamp. Rate tier `ADMIN`.

### GET /ws

JWT required (passed as `?token=` query). Upgrades to WebSocket. See
§6 for event contract. No rate limiting on frames after upgrade — but
backpressure applies; slow consumers are disconnected with code 4008.

## 11. Client Health Checks

Every client MUST check the backend reachable before offering UI:

```bash
curl -fsS --max-time 3 $BASE_URL/api/v1/health
# → {"data":{"status":"ok","version":"2.3.0"},"meta":{...}}
```

If the check fails, clients MUST:
1. Display a backend-unreachable banner
2. Queue mutations for replay when connectivity returns
3. Serve read requests from local cache when possible
4. Never hang the UI while waiting
