package com.catalogizer.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.repository.MediaRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class SearchViewModel(
    private val mediaRepository: MediaRepository
) : ViewModel() {

    private val _query = MutableStateFlow("")
    val query: StateFlow<String> = _query.asStateFlow()

    private val _searchResults = MutableStateFlow<List<MediaItem>>(emptyList())
    val searchResults: StateFlow<List<MediaItem>> = _searchResults.asStateFlow()

    private val _isSearching = MutableStateFlow(false)
    val isSearching: StateFlow<Boolean> = _isSearching.asStateFlow()

    init {
        // Setup search pipeline with debouncing
        _query
            .debounce(300) // 300ms debounce
            .distinctUntilChanged()
            .onEach { query ->
                if (query.isBlank()) {
                    _searchResults.value = emptyList()
                    _isSearching.value = false
                    return@onEach
                }
                
                _isSearching.value = true
                try {
                    val result = mediaRepository.getRecentMedia(50)
                    if (result.isSuccess) {
                        _searchResults.value = (result.data ?: emptyList()).filter {
                            it.title.contains(query, ignoreCase = true) ||
                                it.mediaType.contains(query, ignoreCase = true) ||
                                it.description?.contains(query, ignoreCase = true) == true
                        }
                    }
                } catch (e: Exception) {
                    _searchResults.value = emptyList()
                } finally {
                    _isSearching.value = false
                }
            }
            .catch { e ->
                _searchResults.value = emptyList()
                _isSearching.value = false
            }
            .launchIn(viewModelScope)
    }

    fun search(query: String) {
        _query.value = query
    }
}
