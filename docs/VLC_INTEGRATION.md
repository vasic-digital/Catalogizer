# VLC Media Player Integration

## Overview

Catalogizer now includes full VLC Media Player integration for both Android TV and Desktop platforms, providing robust video/audio playback with advanced features like track selection, speed control, and cross-platform watch progress synchronization.

## Features

### Core Playback
- **Play/Pause/Stop**: Full playback control
- **Seek**: Jump to any position in the video
- **Volume**: Adjust volume with mute toggle
- **Speed Control**: Playback speed from 0.25x to 3x
- **Fullscreen**: Toggle fullscreen mode (Desktop)

### Track Selection
- **Audio Tracks**: Switch between available audio languages
- **Subtitle Tracks**: Enable/disable and select subtitle tracks
- **Video Tracks**: View available video tracks

### Watch Progress
- **Auto-save**: Progress saved every 5 seconds during playback
- **Resume**: Automatically resume from last position (if >5% and <95%)
- **Cross-platform**: Progress syncs across Android TV and Desktop
- **Smart Filtering**: Avoids saving at 0% or 100% to prevent spoilers

## Android TV

### Architecture
```
MediaCard (click)
    ↓
TVNavigation (Player route)
    ↓
VLCPlayerActivity
    ↓
VLCPlayer (libvlc wrapper)
    ↓
VLCVideoLayout (native rendering)
```

### Key Components

#### VLCPlayer.kt
Core wrapper around libvlc-android:
- Manages libVLC lifecycle
- Provides StateFlow-based state management
- Handles track enumeration and selection
- Supports aspect ratio and speed control

```kotlin
// Usage
val vlcPlayer = VLCPlayer(context)
vlcPlayer.play(streamUrl)
vlcPlayer.seek(0.5f) // 50%
vlcPlayer.setSpeed(1.5f)
vlcPlayer.setAudioTrack(trackId)
```

#### VLCPlayerActivity.kt
Full-screen player activity:
- TV-optimized UI with D-pad navigation
- Auto-hiding controls (5 second timeout)
- Remote control key handling
- Progress tracking and resume

```kotlin
// Launch from navigation
val intent = Intent(context, VLCPlayerActivity::class.java).apply {
    putExtra(VLCPlayerActivity.EXTRA_MEDIA_ID, mediaId)
}
context.startActivity(intent)
```

### UI Features
- **Progress Bar**: Netflix-style progress indicator on media cards
- **Controls Overlay**: Play/pause, seek, volume, speed, tracks
- **Settings Menu**: Aspect ratio, playback speed
- **Track Menus**: Audio and subtitle selection

### Remote Control Keys
| Key | Action |
|-----|--------|
| D-pad Center | Play/Pause |
| D-pad Left/Right | Seek ±30s |
| Back | Exit player (saves progress) |
| Play/Pause | Toggle playback |
| Fast Forward/Rewind | Seek ±30s |

## Desktop

### Architecture
```
MediaDetailPage (Play button)
    ↓
VLCPlayer (React component)
    ↓
useVLCPlayer (hook)
    ↓
Tauri Commands
    ↓
VLCPlayer (Rust)
    ↓
libvlc (native)
```

### Key Components

#### useVLCPlayer.ts
React hook for VLC control:
```typescript
const {
  status,
  play,
  pause,
  seek,
  setVolume,
  setAudioTrack,
  saveProgress
} = useVLCPlayer();

// Play media
await play(streamUrl, mediaId, title);

// Save progress manually
await saveProgress(mediaId);
```

#### VLCPlayer.tsx
React component with Netflix-style UI:
- Full-screen overlay
- Progress bar with time display
- Volume slider
- Speed selector
- Track selection menus

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| Space / K | Play/Pause |
| ← / → | Seek ±10s |
| ↑ / ↓ | Volume ±5% |
| F | Toggle fullscreen |
| M | Toggle mute |
| Esc | Close menus / exit fullscreen |

## API Integration

### Stream Endpoint
```
GET /api/v1/entities/:id/stream
Response: {
  "entity_id": 123,
  "file_id": 456,
  "stream_url": "/api/v1/stream/456"
}
```

### Watch Progress Endpoints
```
PUT /api/v1/media/:id/progress
Body: {
  "media_id": 123,
  "position": 360000,  // milliseconds
  "duration": 7200000, // milliseconds
  "timestamp": 1712581200000
}
```

## Build Requirements

### Android TV
```groovy
// build.gradle.kts
dependencies {
    implementation("org.videolan.android:libvlc-all:3.6.0")
}
```

### Desktop
```toml
# Cargo.toml
[dependencies.libvlc-sys]
version = "0.2"
optional = true

[features]
vlc-player = ["libvlc-sys"]
```

### Builder Container
Dockerfile includes VLC libraries:
```dockerfile
RUN apt-get install -y libvlc-dev libvlccore-dev vlc
```

## Configuration

### Android TV
No additional configuration needed. VLC is bundled with the APK.

### Desktop
VLC must be installed on the system:
- **Ubuntu/Debian**: `sudo apt install vlc libvlc-dev`
- **macOS**: `brew install vlc`
- **Windows**: Include VLC SDK in build path

## Troubleshooting

### Android TV

#### Playback fails to start
- Check stream URL accessibility
- Verify API authentication token
- Check network connectivity to server

#### No audio/subtitle tracks
- Verify media file has multiple tracks
- Check VLC options in VLCPlayer.kt

#### Progress not saving
- Check API connectivity
- Verify mediaId is valid
- Check logcat for errors

### Desktop

#### VLC not found
- Install VLC on system
- Check libvlc-sys feature is enabled

#### Window not showing
- Check Tauri configuration
- Verify frontend build succeeded

#### Progress not syncing
- Check API service configuration
- Verify mediaId is passed correctly

## Performance

### Android TV
- VLC uses hardware acceleration by default
- 1080p playback: ~100MB RAM
- 4K playback: ~200MB RAM

### Desktop
- VLC runs in separate native window
- Memory usage depends on video resolution
- Supports hardware acceleration via VLC settings

## Future Enhancements

- [ ] Picture-in-picture mode
- [ ] Audio visualization
- [ ] Playlist support
- [ ] Casting/Chromecast
- [ ] Subtitle download
- [ ] Audio boost/normalization

## References

- [libvlc Android Documentation](https://wiki.videolan.org/Android_libVLC/)
- [VLC Stream Output](https://wiki.videolan.org/Documentation:Streaming_HowTo/)
- [Tauri Command Pattern](https://tauri.app/v1/guides/features/command/)
