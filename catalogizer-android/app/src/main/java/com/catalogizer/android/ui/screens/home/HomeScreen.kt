package com.catalogizer.android.ui.screens.home

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.playback.PlaybackRepository
import com.catalogizer.android.data.playback.UiPlaybackProgress
import com.catalogizer.android.ui.components.HistoryDialog
import com.catalogizer.android.ui.components.ProgressBadge
import com.catalogizer.android.ui.viewmodel.HomeViewModel

/**
 * Enterprise-grade home screen with adaptive layout for phone and tablet.
 * Uses WindowSizeClass via screen width to adjust card sizes and grid columns.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    viewModel: HomeViewModel,
    onNavigateToSearch: () -> Unit,
    onNavigateToSettings: () -> Unit,
    onNavigateToMediaDetail: (Long) -> Unit
) {
    val recentMedia by viewModel.recentMedia.collectAsStateWithLifecycle()
    val favoriteMedia by viewModel.favoriteMedia.collectAsStateWithLifecycle()
    val progressById by viewModel.progressById.collectAsStateWithLifecycle()
    val isLoading by viewModel.isLoading.collectAsStateWithLifecycle()
    val error by viewModel.error.collectAsStateWithLifecycle()
    var historyTarget by remember { mutableStateOf<MediaItem?>(null) }
    val container = com.catalogizer.android.DependencyContainer.getInstance(
        androidx.compose.ui.platform.LocalContext.current
    )
    val config = LocalConfiguration.current
    val isTablet = config.screenWidthDp >= 600

    LaunchedEffect(Unit) {
        viewModel.loadHomeData()
    }

    Scaffold(
        topBar = {
            LargeTopAppBar(
                title = {
                    Column {
                        Text(
                            "Catalogizer",
                            fontWeight = FontWeight.Bold,
                        )
                        Text(
                            "Your media collection",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                },
                actions = {
                    IconButton(onClick = onNavigateToSearch) {
                        Icon(Icons.Default.Search, contentDescription = "Search")
                    }
                    IconButton(onClick = onNavigateToSettings) {
                        Icon(Icons.Default.Settings, contentDescription = "Settings")
                    }
                },
                colors = TopAppBarDefaults.largeTopAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface,
                ),
            )
        }
    ) { paddingValues ->
        if (isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues),
                contentAlignment = Alignment.Center
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    CircularProgressIndicator()
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        "Loading your media collection…",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(paddingValues),
                verticalArrangement = Arrangement.spacedBy(24.dp),
                contentPadding = PaddingValues(vertical = 16.dp)
            ) {
                error?.let { errorMessage ->
                    item {
                        Card(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 16.dp),
                            colors = CardDefaults.cardColors(
                                containerColor = MaterialTheme.colorScheme.errorContainer
                            ),
                            shape = RoundedCornerShape(16.dp),
                        ) {
                            Row(
                                modifier = Modifier.padding(16.dp),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Text(
                                    text = errorMessage,
                                    modifier = Modifier.weight(1f),
                                    color = MaterialTheme.colorScheme.onErrorContainer,
                                )
                                TextButton(onClick = { viewModel.loadHomeData() }) {
                                    Text("Retry")
                                }
                            }
                        }
                    }
                }

                if (recentMedia.isNotEmpty()) {
                    item {
                        MediaSection(
                            title = "Recently Added",
                            items = recentMedia,
                            progressById = progressById,
                            onItemClick = onNavigateToMediaDetail,
                            onItemLongPress = { item -> historyTarget = item },
                            isTablet = isTablet,
                        )
                    }
                }

                if (favoriteMedia.isNotEmpty()) {
                    item {
                        MediaSection(
                            title = "Popular",
                            items = favoriteMedia,
                            progressById = progressById,
                            onItemClick = onNavigateToMediaDetail,
                            onItemLongPress = { item -> historyTarget = item },
                            isTablet = isTablet,
                        )
                    }
                }

                if (recentMedia.isEmpty() && favoriteMedia.isEmpty() && error == null) {
                    item {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(48.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Text(
                                    text = "No media found",
                                    style = MaterialTheme.typography.titleMedium,
                                    fontWeight = FontWeight.SemiBold,
                                )
                                Spacer(modifier = Modifier.height(8.dp))
                                Text(
                                    text = "Your media library is empty. Configure storage sources to get started.",
                                    style = MaterialTheme.typography.bodyMedium,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                                Spacer(modifier = Modifier.height(24.dp))
                                FilledTonalButton(onClick = { viewModel.loadHomeData() }) {
                                    Text("Refresh")
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    historyTarget?.let { target ->
        HistoryDialog(
            mediaItemId = target.id,
            mediaTitle = target.title,
            repository = container.playbackRepository,
            onDismiss = { historyTarget = null },
        )
    }
}

@Composable
private fun MediaSection(
    title: String,
    items: List<MediaItem>,
    progressById: Map<Long, UiPlaybackProgress>,
    onItemClick: (Long) -> Unit,
    onItemLongPress: (MediaItem) -> Unit,
    isTablet: Boolean,
) {
    val cardWidth = if (isTablet) 180.dp else 140.dp
    val imageHeight = if (isTablet) 240.dp else 190.dp

    Column {
        Text(
            text = title,
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp)
        )

        LazyRow(
            contentPadding = PaddingValues(horizontal = 16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            items(items, key = { it.id }) { mediaItem ->
                MediaCard(
                    item = mediaItem,
                    progress = progressById[mediaItem.id],
                    onClick = { onItemClick(mediaItem.id) },
                    onLongClick = { onItemLongPress(mediaItem) },
                    cardWidth = cardWidth,
                    imageHeight = imageHeight,
                )
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun MediaCard(
    item: MediaItem,
    progress: UiPlaybackProgress?,
    onClick: () -> Unit,
    onLongClick: () -> Unit,
    cardWidth: androidx.compose.ui.unit.Dp,
    imageHeight: androidx.compose.ui.unit.Dp,
) {
    Card(
        modifier = Modifier
            .width(cardWidth)
            .combinedClickable(onClick = onClick, onLongClick = onLongClick),
        shape = RoundedCornerShape(12.dp),
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp),
    ) {
        Column {
            // Cover art with gradient overlay
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(imageHeight)
                    .clip(RoundedCornerShape(topStart = 12.dp, topEnd = 12.dp)),
                contentAlignment = Alignment.Center
            ) {
                val coverUrl = item.coverImage
                    ?: item.externalMetadata?.firstOrNull()?.posterUrl
                if (coverUrl != null) {
                    val container = com.catalogizer.android.DependencyContainer.getInstance(
                        androidx.compose.ui.platform.LocalContext.current
                    )
                    val resolvedUrl = when {
                        coverUrl.startsWith("/") ->
                            container.getServerUrl().trimEnd('/') + coverUrl
                        coverUrl.contains("image.tmdb.org") -> {
                            val encoded = java.net.URLEncoder.encode(coverUrl, "UTF-8")
                            container.getServerUrl().trimEnd('/') +
                                "/api/v1/image-proxy?url=$encoded"
                        }
                        else -> coverUrl
                    }
                    AsyncImage(
                        model = resolvedUrl,
                        contentDescription = item.title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    // Gradient placeholder with type initial
                    val gradient = mediaTypeGradient(item.mediaType)
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(Brush.verticalGradient(gradient)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = item.mediaType.take(1).uppercase(),
                            style = MaterialTheme.typography.headlineLarge,
                            color = Color.White.copy(alpha = 0.7f),
                            fontWeight = FontWeight.Bold,
                        )
                    }
                }

                // Bottom gradient for text readability
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(60.dp)
                        .align(Alignment.BottomCenter)
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(Color.Transparent, Color.Black.copy(alpha = 0.6f))
                            )
                        )
                )

                // Media type badge
                Surface(
                    modifier = Modifier
                        .align(Alignment.TopStart)
                        .padding(8.dp),
                    shape = RoundedCornerShape(6.dp),
                    color = Color.Black.copy(alpha = 0.6f),
                ) {
                    Text(
                        text = item.mediaType.replace("_", " "),
                        modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                        style = MaterialTheme.typography.labelSmall,
                        color = Color.White,
                        fontSize = 9.sp,
                    )
                }

                // Nexus cover-quality debug badge — only shown in debug
                // builds. Enabled by default when BuildConfig.DEBUG.
                com.catalogizer.android.ui.debug.CoverQualityBadge(
                    coverId = item.id,
                    enabled = true,
                    modifier = Modifier
                        .align(Alignment.TopEnd)
                        .padding(8.dp),
                )

                // Rating badge
                item.rating?.let { rating ->
                    Surface(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(8.dp),
                        shape = RoundedCornerShape(6.dp),
                        color = Color.Black.copy(alpha = 0.6f),
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(2.dp),
                        ) {
                            Icon(
                                Icons.Default.Star,
                                contentDescription = null,
                                modifier = Modifier.size(10.dp),
                                tint = Color(0xFFFFC107),
                            )
                            Text(
                                text = "%.1f".format(rating),
                                style = MaterialTheme.typography.labelSmall,
                                color = Color.White,
                                fontSize = 10.sp,
                            )
                        }
                    }
                }

                // Progress badge
                if (progress != null) {
                    ProgressBadge(
                        progress = progress,
                        modifier = Modifier.align(Alignment.BottomStart),
                    )
                }
            }

            // Card content
            Column(
                modifier = Modifier.padding(horizontal = 10.dp, vertical = 8.dp)
            ) {
                Text(
                    text = item.title,
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    lineHeight = 18.sp,
                )
                Spacer(modifier = Modifier.height(2.dp))
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    item.year?.let { year ->
                        Text(
                            text = year.toString(),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
            }
        }
    }
}

private fun mediaTypeGradient(type: String): List<Color> = when (type.lowercase()) {
    "movie" -> listOf(Color(0xFF1A237E), Color(0xFF283593))
    "tv_show", "tv_season", "tv_episode" -> listOf(Color(0xFF4A148C), Color(0xFF6A1B9A))
    "music_album", "song", "music" -> listOf(Color(0xFF1B5E20), Color(0xFF2E7D32))
    "game", "software" -> listOf(Color(0xFFE65100), Color(0xFFF57C00))
    "book", "comic", "ebook" -> listOf(Color(0xFF4E342E), Color(0xFF6D4C41))
    else -> listOf(Color(0xFF37474F), Color(0xFF455A64))
}
