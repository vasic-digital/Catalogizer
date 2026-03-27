# Catalogizer -- Desktop User Manual

## Table of Contents

1. [Installation](#installation)
2. [First Launch and Connecting to Server](#first-launch-and-connecting-to-server)
3. [File Browser and Media Library](#file-browser-and-media-library)
4. [Media Playback](#media-playback)
5. [Keyboard Shortcuts](#keyboard-shortcuts)
6. [System Tray Integration](#system-tray-integration)
7. [Auto-Update Mechanism](#auto-update-mechanism)
8. [Troubleshooting](#troubleshooting)

---

## Installation

Catalogizer Desktop is a Tauri-based application available for Windows, macOS, and Linux. The installer wizard (a separate application) handles initial server setup; the desktop app is for day-to-day media browsing and playback.

### Windows

1. Download the `.msi` installer from the releases page or your administrator.
2. Double-click the installer and follow the on-screen prompts. The default installation directory is `C:\Program Files\Catalogizer`.
3. Windows SmartScreen may display a warning for unsigned builds. Click **More info** and then **Run anyway** to proceed.
4. After installation, Catalogizer appears in the Start Menu and on the Desktop (if the shortcut option was selected).

### macOS

1. Download the `.dmg` disk image.
2. Open the disk image and drag the **Catalogizer** app to your **Applications** folder.
3. On first launch, macOS Gatekeeper may block the app. Open **System Preferences > Security & Privacy > General** and click **Open Anyway**.
4. Catalogizer appears in Launchpad and Spotlight search after installation.

### Linux

Catalogizer is distributed as an AppImage and as a `.deb` package:

- **AppImage**: Download the `.AppImage` file, make it executable (`chmod +x Catalogizer.AppImage`), and run it directly. No installation required.
- **Debian/Ubuntu**: Download the `.deb` package and install with `sudo dpkg -i catalogizer_*.deb`. Dependencies are resolved automatically if you run `sudo apt-get install -f` afterward.

On all platforms, the application requires approximately 150 MB of disk space.

---

## First Launch and Connecting to Server

### Server Connection Dialog

On first launch, the app presents a server connection dialog. Enter the URL of your Catalogizer API server, including the port:

```
https://192.168.0.100:8080
```

The app tests the connection by querying the server's health endpoint. A green indicator confirms a successful connection. The server URL is saved locally and used for all subsequent launches.

### Login

After connecting to the server, the login form appears. Enter your username and password. The app stores a JWT token in the system keychain (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux) for secure session persistence. You remain logged in across restarts until the token expires.

### Multiple Servers

You can connect to different Catalogizer servers by opening **File > Switch Server** (or the equivalent menu item on your platform). The app maintains a list of recently used servers for quick switching.

---

## File Browser and Media Library

The main window is divided into a sidebar and a content area.

### Sidebar

The sidebar provides navigation between the major sections:

| Section | Description |
|---------|-------------|
| **Dashboard** | Overview with recently added items, library statistics, and quick actions |
| **Browse** | Full entity browser with type filters and sorting |
| **Storage** | Direct access to storage roots and their directory trees |
| **Favorites** | Your favorited entities |
| **Playlists** | Your custom playlists |
| **Collections** | Shared and personal collections |
| **Search** | Advanced search with filters |

### Entity Browser

The Browse section displays entities in a grid or list layout. Use the toolbar at the top to:

- **Filter by type** -- Click type badges to toggle visibility of each media type.
- **Sort** -- Choose between title, year, date added, or rating.
- **Toggle view** -- Switch between grid (poster cards) and list (compact table rows).

Double-click an entity card to open its detail view. Right-click for a context menu with options: Play, Add to Favorites, Add to Playlist, Add to Collection, and Show in Storage.

### Storage Browser

The Storage section shows your configured storage roots as a directory tree. Expand directories to browse raw files and folders. Files that have been aggregated into entities display their entity metadata inline. Unaggregated files are shown in a muted style.

### Detail View

The detail view opens in the content area (or as a separate panel, depending on your layout preference). It shows the full entity metadata: poster, title, year, genres, rating, plot summary, cast and crew, and the linked file list. From here you can start playback, manage metadata, or add the entity to collections.

---

## Media Playback

### Built-in Player

Catalogizer Desktop includes a built-in media player powered by the system's native media framework (Media Foundation on Windows, AVFoundation on macOS, GStreamer on Linux).

1. Open an entity detail view.
2. Click the **Play** button, or double-click a specific file in the file list.
3. The player opens in the content area. Click the expand icon to enter fullscreen mode.

Player controls:

- **Play / Pause** -- Click the button or press Space.
- **Seek** -- Click anywhere on the progress bar, or drag the seek handle.
- **Volume** -- Adjust with the volume slider or mouse scroll wheel over the player.
- **Fullscreen** -- Press F11 or click the fullscreen icon.
- **Subtitles** -- Click the subtitle icon to select a track or load an external subtitle file.
- **Audio track** -- Click the audio icon to switch between audio streams.

### External Player Integration

If you prefer a dedicated media player, configure it in **Settings > Playback > External Player**. Provide the path to the player executable (e.g., VLC, mpv). When configured, the **Open Externally** button appears next to the Play button, launching the file in your chosen player.

### Picture-in-Picture

On macOS and Windows, the player supports picture-in-picture mode. Click the PiP icon in the player controls to detach the video into a floating window that stays on top while you browse the rest of the app.

---

## Keyboard Shortcuts

Catalogizer Desktop supports keyboard shortcuts for efficient navigation. All shortcuts can be viewed from **Help > Keyboard Shortcuts**.

### Global Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+F` / `Cmd+F` | Focus the search bar |
| `Ctrl+,` / `Cmd+,` | Open settings |
| `Ctrl+Q` / `Cmd+Q` | Quit the application |
| `F5` | Refresh the current view |
| `Ctrl+1` through `Ctrl+6` | Switch between sidebar sections |
| `Escape` | Close the current dialog or panel |

### Playback Shortcuts

| Shortcut | Action |
|----------|--------|
| `Space` | Play / pause |
| `Left Arrow` | Rewind 10 seconds |
| `Right Arrow` | Fast-forward 10 seconds |
| `Up Arrow` | Increase volume |
| `Down Arrow` | Decrease volume |
| `M` | Mute / unmute |
| `F11` | Toggle fullscreen |
| `S` | Cycle through subtitle tracks |
| `A` | Cycle through audio tracks |

### Browse Shortcuts

| Shortcut | Action |
|----------|--------|
| `Enter` | Open the selected entity detail |
| `Backspace` | Navigate back |
| `Ctrl+G` / `Cmd+G` | Toggle grid / list view |

---

## System Tray Integration

Catalogizer runs in the system tray (notification area on Windows/Linux, menu bar on macOS) when you close the main window. This allows background sync and quick access without keeping the full window open.

### Tray Icon Menu

Right-click (or click on macOS) the tray icon to access:

- **Open Catalogizer** -- Restore the main window.
- **Sync Now** -- Trigger an immediate metadata sync with the server.
- **Recently Added** -- Submenu showing the 5 most recently added entities. Click one to open it directly.
- **Settings** -- Open the settings window.
- **Quit** -- Exit the application completely.

### Notifications

The tray icon displays desktop notifications for:

- **New media detected** -- When the server completes a scan and new entities are found.
- **Sync complete** -- When a background metadata sync finishes.
- **Update available** -- When a new version of the app is ready to install.

Notification preferences can be configured in **Settings > Notifications**.

### Minimize to Tray

By default, closing the window minimizes the app to the system tray instead of quitting. To change this behavior, go to **Settings > General > Close action** and select either "Minimize to tray" or "Quit application".

---

## Auto-Update Mechanism

Catalogizer Desktop includes a built-in auto-update system powered by Tauri's updater module.

### How Updates Work

1. On launch (and periodically while running), the app checks the update server for new versions.
2. If an update is available, a notification appears with the version number and changelog summary.
3. Click **Update Now** to download and install the update. The app restarts automatically after installation.
4. Click **Later** to defer the update. You will be reminded on the next launch.

### Update Channels

- **Stable** -- Production releases, tested and verified (default).
- **Beta** -- Pre-release versions for early testing. Enable in **Settings > General > Update channel**.

### Manual Update Check

Open **Help > Check for Updates** to trigger an immediate update check regardless of the automatic schedule.

### Offline Environments

If the desktop app is deployed in an air-gapped environment without internet access, disable auto-update in **Settings > General > Auto-update** and distribute updates manually via the installer packages.

---

## Troubleshooting

### Application Won't Start

- **Windows**: Check that the Visual C++ Redistributable (2015-2022) is installed. Download it from Microsoft's website if missing.
- **macOS**: Ensure you are running macOS 10.15 (Catalina) or later. Open the app from Finder (not from the DMG directly).
- **Linux**: For AppImage issues, ensure FUSE is installed (`sudo apt install libfuse2` on Ubuntu). Alternatively, set `APPIMAGE_EXTRACT_AND_RUN=1` as an environment variable.

### Cannot Connect to Server

- Verify the server URL and port are correct.
- Ensure the Catalogizer API container is running on the server.
- Check that your firewall allows outbound connections to the server port.
- If the server uses self-signed TLS certificates, you may need to accept the certificate on first connection. The app prompts you to trust unknown certificates.

### Login Session Expires Frequently

- JWT tokens have a configurable expiration time set by the server. Ask your administrator to increase the token lifetime if sessions are too short.
- Ensure your system clock is accurate. Clock skew can cause premature token rejection.

### Video Playback Issues

- If video plays in an external player but not the built-in player, the issue is likely a missing codec. On Linux, install GStreamer plugins: `sudo apt install gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly gstreamer1.0-libav`.
- On Windows, install the K-Lite Codec Pack or similar for extended format support.
- If hardware acceleration causes artifacts, disable it in **Settings > Playback > Hardware acceleration**.

### High Memory Usage

- Large libraries with many posters can consume significant memory. Reduce the cache size in **Settings > Storage > Cache limit**.
- Close the storage browser when not in use, as expanding large directory trees loads metadata for all visible files.

### Tray Icon Not Visible (Linux)

- Some Linux desktop environments (e.g., GNOME) do not display tray icons by default. Install the **AppIndicator** extension for GNOME Shell, or use a desktop environment with native tray support (KDE Plasma, XFCE, Cinnamon).
