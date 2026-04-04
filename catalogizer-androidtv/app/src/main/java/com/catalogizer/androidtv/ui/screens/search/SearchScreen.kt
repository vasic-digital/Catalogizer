@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.search

import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Search
// material3-only composables (no TV equivalent)
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Icon as M3Icon
import androidx.compose.material3.Text as M3Text
import androidx.compose.material3.IconButton as M3IconButton
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.input.key.KeyEventType
import androidx.compose.ui.input.key.key
import androidx.compose.ui.input.key.onKeyEvent
import androidx.compose.ui.input.key.type
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
// TV material3 for everything else
import androidx.tv.material3.*
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.ui.components.CompactMediaCard
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * TV-optimized search screen with D-pad-navigable text field and search button.
 * Displays results as [CompactMediaCard] items in a scrollable list with
 * loading overlay and empty-state messaging.
 */
@Composable
fun SearchScreen(
    viewModel: SearchViewModel,
    onNavigateBack: () -> Unit,
    onNavigateToMediaDetail: (Long) -> Unit
) {
    val searchQuery by viewModel.searchQuery
    val searchResults by viewModel.searchResults
    val isLoading by viewModel.isLoading
    val error by viewModel.error
    val hasSearched by viewModel.hasSearched
    val scope = rememberCoroutineScope()
    val focusRequester = remember { FocusRequester() }
    val searchButtonFocusRequester = remember { FocusRequester() }
    val keyboardController = LocalSoftwareKeyboardController.current

    // Colors derived from TV MaterialTheme for the OutlinedTextField
    val tvColorScheme = MaterialTheme.colorScheme
    val focusedBorderColor = tvColorScheme.primary
    val unfocusedBorderColor = tvColorScheme.onSurface.copy(alpha = 0.5f)
    val cursorColor = tvColorScheme.primary
    val textColor = tvColorScheme.onSurface
    val placeholderColor = tvColorScheme.onSurface.copy(alpha = 0.5f)
    val containerColor = tvColorScheme.surface

    LaunchedEffect(Unit) {
        delay(100) // Small delay to ensure composable is laid out
        focusRequester.requestFocus()
    }

    Box(modifier = Modifier.fillMaxSize()) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp)
        ) {
            // Search Header
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                OutlinedTextField(
                    value = searchQuery,
                    onValueChange = { newValue: String -> viewModel.updateSearchQuery(newValue) },
                    placeholder = {
                        M3Text(
                            text = "Search movies, shows, music, games, books...",
                            color = placeholderColor
                        )
                    },
                    leadingIcon = {
                        M3Icon(
                            imageVector = Icons.Default.Search,
                            contentDescription = "Search",
                            tint = unfocusedBorderColor
                        )
                    },
                    trailingIcon = {
                        if (searchQuery.isNotEmpty()) {
                            M3IconButton(onClick = { viewModel.clearResults() }) {
                                M3Icon(
                                    imageVector = Icons.Default.Clear,
                                    contentDescription = "Clear",
                                    tint = unfocusedBorderColor
                                )
                            }
                        }
                    },
                    modifier = Modifier
                        .weight(1f)
                        .height(56.dp)
                        .focusRequester(focusRequester)
                        .focusable()
                        .onKeyEvent { keyEvent ->
                            if (keyEvent.type == KeyEventType.KeyDown && keyEvent.key == Key.DirectionRight) {
                                searchButtonFocusRequester.requestFocus()
                                true
                            } else if (keyEvent.type == KeyEventType.KeyDown && keyEvent.key == Key.Enter) {
                                keyboardController?.hide()
                                viewModel.search()
                                true
                            } else {
                                false
                            }
                        },
                    keyboardOptions = KeyboardOptions(
                        imeAction = ImeAction.Search
                    ),
                    keyboardActions = KeyboardActions(
                        onSearch = {
                            keyboardController?.hide()
                            viewModel.search()
                        }
                    ),
                    singleLine = true,
                    shape = RoundedCornerShape(12.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedTextColor = textColor,
                        unfocusedTextColor = textColor,
                        focusedBorderColor = focusedBorderColor,
                        unfocusedBorderColor = unfocusedBorderColor,
                        cursorColor = cursorColor,
                        focusedContainerColor = containerColor,
                        unfocusedContainerColor = containerColor,
                        focusedLeadingIconColor = focusedBorderColor,
                        unfocusedLeadingIconColor = unfocusedBorderColor,
                        focusedTrailingIconColor = focusedBorderColor,
                        unfocusedTrailingIconColor = unfocusedBorderColor
                    )
                )
                Button(
                    onClick = {
                        keyboardController?.hide()
                        viewModel.search()
                    },
                    enabled = searchQuery.isNotBlank() && !isLoading,
                    modifier = Modifier
                        .height(56.dp)
                        .focusRequester(searchButtonFocusRequester)
                        .onKeyEvent { keyEvent ->
                            if (keyEvent.type == KeyEventType.KeyDown && keyEvent.key == Key.DirectionLeft) {
                                focusRequester.requestFocus()
                                true
                            } else {
                                false
                            }
                        }
                ) {
                    if (isLoading) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(24.dp),
                            strokeWidth = 2.dp,
                            color = tvColorScheme.onPrimary
                        )
                    } else {
                        Text("Search")
                    }
                }
            }

            // Error Message
            error?.let { errorMessage ->
                Surface(
                    onClick = {},
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 16.dp),
                    shape = ClickableSurfaceDefaults.shape(),
                    colors = ClickableSurfaceDefaults.colors(
                        containerColor = MaterialTheme.colorScheme.errorContainer
                    )
                ) {
                    Text(
                        text = errorMessage,
                        modifier = Modifier.padding(16.dp),
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }

            // Search Results
            if (searchResults.isNotEmpty()) {
                Text(
                    text = "${searchResults.size} results found",
                    modifier = Modifier.padding(bottom = 16.dp),
                    style = MaterialTheme.typography.labelLarge,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                )

                LazyColumn(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(searchResults) { mediaItem ->
                        CompactMediaCard(
                            mediaItem = mediaItem,
                            onClick = { onNavigateToMediaDetail(mediaItem.id) },
                            modifier = Modifier.fillMaxWidth()
                        )
                    }
                }
            } else if (hasSearched && searchQuery.isNotBlank() && !isLoading) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        Text(
                            text = "No results found for \"$searchQuery\"",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                        )
                        Text(
                            text = "Try a different title, keyword, or check your spelling",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }
                }
            } else if (searchQuery.isBlank() && !hasSearched) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f),
                    contentAlignment = Alignment.Center
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        Text(
                            text = "Search for Media",
                            style = MaterialTheme.typography.headlineMedium
                        )
                        Text(
                            text = "Type a title, genre, or keyword above and press Search",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }
                }
            }
        }

        // Loading Overlay
        if (isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black.copy(alpha = 0.5f)),
                contentAlignment = Alignment.Center
            ) {
                Surface(
                    onClick = {},
                    shape = ClickableSurfaceDefaults.shape(),
                    colors = ClickableSurfaceDefaults.colors(
                        containerColor = MaterialTheme.colorScheme.surface
                    )
                ) {
                    Column(
                        modifier = Modifier.padding(24.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        CircularProgressIndicator(
                            color = tvColorScheme.primary
                        )
                        Text("Searching...")
                    }
                }
            }
        }
    }
}

/**
 * ViewModel for media search with [MediaRepository] integration.
 * Manages query state, search execution, loading/error indicators, and result clearing.
 */
class SearchViewModel(
    private val mediaRepository: MediaRepository
) : androidx.lifecycle.ViewModel() {
    private val _searchQuery = mutableStateOf("")
    val searchQuery = _searchQuery

    private val _searchResults = mutableStateOf<List<MediaItem>>(emptyList())
    val searchResults = _searchResults

    private val _isLoading = mutableStateOf(false)
    val isLoading = _isLoading

    private val _error = mutableStateOf<String?>(null)
    val error = _error

    private val _hasSearched = mutableStateOf(false)
    val hasSearched = _hasSearched

    fun updateSearchQuery(query: String) {
        _searchQuery.value = query
        _error.value = null
    }

    fun search() {
        if (searchQuery.value.isBlank()) return

        _isLoading.value = true
        _error.value = null

        viewModelScope.launch {
            try {
                val request = MediaSearchRequest(
                    query = searchQuery.value,
                    limit = 50
                )
                // Search entities (titles) instead of raw files.
                // The entity endpoint returns aggregated media items
                // with titles, covers, and metadata — much more
                // useful for user-facing search than file names.
                val results = mediaRepository.searchEntities(request).first()
                _searchResults.value = results
                _hasSearched.value = true
            } catch (e: Exception) {
                _error.value = "Search failed: ${e.message}"
                _hasSearched.value = true
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun clearResults() {
        _searchResults.value = emptyList()
        _searchQuery.value = ""
        _error.value = null
        _hasSearched.value = false
    }
}
