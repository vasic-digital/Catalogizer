package com.catalogizer.androidtv

import android.app.Application

class CatalogizerTVApplication : Application() {

    val dependencyContainer by lazy { DependencyContainer.getInstance(this) }

    override fun onCreate() {
        super.onCreate()

        // Initialize any TV-specific configurations
        initializeTVSettings()
    }

    private fun initializeTVSettings() {
        // Configure for TV environment
        // Set up media session callbacks
        // Initialize background tasks
    }
}