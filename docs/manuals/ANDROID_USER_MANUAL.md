# Catalogizer -- Android User Manual

## Table of Contents

1. [Installation](#installation)
2. [First Launch and Login](#first-launch-and-login)
3. [Browsing Your Media Library](#browsing-your-media-library)
4. [Playing Media](#playing-media)
5. [Managing Favorites and Playlists](#managing-favorites-and-playlists)
6. [Offline Mode and Sync](#offline-mode-and-sync)
7. [Settings and Configuration](#settings-and-configuration)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Installing from APK

If you have received a direct APK file (for example, from an internal release or the project's releases page), follow these steps:

1. Transfer the `.apk` file to your Android device via USB, cloud storage, or direct download.
2. Open the file manager on your device and navigate to the APK.
3. Tap the APK file to begin installation. Android will prompt you to allow installation from unknown sources if this is the first time.
4. On Android 8+, you will be asked to grant the file manager (or browser) permission to install unknown apps. Toggle the permission on and return to the installation prompt.
5. Tap **Install** and wait for the process to complete.
6. Tap **Open** to launch Catalogizer immediately, or find it in your app drawer.

### Installing from Google Play Store

1. Open the Google Play Store on your device.
2. Search for **Catalogizer** by Vasic Digital.
3. Tap **Install** and wait for the download and installation to finish.
4. Tap **Open** to launch the app.

Play Store installations receive automatic updates. APK installations must be updated manually by installing the newer APK over the existing one.

### System Requirements

| Requirement | Minimum |
|-------------|---------|
| Android version | 8.0 (Oreo) / API 26 |
| RAM | 2 GB |
| Storage | 100 MB free for the app |
| Network | Wi-Fi or LAN access to Catalogizer server |

---

## First Launch and Login

When you open Catalogizer for the first time, you are presented with the login screen.

### Connecting to Your Server

1. Enter the **Server URL** in the field at the top of the login screen. This is the address of your Catalogizer API server, for example `https://192.168.0.100:8080` or `https://catalogizer.local:8080`.
2. The app validates the connection by performing a health check against the server. A green checkmark appears if the server is reachable.
3. If the server uses HTTP/3 (QUIC), the app negotiates the connection automatically. Fallback to HTTP/2 occurs transparently if QUIC is unavailable.

### Logging In

1. Enter your **username** and **password**. These are the credentials created by your administrator or during the installer wizard setup.
2. Tap **Log In**.
3. The app stores a JWT token locally for subsequent sessions. You remain logged in until the token expires or you log out manually.

### First-Time Sync

After login, the app performs an initial sync to download your media library metadata. Depending on library size, this may take a few seconds to a few minutes. A progress indicator shows the sync status. Media thumbnails and posters are downloaded on demand as you browse.

---

## Browsing Your Media Library

The home screen displays your media library organized by type. The bottom navigation bar provides access to the main sections of the app.

### Bottom Navigation

| Tab | Description |
|-----|-------------|
| **Home** | Dashboard with recently added items and quick access cards |
| **Browse** | Full media library browser with type filters |
| **Search** | Global search across all entities |
| **Favorites** | Your favorited items |
| **Profile** | Account settings, server connection, and app preferences |

### Media Browser

The Browse tab presents a grid of entity cards, each showing the poster image, title, year, and media type badge. Pull down to refresh the library from the server.

### Filtering and Sorting

Tap the filter icon in the top toolbar to access filtering options:

- **Media type** -- Show only movies, TV shows, music, games, books, or other types.
- **Year range** -- Restrict results to a specific time period.
- **Rating** -- Set a minimum rating threshold.

Tap the sort icon to change ordering. Available sort options: title (A-Z, Z-A), year (newest, oldest), date added, and rating.

### Searching

The Search tab provides a text field with instant results. Type a partial title and matching entities appear as you type. Results are grouped by media type for clarity. Tap any result to open the entity detail screen.

---

## Playing Media

Catalogizer supports playback of video and audio files directly from your storage roots over the network.

### Video Playback

1. Navigate to a movie or TV episode entity.
2. In the file list, tap the file you want to play.
3. The built-in video player opens with full playback controls: play/pause, seek bar, volume, and fullscreen toggle.
4. Swipe left/right on the player to seek forward/backward by 10 seconds.
5. Swipe up/down on the left side to adjust brightness, or on the right side to adjust volume.

The player supports common video formats including MKV, MP4, AVI, and WebM. Codec support depends on your device hardware. Hardware-accelerated decoding is used when available.

### Audio Playback

1. Navigate to a song or music album entity.
2. Tap a track to begin playback. The mini player appears at the bottom of the screen.
3. Tap the mini player to expand it to the full-screen player view with album art, track progress, and playback controls.
4. Supported formats include FLAC, MP3, AAC, OGG, and WAV.

### Subtitle Support

During video playback, tap the subtitle icon in the player controls to select from available subtitle tracks. If the video file contains embedded subtitles, they are listed automatically. External subtitle files (SRT, ASS, VTT) stored alongside the video are also detected.

### External Player

If you prefer to use a third-party player such as VLC or MX Player, long-press a file entry and select **Open with...** from the context menu. The file URL is passed to the external player for streaming playback.

---

## Managing Favorites and Playlists

### Favorites

Tap the heart icon on any entity card or detail screen to add the item to your favorites. Favorited items appear in the Favorites tab for quick access. Favorites are synced to the server and available across all your devices.

To remove a favorite, tap the heart icon again on the entity card or in the Favorites tab.

### Playlists

Playlists allow you to create ordered lists of media for sequential playback or personal organization.

1. Navigate to the entity you want to add.
2. Tap the three-dot menu and select **Add to Playlist**.
3. Choose an existing playlist or tap **Create New Playlist** to start a fresh one.
4. Enter a name for the new playlist and tap **Create**.

To manage playlists, go to **Profile > My Playlists**. From there you can reorder items by drag-and-drop, remove individual items, rename the playlist, or delete it entirely.

---

## Offline Mode and Sync

Catalogizer supports offline access to your media library metadata and selected media files.

### Metadata Sync

Your library metadata (entity titles, posters, ratings, and hierarchy) is cached locally on the device using Room database storage. When the server is unreachable, you can still browse your library, view entity details, and manage favorites. Changes made offline (favorites, playlist edits) are queued and synced when connectivity is restored.

### Downloading for Offline Playback

To download a media file for offline access:

1. Open the entity detail screen.
2. In the file list, tap the download icon next to the file.
3. Select the download quality (original, or a transcoded version if available).
4. The download progress appears in the notification bar.

Downloaded files are stored in the app's internal storage and are accessible from the **Profile > Downloads** screen. Downloaded items display a checkmark badge in the library browser.

### Sync Settings

Control sync behavior in **Settings > Sync**:

- **Auto-sync on Wi-Fi** -- Automatically refresh library metadata when connected to Wi-Fi (enabled by default).
- **Sync frequency** -- Choose between manual, every hour, every 6 hours, or daily.
- **Download on Wi-Fi only** -- Prevent large file downloads over mobile data (enabled by default).

---

## Settings and Configuration

Access settings from the Profile tab by tapping the gear icon.

### Server Connection

- **Server URL** -- Change the connected Catalogizer server.
- **Connection timeout** -- Adjust the timeout for server requests (default: 30 seconds).
- **HTTP/3 (QUIC)** -- Toggle HTTP/3 transport. Disable if your network blocks UDP traffic.

### Playback

- **Default quality** -- Preferred playback quality when multiple files are linked to an entity.
- **Resume playback** -- Resume video from last position (enabled by default).
- **Hardware decoding** -- Enable or disable hardware-accelerated video decoding.
- **External player** -- Set a preferred external player app for the "Open with" action.

### Appearance

- **Theme** -- Light, dark, or system default.
- **Grid columns** -- Number of columns in the media grid (auto, 2, 3, or 4).
- **Language** -- App interface language.

### Storage

- **Cache size limit** -- Maximum disk space for cached thumbnails and metadata (default: 500 MB).
- **Clear cache** -- Remove all cached data. The next browse session will re-download thumbnails.
- **Downloaded files** -- View and manage files saved for offline playback.

### Notifications

- **New content alerts** -- Receive a notification when new media is added to the library.
- **Download complete** -- Notification when an offline download finishes.

---

## Troubleshooting

### Cannot Connect to Server

- Verify that the server URL is correct and includes the port number (e.g., `https://192.168.0.100:8080`).
- Ensure your device is on the same network as the Catalogizer server, or that the server is accessible over the internet.
- Check that the Catalogizer API container is running (`podman ps` on the server).
- If using HTTP/3, try disabling it in settings to fall back to HTTP/2. Some networks block UDP traffic required by QUIC.

### Login Fails

- Double-check your username and password. Passwords are case-sensitive.
- If you have forgotten your password, ask your administrator to reset it from the admin panel.
- Ensure the server clock is reasonably synchronized with your device. JWT validation may fail with significant clock skew.

### Media Won't Play

- Verify that the storage root containing the file is accessible from the server (not just from your device).
- Check that the file format is supported by your device. Try opening the file with an external player for comparison.
- For network playback issues, try downloading the file for offline playback and playing locally.
- If you see a black screen with audio, disable hardware decoding in playback settings.

### Sync Not Working

- Confirm that auto-sync is enabled in Settings > Sync.
- Check your network connection. The sync indicator in the Profile tab shows the last successful sync time.
- Force a manual sync by pulling down on the Browse screen.

### App Crashes

- Ensure you are running the latest version of the app. Update via the Play Store or install the latest APK.
- Clear the app cache from Android Settings > Apps > Catalogizer > Storage > Clear Cache.
- If the issue persists, clear all app data (this will log you out and remove downloads) and log in again.
- Report persistent crashes to your administrator with the device model, Android version, and steps to reproduce.
