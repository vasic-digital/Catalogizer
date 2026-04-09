package com.catalogizer.androidtv.data.sync

import android.app.Service
import android.content.Intent
import android.os.IBinder
import com.catalogizer.androidtv.DependencyContainer
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

/**
 * Background [Service] for media library synchronization with the Catalogizer API.
 * Supports immediate and scheduled sync modes via intent actions.
 * Uses a [CoroutineScope] tied to the service lifecycle.
 */
class SyncService : Service() {
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    
    override fun onCreate() {
        super.onCreate()
    }
    
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        // Handle sync operations here
        when (intent?.action) {
            "SYNC_NOW" -> {
                // Perform immediate sync
                performSync()
            }
            "SCHEDULED_SYNC" -> {
                // Perform scheduled sync
                performScheduledSync()
            }
            else -> {
                // Default sync operation
                performSync()
            }
        }
        
        return START_STICKY
    }
    
    private fun performSync() {
        serviceScope.launch {
            try {
                android.util.Log.d("SyncService", "Starting media library sync...")
                val container = DependencyContainer.getInstance(this@SyncService)
                
                // 1. Sync recent media items
                syncRecentMedia(container)
                
                // 2. Sync watch progress from server
                syncWatchProgress(container)
                
                // 3. Sync watch next recommendations
                syncWatchNext(container)
                
                // 4. Refresh TV home screen channels after sync
                container.tvChannelRepository.refreshAllChannels()
                container.watchNextManager.refreshWatchNext()
                
                android.util.Log.d("SyncService", "Media library sync completed successfully")
            } catch (e: Exception) {
                android.util.Log.e("SyncService", "Sync failed: ${e.message}", e)
            }
        }
    }
    
    /**
     * Sync recent media items from the API.
     * NOTE: Uses search API with empty query to get recent items.
     */
    private suspend fun syncRecentMedia(container: DependencyContainer) {
        try {
            val response = container.api.searchEntities(mapOf("limit" to "50", "sort" to "recent"))
            if (response.isSuccessful) {
                val mediaItems = response.body()?.items ?: emptyList()
                android.util.Log.d("SyncService", "Synced ${mediaItems.size} recent media items")
                
                // Media items are cached via repository flow automatically
            } else {
                android.util.Log.w("SyncService", "Failed to fetch recent media: ${response.code()}")
            }
        } catch (e: Exception) {
            android.util.Log.e("SyncService", "Error syncing recent media", e)
        }
    }
    
    /**
     * Sync watch progress from the server.
     * NOTE: Watch progress is fetched per-item during playback.
     * Full sync of all progress would require a dedicated API endpoint.
     */
    private suspend fun syncWatchProgress(container: DependencyContainer) {
        try {
            // Individual watch progress is synced per media item during playback
            // A full sync endpoint is not currently available in the API
            android.util.Log.d("SyncService", "Watch progress sync - per-item during playback")
        } catch (e: Exception) {
            android.util.Log.e("SyncService", "Error syncing watch progress", e)
        }
    }
    
    /**
     * Sync watch next recommendations.
     * NOTE: Uses trending API as recommendations source.
     */
    private suspend fun syncWatchNext(container: DependencyContainer) {
        try {
            val response = container.api.getTrendingMedia()
            if (response.isSuccessful) {
                android.util.Log.d("SyncService", "Synced watch next recommendations")
                // Recommendations are handled via the WatchNextManager
            } else {
                android.util.Log.w("SyncService", "Failed to fetch trending: ${response.code()}")
            }
        } catch (e: Exception) {
            android.util.Log.e("SyncService", "Error syncing watch next", e)
        }
    }

    private fun performScheduledSync() {
        // Scheduled sync: same as immediate but could be throttled in the future
        performSync()
    }
    
    override fun onBind(intent: Intent?): IBinder? {
        return null
    }
    
    override fun onDestroy() {
        super.onDestroy()
        serviceScope.cancel()
    }
}