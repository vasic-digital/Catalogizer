package com.catalogizer.androidtv.ui.screens.player

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.media3.common.MediaItem
import androidx.media3.common.util.UnstableApi
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.PlayerView
import androidx.tv.material3.*

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
    var streamError by remember { mutableStateOf<String?>(null) }

    // Fetch stream URL from entity endpoint if not provided
    LaunchedEffect(mediaId) {
        if (resolvedUrl.isEmpty()) {
            try {
                val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
                val baseUrl = container.getServerUrl().trimEnd('/')
                // Call /api/v1/entities/:id/stream to get the stream URL
                val response = container.api.getEntityStream(mediaId)
                if (response.isSuccessful) {
                    val body = response.body()
                    val streamPath = body?.get("stream_url")?.asString
                    if (streamPath != null) {
                        resolvedUrl = if (streamPath.startsWith("/")) "$baseUrl$streamPath" else streamPath
                        android.util.Log.d("Player", "Stream URL resolved: $resolvedUrl")
                    } else {
                        streamError = "No stream URL in response"
                    }
                } else {
                    streamError = "Stream unavailable (${response.code()})"
                }
            } catch (e: Exception) {
                streamError = "Failed to get stream: ${e.message}"
                android.util.Log.e("Player", "Stream fetch error", e)
            }
        }
    }

    // Initialize ExoPlayer when URL is resolved, with auth headers
    DisposableEffect(resolvedUrl) {
        var player: ExoPlayer? = null
        if (resolvedUrl.isNotEmpty()) {
            try {
                // Get auth token for authenticated streaming
                val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
                val token = container.authRepository.authState.value.token

                // Build data source factory with Authorization header
                val dataSourceFactory = androidx.media3.datasource.DefaultHttpDataSource.Factory()
                if (token != null) {
                    dataSourceFactory.setDefaultRequestProperties(
                        mapOf("Authorization" to "Bearer $token")
                    )
                }

                val mediaSourceFactory = androidx.media3.exoplayer.source.DefaultMediaSourceFactory(dataSourceFactory)

                player = ExoPlayer.Builder(context)
                    .setMediaSourceFactory(mediaSourceFactory)
                    .build().apply {
                        setMediaItem(MediaItem.fromUri(resolvedUrl))
                        prepare()
                        playWhenReady = true
                    }
                exoPlayer = player
                android.util.Log.d("Player", "ExoPlayer initialized with auth, URL: $resolvedUrl")
            } catch (e: Exception) {
                streamError = "Player error: ${e.message}"
                android.util.Log.e("Player", "ExoPlayer init failed", e)
            }
        }
        onDispose {
            player?.release()
            if (exoPlayer == player) {
                exoPlayer = null
            }
        }
    }

    // Update position
    LaunchedEffect(exoPlayer) {
        exoPlayer?.let { player ->
            while (true) {
                currentPosition = player.currentPosition
                duration = player.duration
                isPlaying = player.isPlaying
                kotlinx.coroutines.delay(1000)
            }
        }
    }

    Box(modifier = Modifier.fillMaxSize()) {
        // ExoPlayer View
        AndroidView(
            factory = { ctx ->
                PlayerView(ctx).apply {
                    player = exoPlayer
                    useController = true
                    controllerAutoShow = true
                }
            },
            modifier = Modifier.fillMaxSize()
        )

        // Overlay controls for TV remote
        if (exoPlayer != null) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(16.dp),
                verticalArrangement = Arrangement.SpaceBetween
            ) {
                // Top bar with title and back button
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(Color.Black.copy(alpha = 0.5f))
                        .padding(16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = mediaTitle,
                        style = MaterialTheme.typography.headlineMedium,
                        color = Color.White
                    )
                    Button(onClick = onNavigateBack) {
                        Text("Back")
                    }
                }

                // Bottom info (could be expanded with more controls)
                if (duration > 0) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .background(Color.Black.copy(alpha = 0.5f))
                            .padding(16.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = "${formatTime(currentPosition)} / ${formatTime(duration)}",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.White
                        )
                        Text(
                            text = if (isPlaying) "Playing" else "Paused",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.White
                        )
                    }
                }
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
                    Text(
                        text = streamError ?: "Playback unavailable",
                        style = MaterialTheme.typography.headlineSmall
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "Media ID: $mediaId",
                        style = MaterialTheme.typography.bodyLarge
                    )
                } else if (resolvedUrl.isEmpty()) {
                    CircularProgressIndicator()
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "Loading stream...",
                        style = MaterialTheme.typography.bodyLarge
                    )
                } else {
                    CircularProgressIndicator()
                    Text(
                        text = "Buffering...",
                        style = MaterialTheme.typography.bodyLarge
                    )
                }
                Button(onClick = onNavigateBack) {
                    Text("Back")
                }
            }
        }
    }

}


private fun formatTime(timeMs: Long): String {
    val seconds = (timeMs / 1000) % 60
    val minutes = (timeMs / (1000 * 60)) % 60
    val hours = timeMs / (1000 * 60 * 60)
    return if (hours > 0) {
        String.format("%d:%02d:%02d", hours, minutes, seconds)
    } else {
        String.format("%02d:%02d", minutes, seconds)
    }
}