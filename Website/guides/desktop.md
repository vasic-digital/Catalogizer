---
title: Desktop App Guide
description: Using the Catalogizer desktop application on Windows, macOS, and Linux
---

# Desktop App Guide

The Catalogizer desktop application is a native cross-platform app built with Tauri (Rust backend + React frontend). It provides the full Catalogizer experience as a standalone desktop application with system integration features.

---

## Supported Platforms

| Platform | Installer Format |
|----------|-----------------|
| Windows | MSI installer |
| macOS | DMG disk image |
| Linux | AppImage or .deb package |

Download the appropriate installer from the [Download](/download) page.

---

## Installation

### Windows
1. Download the `.msi` installer
2. Double-click to run the installer
3. Follow the setup wizard prompts
4. Launch Catalogizer from the Start Menu or desktop shortcut

### macOS
1. Download the `.dmg` disk image
2. Open the DMG and drag Catalogizer to your Applications folder
3. On first launch, right-click and select Open to bypass Gatekeeper
4. Catalogizer appears in Launchpad and Applications

### Linux
**AppImage:**
1. Download the `.AppImage` file
2. Make it executable: `chmod +x Catalogizer-*.AppImage`
3. Run it: `./Catalogizer-*.AppImage`

**Debian/Ubuntu (.deb):**
1. Download the `.deb` package
2. Install: `sudo dpkg -i catalogizer_*.deb`
3. Launch from your application menu

---

## First Launch

On first launch, the app prompts you to connect to a Catalogizer server.

1. Enter the server URL (e.g., `https://your-server:8080`)
2. Log in with your username and password
3. The app stores your credentials securely in the system keychain

If you are running the server locally, use `http://localhost:8080` or the port shown in the `.service-port` file.

---

## Features

The desktop app provides the same functionality as the web application plus native integrations.

### Media Browsing
- Browse your catalog with grid, list, and detail views
- Filter by media type, quality, source, and year
- Full-text search across all metadata
- Entity hierarchy navigation (TV shows, music albums)

### Media Playback
- Built-in player for video and audio
- Subtitle support with track selection
- Resume from last position
- Playlist playback with auto-advance

### Collections and Organization
- Create and manage Manual, Smart, and Dynamic collections
- Favorites with export/import
- Playlist creation and reordering

### System Integration
- **System tray**: Minimize to tray for background operation
- **Native notifications**: OS-level notifications for scan completions and new media
- **Keyboard shortcuts**: Standard shortcuts for navigation and playback controls
- **File associations**: Open supported media files directly in Catalogizer

---

## Server Connection

The desktop app communicates with the Catalogizer backend API over HTTP/3 (QUIC) with Brotli compression, falling back to HTTP/2 when HTTP/3 is unavailable.

### Connection Settings

Access connection settings from the menu bar or settings page:

- **Server URL**: The address of your Catalogizer server
- **Auto-reconnect**: Automatically reconnects if the connection drops
- **Connection timeout**: Configurable timeout for API requests

### Multiple Servers

You can configure multiple server profiles and switch between them. This is useful if you manage more than one Catalogizer instance (e.g., home and office).

---

## Installer Wizard

A separate Installer Wizard application (`installer-wizard`) guides first-time setup.

- Discovers SMB devices on your local network automatically
- Provides a visual interface for configuring storage sources
- Tests connections in real time before saving
- Exports configuration files for the Catalogizer server
- Useful for non-technical users setting up their first instance

The wizard is also built with Tauri and is available as a standalone download.

---

## Updates

Check for updates from the Help menu. The app displays a notification when a new version is available. Download and install the updated package from the [Download](/download) page.

---

## Troubleshooting

**App does not start on Linux:**
Ensure required system libraries are installed. For AppImage, verify FUSE is available or set `APPIMAGE_EXTRACT_AND_RUN=1` as an environment variable.

**Cannot connect to server:**
Verify the server URL and that the backend is running. Check firewall rules for the server port. Try accessing the health endpoint in a browser: `http://your-server:8080/health`.

**Blank screen after login:**
Clear the app cache from Settings or delete the app data directory and restart.
