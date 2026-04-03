package com.catalogizer.androidtv.data.models

import com.catalogizer.androidtv.BuildConfig

/**
 * Server entry represents a saved API endpoint.
 */
data class ServerEntry(
    val url: String,
    val name: String = "",
    val isDefault: Boolean = false,
    val isDiscovered: Boolean = false,
    val lastConnected: Long = 0L
)

/**
 * Immutable data class holding all user-configurable settings including playback,
 * subtitle preferences, server URL, discovered servers, and auto-discovery toggle.
 */
data class Settings(
    val enableNotifications: Boolean,
    val enableAutoPlay: Boolean,
    val streamingQuality: String,
    val enableSubtitles: Boolean,
    val subtitleLanguage: String,
    val serverUrl: String = DEFAULT_SERVER_URL,
    val savedServers: List<ServerEntry> = emptyList(),
    val autoDiscovery: Boolean = true,
    val lastUsername: String? = null
) {
    companion object {
        /** Empty by default — forces user to configure or discover the server URL. */
        val DEFAULT_SERVER_URL: String = BuildConfig.API_BASE_URL.ifBlank { "" }
    }
}