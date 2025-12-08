package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaSearchResponse
import com.catalogizer.androidtv.data.models.PlaybackProgress
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.flow.first

class MediaRepository(private val context: Context, private val api: CatalogizerApi) {
    
    suspend fun searchMedia(request: MediaSearchRequest): Flow<List<MediaItem>> {
        try {
            val params = mutableMapOf<String, String>()
            request.query?.let { params["q"] = it }
            request.limit?.let { params["limit"] = it.toString() }
            request.offset?.let { params["offset"] = it.toString() }
            request.mediaType?.let { params["media_type"] = it }

            val response = api.searchMedia(params)
            if (response.isSuccessful) {
                val mediaItems = response.body() ?: emptyList()
                return flowOf(mediaItems)
            } else {
                return flowOf(emptyList())
            }
        } catch (e: Exception) {
            // Handle error and return empty list
            return flowOf(emptyList())
        }
    }

    suspend fun getMediaById(mediaId: Long): Flow<MediaItem?> {
        try {
            val response = api.getMediaById(mediaId)
            if (response.isSuccessful) {
                val mediaItem = response.body()
                return flowOf(mediaItem)
            } else {
                return flowOf(null)
            }
        } catch (e: Exception) {
            // Handle error and return null
            return flowOf(null)
        }
    }

    suspend fun updateWatchProgress(mediaId: Long, progress: Double) {
        try {
            val progressBody = mapOf("progress" to progress)
            val response = api.updateWatchProgress(mediaId, progressBody)
            if (!response.isSuccessful) {
                throw Exception("Failed to update watch progress: ${response.message()}")
            }
        } catch (e: Exception) {
            // Handle error
            throw e
        }
    }

    suspend fun updateFavoriteStatus(mediaId: Long, isFavorite: Boolean) {
        try {
            val favoriteBody = mapOf("favorite" to isFavorite)
            val response = api.updateFavoriteStatus(mediaId, favoriteBody)
            if (!response.isSuccessful) {
                throw Exception("Failed to update favorite status: ${response.message()}")
            }
        } catch (e: Exception) {
            // Handle error
            throw e
        }
    }
}