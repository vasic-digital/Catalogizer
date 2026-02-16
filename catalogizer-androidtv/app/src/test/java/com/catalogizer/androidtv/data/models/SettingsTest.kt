package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class SettingsTest {

    @Test
    fun `Settings holds correct values`() {
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
    fun `Settings copy updates correctly`() {
        val original = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        val updated = original.copy(
            enableAutoPlay = true,
            streamingQuality = "4K"
        )

        assertTrue(updated.enableAutoPlay)
        assertEquals("4K", updated.streamingQuality)
        assertEquals(original.enableNotifications, updated.enableNotifications)
        assertEquals(original.enableSubtitles, updated.enableSubtitles)
    }

    @Test
    fun `Settings equality works correctly`() {
        val s1 = Settings(true, false, "Auto", true, "English")
        val s2 = Settings(true, false, "Auto", true, "English")
        val s3 = Settings(true, true, "Auto", true, "English")

        assertEquals(s1, s2)
        assertNotEquals(s1, s3)
    }

    @Test
    fun `Settings with all disabled`() {
        val settings = Settings(
            enableNotifications = false,
            enableAutoPlay = false,
            streamingQuality = "SD",
            enableSubtitles = false,
            subtitleLanguage = ""
        )

        assertFalse(settings.enableNotifications)
        assertFalse(settings.enableAutoPlay)
        assertFalse(settings.enableSubtitles)
    }

    @Test
    fun `Settings with different streaming qualities`() {
        val qualities = listOf("Auto", "SD", "720p", "1080p", "4K")
        for (quality in qualities) {
            val settings = Settings(true, false, quality, true, "English")
            assertEquals(quality, settings.streamingQuality)
        }
    }
}
