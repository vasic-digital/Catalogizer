package com.catalogizer.androidtv.player

import android.content.Context
import android.net.Uri
import android.util.Log
import android.view.SurfaceView
import androidx.annotation.OptIn
import androidx.media3.common.util.UnstableApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.Media
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.util.VLCVideoLayout

/**
 * VLC Player wrapper for Android TV.
 * Provides a clean interface to VLC's media playback capabilities with
 * support for subtitles, audio tracks, video tracks, and playback controls.
 */
class VLCPlayer(private val context: Context) {
    companion object {
        private const val TAG = "VLCPlayer"
        
        // VLC options for optimized TV playback
        private val VLC_OPTIONS = listOf(
            "--rtsp-tcp",           // Use TCP for RTSP (more reliable)
            "--network-caching=1500", // Network buffer (ms)
            "--file-caching=300",     // File buffer (ms)
            "--live-caching=1500",    // Live stream buffer (ms)
            "--sout-mux-caching=1500", // Stream output buffer (ms)
            "--no-drop-late-frames",  // Don't drop late frames
            "--no-skip-frames",       // Don't skip frames
            "--audio-time-stretch",   // Enable audio time stretching
            "--avcodec-hw",           // Enable hardware decoding
            "-- subsdec-encoding=UTF-8" // Subtitle encoding
        )
    }

    private var libVLC: LibVLC? = null
    private var mediaPlayer: MediaPlayer? = null
    private var videoLayout: VLCVideoLayout? = null
    
    // Playback state
    private val _playbackState = MutableStateFlow(PlaybackState.IDLE)
    val playbackState: StateFlow<PlaybackState> = _playbackState.asStateFlow()
    
    private val _currentPosition = MutableStateFlow(0L)
    val currentPosition: StateFlow<Long> = _currentPosition.asStateFlow()
    
    private val _duration = MutableStateFlow(0L)
    val duration: StateFlow<Long> = _duration.asStateFlow()
    
    private val _isPlaying = MutableStateFlow(false)
    val isPlaying: StateFlow<Boolean> = _isPlaying.asStateFlow()
    
    private val _volume = MutableStateFlow(100)
    val volume: StateFlow<Int> = _volume.asStateFlow()
    
    // Tracks
    private val _audioTracks = MutableStateFlow<List<Track>>(emptyList())
    val audioTracks: StateFlow<List<Track>> = _audioTracks.asStateFlow()
    
    private val _subtitleTracks = MutableStateFlow<List<Track>>(emptyList())
    val subtitleTracks: StateFlow<List<Track>> = _subtitleTracks.asStateFlow()
    
    private val _videoTracks = MutableStateFlow<List<Track>>(emptyList())
    val videoTracks: StateFlow<List<Track>> = _videoTracks.asStateFlow()
    
    private val _currentAudioTrack = MutableStateFlow<Track?>(null)
    val currentAudioTrack: StateFlow<Track?> = _currentAudioTrack.asStateFlow()
    
    private val _currentSubtitleTrack = MutableStateFlow<Track?>(null)
    val currentSubtitleTrack: StateFlow<Track?> = _currentSubtitleTrack.asStateFlow()

    private var eventListener: MediaPlayer.EventListener? = null

    init {
        initialize()
    }

    /**
     * Initialize VLC with TV-optimized options
     */
    private fun initialize() {
        try {
            libVLC = LibVLC(context, VLC_OPTIONS)
            mediaPlayer = MediaPlayer(libVLC)
            setupEventListener()
            Log.d(TAG, "VLC initialized successfully")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize VLC: ${e.message}", e)
            _playbackState.value = PlaybackState.ERROR
        }
    }

    /**
     * Set up VLC event listener for state updates
     */
    private fun setupEventListener() {
        eventListener = MediaPlayer.EventListener { event ->
            when (event.type) {
                MediaPlayer.Event.Opening -> {
                    _playbackState.value = PlaybackState.BUFFERING
                    Log.d(TAG, "Media opening...")
                }
                MediaPlayer.Event.Playing -> {
                    _playbackState.value = PlaybackState.PLAYING
                    _isPlaying.value = true
                    Log.d(TAG, "Playback started")
                }
                MediaPlayer.Event.Paused -> {
                    _playbackState.value = PlaybackState.PAUSED
                    _isPlaying.value = false
                    Log.d(TAG, "Playback paused")
                }
                MediaPlayer.Event.Stopped -> {
                    _playbackState.value = PlaybackState.STOPPED
                    _isPlaying.value = false
                    Log.d(TAG, "Playback stopped")
                }
                MediaPlayer.Event.EndReached -> {
                    _playbackState.value = PlaybackState.COMPLETED
                    _isPlaying.value = false
                    Log.d(TAG, "Playback completed")
                }
                MediaPlayer.Event.EncounteredError -> {
                    _playbackState.value = PlaybackState.ERROR
                    _isPlaying.value = false
                    Log.e(TAG, "Playback error occurred")
                }
                MediaPlayer.Event.TimeChanged -> {
                    _currentPosition.value = event.timeChanged
                }
                MediaPlayer.Event.LengthChanged -> {
                    _duration.value = event.lengthChanged
                }
                MediaPlayer.Event.Vout -> {
                    // Video output started - update tracks
                    updateTracks()
                }
            }
        }
        mediaPlayer?.setEventListener(eventListener)
    }

    /**
     * Attach video layout for display
     */
    fun attachView(videoLayout: VLCVideoLayout) {
        this.videoLayout = videoLayout
        mediaPlayer?.attachViews(videoLayout, null, false, false)
    }

    /**
     * Load and play media from URI
     */
    fun play(uri: String) {
        try {
            val media = Media(libVLC, Uri.parse(uri))
            
            // Enable all tracks
            media.setBooleanOption(Media.Option.Trusted, true)
            media.setBooleanOption(Media.Option.TextRenderer, true)
            
            // Hardware acceleration
            media.setHWDecoderEnabled(true, false)
            
            mediaPlayer?.media = media
            mediaPlayer?.play()
            
            media.release()
            
            _playbackState.value = PlaybackState.BUFFERING
            Log.d(TAG, "Playing: $uri")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to play media: ${e.message}", e)
            _playbackState.value = PlaybackState.ERROR
        }
    }

    /**
     * Pause playback
     */
    fun pause() {
        mediaPlayer?.pause()
        Log.d(TAG, "Pause requested")
    }

    /**
     * Resume playback
     */
    fun resume() {
        mediaPlayer?.play()
        Log.d(TAG, "Resume requested")
    }

    /**
     * Stop playback
     */
    fun stop() {
        mediaPlayer?.stop()
        _playbackState.value = PlaybackState.STOPPED
        _isPlaying.value = false
        Log.d(TAG, "Stop requested")
    }

    /**
     * Seek to position (0.0 to 1.0)
     */
    fun seek(position: Float) {
        val targetPosition = (position * (_duration.value)).toLong()
        mediaPlayer?.time = targetPosition
        Log.d(TAG, "Seek to: $targetPosition ms")
    }

    /**
     * Seek forward by milliseconds
     */
    fun seekForward(milliseconds: Long = 10000) {
        val newTime = (_currentPosition.value + milliseconds).coerceAtMost(_duration.value)
        mediaPlayer?.time = newTime
        Log.d(TAG, "Seek forward to: $newTime ms")
    }

    /**
     * Seek backward by milliseconds
     */
    fun seekBackward(milliseconds: Long = 10000) {
        val newTime = (_currentPosition.value - milliseconds).coerceAtLeast(0)
        mediaPlayer?.time = newTime
        Log.d(TAG, "Seek backward to: $newTime ms")
    }

    /**
     * Set playback speed (0.25 to 4.0)
     */
    fun setSpeed(speed: Float) {
        val clampedSpeed = speed.coerceIn(0.25f, 4.0f)
        mediaPlayer?.rate = clampedSpeed
        Log.d(TAG, "Speed set to: $clampedSpeed")
    }

    /**
     * Set volume (0 to 100)
     */
    fun setVolume(volume: Int) {
        val clampedVolume = volume.coerceIn(0, 100)
        mediaPlayer?.volume = clampedVolume
        _volume.value = clampedVolume
        Log.d(TAG, "Volume set to: $clampedVolume")
    }

    /**
     * Set audio track by ID
     */
    fun setAudioTrack(trackId: Int) {
        mediaPlayer?.audioTrack = trackId
        updateCurrentTracks()
        Log.d(TAG, "Audio track set to: $trackId")
    }

    /**
     * Set subtitle track by ID (-1 to disable)
     */
    fun setSubtitleTrack(trackId: Int) {
        mediaPlayer?.spuTrack = trackId
        updateCurrentTracks()
        Log.d(TAG, "Subtitle track set to: $trackId")
    }

    /**
     * Set subtitle delay in milliseconds
     */
    fun setSubtitleDelay(delayMs: Long) {
        mediaPlayer?.setSpuDelay(delayMs * 1000) // VLC uses microseconds
        Log.d(TAG, "Subtitle delay set to: $delayMs ms")
    }

    /**
     * Set audio delay in milliseconds
     */
    fun setAudioDelay(delayMs: Long) {
        mediaPlayer?.setAudioDelay(delayMs * 1000) // VLC uses microseconds
        Log.d(TAG, "Audio delay set to: $delayMs ms")
    }

    /**
     * Set crop/aspect ratio
     */
    fun setAspectRatio(aspectRatio: String) {
        mediaPlayer?.aspectRatio = aspectRatio
        Log.d(TAG, "Aspect ratio set to: $aspectRatio")
    }

    /**
     * Set video scale
     */
    fun setVideoScale(scale: Float) {
        mediaPlayer?.scale = scale
        Log.d(TAG, "Video scale set to: $scale")
    }

    /**
     * Get available tracks
     */
    private fun updateTracks() {
        val player = mediaPlayer ?: return
        
        // Audio tracks
        val audioTracksList = mutableListOf<Track>()
        val audioTrackIds = player.audioTracks
        val audioTrackNames = player.audioTrackNames
        
        audioTrackIds?.forEachIndexed { index, id ->
            audioTracksList.add(
                Track(
                    id = id,
                    name = audioTrackNames?.getOrNull(index) ?: "Audio Track ${index + 1}",
                    type = TrackType.AUDIO
                )
            )
        }
        _audioTracks.value = audioTracksList
        
        // Subtitle tracks
        val subtitleTracksList = mutableListOf<Track>()
        val spuTrackIds = player.spuTracks
        val spuTrackNames = player.spuTrackNames
        
        spuTrackIds?.forEachIndexed { index, id ->
            subtitleTracksList.add(
                Track(
                    id = id,
                    name = spuTrackNames?.getOrNull(index) ?: "Subtitle ${index + 1}",
                    type = TrackType.SUBTITLE
                )
            )
        }
        _subtitleTracks.value = subtitleTracksList
        
        // Video tracks (usually only one)
        val videoTracksList = mutableListOf<Track>()
        player.currentVideoTrack?.let { track ->
            videoTracksList.add(
                Track(
                    id = track.id,
                    name = "${track.width}x${track.height}",
                    type = TrackType.VIDEO
                )
            )
        }
        _videoTracks.value = videoTracksList
        
        updateCurrentTracks()
    }

    private fun updateCurrentTracks() {
        val player = mediaPlayer ?: return
        
        _currentAudioTrack.value = _audioTracks.value.find { it.id == player.audioTrack }
        _currentSubtitleTrack.value = _subtitleTracks.value.find { it.id == player.spuTrack }
    }

    /**
     * Take a screenshot of current video frame
     */
    fun takeScreenshot(): ByteArray? {
        return try {
            val videoLayout = this.videoLayout ?: return null
            // Implementation depends on VLC's screenshot capabilities
            // This is a placeholder - VLC doesn't directly support screenshot from video layout
            null
        } catch (e: Exception) {
            Log.e(TAG, "Screenshot failed: ${e.message}")
            null
        }
    }

    /**
     * Release resources
     */
    fun release() {
        Log.d(TAG, "Releasing VLC resources")
        
        mediaPlayer?.stop()
        mediaPlayer?.detachViews()
        mediaPlayer?.setEventListener(null)
        mediaPlayer?.release()
        mediaPlayer = null
        
        libVLC?.release()
        libVLC = null
        
        _playbackState.value = PlaybackState.IDLE
        _isPlaying.value = false
    }

    /**
     * Playback states
     */
    enum class PlaybackState {
        IDLE,
        BUFFERING,
        PLAYING,
        PAUSED,
        STOPPED,
        COMPLETED,
        ERROR
    }

    /**
     * Track types
     */
    enum class TrackType {
        AUDIO,
        VIDEO,
        SUBTITLE
    }

    /**
     * Track data class
     */
    data class Track(
        val id: Int,
        val name: String,
        val type: TrackType
    )
}
