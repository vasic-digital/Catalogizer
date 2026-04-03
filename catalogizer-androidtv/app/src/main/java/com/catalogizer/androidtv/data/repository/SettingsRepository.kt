package com.catalogizer.androidtv.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import com.catalogizer.androidtv.data.models.ServerEntry
import com.catalogizer.androidtv.data.models.Settings
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import org.json.JSONArray
import org.json.JSONObject

/**
 * Persists and retrieves user settings via [DataStore] including playback preferences,
 * subtitle configuration, server URL management, and saved server entries.
 * Maintains an in-memory [cachedSettings] for synchronous reads.
 */
class SettingsRepository(private val dataStore: DataStore<Preferences>) {

    companion object {
        private val ENABLE_NOTIFICATIONS = booleanPreferencesKey("enable_notifications")
        private val ENABLE_AUTO_PLAY = booleanPreferencesKey("enable_auto_play")
        private val STREAMING_QUALITY = stringPreferencesKey("streaming_quality")
        private val ENABLE_SUBTITLES = booleanPreferencesKey("enable_subtitles")
        private val SUBTITLE_LANGUAGE = stringPreferencesKey("subtitle_language")
        private val SERVER_URL = stringPreferencesKey("server_url")
        private val SAVED_SERVERS = stringPreferencesKey("saved_servers_json")
        private val AUTO_DISCOVERY = booleanPreferencesKey("auto_discovery")
        private val LAST_USERNAME = stringPreferencesKey("last_username")
    }

    private var cachedSettings: Settings = Settings(
        enableNotifications = true,
        enableAutoPlay = false,
        streamingQuality = "Auto",
        enableSubtitles = true,
        subtitleLanguage = "English"
    )

    val settingsFlow: Flow<Settings> = dataStore.data.map { preferences ->
        Settings(
            enableNotifications = preferences[ENABLE_NOTIFICATIONS] ?: true,
            enableAutoPlay = preferences[ENABLE_AUTO_PLAY] ?: false,
            streamingQuality = preferences[STREAMING_QUALITY] ?: "Auto",
            enableSubtitles = preferences[ENABLE_SUBTITLES] ?: true,
            subtitleLanguage = preferences[SUBTITLE_LANGUAGE] ?: "English",
            serverUrl = preferences[SERVER_URL] ?: Settings.DEFAULT_SERVER_URL,
            savedServers = deserializeServers(preferences[SAVED_SERVERS]),
            autoDiscovery = preferences[AUTO_DISCOVERY] ?: true,
            lastUsername = preferences[LAST_USERNAME]
        ).also { cachedSettings = it }
    }

    fun getSettings(): Settings = cachedSettings

    suspend fun getSettingsAsync(): Settings = settingsFlow.first()

    suspend fun saveSettings(settings: Settings) {
        dataStore.edit { preferences ->
            preferences[ENABLE_NOTIFICATIONS] = settings.enableNotifications
            preferences[ENABLE_AUTO_PLAY] = settings.enableAutoPlay
            preferences[STREAMING_QUALITY] = settings.streamingQuality
            preferences[ENABLE_SUBTITLES] = settings.enableSubtitles
            preferences[SUBTITLE_LANGUAGE] = settings.subtitleLanguage
            preferences[SERVER_URL] = settings.serverUrl
            preferences[SAVED_SERVERS] = serializeServers(settings.savedServers)
            preferences[AUTO_DISCOVERY] = settings.autoDiscovery
        }
        cachedSettings = settings
    }

    // ─── Server URL Management ──────────────────────────────────────────

    /**
     * Get the current server URL. Returns cached value for instant access.
     */
    fun getServerUrl(): String = cachedSettings.serverUrl

    /**
     * Update the active server URL and persist it.
     */
    suspend fun updateServerUrl(url: String) {
        dataStore.edit { preferences ->
            preferences[SERVER_URL] = url
        }
        cachedSettings = cachedSettings.copy(serverUrl = url)
    }

    suspend fun updateLastUsername(username: String) {
        dataStore.edit { preferences ->
            preferences[LAST_USERNAME] = username
        }
        cachedSettings = cachedSettings.copy(lastUsername = username)
    }

    /**
     * Add a new server entry to the saved list.
     */
    suspend fun addServer(entry: ServerEntry) {
        val current = cachedSettings.savedServers.toMutableList()
        // Remove existing entry with same URL to avoid duplicates
        current.removeAll { it.url == entry.url }
        current.add(entry)
        val updated = cachedSettings.copy(savedServers = current)
        dataStore.edit { preferences ->
            preferences[SAVED_SERVERS] = serializeServers(current)
        }
        cachedSettings = updated
    }

    /**
     * Remove a server entry by URL.
     */
    suspend fun removeServer(url: String) {
        val current = cachedSettings.savedServers.filter { it.url != url }
        dataStore.edit { preferences ->
            preferences[SAVED_SERVERS] = serializeServers(current)
        }
        cachedSettings = cachedSettings.copy(savedServers = current)
    }

    /**
     * Get all saved server entries.
     */
    fun getSavedServers(): List<ServerEntry> = cachedSettings.savedServers

    /**
     * Update auto-discovery preference.
     */
    suspend fun updateAutoDiscovery(enabled: Boolean) {
        dataStore.edit { preferences ->
            preferences[AUTO_DISCOVERY] = enabled
        }
        cachedSettings = cachedSettings.copy(autoDiscovery = enabled)
    }

    // ─── Existing Setting Updates ───────────────────────────────────────

    suspend fun updateNotifications(enabled: Boolean) {
        dataStore.edit { it[ENABLE_NOTIFICATIONS] = enabled }
        cachedSettings = cachedSettings.copy(enableNotifications = enabled)
    }

    suspend fun updateAutoPlay(enabled: Boolean) {
        dataStore.edit { it[ENABLE_AUTO_PLAY] = enabled }
        cachedSettings = cachedSettings.copy(enableAutoPlay = enabled)
    }

    suspend fun updateStreamingQuality(quality: String) {
        dataStore.edit { it[STREAMING_QUALITY] = quality }
        cachedSettings = cachedSettings.copy(streamingQuality = quality)
    }

    suspend fun updateSubtitles(enabled: Boolean) {
        dataStore.edit { it[ENABLE_SUBTITLES] = enabled }
        cachedSettings = cachedSettings.copy(enableSubtitles = enabled)
    }

    suspend fun updateSubtitleLanguage(language: String) {
        dataStore.edit { it[SUBTITLE_LANGUAGE] = language }
        cachedSettings = cachedSettings.copy(subtitleLanguage = language)
    }

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

    // ─── JSON Serialization for ServerEntry list ────────────────────────

    private fun serializeServers(servers: List<ServerEntry>): String {
        if (servers.isEmpty()) return "[]"
        val arr = JSONArray()
        for (s in servers) {
            val obj = JSONObject()
            obj.put("url", s.url)
            obj.put("name", s.name)
            obj.put("isDefault", s.isDefault)
            obj.put("isDiscovered", s.isDiscovered)
            obj.put("lastConnected", s.lastConnected)
            arr.put(obj)
        }
        @Suppress("USELESS_ELVIS")
        return arr.toString() ?: "[]"
    }

    private fun deserializeServers(json: String?): List<ServerEntry> {
        if (json.isNullOrBlank()) return emptyList()
        return try {
            val arr = JSONArray(json)
            (0 until arr.length()).map { i ->
                val obj = arr.getJSONObject(i)
                ServerEntry(
                    url = obj.getString("url"),
                    name = obj.optString("name", ""),
                    isDefault = obj.optBoolean("isDefault", false),
                    isDiscovered = obj.optBoolean("isDiscovered", false),
                    lastConnected = obj.optLong("lastConnected", 0L)
                )
            }
        } catch (e: Exception) {
            emptyList()
        }
    }
}
