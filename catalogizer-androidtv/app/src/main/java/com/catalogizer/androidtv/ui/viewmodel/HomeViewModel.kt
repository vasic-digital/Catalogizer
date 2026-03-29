package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.repository.MediaRepository
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

/**
 * Represents a single content rail on the home screen.
 * Each rail has a title and a list of media items.
 */
data class ContentRail(
    val title: String,
    val items: List<MediaItem> = emptyList()
)

data class HomeUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val continueWatching: List<MediaItem> = emptyList(),
    val recentlyAdded: List<MediaItem> = emptyList(),
    val movies: List<MediaItem> = emptyList(),
    val tvShows: List<MediaItem> = emptyList(),
    val music: List<MediaItem> = emptyList(),
    val documents: List<MediaItem> = emptyList(),
    val featuredItem: MediaItem? = null,
    // Category-specific rails: recently added
    val recentMovies: List<MediaItem> = emptyList(),
    val recentMusicAlbums: List<MediaItem> = emptyList(),
    val recentTvShows: List<MediaItem> = emptyList(),
    val recentGames: List<MediaItem> = emptyList(),
    val recentConcerts: List<MediaItem> = emptyList(),
    val recentBooks: List<MediaItem> = emptyList(),
    val recentSoftware: List<MediaItem> = emptyList(),
    // Recently played rails
    val playedMovies: List<MediaItem> = emptyList(),
    val playedMusicAlbums: List<MediaItem> = emptyList(),
    val playedTvShows: List<MediaItem> = emptyList(),
    val playedGames: List<MediaItem> = emptyList(),
    val playedConcerts: List<MediaItem> = emptyList(),
    // Catalog stats
    val totalEntities: Int = 0,
    val statsByType: Map<String, Int> = emptyMap()
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

                // Category-specific "Recently Added" rails
                val recentMovies = loadRecentByType("movie")
                val recentMusicAlbums = loadRecentByType("music_album")
                val recentTvShows = loadRecentByType("tv_show")
                val recentGames = loadRecentByType("game")
                val recentConcerts = loadRecentByType("concert")
                val recentBooks = loadRecentByType("book")
                val recentSoftware = loadRecentByType("software")

                // "Recently Played" rails
                val playedMovies = loadRecentlyPlayedByType("movie")
                val playedMusicAlbums = loadRecentlyPlayedByType("music_album")
                val playedTvShows = loadRecentlyPlayedByType("tv_show")
                val playedGames = loadRecentlyPlayedByType("game")
                val playedConcerts = loadRecentlyPlayedByType("concert")

                // Set featured item from recently added or continue watching
                val featuredItem = continueWatching.firstOrNull() ?: recentlyAdded.firstOrNull()

                // Load catalog stats
                val stats = loadStats()

                _uiState.update {
                    it.copy(
                        isLoading = false,
                        continueWatching = continueWatching,
                        recentlyAdded = recentlyAdded,
                        movies = movies,
                        tvShows = tvShows,
                        music = music,
                        documents = documents,
                        featuredItem = featuredItem,
                        recentMovies = recentMovies,
                        recentMusicAlbums = recentMusicAlbums,
                        recentTvShows = recentTvShows,
                        recentGames = recentGames,
                        recentConcerts = recentConcerts,
                        recentBooks = recentBooks,
                        recentSoftware = recentSoftware,
                        playedMovies = playedMovies,
                        playedMusicAlbums = playedMusicAlbums,
                        playedTvShows = playedTvShows,
                        playedGames = playedGames,
                        playedConcerts = playedConcerts,
                        totalEntities = stats.first,
                        statsByType = stats.second
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
                    sortBy = "created",
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
                    sortBy = "created",
                    sortOrder = "desc",
                    limit = 50
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadMovies(): List<MediaItem> {
        return try {
            // Use entity browse to get TMDB poster URLs
            mediaRepository.browseEntities("movie", limit = 50, sortBy = "rating", sortOrder = "desc").first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadTVShows(): List<MediaItem> {
        return try {
            mediaRepository.browseEntities("tv_show", limit = 50, sortBy = "rating", sortOrder = "desc").first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    private suspend fun loadMusic(): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    mediaType = MediaType.MUSIC.value,
                    sortBy = "created",
                    sortOrder = "desc",
                    limit = 50
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
                    sortBy = "created",
                    sortOrder = "desc",
                    limit = 50
                )
            ).first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    /**
     * Load recently added items for a specific entity type via the browse endpoint.
     */
    private suspend fun loadRecentByType(type: String): List<MediaItem> {
        return try {
            mediaRepository.browseEntities(type, limit = 50, sortBy = "created", sortOrder = "desc").first()
        } catch (e: Exception) {
            emptyList()
        }
    }

    /**
     * Load recently played items for a specific entity type.
     * Uses the search endpoint with last_played sort — items with watch progress.
     */
    private suspend fun loadRecentlyPlayedByType(type: String): List<MediaItem> {
        return try {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    mediaType = type,
                    sortBy = "updated_at",
                    sortOrder = "desc",
                    limit = 50
                )
            ).first().filter { it.hasWatchProgress }
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

    private suspend fun loadStats(): Pair<Int, Map<String, Int>> {
        return try {
            val response = mediaRepository.getEntityStats()
            Pair(response.first, response.second)
        } catch (_: Exception) {
            Pair(0, emptyMap())
        }
    }

    fun toggleFavorite(mediaId: Long) {
        viewModelScope.launch {
            try {
                val mediaItem = mediaRepository.getMediaById(mediaId).first()
                mediaItem?.let {
                    val entityType = it.mediaType ?: "movie"
                    val isFav = mediaRepository.checkFavorite(
                        entityType, mediaId
                    )
                    mediaRepository.toggleFavorite(
                        entityType, mediaId, isFav
                    )
                    loadHomeData()
                }
            } catch (_: Exception) {
                // Handle error
            }
        }
    }
}