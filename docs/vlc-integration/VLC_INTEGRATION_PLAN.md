# VLC Media Player Integration Plan

## Overview
Integrate VLC as the primary media playback engine for Catalogizer Android and Android TV apps, providing:
- Full video/audio playback support
- Subtitle management (selection, delay adjustment)
- Audio/video track selection
- Playback controls (play, pause, stop, seek)
- Advanced features (speed control, aspect ratio, audio sync)

## Architecture Decision

### Approach: VLC Library Integration (Recommended)
Instead of cloning the entire VLC codebase, we'll use VLC as a library dependency:
- **libvlc-android**: Core VLC engine via Maven/Gradle
- **Custom UI Layer**: Catalogizer-specific player UI
- **Wrapper Module**: Abstraction layer for VLC functionality

### Benefits:
1. **Maintainability**: VLC updates via dependency management
2. **Size**: Only include necessary components
3. **Integration**: Clean separation of concerns
4. **Testing**: Isolated player testing

## Implementation Phases

### Phase 1: VLC Library Setup (Week 1)
- [ ] Add libvlc-android dependencies to build.gradle
- [ ] Configure native library loading
- [ ] Create VLCPlayer wrapper module
- [ ] Basic initialization and lifecycle management

### Phase 2: Core Player UI (Week 2)
- [ ] Create VLCPlayerActivity for Android TV
- [ ] Implement playback controls overlay
- [ ] Basic play/pause/stop/seek functionality
- [ ] Loading and error states

### Phase 3: Advanced Features (Week 3)
- [ ] Subtitle track selection and delay
- [ ] Audio track selection
- [ ] Video track/quality selection
- [ ] Aspect ratio and crop modes
- [ ] Playback speed control

### Phase 4: TV-Optimized UI (Week 4)
- [ ] D-pad navigation for controls
- [ ] Focus management
- [ ] Settings integration
- [ ] Picture-in-picture support

### Phase 5: Testing & Challenges (Week 5)
- [ ] Unit tests for player wrapper
- [ ] UI tests for player controls
- [ ] HelixQA automated test cases
- [ ] Performance testing

## Technical Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Catalogizer App                          │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ MediaDetail  │  │   Search     │  │   Browse     │     │
│  └──────┬───────┘  └──────────────┘  └──────────────┘     │
│         │                                                   │
│         ▼                                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │         VLCPlayerActivity (TV-Optimized)           │    │
│  ├────────────────────────────────────────────────────┤    │
│  │ ┌──────────────┐  ┌──────────────────────────────┐ │    │
│  │ │ VLC Controls │  │       VLC SurfaceView        │ │    │
│  │ │  (Overlay)   │  │                              │ │    │
│  │ └──────────────┘  └──────────────────────────────┘ │    │
│  └────────────────────┬───────────────────────────────┘    │
│                       │                                     │
│         ┌─────────────┴──────────────┐                     │
│         ▼                            ▼                     │
│  ┌──────────────┐           ┌──────────────────┐          │
│  │ VLCPlayer    │◄─────────►│    libVLC        │          │
│  │   Wrapper    │           │  (Native Engine) │          │
│  └──────────────┘           └──────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

## Dependencies

### Gradle Configuration
```kotlin
// VLC Core
dependencies {
    implementation "org.videolan.android:libvlc-all:3.6.0"
    
    // Optional: Media library for indexing
    implementation "org.videolan.android:medialibrary:0.13.2"
}
```

## Key Components

### 1. VLCPlayer Wrapper
```kotlin
class VLCPlayer(context: Context) {
    private var libVLC: LibVLC
    private var mediaPlayer: MediaPlayer
    
    fun play(uri: String)
    fun pause()
    fun stop()
    fun seek(position: Long)
    fun setSubtitleTrack(trackId: Int)
    fun setAudioTrack(trackId: Int)
    fun setPlaybackSpeed(speed: Float)
}
```

### 2. VLCPlayerActivity (TV-Optimized)
- Full-screen player with overlay controls
- D-pad navigation support
- Focus-based UI interactions
- Lifecycle management

### 3. Player Controls Overlay
- Play/Pause button
- Seek bar (progress)
- Current/Total time display
- Audio track selector
- Subtitle track selector
- Settings menu (aspect ratio, speed, etc.)

## Testing Strategy

### Unit Tests
- VLCPlayer wrapper functionality
- State management
- Error handling

### UI Tests  
- Playback controls
- Navigation flow
- Settings changes

### HelixQA Integration
- Automated playback scenarios
- Error condition testing
- Performance monitoring

## Challenges Integration

Create challenge types for:
1. **VideoPlaybackChallenge**: Test video file playback
2. **SubtitleChallenge**: Verify subtitle rendering
3. **AudioTrackChallenge**: Test audio switching
4. **PlayerUIChallenge**: Test control interactions
5. **ErrorRecoveryChallenge**: Test error handling

## Documentation

- API Documentation for VLC wrapper
- UI/UX guidelines for TV controls
- Troubleshooting guide
- Integration examples

## Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| 1 | 1 week | Library setup, basic player |
| 2 | 1 week | Core UI, basic controls |
| 3 | 1 week | Advanced features |
| 4 | 1 week | TV optimization |
| 5 | 1 week | Testing, documentation |

**Total: 5 weeks**
