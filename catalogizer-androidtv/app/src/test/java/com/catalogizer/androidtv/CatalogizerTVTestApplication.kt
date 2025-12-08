package com.catalogizer.androidtv

import android.app.Application

/**
 * Test application class for Android TV testing
 * Provides test-specific configurations and mocked dependencies
 */
class CatalogizerTVTestApplication : Application() {

    override fun onCreate() {
        super.onCreate()
        // Skip TV-specific initialization in tests
    }
}