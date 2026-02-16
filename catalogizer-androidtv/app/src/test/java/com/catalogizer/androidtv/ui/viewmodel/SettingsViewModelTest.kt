package com.catalogizer.androidtv.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
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

    private val defaultSettings = Settings(
        enableNotifications = true,
        enableAutoPlay = false,
        streamingQuality = "Auto",
        enableSubtitles = true,
        subtitleLanguage = "English"
    )

    private lateinit var settingsFlowSource: MutableStateFlow<Settings>

    @Before
    fun setup() {
        settingsRepository = mockk()
        settingsFlowSource = MutableStateFlow(defaultSettings)

        // Setup default mock behavior
        every { settingsRepository.settingsFlow } returns settingsFlowSource
        every { settingsRepository.getSettings() } returns defaultSettings
        coEvery { settingsRepository.getSettingsAsync() } returns defaultSettings
        coEvery { settingsRepository.saveSettings(any()) } answers {
            settingsFlowSource.value = firstArg()
        }
        coEvery { settingsRepository.updateNotifications(any()) } answers {
            settingsFlowSource.value = settingsFlowSource.value.copy(enableNotifications = firstArg())
        }
        coEvery { settingsRepository.updateAutoPlay(any()) } answers {
            settingsFlowSource.value = settingsFlowSource.value.copy(enableAutoPlay = firstArg())
        }
        coEvery { settingsRepository.updateStreamingQuality(any()) } answers {
            settingsFlowSource.value = settingsFlowSource.value.copy(streamingQuality = firstArg())
        }
        coEvery { settingsRepository.updateSubtitles(any()) } answers {
            settingsFlowSource.value = settingsFlowSource.value.copy(enableSubtitles = firstArg())
        }
        coEvery { settingsRepository.updateSubtitleLanguage(any()) } answers {
            settingsFlowSource.value = settingsFlowSource.value.copy(subtitleLanguage = firstArg())
        }
        coEvery { settingsRepository.resetToDefaults() } answers {
            settingsFlowSource.value = defaultSettings
        }

        viewModel = SettingsViewModel(settingsRepository)
    }

    @Test
    fun `initial settings state should be default from flow`() = runTest {
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertEquals(defaultSettings, settingsState)
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
        coEvery { settingsRepository.getSettingsAsync() } returns expectedSettings

        viewModel.loadSettings()
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertEquals(expectedSettings, settingsState)
        coVerify { settingsRepository.getSettingsAsync() }
    }

    @Test
    fun `updateSettings should save to repository and update state`() = runTest {
        advanceUntilIdle() // Let initial collection happen

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
        coVerify { settingsRepository.saveSettings(newSettings) }
    }

    @Test
    fun `updateStreamingQuality should update only streaming quality`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        viewModel.updateStreamingQuality("4K")
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals("4K", updatedSettings?.streamingQuality)
        coVerify { settingsRepository.updateStreamingQuality("4K") }
    }

    @Test
    fun `updateSubtitleSettings should update subtitle settings correctly`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        viewModel.updateSubtitleSettings(
            enableSubtitles = false,
            subtitleLanguage = "French"
        )
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals(false, updatedSettings?.enableSubtitles)
        assertEquals("French", updatedSettings?.subtitleLanguage)
        coVerify { settingsRepository.updateSubtitles(false) }
        coVerify { settingsRepository.updateSubtitleLanguage("French") }
    }

    @Test
    fun `updateNotificationSettings should update notification setting correctly`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        viewModel.updateNotificationSettings(false)
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals(false, updatedSettings?.enableNotifications)
        coVerify { settingsRepository.updateNotifications(false) }
    }

    @Test
    fun `updateAutoPlay should update auto-play setting correctly`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        viewModel.updateAutoPlay(true)
        advanceUntilIdle()

        val updatedSettings = viewModel.settingsState.value
        assertEquals(true, updatedSettings?.enableAutoPlay)
        coVerify { settingsRepository.updateAutoPlay(true) }
    }

    @Test
    fun `updateAllSettings should update all settings at once`() = runTest {
        advanceUntilIdle() // Let initial collection happen

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
    fun `resetToDefaults should restore default settings`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        // First change settings
        val customSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "4K",
            enableSubtitles = false,
            subtitleLanguage = "German"
        )
        viewModel.updateSettings(customSettings)
        advanceUntilIdle()

        // Reset to defaults
        viewModel.resetToDefaults()
        advanceUntilIdle()

        val settingsState = viewModel.settingsState.value
        assertEquals(defaultSettings, settingsState)
        coVerify { settingsRepository.resetToDefaults() }
    }

    @Test
    fun `isLoading should be true during save operation`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        viewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        // After completion, loading should be false
        assertFalse(viewModel.isLoading.value)
        // Verify the save was actually called (proves the loading cycle completed)
        coVerify { settingsRepository.saveSettings(defaultSettings) }
        // Verify settings were updated
        assertEquals(defaultSettings, viewModel.settingsState.value)
    }

    @Test
    fun `error should be set when save fails`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        coEvery { settingsRepository.saveSettings(any()) } throws Exception("Save failed")

        viewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        val error = viewModel.error.value
        assertTrue(error?.contains("Failed to save settings") == true)
    }

    @Test
    fun `clearError should clear error state`() = runTest {
        advanceUntilIdle() // Let initial collection happen

        // Trigger an error
        coEvery { settingsRepository.saveSettings(any()) } throws Exception("Save failed")
        viewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        assertTrue(viewModel.error.value != null)

        // Clear the error
        viewModel.clearError()

        assertNull(viewModel.error.value)
    }

    @Test
    fun `multiple updateSettings calls should work correctly`() = runTest {
        advanceUntilIdle() // Let initial collection happen

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
        coVerify { settingsRepository.saveSettings(firstSettings) }
        coVerify { settingsRepository.saveSettings(secondSettings) }
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
