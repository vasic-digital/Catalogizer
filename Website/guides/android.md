---
title: Android App Guide
description: Using the Catalogizer Android app to browse and manage your media collection on the go
---

# Android App Guide

The Catalogizer Android app brings your media catalog to your phone or tablet. Built with Kotlin and Jetpack Compose, it follows Material Design 3 guidelines and supports offline access to your library.

---

## Requirements

- Android 8.0 (API 26) or higher
- Network access to your Catalogizer server (Wi-Fi or mobile data)

---

## Installation

1. Download the APK from the [Download](/download) page
2. Enable "Install from unknown sources" in your device settings if prompted
3. Open the downloaded APK and tap Install
4. Launch Catalogizer from your app drawer

---

## First Setup

On first launch, the app guides you through initial configuration.

1. **Server URL**: Enter the address of your Catalogizer server (e.g., `https://your-server:8080`)
2. **Login**: Enter your username and password
3. **Sync settings**: Choose what to cache for offline access
4. **Done**: The app loads your catalog

Your credentials are stored securely in the Android Keystore.

---

## Browsing Your Catalog

The main screen displays your media catalog organized by type.

### Navigation

- **Home**: Dashboard with recent additions and quick stats
- **Browse**: Full catalog with filtering and sorting
- **Search**: Full-text search across all media metadata
- **Collections**: Your collections and playlists
- **Profile**: Account settings and app preferences

### Media Types

Browse by category using the tabs or chips at the top of the Browse screen:

- Movies, TV Shows, Music, Games, Software, Books, Comics

### Detail View

Tap any media item to see its full details:

- Cover art, title, year, and description
- Quality information (resolution, codec, bitrate)
- Associated files and storage source
- Play, favorite, and add-to-collection actions

---

## Media Playback

Play video and audio directly from the app.

- Stream from any connected storage source
- Playback position syncs with the server -- resume on any device
- Subtitle selection for video content
- Background audio playback for music
- Picture-in-picture mode for video on supported devices

---

## Offline Mode

The app caches metadata locally using a Room database so you can browse your catalog without a network connection.

### What Works Offline

- Browsing previously loaded catalog metadata
- Viewing media details and cover art
- Searching cached items
- Viewing collections, playlists, and favorites

### What Requires a Connection

- Streaming media playback
- Triggering scans
- Syncing new changes from the server
- Adding new storage sources

### Cache Settings

Configure caching behavior in Settings:

- **Wi-Fi only sync**: Only download metadata updates on Wi-Fi
- **Storage limit**: Set a maximum cache size (cover art and metadata)
- **Auto-sync interval**: How often the app checks for updates
- **Clear cache**: Remove all cached data and re-sync

---

## Sync

The app synchronizes with the server to keep your catalog up to date.

- **Automatic sync**: Runs on a configurable interval when the app is open
- **Pull-to-refresh**: Swipe down on any list to trigger a manual sync
- **Background sync**: Optional background sync when the app is not in the foreground
- **Conflict resolution**: Server data takes precedence; local changes (favorites, collections) are pushed to the server on next sync

---

## Notifications

The app can display notifications for:

- Scan completions on the server
- New media detected in your catalog
- Sync status updates

Configure notification preferences in Settings.

---

## Settings

Access settings from the Profile tab.

- **Server**: Change server URL or switch accounts
- **Cache**: Wi-Fi only, storage limit, auto-sync interval
- **Appearance**: Theme (light, dark, system default)
- **Notifications**: Enable or disable notification categories
- **About**: App version, server version, and open-source licenses

---

## Troubleshooting

**Cannot connect to server:**
Verify the server URL is correct and reachable from your device. Check that the backend is running and accessible on your network. If using HTTPS, ensure the certificate is trusted or add it to your device.

**Slow loading:**
Large catalogs may take time on first sync. Subsequent loads use cached data. Reduce the cache storage limit if your device has limited space.

**Playback fails:**
Verify the storage source is online and accessible from the server. Check that the file format is supported by your device. Try a different file to isolate the issue.
