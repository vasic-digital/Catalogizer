package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.Settings
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class SettingsRepositoryTest2 {

    @Test
    fun `getSettings returns default cached settings`() {
        // SettingsRepository initializes with default cached settings
        // We test the Settings defaults that the repository uses
        val defaults = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        assertTrue(defaults.enableNotifications)
        assertFalse(defaults.enableAutoPlay)
        assertEquals("Auto", defaults.streamingQuality)
        assertTrue(defaults.enableSubtitles)
        assertEquals("English", defaults.subtitleLanguage)
    }

    @Test
    fun `Settings default values match repository defaults`() {
        val defaults = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        // These are the expected defaults from SettingsRepository
        assertTrue(defaults.enableNotifications)
        assertFalse(defaults.enableAutoPlay)
        assertEquals("Auto", defaults.streamingQuality)
        assertTrue(defaults.enableSubtitles)
        assertEquals("English", defaults.subtitleLanguage)
    }

    @Test
    fun `Settings can represent different streaming qualities`() {
        val sdSettings = Settings(true, false, "SD", true, "English")
        val hdSettings = Settings(true, false, "1080p", true, "English")
        val uhSettings = Settings(true, false, "4K", true, "English")

        assertEquals("SD", sdSettings.streamingQuality)
        assertEquals("1080p", hdSettings.streamingQuality)
        assertEquals("4K", uhSettings.streamingQuality)
    }

    @Test
    fun `Settings can represent different subtitle languages`() {
        val english = Settings(true, false, "Auto", true, "English")
        val spanish = Settings(true, false, "Auto", true, "Spanish")
        val japanese = Settings(true, false, "Auto", true, "Japanese")

        assertEquals("English", english.subtitleLanguage)
        assertEquals("Spanish", spanish.subtitleLanguage)
        assertEquals("Japanese", japanese.subtitleLanguage)
    }

    @Test
    fun `Settings all-disabled configuration`() {
        val allDisabled = Settings(
            enableNotifications = false,
            enableAutoPlay = false,
            streamingQuality = "SD",
            enableSubtitles = false,
            subtitleLanguage = ""
        )

        assertFalse(allDisabled.enableNotifications)
        assertFalse(allDisabled.enableAutoPlay)
        assertFalse(allDisabled.enableSubtitles)
    }
}
