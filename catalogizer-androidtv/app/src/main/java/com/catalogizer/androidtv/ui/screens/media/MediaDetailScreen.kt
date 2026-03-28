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
import coil.compose.AsyncImage
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.models.MediaItem
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

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
    val coroutineScope = rememberCoroutineScope()
    val scrollState = rememberScrollState()
    val playButtonFocus = remember { FocusRequester() }
    val config = LocalConfiguration.current
    val isCompact = config.screenWidthDp < 600

    // Fetch entity details from API
    LaunchedEffect(mediaId, retryCount) {
        isLoading = true
        try {
            val result = container.mediaRepository.getMediaById(mediaId).first()
            mediaItem = result
            isFavorite = result?.isFavorite ?: false
            error = if (result == null) "Media not found" else null
        } catch (e: Exception) {
            error = "Failed to load: ${e.message}"
        }
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
                    Spacer(Modifier.height(12.dp))
                    Text(
                        error ?: "An unknown error occurred",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
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
                        Button(onClick = onNavigateBack) {
                            Text("Back to Library")
                        }
                    }
                }
            }
            mediaItem != null -> {
                val item = mediaItem!!
                val coverUrl = item.thumbnailUrl?.let { url ->
                    if (url.startsWith("/")) container.getServerUrl().trimEnd('/') + url else url
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
                            .height(if (isCompact) 280.dp else 400.dp)
                    ) {
                        if (coverUrl != null) {
                            AsyncImage(
                                model = coverUrl,
                                contentDescription = item.title,
                                modifier = Modifier.fillMaxSize(),
                                contentScale = ContentScale.Crop
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
                        // Back button
                        Button(
                            onClick = onNavigateBack,
                            modifier = Modifier.padding(16.dp).align(Alignment.TopStart)
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

                    // Metadata content
                    Column(modifier = Modifier.padding(horizontal = if (isCompact) 24.dp else 48.dp)) {
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

                        Spacer(Modifier.height(8.dp))

                        // Year + Rating + Type badges
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            item.year?.let {
                                Text(it.toString(), color = Color.White.copy(0.7f), style = MaterialTheme.typography.bodyMedium)
                            }
                            item.rating?.let { r ->
                                if (r > 0) {
                                    Row(verticalAlignment = Alignment.CenterVertically) {
                                        M3Icon(Icons.Default.Star, "Rating", Modifier.size(16.dp), tint = Color(0xFFFFD700))
                                        Spacer(Modifier.width(4.dp))
                                        Text(String.format("%.1f", r), color = Color(0xFFFFD700),
                                            style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold)
                                    }
                                }
                            }
                            Box(
                                Modifier.background(Color.White.copy(0.15f), RoundedCornerShape(4.dp))
                                    .padding(horizontal = 8.dp, vertical = 2.dp)
                            ) {
                                Text((item.mediaType ?: "unknown").replace("_", " ").uppercase(),
                                    color = Color.White.copy(0.8f), style = MaterialTheme.typography.labelSmall)
                            }
                            item.quality?.let { q ->
                                Box(
                                    Modifier.background(Color(0xFF1B5E20).copy(0.8f), RoundedCornerShape(4.dp))
                                        .padding(horizontal = 8.dp, vertical = 2.dp)
                                ) {
                                    Text(q.uppercase(), color = Color.White, style = MaterialTheme.typography.labelSmall)
                                }
                            }
                        }

                        Spacer(Modifier.height(20.dp))

                        // Play + Back buttons (Play gets auto-focus for D-pad)
                        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                            Button(
                                onClick = { onNavigateToPlayer(mediaId) },
                                modifier = Modifier.height(48.dp).focusRequester(playButtonFocus).focusable()
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    M3Icon(Icons.Default.PlayArrow, "Play", Modifier.size(24.dp))
                                    Text("Play", style = MaterialTheme.typography.titleSmall)
                                }
                            }
                            Button(onClick = onNavigateBack, modifier = Modifier.height(48.dp)) {
                                Text("Back to Library", style = MaterialTheme.typography.titleSmall)
                            }
                            Button(
                                onClick = {
                                    val newState = !isFavorite
                                    isFavorite = newState
                                    coroutineScope.launch {
                                        try {
                                            container.mediaRepository.updateFavoriteStatus(mediaId, newState)
                                        } catch (e: Exception) {
                                            isFavorite = !newState // revert on failure
                                            android.util.Log.e("MediaDetail", "Favorite toggle failed", e)
                                        }
                                    }
                                },
                                modifier = Modifier.height(48.dp)
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    M3Icon(
                                        if (isFavorite) Icons.Default.Favorite else Icons.Default.FavoriteBorder,
                                        contentDescription = if (isFavorite) "Remove from Favorites" else "Add to Favorites",
                                        modifier = Modifier.size(24.dp)
                                    )
                                    Text("Favorite", style = MaterialTheme.typography.titleSmall)
                                }
                            }
                        }

                        Spacer(Modifier.height(24.dp))

                        // Synopsis
                        item.description?.let { desc ->
                            if (desc.isNotBlank()) {
                                Text("Synopsis", style = MaterialTheme.typography.titleMedium,
                                    color = Color.White, fontWeight = FontWeight.Bold)
                                Spacer(Modifier.height(8.dp))
                                Text(desc, style = MaterialTheme.typography.bodyMedium,
                                    color = Color.White.copy(0.8f))
                                Spacer(Modifier.height(24.dp))
                            }
                        }

                        // File info
                        item.fileSize?.let { size ->
                            if (size > 0) {
                                val s = when {
                                    size > 1073741824 -> String.format("%.1f GB", size.toFloat() / 1073741824)
                                    size > 1048576 -> String.format("%.1f MB", size.toFloat() / 1048576)
                                    else -> String.format("%.1f KB", size.toFloat() / 1024)
                                }
                                Text("File Size: $s", color = Color.White.copy(0.7f), style = MaterialTheme.typography.bodySmall)
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
}
