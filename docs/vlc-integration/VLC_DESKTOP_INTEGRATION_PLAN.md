# VLC Media Player Integration Plan - Desktop (Tauri)

## Overview
Integrate VLC as the primary media playback engine for Catalogizer Desktop application using Tauri's Rust backend with embedded VLC.

## Architecture Decision

### Approach: libvlc-sys with Tauri Commands
For Tauri-based Desktop apps, we'll use VLC through Rust FFI bindings:
- **libvlc-sys**: Rust FFI bindings to libVLC
- **Tauri Commands**: Bridge between frontend and VLC
- **WebView Integration**: Video rendering in the UI

### Alternative: WebChimera.js / Web VLC
- Use VLC's Web plugin (if available)
- Or use a custom protocol handler

## Technical Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     Catalogizer Desktop (Tauri)                             │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    WebView Frontend (React)                         │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │   │
│  │  │ Player UI    │  │   Controls   │  │   Settings/Tracks        │  │   │
│  │  └──────┬───────┘  └──────┬───────┘  └────────────┬─────────────┘  │   │
│  │         │                  │                       │                │   │
│  │         └──────────────────┴───────────────────────┘                │   │
│  │                          │                                          │   │
│  │                    ┌─────┴──────┐                                   │   │
│  │                    │ Video Tag  │ (canvas/webgl for VLC output)      │   │
│  │                    └─────┬──────┘                                   │   │
│  └──────────────────────────┼──────────────────────────────────────────┘   │
│                             │                                               │
│         Tauri IPC (Commands)│                                               │
│                             ▼                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    Rust Backend (Tauri)                             │   │
│  │  ┌───────────────────────────────────────────────────────────────┐  │   │
│  │  │                   VLCHandler (Rust)                           │  │   │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │  │   │
│  │  │  │ libvlc-sys   │  │   Event      │  │   Stream         │   │  │   │
│  │  │  │   Wrapper    │  │   Handler    │  │   Management     │   │  │   │
│  │  │  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │  │   │
│  │  │         │                 │                    │             │  │   │
│  │  │         └─────────────────┴────────────────────┘             │  │   │
│  │  │                          │                                   │  │   │
│  │  └──────────────────────────┼───────────────────────────────────┘  │   │
│  │                             │                                       │   │
│  │         FFI Call            │                                       │   │
│  │                             ▼                                       │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │                    libVLC (System)                          │   │   │
│  │  │         (Installed via system package manager)              │   │   │
│  │  └─────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Rust Backend Setup (Week 1)
- [ ] Add libvlc-sys dependency to Cargo.toml
- [ ] Create VLCHandler struct in Rust
- [ ] Implement basic VLC initialization
- [ ] Tauri commands: play, pause, stop, seek
- [ ] Event handling (position, state changes)

### Phase 2: Frontend Player UI (Week 2)
- [ ] Create VideoPlayer component
- [ ] Implement player controls (React)
- [ ] Tauri command integration
- [ ] Loading and error states

### Phase 3: Advanced Features (Week 3)
- [ ] Subtitle track selection
- [ ] Audio track selection
- [ ] Playback speed control
- [ ] Aspect ratio settings
- [ ] Volume control

### Phase 4: Desktop-Specific Features (Week 4)
- [ ] Keyboard shortcuts
- [ ] Window fullscreen mode
- [ ] Drag and drop support
- [ ] File association
- [ ] Picture-in-picture

### Phase 5: Testing & Documentation (Week 5)
- [ ] Rust unit tests for VLCHandler
- [ ] UI component tests
- [ ] HelixQA test cases
- [ ] User documentation

## Dependencies

### Rust (Cargo.toml)
```toml
[dependencies]
# VLC bindings
libvlc-sys = "0.2"

# Tauri
tauri = { version = "1.5", features = ["shell-open"] }

# Async runtime
tokio = { version = "1.0", features = ["full"] }

# Serialization
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

# Error handling
anyhow = "1.0"
thiserror = "1.0"
```

### System Requirements
Users need VLC installed:
- **Linux**: `sudo apt install libvlc-dev` (Debian/Ubuntu) or `sudo dnf install vlc-devel` (Fedora)
- **macOS**: `brew install vlc`
- **Windows**: VLC installed with SDK

## Key Components

### 1. VLCHandler (Rust)
```rust
// src-tauri/src/vlc.rs
use libvlc_sys::*;

pub struct VLCHandler {
    instance: *mut libvlc_instance_t,
    media_player: *mut libvlc_media_player_t,
}

impl VLCHandler {
    pub fn new() -> Result<Self, VLCError> {
        // Initialize libVLC
    }
    
    pub fn play(&self, uri: &str) -> Result<(), VLCError> {
        // Load and play media
    }
    
    pub fn pause(&self) {
        // Pause playback
    }
    
    pub fn stop(&self) {
        // Stop playback
    }
    
    pub fn seek(&self, position: f64) {
        // Seek to position (0.0 - 1.0)
    }
    
    pub fn get_tracks(&self) -> Vec<Track> {
        // Get audio/video/subtitle tracks
    }
    
    pub fn set_audio_track(&self, track_id: i32) {
        // Switch audio track
    }
    
    pub fn set_subtitle_track(&self, track_id: i32) {
        // Switch subtitle track
    }
}
```

### 2. Tauri Commands
```rust
// src-tauri/src/main.rs
#[tauri::command]
async fn vlc_play(uri: String, state: tauri::State<'_, VLCHandler>) -> Result<(), String> {
    state.play(&uri).map_err(|e| e.to_string())
}

#[tauri::command]
async fn vlc_pause(state: tauri::State<'_, VLCHandler>) {
    state.pause();
}

#[tauri::command]
async fn vlc_get_tracks(state: tauri::State<'_, VLCHandler>) -> Vec<Track> {
    state.get_tracks()
}
```

### 3. Frontend VideoPlayer (React)
```tsx
// src/components/VideoPlayer.tsx
import { useEffect, useRef, useState } from 'react';
import { invoke } from '@tauri-apps/api/tauri';

interface VideoPlayerProps {
  mediaId: string;
  streamUrl: string;
}

export function VideoPlayer({ mediaId, streamUrl }: VideoPlayerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [position, setPosition] = useState(0);
  const [tracks, setTracks] = useState<Track[]>([]);

  useEffect(() => {
    // Initialize VLC and start playback
    invoke('vlc_play', { uri: streamUrl });
    
    // Set up event listeners for VLC events
    const unlisten = listen('vlc-event', (event) => {
      // Handle position updates, state changes, etc.
    });
    
    return () => {
      invoke('vlc_stop');
      unlisten.then(f => f());
    };
  }, [streamUrl]);

  const handlePlayPause = () => {
    if (isPlaying) {
      invoke('vlc_pause');
    } else {
      invoke('vlc_resume');
    }
    setIsPlaying(!isPlaying);
  };

  const handleSeek = (newPosition: number) => {
    invoke('vlc_seek', { position: newPosition });
  };

  return (
    <div className="video-player">
      <canvas ref={canvasRef} className="video-canvas" />
      
      <div className="controls">
        <button onClick={handlePlayPause}>
          {isPlaying ? 'Pause' : 'Play'}
        </button>
        
        <input
          type="range"
          min={0}
          max={1}
          step={0.001}
          value={position}
          onChange={(e) => handleSeek(parseFloat(e.target.value))}
        />
        
        <TrackSelector
          tracks={tracks}
          onAudioChange={(id) => invoke('vlc_set_audio_track', { trackId: id })}
          onSubtitleChange={(id) => invoke('vlc_set_subtitle_track', { trackId: id })}
        />
      </div>
    </div>
  );
}
```

## Testing Strategy

### Rust Tests
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_vlc_initialization() {
        let vlc = VLCHandler::new().unwrap();
        assert!(vlc.is_initialized());
    }

    #[test]
    fn test_playback_controls() {
        let vlc = VLCHandler::new().unwrap();
        vlc.play("file:///test.mp4").unwrap();
        vlc.pause();
        assert_eq!(vlc.get_state(), PlayerState::Paused);
    }
}
```

### UI Tests
- Component rendering
- Control interactions
- Error handling

### HelixQA Integration
- Automated playback scenarios
- Track switching
- Error recovery

## Platform-Specific Notes

### Linux
- Requires `libvlc-dev` package
- X11/Wayland display support
- Hardware acceleration via VA-API/VDPAU

### macOS
- Requires VLC app with framework
- Metal/OpenGL rendering
- Sandboxing considerations

### Windows
- VLC SDK installation
- DirectX rendering
- File path handling

## Documentation

- Installation guide (VLC dependencies)
- API documentation
- Troubleshooting guide
- Build instructions

## Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| 1 | 1 week | Rust backend, basic playback |
| 2 | 1 week | React UI, controls |
| 3 | 1 week | Advanced features |
| 4 | 1 week | Desktop features |
| 5 | 1 week | Testing, docs |

**Total: 5 weeks (parallel with Android)**
