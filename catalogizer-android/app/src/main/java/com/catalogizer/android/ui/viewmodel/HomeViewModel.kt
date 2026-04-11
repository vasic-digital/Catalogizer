package com.catalogizer.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.playback.PlaybackRepository
import com.catalogizer.android.data.playback.UiPlaybackProgress
import com.catalogizer.android.data.repository.MediaRepository
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * Provides home screen data including recent and popular media lists.
 * Fetches data from [MediaRepository] and exposes loading/error state
 * as [StateFlow] for reactive UI composition.
 *
 * After the main lists land, spawns a background fetch against
 * [PlaybackRepository] to load per-item reproduction progress used by
 * ProgressBadge overlays on every card.
 */
class HomeViewModel(
    private val mediaRepository: MediaRepository,
    private val playbackRepository: PlaybackRepository? = null,
) : ViewModel() {

    private val _recentMedia = MutableStateFlow<List<MediaItem>>(emptyList())
    val recentMedia: StateFlow<List<MediaItem>> = _recentMedia.asStateFlow()

    private val _favoriteMedia = MutableStateFlow<List<MediaItem>>(emptyList())
    val favoriteMedia: StateFlow<List<MediaItem>> = _favoriteMedia.asStateFlow()

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()

    private val _progressById = MutableStateFlow<Map<Long, UiPlaybackProgress>>(emptyMap())
    val progressById: StateFlow<Map<Long, UiPlaybackProgress>> = _progressById.asStateFlow()

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

            launch { loadProgressForVisibleItems() }
        }
    }

    private suspend fun loadProgressForVisibleItems() {
        val repo = playbackRepository ?: return
        val allItems = (_recentMedia.value + _favoriteMedia.value).distinctBy { it.id }
        if (allItems.isEmpty()) return
        try {
            coroutineScope {
                val results = allItems.map { item ->
                    async {
                        try {
                            item.id to repo.getProgress(item.id)
                        } catch (_: Throwable) {
                            item.id to null
                        }
                    }
                }.map { it.await() }

                val map = results.mapNotNull { (id, p) -> p?.let { id to it } }.toMap()
                if (map.isNotEmpty()) _progressById.value = map
            }
        } catch (_: Exception) {
            // Best-effort — badges stay hidden on failure
        }
    }
}
