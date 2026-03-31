# Architecture -- catalogizer-api-client

## Purpose

Cross-platform TypeScript client library for the Catalogizer REST API. Wraps HTTP (via Axios) and WebSocket communication, providing typed service classes for authentication, media operations, SMB management, and real-time updates. Used by catalog-web, catalogizer-desktop, and installer-wizard.

## Structure

```
src/
  index.ts                    CatalogizerClient main class and re-exports
  services/
    AuthService.ts            Login, logout, register, token refresh, profile
    MediaService.ts           Media CRUD, search, stats, playback progress
    SMBService.ts             SMB config management, status, scanning
  utils/
    http.ts                   HttpClient -- Axios wrapper with auth token, retry, refresh
    websocket.ts              WebSocketClient -- reconnecting WebSocket with typed events
  types/index.ts              All type/interface definitions and custom error classes
```

## Key Components

- **`CatalogizerClient`** -- Main entry point extending EventEmitter. Owns auth, media, smb service instances. Manages HTTP + WebSocket lifecycle via connect()/disconnect()
- **`HttpClient`** -- Axios wrapper with bearer token injection, automatic 401 refresh, and configurable retry
- **`WebSocketClient`** -- Auto-reconnecting WebSocket emitting download:progress and scan:progress events
- **`AuthService`** / **`MediaService`** / **`SMBService`** -- Domain service classes for API endpoints
- **Error hierarchy** -- CatalogizerError > AuthenticationError, NetworkError, ValidationError

## Data Flow

```
CatalogizerClient.connect(credentials)
    |
    AuthService.login() -> store token -> HttpClient injects Authorization header
    |
    WebSocketClient.connect() -> listen for download:progress, scan:progress events
    |
    MediaService.search(query) -> HttpClient.get("/api/v1/media/search", params)
        |
        Axios interceptor: inject token -> 401? auto-refresh -> retry
        |
        error -> classify: 400=ValidationError, 401=AuthenticationError, network=NetworkError
```

## Dependencies

- `axios` -- HTTP client
- `ws` -- WebSocket client (Node.js)
- `vitest` for testing

## Testing Strategy

Vitest with Node environment. Service tests verify request construction, URL building, token injection, retry behavior, and error classification. Co-located in `__tests__/` directories.
