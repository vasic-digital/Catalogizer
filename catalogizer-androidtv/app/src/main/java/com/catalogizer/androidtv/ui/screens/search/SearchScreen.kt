@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.search

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
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
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.Search
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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.tv.material3.*
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.ui.components.CompactMediaCard
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

// Search history data store key
private const val SEARCH_HISTORY_KEY = "search_history"
private const val MAX_HISTORY_ITEMS = 10

/**
 * TV-optimized search screen with D-pad-navigable text field, search button,
 * search history, and suggestions. Displays results as [CompactMediaCard] items
 * in a scrollable list with loading overlay and empty-state messaging.
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
    val searchHistory by viewModel.searchHistory
    val suggestions by viewModel.suggestions
    
    val scope = rememberCoroutineScope()
    val focusRequester = remember { FocusRequester() }
    val searchButtonFocusRequester = remember { FocusRequester() }
    val historyFocusRequester = remember { FocusRequester() }
    val keyboardController = LocalSoftwareKeyboardController.current

    // Colors derived from TV MaterialTheme for the OutlinedTextField
    val tvColorScheme = MaterialTheme.colorScheme
    val focusedBorderColor = tvColorScheme.primary
    val unfocusedBorderColor = tvColorScheme.onSurface.copy(alpha = 0.6f)
    val cursorColor = tvColorScheme.primary
    val textColor = tvColorScheme.onSurface
    val placeholderColor = tvColorScheme.onSurface.copy(alpha = 0.6f)
    val containerColor = tvColorScheme.surface

    // Load search history on launch
    LaunchedEffect(Unit) {
        viewModel.loadSearchHistory()
        delay(100)
        focusRequester.requestFocus()
    }

    Box(modifier = Modifier.fillMaxSize()) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 48.dp, vertical = 24.dp)
        ) {
            // Search Header with improved spacing
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 24.dp),
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                OutlinedTextField(
                    value = searchQuery,
                    onValueChange = { newValue: String -> 
                        viewModel.updateSearchQuery(newValue)
                        viewModel.updateSuggestions(newValue)
                    },
                    placeholder = {
                        M3Text(
                            text = "Search movies, shows, music...",
                            color = placeholderColor,
                            fontSize = 16.sp
                        )
                    },
                    leadingIcon = {
                        M3Icon(
                            imageVector = Icons.Default.Search,
                            contentDescription = "Search icon",
                            tint = unfocusedBorderColor
                        )
                    },
                    trailingIcon = {
                        if (searchQuery.isNotEmpty()) {
                            M3IconButton(onClick = { viewModel.clearResults() }) {
                                M3Icon(
                                    imageVector = Icons.Default.Clear,
                                    contentDescription = "Clear search",
                                    tint = unfocusedBorderColor
                                )
                            }
                        }
                    },
                    modifier = Modifier
                        .weight(1f)
                        .height(64.dp)
                        .focusRequester(focusRequester)
                        .focusable()
                        .onKeyEvent { keyEvent ->
                            when {
                                keyEvent.type == KeyEventType.KeyDown && keyEvent.key == Key.DirectionRight -> {
                                    if (searchQuery.isNotBlank()) {
                                        searchButtonFocusRequester.requestFocus()
                                    }
                                    true
                                }
                                keyEvent.type == KeyEventType.KeyDown && keyEvent.key == Key.DirectionDown -> {
                                    if (searchHistory.isNotEmpty() && !hasSearched) {
                                        historyFocusRequester.requestFocus()
                                        true
                                    } else false
                                }
                                keyEvent.type == KeyEventType.KeyDown && keyEvent.key == Key.Enter -> {
                                    keyboardController?.hide()
                                    viewModel.search()
                                    true
                                }
                                else -> false
                            }
                        },
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search),
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
                
                // Search button with clear call-to-action
                Button(
                    onClick = {
                        keyboardController?.hide()
                        viewModel.search()
                    },
                    enabled = searchQuery.isNotBlank() && !isLoading,
                    modifier = Modifier
                        .height(64.dp)
                        .widthIn(min = 120.dp)
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
                        Text(
                            text = "Search",
                            fontSize = 16.sp,
                            fontWeight = FontWeight.Medium
                        )
                    }
                }
            }

            // Suggestions dropdown
            AnimatedVisibility(
                visible = suggestions.isNotEmpty() && !hasSearched,
                enter = fadeIn(),
                exit = fadeOut()
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 16.dp)
                ) {
                    Text(
                        text = "Suggestions",
                        style = MaterialTheme.typography.labelLarge,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                        modifier = Modifier.padding(bottom = 8.dp)
                    )
                    LazyRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        items(suggestions) { suggestion ->
                            SuggestionChip(
                                suggestion = suggestion,
                                onClick = {
                                    viewModel.updateSearchQuery(suggestion)
                                    viewModel.search()
                                }
                            )
                        }
                    }
                }
            }

            // Search History Section
            AnimatedVisibility(
                visible = searchHistory.isNotEmpty() && !hasSearched,
                enter = fadeIn(),
                exit = fadeOut()
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 24.dp)
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(bottom = 12.dp),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(
                            text = "Recent Searches",
                            style = MaterialTheme.typography.labelLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                        TextButton(
                            onClick = { viewModel.clearSearchHistory() }
                        ) {
                            Text(
                                text = "Clear All",
                                fontSize = 14.sp
                            )
                        }
                    }
                    
                    LazyRow(
                        horizontalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        items(searchHistory) { historyItem ->
                            HistoryChip(
                                query = historyItem,
                                onClick = {
                                    viewModel.updateSearchQuery(historyItem)
                                    viewModel.search()
                                },
                                modifier = if (historyItem == searchHistory.first()) {
                                    Modifier.focusRequester(historyFocusRequester)
                                } else Modifier
                            )
                        }
                    }
                }
            }

            // Error Message with improved visibility
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
                    Row(
                        modifier = Modifier.padding(16.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        M3Icon(
                            imageVector = Icons.Default.Search,
                            contentDescription = null,
                            tint = MaterialTheme.colorScheme.error
                        )
                        Text(
                            text = errorMessage,
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onErrorContainer
                        )
                    }
                }
            }

            // Search Results
            if (searchResults.isNotEmpty()) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "${searchResults.size} results found",
                        style = MaterialTheme.typography.labelLarge,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                    )
                    TextButton(onClick = { viewModel.clearResults() }) {
                        Text("Clear Results")
                    }
                }

                LazyColumn(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
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
                // Empty state with helpful messaging
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
                        M3Icon(
                            imageVector = Icons.Default.Search,
                            contentDescription = null,
                            modifier = Modifier.size(64.dp),
                            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f)
                        )
                        Text(
                            text = "No results found",
                            style = MaterialTheme.typography.headlineSmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                        )
                        Text(
                            text = "Try different keywords or check your spelling",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                        if (searchHistory.isNotEmpty()) {
                            TextButton(
                                onClick = { viewModel.clearResults() }
                            ) {
                                Text("Try a recent search")
                            }
                        }
                    }
                }
            } else if (searchQuery.isBlank() && !hasSearched) {
                // Initial state
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
                        M3Icon(
                            imageVector = Icons.Default.Search,
                            contentDescription = null,
                            modifier = Modifier.size(80.dp),
                            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.2f)
                        )
                        Text(
                            text = "Search Your Media",
                            style = MaterialTheme.typography.headlineMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.9f)
                        )
                        Text(
                            text = "Type above and press Search to find movies, shows, music, and more",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
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
                        modifier = Modifier.padding(32.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        CircularProgressIndicator(
                            color = tvColorScheme.primary,
                            modifier = Modifier.size(48.dp)
                        )
                        Text(
                            text = "Searching...",
                            style = MaterialTheme.typography.titleMedium
                        )
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun HistoryChip(
    query: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        onClick = onClick,
        modifier = modifier,
        scale = CardDefaults.scale(focusedScale = 1.05f),
        colors = CardDefaults.colors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
            focusedContainerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            M3Icon(
                imageVector = Icons.Default.History,
                contentDescription = null,
                modifier = Modifier.size(18.dp),
                tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
            )
            Text(
                text = query,
                style = MaterialTheme.typography.bodyMedium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun SuggestionChip(
    suggestion: String,
    onClick: () -> Unit
) {
    Button(
        onClick = onClick,
        scale = ButtonDefaults.scale(focusedScale = 1.05f)
    ) {
        Text(
            text = suggestion,
            style = MaterialTheme.typography.bodyMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis
        )
    }
}

/**
 * Enhanced ViewModel for media search with history and suggestions support.
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

    private val _searchHistory = mutableStateOf<List<String>>(emptyList())
    val searchHistory = _searchHistory

    private val _suggestions = mutableStateOf<List<String>>(emptyList())
    val suggestions = _suggestions

    // Mock suggestions based on query
    private val commonSuggestions = listOf(
        "Action Movies", "Comedy Shows", "Latest Music", 
        "Top Rated", "Recently Added", "Trending Now"
    )

    fun updateSearchQuery(query: String) {
        _searchQuery.value = query
        _error.value = null
        updateSuggestions(query)
    }

    fun updateSuggestions(query: String) {
        if (query.length >= 2) {
            _suggestions.value = commonSuggestions.filter { 
                it.contains(query, ignoreCase = true) 
            }.take(5)
        } else {
            _suggestions.value = emptyList()
        }
    }

    fun loadSearchHistory() {
        // In a real implementation, this would load from DataStore
        // For now, using empty list as default
        _searchHistory.value = emptyList()
    }

    fun clearSearchHistory() {
        _searchHistory.value = emptyList()
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
                val results = mediaRepository.searchEntities(request).first()
                _searchResults.value = results
                _hasSearched.value = true
                
                // Add to search history
                addToHistory(searchQuery.value)
            } catch (e: Exception) {
                _error.value = "Search failed: ${e.message}"
                _hasSearched.value = true
            } finally {
                _isLoading.value = false
            }
        }
    }

    private fun addToHistory(query: String) {
        val current = _searchHistory.value.toMutableList()
        current.remove(query) // Remove if exists to move to top
        current.add(0, query)
        _searchHistory.value = current.take(MAX_HISTORY_ITEMS)
    }

    fun clearResults() {
        _searchResults.value = emptyList()
        _searchQuery.value = ""
        _error.value = null
        _hasSearched.value = false
        _suggestions.value = emptyList()
    }
}
