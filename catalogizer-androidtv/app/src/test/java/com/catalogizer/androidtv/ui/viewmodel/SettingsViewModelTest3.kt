package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SettingsViewModelTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockSettingsRepository: SettingsRepository
    private lateinit var viewModel: SettingsViewModel

    private val defaultSettings = Settings(
        enableNotifications = true,
        enableAutoPlay = false,
        streamingQuality = "Auto",
        enableSubtitles = true,
        subtitleLanguage = "English"
    )

    @Before
    fun setup() {
        mockSettingsRepository = mockk(relaxed = true)
        every { mockSettingsRepository.settingsFlow } returns flowOf(defaultSettings)
        coEvery { mockSettingsRepository.getSettingsAsync() } returns defaultSettings
        viewModel = SettingsViewModel(mockSettingsRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial settings state collects from flow`() = runTest {
        advanceUntilIdle()
        val state = viewModel.settingsState.value
        assertNotNull(state)
        assertTrue(state!!.enableNotifications)
        assertEquals("Auto", state.streamingQuality)
    }

    @Test
    fun `initial isLoading is false`() {
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `initial error is null`() {
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadSettings fetches from repository`() = runTest {
        viewModel.loadSettings()
        advanceUntilIdle()

        assertNotNull(viewModel.settingsState.value)
        assertFalse(viewModel.isLoading.value)
        coVerify { mockSettingsRepository.getSettingsAsync() }
    }

    @Test
    fun `loadSettings sets error on failure`() = runTest {
        coEvery { mockSettingsRepository.getSettingsAsync() } throws RuntimeException("IO error")

        viewModel.loadSettings()
        advanceUntilIdle()

        assertNotNull(viewModel.error.value)
        assertTrue(viewModel.error.value!!.contains("Failed to load settings"))
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `updateSettings saves to repository`() = runTest {
        val newSettings = defaultSettings.copy(streamingQuality = "1080p")

        viewModel.updateSettings(newSettings)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.saveSettings(newSettings) }
        assertEquals("1080p", viewModel.settingsState.value?.streamingQuality)
    }

    @Test
    fun `updateSettings sets error on failure`() = runTest {
        coEvery { mockSettingsRepository.saveSettings(any()) } throws RuntimeException("Save failed")

        viewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        assertNotNull(viewModel.error.value)
        assertTrue(viewModel.error.value!!.contains("Failed to save settings"))
    }

    @Test
    fun `updateStreamingQuality calls repository`() = runTest {
        viewModel.updateStreamingQuality("4K")
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateStreamingQuality("4K") }
    }

    @Test
    fun `updateSubtitleSettings calls repository`() = runTest {
        viewModel.updateSubtitleSettings(true, "French")
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateSubtitles(true) }
        coVerify { mockSettingsRepository.updateSubtitleLanguage("French") }
    }

    @Test
    fun `updateNotificationSettings calls repository`() = runTest {
        viewModel.updateNotificationSettings(false)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateNotifications(false) }
    }

    @Test
    fun `updateAutoPlay calls repository`() = runTest {
        viewModel.updateAutoPlay(true)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateAutoPlay(true) }
    }

    @Test
    fun `resetToDefaults calls repository`() = runTest {
        viewModel.resetToDefaults()
        advanceUntilIdle()

        coVerify { mockSettingsRepository.resetToDefaults() }
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `clearError sets error to null`() {
        viewModel.clearError()
        assertNull(viewModel.error.value)
    }

    @Test
    fun `updateAllSettings creates correct Settings and saves`() = runTest {
        viewModel.updateAllSettings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "720p",
            enableSubtitles = false,
            subtitleLanguage = "Spanish"
        )
        advanceUntilIdle()

        coVerify {
            mockSettingsRepository.saveSettings(match {
                !it.enableNotifications &&
                    it.enableAutoPlay &&
                    it.streamingQuality == "720p" &&
                    !it.enableSubtitles &&
                    it.subtitleLanguage == "Spanish"
            })
        }
    }
}
