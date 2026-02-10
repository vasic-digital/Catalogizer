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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.ui.components.MediaCard
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
    val scope = rememberCoroutineScope()
    val focusRequester = remember { FocusRequester() }
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
                    modifier = Modifier
                        .weight(1f)
                        .focusRequester(focusRequester)
                        .focusable(),
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
                    enabled = searchQuery.isNotBlank() && !isLoading
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
                    style = androidx.tv.material3.MaterialTheme.typography.bodyLarge
                )
                
                LazyColumn(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    items(searchResults) { mediaItem ->
                        MediaCard(
                            mediaItem = mediaItem,
                            onClick = { onNavigateToMediaDetail(mediaItem.id) },
                            onFocus = { /* Handle focus if needed */ },
                            modifier = Modifier.fillMaxWidth()
                        )
                    }
                }
            } else if (searchQuery.isNotBlank() && !isLoading) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .weight(1f),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "No results found for \"$searchQuery\"",
                        style = androidx.tv.material3.MaterialTheme.typography.bodyLarge
                    )
                }
            } else if (searchQuery.isBlank()) {
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
                            style = androidx.tv.material3.MaterialTheme.typography.bodyMedium
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
            } catch (e: Exception) {
                _error.value = "Search failed: ${e.message}"
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun clearResults() {
        _searchResults.value = emptyList()
        _searchQuery.value = ""
        _error.value = null
    }
}