# Architecture -- catalogizer-desktop

## Purpose

Cross-platform desktop application for browsing and managing Catalogizer media collections. Built with Tauri 2 (Rust backend + React 18 frontend). The Rust backend handles configuration persistence, HTTP proxying with SSRF protection, and platform detection via IPC commands.

## Structure

```
src/
  pages/               Route pages: Home, Library, Login, MediaDetail, Search, Settings
  components/          Shared UI: Layout, LoadingScreen
  services/apiService.ts  HTTP client for catalog-api communication
  stores/              Zustand stores: authStore, configStore
  types/index.ts       Media, auth, playback, config type definitions
  utils/cn.ts          Tailwind class merge utility
  test-utils/          Test helpers, mock data, custom render
src-tauri/
  src/main.rs          Rust backend: IPC commands, config state, HTTP proxy
  tauri.conf.json      Tauri app config (window 1200x800, CSP, bundle)
  Cargo.toml           Rust dependencies (tauri 2, reqwest, tokio, serde)
```

## Key Components

- **IPC commands (Rust)** -- get_config/update_config, set_server_url/set_auth_token, make_http_request (with SSRF validation), get_app_version/get_platform/get_arch
- **React frontend** -- React Router 6 pages, React Query 4 for server state, Zustand for client state
- **SSRF protection** -- make_http_request validates that URL matches configured server
- **Tailwind CSS** -- Utility-first styling with Lucide icons

## Data Flow

```
React Page -> apiService -> Tauri IPC (invoke) -> Rust make_http_request
    |                                                    |
    Zustand stores (auth, config)                  reqwest HTTP -> catalog-api
    |                                                    |
    React Query cache                              validate URL matches server config (SSRF check)
```

## Dependencies

- **Frontend**: React 18, React Router 6, React Query 4, Zustand, Tailwind CSS, Lucide icons
- **Rust**: tauri 2, reqwest 0.11, tokio, serde, env_logger

## Testing Strategy

Vitest + React Testing Library for frontend component and store tests. Rust `#[cfg(test)]` modules for backend unit tests. Custom test-utils with mock data for consistent testing.
