package com.catalogizer.androidtv.ui.screens.home

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.tv.material3.Text
import androidx.tv.material3.MaterialTheme
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.foundation.lazy.list.TvLazyColumn
import androidx.tv.foundation.lazy.list.items
import androidx.tv.material3.*
import androidx.compose.foundation.background
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.ui.components.MediaCarousel
import com.catalogizer.androidtv.ui.components.MediaCard
import com.catalogizer.androidtv.ui.components.TopBar
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel

/**
 * Internal data holder for building the home screen rail list dynamically.
 * @param navigateToPlayer true for "Continue Watching" / "Recently Played" rails
 *                         (clicking opens the player directly).
 */
private data class ContentRailData(
    val title: String,
    val items: List<MediaItem>,
    val navigateToPlayer: Boolean = false
)

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun HomeScreen(
    onNavigateToSearch: () -> Unit,
    onNavigateToSettings: () -> Unit,
    onNavigateToMediaDetail: (Long) -> Unit,
    onNavigateToPlayer: (Long) -> Unit,
    viewModel: HomeViewModel
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.loadHomeData()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 48.dp)
    ) {
        TopBar(
            title = "Catalogizer",
            onSearchClick = onNavigateToSearch,
            onSettingsClick = onNavigateToSettings,
            modifier = Modifier.padding(vertical = 24.dp)
        )

        when {
            uiState.isLoading -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = "Loading your media collection…",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }
                }
            }

            uiState.error != null -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "Error loading content",
                            style = MaterialTheme.typography.headlineSmall
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Please check your network connection and try again.",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = uiState.error ?: "Unknown error",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                        Spacer(modifier = Modifier.height(24.dp))
                        Button(
                            onClick = { viewModel.loadHomeData() }
                        ) {
                            Text("Retry")
                        }
                    }
                }
            }

            uiState.continueWatching.isEmpty() &&
                uiState.recentMovies.isEmpty() && uiState.recentTvShows.isEmpty() &&
                uiState.recentMusicAlbums.isEmpty() && uiState.recentGames.isEmpty() &&
                uiState.recentBooks.isEmpty() && uiState.recentComics.isEmpty() &&
                uiState.recentSoftware.isEmpty() && uiState.recentConcerts.isEmpty() &&
                uiState.recommended.isEmpty() && uiState.trending.isEmpty() &&
                uiState.topRatedMovies.isEmpty() && uiState.topRatedTvShows.isEmpty() &&
                uiState.topRatedMusic.isEmpty() && uiState.topRatedDocuments.isEmpty() -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "Your Library is Empty",
                            style = MaterialTheme.typography.headlineSmall
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            text = "Add storage roots and scan your media collection to get started.",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Use the web dashboard to configure storage and trigger scans.",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                        )
                        Spacer(modifier = Modifier.height(24.dp))
                        Button(
                            onClick = { viewModel.loadHomeData() }
                        ) {
                            Text("Refresh")
                        }
                    }
                }
            }

            else -> {
                // Build the list of content rails dynamically — only non-empty rails are shown
                val rails = buildList {
                    if (uiState.continueWatching.isNotEmpty())
                        add(ContentRailData("Continue Watching", uiState.continueWatching, navigateToPlayer = true))

                    if (uiState.recentMovies.isNotEmpty())
                        add(ContentRailData("Recently Added Movies", uiState.recentMovies))
                    if (uiState.recentTvShows.isNotEmpty())
                        add(ContentRailData("Recently Added TV Shows", uiState.recentTvShows))
                    if (uiState.recentMusicAlbums.isNotEmpty())
                        add(ContentRailData("Recently Added Music Albums", uiState.recentMusicAlbums))
                    if (uiState.recentGames.isNotEmpty())
                        add(ContentRailData("Recently Added Games", uiState.recentGames))
                    if (uiState.recentBooks.isNotEmpty())
                        add(ContentRailData("Recently Added Books", uiState.recentBooks))
                    if (uiState.recentComics.isNotEmpty())
                        add(ContentRailData("Recently Added Comics", uiState.recentComics))
                    if (uiState.recentSoftware.isNotEmpty())
                        add(ContentRailData("Recently Added Software", uiState.recentSoftware))
                    if (uiState.recentConcerts.isNotEmpty())
                        add(ContentRailData("Recently Added Concerts", uiState.recentConcerts))

                    if (uiState.recommended.isNotEmpty())
                        add(ContentRailData("Recommended for You", uiState.recommended))
                    if (uiState.trending.isNotEmpty())
                        add(ContentRailData("Trending", uiState.trending))

                    if (uiState.topRatedMovies.isNotEmpty())
                        add(ContentRailData("Top Rated Movies", uiState.topRatedMovies))
                    if (uiState.topRatedTvShows.isNotEmpty())
                        add(ContentRailData("Top Rated TV Shows", uiState.topRatedTvShows))
                    if (uiState.topRatedMusic.isNotEmpty())
                        add(ContentRailData("Top Rated Music", uiState.topRatedMusic))
                    if (uiState.topRatedDocuments.isNotEmpty())
                        add(ContentRailData("Top Rated Documents", uiState.topRatedDocuments))
                }

                TvLazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(24.dp)
                ) {
                    // Stats header
                    if (uiState.totalEntities > 0) {
                        item {
                            CatalogStatsHeader(
                                totalEntities = uiState.totalEntities,
                                statsByType = uiState.statsByType
                            )
                        }
                    }

                    items(rails.size) { index ->
                        val rail = rails[index]
                        MediaSection(
                            title = rail.title,
                            items = rail.items,
                            onItemClick = if (rail.navigateToPlayer) onNavigateToPlayer else onNavigateToMediaDetail,
                            onItemFocus = { /* Handle focus */ }
                        )
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun MediaSection(
    title: String,
    items: List<MediaItem>,
    onItemClick: (Long) -> Unit,
    onItemFocus: (MediaItem) -> Unit,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        Text(
            text = title,
            style = MaterialTheme.typography.headlineMedium,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(16.dp),
            contentPadding = PaddingValues(end = 48.dp)
        ) {
            items(items) { item ->
                MediaCard(
                    mediaItem = item,
                    onClick = { onItemClick(item.id) },
                    onFocus = { onItemFocus(item) },
                    modifier = Modifier.width(240.dp)
                )
            }
        }
    }
}

private val statsTypeLabels = mapOf(
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
    "comic" to "Comics"
)

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun CatalogStatsHeader(
    totalEntities: Int,
    statsByType: Map<String, Int>,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "Your Library — $totalEntities items",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )
        }

        Spacer(modifier = Modifier.height(12.dp))

        // Stats chips in a horizontal row
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            val nonZero = statsByType.filter { it.value > 0 }
                .entries.sortedByDescending { it.value }
            items(nonZero.size) { index ->
                val entry = nonZero[index]
                val label = statsTypeLabels[entry.key]
                    ?: entry.key.replace("_", " ")
                        .replaceFirstChar { it.uppercase() }
                Box(
                    modifier = Modifier
                        .background(
                            color = Color.White.copy(alpha = 0.1f),
                            shape = RoundedCornerShape(8.dp)
                        )
                        .padding(horizontal = 16.dp, vertical = 8.dp)
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Text(
                            text = "${entry.value}",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                            color = MaterialTheme.colorScheme.primary
                        )
                        Text(
                            text = label,
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.White.copy(alpha = 0.8f)
                        )
                    }
                }
            }
        }
    }
}