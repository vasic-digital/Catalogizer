# catalog-web Architecture

## Overview

React 18 single-page application built with TypeScript and Vite. Serves as the primary web interface for Catalogizer media management, providing media browsing, collections, analytics, and playback.

## Component Hierarchy

```
ErrorBoundary
  AuthProvider (JWT auth context)
    WebSocketProvider (real-time events)
      BrowserRouter
        ConnectionStatus (global WS indicator)
        Suspense (code-split loading)
          Routes
            /login, /register          -- public
            Layout (sidebar + header)
              ProtectedRoute            -- auth gate with optional permission check
                PageErrorBoundary
                  <Page />              -- lazy-loaded page component
```

All protected routes are wrapped in `ProtectedRoute`, which supports `requiredPermission` strings (e.g., `read:media`, `manage:subtitles`) and `requireAdmin` for admin-only pages.

## State Management

| Layer | Tool | Scope |
|-------|------|-------|
| Server state | React Query (`@tanstack/react-query`) | API data fetching, caching, mutations |
| Client state | Zustand | UI preferences, transient state |
| Auth state | React Context (`AuthContext`) | JWT token, user info, permissions |
| Real-time | React Context (`WebSocketContext`) | WebSocket connection, event subscriptions |
| Form state | React Hook Form + Zod | Validated form inputs |

## Routing

React Router v6 with `createBrowserRouter` semantics. The `Layout` component renders a shared sidebar and header with an `<Outlet />` for nested page content. 15 protected routes cover Dashboard, Media Browser, Entity Browser, Analytics, Collections, Favorites, Playlists, Conversion Tools, Subtitles, Settings, AI Dashboard, and Admin.

All page components are lazy-loaded via `React.lazy()` for automatic code splitting.

## Data Flow

```
User Action -> React Hook Form (validate) -> API call (axios via lib/)
  -> React Query mutation -> cache invalidation -> re-render

WebSocket event -> WebSocketContext -> subscriber callbacks -> state update
```

API functions live in `src/lib/` (e.g., `mediaApi.ts`, `collectionsApi.ts`). Each module exports typed functions that call the backend via axios. The Vite dev server proxies `/api` and `/ws` to the catalog-api backend using a dynamically discovered port from `.service-port`.

## Submodule Dependencies

Nine `@vasic-digital/*` packages are linked via `file:../` in `package.json`:

- **auth-context** -- Auth context provider
- **websocket-client** -- WebSocket with reconnection + React hooks
- **ui-components** -- Shared UI component library
- **media-types** -- Shared TypeScript type definitions
- **catalogizer-api-client** -- Typed API client
- **media-browser** -- Media browsing components
- **media-player** -- Media playback components
- **collection-manager** -- Collection management UI
- **dashboard-analytics** -- Dashboard and analytics widgets

## Build Output

Production builds split into vendor chunks: `vendor` (react/react-dom), `router`, `ui` (framer-motion, lucide, headlessui), `charts` (recharts), `utils` (axios, date-fns, clsx). Output directory is `dist/` with sourcemaps enabled.

## Key Design Decisions

- **Code splitting** via `React.lazy` keeps initial bundle small; each page loads on demand.
- **Error boundaries** at two levels: global `ErrorBoundary` and per-page `PageErrorBoundary` for graceful degradation.
- **SplashScreen** gate delays router mount until branding animation completes.
- **Zero-warning policy**: ESLint enforces `--max-warnings 0`. No console errors or failed network requests in any environment.
