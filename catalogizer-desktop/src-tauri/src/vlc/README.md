# VLC Player Integration for Catalogizer Desktop

## Overview

This module provides VLC media player integration for the Catalogizer Desktop application using the `libvlc-sys` Rust bindings.

## Architecture

```
React Frontend (TypeScript)
    ↓ Tauri Commands
src-tauri/src/vlc/commands.rs
    ↓ Rust FFI
libvlc-sys (Rust bindings)
    ↓ C ABI
libvlc (VLC native library)
```

## Files

- `mod.rs` - Core VLCPlayer struct and libvlc wrapper
- `commands.rs` - Tauri commands for frontend integration
- `README.md` - This documentation

## Tauri Commands

| Command | Description |
|---------|-------------|
| `vlc_play` | Start playback from URL |
| `vlc_pause` | Pause playback |
| `vlc_resume` | Resume playback |
| `vlc_stop` | Stop playback |
| `vlc_seek` | Seek to position (0.0-1.0) |
| `vlc_get_status` | Get current playback status |
| `vlc_set_volume` | Set volume (0-100) |
| `vlc_toggle_mute` | Toggle mute state |
| `vlc_set_rate` | Set playback speed |
| `vlc_get_tracks` | Get audio/subtitle/video tracks |
| `vlc_set_audio_track` | Set active audio track |
| `vlc_set_subtitle_track` | Set active subtitle track |

## Usage

### Frontend (React)

```typescript
import { invoke } from '@tauri-apps/api/core';

// Play media
await invoke('vlc_play', { 
  request: { 
    url: 'https://api.example.com/media/123/stream',
    mediaId: 123,
    title: 'My Movie'
  } 
});

// Get status
const status = await invoke('vlc_get_status');
console.log(status.isPlaying, status.time);

// Control playback
await invoke('vlc_pause');
await invoke('vlc_seek', { position: 0.5 });
```

### Backend (Rust)

```rust
use crate::vlc::{VLCPlayer, PlaybackState};

let player = VLCPlayer::new()?;
player.play("https://example.com/video.mp4")?;
player.set_volume(80);
player.seek(0.5);
```

## Building

The VLC integration is enabled by the `vlc-player` feature (enabled by default).

```bash
# Build with VLC support
cd catalogizer-desktop/src-tauri
cargo build --features vlc-player

# Build without VLC
cargo build --no-default-features
```

## Dependencies

- `libvlc-sys` - Rust FFI bindings to libvlc
- System VLC libraries must be installed:
  - Ubuntu/Debian: `sudo apt install libvlc-dev`
  - macOS: `brew install vlc`
  - Windows: Include VLC SDK in build path

## License

Same as Catalogizer Desktop (TBD)
