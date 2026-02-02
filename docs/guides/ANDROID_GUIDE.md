# Catalogizer Android Mobile App Guide

This guide covers the Catalogizer Android mobile app, built with Kotlin and Jetpack Compose following the MVVM architecture.

## Table of Contents

1. [Installation](#installation)
2. [Login](#login)
3. [Home Screen](#home-screen)
4. [Search](#search)
5. [Offline Mode](#offline-mode)
6. [Settings](#settings)
7. [Common Workflows](#common-workflows)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### From Google Play Store

1. Open the Google Play Store on your Android device.
2. Search for "Catalogizer".
3. Tap **Install**.
4. Once installed, tap **Open** or find the app in your app drawer.

### From APK (Enterprise/Sideload)

1. Download the APK file from your organization's distribution channel.
2. On your device, go to **Settings > Security** and enable **Install from unknown sources** (or grant permission to your file manager/browser).
3. Open the downloaded APK file and tap **Install**.
4. Launch Catalogizer from your app drawer.

### System Requirements

- Android 8.0 (API level 26) or higher
- Minimum 2 GB RAM recommended
- Network connectivity (Wi-Fi or cellular) for initial setup and online features
- Storage space varies based on offline caching preferences

---

## Login

When you first launch Catalogizer, you are presented with the Login screen.

### Connecting to Your Server

1. Enter the **Server URL** of your Catalogizer instance (e.g. `http://192.168.1.100:8080` or `https://catalogizer.example.com`).
2. Enter your **Username** and **Password**.
3. Tap **Sign In**.

If authentication is successful, you are redirected to the Home screen. Your session token is stored securely on the device so you remain logged in across app restarts.

### Authentication Errors

- **Invalid credentials** -- verify your username and password are correct
- **Connection failed** -- check that the server URL is correct and the server is reachable from your network
- **Server timeout** -- the server may be under heavy load; try again after a few moments

---

## Home Screen

The Home screen is the main hub of the app, showing your media at a glance.

### Navigation

The top app bar includes:

- **Title**: "Catalogizer"
- **Search icon** -- navigates to the Search screen
- **Settings icon** -- navigates to the Settings screen

### Recently Added Section

A horizontally scrollable row of media cards showing items recently added to your library. Each card displays:

- A placeholder thumbnail with the first letter of the media type
- Media title (up to 2 lines)
- Year of release (if available)
- Rating (if available)

Tap a card to view the media detail (feature planned for a future update).

### Favorites Section

A horizontally scrollable row showing your favorite media items. This section only appears if you have favorited at least one item.

### Empty State

If your library is empty, you will see a message: "No media found. Your media library is empty. Configure storage sources to get started." This indicates you need to configure storage sources on the server via the web app or installer wizard.

### Loading State

While data is loading, a centered progress spinner is displayed. An error card appears at the top if loading fails, showing the error message.

---

## Search

The Search screen allows you to find specific media items.

### Using Search

1. Tap the search icon on the Home screen.
2. Type your search query in the text field.
3. Results appear as you type (or after submitting).
4. Tap a result to view details.

### Search Capabilities

- Search by title
- Search by media type
- Results come from both the server (when online) and cached data (when offline)

### Returning to Home

Tap the back arrow or use the system back gesture to return to the Home screen.

---

## Offline Mode

One of Catalogizer Android's key features is robust offline support, allowing you to browse and interact with your media library without an internet connection.

### How Offline Mode Works

The app uses a local Room database to cache media data. When you are online:

- Media items you browse are automatically cached locally.
- Search results are cached for offline access.
- Your actions (favorites, ratings, watch progress) are stored locally.

When you go offline:

- You can browse all previously cached media items.
- Search returns results from the local cache.
- Actions you take (favoriting, rating) are queued as **sync operations**.
- When connectivity returns, queued operations are automatically synced to the server.

### Offline Settings

The following offline settings are available (configurable via the offline preferences):

| Setting | Default | Description |
|---------|---------|-------------|
| Offline Mode | Off | Enables explicit offline mode with periodic sync |
| Auto Download | Off | Automatically download media metadata for offline use |
| Download Quality | 1080p | Quality level for downloaded content |
| Wi-Fi Only | On | Only sync and download when connected to Wi-Fi |
| Storage Limit | 5 GB | Maximum local storage for cached content |

### Sync Operations

When offline, the following actions are queued for later sync:

- **Favorite toggles** -- adding or removing favorites
- **Rating updates** -- rating media items
- **Watch progress** -- tracking how far you watched a video

When the device regains network connectivity, queued sync operations are automatically processed. You can also trigger a manual sync.

### Cache Management

- Cached items older than 30 days (that are not favorites or recently watched) are automatically cleaned up.
- You can view offline statistics including: cached items count, pending sync operations, failed sync operations, storage usage.
- The storage usage percentage is displayed relative to your configured storage limit.

### Data Export and Import

For backup purposes, you can export your offline data (cached media, sync operations, search queries) to a JSON file and import it later on the same or a different device.

---

## Settings

The Settings screen is accessible via the gear icon on the Home screen.

### About Section

Displays app identification:

- App name: "Catalogizer for Android"
- Description: "Multi-platform media collection manager"

### Sign Out

The Sign Out section includes a prominent button that:

1. Clears your authentication token.
2. Returns you to the Login screen.
3. Stops any background sync operations.

---

## Common Workflows

### Browsing Your Library on Mobile

1. Open the app and view the Home screen.
2. Scroll through "Recently Added" and "Favorites" sections.
3. Tap any media card to see more details.

### Searching for Specific Content

1. Tap the search icon on the Home screen.
2. Enter your search term (e.g. movie title, genre keyword).
3. Browse the results and tap to view details.

### Using the App Offline

1. While online, browse your library to cache content locally.
2. When you lose connectivity, the app seamlessly switches to offline mode.
3. Continue browsing cached content and performing actions.
4. When connectivity returns, pending changes sync automatically.

### Managing Favorites While Offline

1. Browse cached media items.
2. Toggle favorites -- the change is saved locally and queued for sync.
3. Once online, the server is updated automatically.

---

## Troubleshooting

### App Cannot Connect to Server

- Verify the server URL is correct and includes the protocol (http:// or https://).
- Ensure the server is running and accessible from your network.
- If on cellular data, check that the server is publicly accessible or that you are connected via VPN.
- Check that your device clock is synchronized (required for JWT token validation).

### Slow Performance

- Clear the app cache: Settings > Apps > Catalogizer > Storage > Clear Cache.
- If the local database has grown large, trigger a cache cleanup.
- Ensure you have sufficient free storage on your device.

### Sync Failures

- Check the offline stats for failed sync operations count.
- Verify that the server is reachable.
- Force a manual sync when connectivity is stable.
- If issues persist, try logging out and back in to refresh your authentication token.

### App Crashes on Launch

- Ensure your device meets the minimum Android version requirement (8.0+).
- Try clearing app data: Settings > Apps > Catalogizer > Storage > Clear Data.
- Update the app to the latest version.
- If the issue persists, uninstall and reinstall the app.
