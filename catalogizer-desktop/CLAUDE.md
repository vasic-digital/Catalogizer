# CLAUDE.md - Catalogizer Desktop

## Overview

Tauri 2 desktop application for browsing and managing Catalogizer media collections. React 18 frontend with a Rust backend that handles configuration persistence, HTTP proxying (with SSRF protection), and platform detection via IPC commands.

**Identifier**: `com.catalogizer.desktop` (Tauri 2 / Rust + React 18 / TypeScript / Vite)

## Build & Test

```bash
# Frontend
npm install
npm run dev                # Vite dev server (:1420)
npm run build              # tsc + vite build
npm run test               # vitest run
npm run test:watch         # vitest --watch
npm run test:coverage      # vitest --coverage

# Tauri (full desktop app)
npm run tauri:dev          # dev mode with hot reload
npm run tauri:build        # production build (AppImage/dmg/msi)

# Rust backend only
cd src-tauri
cargo test                 # unit tests
cargo build                # debug build
```

## Code Style

- **TypeScript**: strict mode, PascalCase components, camelCase functions. Tailwind CSS for styling
- **Rust**: edition 2021, `snake_case` functions, `PascalCase` structs. Serde for serialization
- Imports grouped: React/framework, third-party, internal
- Tests: Vitest + React Testing Library (frontend), `#[cfg(test)]` modules (Rust)

## Directory Structure

| Path | Purpose |
|------|---------|
| `src/pages/` | Route pages: Home, Library, Login, MediaDetail, Search, Settings |
| `src/components/` | Shared UI: Layout, LoadingScreen |
| `src/services/apiService.ts` | HTTP client for catalog-api communication |
| `src/stores/` | Zustand stores: `authStore`, `configStore` |
| `src/types/index.ts` | Media, auth, playback, config type definitions |
| `src/utils/cn.ts` | Tailwind class merge utility |
| `src/test-utils/` | Test helpers, mock data, custom render |
| `src-tauri/src/main.rs` | Rust backend: IPC commands, config state, HTTP proxy |
| `src-tauri/tauri.conf.json` | Tauri app config (window 1200x800, CSP, bundle) |
| `src-tauri/Cargo.toml` | Rust dependencies (tauri 2, reqwest, tokio, serde) |

## Key IPC Commands (Rust)

- `get_config` / `update_config` -- Read/write full `AppConfig` (server_url, auth_token, theme, auto_start)
- `set_server_url` / `set_auth_token` / `clear_auth_token` -- Granular config mutations
- `make_http_request` -- Proxied HTTP with SSRF validation (URL must match configured server)
- `get_app_version` / `get_platform` / `get_arch` -- System info

## Key Frontend Types

- `MediaItem`, `MediaVersion`, `ExternalMetadata` -- Media domain
- `User`, `LoginRequest`, `LoginResponse`, `AuthStatus` -- Authentication
- `AppConfig`, `SMBConfig`, `PlaybackProgress`, `DownloadJob` -- App state
- `MediaType`, `QualityLevel`, `SortOption`, `Theme` -- Union types

## Dependencies

- **Frontend**: React 18, React Router 6, React Query 4, Zustand, Tailwind CSS, Lucide icons
- **Rust**: tauri 2, reqwest 0.11, tokio, serde, env_logger

## Commit Style

Conventional Commits: `feat(desktop): add media detail view`
