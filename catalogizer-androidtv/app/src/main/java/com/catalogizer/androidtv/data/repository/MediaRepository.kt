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

    /**
     * Browse entities by type (movie, tv_show, etc.) — includes external metadata with cover URLs.
     */
    suspend fun browseEntities(type: String, limit: Int = 20, sortBy: String = "title", sortOrder: String = "asc"): Flow<List<MediaItem>> {
        try {
            val params = mapOf("limit" to limit.toString(), "sort_by" to sortBy, "sort_order" to sortOrder)
            val response = api.browseEntities(type, params)
            if (response.isSuccessful) {
                val items = response.body()?.items ?: emptyList()
                return flowOf(items)
            } else {
                android.util.Log.w("MediaRepo", "Browse entities failed: ${response.code()}")
                return flowOf(emptyList())
            }
        } catch (e: Exception) {
            android.util.Log.w("MediaRepo", "Browse entities error: ${e.message}")
            return flowOf(emptyList())
        }
    }

    suspend fun searchMedia(request: MediaSearchRequest): Flow<List<MediaItem>> {
        try {
            val params = mutableMapOf<String, String>()
            request.query?.let { params["query"] = it }
            request.limit?.let { params["limit"] = it.toString() }
            request.offset?.let { params["offset"] = it.toString() }
            request.mediaType?.let { params["media_type"] = it }
            request.sortBy?.let { params["sort_by"] = it }
            request.sortOrder?.let { params["sort_order"] = it }

            val response = api.searchMedia(params)
            if (response.isSuccessful) {
                val searchResponse = response.body()
                val mediaItems = searchResponse?.items ?: emptyList()
                return flowOf(mediaItems)
            } else {
                android.util.Log.w("MediaRepo", "Search failed: ${response.code()} ${response.message()}")
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