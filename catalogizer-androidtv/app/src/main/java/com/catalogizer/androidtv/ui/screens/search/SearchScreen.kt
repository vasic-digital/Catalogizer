@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.search

import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
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
import androidx.tv.material3.*
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.ui.components.CompactMediaCard
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

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

    LaunchedEffect(Unit) {
        delay(100) // Small delay to ensure composable is laid out
        focusRequester.requestFocus()
    }

    Box(modifier = Modifier.fillMaxSize()) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(16.dp)
        ) {
            // Search Header
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                @OptIn(ExperimentalTvMaterial3Api::class)
                TextField(
                    value = searchQuery,
                    onValueChange = { newValue: String -> viewModel.updateSearchQuery(newValue) },
                    label = { Text("Search Media") },
                    placeholder = {
                        Text(
                            text = "Search movies, music, books...",
                            color = Color.White.copy(alpha = 0.5f)
                        )
                    },
                    modifier = Modifier
                        .weight(1f)
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
                    singleLine = true
                )
                Button(
                    onClick = {
                        keyboardController?.hide()
                        viewModel.search()
                    },
                    enabled = searchQuery.isNotBlank() && !isLoading,
                    modifier = Modifier
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
                            strokeWidth = 2.dp
                        )
                    } else {
                        Text("Search")
                    }
                }
            }

            // Error Message
            error?.let { errorMessage ->
                Surface(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 16.dp),
                    shape = androidx.tv.material3.MaterialTheme.shapes.medium,
                    color = androidx.tv.material3.MaterialTheme.colorScheme.errorContainer,
                    onClick = {} // Empty onClick for compatibility
                ) {
                    Text(
                        text = errorMessage,
                        modifier = Modifier.padding(16.dp),
                        style = androidx.tv.material3.MaterialTheme.typography.bodyMedium
                    )
                }
            }

            // Search Results
            if (searchResults.isNotEmpty()) {
                Text(
                    text = "${searchResults.size} results found",
                    modifier = Modifier.padding(bottom = 16.dp),
                    style = androidx.tv.material3.MaterialTheme.typography.bodyLarge,
                    color = androidx.tv.material3.MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
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
                            style = androidx.tv.material3.MaterialTheme.typography.bodyLarge,
                            color = androidx.tv.material3.MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                        )
                        Text(
                            text = "Try a different title, keyword, or check your spelling",
                            style = androidx.tv.material3.MaterialTheme.typography.bodyMedium,
                            color = androidx.tv.material3.MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
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
                            style = androidx.tv.material3.MaterialTheme.typography.headlineMedium
                        )
                        Text(
                            text = "Enter a title, actor, or keyword to find media",
                            style = androidx.tv.material3.MaterialTheme.typography.bodyMedium,
                            color = androidx.tv.material3.MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
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
                    shape = androidx.tv.material3.MaterialTheme.shapes.medium,
                    color = androidx.tv.material3.MaterialTheme.colorScheme.surface,
                    onClick = {}
                ) {
                    Column(
                        modifier = Modifier.padding(24.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        CircularProgressIndicator()
                        Text("Searching...")
                    }
                }
            }
        }
    }
}

// ViewModel for search functionality with repository integration
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
                val results = mediaRepository.searchMedia(request).first()
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