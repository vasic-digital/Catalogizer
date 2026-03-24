package com.catalogizer.androidtv

import android.app.Application
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class CatalogizerTVApplication : Application() {

    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }
    private val appScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override fun onCreate() {
        super.onCreate()

        // Initialize dependency container and load persisted settings (server URL, etc.)
        appScope.launch {
            dependencyContainer.initializeAsync()
        }
    }
}
