package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SettingsViewModelTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockSettingsRepository = mockk<SettingsRepository>(relaxed = true)

    private val defaultSettings = Settings(
        enableNotifications = true,
        enableAutoPlay = false,
        streamingQuality = "Auto",
        enableSubtitles = true,
        subtitleLanguage = "English"
    )

    @Before
    fun setup() {
        every { mockSettingsRepository.settingsFlow } returns flowOf(defaultSettings)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state has null settings`() = runTest {
        val viewModel = SettingsViewModel(mockSettingsRepository)
        // Before collection starts, settingsState is null
        // After init block runs and collects, it should have settings
        advanceUntilIdle()

        assertNotNull(viewModel.settingsState.value)
        assertEquals(defaultSettings, viewModel.settingsState.value)
    }

    @Test
    fun `loadSettings fetches from repository`() = runTest {
        coEvery { mockSettingsRepository.getSettingsAsync() } returns defaultSettings

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.loadSettings()
        advanceUntilIdle()

        assertEquals(defaultSettings, viewModel.settingsState.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `updateSettings saves to repository`() = runTest {
        coEvery { mockSettingsRepository.saveSettings(any()) } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        val newSettings = defaultSettings.copy(streamingQuality = "4K", enableAutoPlay = true)
        viewModel.updateSettings(newSettings)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.saveSettings(newSettings) }
        assertEquals(newSettings, viewModel.settingsState.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `updateStreamingQuality calls repository`() = runTest {
        coEvery { mockSettingsRepository.updateStreamingQuality(any()) } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.updateStreamingQuality("1080p")
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateStreamingQuality("1080p") }
        assertNull(viewModel.error.value)
    }

    @Test
    fun `updateSubtitleSettings calls repository`() = runTest {
        coEvery { mockSettingsRepository.updateSubtitles(any()) } just Runs
        coEvery { mockSettingsRepository.updateSubtitleLanguage(any()) } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.updateSubtitleSettings(true, "Spanish")
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateSubtitles(true) }
        coVerify { mockSettingsRepository.updateSubtitleLanguage("Spanish") }
    }

    @Test
    fun `updateNotificationSettings calls repository`() = runTest {
        coEvery { mockSettingsRepository.updateNotifications(any()) } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.updateNotificationSettings(false)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateNotifications(false) }
    }

    @Test
    fun `updateAutoPlay calls repository`() = runTest {
        coEvery { mockSettingsRepository.updateAutoPlay(any()) } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.updateAutoPlay(true)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateAutoPlay(true) }
    }

    @Test
    fun `resetToDefaults calls repository`() = runTest {
        coEvery { mockSettingsRepository.resetToDefaults() } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.resetToDefaults()
        advanceUntilIdle()

        coVerify { mockSettingsRepository.resetToDefaults() }
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `clearError clears error state`() = runTest {
        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.clearError()

        assertNull(viewModel.error.value)
    }

    @Test
    fun `updateSettings failure sets error`() = runTest {
        coEvery { mockSettingsRepository.saveSettings(any()) } throws RuntimeException("Save failed")

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        assertNotNull(viewModel.error.value)
        assertTrue(viewModel.error.value?.contains("Failed to save settings") == true)
    }

    @Test
    fun `updateAllSettings creates correct Settings object`() = runTest {
        coEvery { mockSettingsRepository.saveSettings(any()) } just Runs

        val viewModel = SettingsViewModel(mockSettingsRepository)
        advanceUntilIdle()

        viewModel.updateAllSettings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "4K",
            enableSubtitles = false,
            subtitleLanguage = "Japanese"
        )
        advanceUntilIdle()

        coVerify {
            mockSettingsRepository.saveSettings(match { settings ->
                !settings.enableNotifications &&
                    settings.enableAutoPlay &&
                    settings.streamingQuality == "4K" &&
                    !settings.enableSubtitles &&
                    settings.subtitleLanguage == "Japanese"
            })
        }
    }
}
