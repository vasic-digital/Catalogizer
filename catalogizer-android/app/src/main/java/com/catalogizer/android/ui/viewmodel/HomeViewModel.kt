package com.catalogizer.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.repository.MediaRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class HomeViewModel(
    private val mediaRepository: MediaRepository
) : ViewModel() {

    private val _recentMedia = MutableStateFlow<List<MediaItem>>(emptyList())
    val recentMedia: StateFlow<List<MediaItem>> = _recentMedia.asStateFlow()

    private val _favoriteMedia = MutableStateFlow<List<MediaItem>>(emptyList())
    val favoriteMedia: StateFlow<List<MediaItem>> = _favoriteMedia.asStateFlow()

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()

    fun loadHomeData() {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            try {
                val recentResult = mediaRepository.getRecentMedia(20)
                if (recentResult.isSuccess) {
                    _recentMedia.value = recentResult.data ?: emptyList()
                }

                val favoritesResult = mediaRepository.getPopularMedia(20)
                if (favoritesResult.isSuccess) {
                    _favoriteMedia.value = favoritesResult.data ?: emptyList()
                }
            } catch (e: Exception) {
                _error.value = e.message ?: "Failed to load media"
            } finally {
                _isLoading.value = false
            }
        }
    }
}
