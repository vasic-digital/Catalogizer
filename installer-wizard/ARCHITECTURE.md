# Architecture -- installer-wizard

## Purpose

Tauri 2 desktop wizard for configuring Catalogizer storage sources. Guides users step-by-step through network scanning, protocol selection (SMB, FTP, NFS, WebDAV, Local), connection testing, and configuration file generation.

## Structure

```
src/
  components/
    wizard/            Step components: Welcome, ProtocolSelection, NetworkScan, SMB/FTP/NFS/WebDAV/Local config, ConfigurationManagement, Summary
    ui/                Reusable UI primitives: Button, Card, Input
    layout/            WizardLayout -- step navigation chrome
  contexts/            WizardContext (step state), ConfigurationContext (source/access state)
  services/tauri.ts    Tauri IPC bridge for Rust commands
  types/index.ts       Network, config, wizard, per-protocol connection types
src-tauri/
  src/main.rs          Rust entry: IPC command registration, domain structs
  src/network.rs       Network scanning (trust-dns, ipnetwork)
  src/smb.rs           SMB share scanning, browsing, connection testing
  src/ftp.rs           FTP connection testing
  src/nfs.rs           NFS connection testing
  src/webdav.rs        WebDAV connection testing
  src/local.rs         Local filesystem validation
```

## Key Components

- **Wizard steps** -- Welcome -> ProtocolSelection -> NetworkScan -> Protocol-specific config -> Summary
- **IPC commands (Rust)** -- scan_network, scan_smb_shares/browse_smb_share/test_smb_connection, test_ftp/nfs/webdav/local_connection, load/save_configuration, get_default_config_path
- **WizardContext** -- Manages current step and navigation
- **ConfigurationContext** -- Manages selected protocol, connection settings, and access configuration
- **React Hook Form + Zod** -- Form validation for connection settings

## Data Flow

```
WizardLayout -> current step component
    |
    NetworkScan: invoke("scan_network") -> Rust trust-dns resolver -> Vec<NetworkHost>
    |
    SMBConfig: invoke("scan_smb_shares", {host, user, pass}) -> Rust SMB2 -> Vec<SMBShare>
    |
    TestConnection: invoke("test_*_connection", config) -> Rust protocol test -> success/error
    |
    Summary: invoke("save_configuration", config) -> Rust JSON file write -> ~/.catalogizer/config.json
```

## Dependencies

- **Frontend**: React 18, React Router 6, React Query 4, React Hook Form, Zod, Zustand, Tailwind CSS
- **Rust**: tauri 2 + plugins (shell, dialog, fs), reqwest, tokio, trust-dns-resolver, ipnetwork, network-interface

## Testing Strategy

Vitest + React Testing Library for frontend (30 tests, 93% coverage). Rust `#[cfg(test)]` with `#[tokio::test]` for backend tests (struct serialization, local connection validation, config path resolution).
