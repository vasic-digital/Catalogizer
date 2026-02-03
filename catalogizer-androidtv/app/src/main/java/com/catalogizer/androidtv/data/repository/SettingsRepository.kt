package com.catalogizer.androidtv.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import com.catalogizer.androidtv.data.models.Settings
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

class SettingsRepository(private val dataStore: DataStore<Preferences>) {

    companion object {
        private val ENABLE_NOTIFICATIONS = booleanPreferencesKey("enable_notifications")
        private val ENABLE_AUTO_PLAY = booleanPreferencesKey("enable_auto_play")
        private val STREAMING_QUALITY = stringPreferencesKey("streaming_quality")
        private val ENABLE_SUBTITLES = booleanPreferencesKey("enable_subtitles")
        private val SUBTITLE_LANGUAGE = stringPreferencesKey("subtitle_language")
    }

    // In-memory cache for synchronous access
    private var cachedSettings: Settings = Settings(
        enableNotifications = true,
        enableAutoPlay = false,
        streamingQuality = "Auto",
        enableSubtitles = true,
        subtitleLanguage = "English"
    )

    /**
     * Get settings as a Flow for reactive updates
     */
    val settingsFlow: Flow<Settings> = dataStore.data.map { preferences ->
        Settings(
            enableNotifications = preferences[ENABLE_NOTIFICATIONS] ?: true,
            enableAutoPlay = preferences[ENABLE_AUTO_PLAY] ?: false,
            streamingQuality = preferences[STREAMING_QUALITY] ?: "Auto",
            enableSubtitles = preferences[ENABLE_SUBTITLES] ?: true,
            subtitleLanguage = preferences[SUBTITLE_LANGUAGE] ?: "English"
        ).also { cachedSettings = it }
    }

    /**
     * Get current settings (synchronous, uses cache)
     * For reactive updates, use settingsFlow instead
     */
    fun getSettings(): Settings {
        return cachedSettings
    }

    /**
     * Get settings from DataStore (suspend function)
     */
    suspend fun getSettingsAsync(): Settings {
        return settingsFlow.first()
    }

    /**
     * Save settings to DataStore
     */
    suspend fun saveSettings(settings: Settings) {
        dataStore.edit { preferences ->
            preferences[ENABLE_NOTIFICATIONS] = settings.enableNotifications
            preferences[ENABLE_AUTO_PLAY] = settings.enableAutoPlay
            preferences[STREAMING_QUALITY] = settings.streamingQuality
            preferences[ENABLE_SUBTITLES] = settings.enableSubtitles
            preferences[SUBTITLE_LANGUAGE] = settings.subtitleLanguage
        }
        cachedSettings = settings
    }

    /**
     * Update a single setting
     */
    suspend fun updateNotifications(enabled: Boolean) {
        dataStore.edit { preferences ->
            preferences[ENABLE_NOTIFICATIONS] = enabled
        }
        cachedSettings = cachedSettings.copy(enableNotifications = enabled)
    }

    suspend fun updateAutoPlay(enabled: Boolean) {
        dataStore.edit { preferences ->
            preferences[ENABLE_AUTO_PLAY] = enabled
        }
        cachedSettings = cachedSettings.copy(enableAutoPlay = enabled)
    }

    suspend fun updateStreamingQuality(quality: String) {
        dataStore.edit { preferences ->
            preferences[STREAMING_QUALITY] = quality
        }
        cachedSettings = cachedSettings.copy(streamingQuality = quality)
    }

    suspend fun updateSubtitles(enabled: Boolean) {
        dataStore.edit { preferences ->
            preferences[ENABLE_SUBTITLES] = enabled
        }
        cachedSettings = cachedSettings.copy(enableSubtitles = enabled)
    }

    suspend fun updateSubtitleLanguage(language: String) {
        dataStore.edit { preferences ->
            preferences[SUBTITLE_LANGUAGE] = language
        }
        cachedSettings = cachedSettings.copy(subtitleLanguage = language)
    }

    /**
     * Reset settings to defaults
     */
    suspend fun resetToDefaults() {
        val defaults = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        saveSettings(defaults)
    }
}
