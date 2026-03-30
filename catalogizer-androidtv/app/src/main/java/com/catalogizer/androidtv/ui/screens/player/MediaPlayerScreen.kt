package com.catalogizer.androidtv.ui.screens.player

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon as M3Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
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
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.common.Tracks
import androidx.media3.common.util.UnstableApi
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.PlayerView
import androidx.tv.material3.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first

// VLC orange accent color
private val VLCOrange = Color(0xFFFF6600)
private val VLCOrangeDark = Color(0xFFCC5200)
private val VLCOrangeLight = Color(0xFFFF8533)
private val OverlayBg = Color.Black.copy(alpha = 0.5f)

// Playback speed presets
private val speedOptions = listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f, 3.0f)

@OptIn(ExperimentalTvMaterial3Api::class)
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
    var isBuffering by remember { mutableStateOf(false) }

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
                    val streamPath = body?.get("stream_url")?.asString
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
                    .setConnectTimeoutMs(15_000)
                    .setReadTimeoutMs(30_000)
                    .setAllowCrossProtocolRedirects(true)
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

    // Listen to player events: errors, buffering, tracks, position
    LaunchedEffect(exoPlayer) {
        exoPlayer?.let { player ->
            player.addListener(object : Player.Listener {
                override fun onPlayerError(error: PlaybackException) {
                    val cause = error.cause?.message ?: error.message ?: "Unknown playback error"
                    streamError = when (error.errorCode) {
                        PlaybackException.ERROR_CODE_IO_NETWORK_CONNECTION_FAILED ->
                            "Network connection failed. Check your server connection."
                        PlaybackException.ERROR_CODE_IO_NETWORK_CONNECTION_TIMEOUT ->
                            "Connection timed out. The server may be slow or unreachable."
                        PlaybackException.ERROR_CODE_IO_BAD_HTTP_STATUS ->
                            "Server returned an error. The file may not be accessible."
                        PlaybackException.ERROR_CODE_IO_FILE_NOT_FOUND ->
                            "File not found on the server."
                        PlaybackException.ERROR_CODE_DECODER_INIT_FAILED ->
                            "Cannot decode this media format."
                        PlaybackException.ERROR_CODE_DECODING_FAILED ->
                            "Decoding error. The file may be corrupted."
                        else -> "Playback error: $cause"
                    }
                }

                override fun onPlaybackStateChanged(playbackState: Int) {
                    isBuffering = playbackState == Player.STATE_BUFFERING
                }

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

            // Position/duration polling loop
            while (true) {
                currentPosition = player.currentPosition
                duration = player.duration.coerceAtLeast(0)
                bufferedPosition = player.bufferedPosition
                isPlaying = player.isPlaying
                delay(500)
            }
        }
    }

    // Auto-hide controls after 5 seconds of playback
    LaunchedEffect(showControls, isPlaying) {
        if (showControls && isPlaying && !showSpeedMenu && !showAudioMenu && !showSubtitleMenu) {
            delay(5000)
            showControls = false
        }
    }

    // ---- Main layout ----
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .focusRequester(controlsFocus)
            .focusable()
            .onKeyEvent { event ->
                if (event.type != KeyEventType.KeyDown) return@onKeyEvent false
                val player = exoPlayer ?: return@onKeyEvent false

                // Any key press shows controls
                if (!showControls) {
                    showControls = true
                    return@onKeyEvent true
                }

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
                    Key.DirectionUp, Key.DirectionDown -> {
                        showControls = true
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
        // ---- Video surface (ExoPlayer) ----
        if (exoPlayer != null) {
            AndroidView(
                factory = { ctx ->
                    PlayerView(ctx).apply {
                        this.player = exoPlayer
                        useController = false
                        setShowBuffering(PlayerView.SHOW_BUFFERING_NEVER)
                    }
                },
                update = { playerView ->
                    playerView.player = exoPlayer
                },
                modifier = Modifier.fillMaxSize()
            )
        }

        // ---- VLC-style controls overlay ----
        if (exoPlayer != null && streamError == null) {
            AnimatedVisibility(
                visible = showControls,
                enter = fadeIn(),
                exit = fadeOut()
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(OverlayBg)
                ) {
                    // Top bar: Back arrow + Title
                    VLCTopBar(
                        title = resolvedTitle,
                        playbackSpeed = playbackSpeed,
                        onBack = onNavigateBack,
                        modifier = Modifier.align(Alignment.TopCenter)
                    )

                    // Center: Large play/pause button + buffering indicator
                    VLCCenterControls(
                        isPlaying = isPlaying,
                        isBuffering = isBuffering,
                        onTogglePlayPause = {
                            val p = exoPlayer ?: return@VLCCenterControls
                            if (p.isPlaying) p.pause() else p.play()
                        },
                        onSeekBack = { exoPlayer?.seekBack() },
                        onSeekForward = { exoPlayer?.seekForward() },
                        modifier = Modifier.align(Alignment.Center)
                    )

                    // Bottom bar: Seekbar + time + control buttons
                    VLCBottomBar(
                        currentPosition = currentPosition,
                        duration = duration,
                        bufferedPosition = bufferedPosition,
                        playbackSpeed = playbackSpeed,
                        selectedSubtitleIndex = selectedSubtitleIndex,
                        onSeek = { exoPlayer?.seekTo(it) },
                        onSpeedClick = {
                            showSpeedMenu = !showSpeedMenu
                            showAudioMenu = false
                            showSubtitleMenu = false
                        },
                        onAudioClick = {
                            showAudioMenu = !showAudioMenu
                            showSpeedMenu = false
                            showSubtitleMenu = false
                        },
                        onSubtitleClick = {
                            showSubtitleMenu = !showSubtitleMenu
                            showSpeedMenu = false
                            showAudioMenu = false
                        },
                        modifier = Modifier.align(Alignment.BottomCenter)
                    )
                }
            }

            // Buffering spinner shown even when controls are hidden
            if (isBuffering && !showControls) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(
                        color = VLCOrange,
                        strokeWidth = 3.dp,
                        modifier = Modifier.size(48.dp)
                    )
                }
            }
        }

        // ---- Popup menus ----
        if (showSpeedMenu) {
            VLCOverlayMenu(
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

        if (showAudioMenu) {
            val labels = if (audioTracks.isEmpty()) listOf("Default") else audioTracks.map { it.label }
            VLCOverlayMenu(
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

        if (showSubtitleMenu) {
            val labels = listOf("Off") + subtitleTracks.map { it.label }
            VLCOverlayMenu(
                title = "Subtitles",
                items = labels,
                selectedIndex = selectedSubtitleIndex + 1,
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

        // ---- Loading / error state (no player yet) ----
        if (exoPlayer == null || streamError != null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black),
                contentAlignment = Alignment.Center
            ) {
                if (streamError != null) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier.padding(48.dp)
                    ) {
                        Box(
                            modifier = Modifier
                                .size(72.dp)
                                .background(Color(0xFF2A2A2A), CircleShape),
                            contentAlignment = Alignment.Center
                        ) {
                            M3Icon(
                                Icons.Default.ErrorOutline,
                                contentDescription = "Error",
                                tint = VLCOrange,
                                modifier = Modifier.size(40.dp)
                            )
                        }
                        Spacer(modifier = Modifier.height(20.dp))
                        Text(
                            text = "Unable to Play Media",
                            style = MaterialTheme.typography.headlineSmall,
                            color = Color.White,
                            fontWeight = FontWeight.Bold
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = streamError ?: "Playback unavailable",
                            style = MaterialTheme.typography.bodyLarge,
                            color = Color.White.copy(alpha = 0.6f)
                        )
                        Spacer(modifier = Modifier.height(32.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                            Button(onClick = {
                                streamError = null
                                resolvedUrl = ""
                                retryCount++
                            }) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    M3Icon(Icons.Default.Refresh, "Retry", Modifier.size(18.dp))
                                    Text("Retry")
                                }
                            }
                            Button(onClick = onNavigateBack) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    M3Icon(Icons.Default.ArrowBack, "Back", Modifier.size(18.dp))
                                    Text("Back")
                                }
                            }
                        }
                    }
                } else if (resolvedUrl.isEmpty()) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        CircularProgressIndicator(
                            color = VLCOrange,
                            strokeWidth = 3.dp,
                            modifier = Modifier.size(48.dp)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            "Loading stream...",
                            style = MaterialTheme.typography.bodyLarge,
                            color = Color.White.copy(alpha = 0.7f)
                        )
                    }
                } else {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        CircularProgressIndicator(
                            color = VLCOrange,
                            strokeWidth = 3.dp,
                            modifier = Modifier.size(48.dp)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            "Preparing playback...",
                            style = MaterialTheme.typography.bodyLarge,
                            color = Color.White.copy(alpha = 0.7f)
                        )
                    }
                }
            }
        }
    }

    // Auto-focus controls when player is ready
    LaunchedEffect(exoPlayer) {
        if (exoPlayer != null) {
            delay(500)
            try { controlsFocus.requestFocus() } catch (_: Exception) {}
        }
    }
}

// ---- VLC-style Top Bar ----
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun VLCTopBar(
    title: String,
    playbackSpeed: Float,
    onBack: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(
                Brush.verticalGradient(
                    listOf(Color.Black.copy(alpha = 0.8f), Color.Transparent)
                )
            )
            .padding(horizontal = 32.dp, vertical = 20.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Back arrow button
            Surface(
                onClick = onBack,
                modifier = Modifier.size(40.dp),
                shape = ClickableSurfaceDefaults.shape(
                    shape = CircleShape,
                    focusedShape = CircleShape,
                    pressedShape = CircleShape
                ),
                colors = ClickableSurfaceDefaults.colors(
                    containerColor = Color.White.copy(alpha = 0.1f),
                    focusedContainerColor = Color.White.copy(alpha = 0.2f),
                    pressedContainerColor = Color.White.copy(alpha = 0.3f)
                )
            ) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    M3Icon(
                        Icons.Default.ArrowBack,
                        contentDescription = "Back",
                        tint = Color.White,
                        modifier = Modifier.size(22.dp)
                    )
                }
            }

            Spacer(modifier = Modifier.width(16.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleLarge,
                    color = Color.White,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    fontSize = 20.sp
                )
                if (playbackSpeed != 1.0f) {
                    Text(
                        text = "${playbackSpeed}x speed",
                        style = MaterialTheme.typography.bodySmall,
                        color = VLCOrangeLight,
                        fontSize = 13.sp
                    )
                }
            }
        }
    }
}

// ---- VLC-style Center Controls (seek back, play/pause, seek forward) ----
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun VLCCenterControls(
    isPlaying: Boolean,
    isBuffering: Boolean,
    onTogglePlayPause: () -> Unit,
    onSeekBack: () -> Unit,
    onSeekForward: () -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier,
        horizontalArrangement = Arrangement.spacedBy(40.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Seek back 10s
        VLCSurfaceCircleButton(
            icon = Icons.Default.Replay10,
            contentDescription = "Rewind 10 seconds",
            size = 52,
            iconSize = 28,
            bgColor = Color.White.copy(alpha = 0.1f),
            onClick = onSeekBack
        )

        // Large play/pause button
        Box(contentAlignment = Alignment.Center) {
            if (isBuffering) {
                CircularProgressIndicator(
                    color = VLCOrange,
                    strokeWidth = 3.dp,
                    modifier = Modifier.size(84.dp)
                )
            }
            VLCSurfaceCircleButton(
                icon = if (isPlaying) Icons.Default.Pause else Icons.Default.PlayArrow,
                contentDescription = if (isPlaying) "Pause" else "Play",
                size = 72,
                iconSize = 40,
                bgColor = VLCOrange,
                focusBgColor = VLCOrangeDark,
                onClick = onTogglePlayPause
            )
        }

        // Seek forward 10s
        VLCSurfaceCircleButton(
            icon = Icons.Default.Forward10,
            contentDescription = "Forward 10 seconds",
            size = 52,
            iconSize = 28,
            bgColor = Color.White.copy(alpha = 0.1f),
            onClick = onSeekForward
        )
    }
}

// ---- VLC-style Bottom Bar ----
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun VLCBottomBar(
    currentPosition: Long,
    duration: Long,
    bufferedPosition: Long,
    playbackSpeed: Float,
    selectedSubtitleIndex: Int,
    onSeek: (Long) -> Unit,
    onSpeedClick: () -> Unit,
    onAudioClick: () -> Unit,
    onSubtitleClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(
                Brush.verticalGradient(
                    listOf(Color.Transparent, Color.Black.copy(alpha = 0.85f))
                )
            )
            .padding(horizontal = 32.dp)
            .padding(bottom = 24.dp, top = 40.dp)
    ) {
        // Seekbar with buffer progress
        if (duration > 0) {
            Box(modifier = Modifier.fillMaxWidth().height(24.dp)) {
                // Buffer progress (behind seekbar)
                LinearProgressIndicator(
                    progress = (bufferedPosition.toFloat() / duration.toFloat().coerceAtLeast(1f)).coerceIn(0f, 1f),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(4.dp)
                        .align(Alignment.Center)
                        .clip(RoundedCornerShape(2.dp)),
                    color = Color.White.copy(alpha = 0.3f),
                    trackColor = Color.White.copy(alpha = 0.1f)
                )
                // Seekbar
                Slider(
                    value = currentPosition.toFloat(),
                    onValueChange = { onSeek(it.toLong()) },
                    valueRange = 0f..duration.toFloat().coerceAtLeast(1f),
                    modifier = Modifier.fillMaxWidth(),
                    colors = SliderDefaults.colors(
                        thumbColor = VLCOrange,
                        activeTrackColor = VLCOrange,
                        inactiveTrackColor = Color.White.copy(alpha = 0.15f)
                    )
                )
            }
        }

        Spacer(modifier = Modifier.height(4.dp))

        // Time labels + control buttons row
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Left: Current time / Duration
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                Text(
                    text = formatTime(currentPosition),
                    color = Color.White,
                    fontSize = 14.sp,
                    fontWeight = FontWeight.Medium
                )
                Text(
                    text = "/",
                    color = Color.White.copy(alpha = 0.5f),
                    fontSize = 14.sp
                )
                Text(
                    text = formatTime(duration),
                    color = Color.White.copy(alpha = 0.7f),
                    fontSize = 14.sp
                )
            }

            // Right: Control buttons
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Speed
                VLCSmallSurfaceButton(
                    icon = Icons.Default.Speed,
                    label = if (playbackSpeed != 1.0f) "${playbackSpeed}x" else null,
                    contentDescription = "Speed",
                    onClick = onSpeedClick
                )

                // Audio track
                VLCSmallSurfaceButton(
                    icon = Icons.Default.Audiotrack,
                    label = null,
                    contentDescription = "Audio track",
                    onClick = onAudioClick
                )

                // Subtitles
                VLCSmallSurfaceButton(
                    icon = Icons.Default.Subtitles,
                    label = null,
                    contentDescription = "Subtitles",
                    onClick = onSubtitleClick,
                    active = selectedSubtitleIndex >= 0
                )
            }
        }
    }
}

// ---- Reusable VLC-style circle button using Surface ----
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun VLCSurfaceCircleButton(
    icon: ImageVector,
    contentDescription: String,
    size: Int,
    iconSize: Int,
    bgColor: Color,
    focusBgColor: Color = Color.White.copy(alpha = 0.2f),
    onClick: () -> Unit
) {
    var isFocused by remember { mutableStateOf(false) }

    Surface(
        onClick = onClick,
        modifier = Modifier
            .size(size.dp)
            .onFocusChanged { isFocused = it.isFocused }
            .then(
                if (isFocused) Modifier.border(
                    BorderStroke(2.dp, VLCOrange),
                    shape = CircleShape
                ) else Modifier
            ),
        shape = ClickableSurfaceDefaults.shape(
            shape = CircleShape,
            focusedShape = CircleShape,
            pressedShape = CircleShape
        ),
        colors = ClickableSurfaceDefaults.colors(
            containerColor = bgColor,
            focusedContainerColor = focusBgColor,
            pressedContainerColor = bgColor.copy(alpha = (bgColor.alpha + 0.2f).coerceAtMost(1f))
        )
    ) {
        Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            M3Icon(
                icon,
                contentDescription = contentDescription,
                tint = Color.White,
                modifier = Modifier.size(iconSize.dp)
            )
        }
    }
}

// ---- Small bottom-bar control button using Surface ----
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun VLCSmallSurfaceButton(
    icon: ImageVector,
    label: String?,
    contentDescription: String,
    onClick: () -> Unit,
    active: Boolean = false
) {
    var isFocused by remember { mutableStateOf(false) }
    val bgColor = when {
        active -> VLCOrange.copy(alpha = 0.15f)
        else -> Color.White.copy(alpha = 0.08f)
    }
    val iconTint = if (active) VLCOrange else Color.White.copy(alpha = 0.85f)

    Surface(
        onClick = onClick,
        modifier = Modifier
            .height(36.dp)
            .onFocusChanged { isFocused = it.isFocused }
            .then(
                if (isFocused) Modifier.border(
                    BorderStroke(1.dp, VLCOrange),
                    shape = RoundedCornerShape(8.dp)
                ) else Modifier
            ),
        shape = ClickableSurfaceDefaults.shape(
            shape = RoundedCornerShape(8.dp),
            focusedShape = RoundedCornerShape(8.dp),
            pressedShape = RoundedCornerShape(8.dp)
        ),
        colors = ClickableSurfaceDefaults.colors(
            containerColor = bgColor,
            focusedContainerColor = VLCOrange.copy(alpha = 0.3f),
            pressedContainerColor = VLCOrange.copy(alpha = 0.4f)
        )
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            M3Icon(
                icon,
                contentDescription = contentDescription,
                tint = iconTint,
                modifier = Modifier.size(18.dp)
            )
            if (label != null) {
                Text(
                    text = label,
                    color = iconTint,
                    fontSize = 13.sp,
                    fontWeight = FontWeight.Medium
                )
            }
        }
    }
}

// ---- VLC-style overlay menu (speed / audio / subtitles) — right-side panel ----
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun VLCOverlayMenu(
    title: String,
    items: List<String>,
    selectedIndex: Int,
    onSelect: (Int) -> Unit,
    onDismiss: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.8f)),
        contentAlignment = Alignment.CenterEnd
    ) {
        // Right-side panel (VLC-style)
        Column(
            modifier = Modifier
                .width(320.dp)
                .fillMaxHeight()
                .background(Color(0xFF1A1A1A))
                .padding(24.dp)
        ) {
            // Title
            Text(
                text = title,
                style = MaterialTheme.typography.titleLarge,
                color = VLCOrange,
                fontWeight = FontWeight.Bold,
                fontSize = 18.sp
            )

            Spacer(modifier = Modifier.height(4.dp))

            // Subtle divider
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(1.dp)
                    .background(VLCOrange.copy(alpha = 0.3f))
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Menu items
            items.forEachIndexed { index, label ->
                val isSelected = index == selectedIndex

                Surface(
                    onClick = { onSelect(index) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(44.dp),
                    shape = ClickableSurfaceDefaults.shape(
                        shape = RoundedCornerShape(8.dp),
                        focusedShape = RoundedCornerShape(8.dp),
                        pressedShape = RoundedCornerShape(8.dp)
                    ),
                    colors = ClickableSurfaceDefaults.colors(
                        containerColor = if (isSelected) VLCOrange.copy(alpha = 0.15f) else Color.Transparent,
                        focusedContainerColor = if (isSelected) VLCOrange.copy(alpha = 0.35f) else Color.White.copy(alpha = 0.15f),
                        pressedContainerColor = VLCOrange.copy(alpha = 0.4f)
                    )
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(horizontal = 16.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = label,
                            color = if (isSelected) VLCOrange else Color.White.copy(alpha = 0.85f),
                            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Normal,
                            fontSize = 15.sp
                        )
                        if (isSelected) {
                            M3Icon(
                                Icons.Default.Check,
                                "Selected",
                                modifier = Modifier.size(18.dp),
                                tint = VLCOrange
                            )
                        }
                    }
                }

                if (index < items.lastIndex) {
                    Spacer(modifier = Modifier.height(2.dp))
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            // Close button at bottom
            Surface(
                onClick = onDismiss,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(44.dp),
                shape = ClickableSurfaceDefaults.shape(
                    shape = RoundedCornerShape(8.dp),
                    focusedShape = RoundedCornerShape(8.dp),
                    pressedShape = RoundedCornerShape(8.dp)
                ),
                colors = ClickableSurfaceDefaults.colors(
                    containerColor = Color.White.copy(alpha = 0.08f),
                    focusedContainerColor = Color.White.copy(alpha = 0.15f),
                    pressedContainerColor = Color.White.copy(alpha = 0.2f)
                )
            ) {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text("Close", color = Color.White.copy(alpha = 0.7f), fontSize = 14.sp)
                }
            }
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
