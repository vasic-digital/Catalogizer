//! Tauri commands for VLC player control
//! 
//! These commands provide the bridge between the React frontend
//! and the Rust VLC backend.

use tauri::State;
use std::sync::Mutex;
use serde::{Deserialize, Serialize};
use crate::vlc::{VLCPlayer, VLCError, PlaybackState, Track, PlayerEvent};

/// Shared VLC player state managed by Tauri
pub struct VLCPlayerState(pub Mutex<VLCPlayer>);

/// Request to play media
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlayRequest {
    pub url: String,
    pub media_id: Option<i64>,
    pub title: Option<String>,
}

/// Current playback status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PlaybackStatus {
    pub state: PlaybackState,
    pub position: f32,
    pub time: i64,
    pub duration: i64,
    pub volume: i32,
    pub is_playing: bool,
    pub is_muted: bool,
    pub rate: f32,
}

/// Track list response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackListResponse {
    pub audio_tracks: Vec<TrackInfo>,
    pub subtitle_tracks: Vec<TrackInfo>,
    pub video_tracks: Vec<TrackInfo>,
}

/// Track information for frontend
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrackInfo {
    pub id: i32,
    pub name: String,
    pub language: Option<String>,
    pub is_selected: bool,
}

impl From<Track> for TrackInfo {
    fn from(track: Track) -> Self {
        TrackInfo {
            id: track.id,
            name: track.name,
            language: track.language,
            is_selected: track.is_selected,
        }
    }
}

/// Initialize VLC player
#[tauri::command]
pub fn vlc_initialize(state: State<'_, VLCPlayerState>) -> Result<(), String> {
    // VLC is already initialized in the state
    Ok(())
}

/// Play media from URL
#[tauri::command]
pub fn vlc_play(
    request: PlayRequest,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let mut player = state.0.lock().map_err(|e| e.to_string())?;
    
    player.play(&request.url)
        .map_err(|e| e.to_string())
}

/// Pause playback
#[tauri::command]
pub fn vlc_pause(state: State<'_, VLCPlayerState>) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.pause();
    Ok(())
}

/// Resume playback
#[tauri::command]
pub fn vlc_resume(state: State<'_, VLCPlayerState>) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.resume();
    Ok(())
}

/// Stop playback
#[tauri::command]
pub fn vlc_stop(state: State<'_, VLCPlayerState>) -> Result<(), String> {
    let mut player = state.0.lock().map_err(|e| e.to_string())?;
    player.stop();
    Ok(())
}

/// Toggle play/pause
#[tauri::command]
pub fn vlc_toggle_playback(state: State<'_, VLCPlayerState>) -> Result<bool, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    
    if player.is_playing() {
        player.pause();
        Ok(false) // Now paused
    } else {
        player.resume();
        Ok(true) // Now playing
    }
}

/// Seek to position (0.0 to 1.0)
#[tauri::command]
pub fn vlc_seek(position: f32, state: State<'_, VLCPlayerState>) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.seek(position);
    Ok(())
}

/// Seek forward by milliseconds
#[tauri::command]
pub fn vlc_seek_forward(
    milliseconds: i64,
    state: State<'_, VLCPlayerState>
) -> Result<i64, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    
    let current = player.get_time();
    let duration = player.get_duration();
    let new_time = (current + milliseconds).min(duration);
    
    let position = if duration > 0 {
        new_time as f32 / duration as f32
    } else {
        0.0
    };
    
    player.seek(position);
    Ok(new_time)
}

/// Seek backward by milliseconds
#[tauri::command]
pub fn vlc_seek_backward(
    milliseconds: i64,
    state: State<'_, VLCPlayerState>
) -> Result<i64, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    
    let current = player.get_time();
    let new_time = (current - milliseconds).max(0);
    
    let duration = player.get_duration();
    let position = if duration > 0 {
        new_time as f32 / duration as f32
    } else {
        0.0
    };
    
    player.seek(position);
    Ok(new_time)
}

/// Get current playback status
#[tauri::command]
pub fn vlc_get_status(state: State<'_, VLCPlayerState>) -> Result<PlaybackStatus, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    
    Ok(PlaybackStatus {
        state: player.get_state(),
        position: player.get_position(),
        time: player.get_time(),
        duration: player.get_duration(),
        volume: player.get_volume(),
        is_playing: player.is_playing(),
        is_muted: player.is_muted(),
        rate: player.get_rate(),
    })
}

/// Set volume (0 to 100)
#[tauri::command]
pub fn vlc_set_volume(volume: i32, state: State<'_, VLCPlayerState>) -> Result<i32, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.set_volume(volume);
    Ok(player.get_volume())
}

/// Toggle mute
#[tauri::command]
pub fn vlc_toggle_mute(state: State<'_, VLCPlayerState>) -> Result<bool, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    
    let new_mute = !player.is_muted();
    player.set_mute(new_mute);
    Ok(new_mute)
}

/// Set playback rate (0.25 to 4.0)
#[tauri::command]
pub fn vlc_set_rate(rate: f32, state: State<'_, VLCPlayerState>) -> Result<f32, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.set_rate(rate);
    Ok(player.get_rate())
}

/// Get available tracks
#[tauri::command]
pub fn vlc_get_tracks(state: State<'_, VLCPlayerState>) -> Result<TrackListResponse, String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    
    let audio_tracks: Vec<TrackInfo> = player.get_audio_tracks()
        .into_iter()
        .map(TrackInfo::from)
        .collect();
    
    let subtitle_tracks: Vec<TrackInfo> = player.get_subtitle_tracks()
        .into_iter()
        .map(TrackInfo::from)
        .collect();
    
    let video_tracks: Vec<TrackInfo> = player.get_video_tracks()
        .into_iter()
        .map(TrackInfo::from)
        .collect();
    
    Ok(TrackListResponse {
        audio_tracks,
        subtitle_tracks,
        video_tracks,
    })
}

/// Set audio track
#[tauri::command]
pub fn vlc_set_audio_track(
    track_id: i32,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.set_audio_track(track_id);
    Ok(())
}

/// Set subtitle track
#[tauri::command]
pub fn vlc_set_subtitle_track(
    track_id: i32,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.set_subtitle_track(track_id);
    Ok(())
}

/// Set subtitle delay (milliseconds)
#[tauri::command]
pub fn vlc_set_subtitle_delay(
    delay_ms: i64,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    // Convert to microseconds for VLC
    player.set_subtitle_delay(delay_ms * 1000);
    Ok(())
}

/// Set audio delay (milliseconds)
#[tauri::command]
pub fn vlc_set_audio_delay(
    delay_ms: i64,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    // Convert to microseconds for VLC
    player.set_audio_delay(delay_ms * 1000);
    Ok(())
}

/// Set aspect ratio
#[tauri::command]
pub fn vlc_set_aspect_ratio(
    ratio: String,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.set_aspect_ratio(&ratio);
    Ok(())
}

/// Take screenshot
#[tauri::command]
pub fn vlc_take_snapshot(
    width: u32,
    height: u32,
    filepath: String,
    state: State<'_, VLCPlayerState>
) -> Result<(), String> {
    let player = state.0.lock().map_err(|e| e.to_string())?;
    player.take_snapshot(width, height, &filepath)
        .map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    /// Tests for PlayRequest struct
    mod play_request_tests {
        use super::*;

        #[test]
        fn test_play_request_serialization() {
            let request = PlayRequest {
                url: "http://localhost:8080/api/v1/stream/123".to_string(),
                media_id: Some(123),
                title: Some("Test Movie".to_string()),
            };

            let json = serde_json::to_string(&request).expect("Should serialize");
            let deserialized: PlayRequest =
                serde_json::from_str(&json).expect("Should deserialize");

            assert_eq!(deserialized.url, "http://localhost:8080/api/v1/stream/123");
            assert_eq!(deserialized.media_id, Some(123));
            assert_eq!(deserialized.title, Some("Test Movie".to_string()));
        }

        #[test]
        fn test_play_request_minimal() {
            let request = PlayRequest {
                url: "http://example.com/video.mp4".to_string(),
                media_id: None,
                title: None,
            };

            let json = serde_json::to_string(&request).expect("Should serialize");
            let deserialized: PlayRequest =
                serde_json::from_str(&json).expect("Should deserialize");

            assert_eq!(deserialized.url, "http://example.com/video.mp4");
            assert!(deserialized.media_id.is_none());
            assert!(deserialized.title.is_none());
        }

        #[test]
        fn test_play_request_from_json() {
            let json = r#"{
                "url": "smb://nas/movies/film.mkv",
                "media_id": 456,
                "title": "My Film"
            }"#;

            let request: PlayRequest =
                serde_json::from_str(json).expect("Should deserialize");
            assert_eq!(request.url, "smb://nas/movies/film.mkv");
            assert_eq!(request.media_id, Some(456));
            assert_eq!(request.title, Some("My Film".to_string()));
        }

        #[test]
        fn test_play_request_from_json_nulls() {
            let json = r#"{
                "url": "http://example.com/video.mp4",
                "media_id": null,
                "title": null
            }"#;

            let request: PlayRequest =
                serde_json::from_str(json).expect("Should deserialize");
            assert!(request.media_id.is_none());
            assert!(request.title.is_none());
        }

        #[test]
        fn test_play_request_clone() {
            let original = PlayRequest {
                url: "http://example.com/test.mp4".to_string(),
                media_id: Some(1),
                title: Some("Title".to_string()),
            };

            let cloned = original.clone();
            assert_eq!(original.url, cloned.url);
            assert_eq!(original.media_id, cloned.media_id);
            assert_eq!(original.title, cloned.title);
        }

        #[test]
        fn test_play_request_debug() {
            let request = PlayRequest {
                url: "http://test.com".to_string(),
                media_id: None,
                title: None,
            };
            let debug = format!("{:?}", request);
            assert!(debug.contains("PlayRequest"));
            assert!(debug.contains("http://test.com"));
        }
    }

    /// Tests for PlaybackStatus struct
    mod playback_status_tests {
        use super::*;

        #[test]
        fn test_playback_status_serialization() {
            let status = PlaybackStatus {
                state: PlaybackState::Playing,
                position: 0.5,
                time: 60000,
                duration: 120000,
                volume: 75,
                is_playing: true,
                is_muted: false,
                rate: 1.0,
            };

            let json = serde_json::to_string(&status).expect("Should serialize");
            let deserialized: PlaybackStatus =
                serde_json::from_str(&json).expect("Should deserialize");

            assert_eq!(deserialized.position, 0.5);
            assert_eq!(deserialized.time, 60000);
            assert_eq!(deserialized.duration, 120000);
            assert_eq!(deserialized.volume, 75);
            assert!(deserialized.is_playing);
            assert!(!deserialized.is_muted);
            assert_eq!(deserialized.rate, 1.0);
        }

        #[test]
        fn test_playback_status_idle() {
            let status = PlaybackStatus {
                state: PlaybackState::Idle,
                position: 0.0,
                time: 0,
                duration: 0,
                volume: 100,
                is_playing: false,
                is_muted: false,
                rate: 1.0,
            };

            assert!(!status.is_playing);
            assert_eq!(status.position, 0.0);
            assert_eq!(status.time, 0);
        }

        #[test]
        fn test_playback_status_paused() {
            let status = PlaybackStatus {
                state: PlaybackState::Paused,
                position: 0.3,
                time: 30000,
                duration: 100000,
                volume: 50,
                is_playing: false,
                is_muted: false,
                rate: 1.0,
            };

            assert!(!status.is_playing);
            assert_eq!(status.time, 30000);
        }

        #[test]
        fn test_playback_status_muted() {
            let status = PlaybackStatus {
                state: PlaybackState::Playing,
                position: 0.1,
                time: 10000,
                duration: 100000,
                volume: 0,
                is_playing: true,
                is_muted: true,
                rate: 1.0,
            };

            assert!(status.is_muted);
        }

        #[test]
        fn test_playback_status_fast_rate() {
            let status = PlaybackStatus {
                state: PlaybackState::Playing,
                position: 0.5,
                time: 50000,
                duration: 100000,
                volume: 50,
                is_playing: true,
                is_muted: false,
                rate: 2.0,
            };

            assert_eq!(status.rate, 2.0);
        }

        #[test]
        fn test_playback_status_clone() {
            let original = PlaybackStatus {
                state: PlaybackState::Playing,
                position: 0.75,
                time: 75000,
                duration: 100000,
                volume: 80,
                is_playing: true,
                is_muted: false,
                rate: 1.5,
            };

            let cloned = original.clone();
            assert_eq!(original.position, cloned.position);
            assert_eq!(original.time, cloned.time);
            assert_eq!(original.volume, cloned.volume);
            assert_eq!(original.rate, cloned.rate);
        }
    }

    /// Tests for TrackListResponse struct
    mod track_list_tests {
        use super::*;

        #[test]
        fn test_track_list_response_empty() {
            let response = TrackListResponse {
                audio_tracks: vec![],
                subtitle_tracks: vec![],
                video_tracks: vec![],
            };

            let json = serde_json::to_string(&response).expect("Should serialize");
            let deserialized: TrackListResponse =
                serde_json::from_str(&json).expect("Should deserialize");

            assert!(deserialized.audio_tracks.is_empty());
            assert!(deserialized.subtitle_tracks.is_empty());
            assert!(deserialized.video_tracks.is_empty());
        }

        #[test]
        fn test_track_list_response_with_tracks() {
            let response = TrackListResponse {
                audio_tracks: vec![
                    TrackInfo {
                        id: 0,
                        name: "English".to_string(),
                        language: Some("en".to_string()),
                        is_selected: true,
                    },
                    TrackInfo {
                        id: 1,
                        name: "Russian".to_string(),
                        language: Some("ru".to_string()),
                        is_selected: false,
                    },
                ],
                subtitle_tracks: vec![TrackInfo {
                    id: 0,
                    name: "English Subtitles".to_string(),
                    language: Some("en".to_string()),
                    is_selected: false,
                }],
                video_tracks: vec![TrackInfo {
                    id: 0,
                    name: "Video Track 1".to_string(),
                    language: None,
                    is_selected: true,
                }],
            };

            assert_eq!(response.audio_tracks.len(), 2);
            assert_eq!(response.subtitle_tracks.len(), 1);
            assert_eq!(response.video_tracks.len(), 1);
        }

        #[test]
        fn test_track_list_response_serialization_roundtrip() {
            let response = TrackListResponse {
                audio_tracks: vec![TrackInfo {
                    id: 0,
                    name: "Stereo".to_string(),
                    language: Some("en".to_string()),
                    is_selected: true,
                }],
                subtitle_tracks: vec![],
                video_tracks: vec![],
            };

            let json = serde_json::to_string(&response).expect("Should serialize");
            let deserialized: TrackListResponse =
                serde_json::from_str(&json).expect("Should deserialize");

            assert_eq!(deserialized.audio_tracks.len(), 1);
            assert_eq!(deserialized.audio_tracks[0].name, "Stereo");
            assert!(deserialized.audio_tracks[0].is_selected);
        }
    }

    /// Tests for TrackInfo struct
    mod track_info_tests {
        use super::*;

        #[test]
        fn test_track_info_serialization() {
            let info = TrackInfo {
                id: 42,
                name: "DTS-HD Master Audio".to_string(),
                language: Some("en".to_string()),
                is_selected: true,
            };

            let json = serde_json::to_string(&info).expect("Should serialize");
            let deserialized: TrackInfo =
                serde_json::from_str(&json).expect("Should deserialize");

            assert_eq!(deserialized.id, 42);
            assert_eq!(deserialized.name, "DTS-HD Master Audio");
            assert_eq!(deserialized.language, Some("en".to_string()));
            assert!(deserialized.is_selected);
        }

        #[test]
        fn test_track_info_without_language() {
            let info = TrackInfo {
                id: 0,
                name: "Track 1".to_string(),
                language: None,
                is_selected: false,
            };

            let json = serde_json::to_string(&info).expect("Should serialize");
            assert!(json.contains("\"language\":null"));
        }

        #[test]
        fn test_track_info_from_track() {
            let track = Track {
                id: 5,
                track_type: TrackType::Audio,
                name: "5.1 Surround".to_string(),
                language: Some("en".to_string()),
                codec: Some("aac".to_string()),
                is_selected: true,
            };

            let info = TrackInfo::from(track);
            assert_eq!(info.id, 5);
            assert_eq!(info.name, "5.1 Surround");
            assert_eq!(info.language, Some("en".to_string()));
            assert!(info.is_selected);
        }

        #[test]
        fn test_track_info_from_track_no_language() {
            let track = Track {
                id: 0,
                track_type: TrackType::Video,
                name: "Video".to_string(),
                language: None,
                codec: None,
                is_selected: true,
            };

            let info = TrackInfo::from(track);
            assert_eq!(info.id, 0);
            assert!(info.language.is_none());
        }

        #[test]
        fn test_track_info_clone() {
            let original = TrackInfo {
                id: 1,
                name: "Audio".to_string(),
                language: Some("fr".to_string()),
                is_selected: false,
            };

            let cloned = original.clone();
            assert_eq!(original.id, cloned.id);
            assert_eq!(original.name, cloned.name);
            assert_eq!(original.language, cloned.language);
        }

        #[test]
        fn test_track_info_debug() {
            let info = TrackInfo {
                id: 0,
                name: "Test".to_string(),
                language: None,
                is_selected: false,
            };
            let debug = format!("{:?}", info);
            assert!(debug.contains("TrackInfo"));
        }
    }

    /// Tests for seek calculation logic
    mod seek_calculation_tests {
        #[test]
        fn test_seek_forward_calculation() {
            let current: i64 = 30000; // 30 seconds
            let duration: i64 = 120000; // 2 minutes
            let skip: i64 = 10000; // 10 seconds

            let new_time = (current + skip).min(duration);
            assert_eq!(new_time, 40000);

            let position = if duration > 0 {
                new_time as f32 / duration as f32
            } else {
                0.0
            };
            let expected_pos = 40000.0 / 120000.0;
            assert!((position - expected_pos).abs() < 0.001);
        }

        #[test]
        fn test_seek_forward_clamped_to_duration() {
            let current: i64 = 115000;
            let duration: i64 = 120000;
            let skip: i64 = 10000;

            let new_time = (current + skip).min(duration);
            assert_eq!(new_time, 120000);
        }

        #[test]
        fn test_seek_backward_calculation() {
            let current: i64 = 30000;
            let skip: i64 = 10000;

            let new_time = (current - skip).max(0);
            assert_eq!(new_time, 20000);
        }

        #[test]
        fn test_seek_backward_clamped_to_zero() {
            let current: i64 = 5000;
            let skip: i64 = 10000;

            let new_time = (current - skip).max(0);
            assert_eq!(new_time, 0);
        }

        #[test]
        fn test_seek_position_calculation_zero_duration() {
            let duration: i64 = 0;
            let new_time: i64 = 0;

            let position = if duration > 0 {
                new_time as f32 / duration as f32
            } else {
                0.0
            };

            assert_eq!(position, 0.0);
        }

        #[test]
        fn test_seek_position_at_start() {
            let duration: i64 = 100000;
            let time: i64 = 0;

            let position = time as f32 / duration as f32;
            assert_eq!(position, 0.0);
        }

        #[test]
        fn test_seek_position_at_end() {
            let duration: i64 = 100000;
            let time: i64 = 100000;

            let position = time as f32 / duration as f32;
            assert_eq!(position, 1.0);
        }

        #[test]
        fn test_seek_position_at_midpoint() {
            let duration: i64 = 100000;
            let time: i64 = 50000;

            let position = time as f32 / duration as f32;
            assert_eq!(position, 0.5);
        }
    }

    /// Tests for delay conversion logic
    mod delay_conversion_tests {
        #[test]
        fn test_milliseconds_to_microseconds() {
            let delay_ms: i64 = 500;
            let delay_us = delay_ms * 1000;
            assert_eq!(delay_us, 500000);
        }

        #[test]
        fn test_negative_delay_conversion() {
            let delay_ms: i64 = -200;
            let delay_us = delay_ms * 1000;
            assert_eq!(delay_us, -200000);
        }

        #[test]
        fn test_zero_delay_conversion() {
            let delay_ms: i64 = 0;
            let delay_us = delay_ms * 1000;
            assert_eq!(delay_us, 0);
        }

        #[test]
        fn test_large_delay_conversion() {
            let delay_ms: i64 = 5000;
            let delay_us = delay_ms * 1000;
            assert_eq!(delay_us, 5_000_000);
        }
    }
}
