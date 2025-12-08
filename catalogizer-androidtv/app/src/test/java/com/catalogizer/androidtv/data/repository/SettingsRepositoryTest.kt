package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.Settings
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class SettingsRepositoryTest {

    private lateinit var repository: SettingsRepository

    @Before
    fun setup() {
        repository = SettingsRepository()
    }

    @Test
    fun `getSettings returns default settings initially`() {
        val settings = repository.getSettings()

        assertTrue(settings.enableNotifications)
        assertFalse(settings.enableAutoPlay)
        assertEquals("Auto", settings.streamingQuality)
        assertTrue(settings.enableSubtitles)
        assertEquals("English", settings.subtitleLanguage)
    }

    @Test
    fun `saveSettings updates current settings`() {
        val newSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "High",
            enableSubtitles = false,
            subtitleLanguage = "Spanish"
        )

        repository.saveSettings(newSettings)
        val retrievedSettings = repository.getSettings()

        assertEquals(newSettings, retrievedSettings)
        assertFalse(retrievedSettings.enableNotifications)
        assertTrue(retrievedSettings.enableAutoPlay)
        assertEquals("High", retrievedSettings.streamingQuality)
        assertFalse(retrievedSettings.enableSubtitles)
        assertEquals("Spanish", retrievedSettings.subtitleLanguage)
    }

    @Test
    fun `saveSettings followed by getSettings returns saved settings`() {
        val originalSettings = repository.getSettings()
        assertTrue(originalSettings.enableNotifications)

        val updatedSettings = originalSettings.copy(enableNotifications = false)
        repository.saveSettings(updatedSettings)

        val retrievedSettings = repository.getSettings()
        assertFalse(retrievedSettings.enableNotifications)
        assertEquals(updatedSettings, retrievedSettings)
    }

    @Test
    fun `multiple saveSettings calls update settings correctly`() {
        // First save
        val firstSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "Low",
            enableSubtitles = true,
            subtitleLanguage = "French"
        )
        repository.saveSettings(firstSettings)

        var retrieved = repository.getSettings()
        assertEquals(firstSettings, retrieved)

        // Second save
        val secondSettings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Ultra",
            enableSubtitles = false,
            subtitleLanguage = "German"
        )
        repository.saveSettings(secondSettings)

        retrieved = repository.getSettings()
        assertEquals(secondSettings, retrieved)
        assertTrue(retrieved.enableNotifications)
        assertFalse(retrieved.enableAutoPlay)
        assertEquals("Ultra", retrieved.streamingQuality)
        assertFalse(retrieved.enableSubtitles)
        assertEquals("German", retrieved.subtitleLanguage)
    }

    @Test
    fun `getSettings returns copy not reference to internal state`() {
        val retrieved1 = repository.getSettings()
        val retrieved2 = repository.getSettings()

        // They should be equal but not the same instance
        assertEquals(retrieved1, retrieved2)

        // Modifying one shouldn't affect the other (since data classes are immutable)
        // But let's verify the repository's internal state isn't affected
        val newSettings = retrieved1.copy(enableNotifications = false)
        repository.saveSettings(newSettings)

        val retrieved3 = repository.getSettings()
        assertFalse(retrieved3.enableNotifications)
        assertTrue(retrieved1.enableNotifications) // Original should still be true
    }

    @Test
    fun `settings with same values are equal`() {
        val settings1 = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        val settings2 = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        assertEquals(settings1, settings2)
    }

    @Test
    fun `settings with different values are not equal`() {
        val settings1 = repository.getSettings()
        val settings2 = settings1.copy(enableNotifications = false)

        assertFalse(settings1 == settings2)
    }

    @Test
    fun `copy method works correctly for settings modification`() {
        val original = repository.getSettings()
        val modified = original.copy(
            enableNotifications = false,
            streamingQuality = "4K"
        )

        assertTrue(original.enableNotifications)
        assertEquals("Auto", original.streamingQuality)

        assertFalse(modified.enableNotifications)
        assertEquals("4K", modified.streamingQuality)

        // Other fields should remain the same
        assertEquals(original.enableAutoPlay, modified.enableAutoPlay)
        assertEquals(original.enableSubtitles, modified.enableSubtitles)
        assertEquals(original.subtitleLanguage, modified.subtitleLanguage)
    }
}