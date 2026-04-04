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
            // TODO: Add media library sync logic here (e.g., fetch updated catalog from API)

            // Refresh TV home screen channels after sync
            try {
                val container = DependencyContainer.getInstance(this@SyncService)
                container.tvChannelRepository.refreshAllChannels()
                container.watchNextManager.refreshWatchNext()
            } catch (e: Exception) {
                android.util.Log.w("SyncService", "Channel refresh failed: ${e.message}")
            }
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