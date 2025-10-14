package com.catalogizer.androidtv.ui.screens.home

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.platform.LocalContext
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
    val searchFocusRequester = remember { FocusRequester() }

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

            else -> {
                TvLazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(24.dp)
                ) {
                    // Continue Watching Section
                    if (uiState.continueWatching.isNotEmpty()) {
                        item {
                            MediaSection(
                                title = "Continue Watching",
                                items = uiState.continueWatching,
                                onItemClick = onNavigateToPlayer,
                                onItemFocus = { /* Handle focus */ }
                            )
                        }
                    }

                    // Recently Added Section
                    if (uiState.recentlyAdded.isNotEmpty()) {
                        item {
                            MediaSection(
                                title = "Recently Added",
                                items = uiState.recentlyAdded,
                                onItemClick = onNavigateToMediaDetail,
                                onItemFocus = { /* Handle focus */ }
                            )
                        }
                    }

                    // Movies Section
                    if (uiState.movies.isNotEmpty()) {
                        item {
                            MediaSection(
                                title = "Movies",
                                items = uiState.movies,
                                onItemClick = onNavigateToMediaDetail,
                                onItemFocus = { /* Handle focus */ }
                            )
                        }
                    }

                    // TV Shows Section
                    if (uiState.tvShows.isNotEmpty()) {
                        item {
                            MediaSection(
                                title = "TV Shows",
                                items = uiState.tvShows,
                                onItemClick = onNavigateToMediaDetail,
                                onItemFocus = { /* Handle focus */ }
                            )
                        }
                    }

                    // Music Section
                    if (uiState.music.isNotEmpty()) {
                        item {
                            MediaSection(
                                title = "Music",
                                items = uiState.music,
                                onItemClick = onNavigateToMediaDetail,
                                onItemFocus = { /* Handle focus */ }
                            )
                        }
                    }

                    // Documents Section
                    if (uiState.documents.isNotEmpty()) {
                        item {
                            MediaSection(
                                title = "Documents",
                                items = uiState.documents,
                                onItemClick = onNavigateToMediaDetail,
                                onItemFocus = { /* Handle focus */ }
                            )
                        }
                    }
                }
            }
        }
    }
}

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