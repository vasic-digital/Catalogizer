package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class SettingsTest2 {

    @Test
    fun `ServerEntry has correct defaults`() {
        val entry = ServerEntry(url = "http://localhost:8080")
        assertEquals("http://localhost:8080", entry.url)
        assertEquals("", entry.name)
        assertFalse(entry.isDefault)
        assertFalse(entry.isDiscovered)
        assertEquals(0L, entry.lastConnected)
    }

    @Test
    fun `ServerEntry with all fields`() {
        val entry = ServerEntry(
            url = "http://192.168.0.100:8080",
            name = "Home Server",
            isDefault = true,
            isDiscovered = true,
            lastConnected = 1735689600000L
        )
        assertEquals("http://192.168.0.100:8080", entry.url)
        assertEquals("Home Server", entry.name)
        assertTrue(entry.isDefault)
        assertTrue(entry.isDiscovered)
        assertEquals(1735689600000L, entry.lastConnected)
    }

    @Test
    fun `Settings has correct field values`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "1080p",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        assertTrue(settings.enableNotifications)
        assertFalse(settings.enableAutoPlay)
        assertEquals("1080p", settings.streamingQuality)
        assertTrue(settings.enableSubtitles)
        assertEquals("English", settings.subtitleLanguage)
    }

    @Test
    fun `Settings has default server URL`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        // Default server URL comes from BuildConfig - in tests it will be empty
        assertNotNull(settings.serverUrl)
    }

    @Test
    fun `Settings with custom server URL`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English",
            serverUrl = "http://myserver.local:8080"
        )
        assertEquals("http://myserver.local:8080", settings.serverUrl)
    }

    @Test
    fun `Settings with saved servers`() {
        val servers = listOf(
            ServerEntry(url = "http://server1:8080", name = "Server 1"),
            ServerEntry(url = "http://server2:8080", name = "Server 2")
        )
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English",
            savedServers = servers
        )
        assertEquals(2, settings.savedServers.size)
    }

    @Test
    fun `Settings default savedServers is empty`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        assertTrue(settings.savedServers.isEmpty())
    }

    @Test
    fun `Settings autoDiscovery defaults to true`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        assertTrue(settings.autoDiscovery)
    }

    @Test
    fun `Settings lastUsername defaults to null`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        assertNull(settings.lastUsername)
    }

    @Test
    fun `Settings with lastUsername set`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English",
            lastUsername = "admin"
        )
        assertEquals("admin", settings.lastUsername)
    }
}
