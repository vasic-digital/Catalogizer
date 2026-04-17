package com.catalogizer.androidtv.data.tv

import android.content.ContentValues
import android.content.Context
import android.util.Log
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock

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

    private val channelMutex = Mutex()

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

    /**
     * Queries the current display name of a channel by its system ID.
     */
    private fun getChannelDisplayName(channelId: Long): String? {
        return try {
            val uri = TvContractCompat.buildChannelUri(channelId)
            context.contentResolver.query(
                uri,
                arrayOf(TvContractCompat.Channels.COLUMN_DISPLAY_NAME),
                null, null, null
            )?.use { cursor ->
                if (cursor.moveToFirst()) {
                    cursor.getString(0)
                } else null
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to query channel display name for $channelId: ${e.message}")
            null
        }
    }

    suspend fun initializeDefaultChannel() {
        val existingId = settingsRepository.getChannelId(DEFAULT_CHANNEL_KEY)
        if (existingId != null) {
            val currentName = getChannelDisplayName(existingId)
            if (currentName != null && currentName != DEFAULT_CHANNEL_DISPLAY_NAME) {
                // Name is stale; delete and recreate to force the launcher to refresh
                Log.d(TAG, "Default channel name mismatch '$currentName' != '$DEFAULT_CHANNEL_DISPLAY_NAME'; recreating")
                deleteChannel(DEFAULT_CHANNEL_KEY)
            } else {
                // Ensure other metadata is up to date
                try {
                    val channelValues = ContentValues().apply {
                        put(TvContractCompat.Channels.COLUMN_DISPLAY_NAME, DEFAULT_CHANNEL_DISPLAY_NAME)
                        put(TvContractCompat.Channels.COLUMN_APP_LINK_INTENT_URI, "catalogizer://home")
                        put(TvContractCompat.Channels.COLUMN_TYPE, TvContractCompat.Channels.TYPE_PREVIEW)
                    }
                    val channelUri = TvContractCompat.buildChannelUri(existingId)
                    val updated = context.contentResolver.update(channelUri, channelValues, null, null)
                    if (updated > 0) {
                        Log.d(TAG, "Default channel updated: $existingId")
                    }
                } catch (e: Exception) {
                    Log.w(TAG, "Failed to update default channel: ${e.message}")
                }
                return
            }
        }

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
        if (existingId != null) {
            val currentName = getChannelDisplayName(existingId)
            if (currentName != null && currentName != displayName) {
                // Name is stale/incorrect; delete and recreate to force the launcher to refresh
                Log.d(TAG, "Category channel name mismatch '$currentName' != '$displayName' for $mediaType; recreating")
                deleteChannel(mediaType)
            } else {
                // Update metadata in case intent URI or internal provider ID changed
                try {
                    val channelValues = ContentValues().apply {
                        put(TvContractCompat.Channels.COLUMN_DISPLAY_NAME, displayName)
                        put(TvContractCompat.Channels.COLUMN_APP_LINK_INTENT_URI, "catalogizer://browse/$mediaType")
                        put(TvContractCompat.Channels.COLUMN_TYPE, TvContractCompat.Channels.TYPE_PREVIEW)
                        put(TvContractCompat.Channels.COLUMN_INTERNAL_PROVIDER_ID, mediaType)
                    }
                    val channelUri = TvContractCompat.buildChannelUri(existingId)
                    val updated = context.contentResolver.update(channelUri, channelValues, null, null)
                    if (updated > 0) {
                        Log.d(TAG, "Category channel updated: $mediaType -> $existingId")
                    }
                } catch (e: Exception) {
                    Log.w(TAG, "Failed to update category channel $mediaType: ${e.message}")
                }
                return existingId
            }
        }

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

        val serverUrl = try {
            DependencyContainer.getInstance(context).getServerUrl()
        } catch (_: Exception) { null }

        for (item in items) {
            try {
                val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId, serverUrl)
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
        channelMutex.withLock {
            deleteChannel(DEFAULT_CHANNEL_KEY)
            val allTypes = MediaType.values().map { it.value }
            for (type in allTypes) {
                deleteChannel(type)
            }
            settingsRepository.clearAllChannelIds()
            Log.d(TAG, "All channels deleted")
        }
    }

    // ─── Orchestration ──────────────────────────────────────────────────

    suspend fun refreshAllChannels() {
        channelMutex.withLock {
            refreshDefaultChannel()
            createCategoryChannels()
            removeStaleCategoryChannels()
            removeOrphanedChannels()
            Log.d(TAG, "All channels refreshed")
        }
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

    /**
     * Queries channel IDs by their internal provider ID.
     */
    private fun queryChannelIdsByProviderId(providerId: String): List<Long> {
        return try {
            context.contentResolver.query(
                TvContractCompat.Channels.CONTENT_URI,
                arrayOf(TvContractCompat.Channels._ID),
                "${TvContractCompat.Channels.COLUMN_INTERNAL_PROVIDER_ID} = ?",
                arrayOf(providerId),
                null
            )?.use { cursor ->
                val idIndex = cursor.getColumnIndex(TvContractCompat.Channels._ID)
                buildList {
                    while (cursor.moveToNext()) {
                        if (idIndex >= 0) add(cursor.getLong(idIndex))
                    }
                }
            } ?: emptyList()
        } catch (e: Exception) {
            Log.w(TAG, "Failed to query channels for providerId=$providerId: ${e.message}")
            emptyList()
        }
    }

    /**
     * Deletes any system channels that have the same internal provider ID as our
     * known channels but different IDs (duplicates), or channels for provider IDs
     * that we no longer manage. This cleans up stale channels left behind by
     * previous installs without requiring package-name filtering.
     */
    suspend fun removeOrphanedChannels() {
        try {
            val knownIds = settingsRepository.getAllChannelIds().toSet()

            // 1. Remove duplicate channels for each category we manage
            val managedProviderIds = MediaType.values().map { it.value } + DEFAULT_CHANNEL_KEY
            for (providerId in managedProviderIds) {
                val systemIds = queryChannelIdsByProviderId(providerId)
                for (channelId in systemIds) {
                    if (channelId !in knownIds) {
                        val uri = TvContractCompat.buildChannelUri(channelId)
                        context.contentResolver.delete(uri, null, null)
                        Log.d(TAG, "Removed orphaned/duplicate channel: $channelId (provider=$providerId)")
                    }
                }
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to remove orphaned channels: ${e.message}")
        }
    }
}
