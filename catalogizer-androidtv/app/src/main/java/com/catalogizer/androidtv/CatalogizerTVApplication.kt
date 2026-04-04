package com.catalogizer.androidtv

import android.app.Application
import com.catalogizer.androidtv.data.tv.TvChannelSyncWorker
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

/**
 * Application-level class for the Android TV variant. Initializes the
 * [DependencyContainer] and loads persisted server settings asynchronously
 * during application startup.
 */
class CatalogizerTVApplication : Application() {

    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }
    private val appScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override fun onCreate() {
        super.onCreate()

        // Initialize dependency container and load persisted settings (server URL, etc.)
        appScope.launch {
            dependencyContainer.initializeAsync()

            // Initialize default channel and enqueue periodic sync
            try {
                dependencyContainer.tvChannelRepository.initializeDefaultChannel()
                TvChannelSyncWorker.enqueue(this@CatalogizerTVApplication)
            } catch (e: Exception) {
                // Channel initialization can fail if not authenticated yet — that's OK
                android.util.Log.w("CatalogizerTV", "Channel init deferred: ${e.message}")
            }
        }
    }
}
