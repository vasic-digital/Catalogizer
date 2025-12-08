package com.catalogizer.androidtv.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@ExperimentalCoroutinesApi
class SettingsViewModelTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var settingsRepository: SettingsRepository
    private lateinit var viewModel: SettingsViewModel

    @Before
    fun setup() {
        settingsRepository = mockk()
        viewModel = SettingsViewModel(settingsRepository)
    }

    @Test
    fun `initial settings state should be null`() = runTest {
        val initialState = viewModel.settingsState.value
        assertNull(initialState)
    }

    @Test
    fun `loadSettings should update state with repository settings`() = runTest {
        val expectedSettings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "High",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        every { settingsRepository.getSettings() } returns expectedSettings

        viewModel.loadSettings()
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertEquals(expectedSettings, settingsState)
        verify { settingsRepository.getSettings() }
    }

    @Test
    fun `updateSettings should save to repository and update state`() = runTest {
        val newSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "Low",
            enableSubtitles = false,
            subtitleLanguage = "Spanish"
        )

        viewModel.updateSettings(newSettings)
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertEquals(newSettings, settingsState)
        verify { settingsRepository.saveSettings(newSettings) }
    }

    @Test
    fun `updateStreamingQuality should update only streaming quality`() = runTest {
        // First set initial settings
        val initialSettings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        viewModel.updateSettings(initialSettings)
        advanceUntilIdle()

        // Update streaming quality
        viewModel.updateStreamingQuality("4K")
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals("4K", updatedSettings?.streamingQuality)
        // Other fields should remain unchanged
        assertEquals(true, updatedSettings?.enableNotifications)
        assertEquals(false, updatedSettings?.enableAutoPlay)
        assertEquals(true, updatedSettings?.enableSubtitles)
        assertEquals("English", updatedSettings?.subtitleLanguage)
    }

    @Test
    fun `updateStreamingQuality with null current settings should do nothing`() = runTest {
        // Don't set initial settings (state remains null)

        viewModel.updateStreamingQuality("4K")
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertNull(settingsState) // Should remain null
    }

    @Test
    fun `updateSubtitleSettings should update subtitle settings correctly`() = runTest {
        // First set initial settings
        val initialSettings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        viewModel.updateSettings(initialSettings)
        advanceUntilIdle()

        // Update subtitle settings
        viewModel.updateSubtitleSettings(
            enableSubtitles = false,
            subtitleLanguage = "French"
        )
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals(false, updatedSettings?.enableSubtitles)
        assertEquals("French", updatedSettings?.subtitleLanguage)
        // Other fields should remain unchanged
        assertEquals(true, updatedSettings?.enableNotifications)
        assertEquals(false, updatedSettings?.enableAutoPlay)
        assertEquals("Auto", updatedSettings?.streamingQuality)
    }

    @Test
    fun `updateSubtitleSettings with null current settings should do nothing`() = runTest {
        viewModel.updateSubtitleSettings(
            enableSubtitles = false,
            subtitleLanguage = "French"
        )
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertNull(settingsState) // Should remain null
    }

    @Test
    fun `updateNotificationSettings should update notification setting correctly`() = runTest {
        // First set initial settings with notifications enabled
        val initialSettings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        viewModel.updateSettings(initialSettings)
        advanceUntilIdle()

        // Disable notifications
        viewModel.updateNotificationSettings(false)
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals(false, updatedSettings?.enableNotifications)
        // Other fields should remain unchanged
        assertEquals(false, updatedSettings?.enableAutoPlay)
        assertEquals("Auto", updatedSettings?.streamingQuality)
        assertEquals(true, updatedSettings?.enableSubtitles)
        assertEquals("English", updatedSettings?.subtitleLanguage)
    }

    @Test
    fun `updateNotificationSettings with null current settings should do nothing`() = runTest {
        viewModel.updateNotificationSettings(false)
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertNull(settingsState) // Should remain null
    }

    @Test
    fun `updateAllSettings should update all settings at once`() = runTest {
        viewModel.updateAllSettings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "Ultra",
            enableSubtitles = false,
            subtitleLanguage = "German"
        )
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals(false, updatedSettings?.enableNotifications)
        assertEquals(true, updatedSettings?.enableAutoPlay)
        assertEquals("Ultra", updatedSettings?.streamingQuality)
        assertEquals(false, updatedSettings?.enableSubtitles)
        assertEquals("German", updatedSettings?.subtitleLanguage)
    }

    @Test
    fun `multiple updateSettings calls should work correctly`() = runTest {
        val firstSettings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Low",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        val secondSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "High",
            enableSubtitles = false,
            subtitleLanguage = "Spanish"
        )

        // First update
        viewModel.updateSettings(firstSettings)
        advanceUntilIdle()

        var currentSettings = viewModel.settingsState.value
        assertEquals(firstSettings, currentSettings)

        // Second update
        viewModel.updateSettings(secondSettings)
        advanceUntilIdle()

        currentSettings = viewModel.settingsState.value
        assertEquals(secondSettings, currentSettings)

        // Verify both saves were called
        verify { settingsRepository.saveSettings(firstSettings) }
        verify { settingsRepository.saveSettings(secondSettings) }
    }

    @Test
    fun `settings state flow should emit correct values`() = runTest {
        val emittedValues = mutableListOf<Settings?>()

        val job = launch {
            viewModel.settingsState.collect { emittedValues.add(it) }
        }

        // Initial emission should be null
        assertEquals(1, emittedValues.size)
        assertNull(emittedValues[0])

        // Load settings
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )
        every { settingsRepository.getSettings() } returns settings

        viewModel.loadSettings()
        advanceUntilIdle()

        // Should have emitted the loaded settings
        assertEquals(2, emittedValues.size)
        assertEquals(settings, emittedValues[1])

        job.cancel()
    }

    @Test
    fun `copy method works correctly for settings updates`() = runTest {
        val original = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        val copied = original.copy(
            enableNotifications = false,
            streamingQuality = "4K"
        )

        assertEquals(false, copied.enableNotifications)
        assertEquals("4K", copied.streamingQuality)
        // Other fields should remain the same
        assertEquals(original.enableAutoPlay, copied.enableAutoPlay)
        assertEquals(original.enableSubtitles, copied.enableSubtitles)
        assertEquals(original.subtitleLanguage, copied.subtitleLanguage)
    }
}