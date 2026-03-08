# CLAUDE.md - Catalogizer API Client

## Overview

`@catalogizer/api-client` is a cross-platform TypeScript client library for the Catalogizer REST API. It wraps HTTP (via Axios) and WebSocket communication, providing typed service classes for authentication, media operations, and SMB management. Used by catalog-web, catalogizer-desktop, and installer-wizard.

**Package**: `@catalogizer/api-client` (TypeScript, Node.js)

## Build & Test

```bash
npm install
npm run build        # tsc (outputs to dist/)
npm run dev          # tsc --watch
npm run test         # vitest run
npm run lint         # eslint src --ext .ts
```

## Code Style

- TypeScript strict mode, target ES2020, CommonJS output
- PascalCase classes, camelCase functions/variables
- Imports grouped: Node.js builtins, third-party, internal
- Tests: Vitest with Node environment, co-located in `__tests__/` directories

## Package Structure

| Path | Purpose |
|------|---------|
| `src/index.ts` | `CatalogizerClient` main class and re-exports |
| `src/services/AuthService.ts` | Login, logout, register, token refresh, profile |
| `src/services/MediaService.ts` | Media CRUD, search, stats, playback progress |
| `src/services/SMBService.ts` | SMB config management, status, scanning |
| `src/utils/http.ts` | `HttpClient` -- Axios wrapper with auth token, retry, refresh |
| `src/utils/websocket.ts` | `WebSocketClient` -- reconnecting WebSocket with typed events |
| `src/types/index.ts` | All type/interface definitions and custom error classes |

## Key Exports

- `CatalogizerClient` -- Main entry point; extends `EventEmitter` with typed events. Owns `auth`, `media`, `smb` service instances. Manages HTTP + WebSocket lifecycle via `connect()`/`disconnect()`
- `HttpClient` -- Axios wrapper with token injection, automatic refresh, configurable retry
- `WebSocketClient` -- Auto-reconnecting WebSocket emitting `download:progress`, `scan:progress` events
- `AuthService` / `MediaService` / `SMBService` -- Domain service classes
- Error hierarchy: `CatalogizerError` > `AuthenticationError`, `NetworkError`, `ValidationError`
- Types: `ClientConfig`, `MediaItem`, `User`, `LoginRequest`, `SMBConfig`, `ClientEvents`, etc.

## Dependencies

- **Runtime**: `axios ^1.4.0`, `ws ^8.13.0`
- **Dev**: `typescript ^5.0`, `vitest`, `eslint`

## Commit Style

Conventional Commits: `feat(api-client): add collection endpoints`
