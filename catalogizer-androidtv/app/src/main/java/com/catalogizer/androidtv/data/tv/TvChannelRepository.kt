package com.catalogizer.androidtv.data.tv

import android.content.ContentValues
import android.content.Context
import android.util.Log
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.flow.first

/**
 * Manages Android TV Home Screen channels and programs via [TvContractCompat].
 * Creates a default curated channel, dynamic per-category channels, and populates
 * them with [MediaItem] data from [MediaRepository].
 */
class TvChannelRepository(
    private val context: Context,
    private val mediaRepository: MediaRepository,
    private val settingsRepository: SettingsRepository
) {
    companion object {
        private const val TAG = "TvChannelRepo"
        const val MAX_PROGRAMS_PER_CHANNEL = 30
        const val DEFAULT_CHANNEL_KEY = "default"
        const val DEFAULT_CHANNEL_DISPLAY_NAME = "Catalogizer Picks"
    }

    // ─── Content Building (testable, no ContentResolver dependency) ─────

    suspend fun buildDefaultChannelContent(): List<MediaItem> {
        val seen = mutableSetOf<Long>()
        val result = mutableListOf<MediaItem>()

        // 1. Continue watching (partially watched, not completed)
        try {
            val continueWatching = mediaRepository.searchMedia(
                MediaSearchRequest(sortBy = "created", sortOrder = "desc", limit = 10)
            ).first().filter { it.watchProgress > 0.05 && it.watchProgress < 0.9 }
            for (item in continueWatching) {
                if (seen.add(item.id)) result.add(item)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load continue watching: ${e.message}")
        }

        // 2. Recently added across all types
        try {
            val recent = mediaRepository.browseEntities(
                "all", limit = 20, sortBy = "created", sortOrder = "desc"
            ).first()
            for (item in recent) {
                if (seen.add(item.id)) result.add(item)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load recent: ${e.message}")
        }

        // 3. Trending
        try {
            val trending = mediaRepository.getTrendingItems(10)
            for (item in trending) {
                if (seen.add(item.id)) result.add(item)
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load trending: ${e.message}")
        }

        return result.take(MAX_PROGRAMS_PER_CHANNEL)
    }

    suspend fun buildCategoryContent(mediaType: String): List<MediaItem> {
        return try {
            mediaRepository.browseEntities(
                mediaType, limit = MAX_PROGRAMS_PER_CHANNEL, sortBy = "created", sortOrder = "desc"
            ).first().take(MAX_PROGRAMS_PER_CHANNEL)
        } catch (e: Exception) {
            Log.w(TAG, "Failed to load category $mediaType: ${e.message}")
            emptyList()
        }
    }

    suspend fun getActiveMediaTypes(): List<String> {
        return try {
            val (_, byType) = mediaRepository.getEntityStats()
            byType.filter { it.value > 0 }.keys.toList()
        } catch (e: Exception) {
            Log.w(TAG, "Failed to get entity stats: ${e.message}")
            emptyList()
        }
    }

    // ─── Channel CRUD (interacts with system ContentResolver) ───────────

    suspend fun initializeDefaultChannel() {
        val existingId = settingsRepository.getChannelId(DEFAULT_CHANNEL_KEY)
        if (existingId != null) return

        val channelValues = ContentValues().apply {
            put(TvContractCompat.Channels.COLUMN_DISPLAY_NAME, DEFAULT_CHANNEL_DISPLAY_NAME)
            put(TvContractCompat.Channels.COLUMN_APP_LINK_INTENT_URI, "catalogizer://home")
            put(TvContractCompat.Channels.COLUMN_TYPE, TvContractCompat.Channels.TYPE_PREVIEW)
        }

        try {
            val channelUri = context.contentResolver.insert(
                TvContractCompat.Channels.CONTENT_URI, channelValues
            )
            val channelId = channelUri?.let { android.content.ContentUris.parseId(it) }
            if (channelId != null && channelId > 0) {
                settingsRepository.saveChannelId(DEFAULT_CHANNEL_KEY, channelId)
                TvContractCompat.requestChannelBrowsable(context, channelId)
                Log.d(TAG, "Default channel created: $channelId")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to create default channel: ${e.message}")
        }
    }

    suspend fun createCategoryChannel(mediaType: String, displayName: String): Long? {
        val existingId = settingsRepository.getChannelId(mediaType)
        if (existingId != null) return existingId

        val channelValues = ContentValues().apply {
            put(TvContractCompat.Channels.COLUMN_DISPLAY_NAME, displayName)
            put(TvContractCompat.Channels.COLUMN_APP_LINK_INTENT_URI, "catalogizer://browse/$mediaType")
            put(TvContractCompat.Channels.COLUMN_TYPE, TvContractCompat.Channels.TYPE_PREVIEW)
            put(TvContractCompat.Channels.COLUMN_INTERNAL_PROVIDER_ID, mediaType)
        }

        return try {
            val channelUri = context.contentResolver.insert(
                TvContractCompat.Channels.CONTENT_URI, channelValues
            )
            val channelId = channelUri?.let { android.content.ContentUris.parseId(it) }
            if (channelId != null && channelId > 0) {
                settingsRepository.saveChannelId(mediaType, channelId)
                Log.d(TAG, "Category channel created: $mediaType -> $channelId")
            }
            channelId
        } catch (e: Exception) {
            Log.e(TAG, "Failed to create category channel $mediaType: ${e.message}")
            null
        }
    }

    suspend fun refreshChannelPrograms(channelId: Long, items: List<MediaItem>) {
        try {
            context.contentResolver.delete(
                TvContractCompat.PreviewPrograms.CONTENT_URI,
                "${TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID} = ?",
                arrayOf(channelId.toString())
            )
        } catch (e: Exception) {
            Log.w(TAG, "Failed to delete old programs for channel $channelId: ${e.message}")
        }

        for (item in items) {
            try {
                val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId)
                context.contentResolver.insert(TvContractCompat.PreviewPrograms.CONTENT_URI, values)
            } catch (e: Exception) {
                Log.w(TAG, "Failed to insert program ${item.id}: ${e.message}")
            }
        }
    }

    suspend fun deleteChannel(channelKey: String) {
        val channelId = settingsRepository.getChannelId(channelKey) ?: return
        try {
            val channelUri = TvContractCompat.buildChannelUri(channelId)
            context.contentResolver.delete(channelUri, null, null)
            settingsRepository.removeChannelId(channelKey)
            Log.d(TAG, "Deleted channel: $channelKey ($channelId)")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to delete channel $channelKey: ${e.message}")
        }
    }

    suspend fun deleteAllChannels() {
        deleteChannel(DEFAULT_CHANNEL_KEY)
        val allTypes = MediaType.values().map { it.value }
        for (type in allTypes) {
            deleteChannel(type)
        }
        settingsRepository.clearAllChannelIds()
        Log.d(TAG, "All channels deleted")
    }

    // ─── Orchestration ──────────────────────────────────────────────────

    suspend fun refreshAllChannels() {
        refreshDefaultChannel()
        createCategoryChannels()
        removeStaleCategoryChannels()
        Log.d(TAG, "All channels refreshed")
    }

    suspend fun refreshDefaultChannel() {
        val channelId = settingsRepository.getChannelId(DEFAULT_CHANNEL_KEY) ?: return
        val content = buildDefaultChannelContent()
        refreshChannelPrograms(channelId, content)
    }

    suspend fun createCategoryChannels() {
        val activeTypes = getActiveMediaTypes()
        for (type in activeTypes) {
            val displayName = MediaType.fromValue(type).displayName
            val channelId = createCategoryChannel(type, displayName) ?: continue
            val content = buildCategoryContent(type)
            refreshChannelPrograms(channelId, content)
        }
    }

    suspend fun removeStaleCategoryChannels() {
        val activeTypes = getActiveMediaTypes().toSet()
        val allTypes = MediaType.values().map { it.value }
        for (type in allTypes) {
            if (type !in activeTypes) {
                val channelId = settingsRepository.getChannelId(type)
                if (channelId != null) {
                    deleteChannel(type)
                }
            }
        }
    }
}
