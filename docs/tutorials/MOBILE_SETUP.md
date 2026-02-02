# Mobile Setup (Android)

This tutorial covers installing the Catalogizer Android app, connecting it to your server, and configuring offline mode for use without a network connection.

## Prerequisites

- Catalogizer API server running and accessible from your network (see [Quick Start](QUICK_START.md))
- Android device running Android 8.0 (API 26) or later
- The server must be reachable from the device's network (same LAN or VPN)

## Step 1: Install the Android App

### Option A: Build from Source

If you have Android Studio installed:

```bash
cd catalogizer-android
./gradlew assembleDebug
```

Transfer the APK to your device:

```bash
# The APK is located at:
# catalogizer-android/app/build/outputs/apk/debug/app-debug.apk
adb install app/build/outputs/apk/debug/app-debug.apk
```

### Option B: Install via Android Studio

1. Open Android Studio
2. Open the `catalogizer-android` project directory
3. Connect your device via USB (enable USB debugging in Developer Options)
4. Click **Run** > **Run 'app'**

**Expected result:** The Catalogizer app installs and launches on your device, showing the login screen.

## Step 2: Connect to Your Server

1. Open the Catalogizer app on your device
2. On the login screen, tap **Server Settings** or the gear icon
3. Enter your server details:
   - **Server URL:** `http://<your-server-ip>:8080` (use your LAN IP, not `localhost`)
   - Example: `http://192.168.1.50:8080`
4. Tap **Save** or **Test Connection**
5. Enter your credentials (the same username/password you created on the server)
6. Tap **Login**

**Expected result:** The app connects to your server and displays the main media catalog screen.

To find your server's LAN IP:

```bash
# Linux
hostname -I

# macOS
ipconfig getifaddr en0
```

## Step 3: Browse Your Media Catalog

After logging in, you can:

- **Browse by category:** Movies, TV Shows, Music, Games, Software, and more
- **Search:** Use the search bar to find specific media by title
- **Filter:** Apply filters for quality, year, genre, and rating
- **View details:** Tap any media item for full metadata, quality info, and external provider data

**Expected result:** Your media catalog loads from the server. Items appear with titles, types, and quality information.

## Step 4: Configure Offline Mode

The Android app includes offline support powered by a local Room database and a sync manager. Configure offline mode to access your catalog without a network connection.

1. Open the app's **Settings** (gear icon or menu)
2. Navigate to **Offline Settings**
3. Configure the following options:
   - **Offline Mode:** Toggle on to enable local caching
   - **Auto Download:** Automatically cache media metadata when connected
   - **Download Quality:** Select preferred quality for cached thumbnails (e.g., 1080p)
   - **Wi-Fi Only:** Restrict sync operations to Wi-Fi connections
   - **Storage Limit:** Set maximum storage for cached data (in MB)

**Expected result:** When enabled, the app caches media metadata locally. Browsing and searching work without a network connection using the cached data.

## Step 5: Sync Data for Offline Use

With offline mode enabled:

1. Ensure you are connected to your server (on the same network)
2. Navigate to **Settings** > **Offline Settings** > **Sync Now** (or wait for auto-sync if Auto Download is enabled)
3. The app downloads metadata for your media catalog to the local database

**Expected result:** A progress indicator shows the sync operation. Once complete, you can disconnect from the network and still browse your catalog.

The sync manager handles:
- Incremental updates (only syncs changes since the last sync)
- Conflict resolution between local and server data
- Queued operations that execute when connectivity returns

## Step 6: Verify Offline Access

1. Disconnect from your network (disable Wi-Fi and mobile data)
2. Open the Catalogizer app
3. Browse and search your catalog

**Expected result:** The app displays cached media items. Some features that require server connectivity (like fetching new external metadata) will show an offline indicator, but browsing and searching cached data works normally.

## Android TV Setup

For Android TV devices, use the `catalogizer-androidtv` app instead:

```bash
cd catalogizer-androidtv
./gradlew assembleDebug
adb install app/build/outputs/apk/debug/app-debug.apk
```

The Android TV app provides:
- **Leanback UI** optimized for TV screens
- **D-pad navigation** for remote control use
- **Voice search** via Google Assistant integration
- **Recommendations** on the Android TV home screen

Connection and configuration steps are the same as the mobile app.

## Troubleshooting

### "Connection failed" when entering server URL

- Ensure the server URL uses your LAN IP address, not `localhost` or `127.0.0.1`
- Verify the device is on the same network as the server
- Check that port 8080 is not blocked by a firewall
- If using HTTPS, ensure the certificate is valid or add a security exception
- Test connectivity by opening `http://<server-ip>:8080/health` in the device's browser

### Login succeeds but catalog is empty

- The server may still be scanning storage sources
- Check the server-side scan status via the API or web interface
- Pull to refresh in the app to reload data

### Offline mode shows outdated data

- Trigger a manual sync from Settings when connected to the network
- Verify Auto Download is enabled for automatic updates
- Check the Storage Limit setting -- if the limit is reached, older cached data may be purged

### App crashes on startup

- Ensure Android 8.0+ is installed on the device
- Clear the app's data: Settings > Apps > Catalogizer > Clear Data
- If building from source, ensure you are using the latest code and dependencies

### Sync operations are slow

- Enable the **Wi-Fi Only** option to avoid slow mobile data sync
- Reduce the amount of cached data by setting a lower Storage Limit
- Large catalogs may take several minutes for the initial sync
