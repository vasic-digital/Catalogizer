@file:SuppressLint("RestrictedApi")

package com.catalogizer.androidtv.data.tv

import android.annotation.SuppressLint
import android.content.Context
import android.util.Log
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import kotlinx.coroutines.flow.first

/**
 * Manages the system Watch Next row on the Android TV home screen.
 * Adds partially-watched items (5%-90%), auto-surfaces next episodes for TV series,
 * and removes completed or stale entries.
 */
class WatchNextManager(
    private val context: Context,
    private val mediaRepository: MediaRepository
) {
    companion object {
        private const val TAG = "WatchNextMgr"
        private const val PROGRESS_MIN = 0.05
        private const val PROGRESS_MAX = 0.90
        private const val STALE_THRESHOLD_MS = 30L * 24 * 60 * 60 * 1000 // 30 days

        fun filterContinueWatching(items: List<MediaItem>): List<MediaItem> {
            return items.filter { it.watchProgress >= PROGRESS_MIN && it.watchProgress < PROGRESS_MAX }
        }

        fun isStale(lastEngagementTimeMs: Long): Boolean {
            if (lastEngagementTimeMs == 0L) return false
            return (System.currentTimeMillis() - lastEngagementTimeMs) > STALE_THRESHOLD_MS
        }

        fun isCompleted(progress: Double): Boolean {
            return progress > PROGRESS_MAX
        }

        fun isTvSeriesType(mediaType: String?): Boolean {
            return mediaType == "tv_show" || mediaType == "tv_episode"
        }
    }

    suspend fun refreshWatchNext() {
        val allItems = try {
            mediaRepository.searchMedia(
                MediaSearchRequest(sortBy = "updated_at", sortOrder = "desc", limit = 50)
            ).first().filter { it.watchProgress > 0.0 }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to fetch watched items: ${e.message}")
            return
        }

        removeAll()

        val continueItems = filterContinueWatching(allItems)
        for (item in continueItems) {
            addWatchNextProgram(item, TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE)
        }

        val completedEpisodes = allItems.filter { isCompleted(it.watchProgress) && isTvSeriesType(it.mediaType) }
        for (episode in completedEpisodes) {
            val nextEpisode = findNextEpisode(episode)
            if (nextEpisode != null) {
                addWatchNextProgram(nextEpisode, TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_NEXT)
            }
        }

        Log.d(TAG, "Watch Next refreshed: ${continueItems.size} continue, ${completedEpisodes.size} completed episodes checked")
    }

    private suspend fun findNextEpisode(currentEpisode: MediaItem): MediaItem? {
        return try {
            val similar = mediaRepository.getSimilarItems(currentEpisode.id)
            similar.firstOrNull { it.mediaType == "tv_episode" && it.id != currentEpisode.id }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to find next episode for ${currentEpisode.id}: ${e.message}")
            null
        }
    }

    private fun addWatchNextProgram(item: MediaItem, watchNextType: Int) {
        try {
            val serverUrl = try {
                DependencyContainer.getInstance(context).getServerUrl()
            } catch (_: Exception) { null }
            val values = ChannelProgramMapper.toWatchNextValues(item, watchNextType, serverUrl)
            context.contentResolver.insert(TvContractCompat.WatchNextPrograms.CONTENT_URI, values)
        } catch (e: Exception) {
            Log.w(TAG, "Failed to add Watch Next for ${item.id}: ${e.message}")
        }
    }

    fun removeAll() {
        try {
            // Only delete Catalogizer's Watch Next entries (identified by our URI scheme)
            context.contentResolver.delete(
                TvContractCompat.WatchNextPrograms.CONTENT_URI,
                "${TvContractCompat.WatchNextPrograms.COLUMN_INTENT_URI} LIKE ?",
                arrayOf("catalogizer://%")
            )
        } catch (e: Exception) {
            Log.w(TAG, "Failed to clear Watch Next: ${e.message}")
        }
    }
}
