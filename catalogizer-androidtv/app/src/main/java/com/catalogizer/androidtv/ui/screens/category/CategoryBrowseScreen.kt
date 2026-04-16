@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.screens.category

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon as M3Icon
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.ui.components.MediaCard
import kotlinx.coroutines.flow.first

private val typeDisplayNames = mapOf(
    "movie" to "Movies",
    "tv_show" to "TV Shows",
    "tv_season" to "Seasons",
    "tv_episode" to "Episodes",
    "music_artist" to "Artists",
    "music_album" to "Albums",
    "song" to "Songs",
    "game" to "Games",
    "software" to "Software",
    "book" to "Books",
    "comic" to "Comics",
    "concert" to "Concerts"
)

/**
 * Category browse screen showing a grid of media items filtered by type.
 * Clicking a card navigates to the detail screen (or player for playable types).
 */
@Composable
fun CategoryBrowseScreen(
    mediaType: String,
    onNavigateBack: () -> Unit,
    onNavigateToMediaDetail: (Long) -> Unit,
    onNavigateToPlayer: (Long) -> Unit
) {
    val container = DependencyContainer.getInstance(
        androidx.compose.ui.platform.LocalContext.current
    )
    var items by remember { mutableStateOf<List<MediaItem>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }
    var error by remember { mutableStateOf<String?>(null) }
    var retryTrigger by remember { mutableStateOf(0) }

    LaunchedEffect(mediaType, retryTrigger) {
        isLoading = true
        error = null
        try {
            val results = container.mediaRepository.browseEntities(
                type = mediaType,
                limit = 100,
                sortBy = "rating",
                sortOrder = "desc"
            ).first()
            items = results
        } catch (e: Exception) {
            error = e.message ?: "Failed to load category"
        } finally {
            isLoading = false
        }
    }

    val displayName = typeDisplayNames[mediaType]
        ?: mediaType.replace("_", " ").replaceFirstChar { it.uppercase() }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 48.dp, vertical = 24.dp)
    ) {
        // Header with back button
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 24.dp),
            horizontalArrangement = Arrangement.spacedBy(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Button(
                onClick = onNavigateBack,
                modifier = Modifier.height(44.dp),
                scale = ButtonDefaults.scale(focusedScale = 1.05f),
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

            Text(
                text = displayName,
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
        }

        when {
            isLoading -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
                }
            }
            error != null -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "Error loading $displayName",
                            style = MaterialTheme.typography.headlineSmall
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = error ?: "Unknown error",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(
                            onClick = { retryTrigger++ },
                            scale = ButtonDefaults.scale(focusedScale = 1.08f),
                            glow = ButtonDefaults.glow(focusedGlow = Glow(elevation = 6.dp, elevationColor = MaterialTheme.colorScheme.primary))
                        ) {
                            Box(
                                modifier = Modifier.fillMaxSize(),
                                contentAlignment = Alignment.Center
                            ) {
                                Text("Retry")
                            }
                        }
                    }
                }
            }
            items.isEmpty() -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "No $displayName found",
                            style = MaterialTheme.typography.headlineSmall
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "This category is empty in your library.",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                }
            }
            else -> {
                Text(
                    text = "${items.size} items",
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                    modifier = Modifier.padding(bottom = 16.dp)
                )

                LazyVerticalGrid(
                    columns = GridCells.Adaptive(minSize = 140.dp),
                    modifier = Modifier.fillMaxSize(),
                    horizontalArrangement = Arrangement.spacedBy(16.dp),
                    verticalArrangement = Arrangement.spacedBy(24.dp),
                    contentPadding = PaddingValues(bottom = 48.dp)
                ) {
                    items(items, key = { it.id }) { item ->
                        val isPlayable = item.mediaType in setOf(
                            "movie", "tv_show", "tv_episode", "concert", "song"
                        )
                        MediaCard(
                            mediaItem = item,
                            onClick = {
                                if (isPlayable) {
                                    onNavigateToPlayer(item.id)
                                } else {
                                    onNavigateToMediaDetail(item.id)
                                }
                            },
                            onFocus = { },
                            modifier = Modifier.width(140.dp)
                        )
                    }
                }
            }
        }
    }
}
