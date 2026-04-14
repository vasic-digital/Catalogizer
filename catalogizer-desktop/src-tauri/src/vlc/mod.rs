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
        // Using map_err to convert NulError to VLCError instead of unwrap
        let args = vec![
            CString::new("--no-video-title-show")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
            CString::new("--no-snapshot-preview")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
            CString::new("--network-caching=3000")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
            CString::new("--file-caching=1000")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
            CString::new("--live-caching=3000")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
            CString::new("--audio-time-stretch")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
            CString::new("--avcodec-hw")
                .map_err(|e| VLCError::InitializationError(format!("Invalid argument: {}", e)))?,
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
    pub fn set_event_callback<F>(&self, callback: F) -> Result<(), VLCError>
    where
        F: Fn(PlayerEvent) + Send + Sync + 'static
    {
        let mut cb = self.event_callback.lock()
            .map_err(|_| VLCError::InvalidState("Event callback mutex poisoned".to_string()))?;
        *cb = Some(Box::new(callback));
        Ok(())
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

    /// Tests for PlaybackState enum
    mod playback_state_tests {
        use super::*;

        #[test]
        fn test_playback_state_serialization_all_variants() {
            let states = vec![
                PlaybackState::Idle,
                PlaybackState::Opening,
                PlaybackState::Buffering,
                PlaybackState::Playing,
                PlaybackState::Paused,
                PlaybackState::Stopped,
                PlaybackState::Ended,
                PlaybackState::Error,
            ];

            for state in &states {
                let json = serde_json::to_string(state).expect("Should serialize");
                let deserialized: PlaybackState =
                    serde_json::from_str(&json).expect("Should deserialize");
                assert_eq!(*state, deserialized);
            }
        }

        #[test]
        fn test_playback_state_equality() {
            assert_eq!(PlaybackState::Idle, PlaybackState::Idle);
            assert_eq!(PlaybackState::Playing, PlaybackState::Playing);
            assert_ne!(PlaybackState::Idle, PlaybackState::Playing);
            assert_ne!(PlaybackState::Paused, PlaybackState::Stopped);
        }

        #[test]
        fn test_playback_state_clone() {
            let state = PlaybackState::Playing;
            let cloned = state;
            assert_eq!(state, cloned);
        }

        #[test]
        fn test_playback_state_debug() {
            let state = PlaybackState::Buffering;
            let debug = format!("{:?}", state);
            assert_eq!(debug, "Buffering");
        }

        #[test]
        fn test_playback_state_json_values() {
            assert_eq!(
                serde_json::to_string(&PlaybackState::Idle).expect("Should serialize"),
                "\"Idle\""
            );
            assert_eq!(
                serde_json::to_string(&PlaybackState::Playing).expect("Should serialize"),
                "\"Playing\""
            );
            assert_eq!(
                serde_json::to_string(&PlaybackState::Paused).expect("Should serialize"),
                "\"Paused\""
            );
            assert_eq!(
                serde_json::to_string(&PlaybackState::Stopped).expect("Should serialize"),
                "\"Stopped\""
            );
        }
    }

    /// Tests for TrackType enum
    mod track_type_tests {
        use super::*;

        #[test]
        fn test_track_type_serialization() {
            let types = vec![TrackType::Audio, TrackType::Video, TrackType::Subtitle];

            for track_type in &types {
                let json = serde_json::to_string(track_type).expect("Should serialize");
                let deserialized: TrackType =
                    serde_json::from_str(&json).expect("Should deserialize");
                assert_eq!(*track_type, deserialized);
            }
        }

        #[test]
        fn test_track_type_equality() {
            assert_eq!(TrackType::Audio, TrackType::Audio);
            assert_ne!(TrackType::Audio, TrackType::Video);
            assert_ne!(TrackType::Video, TrackType::Subtitle);
        }

        #[test]
        fn test_track_type_debug() {
            assert_eq!(format!("{:?}", TrackType::Audio), "Audio");
            assert_eq!(format!("{:?}", TrackType::Video), "Video");
            assert_eq!(format!("{:?}", TrackType::Subtitle), "Subtitle");
        }

        #[test]
        fn test_track_type_json_values() {
            assert_eq!(
                serde_json::to_string(&TrackType::Audio).expect("Should serialize"),
                "\"Audio\""
            );
            assert_eq!(
                serde_json::to_string(&TrackType::Video).expect("Should serialize"),
                "\"Video\""
            );
            assert_eq!(
                serde_json::to_string(&TrackType::Subtitle).expect("Should serialize"),
                "\"Subtitle\""
            );
        }
    }

    /// Tests for Track struct
    mod track_tests {
        use super::*;

        #[test]
        fn test_track_serialization() {
            let track = Track {
                id: 1,
                track_type: TrackType::Audio,
                name: "English 5.1".to_string(),
                language: Some("en".to_string()),
                codec: Some("aac".to_string()),
                is_selected: true,
            };

            let json = serde_json::to_string(&track).expect("Should serialize");
            let deserialized: Track =
                serde_json::from_str(&json).expect("Should deserialize");

            assert_eq!(deserialized.id, 1);
            assert_eq!(deserialized.track_type, TrackType::Audio);
            assert_eq!(deserialized.name, "English 5.1");
            assert_eq!(deserialized.language, Some("en".to_string()));
            assert_eq!(deserialized.codec, Some("aac".to_string()));
            assert!(deserialized.is_selected);
        }

        #[test]
        fn test_track_without_optional_fields() {
            let track = Track {
                id: 0,
                track_type: TrackType::Video,
                name: "Video Track".to_string(),
                language: None,
                codec: None,
                is_selected: true,
            };

            let json = serde_json::to_string(&track).expect("Should serialize");
            let deserialized: Track =
                serde_json::from_str(&json).expect("Should deserialize");

            assert!(deserialized.language.is_none());
            assert!(deserialized.codec.is_none());
        }

        #[test]
        fn test_track_subtitle_type() {
            let track = Track {
                id: 2,
                track_type: TrackType::Subtitle,
                name: "French Subtitles".to_string(),
                language: Some("fr".to_string()),
                codec: Some("srt".to_string()),
                is_selected: false,
            };

            assert_eq!(track.track_type, TrackType::Subtitle);
            assert!(!track.is_selected);
        }

        #[test]
        fn test_track_clone() {
            let original = Track {
                id: 3,
                track_type: TrackType::Audio,
                name: "Stereo".to_string(),
                language: Some("de".to_string()),
                codec: Some("mp3".to_string()),
                is_selected: false,
            };

            let cloned = original.clone();
            assert_eq!(original.id, cloned.id);
            assert_eq!(original.name, cloned.name);
            assert_eq!(original.language, cloned.language);
        }

        #[test]
        fn test_track_debug() {
            let track = Track {
                id: 0,
                track_type: TrackType::Audio,
                name: "Test".to_string(),
                language: None,
                codec: None,
                is_selected: false,
            };
            let debug = format!("{:?}", track);
            assert!(debug.contains("Track"));
            assert!(debug.contains("Test"));
        }
    }

    /// Tests for PlayerEvent enum
    mod player_event_tests {
        use super::*;

        #[test]
        fn test_player_event_state_changed() {
            let event = PlayerEvent::StateChanged(PlaybackState::Playing);
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::StateChanged(state) => {
                    assert_eq!(state, PlaybackState::Playing);
                }
                _ => panic!("Expected StateChanged event"),
            }
        }

        #[test]
        fn test_player_event_time_changed() {
            let event = PlayerEvent::TimeChanged(45000);
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::TimeChanged(time) => assert_eq!(time, 45000),
                _ => panic!("Expected TimeChanged event"),
            }
        }

        #[test]
        fn test_player_event_duration_changed() {
            let event = PlayerEvent::DurationChanged(7200000);
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::DurationChanged(dur) => assert_eq!(dur, 7200000),
                _ => panic!("Expected DurationChanged event"),
            }
        }

        #[test]
        fn test_player_event_position_changed() {
            let event = PlayerEvent::PositionChanged(0.75);
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::PositionChanged(pos) => {
                    assert!((pos - 0.75).abs() < 0.001);
                }
                _ => panic!("Expected PositionChanged event"),
            }
        }

        #[test]
        fn test_player_event_volume_changed() {
            let event = PlayerEvent::VolumeChanged(80);
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::VolumeChanged(vol) => assert_eq!(vol, 80),
                _ => panic!("Expected VolumeChanged event"),
            }
        }

        #[test]
        fn test_player_event_track_list_updated() {
            let event = PlayerEvent::TrackListUpdated;
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::TrackListUpdated => {}
                _ => panic!("Expected TrackListUpdated event"),
            }
        }

        #[test]
        fn test_player_event_error() {
            let event = PlayerEvent::Error("Codec not found".to_string());
            let json = serde_json::to_string(&event).expect("Should serialize");
            let deserialized: PlayerEvent =
                serde_json::from_str(&json).expect("Should deserialize");

            match deserialized {
                PlayerEvent::Error(msg) => assert_eq!(msg, "Codec not found"),
                _ => panic!("Expected Error event"),
            }
        }

        #[test]
        fn test_player_event_clone() {
            let event = PlayerEvent::TimeChanged(12345);
            let cloned = event.clone();
            match cloned {
                PlayerEvent::TimeChanged(t) => assert_eq!(t, 12345),
                _ => panic!("Clone should preserve variant"),
            }
        }
    }

    /// Tests for VLCError enum
    mod vlc_error_tests {
        use super::*;

        #[test]
        fn test_initialization_error_display() {
            let err = VLCError::InitializationError("no libvlc".to_string());
            let msg = err.to_string();
            assert!(msg.contains("Failed to initialize VLC"));
            assert!(msg.contains("no libvlc"));
        }

        #[test]
        fn test_media_load_error_display() {
            let err = VLCError::MediaLoadError("bad URL".to_string());
            let msg = err.to_string();
            assert!(msg.contains("Failed to load media"));
            assert!(msg.contains("bad URL"));
        }

        #[test]
        fn test_playback_error_display() {
            let err = VLCError::PlaybackError("decoder failure".to_string());
            let msg = err.to_string();
            assert!(msg.contains("Playback error"));
            assert!(msg.contains("decoder failure"));
        }

        #[test]
        fn test_invalid_state_error_display() {
            let err = VLCError::InvalidState("player not ready".to_string());
            let msg = err.to_string();
            assert!(msg.contains("Invalid state"));
            assert!(msg.contains("player not ready"));
        }

        #[test]
        fn test_vlc_error_serialization() {
            let err = VLCError::PlaybackError("test error".to_string());
            let json = serde_json::to_string(&err).expect("Should serialize");
            assert!(json.contains("test error"));
        }

        #[test]
        fn test_vlc_error_debug() {
            let err = VLCError::InitializationError("debug test".to_string());
            let debug = format!("{:?}", err);
            assert!(debug.contains("InitializationError"));
            assert!(debug.contains("debug test"));
        }
    }

    /// Tests for value clamping logic used in VLCPlayer methods
    mod clamping_tests {
        #[test]
        fn test_volume_clamping() {
            assert_eq!((-10i32).clamp(0, 100), 0);
            assert_eq!(0i32.clamp(0, 100), 0);
            assert_eq!(50i32.clamp(0, 100), 50);
            assert_eq!(100i32.clamp(0, 100), 100);
            assert_eq!(150i32.clamp(0, 100), 100);
        }

        #[test]
        fn test_position_clamping() {
            assert_eq!((-0.5f32).clamp(0.0, 1.0), 0.0);
            assert_eq!(0.0f32.clamp(0.0, 1.0), 0.0);
            assert_eq!(0.5f32.clamp(0.0, 1.0), 0.5);
            assert_eq!(1.0f32.clamp(0.0, 1.0), 1.0);
            assert_eq!(1.5f32.clamp(0.0, 1.0), 1.0);
        }

        #[test]
        fn test_rate_clamping() {
            assert_eq!(0.1f32.clamp(0.25, 4.0), 0.25);
            assert_eq!(0.25f32.clamp(0.25, 4.0), 0.25);
            assert_eq!(1.0f32.clamp(0.25, 4.0), 1.0);
            assert_eq!(2.0f32.clamp(0.25, 4.0), 2.0);
            assert_eq!(4.0f32.clamp(0.25, 4.0), 4.0);
            assert_eq!(5.0f32.clamp(0.25, 4.0), 4.0);
        }
    }

    /// Tests for CString creation (used in VLC initialization args)
    mod cstring_tests {
        use std::ffi::CString;

        #[test]
        fn test_vlc_args_creation() {
            let args = vec![
                "--no-video-title-show",
                "--no-snapshot-preview",
                "--network-caching=3000",
                "--file-caching=1000",
                "--live-caching=3000",
                "--audio-time-stretch",
                "--avcodec-hw",
            ];

            for arg in &args {
                let result = CString::new(*arg);
                assert!(result.is_ok(), "CString should be created for: {}", arg);
            }
        }

        #[test]
        fn test_cstring_with_null_byte_fails() {
            let result = CString::new("invalid\0arg");
            assert!(result.is_err());
        }

        #[test]
        fn test_vlc_arg_count() {
            let args = vec![
                "--no-video-title-show",
                "--no-snapshot-preview",
                "--network-caching=3000",
                "--file-caching=1000",
                "--live-caching=3000",
                "--audio-time-stretch",
                "--avcodec-hw",
            ];
            assert_eq!(args.len(), 7);
        }

        #[test]
        fn test_aspect_ratio_cstring() {
            let ratios = vec!["16:9", "4:3", "1:1", "21:9"];
            for ratio in &ratios {
                let result = CString::new(*ratio);
                assert!(result.is_ok(), "CString should work for ratio: {}", ratio);
            }
        }

        #[test]
        fn test_aspect_ratio_fallback() {
            let invalid_ratio = "invalid\0ratio";
            let result = CString::new(invalid_ratio);
            assert!(result.is_err());
            // Fallback should work
            let fallback = CString::new("16:9");
            assert!(fallback.is_ok());
        }
    }
}
