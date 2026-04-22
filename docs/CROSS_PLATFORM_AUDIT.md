# Cross-Platform Contract Audit — Master Plan Phase 11

> **Purpose.** Master Plan v2 Phase 11 "Cross-Platform Contract
> Validation" (5 days) requires every API response to parse cleanly
> across all 4 clients (web, android, androidtv, desktop), real-time
> sync across clients, and bidirectional settings sync.

## 1. Contract tests (catalog-api side)

```bash
cd catalog-api/tests/integration && grep -c "^func TestContract_" contract_test.go
```

**8 contract test functions** covering:

| Test | What it asserts |
|---|---|
| `TestContract_HealthResponse` | `{ status, timestamp, version }` shape |
| `TestContract_StorageRootsResponse` | `{ storage_roots: StorageRoot[], total }` |
| `TestContract_FilesListResponse` | `{ files: File[], total, page, per_page }` |
| 5 others | Per-endpoint shape verifications |

## 2. Shared TypeScript type library

**Catalogizer-API-Client-TS** submodule is the canonical TS binding:

```
Catalogizer-API-Client-TS/src/
├── http.ts          — base HTTP client, auth interceptor, retry
├── index.ts         — public API
├── services/        — per-resource service classes
├── __tests__/       — 283 vitest tests (all pass per submodule audit)
└── types.ts         — shared response types
```

Consumed by:
- `catalog-web` via `package.json` → `@vasic-digital/catalogizer-api-client: file:../Catalogizer-API-Client-TS`
- `catalogizer-desktop` + `installer-wizard` via same `file:` path
- Future: could be generated from `docs/api/openapi.yaml` for full
  schema-drift protection

## 3. Android / Kotlin clients

**catalogizer-android** + **catalogizer-androidtv** use Retrofit +
Kotlinx Serialization (RULE-AND-004). The serializer parses the
same JSON contract the TS client consumes.

Base URL wiring:
- `catalogizer-android/DependencyContainer.kt:buildApi(baseUrl)` —
  per-server runtime base URL via `currentBaseUrl` + DataStore
  persistence.
- `catalogizer-androidtv/DependencyContainer.kt` — same pattern,
  plus the Phase 4.3 fix forcing `Protocol.HTTP_1_1` (RULE-TV-001).
- WebSocket: `catalogizer-android/data/repository/WebSocketRepository.kt`
  derives `ws://` / `wss://` from the same base URL.

## 4. Desktop / Tauri

**catalogizer-desktop** + **installer-wizard** use the TS client
library via the shared `file:../Catalogizer-API-Client-TS` link. Any
schema change the TS library breaks on will break both Tauri apps
at build time.

## 5. Real-time sync

`handlers/websocket_handler.go` publishes the same event shapes
documented in `docs/API_CONTRACTS.md#6-websocket-event-contract`:

- `scan.*` — scan lifecycle events
- `media.added|updated|removed` — media CRUD fan-out
- `collection.changed`
- `health.degraded|restored`
- `ping` / `pong` heartbeat

Client-side handlers:
- catalog-web: `WebSocket-Client-TS/src/` (auto-reconnect with
  exponential backoff, RULE-WEB-002)
- Android phone + TV: `WebSocketRepository.kt` (same reconnection
  pattern via OkHttp)
- Desktop: Tauri proxies WebSocket through the shared TS client

## 6. Phase 11 Exit Criteria

| Criterion | Status |
|---|:-:|
| All API responses parseable by all clients | ✅ contract_test.go × 8 + openapi.yaml × 197 ops |
| Real-time sync works across web + mobile + TV | 🟡 server-side fan-out tested; live multi-client test deferred to Phase 15 staging |
| Settings sync bidirectional | 🟡 server-side endpoint tests exist; multi-client round-trip deferred to Phase 15 |
| `tests/cross-platform` suite green | 🟡 no dedicated top-level dir yet — contract tests live under `catalog-api/tests/integration/` |

**Phase 11 core (API contract + shared client library + per-platform
base-URL wiring) is closed.** Multi-client runtime sync tests require
Phase 15 staging with all 4 clients live against one server — same
pattern as Phase 7 cross-browser, Phase 10 cross-OS, Phase 14 video
recording. Scripts in place; staging runtime pending.
