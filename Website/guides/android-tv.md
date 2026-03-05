---
title: Android TV Guide
description: Using Catalogizer on Android TV with the lean-back interface and remote control navigation
---

# Android TV Guide

The Catalogizer Android TV app is optimized for large-screen viewing with a lean-back interface designed for remote control navigation. Browse your media catalog from your couch and play content directly on your TV.

---

## Requirements

- Android TV device running Android 8.0 (API 26) or higher
- Network access to your Catalogizer server

---

## Installation

1. Sideload the APK from the [Download](/download) page using a file manager or ADB
2. The app appears in your Android TV launcher under Apps

Alternatively, transfer the APK to your TV via USB drive or network share.

---

## First Setup

1. Launch Catalogizer from the Apps row
2. Enter your server URL using the on-screen keyboard
3. Log in with your credentials
4. The app loads your catalog and displays the home screen

---

## Lean-Back Interface

The TV interface uses the Android Leanback library for a 10-foot viewing experience.

### Home Screen

- **Continue Watching**: Resume media from where you left off
- **Recently Added**: New items detected in the latest scan
- **Movies**: Browse your movie collection
- **TV Shows**: Browse series with season and episode navigation
- **Music**: Browse artists, albums, and songs
- **Collections**: Your collections and playlists

Each row scrolls horizontally. Items display cover art, title, and a quality badge.

### Detail Screen

Select any item to open its detail screen:

- Full poster art and backdrop image
- Title, year, description, and quality metadata
- Play button to start playback
- Add to favorites or collection

---

## Remote Control Navigation

The interface is designed for D-pad navigation with a standard TV remote.

| Button | Action |
|--------|--------|
| D-pad (arrows) | Navigate between items and rows |
| Select / OK | Open item or confirm action |
| Back | Return to previous screen |
| Play/Pause | Toggle media playback |
| Fast-forward / Rewind | Seek forward or backward during playback |
| Home | Return to the Android TV launcher |

### Search

Press the microphone button on your remote to use Google Assistant voice search. Say a title or phrase and Catalogizer returns matching results from your catalog.

You can also navigate to the search icon on the home screen and type a query using the on-screen keyboard.

---

## Media Playback

- Full-screen video playback optimized for TV displays
- Transport overlay with play, pause, seek, and subtitle controls
- Subtitle track selection via the controls overlay
- Resume from last position, synced with other devices
- Playlist auto-advance for sequential playback

### Playback Controls

During playback, press Select/OK to show the transport overlay. Use the D-pad to seek or select controls. The overlay auto-hides after a few seconds of inactivity.

---

## Recommendations

Catalogizer integrates with the Android TV recommendations system. Recently added and partially watched items appear in the Recommendations row on your TV home screen, allowing quick access without opening the app first.

---

## Troubleshooting

**App does not appear in launcher:**
Ensure the APK is installed correctly. Some TV launchers only show apps with a Leanback launcher intent. Try accessing it from Settings > Apps.

**Remote controls unresponsive:**
Verify the app has focus. Press Home and relaunch the app. If using a Bluetooth remote, check the pairing status.

**Playback stuttering:**
Check network bandwidth between your TV and the Catalogizer server. Wired Ethernet provides more reliable streaming than Wi-Fi for high-bitrate content. Verify the storage source is not under heavy load.
