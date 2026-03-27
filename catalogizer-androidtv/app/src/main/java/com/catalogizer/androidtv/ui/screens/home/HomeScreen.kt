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
                    CircularProgressIndicator()
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
                            text = uiState.error ?: "Unknown error",
                            style = MaterialTheme.typography.bodyMedium
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(
                            onClick = { viewModel.loadHomeData() }
                        ) {
                            Text("Retry")
                        }
                    }
                }
            }

            uiState.continueWatching.isEmpty() && uiState.recentlyAdded.isEmpty() &&
                uiState.movies.isEmpty() && uiState.tvShows.isEmpty() &&
                uiState.music.isEmpty() && uiState.documents.isEmpty() &&
                uiState.recentMovies.isEmpty() && uiState.recentMusicAlbums.isEmpty() &&
                uiState.recentTvShows.isEmpty() && uiState.recentGames.isEmpty() &&
                uiState.recentConcerts.isEmpty() && uiState.recentBooks.isEmpty() &&
                uiState.recentSoftware.isEmpty() -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "No content available",
                            style = MaterialTheme.typography.headlineSmall
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Scan your media collection to populate the catalog",
                            style = MaterialTheme.typography.bodyMedium
                        )
                        Spacer(modifier = Modifier.height(16.dp))
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
                    if (uiState.recentlyAdded.isNotEmpty())
                        add(ContentRailData("Recently Added", uiState.recentlyAdded))

                    // Recently Added by category
                    if (uiState.recentMovies.isNotEmpty())
                        add(ContentRailData("Recently Added Movies", uiState.recentMovies))
                    if (uiState.recentMusicAlbums.isNotEmpty())
                        add(ContentRailData("Recently Added Music Albums", uiState.recentMusicAlbums))
                    if (uiState.recentTvShows.isNotEmpty())
                        add(ContentRailData("Recently Added TV Shows", uiState.recentTvShows))
                    if (uiState.recentGames.isNotEmpty())
                        add(ContentRailData("Recently Added Games", uiState.recentGames))
                    if (uiState.recentConcerts.isNotEmpty())
                        add(ContentRailData("Recently Added Concerts", uiState.recentConcerts))
                    if (uiState.recentBooks.isNotEmpty())
                        add(ContentRailData("Recently Added Books", uiState.recentBooks))
                    if (uiState.recentSoftware.isNotEmpty())
                        add(ContentRailData("Recently Added Software", uiState.recentSoftware))

                    // Recently Played by category
                    if (uiState.playedMovies.isNotEmpty())
                        add(ContentRailData("Recently Played Movies", uiState.playedMovies, navigateToPlayer = true))
                    if (uiState.playedMusicAlbums.isNotEmpty())
                        add(ContentRailData("Recently Played Music Albums", uiState.playedMusicAlbums, navigateToPlayer = true))
                    if (uiState.playedTvShows.isNotEmpty())
                        add(ContentRailData("Recently Played TV Shows", uiState.playedTvShows, navigateToPlayer = true))
                    if (uiState.playedGames.isNotEmpty())
                        add(ContentRailData("Recently Played Games", uiState.playedGames, navigateToPlayer = true))
                    if (uiState.playedConcerts.isNotEmpty())
                        add(ContentRailData("Recently Played Concerts", uiState.playedConcerts, navigateToPlayer = true))

                    // Top-rated category rails (original rails)
                    if (uiState.movies.isNotEmpty())
                        add(ContentRailData("Top Rated Movies", uiState.movies))
                    if (uiState.tvShows.isNotEmpty())
                        add(ContentRailData("Top Rated TV Shows", uiState.tvShows))
                    if (uiState.music.isNotEmpty())
                        add(ContentRailData("Music", uiState.music))
                    if (uiState.documents.isNotEmpty())
                        add(ContentRailData("Documents", uiState.documents))
                }

                TvLazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(24.dp)
                ) {
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
                    modifier = Modifier.width(200.dp)
                )
            }
        }
    }
}