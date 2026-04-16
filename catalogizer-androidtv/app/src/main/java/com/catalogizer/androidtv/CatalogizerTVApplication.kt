package com.catalogizer.androidtv

import android.app.Application
import android.util.Log
import coil.ImageLoader
import coil.ImageLoaderFactory
import coil.decode.SvgDecoder
import com.catalogizer.androidtv.data.tv.TvChannelSyncWorker
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull

/**
 * Application-level class for the Android TV variant. Initializes the
 * [DependencyContainer] and loads persisted server settings asynchronously
 * during application startup.
 * 
 * CRITICAL: All initialization must complete in <2 seconds to prevent ANR.
 * Heavy operations are deferred or run on background threads.
 */
class CatalogizerTVApplication : Application(), ImageLoaderFactory {

    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }
    private val appScope = CoroutineScope(SupervisorJob() + Dispatchers.Default)

    override fun newImageLoader(): ImageLoader {
        return ImageLoader.Builder(this)
            .components {
                add(SvgDecoder.Factory())
            }
            .allowHardware(false)
            .build()
    }

    override fun onCreate() {
        super.onCreate()
        val startTime = System.currentTimeMillis()
        Log.i("CatalogizerTV", "Application.onCreate() starting...")

        // Initialize dependency container and load persisted settings (server URL, etc.)
        // Use Default dispatcher (background thread pool) to avoid blocking main thread
        appScope.launch {
            try {
                // Hard timeout of 3 seconds for initialization to ensure we never block too long
                withTimeoutOrNull(3000L) {
                    dependencyContainer.initializeAsync()
                } ?: Log.w("CatalogizerTV", "Initialization timed out after 3s - continuing anyway")
                
                Log.i("CatalogizerTV", "Container initialization completed")
            } catch (e: Exception) {
                Log.e("CatalogizerTV", "Container initialization failed: ${e.message}")
            }

            // Initialize default channel and enqueue periodic sync
            // This is non-critical and can fail silently
            try {
                dependencyContainer.tvChannelRepository.initializeDefaultChannel()
                TvChannelSyncWorker.enqueue(this@CatalogizerTVApplication)
            } catch (e: Exception) {
                // Channel initialization can fail if not authenticated yet — that's OK
                Log.w("CatalogizerTV", "Channel init deferred: ${e.message}")
            }
            
            val elapsed = System.currentTimeMillis() - startTime
            Log.i("CatalogizerTV", "Application.onCreate() completed in ${elapsed}ms")
        }
    }
}
