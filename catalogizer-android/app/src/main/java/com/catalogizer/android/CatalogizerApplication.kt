package com.catalogizer.android

import android.app.Application
import androidx.work.Configuration

class CatalogizerApplication : Application(), Configuration.Provider {

    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }

    override val workManagerConfiguration: Configuration
        get() = Configuration.Builder()
            .setWorkerFactory(CatalogizerWorkerFactory(dependencyContainer))
            .build()

    override fun onCreate() {
        super.onCreate()
    }
}