# VLC Media Player Integration for Catalogizer

This directory contains comprehensive plans for integrating VLC Media Player as the primary playback engine across all Catalogizer platforms.

## Platforms

### 📱 Android & Android TV
- **Location**: `catalogizer-androidtv/`
- **Approach**: libvlc-android library integration
- **UI**: Custom TV-optimized player interface
- **Plan**: [VLC_INTEGRATION_PLAN.md](VLC_INTEGRATION_PLAN.md)

### 🖥️ Desktop (Tauri)
- **Location**: `catalogizer-desktop/`
- **Approach**: libvlc-sys Rust bindings
- **UI**: React-based player with Rust backend
- **Plan**: [VLC_DESKTOP_INTEGRATION_PLAN.md](VLC_DESKTOP_INTEGRATION_PLAN.md)

## Common Features (All Platforms)

- ✅ Video/Audio playback
- ✅ Subtitle management (selection, delay)
- ✅ Audio track selection
- ✅ Video track/quality selection
- ✅ Playback controls (play, pause, stop, seek)
- ✅ Playback speed control
- ✅ Aspect ratio settings
- ✅ Volume control
- ✅ Fullscreen support

## Implementation Status

| Platform | Phase | Status |
|----------|-------|--------|
| Android TV | Planning | ✅ Plan complete |
| Desktop | Planning | ✅ Plan complete |

## Next Steps

1. **Start Android TV Phase 1**: Add libvlc-android dependencies
2. **Start Desktop Phase 1**: Add libvlc-sys Rust crate
3. **Parallel Development**: Both platforms can proceed simultaneously
4. **Testing**: HelixQA integration for automated testing

## Timeline

- **Android TV**: 5 weeks
- **Desktop**: 5 weeks
- **Parallel total**: 5 weeks
- **QA & Polish**: +1 week

**Estimated completion: 6 weeks from start**

## Resources

- **VLC Android**: https://github.com/videolan/vlc-android
- **VLC Desktop**: https://github.com/videolan/vlc
- **libvlc-sys**: https://docs.rs/libvlc-sys
- **libvlc-android**: https://code.videolan.org/videolan/libvlc-android-samples
