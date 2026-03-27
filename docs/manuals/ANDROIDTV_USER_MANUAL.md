# Catalogizer -- Android TV User Manual

## Table of Contents

1. [Installation on Android TV](#installation-on-android-tv)
2. [Remote Control Navigation](#remote-control-navigation)
3. [Browsing Media on the Big Screen](#browsing-media-on-the-big-screen)
4. [Media Playback with TV Remote](#media-playback-with-tv-remote)
5. [Voice Search](#voice-search)
6. [Settings](#settings)
7. [Leanback UI Specifics](#leanback-ui-specifics)
8. [Troubleshooting](#troubleshooting)

---

## Installation on Android TV

### From Google Play Store

1. On your Android TV home screen, open the **Google Play Store** app.
2. Use the search function to find **Catalogizer** by Vasic Digital.
3. Select the app from the results and press **Install** on your remote.
4. Once installed, the app appears in the "Apps" row on your home screen.

### Sideloading via APK

Some Android TV devices (such as Xiaomi Mi Box or certain projectors) may require sideloading:

1. On the TV, go to **Settings > Security** and enable **Unknown sources** for your file manager or sideload app.
2. Transfer the APK to a USB drive and plug it into the TV, or use a sideload utility such as **Send Files to TV** to transfer the file over your local network.
3. Open the APK using a file manager on the TV and confirm installation.
4. The app appears in the Apps row after installation.

### System Requirements

| Requirement | Minimum |
|-------------|---------|
| Android TV version | 8.0 (Oreo) / API 26 |
| RAM | 2 GB |
| Storage | 100 MB free |
| Network | Wi-Fi or Ethernet to Catalogizer server |
| Remote | Standard Android TV remote with D-pad |

---

## Remote Control Navigation

Catalogizer for Android TV is built around the Leanback UI framework, designed entirely for D-pad and remote control interaction. No touch screen is required.

### D-pad Focus Model

All navigation uses the directional pad (D-pad) on your remote:

| Button | Action |
|--------|--------|
| **Up / Down / Left / Right** | Move focus between items |
| **Select (OK / Center)** | Open the focused item or confirm an action |
| **Back** | Return to the previous screen or close a dialog |
| **Home** | Return to the Android TV home screen |

Focus is indicated by a highlighted border around the currently selected item. As you move between rows and columns, the focused item scales up slightly and the row scrolls to keep it centered.

### Navigation Structure

The app follows a left-to-right, top-to-bottom navigation flow:

1. **Header row** -- Contains the app logo, search icon, and settings icon. Press Up from any content row to reach the header.
2. **Content rows** -- Horizontal carousels organized by category (Continue Watching, Recently Added, Movies, TV Shows, Music, etc.). Press Left/Right to scroll within a row, Up/Down to move between rows.
3. **Detail screens** -- Press Select on any card to open its detail screen. The detail screen uses a vertical layout with the poster, metadata, and action buttons.

### Long-Press Actions

Long-pressing the Select button on a media card opens a context menu with options such as **Add to Favorites**, **Add to Playlist**, and **Mark as Watched**.

---

## Browsing Media on the Big Screen

### Home Screen

The home screen presents your media library in horizontal carousels optimized for the 10-foot viewing experience. Each row represents a category:

- **Continue Watching** -- Items you started but did not finish, with a progress indicator overlay on the poster.
- **Recently Added** -- The newest items added to your library, sorted by scan date.
- **Movies** -- All movie entities.
- **TV Shows** -- All TV show entities. Selecting a show opens its season/episode hierarchy.
- **Music** -- Artists and albums.
- **Games, Books, Software** -- Other media types, each in their own row.

Rows are populated dynamically based on your library contents. Empty categories are hidden automatically.

### Entity Cards

Each card in a row displays:

- **Poster image** scaled to a consistent aspect ratio (2:3 for portrait, 16:9 for landscape banners).
- **Title** below the poster.
- **Year** and **rating badge** (if metadata is available).
- **Progress bar** overlay for partially watched items.

Cards are large enough to be readable from across the room. Text uses a minimum size of 18sp for legibility on TVs.

### Entity Detail Screen

Pressing Select on a card opens the detail screen, which displays:

- Full-resolution backdrop image at the top.
- Title, year, genre tags, and rating.
- Plot summary or description text.
- **Action buttons**: Play, Add to Favorites, Add to Playlist.
- **File list**: All linked files with format, resolution, and size information.
- **Related items**: For TV shows, the season and episode list. For music artists, the album list.

Navigate between sections using Up/Down, and between action buttons or file entries using Left/Right.

---

## Media Playback with TV Remote

### Starting Playback

1. Open an entity detail screen.
2. Press Select on the **Play** button. If multiple files are linked, the highest-quality file is selected by default.
3. To choose a specific file, navigate to the file list and press Select on the desired entry.

### Playback Controls

During playback, the transport controls overlay appears at the bottom of the screen. It auto-hides after 5 seconds of inactivity. Press any D-pad button to show it again.

| Button | Action |
|--------|--------|
| **Select (OK)** | Toggle play / pause |
| **Left** | Rewind 10 seconds |
| **Right** | Fast-forward 10 seconds |
| **Up** | Show transport controls if hidden |
| **Down** | Open subtitle and audio track selector |
| **Back** | Stop playback and return to detail screen |

### Seeking

Use the Left/Right buttons for 10-second jumps. Hold Left or Right to enter fast seek mode, which scrubs through the timeline at increasing speed. Release to resume playback at the current position.

### Subtitle Selection

Press Down during playback to open the track selector panel. Navigate to the subtitle track row, choose a subtitle language or "Off" to disable subtitles, and press Select. Embedded and external subtitles are listed together.

### Audio Track Selection

In the same track selector panel, navigate to the audio track row to switch between available audio streams (e.g., different languages or commentary tracks).

### Resume Playback

If you stop playback before a video finishes, the app remembers the position. The next time you open the same entity, the Play button shows **Resume from XX:XX** instead of **Play**. The Continue Watching row on the home screen also provides quick access to unfinished items.

---

## Voice Search

If your Android TV remote has a built-in microphone button, you can use voice search to find media in your library.

1. Press the **microphone button** on your remote.
2. Speak the name of the movie, show, artist, or any search term.
3. Catalogizer registers as a searchable app with the Android TV system. Results from your library appear in the global search results alongside results from other apps.
4. Select a Catalogizer result to open the entity detail screen directly.

Voice search requires that your TV has Google Assistant or the built-in Android TV search functionality enabled. On devices without a microphone button, use the on-screen search in the app header instead.

---

## Settings

Access settings by navigating to the gear icon in the header row and pressing Select.

### Server Connection

- **Server URL** -- The address of your Catalogizer API server.
- **HTTP/3 (QUIC)** -- Enable or disable QUIC transport.
- **Connection timeout** -- Adjust the request timeout (default: 30 seconds).

### Playback

- **Preferred quality** -- When multiple files exist, automatically select the file matching this quality (4K, 1080p, 720p, or highest available).
- **Resume playback** -- Enable or disable automatic position resumption.
- **Hardware decoding** -- Toggle hardware-accelerated video decoding. Disable if you experience playback artifacts.

### Display

- **Overscan adjustment** -- Shift the UI inward if content is cut off at the edges of your TV. Use the arrow keys to adjust each edge independently.
- **UI scaling** -- Adjust the overall UI scale (90%, 100%, 110%, 120%) for different TV sizes and viewing distances.

### Sync

- **Auto-sync** -- Automatically refresh library metadata on app launch (enabled by default).
- **Sync frequency** -- Manual, every hour, every 6 hours, or daily.

### Account

- **Log out** -- Sign out of the current account. You will need to re-enter credentials on next launch.
- **Switch server** -- Connect to a different Catalogizer server.

---

## Leanback UI Specifics

Catalogizer for Android TV uses the Android Leanback library and Jetpack Compose for TV to provide a native big-screen experience. This section describes behavior specific to the TV interface.

### Focus Management

The Leanback framework manages focus automatically. When a screen loads, focus is placed on the first actionable item. When navigating back from a detail screen, focus returns to the card that was previously selected in the browse row, preserving your position.

### Row Scrolling

Content rows scroll horizontally as you move focus with Left/Right. The focused card is centered on screen when possible. Rows load additional items on demand as you scroll toward the end, so even large libraries remain responsive.

### Background Updates

As focus moves between cards, the background image behind the rows updates to show the backdrop of the currently focused item (for movies and TV shows with backdrop metadata). This provides visual context without requiring you to open the detail screen.

### Recommendations Row

Catalogizer can publish recommendations to the Android TV home screen's "Recommendations" channel (Android 8+) or the "Watch Next" row (Android 10+). Recommendations include recently added items and items you have partially watched. This feature can be enabled or disabled in Settings.

### Screen Saver Integration

If the app is idle for the duration configured in your TV's screen saver settings, the system screen saver activates as normal. Playback is not affected -- the screen saver only triggers when on browse or detail screens, never during active media playback.

### Gamepad and Keyboard Support

In addition to the standard TV remote, Catalogizer supports Bluetooth gamepads and USB keyboards. The D-pad on a gamepad maps to the same navigation controls. A keyboard can be used for text input on the search screen and for playback shortcuts (Space for play/pause, arrow keys for seeking).

---

## Troubleshooting

### App Not Visible on Home Screen

After installation, the app should appear in the Apps row. If it does not, open **Settings > Apps > See all apps** and locate Catalogizer. Press **Open** from there. Some TV launchers require a restart to show newly installed apps.

### Cannot Connect to Server

- Verify the server URL includes the correct port (e.g., `https://192.168.0.100:8080`).
- Ensure the TV is connected to the same network as the Catalogizer server. Check the TV's network settings under **Settings > Network & Internet**.
- If using Ethernet, verify the cable is connected and the link indicator is active.
- Try disabling HTTP/3 in the app settings to fall back to HTTP/2 in case UDP traffic is blocked.

### Playback Stuttering or Buffering

- Check your network bandwidth. Streaming 4K video requires a stable connection of at least 25 Mbps. Ethernet is recommended over Wi-Fi for high-bitrate content.
- Try selecting a lower-quality file if multiple are available.
- Disable hardware decoding in settings if the issue is specific to certain codecs.
- On Mi Box and similar devices with limited hardware, avoid 4K HEVC content or ensure hardware decoding is enabled.

### Remote Buttons Not Responding

- Replace the batteries in your remote.
- Re-pair the remote via **Settings > Remotes & Accessories**.
- Restart the app by pressing Home, then reopening Catalogizer from the Apps row.

### Focus Gets Lost

If D-pad navigation stops working or focus seems stuck, press Back to return to a known screen, then navigate forward again. If the issue persists, restart the app.

### Audio but No Video

- This is typically a codec compatibility issue. Try disabling hardware decoding in playback settings.
- Some older Android TV devices do not support HEVC (H.265) or VP9. Try a file encoded in H.264 if available.

### Login Screen Keyboard Difficult to Use

The on-screen keyboard on Android TV can be tedious for entering long passwords. Consider using the **Android TV Remote Control** app on your phone to type credentials using your phone's keyboard, or pair a Bluetooth keyboard with your TV.
