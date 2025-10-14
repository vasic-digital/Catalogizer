package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class MediaRepository(
    private val api: CatalogizerApi
) {

    fun searchMedia(request: MediaSearchRequest): Flow<List<MediaItem>> = flow {
        try {
            // TODO: Implement API call
            // For now, emit empty list
            emit(emptyList())
        } catch (e: Exception) {
            emit(emptyList())
        }
    }

    fun getMediaById(id: Long): Flow<MediaItem?> = flow {
        try {
            // TODO: Implement API call
            emit(null)
        } catch (e: Exception) {
            emit(null)
        }
    }

    suspend fun updateWatchProgress(mediaId: Long, progress: Double) {
        try {
            // TODO: Implement API call
        } catch (e: Exception) {
            // Handle error
        }
    }

    suspend fun updateFavoriteStatus(mediaId: Long, isFavorite: Boolean) {
        try {
            // TODO: Implement API call
        } catch (e: Exception) {
            // Handle error
        }
    }
}