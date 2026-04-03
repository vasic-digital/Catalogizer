package com.catalogizer.android

import android.app.Application
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ProcessLifecycleOwner
import androidx.work.Configuration
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

class CatalogizerApplication : Application(), Configuration.Provider {

    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }
    private val appScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override val workManagerConfiguration: Configuration
        get() = Configuration.Builder()
            .setWorkerFactory(CatalogizerWorkerFactory(dependencyContainer))
            .build()

    override fun onCreate() {
        super.onCreate()

        // Cancel appScope when the app process is destroyed to prevent leaked coroutines
        ProcessLifecycleOwner.get().lifecycle.addObserver(object : DefaultLifecycleObserver {
            override fun onDestroy(owner: LifecycleOwner) {
                appScope.cancel()
            }
        })

        // Initialize dependency container and load persisted settings (server URL, etc.)
        appScope.launch {
            dependencyContainer.initializeAsync()
        }
    }

    override fun onTerminate() {
        super.onTerminate()
        appScope.cancel()
        dependencyContainer.shutdown()
    }
}