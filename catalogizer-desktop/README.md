# Catalogizer Desktop

Cross-platform desktop application built with Tauri (Rust + React).

## Tech Stack

- **Tauri 2.0** for native desktop shell
- **React 18** with TypeScript
- **Vite** for frontend build
- **TanStack Query** for server state
- **Zustand** for client state
- **Tailwind CSS** for styling
- **Vitest** for testing

## Prerequisites

- **Node.js** 18+
- **Rust** (latest stable)
- **Tauri CLI**: `cargo install tauri-cli`

### Platform-specific requirements:

**Linux:**
```bash
sudo apt install libwebkit2gtk-4.1-dev build-essential curl wget libssl-dev libgtk-3-dev libayatana-appindicator3-dev librsvg2-dev
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
- Visual Studio Build Tools with C++ workload
- WebView2 (usually pre-installed on Windows 10/11)

## Quick Start

```bash
# Install dependencies
npm install

# Start development (opens desktop app with hot reload)
npm run tauri:dev

# Build production app
npm run tauri:build
```

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start Vite dev server (web only) |
| `npm run tauri:dev` | Start Tauri development |
| `npm run tauri:build` | Build production desktop app |
| `npm run test` | Run tests |
| `npm run test:watch` | Run tests in watch mode |
| `npm run build` | Build frontend only |

## Project Structure

```
src/                 # React frontend
├── components/      # UI components
├── hooks/           # Custom hooks
├── lib/             # Tauri API bindings
└── pages/           # Application pages

src-tauri/           # Rust backend
├── src/             # Rust source code
├── Cargo.toml       # Rust dependencies
└── tauri.conf.json  # Tauri configuration
```

## Build Output

Production builds are created in `src-tauri/target/release/bundle/`:
- **Linux**: `.deb`, `.AppImage`
- **macOS**: `.dmg`, `.app`
- **Windows**: `.msi`, `.exe`

## Related Documentation

- [Tauri IPC Guide](/docs/architecture/TAURI_IPC_GUIDE.md)
- [Desktop Guide](/docs/guides/DESKTOP_GUIDE.md)
