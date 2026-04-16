@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.screens.media

import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon as M3Icon
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*

import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.ui.components.CoverImage
import com.catalogizer.androidtv.ui.components.HistoryDialog
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Enhanced media detail screen with improved UI/UX.
 * Features hero poster, metadata badges, action buttons with clear CTAs,
 * synopsis, and file info with proper spacing and visual feedback.
 */
@Composable
fun MediaDetailScreen(
    mediaId: Long,
    onNavigateBack: () -> Unit,
    onNavigateToPlayer: (Long) -> Unit
) {
    val container = DependencyContainer.getInstance(androidx.compose.ui.platform.LocalContext.current)
    var mediaItem by remember { mutableStateOf<MediaItem?>(null) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }
    var retryCount by remember { mutableStateOf(0) }
    var isFavorite by remember { mutableStateOf(false) }
    var showHistory by remember { mutableStateOf(false) }
    val coroutineScope = rememberCoroutineScope()
    val scrollState = rememberScrollState()
    val playButtonFocus = remember { FocusRequester() }
    val config = LocalConfiguration.current
    val isCompact = config.screenWidthDp < 600

    // Fetch entity details from API with retry
    LaunchedEffect(mediaId, retryCount) {
        isLoading = true
        var lastError: String? = null
        for (attempt in 0..2) {
            try {
                val result = container.mediaRepository.getMediaById(mediaId).first()
                if (result != null) {
                    mediaItem = result
                    error = null
                    val entityType = result.mediaType ?: "movie"
                    isFavorite = container.mediaRepository.checkFavorite(entityType, mediaId)
                    lastError = null
                    break
                } else {
                    lastError = "Media item not found (id=$mediaId). The server may still be processing this entry."
                }
            } catch (e: java.net.ConnectException) {
                lastError = "Cannot connect to server. Check that the server is running and reachable."
            } catch (e: java.net.SocketTimeoutException) {
                lastError = "Server request timed out. The server may be busy or unreachable."
            } catch (e: Exception) {
                lastError = "Failed to load media: ${e.message}"
            }
            if (attempt < 2 && lastError != null) {
                kotlinx.coroutines.delay((attempt + 1) * 1000L)
            }
        }
        error = lastError
        isLoading = false
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFF0A0A0A))
    ) {
        when {
            isLoading -> {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = Color.White)
                }
            }
            error != null -> {
                Column(
                    Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Text(
                        "Unable to Load Media",
                        style = MaterialTheme.typography.headlineMedium,
                        color = Color.White,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(Modifier.height(8.dp))
                    Text(
                        "Please check your network connection and verify the server is reachable.",
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color.White.copy(alpha = 0.6f)
                    )
                    Spacer(Modifier.height(12.dp))
                    Text(
                        error ?: "An unknown error occurred",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.85f)
                    )
                    Spacer(Modifier.height(32.dp))
                    Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                        Button(onClick = {
                            error = null
                            isLoading = true
                            retryCount++
                        }) {
                            Text("Retry")
                        }
                        // Simplified back button with full text visibility
                        Button(onClick = onNavigateBack) {
                            Text("Back")
                        }
                    }
                }
            }
            mediaItem != null -> {
                val item = mediaItem!!
                val coverUrl = item.thumbnailUrl?.let { url ->
                    when {
                        url.startsWith("/") ->
                            container.getServerUrl().trimEnd('/') + url
                        url.contains("image.tmdb.org") -> {
                            val encoded = java.net.URLEncoder.encode(url, "UTF-8")
                            container.getServerUrl().trimEnd('/') +
                                "/api/v1/image-proxy?url=$encoded"
                        }
                        else -> url
                    }
                }

                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(scrollState)
                ) {
                    // Hero poster section
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(if (isCompact) 280.dp else 400.dp),
                        contentAlignment = Alignment.TopCenter
                    ) {
                        if (coverUrl != null) {
                            CoverImage(
                                url = coverUrl,
                                contentDescription = item.title,
                                modifier = Modifier.fillMaxWidth(),
                                contentScale = ContentScale.FillWidth
                            )
                        } else {
                            Box(
                                Modifier.fillMaxSize().background(
                                    Brush.verticalGradient(listOf(Color(0xFF1A237E), Color(0xFF0A0A0A)))
                                )
                            )
                        }
                        // Gradient overlay for text readability
                        Box(
                            Modifier.fillMaxSize().background(
                                Brush.verticalGradient(
                                    listOf(Color.Transparent, Color(0xCC0A0A0A), Color(0xFF0A0A0A)),
                                    startY = 100f
                                )
                            )
                        )
                        // Back button with icon and label for clarity
                        Button(
                            onClick = onNavigateBack,
                            modifier = Modifier.padding(16.dp).align(Alignment.TopStart),
                            scale = ButtonDefaults.scale(focusedScale = 1.08f),
                            glow = ButtonDefaults.glow(focusedGlow = Glow(elevation = 6.dp, elevationColor = MaterialTheme.colorScheme.primary))
                        ) {
                            Box(
                                modifier = Modifier.fillMaxSize(),
                                contentAlignment = Alignment.Center
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    M3Icon(Icons.Default.ArrowBack, "Back", Modifier.size(20.dp))
                                    Text("Back")
                                }
                            }
                        }
                    }

                    // Metadata content with improved spacing
                    Column(
                        modifier = Modifier.padding(
                            horizontal = if (isCompact) 24.dp else 48.dp,
                            vertical = 24.dp
                        )
                    ) {
                        // Title
                        Text(
                            text = item.title,
                            style = if (isCompact) MaterialTheme.typography.headlineSmall
                                    else MaterialTheme.typography.headlineMedium,
                            color = Color.White,
                            fontWeight = FontWeight.Bold,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis
                        )

                        Spacer(Modifier.height(12.dp))

                        // Year + Rating + Type badges with improved styling
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(16.dp)
                        ) {
                            item.year?.let {
                                Text(
                                    it.toString(), 
                                    color = Color.White.copy(0.8f), 
                                    style = MaterialTheme.typography.bodyLarge
                                )
                            }
                            item.rating?.let { r ->
                                if (r > 0) {
                                    Row(verticalAlignment = Alignment.CenterVertically) {
                                        M3Icon(
                                            Icons.Default.Star, 
                                            "Rating", 
                                            Modifier.size(18.dp), 
                                            tint = Color(0xFFFFD700)
                                        )
                                        Spacer(Modifier.width(6.dp))
                                        Text(
                                            String.format("%.1f", r), 
                                            color = Color(0xFFFFD700),
                                            style = MaterialTheme.typography.bodyLarge, 
                                            fontWeight = FontWeight.Bold
                                        )
                                    }
                                }
                            }
                            Box(
                                Modifier.background(Color.White.copy(0.15f), RoundedCornerShape(6.dp))
                                    .padding(horizontal = 12.dp, vertical = 4.dp)
                            ) {
                                val typeLabel = when(item.mediaType) {
                                    "movie" -> "Movie"
                                    "tv_show" -> "TV Show"
                                    "tv_episode" -> "Episode"
                                    "music_album" -> "Album"
                                    "song" -> "Song"
                                    "game" -> "Game"
                                    "book" -> "Book"
                                    else -> (item.mediaType ?: "unknown").replace("_", " ")
                                        .replaceFirstChar { it.uppercase() }
                                }
                                Text(
                                    typeLabel,
                                    color = Color.White.copy(0.9f), 
                                    style = MaterialTheme.typography.labelMedium
                                )
                            }
                            item.quality?.let { q ->
                                Box(
                                    Modifier.background(Color(0xFF1B5E20).copy(0.8f), RoundedCornerShape(6.dp))
                                        .padding(horizontal = 12.dp, vertical = 4.dp)
                                ) {
                                    Text(q.uppercase(), color = Color.White, style = MaterialTheme.typography.labelMedium)
                                }
                            }
                        }

                        Spacer(Modifier.height(24.dp))

                        // Action buttons with improved spacing and clear CTAs
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            // Primary Play button with enhanced styling
                            Button(
                                onClick = { onNavigateToPlayer(mediaId) },
                                modifier = Modifier
                                    .height(52.dp)
                                    .focusRequester(playButtonFocus)
                                    .focusable(),
                                scale = ButtonDefaults.scale(focusedScale = 1.08f),
                                glow = ButtonDefaults.glow(focusedGlow = Glow(elevation = 8.dp, elevationColor = MaterialTheme.colorScheme.primary)),
                                border = ButtonDefaults.border(
                                    focusedBorder = Border(
                                        androidx.compose.foundation.BorderStroke(2.dp, MaterialTheme.colorScheme.primary),
                                        shape = RoundedCornerShape(8.dp)
                                    )
                                )
                            ) {
                                Box(
                                    modifier = Modifier.fillMaxSize(),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Row(
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                                    ) {
                                        M3Icon(Icons.Default.PlayArrow, "Play", Modifier.size(28.dp))
                                        Text(
                                            "Play Now",
                                            style = MaterialTheme.typography.titleMedium,
                                            fontWeight = FontWeight.SemiBold
                                        )
                                    }
                                }
                            }
                            
                            // Favorite button with clear state indication
                            Button(
                                onClick = {
                                    val oldState = isFavorite
                                    isFavorite = !oldState
                                    coroutineScope.launch {
                                        try {
                                            val entityType = item.mediaType ?: "movie"
                                            container.mediaRepository.toggleFavorite(entityType, mediaId, oldState)
                                        } catch (e: Exception) {
                                            isFavorite = oldState
                                            android.util.Log.e("MediaDetail", "Favorite toggle failed", e)
                                        }
                                    }
                                },
                                modifier = Modifier.height(52.dp),
                                scale = ButtonDefaults.scale(focusedScale = 1.08f),
                                glow = ButtonDefaults.glow(focusedGlow = Glow(elevation = 8.dp, elevationColor = MaterialTheme.colorScheme.primary)),
                                border = ButtonDefaults.border(
                                    focusedBorder = Border(
                                        androidx.compose.foundation.BorderStroke(2.dp, MaterialTheme.colorScheme.primary),
                                        shape = RoundedCornerShape(8.dp)
                                    )
                                )
                            ) {
                                Box(
                                    modifier = Modifier.fillMaxSize(),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Row(
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                                    ) {
                                        M3Icon(
                                            if (isFavorite) Icons.Default.Favorite else Icons.Default.FavoriteBorder,
                                            contentDescription = if (isFavorite) "Remove from Favorites" else "Add to Favorites",
                                            modifier = Modifier.size(24.dp),
                                            tint = if (isFavorite) Color(0xFFFF6B6B) else Color.White
                                        )
                                        Text(
                                            if (isFavorite) "Favorited" else "Favorite",
                                            style = MaterialTheme.typography.titleMedium
                                        )
                                    }
                                }
                            }

                            // History button — opens full reproduction
                            // history dialog showing aggregate totals + all
                            // session rows.
                            Button(
                                onClick = { showHistory = true },
                                modifier = Modifier.height(52.dp),
                                scale = ButtonDefaults.scale(focusedScale = 1.08f),
                                glow = ButtonDefaults.glow(focusedGlow = Glow(elevation = 8.dp, elevationColor = MaterialTheme.colorScheme.primary)),
                                border = ButtonDefaults.border(
                                    focusedBorder = Border(
                                        androidx.compose.foundation.BorderStroke(2.dp, MaterialTheme.colorScheme.primary),
                                        shape = RoundedCornerShape(8.dp)
                                    )
                                )
                            ) {
                                Box(
                                    modifier = Modifier.fillMaxSize(),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Row(
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                                    ) {
                                        M3Icon(
                                            Icons.Default.History,
                                            contentDescription = "View reproduction history",
                                            modifier = Modifier.size(24.dp)
                                        )
                                        Text(
                                            "History",
                                            style = MaterialTheme.typography.titleMedium
                                        )
                                    }
                                }
                            }
                        }

                        Spacer(Modifier.height(32.dp))

                        // Synopsis section
                        item.description?.let { desc ->
                            if (desc.isNotBlank()) {
                                Text(
                                    "Synopsis",
                                    style = MaterialTheme.typography.titleLarge,
                                    color = Color.White,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(Modifier.height(12.dp))
                                Text(
                                    desc,
                                    style = MaterialTheme.typography.bodyLarge,
                                    color = Color.White.copy(0.85f),
                                    lineHeight = MaterialTheme.typography.bodyLarge.lineHeight * 1.2f
                                )
                                Spacer(Modifier.height(24.dp))
                            }
                        }

                        // File info section
                        item.fileSize?.let { size ->
                            if (size > 0) {
                                val s = when {
                                    size > 1073741824 -> String.format("%.1f GB", size.toFloat() / 1073741824)
                                    size > 1048576 -> String.format("%.1f MB", size.toFloat() / 1048576)
                                    else -> String.format("%.1f KB", size.toFloat() / 1024)
                                }
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    Text(
                                        "File Size:",
                                        color = Color.White.copy(0.6f),
                                        style = MaterialTheme.typography.bodyMedium
                                    )
                                    Text(
                                        s,
                                        color = Color.White.copy(0.9f),
                                        style = MaterialTheme.typography.bodyMedium,
                                        fontWeight = FontWeight.Medium
                                    )
                                }
                            }
                        }
                        Spacer(Modifier.height(48.dp))
                    }
                }
            }
        }
    }

    // Auto-focus Play button when content loads
    LaunchedEffect(mediaItem) {
        if (mediaItem != null) {
            kotlinx.coroutines.delay(300)
            try { playButtonFocus.requestFocus() } catch (_: Exception) {}
        }
    }

    if (showHistory) {
        mediaItem?.let { item ->
            HistoryDialog(
                mediaItemId = mediaId,
                mediaTitle = item.title,
                repository = container.playbackRepository,
                onDismiss = { showHistory = false }
            )
        }
    }
}
