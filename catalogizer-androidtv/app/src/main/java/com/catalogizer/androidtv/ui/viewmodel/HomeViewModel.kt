package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.repository.MediaRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

data class HomeUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val continueWatching: List<MediaItem> = emptyList(),
    val recentlyAdded: List<MediaItem> = emptyList(),
    val movies: List<MediaItem> = emptyList(),
    val tvShows: List<MediaItem> = emptyList(),
    val music: List<MediaItem> = emptyList(),
    val documents: List<MediaItem> = emptyList(),
    val featuredItem: MediaItem? = null
)

class HomeViewModel(
    private val mediaRepository: MediaRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState.asStateFlow()

    fun loadHomeData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                // Load different content sections in parallel
                val continueWatching = loadContinueWatching()
                val recentlyAdded = loadRecentlyAdded()
                val movies = loadMovies()
                val tvShows = loadTVShows()
                val music = loadMusic()
                val documents = loadDocuments()

                // Set featured item from recently added or continue watching
                val featuredItem = continueWatching.firstOrNull() ?: recentlyAdded.firstOrNull()

                _uiState.update {
                    it.copy(
                        isLoading = false,
                        continueWatching = continueWatching,
                        recentlyAdded = recentlyAdded,
                        movies = movies,
                        tvShows = tvShows,
                        music = music,
                        documents = documents,
                        featuredItem = featuredItem
                    )
                }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = e.message ?: "Failed to load content"
                    )
                }
            }
        }
    }

    private suspend fun loadContinueWatching(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    sortBy = "last_watched",
                    sortOrder = "desc",
                    limit = 10
                )
            ).first().filter { it.hasWatchProgress && !it.isCompleted }
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadRecentlyAdded(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    sortBy = "created_at",
                    sortOrder = "desc",
                    limit = 20
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadMovies(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    mediaType = MediaType.MOVIE.value,
                    sortBy = "rating",
                    sortOrder = "desc",
                    limit = 20
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadTVShows(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    mediaType = MediaType.TV_SHOW.value,
                    sortBy = "rating",
                    sortOrder = "desc",
                    limit = 20
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadMusic(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    mediaType = MediaType.MUSIC.value,
                    sortBy = "created_at",
                    sortOrder = "desc",
                    limit = 20
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadDocuments(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    mediaType = MediaType.EBOOK.value,
                    sortBy = "created_at",
                    sortOrder = "desc",
                    limit = 20
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    fun refreshContent() {
        loadHomeData()
    }

    fun markAsWatched(mediaId: Long) {
        viewModelScope.launch {
            try {
                mediaRepository.updateWatchProgress(mediaId, 1.0)
                // Refresh continue watching section
                loadHomeData()
            } catch (e: Exception) {
                // Handle error
            }
        }
    }

    fun updateWatchProgress(mediaId: Long, progress: Double) {
        viewModelScope.launch {
            try {
                mediaRepository.updateWatchProgress(mediaId, progress)
            } catch (e: Exception) {
                // Handle error
            }
        }
    }

    fun toggleFavorite(mediaId: Long) {
        viewModelScope.launch {
            try {
                val mediaItem = mediaRepository.getMediaById(mediaId).first()
                mediaItem?.let {
                    mediaRepository.updateFavoriteStatus(mediaId, !it.isFavorite)
                    loadHomeData()
                }
            } catch (e: Exception) {
                // Handle error
            }
        }
    }
}