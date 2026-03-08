# CLAUDE.md - Catalogizer Installer Wizard

## Overview

Tauri 2 desktop wizard for configuring Catalogizer storage sources. Guides users through network scanning, protocol selection, and connection testing for SMB, FTP, NFS, WebDAV, and local filesystems. Saves/loads JSON configuration files.

**Identifier**: `com.catalogizer.installer-wizard` (Tauri 2 / Rust + React 18 / TypeScript / Vite)

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
npm run tauri:build        # production build

# Rust backend only
cd src-tauri
cargo test                 # unit tests (struct serialization, local connection, config path)
cargo build                # debug build
```

## Code Style

- **TypeScript**: strict mode, PascalCase components, camelCase functions. Tailwind CSS, React Hook Form + Zod
- **Rust**: edition 2021, `snake_case` functions, `PascalCase` structs. `anyhow`/`thiserror` for errors
- Tests: Vitest + React Testing Library (frontend), `#[cfg(test)]` with `#[tokio::test]` (Rust)

## Directory Structure

| Path | Purpose |
|------|---------|
| `src/components/wizard/` | Step components: Welcome, ProtocolSelection, NetworkScan, SMB/FTP/NFS/WebDAV/Local config, ConfigurationManagement, Summary |
| `src/components/ui/` | Reusable UI primitives: Button, Card, Input |
| `src/components/layout/` | `WizardLayout` -- step navigation chrome |
| `src/contexts/` | `WizardContext` (step state), `ConfigurationContext` (source/access state) |
| `src/services/tauri.ts` | Tauri IPC bridge for Rust commands |
| `src/types/index.ts` | Network, config, wizard, per-protocol connection types |
| `src-tauri/src/main.rs` | Rust entry: IPC command registration, domain structs |
| `src-tauri/src/network.rs` | Network scanning (trust-dns, ipnetwork) |
| `src-tauri/src/smb.rs` | SMB share scanning, browsing, connection testing |
| `src-tauri/src/ftp.rs` | FTP connection testing |
| `src-tauri/src/nfs.rs` | NFS connection testing |
| `src-tauri/src/webdav.rs` | WebDAV connection testing |
| `src-tauri/src/local.rs` | Local filesystem validation |

## Key IPC Commands (Rust)

- `scan_network` -- Discover hosts on local network (returns `Vec<NetworkHost>`)
- `scan_smb_shares` / `browse_smb_share` / `test_smb_connection` -- SMB operations
- `test_ftp_connection` / `test_nfs_connection` / `test_webdav_connection` / `test_local_connection` -- Protocol testers
- `load_configuration` / `save_configuration` -- JSON config file I/O
- `get_default_config_path` -- `~/.catalogizer/config.json`

## Key Frontend Types

- `NetworkHost`, `SMBShare`, `FileEntry` -- Network discovery
- `SMBConnectionConfig`, `FTPConnectionConfig`, `NFSConnectionConfig`, `WebDAVConnectionConfig`, `LocalConnectionConfig` -- Per-protocol config
- `Configuration`, `ConfigurationAccess`, `ConfigurationSource` -- Persisted config model
- `WizardState`, `StepProps`, `WizardStep` -- Wizard flow control
- `ConfigValidationResult`, `ValidationError` -- Form validation

## Dependencies

- **Frontend**: React 18, React Router 6, React Query 4, React Hook Form, Zod, Zustand, Tailwind CSS, @catalogizer/api-client
- **Rust**: tauri 2 (+ plugins: shell, dialog, fs), reqwest, tokio, trust-dns-resolver, ipnetwork, network-interface
- **Tauri plugins**: shell (open), dialog (file picker), fs (file access)

## Commit Style

Conventional Commits: `feat(wizard): add WebDAV configuration step`
