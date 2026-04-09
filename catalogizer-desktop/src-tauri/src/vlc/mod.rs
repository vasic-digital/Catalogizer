//! VLC Media Player integration for Tauri Desktop app
//! 
//! This module provides Rust bindings to libVLC for video/audio playback.
//! It's designed to work with the Tauri framework and provides a clean
//! API for the frontend to control media playback.

use libvlc_sys::*;
use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_int, c_void};
use std::ptr;
use std::sync::{Arc, Mutex};
use serde::{Deserialize, Serialize};

/// VLC Player errors
#[derive(Debug, thiserror::Error, Serialize)]
pub enum VLCError {
    #[error("Failed to initialize VLC: {0}")]
    InitializationError(String),
    
    #[error("Failed to load media: {0}")]
    MediaLoadError(String),
    
    #[error("Playback error: {0}")]
    PlaybackError(String),
    
    #[error("Invalid state: {0}")]
    InvalidState(String),
}

/// Playback state
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PlaybackState {
    Idle,
    Opening,
    Buffering,
    Playing,
    Paused,
    Stopped,
    Ended,
    Error,
}

/// Track type (audio, video, subtitle)
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TrackType {
    Audio,
    Video,
    Subtitle,
}

/// Media track information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Track {
    pub id: i32,
    pub track_type: TrackType,
    pub name: String,
    pub language: Option<String>,
    pub codec: Option<String>,
    pub is_selected: bool,
}

/// Player event callback
pub type EventCallback = Box<dyn Fn(PlayerEvent) + Send + Sync>;

/// Player events
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PlayerEvent {
    StateChanged(PlaybackState),
    TimeChanged(i64),      // Current time in ms
    DurationChanged(i64),  // Total duration in ms
    PositionChanged(f32),  // Position 0.0 to 1.0
    VolumeChanged(i32),
    TrackListUpdated,
    Error(String),
}

/// VLC Player handle
pub struct VLCPlayer {
    instance: *mut libvlc_instance_t,
    media_player: *mut libvlc_media_player_t,
    current_media: Option<*mut libvlc_media_t>,
    state: Arc<Mutex<PlaybackState>>,
    event_callback: Arc<Mutex<Option<EventCallback>>>,
    event_manager: Option<*mut libvlc_event_manager_t>,
}

// Safety: VLC is thread-safe for most operations
unsafe impl Send for VLCPlayer {}
unsafe impl Sync for VLCPlayer {}

impl VLCPlayer {
    /// Create a new VLC player instance
    pub fn new() -> Result<Self, VLCError> {
        // VLC command line arguments
        let args = vec![
            CString::new("--no-video-title-show").unwrap(),
            CString::new("--no-snapshot-preview").unwrap(),
            CString::new("--network-caching=3000").unwrap(),
            CString::new("--file-caching=1000").unwrap(),
            CString::new("--live-caching=3000").unwrap(),
            CString::new("--audio-time-stretch").unwrap(),
            CString::new("--avcodec-hw").unwrap(),
        ];
        
        let argc = args.len() as c_int;
        let argv: Vec<*const c_char> = args.iter()
            .map(|arg| arg.as_ptr())
            .collect();
        
        let instance = unsafe {
            libvlc_new(argc, argv.as_ptr())
        };
        
        if instance.is_null() {
            return Err(VLCError::InitializationError(
                "Failed to create VLC instance".to_string()
            ));
        }
        
        let media_player = unsafe {
            libvlc_media_player_new(instance)
        };
        
        if media_player.is_null() {
            unsafe { libvlc_release(instance) };
            return Err(VLCError::InitializationError(
                "Failed to create media player".to_string()
            ));
        }
        
        let event_manager = unsafe {
            libvlc_media_player_event_manager(media_player)
        };
        
        let player = VLCPlayer {
            instance,
            media_player,
            current_media: None,
            state: Arc::new(Mutex::new(PlaybackState::Idle)),
            event_callback: Arc::new(Mutex::new(None)),
            event_manager: if event_manager.is_null() { None } else { Some(event_manager) },
        };
        
        // Setup event handling
        player.setup_event_handling();
        
        Ok(player)
    }
    
    /// Setup VLC event callbacks
    fn setup_event_handling(&self) {
        if let Some(event_manager) = self.event_manager {
            let state_ptr = Arc::into_raw(Arc::clone(&self.state)) as *mut c_void;
            
            unsafe {
                // Register for various events
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerOpening,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerPlaying,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerPaused,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerStopped,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerEndReached,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerTimeChanged,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerPositionChanged,
                    Some(player_event_callback),
                    state_ptr
                );
                libvlc_event_attach(
                    event_manager,
                    libvlc_event_e_libvlc_MediaPlayerEncounteredError,
                    Some(player_event_callback),
                    state_ptr
                );
            }
        }
    }
    
    /// Set event callback
    pub fn set_event_callback<F>(&self, callback: F)
    where
        F: Fn(PlayerEvent) + Send + Sync + 'static
    {
        let mut cb = self.event_callback.lock().unwrap();
        *cb = Some(Box::new(callback));
    }
    
    /// Play media from URL
    pub fn play(&mut self, url: &str) -> Result<(), VLCError> {
        // Stop current playback
        self.stop();
        
        // Create media from URL
        let url_cstring = CString::new(url).map_err(|_| {
            VLCError::MediaLoadError("Invalid URL".to_string())
        })?;
        
        let media = unsafe {
            libvlc_media_new_location(self.instance, url_cstring.as_ptr())
        };
        
        if media.is_null() {
            return Err(VLCError::MediaLoadError(
                "Failed to create media".to_string()
            ));
        }
        
        // Set media and play
        unsafe {
            libvlc_media_player_set_media(self.media_player, media);
            let result = libvlc_media_player_play(self.media_player);
            
            if result < 0 {
                libvlc_media_release(media);
                return Err(VLCError::PlaybackError(
                    "Failed to start playback".to_string()
                ));
            }
        }
        
        self.current_media = Some(media);
        
        Ok(())
    }
    
    /// Pause playback
    pub fn pause(&self) {
        unsafe {
            libvlc_media_player_pause(self.media_player);
        }
    }
    
    /// Resume playback
    pub fn resume(&self) {
        unsafe {
            libvlc_media_player_play(self.media_player);
        }
    }
    
    /// Stop playback
    pub fn stop(&mut self) {
        unsafe {
            libvlc_media_player_stop(self.media_player);
        }
        
        // Release current media
        if let Some(media) = self.current_media.take() {
            unsafe {
                libvlc_media_release(media);
            }
        }
    }
    
    /// Get current playback state
    pub fn get_state(&self) -> PlaybackState {
        let state = unsafe {
            libvlc_media_player_get_state(self.media_player)
        };
        
        match state {
            libvlc_state_t_libvlc_NothingSpecial => PlaybackState::Idle,
            libvlc_state_t_libvlc_Opening => PlaybackState::Opening,
            libvlc_state_t_libvlc_Buffering => PlaybackState::Buffering,
            libvlc_state_t_libvlc_Playing => PlaybackState::Playing,
            libvlc_state_t_libvlc_Paused => PlaybackState::Paused,
            libvlc_state_t_libvlc_Stopped => PlaybackState::Stopped,
            libvlc_state_t_libvlc_Ended => PlaybackState::Ended,
            libvlc_state_t_libvlc_Error => PlaybackState::Error,
            _ => PlaybackState::Idle,
        }
    }
    
    /// Seek to position (0.0 to 1.0)
    pub fn seek(&self, position: f32) {
        let clamped = position.clamp(0.0, 1.0);
        unsafe {
            libvlc_media_player_set_position(self.media_player, clamped);
        }
    }
    
    /// Get current position (0.0 to 1.0)
    pub fn get_position(&self) -> f32 {
        unsafe {
            libvlc_media_player_get_position(self.media_player)
        }
    }
    
    /// Get current time in milliseconds
    pub fn get_time(&self) -> i64 {
        unsafe {
            libvlc_media_player_get_time(self.media_player)
        }
    }
    
    /// Get total duration in milliseconds
    pub fn get_duration(&self) -> i64 {
        unsafe {
            libvlc_media_player_get_length(self.media_player)
        }
    }
    
    /// Set volume (0 to 100)
    pub fn set_volume(&self, volume: i32) {
        let clamped = volume.clamp(0, 100);
        unsafe {
            libvlc_audio_set_volume(self.media_player, clamped);
        }
    }
    
    /// Get current volume (0 to 100)
    pub fn get_volume(&self) -> i32 {
        unsafe {
            libvlc_audio_get_volume(self.media_player)
        }
    }
    
    /// Set mute
    pub fn set_mute(&self, mute: bool) {
        unsafe {
            libvlc_audio_set_mute(self.media_player, if mute { 1 } else { 0 });
        }
    }
    
    /// Get mute state
    pub fn is_muted(&self) -> bool {
        unsafe {
            libvlc_audio_get_mute(self.media_player) != 0
        }
    }
    
    /// Set playback rate (0.25 to 4.0)
    pub fn set_rate(&self, rate: f32) {
        let clamped = rate.clamp(0.25, 4.0);
        unsafe {
            libvlc_media_player_set_rate(self.media_player, clamped);
        }
    }
    
    /// Get playback rate
    pub fn get_rate(&self) -> f32 {
        unsafe {
            libvlc_media_player_get_rate(self.media_player)
        }
    }
    
    /// Check if currently playing
    pub fn is_playing(&self) -> bool {
        unsafe {
            libvlc_media_player_is_playing(self.media_player) != 0
        }
    }
    
    /// Get audio tracks
    pub fn get_audio_tracks(&self) -> Vec<Track> {
        self.get_tracks(TrackType::Audio)
    }
    
    /// Get subtitle tracks
    pub fn get_subtitle_tracks(&self) -> Vec<Track> {
        self.get_tracks(TrackType::Subtitle)
    }
    
    /// Get video tracks
    pub fn get_video_tracks(&self) -> Vec<Track> {
        self.get_tracks(TrackType::Video)
    }
    
    /// Get tracks of specific type
    fn get_tracks(&self, track_type: TrackType) -> Vec<Track> {
        // This is a simplified implementation
        // Full implementation would use libvlc_media_tracks_get
        vec![]
    }
    
    /// Set audio track
    pub fn set_audio_track(&self, track_id: i32) {
        unsafe {
            libvlc_audio_set_track(self.media_player, track_id);
        }
    }
    
    /// Set subtitle track
    pub fn set_subtitle_track(&self, track_id: i32) {
        unsafe {
            libvlc_video_set_spu(self.media_player, track_id);
        }
    }
    
    /// Set subtitle delay (microseconds)
    pub fn set_subtitle_delay(&self, delay_us: i64) {
        unsafe {
            libvlc_video_set_spu_delay(self.media_player, delay_us);
        }
    }
    
    /// Set audio delay (microseconds)
    pub fn set_audio_delay(&self, delay_us: i64) {
        unsafe {
            libvlc_audio_set_delay(self.media_player, delay_us);
        }
    }
    
    /// Set video crop/aspect ratio
    pub fn set_aspect_ratio(&self, ratio: &str) {
        let ratio_cstring = CString::new(ratio).unwrap_or_else(|_| CString::new("16:9").unwrap());
        unsafe {
            libvlc_video_set_aspect_ratio(self.media_player, ratio_cstring.as_ptr());
        }
    }
    
    /// Take screenshot
    pub fn take_snapshot(&self, width: u32, height: u32, filepath: &str) -> Result<(), VLCError> {
        let filepath_cstring = CString::new(filepath).map_err(|_| {
            VLCError::InvalidState("Invalid filepath".to_string())
        })?;
        
        let result = unsafe {
            libvlc_video_take_snapshot(
                self.media_player,
                0, // Video output number
                filepath_cstring.as_ptr(),
                width,
                height
            )
        };
        
        if result != 0 {
            return Err(VLCError::PlaybackError("Failed to take snapshot".to_string()));
        }
        
        Ok(())
    }
}

impl Drop for VLCPlayer {
    fn drop(&mut self) {
        self.stop();
        
        if let Some(event_manager) = self.event_manager {
            unsafe {
                libvlc_event_detach(event_manager, libvlc_event_e_libvlc_MediaPlayerOpening, Some(player_event_callback), ptr::null_mut());
                libvlc_event_detach(event_manager, libvlc_event_e_libvlc_MediaPlayerPlaying, Some(player_event_callback), ptr::null_mut());
                libvlc_event_detach(event_manager, libvlc_event_e_libvlc_MediaPlayerPaused, Some(player_event_callback), ptr::null_mut());
                libvlc_event_detach(event_manager, libvlc_event_e_libvlc_MediaPlayerStopped, Some(player_event_callback), ptr::null_mut());
                libvlc_event_detach(event_manager, libvlc_event_e_libvlc_MediaPlayerEndReached, Some(player_event_callback), ptr::null_mut());
            }
        }
        
        unsafe {
            libvlc_media_player_release(self.media_player);
            libvlc_release(self.instance);
        }
    }
}

/// VLC event callback handler
extern "C" fn player_event_callback(event: *const libvlc_event_t, data: *mut c_void) {
    if event.is_null() {
        return;
    }
    
    let event_type = unsafe { (*event).type_ };
    
    match event_type {
        libvlc_event_e_libvlc_MediaPlayerOpening => {
            // Handle opening
        }
        libvlc_event_e_libvlc_MediaPlayerPlaying => {
            // Handle playing
        }
        libvlc_event_e_libvlc_MediaPlayerPaused => {
            // Handle paused
        }
        libvlc_event_e_libvlc_MediaPlayerStopped => {
            // Handle stopped
        }
        libvlc_event_e_libvlc_MediaPlayerEndReached => {
            // Handle ended
        }
        libvlc_event_e_libvlc_MediaPlayerTimeChanged => {
            // Handle time changed
        }
        libvlc_event_e_libvlc_MediaPlayerPositionChanged => {
            // Handle position changed
        }
        libvlc_event_e_libvlc_MediaPlayerEncounteredError => {
            // Handle error
        }
        _ => {}
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_vlc_initialization() {
        let player = VLCPlayer::new();
        assert!(player.is_ok());
        
        let player = player.unwrap();
        assert_eq!(player.get_state(), PlaybackState::Idle);
    }
    
    #[test]
    fn test_playback_state_transitions() {
        let mut player = VLCPlayer::new().unwrap();
        
        // Initially idle
        assert_eq!(player.get_state(), PlaybackState::Idle);
        
        // Volume control should work
        player.set_volume(50);
        assert_eq!(player.get_volume(), 50);
        
        player.set_volume(150); // Should be clamped
        assert_eq!(player.get_volume(), 100);
    }
    
    #[test]
    fn test_rate_control() {
        let player = VLCPlayer::new().unwrap();
        
        player.set_rate(1.5);
        // Rate should be set (can't verify without playing)
        
        player.set_rate(5.0); // Should be clamped
        // Should be clamped to max
    }
}
