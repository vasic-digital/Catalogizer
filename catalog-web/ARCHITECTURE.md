# Architecture -- catalog-web

## Purpose

React 18 / TypeScript / Vite frontend for Catalogizer. Provides media browsing, search, collection management, dashboards, media playback, and real-time WebSocket updates. Runs on port 3000 and proxies `/api` requests to the catalog-api backend.

## Structure

```
src/
  components/     Reusable UI components
  pages/          Route-level page components
  hooks/          Custom React hooks
  lib/            Utility functions and helpers
  types/          TypeScript type definitions
  contexts/       React context providers (Auth, WebSocket)
  assets/         Static assets (images)
  __tests__/      Test files
e2e/              Playwright E2E test specs
```

## Key Components

- **`AuthProvider`** -- from @vasic-digital/auth-context; wraps app with authentication state
- **`WebSocketProvider`** -- from @vasic-digital/websocket-client; provides real-time updates
- **`ProtectedRoute`** -- Auth-gated route component
- **React Query** -- Server state management for all API data
- **Zustand** -- Client-side state management
- **9 local submodule dependencies** -- auth-context, websocket-client, ui-components, media-types, catalogizer-api-client, media-browser, media-player, collection-manager, dashboard-analytics

## Data Flow

```
App mount -> AuthProvider (check auth status) -> WebSocketProvider (connect WS)
    |
    Router -> ProtectedRoute -> Page component
        |
        useQuery() -> CatalogizerClient.entities.browseByType() -> /api/v1/entities/browse/:type
        |                                                              |
        React Query cache                                    Vite proxy -> catalog-api backend
        |
        WebSocket events -> real-time UI updates (scan progress, new files)
```

## Dependencies

- React 18, TypeScript, Vite 6, React Router DOM v6
- @tanstack/react-query (server state), Zustand (client state)
- Tailwind CSS, framer-motion, Recharts
- React Hook Form + Zod (forms)
- Vitest (unit tests), Playwright (E2E tests)

## Testing Strategy

Vitest with jsdom for unit tests (823+ tests). Playwright for E2E tests. Coverage via @vitest/coverage-v8. Zero-warning ESLint policy with `--max-warnings 0`.
