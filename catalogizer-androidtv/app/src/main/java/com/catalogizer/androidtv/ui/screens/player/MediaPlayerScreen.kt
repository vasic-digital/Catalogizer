@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.screens.player

import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
// material3-only composables (no TV equivalent)
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.runtime.*
import kotlinx.serialization.json.jsonPrimitive
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.input.key.KeyEventType
import androidx.compose.ui.input.key.key
import androidx.compose.ui.input.key.onKeyEvent
import androidx.compose.ui.input.key.type
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.media3.common.C
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.common.TrackSelectionOverride
import androidx.media3.common.Tracks
import androidx.media3.common.util.UnstableApi
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.PlayerView
// TV material3 for everything else
import androidx.tv.material3.*
import kotlinx.coroutines.flow.first

// VLC-style accent color
private val VlcOrange = Color(0xFFFF6600)
private val OverlayDark = Color.Black.copy(alpha = 0.75f)
private val OverlayGradientTop = listOf(Color.Black.copy(alpha = 0.85f), Color.Transparent)
private val OverlayGradientBottom = listOf(Color.Transparent, Color.Black.copy(alpha = 0.85f))

// Playback speed presets (VLC-style)
private val speedOptions = listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f, 3.0f)

/**
 * VLC-style media player screen with ExoPlayer integration, D-pad controls,
 * seekbar, playback speed selection, audio/subtitle track switching, and
 * auto-hiding overlay. Resolves stream URLs from the entity API with auth headers.
 */
@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun MediaPlayerScreen(
    mediaId: Long,
    mediaUrl: String = "",
    mediaTitle: String = "Media $mediaId",
    onNavigateBack: () -> Unit
) {
    val context = androidx.compose.ui.platform.LocalContext.current
    var exoPlayer by remember { mutableStateOf<ExoPlayer?>(null) }
    var isPlaying by remember { mutableStateOf(false) }
    var currentPosition by remember { mutableStateOf(0L) }
    var duration by remember { mutableStateOf(0L) }
    var resolvedUrl by remember { mutableStateOf(mediaUrl) }
    var resolvedTitle by remember { mutableStateOf(mediaTitle) }
    var streamError by remember { mutableStateOf<String?>(null) }
    var retryCount by remember { mutableStateOf(0) }

    // VLC-like controls state
    var playbackSpeed by remember { mutableStateOf(1.0f) }
    var showControls by remember { mutableStateOf(true) }
    var showSpeedMenu by remember { mutableStateOf(false) }
    var showAudioMenu by remember { mutableStateOf(false) }
    var showSubtitleMenu by remember { mutableStateOf(false) }
    var audioTracks by remember { mutableStateOf<List<TrackInfo>>(emptyList()) }
    var subtitleTracks by remember { mutableStateOf<List<TrackInfo>>(emptyList()) }
    var selectedAudioIndex by remember { mutableStateOf(0) }
    var selectedSubtitleIndex by remember { mutableStateOf(-1) } // -1 = off
    var bufferedPosition by remember { mutableStateOf(0L) }

    val controlsFocus = remember { FocusRequester() }

    // Fetch stream URL and real title from entity endpoint
    LaunchedEffect(mediaId, retryCount) {
        if (resolvedUrl.isEmpty()) {
            try {
                val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
                val baseUrl = container.getServerUrl().trimEnd('/')

                // Fetch the real media title from the entity
                try {
                    val mediaFlow = container.mediaRepository.getMediaById(mediaId)
                    val item = mediaFlow.first()
                    if (item != null && item.title.isNotBlank()) {
                        resolvedTitle = item.title
                    }
                } catch (_: Exception) { /* keep default title */ }

                // Call /api/v1/entities/:id/stream to get the stream URL
                val response = container.api.getEntityStream(mediaId)
                if (response.isSuccessful) {
                    val body = response.body()
                    val streamPath = body?.get("stream_url")?.jsonPrimitive?.content
                    if (streamPath != null) {
                        resolvedUrl = if (streamPath.startsWith("/")) "$baseUrl$streamPath" else streamPath
                    } else {
                        streamError = "No stream URL in response"
                    }
                } else {
                    streamError = "Stream unavailable (${response.code()})"
                }
            } catch (e: Exception) {
                streamError = "Failed to get stream: ${e.message}"
            }
        }
    }

    // Initialize ExoPlayer when URL is resolved, with auth headers
    DisposableEffect(resolvedUrl) {
        var player: ExoPlayer? = null
        if (resolvedUrl.isNotEmpty()) {
            try {
                val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
                val token = container.authRepository.authState.value.token

                val dataSourceFactory = androidx.media3.datasource.DefaultHttpDataSource.Factory()
                if (token != null) {
                    dataSourceFactory.setDefaultRequestProperties(
                        mapOf("Authorization" to "Bearer $token")
                    )
                }

                val mediaSourceFactory = androidx.media3.exoplayer.source.DefaultMediaSourceFactory(dataSourceFactory)

                player = ExoPlayer.Builder(context)
                    .setMediaSourceFactory(mediaSourceFactory)
                    .setSeekForwardIncrementMs(10_000)
                    .setSeekBackIncrementMs(10_000)
                    .build().apply {
                        setMediaItem(MediaItem.fromUri(resolvedUrl))
                        prepare()
                        playWhenReady = true
                    }
                exoPlayer = player
            } catch (e: Exception) {
                streamError = "Player error: ${e.message}"
            }
        }
        onDispose {
            player?.release()
            if (exoPlayer == player) {
                exoPlayer = null
            }
        }
    }

    // Update position, buffer, and track info periodically
    LaunchedEffect(exoPlayer) {
        exoPlayer?.let { player ->
            // Extract tracks once player is ready
            player.addListener(object : Player.Listener {
                override fun onTracksChanged(tracks: Tracks) {
                    val audio = mutableListOf<TrackInfo>()
                    val subtitle = mutableListOf<TrackInfo>()
                    for (group in tracks.groups) {
                        for (i in 0 until group.length) {
                            val format = group.getTrackFormat(i)
                            val label = format.label ?: format.language ?: "Track ${i + 1}"
                            when {
                                group.type == C.TRACK_TYPE_AUDIO ->
                                    audio.add(TrackInfo(label, i, group.type, group.isTrackSelected(i)))
                                group.type == C.TRACK_TYPE_TEXT ->
                                    subtitle.add(TrackInfo(label, i, group.type, group.isTrackSelected(i)))
                            }
                        }
                    }
                    audioTracks = audio
                    subtitleTracks = subtitle
                    selectedAudioIndex = audio.indexOfFirst { it.selected }.coerceAtLeast(0)
                    selectedSubtitleIndex = subtitle.indexOfFirst { it.selected }
                }
            })
            while (true) {
                currentPosition = player.currentPosition
                duration = player.duration.coerceAtLeast(0)
                bufferedPosition = player.bufferedPosition
                isPlaying = player.isPlaying
                kotlinx.coroutines.delay(500)
            }
        }
    }

    // Auto-hide controls after 5 seconds of playback
    LaunchedEffect(showControls, isPlaying) {
        if (showControls && isPlaying) {
            kotlinx.coroutines.delay(5000)
            if (!showSpeedMenu && !showAudioMenu && !showSubtitleMenu) {
                showControls = false
            }
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .focusRequester(controlsFocus)
            .focusable()
            .onKeyEvent { event ->
                if (event.type != KeyEventType.KeyDown) return@onKeyEvent false
                val player = exoPlayer ?: return@onKeyEvent false
                showControls = true
                when (event.key) {
                    Key.DirectionCenter, Key.Enter -> {
                        if (player.isPlaying) player.pause() else player.play()
                        true
                    }
                    Key.DirectionLeft -> {
                        player.seekBack()
                        true
                    }
                    Key.DirectionRight -> {
                        player.seekForward()
                        true
                    }
                    Key.Back -> {
                        if (showSpeedMenu || showAudioMenu || showSubtitleMenu) {
                            showSpeedMenu = false
                            showAudioMenu = false
                            showSubtitleMenu = false
                        } else {
                            onNavigateBack()
                        }
                        true
                    }
                    else -> false
                }
            }
    ) {
        // ExoPlayer View (native controls hidden -- we use our own overlay)
        if (exoPlayer != null) {
            AndroidView(
                factory = { ctx ->
                    PlayerView(ctx).apply {
                        this.player = exoPlayer
                        useController = false // We use our custom TV overlay
                    }
                },
                update = { playerView ->
                    playerView.player = exoPlayer
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // VLC-style custom TV overlay controls
        if (exoPlayer != null && showControls) {
            // Full overlay container
            Column(
                modifier = Modifier.fillMaxSize()
            ) {
                // Top bar: Back + Title (VLC gradient)
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Brush.verticalGradient(OverlayGradientTop))
                        .padding(horizontal = 24.dp, vertical = 16.dp)
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        // Back button (VLC style: circular)
                        PlayerIconButton(
                            icon = Icons.Default.ArrowBack,
                            contentDescription = "Back",
                            onClick = onNavigateBack,
                            tint = Color.White,
                            size = 36.dp
                        )
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = resolvedTitle,
                                style = MaterialTheme.typography.headlineSmall,
                                color = Color.White,
                                fontWeight = FontWeight.Bold,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis
                            )
                            if (playbackSpeed != 1.0f) {
                                Text(
                                    text = "${playbackSpeed}x speed",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = VlcOrange
                                )
                            }
                        }
                    }
                }

                Spacer(modifier = Modifier.weight(1f))

                // Center play/pause button (large, VLC-style)
                Box(
                    modifier = Modifier.fillMaxWidth(),
                    contentAlignment = Alignment.Center
                ) {
                    Box(
                        modifier = Modifier
                            .size(72.dp)
                            .clip(CircleShape)
                            .background(VlcOrange.copy(alpha = 0.9f)),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(
                            imageVector = if (isPlaying) Icons.Default.Pause else Icons.Default.PlayArrow,
                            contentDescription = if (isPlaying) "Pause" else "Play",
                            tint = Color.White,
                            modifier = Modifier.size(40.dp)
                        )
                    }
                }

                Spacer(modifier = Modifier.weight(1f))

                // Bottom bar: Seekbar + Controls (VLC gradient)
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Brush.verticalGradient(OverlayGradientBottom))
                        .padding(horizontal = 24.dp, vertical = 16.dp)
                ) {
                    // Progress / seek bar
                    if (duration > 0) {
                        // Buffered progress underneath the slider
                        Box(modifier = Modifier.fillMaxWidth()) {
                            // Buffer indicator (thin bar behind slider)
                            LinearProgressIndicator(
                                progress = (bufferedPosition.toFloat() / duration.toFloat().coerceAtLeast(1f))
                                    .coerceIn(0f, 1f),
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(4.dp)
                                    .align(Alignment.Center),
                                color = Color.White.copy(alpha = 0.3f),
                                trackColor = Color.White.copy(alpha = 0.1f)
                            )
                        }

                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = formatTime(currentPosition),
                                style = MaterialTheme.typography.bodySmall,
                                color = Color.White,
                                modifier = Modifier.width(64.dp)
                            )
                            Slider(
                                value = currentPosition.toFloat(),
                                onValueChange = { exoPlayer?.seekTo(it.toLong()) },
                                valueRange = 0f..duration.toFloat().coerceAtLeast(1f),
                                modifier = Modifier.weight(1f),
                                colors = SliderDefaults.colors(
                                    thumbColor = VlcOrange,
                                    activeTrackColor = VlcOrange,
                                    inactiveTrackColor = Color.White.copy(alpha = 0.3f)
                                )
                            )
                            Text(
                                text = formatTime(duration),
                                style = MaterialTheme.typography.bodySmall,
                                color = Color.White,
                                modifier = Modifier.width(64.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(8.dp))

                    // Control buttons row (VLC-style)
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceEvenly,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        // Seek backward 10s
                        PlayerControlButton(
                            icon = Icons.Default.Replay10,
                            label = "-10s",
                            onClick = { exoPlayer?.seekBack() }
                        )

                        // Play/Pause
                        PlayerControlButton(
                            icon = if (isPlaying) Icons.Default.Pause else Icons.Default.PlayArrow,
                            label = if (isPlaying) "Pause" else "Play",
                            onClick = {
                                val p = exoPlayer ?: return@PlayerControlButton
                                if (p.isPlaying) p.pause() else p.play()
                            },
                            accentColor = VlcOrange
                        )

                        // Seek forward 10s
                        PlayerControlButton(
                            icon = Icons.Default.Forward10,
                            label = "+10s",
                            onClick = { exoPlayer?.seekForward() }
                        )

                        // Speed control
                        PlayerControlButton(
                            icon = Icons.Default.Speed,
                            label = "${playbackSpeed}x",
                            onClick = {
                                showSpeedMenu = !showSpeedMenu
                                showAudioMenu = false
                                showSubtitleMenu = false
                            },
                            accentColor = if (playbackSpeed != 1.0f) VlcOrange else Color.White
                        )

                        // Audio track
                        PlayerControlButton(
                            icon = Icons.Default.Audiotrack,
                            label = "Audio",
                            onClick = {
                                showAudioMenu = !showAudioMenu
                                showSpeedMenu = false
                                showSubtitleMenu = false
                            }
                        )

                        // Subtitles
                        PlayerControlButton(
                            icon = Icons.Default.Subtitles,
                            label = if (selectedSubtitleIndex >= 0) "On" else "Off",
                            onClick = {
                                showSubtitleMenu = !showSubtitleMenu
                                showSpeedMenu = false
                                showAudioMenu = false
                            },
                            accentColor = if (selectedSubtitleIndex >= 0) VlcOrange else Color.White
                        )
                    }
                }
            }

            // Speed menu popup
            if (showSpeedMenu) {
                OverlayMenu(
                    title = "Playback Speed",
                    items = speedOptions.map { "${it}x" },
                    selectedIndex = speedOptions.indexOf(playbackSpeed),
                    onSelect = { index ->
                        playbackSpeed = speedOptions[index]
                        exoPlayer?.setPlaybackSpeed(playbackSpeed)
                        showSpeedMenu = false
                    },
                    onDismiss = { showSpeedMenu = false }
                )
            }

            // Audio track menu popup
            if (showAudioMenu) {
                val labels = if (audioTracks.isEmpty()) listOf("Default") else audioTracks.map { it.label }
                OverlayMenu(
                    title = "Audio Track",
                    items = labels,
                    selectedIndex = selectedAudioIndex,
                    onSelect = { index ->
                        selectedAudioIndex = index
                        showAudioMenu = false
                    },
                    onDismiss = { showAudioMenu = false }
                )
            }

            // Subtitle menu popup
            if (showSubtitleMenu) {
                val labels = listOf("Off") + subtitleTracks.map { it.label }
                OverlayMenu(
                    title = "Subtitles",
                    items = labels,
                    selectedIndex = selectedSubtitleIndex + 1, // +1 for "Off" at index 0
                    onSelect = { index ->
                        selectedSubtitleIndex = index - 1
                        exoPlayer?.let { player ->
                            if (index == 0) {
                                player.trackSelectionParameters = player.trackSelectionParameters
                                    .buildUpon()
                                    .setTrackTypeDisabled(C.TRACK_TYPE_TEXT, true)
                                    .build()
                            } else {
                                player.trackSelectionParameters = player.trackSelectionParameters
                                    .buildUpon()
                                    .setTrackTypeDisabled(C.TRACK_TYPE_TEXT, false)
                                    .build()
                            }
                        }
                        showSubtitleMenu = false
                    },
                    onDismiss = { showSubtitleMenu = false }
                )
            }
        }

        // Loading or error state
        if (exoPlayer == null) {
            Column(
                modifier = Modifier.fillMaxSize(),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                if (streamError != null) {
                    Icon(
                        imageVector = Icons.Default.ErrorOutline,
                        contentDescription = "Error",
                        tint = VlcOrange,
                        modifier = Modifier.size(48.dp)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = "Unable to Play Media",
                        style = MaterialTheme.typography.headlineSmall,
                        color = Color.White
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = streamError ?: "Playback unavailable",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                    Spacer(modifier = Modifier.height(24.dp))
                    Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                        Button(onClick = {
                            streamError = null
                            resolvedUrl = ""
                            retryCount++
                        }) {
                            Text("Retry")
                        }
                        Button(onClick = onNavigateBack) {
                            Text("Back to Library")
                        }
                    }
                } else if (resolvedUrl.isEmpty()) {
                    CircularProgressIndicator(color = VlcOrange)
                    Spacer(modifier = Modifier.height(12.dp))
                    Text(
                        text = "Loading stream...",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                } else {
                    CircularProgressIndicator(color = VlcOrange)
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "Buffering...",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            }
        }
    }

    // Auto-focus controls
    LaunchedEffect(exoPlayer) {
        if (exoPlayer != null) {
            kotlinx.coroutines.delay(500)
            try { controlsFocus.requestFocus() } catch (_: Exception) {}
        }
    }
}

/**
 * VLC-style control button using TV material3 Button with icon + label.
 */
@Composable
private fun PlayerControlButton(
    icon: ImageVector,
    label: String,
    onClick: () -> Unit,
    accentColor: Color = Color.White
) {
    Button(onClick = onClick) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                modifier = Modifier.size(20.dp),
                tint = accentColor
            )
            Text(
                text = label,
                color = accentColor,
                fontSize = 12.sp
            )
        }
    }
}

/**
 * Simple circular icon button for the top bar using a clickable Box.
 */
@Composable
private fun PlayerIconButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    tint: Color = Color.White,
    size: androidx.compose.ui.unit.Dp = 36.dp
) {
    Surface(
        onClick = onClick,
        modifier = Modifier.size(size),
        shape = ClickableSurfaceDefaults.shape(shape = CircleShape),
        colors = ClickableSurfaceDefaults.colors(
            containerColor = Color.White.copy(alpha = 0.15f)
        )
    ) {
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = icon,
                contentDescription = contentDescription,
                tint = tint,
                modifier = Modifier.size(size * 0.6f)
            )
        }
    }
}

// Track info for audio/subtitle selection
private data class TrackInfo(
    val label: String,
    val index: Int,
    val type: Int,
    val selected: Boolean
)

// Overlay menu for speed/audio/subtitle selection (VLC-style)
@Composable
private fun OverlayMenu(
    title: String,
    items: List<String>,
    selectedIndex: Int,
    onSelect: (Int) -> Unit,
    onDismiss: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.7f)),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier
                .widthIn(max = 300.dp)
                .background(Color(0xFF1E1E1E), RoundedCornerShape(12.dp))
                .padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.titleMedium,
                color = VlcOrange,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(12.dp))
            items.forEachIndexed { index, label ->
                val isSelected = index == selectedIndex
                Button(
                    onClick = { onSelect(index) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(44.dp)
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = label,
                            color = if (isSelected) VlcOrange else Color.White
                        )
                        if (isSelected) {
                            Icon(
                                imageVector = Icons.Default.Check,
                                contentDescription = "Selected",
                                modifier = Modifier.size(18.dp),
                                tint = VlcOrange
                            )
                        }
                    }
                }
                if (index < items.lastIndex) {
                    Spacer(modifier = Modifier.height(4.dp))
                }
            }
            Spacer(modifier = Modifier.height(8.dp))
            Button(
                onClick = onDismiss,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(40.dp)
            ) {
                Text("Close")
            }
        }
    }
}

private fun formatTime(timeMs: Long): String {
    if (timeMs <= 0) return "00:00"
    val seconds = (timeMs / 1000) % 60
    val minutes = (timeMs / (1000 * 60)) % 60
    val hours = timeMs / (1000 * 60 * 60)
    return if (hours > 0) {
        String.format("%d:%02d:%02d", hours, minutes, seconds)
    } else {
        String.format("%02d:%02d", minutes, seconds)
    }
}
