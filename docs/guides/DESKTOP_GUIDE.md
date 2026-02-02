# Catalogizer Desktop App Guide

This guide covers the Catalogizer Desktop application, a Tauri-based app combining a React/TypeScript frontend with a Rust backend. It provides a native desktop experience for managing your media collection.

## Table of Contents

1. [Installation](#installation)
2. [Initial Setup](#initial-setup)
3. [Login](#login)
4. [Home Page](#home-page)
5. [Library](#library)
6. [Search](#search)
7. [Media Detail](#media-detail)
8. [Settings](#settings)
9. [IPC Integration](#ipc-integration)
10. [Common Workflows](#common-workflows)
11. [Troubleshooting](#troubleshooting)

---

## Installation

### Windows

1. Download the `.msi` installer from the Catalogizer releases page.
2. Run the installer and follow the prompts.
3. Catalogizer will be added to your Start Menu and optionally to your desktop.

### macOS

1. Download the `.dmg` file from the Catalogizer releases page.
2. Open the DMG and drag Catalogizer to your Applications folder.
3. On first launch, you may need to allow the app in **System Preferences > Security & Privacy** since it is not from the Mac App Store.

### Linux

1. Download the `.AppImage` or `.deb` package from the Catalogizer releases page.
2. For AppImage: make the file executable (`chmod +x Catalogizer.AppImage`) and run it.
3. For Debian/Ubuntu: install with `sudo dpkg -i catalogizer_*.deb`.

### Building from Source

```bash
cd catalogizer-desktop
npm install
npm run tauri:build
```

The built binary will be in `src-tauri/target/release/`.

---

## Initial Setup

When you first launch the desktop app, it needs to know where your Catalogizer server is.

### Server Configuration (Required)

If no server URL is configured, you are automatically redirected to the Settings page:

1. Enter your **Server URL** (e.g. `http://localhost:8080` or `https://catalogizer.example.com`).
2. Click **Test** to verify the connection. A success message appears if the server is reachable, or an error message explains what went wrong.
3. Click **Save Settings**.

You are then redirected to the Login page.

### What the App Stores Locally

The desktop app uses Tauri's secure storage to persist:

- Server URL
- Authentication token
- Theme preference
- Auto-start setting

This data is stored in your OS-specific application data directory and retrieved via IPC commands on startup.

---

## Login

After the server URL is configured, the Login page is shown.

1. Enter your **Username** and **Password**.
2. Click **Sign In**.
3. On success, you are redirected to the Home page and your auth token is stored locally.

If your stored token is still valid from a previous session, you bypass the login screen entirely.

---

## Home Page

The Home page is the main landing screen after login, providing an overview of your media library.

### Features

- Welcome message with library statistics
- Quick access to recently added media
- Summary cards showing total items, storage usage, and recent activity
- Navigation links to Library, Search, and Settings

---

## Library

The Library page is the primary media browsing interface.

### Search and Filters

At the top, a filter bar provides:

- **Search field** -- type to search by title, description, or keywords
- **Type filter** -- dropdown to filter by media type:
  - All Types
  - Movies
  - TV Shows
  - Music
  - Documentaries
  - Anime
- **Sort by** -- dropdown to sort results:
  - Recently Updated
  - Recently Added
  - Title
  - Year
  - Rating
  - File Size
- **Sort order** -- Ascending or Descending
- **View mode toggle** -- switch between Grid and List views

### Grid View

Media items are displayed as visual cards in a responsive grid (2 columns on small screens, up to 6 on large screens). Each card shows:

- Poster image (from TMDB/IMDB metadata, or a play icon placeholder)
- Title
- Year (if available)
- Star rating (if available)

Hover over a card to see a subtle zoom animation. Click to navigate to the Media Detail page.

### List View

Media items are displayed as horizontal rows with:

- Small poster thumbnail
- Title (with hover color change)
- Year, media type, and rating
- Description excerpt (up to 2 lines)

Click any row to navigate to the Media Detail page.

### Result Count

A text indicator shows how many items are displayed vs. the total (e.g. "Showing 50 of 1,247 items").

### Empty State

If no results match your search or filters, a message appears: "No media found. Try adjusting your search or filters."

### Loading State

While data is loading, animated placeholder cards (skeleton loading) are shown in a grid pattern.

---

## Search

The Search page provides a dedicated search interface.

### Features

- Full-text search across your media library
- Search results displayed in a list with media details
- Click on a result to navigate to the Media Detail page
- Back navigation to return to the previous page

---

## Media Detail

The Media Detail page shows comprehensive information about a single media item.

### Information Displayed

- Full title and year
- Media type and quality
- Description / synopsis
- File size and format information
- Rating from external providers
- Poster image (if available)
- External metadata (TMDB, IMDB data)

### Actions

- Navigate back to the Library or Search results
- View detailed metadata

---

## Settings

The Settings page allows you to configure the desktop client. It is accessible from the sidebar navigation or by clicking the settings icon.

### Server Configuration

- **Server URL** -- edit the URL of your Catalogizer server
- **Test Connection** -- verify the server is reachable; shows success/failure with detailed message

### Appearance

- **Theme** -- choose between Light, Dark, or System (follows your OS theme preference)

### Storage Configuration

View and manage storage sources that the server scans for media:

- A list of configured storage sources, each showing:
  - Storage path (e.g. `//server/share` or `/mnt/media`)
  - Date added
  - Delete button to remove the source
- **Add Storage Source** -- click to expand a form:
  - Storage path (e.g. SMB path or local directory)
  - Username (optional, for authenticated protocols)
  - Password (optional)
  - Click **Add Source** to save

Supported protocols include SMB, FTP, NFS, WebDAV, and Local filesystem paths.

### General

- **Auto-start** -- toggle to start Catalogizer automatically when your computer boots

### Saving

Click **Save Settings** to persist all changes. If you are authenticated, clicking Save navigates you back to the previous page. If not authenticated, it navigates to the Login page.

---

## IPC Integration

The desktop app uses Tauri's IPC (Inter-Process Communication) to bridge the React frontend with the Rust backend. This enables native OS capabilities that are not available in a standard web browser.

### How IPC Works

The React frontend calls Rust functions via `invoke()` from the `@tauri-apps/api/core` package. The Rust backend processes these calls and returns results.

### Key IPC Commands

| Command | Purpose |
|---------|---------|
| `get_config` | Retrieves stored configuration (server URL, auth token, theme, auto-start) |
| `save_config` | Persists configuration changes to the OS-specific app data directory |
| `test_connection` | Tests connectivity to the Catalogizer server from the native layer |

### Configuration Flow

On startup:

1. The app loads configuration via the `get_config` IPC command.
2. If an auth token is stored, it is set in the auth store.
3. If no server URL is configured, the user is redirected to Settings.
4. If not authenticated, the user is redirected to Login.

### Benefits of IPC Integration

- **Secure storage** -- credentials and tokens are stored using OS-level secure storage rather than browser localStorage
- **Auto-start** -- the Rust backend can register the app to start with the OS
- **Native file access** -- future features can leverage native filesystem access for direct media operations
- **System tray** -- potential for background operation and system tray integration

---

## Common Workflows

### Setting Up a New Desktop Installation

1. Launch the app -- you are redirected to Settings.
2. Enter your Catalogizer server URL and click **Test** to verify.
3. Configure your preferred theme.
4. Click **Save Settings**.
5. Log in with your credentials.
6. Browse your media library from the Home page.

### Browsing and Finding Media

1. Navigate to the Library page.
2. Use the search bar to type a title or keyword.
3. Apply type filters (e.g. only Movies).
4. Sort by your preferred criteria (e.g. Rating descending).
5. Click on an item to see full details.

### Managing Storage Sources

1. Go to Settings.
2. Scroll to the Storage Configuration section.
3. Click **Add Storage Source**.
4. Enter the path (e.g. `//nas-server/media`).
5. Optionally enter credentials.
6. Click **Add Source**.
7. The server will begin scanning the new source for media.

### Switching Themes

1. Go to Settings.
2. Under Appearance, select Light, Dark, or System.
3. Click **Save Settings**.
4. The theme changes immediately.

---

## Troubleshooting

### App Shows Blank White Screen

- Open the developer console (if available) and check for JavaScript errors.
- Ensure the app is up to date.
- Try clearing the app data and reconfiguring.

### Cannot Connect to Server

- Verify the server URL is correct (include http:// or https://).
- Use the **Test** button in Settings to diagnose the issue.
- Check that your firewall allows connections from the desktop app to the server port.
- If using HTTPS, ensure the server certificate is valid.

### Auto-Start Not Working

- Verify the auto-start toggle is enabled in Settings.
- On Linux, check that the desktop entry was created in `~/.config/autostart/`.
- On macOS, check System Preferences > Users & Groups > Login Items.
- On Windows, check Task Manager > Startup tab.

### Storage Sources Not Appearing

- Ensure you are logged in (storage configuration requires authentication).
- Check the error message displayed below the Storage Configuration section.
- Verify the server has the storage source configured correctly.

### Slow Media Loading

- Check your network connection to the server.
- If loading many items, try using filters to narrow results.
- The app caches API responses; refreshing the page re-fetches fresh data.
