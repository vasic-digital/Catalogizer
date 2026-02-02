# Catalogizer Android TV Guide

This guide covers the Catalogizer Android TV app, designed for large-screen navigation using a remote control or gamepad. The app is built with Kotlin, Jetpack Compose for TV (`androidx.tv`), and ExoPlayer for media playback.

## Table of Contents

1. [Installation](#installation)
2. [Remote Navigation Basics](#remote-navigation-basics)
3. [Login](#login)
4. [Home Screen](#home-screen)
5. [Media Detail](#media-detail)
6. [Media Player](#media-player)
7. [Search](#search)
8. [Settings](#settings)
9. [Common Workflows](#common-workflows)
10. [Troubleshooting](#troubleshooting)

---

## Installation

### From Google Play Store (Android TV)

1. On your Android TV, open the **Google Play Store**.
2. Search for "Catalogizer TV".
3. Select the app and click **Install**.
4. Once installed, find it in your apps list.

### Sideloading via APK

1. Enable **Unknown Sources** on your Android TV: Settings > Security & restrictions > Unknown sources.
2. Transfer the APK to your TV via USB drive, cloud storage, or a sideload utility.
3. Use a file manager app on the TV to locate and install the APK.

### System Requirements

- Android TV OS 8.0 or higher
- Remote control, gamepad, or compatible keyboard
- Network connectivity (Wi-Fi or Ethernet) to connect to your Catalogizer server

---

## Remote Navigation Basics

The Android TV app is built with focus-based navigation optimized for the TV remote:

| Remote Button | Action |
|--------------|--------|
| D-Pad (Up/Down/Left/Right) | Navigate between items and sections |
| Select / OK / Center | Confirm selection, open item, or press button |
| Back | Go to previous screen or close overlay |
| Home | Return to the Android TV home screen |
| Play/Pause | Toggle playback in the media player |
| Fast Forward / Rewind | Seek forward or backward during playback |

Focus moves between items in a predictable pattern:

- **Horizontal rows**: Left/Right to move between items in a row (e.g. media cards)
- **Vertical sections**: Up/Down to move between rows or sections
- The currently focused item is visually highlighted

---

## Login

When launching the app for the first time (or after logging out), you see the Login screen.

### How to Log In

1. Use the D-Pad to navigate to the **Server URL** field.
2. Press Select to open the on-screen keyboard and type your server address.
3. Navigate to the **Username** field and enter your username.
4. Navigate to the **Password** field and enter your password.
5. Navigate to the **Sign In** button and press Select.

After successful authentication, you are taken to the Home screen.

### Tips for TV Login

- If you have a USB or Bluetooth keyboard connected to your TV, typing credentials is much faster.
- The app stores your session token so you only need to log in once until you explicitly sign out or the token expires.

---

## Home Screen

The Home screen is the main browsing experience, organized into horizontal content rows.

### Top Bar

At the top of the screen you will see:

- **Title**: "Catalogizer"
- **Search** button -- navigates to the Search screen
- **Settings** button -- navigates to the Settings screen

### Content Sections

The Home screen displays categorized content rows, each scrollable horizontally:

1. **Continue Watching** -- media you started but have not finished; tapping an item resumes playback directly
2. **Recently Added** -- the latest media added to your library
3. **Movies** -- all movies in your catalog
4. **TV Shows** -- all TV series in your catalog
5. **Music** -- music items
6. **Documents** -- document-type media items

Each section only appears if it contains at least one item.

### Media Cards

Each card in a content row shows:

- A thumbnail or poster image
- Media title
- Year (if available)
- Rating (if available)
- Media type badge

When a card receives focus, it is visually highlighted. Press Select to navigate to the Media Detail screen. For items in the "Continue Watching" row, pressing Select takes you directly to the Media Player.

### Loading and Error States

- A centered spinner appears while content is loading.
- If loading fails, an error message is shown with a **Retry** button.

---

## Media Detail

The Media Detail screen provides comprehensive information about a selected media item.

### Information Displayed

- Full title
- Media type, year, and quality
- Description / synopsis
- Rating and metadata from external providers
- File path and size information

### Actions

- **Play** -- opens the media player
- **Back** -- returns to the previous screen

Navigate between action buttons using the D-Pad and press Select to activate.

---

## Media Player

The Media Player uses ExoPlayer for high-quality video and audio playback, optimized for TV.

### Player Interface

The player fills the entire screen with the following overlay elements:

**Top Bar** (semi-transparent overlay):
- Media title displayed on the left
- **Back** button on the right

**Bottom Bar** (semi-transparent overlay):
- Current position and total duration (e.g. "01:23:45 / 02:15:00")
- Playback state indicator ("Playing" or "Paused")

### Playback Controls

| Remote Button | Action |
|--------------|--------|
| Play/Pause | Toggle between playing and paused states |
| Select / OK | Toggle play/pause |
| Back | Exit the player and return to the previous screen |
| Fast Forward | Seek forward |
| Rewind | Seek backward |
| D-Pad Left/Right | Scrub through the timeline |

### Auto-Play

When playback begins, the player automatically starts playing (playWhenReady is enabled).

### Error Handling

If the media URL is not available or cannot be loaded:

- A message is displayed: "Media URL not available" along with the Media ID.
- A **Back** button allows you to return to the previous screen.
- If loading is in progress, a spinner and "Loading media..." message appear.

---

## Search

The Search screen allows you to find media items across your entire catalog.

### Using Search on TV

1. From the Home screen, navigate to the **Search** button in the top bar and press Select.
2. Use the on-screen keyboard (or a connected physical keyboard) to type your query.
3. Results appear below the search field as you type.
4. Use the D-Pad to navigate through results.
5. Press Select on a result to go to its Media Detail screen.

### Returning Home

Press the **Back** button on your remote to return to the Home screen.

---

## Settings

The Settings screen lets you configure app preferences.

### Available Settings

| Setting | Options | Description |
|---------|---------|-------------|
| Notifications | On / Off | Enable or disable push notifications |
| Auto-Play | On / Off | Automatically start playing the next item |
| Streaming Quality | Auto / 720p / 1080p / 4K | Video quality for streaming |
| Subtitles | On / Off | Enable or disable subtitle display |
| Subtitle Language | English, Spanish, etc. | Preferred subtitle language |

### Navigating Settings

- Use D-Pad Up/Down to move between setting items.
- Use D-Pad Left/Right or Select to toggle switches or open dropdown selectors.

### Sign Out

At the bottom of the Settings screen, the **Sign Out** option logs you out and returns to the Login screen.

---

## Common Workflows

### Watching a Movie

1. From the Home screen, browse the **Movies** row.
2. Focus on the movie you want and press Select.
3. On the Media Detail screen, focus on **Play** and press Select.
4. The movie begins playing. Use your remote to control playback.
5. Press **Back** to exit the player.

### Resuming Playback

1. Look at the **Continue Watching** row on the Home screen.
2. Press Select on the item you want to resume.
3. Playback starts from where you left off.

### Searching for Content

1. Navigate to the Search button on the Home screen.
2. Type your search term.
3. Browse results and select one to view details or play.

### Changing Playback Quality

1. Go to Settings from the Home screen.
2. Find the **Streaming Quality** option.
3. Select your preferred quality (Auto, 720p, 1080p, 4K).
4. The change takes effect for the next playback session.

---

## Troubleshooting

### Black Screen During Playback

- Check that the media file format is supported by ExoPlayer.
- Verify the server is streaming the file correctly by testing in the web app.
- Try a lower streaming quality setting.
- If the media URL is not available, the player will display "Media URL not available" -- ensure the storage source is connected.

### Remote Controls Not Responding

- Ensure your remote is paired and has battery.
- Restart the app from the Android TV settings.
- Some generic remotes may not map all buttons correctly; try a different remote.

### Cannot Connect to Server

- Verify the server URL is correct.
- Check that your TV is connected to the same network as the server.
- If using a local IP address, ensure your TV has Wi-Fi or Ethernet connectivity.
- Try pinging the server from another device on the same network.

### Content Not Loading

- Check the Home screen for error messages.
- Press the **Retry** button if visible.
- Ensure the server is running and storage sources are connected.
- Try signing out and back in to refresh your authentication token.

### Subtitle Not Displaying

- Verify subtitles are enabled in Settings.
- Check that the subtitle language matches an available subtitle for the media.
- Manage subtitles via the web app's Subtitle Manager if needed.
