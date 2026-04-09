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
