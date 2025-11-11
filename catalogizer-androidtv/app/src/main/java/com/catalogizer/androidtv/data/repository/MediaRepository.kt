package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class MediaRepository(
    private val api: CatalogizerApi
) {

    fun searchMedia(request: MediaSearchRequest): Flow<List<MediaItem>> = flow {
        try {
            // Convert search request to query parameters
            val params = buildMap<String, String> {
                request.query?.let { put("query", it) }
                request.mediaType?.let { put("media_type", it) }
                request.yearMin?.let { put("year_min", it.toString()) }
                request.yearMax?.let { put("year_max", it.toString()) }
                request.ratingMin?.let { put("rating_min", it.toString()) }
                request.quality?.let { put("quality", it) }
                request.sortBy?.let { put("sort_by", it) }
                request.sortOrder?.let { put("sort_order", it) }
                put("limit", request.limit.toString())
                put("offset", request.offset.toString())
            }

            // Call API
            val response = api.searchMedia(params)

            if (response.isSuccessful) {
                val items = response.body() ?: emptyList()
                emit(items)
            } else {
                // Log error and emit empty list
                android.util.Log.e("MediaRepository", "Search failed: ${response.code()} ${response.message()}")
                emit(emptyList())
            }
        } catch (e: Exception) {
            // Log exception and emit empty list
            android.util.Log.e("MediaRepository", "Search error", e)
            emit(emptyList())
        }
    }

    fun getMediaById(id: Long): Flow<MediaItem?> = flow {
        try {
            // First, search for the media by ID to get the file path
            val searchParams = mapOf("id" to id.toString())
            val searchResponse = api.searchMedia(searchParams)

            if (searchResponse.isSuccessful) {
                val items = searchResponse.body()
                if (!items.isNullOrEmpty()) {
                    // Found the media item
                    emit(items.first())
                } else {
                    android.util.Log.w("MediaRepository", "Media not found with ID: $id")
                    emit(null)
                }
            } else {
                android.util.Log.e("MediaRepository", "Get media failed: ${searchResponse.code()} ${searchResponse.message()}")
                emit(null)
            }
        } catch (e: Exception) {
            android.util.Log.e("MediaRepository", "Get media error", e)
            emit(null)
        }
    }

    suspend fun updateWatchProgress(mediaId: Long, progress: Double) {
        try {
            val progressData = mapOf("progress" to progress)
            val response = api.updateWatchProgress(mediaId, progressData)

            if (response.isSuccessful) {
                android.util.Log.d("MediaRepository", "Watch progress updated for media $mediaId: $progress")
            } else {
                android.util.Log.e("MediaRepository", "Failed to update watch progress: ${response.code()} ${response.message()}")
            }
        } catch (e: Exception) {
            android.util.Log.e("MediaRepository", "Error updating watch progress", e)
        }
    }

    suspend fun updateFavoriteStatus(mediaId: Long, isFavorite: Boolean) {
        try {
            val favoriteData = mapOf("is_favorite" to isFavorite)
            val response = api.updateFavoriteStatus(mediaId, favoriteData)

            if (response.isSuccessful) {
                android.util.Log.d("MediaRepository", "Favorite status updated for media $mediaId: $isFavorite")
            } else {
                android.util.Log.e("MediaRepository", "Failed to update favorite status: ${response.code()} ${response.message()}")
            }
        } catch (e: Exception) {
            android.util.Log.e("MediaRepository", "Error updating favorite status", e)
        }
    }
}