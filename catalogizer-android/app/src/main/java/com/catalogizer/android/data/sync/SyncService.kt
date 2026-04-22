package com.catalogizer.android.data.sync

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Binder
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.catalogizer.android.R
import com.catalogizer.android.ui.MainActivity
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

/**
 * SyncService provides foreground service for data synchronization.
 * Handles media library sync, metadata updates, and offline content management.
 */
class SyncService : Service() {

    private val binder = LocalBinder()
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var syncJob: Job? = null
    
    private val _syncState = MutableStateFlow<SyncState>(SyncState.Idle)
    val syncState: StateFlow<SyncState> = _syncState
    
    inner class LocalBinder : Binder() {
        fun getService(): SyncService = this@SyncService
    }
    
    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
    }
    
    override fun onBind(intent: Intent): IBinder {
        return binder
    }
    
    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START_SYNC -> {
                startForegroundService()
                performSync()
            }
            ACTION_STOP_SYNC -> {
                stopSync()
            }
        }
        return START_STICKY
    }
    
    private fun startForegroundService() {
        val notification = createNotification("Starting sync...", 0)
        startForeground(NOTIFICATION_ID, notification)
    }
    
    private fun performSync() {
        if (syncJob?.isActive == true) return
        
        syncJob = serviceScope.launch {
            try {
                _syncState.value = SyncState.InProgress(0)
                
                // Simulate sync phases
                updateNotification("Syncing media library...", 25)
                kotlinx.coroutines.delay(1000)
                _syncState.value = SyncState.InProgress(25)
                
                updateNotification("Updating metadata...", 50)
                kotlinx.coroutines.delay(1000)
                _syncState.value = SyncState.InProgress(50)
                
                updateNotification("Checking offline content...", 75)
                kotlinx.coroutines.delay(1000)
                _syncState.value = SyncState.InProgress(75)
                
                updateNotification("Sync completed", 100)
                _syncState.value = SyncState.Completed
                
                // Stop service after completion
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                    stopForeground(STOP_FOREGROUND_REMOVE)
                } else {
                    @Suppress("DEPRECATION")
                    stopForeground(true)
                }
                stopSelf()
                
            } catch (e: Exception) {
                _syncState.value = SyncState.Error(e.message ?: "Unknown error")
                updateNotification("Sync failed: ${e.message}", 0)
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                    stopForeground(STOP_FOREGROUND_REMOVE)
                } else {
                    @Suppress("DEPRECATION")
                    stopForeground(true)
                }
                stopSelf()
            }
        }
    }
    
    private fun stopSync() {
        syncJob?.cancel()
        _syncState.value = SyncState.Idle
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            stopForeground(STOP_FOREGROUND_REMOVE)
        } else {
            @Suppress("DEPRECATION")
            stopForeground(true)
        }
        stopSelf()
    }
    
    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Sync Service",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Background synchronization service"
                setShowBadge(false)
            }
            
            val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            notificationManager.createNotificationChannel(channel)
        }
    }
    
    private fun createNotification(content: String, progress: Int): Notification {
        val intent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            intent,
            PendingIntent.FLAG_IMMUTABLE
        )
        
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("Catalogizer Sync")
            .setContentText(content)
            .setSmallIcon(android.R.drawable.ic_popup_sync)
            .setContentIntent(pendingIntent)
            .setProgress(100, progress, progress == 0)
            .setOngoing(true)
            .build()
    }
    
    private fun updateNotification(content: String, progress: Int) {
        val notification = createNotification(content, progress)
        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        notificationManager.notify(NOTIFICATION_ID, notification)
    }
    
    override fun onDestroy() {
        super.onDestroy()
        serviceScope.cancel()
    }
    
    sealed class SyncState {
        object Idle : SyncState()
        data class InProgress(val progress: Int) : SyncState()
        object Completed : SyncState()
        data class Error(val message: String) : SyncState()
    }
    
    companion object {
        const val ACTION_START_SYNC = "com.catalogizer.android.action.START_SYNC"
        const val ACTION_STOP_SYNC = "com.catalogizer.android.action.STOP_SYNC"
        const val CHANNEL_ID = "sync_service_channel"
        const val NOTIFICATION_ID = 1001
        
        fun start(context: Context) {
            val intent = Intent(context, SyncService::class.java).apply {
                action = ACTION_START_SYNC
            }
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }
        
        fun stop(context: Context) {
            val intent = Intent(context, SyncService::class.java).apply {
                action = ACTION_STOP_SYNC
            }
            context.startService(intent)
        }
    }
}
