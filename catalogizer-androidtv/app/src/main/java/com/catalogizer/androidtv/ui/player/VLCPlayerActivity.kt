package com.catalogizer.androidtv.ui.player

import android.os.Bundle
import android.util.Log
import android.view.KeyEvent
import android.view.View
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.lifecycleScope
import androidx.lifecycle.repeatOnLifecycle
import androidx.tv.material3.*
import com.catalogizer.androidtv.player.VLCPlayer
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.flow.first
import kotlinx.serialization.json.jsonPrimitive
import org.videolan.libvlc.util.VLCVideoLayout

// Progress tracking constants
private const val PROGRESS_SAVE_INTERVAL_MS = 5000L // Save every 5 seconds
private const val MIN_PROGRESS_TO_SAVE = 0.05 // Only save if > 5% watched
private const val MAX_PROGRESS_TO_SAVE = 0.95 // Don't save if > 95% (completed)

/**
 * VLC-based video player activity for Android TV.
 * Provides full-screen video playback with TV-optimized controls.
 */
class VLCPlayerActivity : ComponentActivity() {

    companion object {
        const val EXTRA_MEDIA_ID = "media_id"
        const val EXTRA_STREAM_URL = "stream_url"
        const val EXTRA_MEDIA_TITLE = "media_title"
        private const val TAG = "VLCPlayerActivity"
        private const val CONTROLS_AUTO_HIDE_DELAY = 5000L // 5 seconds
    }

    private lateinit var vlcPlayer: VLCPlayer
    private var mediaId: Long = -1
    private var streamUrl: String? = null
    private var mediaTitle: String = ""
    private var lastSavedProgress: Double = 0.0
    private var progressTrackingJob: kotlinx.coroutines.Job? = null
    private var resumePositionMs: Long = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Extract extras
        mediaId = intent.getLongExtra(EXTRA_MEDIA_ID, -1)
        streamUrl = intent.getStringExtra(EXTRA_STREAM_URL)
        mediaTitle = intent.getStringExtra(EXTRA_MEDIA_TITLE) ?: ""

        // Initialize VLC
        vlcPlayer = VLCPlayer(this)

        setContent {
            VLCPlayerScreen(
                vlcPlayer = vlcPlayer,
                mediaTitle = mediaTitle,
                onBackPressed = { finish() }
            )
        }

        // Resolve stream URL if media_id provided
        if (streamUrl.isNullOrBlank() && mediaId > 0) {
            lifecycleScope.launch {
                resolveAndPlay(mediaId)
            }
        } else if (!streamUrl.isNullOrBlank()) {
            // Start playback with provided URL
            vlcPlayer.play(streamUrl!!)
            Log.d(TAG, "Starting playback: $streamUrl")
        } else {
            Toast.makeText(this, "No media to play", Toast.LENGTH_SHORT).show()
            finish()
        }
    }
    
    private suspend fun resolveAndPlay(mediaId: Long) {
        try {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(this)
            val baseUrl = container.getServerUrl().trimEnd('/')
            
            // Fetch media details including watch progress
            try {
                val mediaFlow = container.mediaRepository.getMediaById(mediaId)
                val item = mediaFlow.first()
                if (item != null) {
                    if (item.title.isNotBlank()) {
                        mediaTitle = item.title
                    }
                    // Get resume position from watch progress
                    val durationMs = item.duration ?: 0
                    if (item.watchProgress > 0 && item.watchProgress < 0.95 && durationMs > 0) {
                        resumePositionMs = (item.watchProgress * durationMs).toLong()
                        Log.d(TAG, "Will resume from ${resumePositionMs}ms (progress: ${item.watchProgress})")
                    }
                }
            } catch (e: Exception) {
                Log.w(TAG, "Could not fetch media details: ${e.message}")
            }
            
            // Fetch stream URL from API
            val response = container.api.getEntityStream(mediaId)
            if (response.isSuccessful) {
                val body = response.body()
                val streamPath = body?.get("stream_url")?.jsonPrimitive?.content
                if (streamPath != null) {
                    val resolvedUrl = if (streamPath.startsWith("/")) "$baseUrl$streamPath" else streamPath
                    streamUrl = resolvedUrl
                    
                    // Start playback
                    vlcPlayer.play(resolvedUrl)
                    
                    // Seek to resume position if available
                    if (resumePositionMs > 0) {
                        delay(1000) // Wait for player to be ready
                        vlcPlayer.seek(resumePositionMs.toFloat() / 1000 / 60) // Approximate seek
                        Log.d(TAG, "Resumed playback at ${resumePositionMs}ms")
                    }
                    
                    Log.d(TAG, "Starting playback for media $mediaId: $resolvedUrl")
                    
                    // Start progress tracking
                    startProgressTracking()
                } else {
                    showErrorAndFinish("No stream URL available")
                }
            } else {
                showErrorAndFinish("Stream unavailable (${response.code()})")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to resolve stream: ${e.message}", e)
            showErrorAndFinish("Failed to load stream: ${e.message}")
        }
    }
    
    private fun startProgressTracking() {
        progressTrackingJob = lifecycleScope.launch {
            while (true) {
                delay(PROGRESS_SAVE_INTERVAL_MS)
                saveWatchProgress()
            }
        }
    }
    
    private suspend fun saveWatchProgress() {
        if (mediaId <= 0) return
        
        val currentPosition = vlcPlayer.currentPosition.value
        val duration = vlcPlayer.duration.value
        
        if (duration <= 0) return
        
        val progress = currentPosition.toDouble() / duration.toDouble()
        
        // Only save if progress changed significantly and is in valid range
        if (progress in MIN_PROGRESS_TO_SAVE..MAX_PROGRESS_TO_SAVE && 
            kotlin.math.abs(progress - lastSavedProgress) > 0.01) {
            
            try {
                val container = com.catalogizer.androidtv.DependencyContainer.getInstance(this)
                container.mediaRepository.updateWatchProgress(mediaId, progress)
                lastSavedProgress = progress
                Log.d(TAG, "Saved watch progress: ${"%.2f".format(progress * 100)}%")
            } catch (e: Exception) {
                Log.w(TAG, "Failed to save watch progress: ${e.message}")
            }
        }
    }
    
    private fun showErrorAndFinish(message: String) {
        Toast.makeText(this, message, Toast.LENGTH_LONG).show()
        finish()
    }

    override fun onPause() {
        super.onPause()
        vlcPlayer.pause()
        // Save progress when pausing
        lifecycleScope.launch {
            saveWatchProgress()
        }
    }

    override fun onResume() {
        super.onResume()
        vlcPlayer.resume()
    }

    override fun onDestroy() {
        super.onDestroy()
        // Save final progress before destroying
        progressTrackingJob?.cancel()
        lifecycleScope.launch {
            saveWatchProgress()
            vlcPlayer.release()
        }
    }

    override fun onKeyDown(keyCode: Int, event: KeyEvent?): Boolean {
        // Handle TV remote buttons
        return when (keyCode) {
            KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE -> {
                togglePlayPause()
                true
            }
            KeyEvent.KEYCODE_MEDIA_PLAY -> {
                vlcPlayer.resume()
                true
            }
            KeyEvent.KEYCODE_MEDIA_PAUSE -> {
                vlcPlayer.pause()
                true
            }
            KeyEvent.KEYCODE_MEDIA_STOP -> {
                vlcPlayer.stop()
                finish()
                true
            }
            KeyEvent.KEYCODE_MEDIA_FAST_FORWARD -> {
                vlcPlayer.seekForward(30000) // 30 seconds
                true
            }
            KeyEvent.KEYCODE_MEDIA_REWIND -> {
                vlcPlayer.seekBackward(30000) // 30 seconds
                true
            }
            KeyEvent.KEYCODE_MEDIA_NEXT -> {
                // Play next episode in series if available
                lifecycleScope.launch {
                    playNextEpisode()
                }
                true
            }
            KeyEvent.KEYCODE_MEDIA_PREVIOUS -> {
                // Play previous episode in series if available
                lifecycleScope.launch {
                    playPreviousEpisode()
                }
                true
            }
            KeyEvent.KEYCODE_DPAD_CENTER -> {
                // Show/hide controls
                true
            }
            KeyEvent.KEYCODE_BACK -> {
                finish()
                true
            }
            else -> super.onKeyDown(keyCode, event)
        }
    }

    /**
     * Play the next episode in the series if available.
     * Queries the API for episodes with the next episode number.
     */
    private suspend fun playNextEpisode() {
        if (mediaId <= 0) {
            Toast.makeText(this, "Cannot determine next episode", Toast.LENGTH_SHORT).show()
            return
        }
        
        try {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(this)
            val response = container.apiService.getEntityDetails(mediaId)
            
            if (response.isSuccessful) {
                val currentEntity = response.body()
                val seriesId = currentEntity?.series_id
                val currentEpisodeNum = currentEntity?.episode_number ?: 0
                
                if (seriesId != null && seriesId > 0) {
                    // Query for next episode
                    val nextEpisodeResponse = container.apiService.getEpisodesBySeries(
                        seriesId = seriesId,
                        seasonNumber = currentEntity?.season_number ?: 1
                    )
                    
                    if (nextEpisodeResponse.isSuccessful) {
                        val episodes = nextEpisodeResponse.body() ?: emptyList()
                        val nextEpisode = episodes.find { it.episode_number == currentEpisodeNum + 1 }
                        
                        if (nextEpisode != null) {
                            // Save progress for current episode before switching
                            saveWatchProgress()
                            
                            // Play next episode
                            mediaId = nextEpisode.id
                            mediaTitle = nextEpisode.title ?: ""
                            resolveAndPlay(mediaId)
                            Toast.makeText(this, "Playing next episode", Toast.LENGTH_SHORT).show()
                        } else {
                            Toast.makeText(this, "No next episode available", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(this, "Failed to fetch episodes", Toast.LENGTH_SHORT).show()
                    }
                } else {
                    Toast.makeText(this, "No series information available", Toast.LENGTH_SHORT).show()
                }
            } else {
                Toast.makeText(this, "Failed to get current media details", Toast.LENGTH_SHORT).show()
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error playing next episode", e)
            Toast.makeText(this, "Error: ${e.message}", Toast.LENGTH_SHORT).show()
        }
    }
    
    /**
     * Play the previous episode in the series if available.
     * Queries the API for episodes with the previous episode number.
     */
    private suspend fun playPreviousEpisode() {
        if (mediaId <= 0) {
            Toast.makeText(this, "Cannot determine previous episode", Toast.LENGTH_SHORT).show()
            return
        }
        
        try {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(this)
            val response = container.apiService.getEntityDetails(mediaId)
            
            if (response.isSuccessful) {
                val currentEntity = response.body()
                val seriesId = currentEntity?.series_id
                val currentEpisodeNum = currentEntity?.episode_number ?: 0
                
                if (seriesId != null && seriesId > 0 && currentEpisodeNum > 1) {
                    // Query for previous episode
                    val prevEpisodeResponse = container.apiService.getEpisodesBySeries(
                        seriesId = seriesId,
                        seasonNumber = currentEntity?.season_number ?: 1
                    )
                    
                    if (prevEpisodeResponse.isSuccessful) {
                        val episodes = prevEpisodeResponse.body() ?: emptyList()
                        val prevEpisode = episodes.find { it.episode_number == currentEpisodeNum - 1 }
                        
                        if (prevEpisode != null) {
                            // Save progress for current episode before switching
                            saveWatchProgress()
                            
                            // Play previous episode
                            mediaId = prevEpisode.id
                            mediaTitle = prevEpisode.title ?: ""
                            resolveAndPlay(mediaId)
                            Toast.makeText(this, "Playing previous episode", Toast.LENGTH_SHORT).show()
                        } else {
                            Toast.makeText(this, "No previous episode available", Toast.LENGTH_SHORT).show()
                        }
                    } else {
                        Toast.makeText(this, "Failed to fetch episodes", Toast.LENGTH_SHORT).show()
                    }
                } else {
                    Toast.makeText(this, "No previous episode available", Toast.LENGTH_SHORT).show()
                }
            } else {
                Toast.makeText(this, "Failed to get current media details", Toast.LENGTH_SHORT).show()
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error playing previous episode", e)
            Toast.makeText(this, "Error: ${e.message}", Toast.LENGTH_SHORT).show()
        }
    }
    
    /**
     * Save current watch progress to the API.
     */
    private suspend fun saveWatchProgress() {
        if (mediaId <= 0) return
        
        try {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(this)
            val currentTime = vlcPlayer.getCurrentTime()
            val duration = vlcPlayer.getDuration()
            
            if (duration > 0) {
                val progress = currentTime.toFloat() / duration.toFloat()
                if (progress in MIN_PROGRESS_TO_SAVE..MAX_PROGRESS_TO_SAVE) {
                    container.mediaRepository.updateWatchProgress(mediaId, progress.toDouble())
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error saving watch progress", e)
        }
    }

    private fun togglePlayPause() {
        if (vlcPlayer.isPlaying.value) {
            vlcPlayer.pause()
        } else {
            vlcPlayer.resume()
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun VLCPlayerScreen(
    vlcPlayer: VLCPlayer,
    mediaTitle: String,
    onBackPressed: () -> Unit
) {
    val context = LocalContext.current
    
    // Collect VLC state
    val isPlaying by vlcPlayer.isPlaying.collectAsState()
    val currentPosition by vlcPlayer.currentPosition.collectAsState()
    val duration by vlcPlayer.duration.collectAsState()
    val playbackState by vlcPlayer.playbackState.collectAsState()
    val volume by vlcPlayer.volume.collectAsState()
    val audioTracks by vlcPlayer.audioTracks.collectAsState()
    val subtitleTracks by vlcPlayer.subtitleTracks.collectAsState()
    val currentAudioTrack by vlcPlayer.currentAudioTrack.collectAsState()
    val currentSubtitleTrack by vlcPlayer.currentSubtitleTrack.collectAsState()

    // UI State
    var showControls by remember { mutableStateOf(true) }
    var showAudioTrackMenu by remember { mutableStateOf(false) }
    var showSubtitleMenu by remember { mutableStateOf(false) }
    var showSettingsMenu by remember { mutableStateOf(false) }
    
    // Auto-hide controls
    LaunchedEffect(showControls) {
        if (showControls) {
            delay(5000) // 5 seconds
            showControls = false
        }
    }

    // Handle state changes
    LaunchedEffect(playbackState) {
        when (playbackState) {
            VLCPlayer.PlaybackState.ERROR -> {
                Toast.makeText(context, "Playback error occurred", Toast.LENGTH_LONG).show()
            }
            VLCPlayer.PlaybackState.COMPLETED -> {
                onBackPressed()
            }
            else -> {}
        }
    }

    Box(
        modifier = Modifier.fillMaxSize()
    ) {
        // Video Surface
        AndroidView(
            factory = { ctx ->
                VLCVideoLayout(ctx).apply {
                    vlcPlayer.attachView(this)
                }
            },
            modifier = Modifier.fillMaxSize()
        )

        // Loading Indicator
        if (playbackState == VLCPlayer.PlaybackState.BUFFERING) {
            CircularProgressIndicator(
                modifier = Modifier.align(Alignment.Center),
                color = Color.White
            )
        }

        // Controls Overlay
        if (showControls) {
            PlayerControlsOverlay(
                mediaTitle = mediaTitle,
                isPlaying = isPlaying,
                currentPosition = currentPosition,
                duration = duration,
                volume = volume,
                onPlayPauseClick = {
                    if (isPlaying) vlcPlayer.pause() else vlcPlayer.resume()
                },
                onSeek = { position -> vlcPlayer.seek(position) },
                onVolumeChange = { vlcPlayer.setVolume(it) },
                onAudioTrackClick = { showAudioTrackMenu = true },
                onSubtitleTrackClick = { showSubtitleMenu = true },
                onSettingsClick = { showSettingsMenu = true },
                onBackClick = onBackPressed,
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
                    .background(Color.Black.copy(alpha = 0.7f))
                    .padding(16.dp)
            )
        }

        // Audio Track Selection Menu
        if (showAudioTrackMenu) {
            TrackSelectionMenu(
                title = "Select Audio Track",
                tracks = audioTracks,
                currentTrack = currentAudioTrack,
                onTrackSelected = { vlcPlayer.setAudioTrack(it.id) },
                onDismiss = { showAudioTrackMenu = false }
            )
        }

        // Subtitle Track Selection Menu
        if (showSubtitleMenu) {
            TrackSelectionMenu(
                title = "Select Subtitle Track",
                tracks = subtitleTracks,
                currentTrack = currentSubtitleTrack,
                onTrackSelected = { vlcPlayer.setSubtitleTrack(it.id) },
                onDismiss = { showSubtitleMenu = false }
            )
        }

        // Settings Menu
        if (showSettingsMenu) {
            PlayerSettingsMenu(
                vlcPlayer = vlcPlayer,
                onDismiss = { showSettingsMenu = false }
            )
        }

        // Tap to show controls
        if (!showControls) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .focusable()
                    .onFocusChanged { 
                        if (it.isFocused || it.hasFocus) {
                            showControls = true
                        }
                    }
            )
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun PlayerControlsOverlay(
    mediaTitle: String,
    isPlaying: Boolean,
    currentPosition: Long,
    duration: Long,
    volume: Int,
    onPlayPauseClick: () -> Unit,
    onSeek: (Float) -> Unit,
    onVolumeChange: (Int) -> Unit,
    onAudioTrackClick: () -> Unit,
    onSubtitleTrackClick: () -> Unit,
    onSettingsClick: () -> Unit,
    onBackClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        // Title
        Text(
            text = mediaTitle,
            style = MaterialTheme.typography.headlineSmall,
            color = Color.White,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        // Progress Bar
        if (duration > 0) {
            val progress = currentPosition.toFloat() / duration.toFloat()
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = formatTime(currentPosition),
                    color = Color.White,
                    modifier = Modifier.width(60.dp)
                )
                
                Slider(
                    value = progress,
                    onValueChange = onSeek,
                    modifier = Modifier.weight(1f).padding(horizontal = 8.dp)
                )
                
                Text(
                    text = formatTime(duration),
                    color = Color.White,
                    modifier = Modifier.width(60.dp)
                )
            }
        }

        // Control Buttons
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(top = 16.dp),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Back button
            IconButton(onClick = onBackClick) {
                Text("←", color = Color.White, style = MaterialTheme.typography.titleLarge)
            }

            // Rewind
            IconButton(onClick = { /* Seek backward */ }) {
                Text("⏪", color = Color.White, style = MaterialTheme.typography.titleLarge)
            }

            // Play/Pause
            IconButton(
                onClick = onPlayPauseClick,
                modifier = Modifier.size(64.dp)
            ) {
                Icon(
                    imageVector = if (isPlaying) Icons.Default.Pause else Icons.Default.PlayArrow,
                    contentDescription = if (isPlaying) "Pause" else "Play",
                    tint = Color.White,
                    modifier = Modifier.size(48.dp)
                )
            }

            // Fast Forward
            IconButton(onClick = { /* Seek forward */ }) {
                Text("⏩", color = Color.White, style = MaterialTheme.typography.titleLarge)
            }

            // Audio Track
            IconButton(onClick = onAudioTrackClick) {
                Text("🎵", style = MaterialTheme.typography.titleLarge)
            }

            // Subtitle Track
            IconButton(onClick = onSubtitleTrackClick) {
                Text("💬", style = MaterialTheme.typography.titleLarge)
            }

            // Settings
            IconButton(onClick = onSettingsClick) {
                Text("⚙️", style = MaterialTheme.typography.titleLarge)
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TrackSelectionMenu(
    title: String,
    tracks: List<VLCPlayer.Track>,
    currentTrack: VLCPlayer.Track?,
    onTrackSelected: (VLCPlayer.Track) -> Unit,
    onDismiss: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.7f))
            .clickable(onClick = onDismiss),
        contentAlignment = Alignment.Center
    ) {
        Surface(
            modifier = Modifier
                .widthIn(max = 400.dp)
                .padding(16.dp),
            shape = SurfaceDefaults.shape,
            colors = SurfaceDefaults.colors(
                containerColor = Color(0xFF1E1E1E)
            )
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleLarge,
                    color = Color(0xFFFF6600),
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.padding(bottom = 16.dp)
                )
                
                tracks.forEach { track ->
                    val isSelected = track.id == currentTrack?.id
                    
                    Surface(
                        onClick = {
                            onTrackSelected(track)
                            onDismiss()
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 4.dp),
                        colors = ClickableSurfaceDefaults.colors(
                            containerColor = if (isSelected) Color(0xFFFF6600).copy(alpha = 0.2f) else Color.Transparent
                        )
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(12.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = track.name,
                                color = if (isSelected) Color(0xFFFF6600) else Color.White
                            )
                            if (isSelected) {
                                Text(
                                    text = "✓",
                                    color = Color(0xFFFF6600)
                                )
                            }
                        }
                    }
                }
                
                Spacer(modifier = Modifier.height(16.dp))
                
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("Close")
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun PlayerSettingsMenu(
    vlcPlayer: VLCPlayer,
    onDismiss: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.7f))
            .clickable(onClick = onDismiss),
        contentAlignment = Alignment.Center
    ) {
        Surface(
            modifier = Modifier
                .widthIn(max = 400.dp)
                .padding(16.dp),
            shape = SurfaceDefaults.shape,
            colors = SurfaceDefaults.colors(
                containerColor = Color(0xFF1E1E1E)
            )
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Text(
                    text = "Player Settings",
                    style = MaterialTheme.typography.titleLarge,
                    color = Color(0xFFFF6600),
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.padding(bottom = 16.dp)
                )
                
                // Playback Speed
                Text(
                    text = "Playback Speed",
                    style = MaterialTheme.typography.titleMedium,
                    color = Color.White,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
                Row(
                    modifier = Modifier.padding(bottom = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    listOf(0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f).forEach { speed ->
                        FilterChip(
                            selected = false,
                            onClick = { vlcPlayer.setSpeed(speed) }
                        ) {
                            Text("${speed}x")
                        }
                    }
                }
                
                // Aspect Ratio
                Text(
                    text = "Aspect Ratio",
                    style = MaterialTheme.typography.titleMedium,
                    color = Color.White,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
                Row(
                    modifier = Modifier.padding(bottom = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    listOf("16:9", "4:3", "1:1", "2.21:1").forEach { ratio ->
                        FilterChip(
                            selected = false,
                            onClick = { vlcPlayer.setAspectRatio(ratio) }
                        ) {
                            Text(ratio)
                        }
                    }
                }
                
                Spacer(modifier = Modifier.height(16.dp))
                
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("Close")
                }
            }
        }
    }
}

private fun formatTime(milliseconds: Long): String {
    val seconds = milliseconds / 1000
    val minutes = seconds / 60
    val hours = minutes / 60
    val remainingMinutes = minutes % 60
    val remainingSeconds = seconds % 60
    
    return if (hours > 0) {
        String.format("%d:%02d:%02d", hours, remainingMinutes, remainingSeconds)
    } else {
        String.format("%02d:%02d", remainingMinutes, remainingSeconds)
    }
}
