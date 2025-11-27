package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.Settings

class SettingsRepository {
    private var currentSettings: Settings = Settings(
        enableNotifications = true,
        enableAutoPlay = false,
        streamingQuality = "Auto",
        enableSubtitles = true,
        subtitleLanguage = "English"
    )

    fun getSettings(): Settings {
        return currentSettings
    }

    fun saveSettings(settings: Settings) {
        currentSettings = settings
        // TODO: Implement persistence
    }
}