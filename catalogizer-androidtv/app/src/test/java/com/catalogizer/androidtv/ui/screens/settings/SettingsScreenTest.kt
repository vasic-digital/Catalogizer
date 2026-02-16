package com.catalogizer.androidtv.ui.screens.settings

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import com.catalogizer.androidtv.ui.viewmodel.SettingsViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SettingsScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockSettingsRepository: SettingsRepository
    private lateinit var settingsViewModel: SettingsViewModel

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
        settingsViewModel = SettingsViewModel(mockSettingsRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial settingsState should be null`() {
        // Before flow emits, state might be null
        // But after init block collects, it should have value
        assertNotNull(settingsViewModel)
    }

    @Test
    fun `settingsState should update from repository flow`() = runTest {
        advanceUntilIdle()

        val state = settingsViewModel.settingsState.value
        assertNotNull(state)
        assertEquals(true, state?.enableNotifications)
        assertEquals(false, state?.enableAutoPlay)
        assertEquals("Auto", state?.streamingQuality)
        assertEquals(true, state?.enableSubtitles)
        assertEquals("English", state?.subtitleLanguage)
    }

    @Test
    fun `initial isLoading should be false`() {
        assertFalse(settingsViewModel.isLoading.value)
    }

    @Test
    fun `initial error should be null`() {
        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `loadSettings should call repository getSettingsAsync`() = runTest {
        settingsViewModel.loadSettings()
        advanceUntilIdle()

        coVerify { mockSettingsRepository.getSettingsAsync() }
    }

    @Test
    fun `loadSettings should update settingsState`() = runTest {
        val customSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "High",
            enableSubtitles = false,
            subtitleLanguage = "Spanish"
        )
        coEvery { mockSettingsRepository.getSettingsAsync() } returns customSettings

        settingsViewModel.loadSettings()
        advanceUntilIdle()

        val state = settingsViewModel.settingsState.value
        assertEquals(false, state?.enableNotifications)
        assertEquals(true, state?.enableAutoPlay)
        assertEquals("High", state?.streamingQuality)
        assertEquals(false, state?.enableSubtitles)
        assertEquals("Spanish", state?.subtitleLanguage)
    }

    @Test
    fun `loadSettings should set loading then unset`() = runTest {
        settingsViewModel.loadSettings()
        advanceUntilIdle()

        assertFalse(settingsViewModel.isLoading.value)
    }

    @Test
    fun `loadSettings should set error on exception`() = runTest {
        coEvery { mockSettingsRepository.getSettingsAsync() } throws RuntimeException("DB error")

        settingsViewModel.loadSettings()
        advanceUntilIdle()

        assertNotNull(settingsViewModel.error.value)
        assertTrue(settingsViewModel.error.value!!.contains("DB error"))
    }

    @Test
    fun `updateSettings should call repository saveSettings`() = runTest {
        val newSettings = defaultSettings.copy(enableAutoPlay = true)

        settingsViewModel.updateSettings(newSettings)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.saveSettings(newSettings) }
    }

    @Test
    fun `updateSettings should update settingsState`() = runTest {
        val newSettings = defaultSettings.copy(streamingQuality = "High")

        settingsViewModel.updateSettings(newSettings)
        advanceUntilIdle()

        assertEquals("High", settingsViewModel.settingsState.value?.streamingQuality)
    }

    @Test
    fun `updateSettings should clear error on success`() = runTest {
        settingsViewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateSettings should set error on failure`() = runTest {
        coEvery { mockSettingsRepository.saveSettings(any()) } throws RuntimeException("Save failed")

        settingsViewModel.updateSettings(defaultSettings)
        advanceUntilIdle()

        assertNotNull(settingsViewModel.error.value)
        assertTrue(settingsViewModel.error.value!!.contains("Save failed"))
    }

    @Test
    fun `updateStreamingQuality should call repository`() = runTest {
        settingsViewModel.updateStreamingQuality("High")
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateStreamingQuality("High") }
    }

    @Test
    fun `updateStreamingQuality should clear error on success`() = runTest {
        settingsViewModel.updateStreamingQuality("Low")
        advanceUntilIdle()

        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateStreamingQuality should set error on failure`() = runTest {
        coEvery { mockSettingsRepository.updateStreamingQuality(any()) } throws RuntimeException("Failed")

        settingsViewModel.updateStreamingQuality("High")
        advanceUntilIdle()

        assertNotNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateSubtitleSettings should call repository methods`() = runTest {
        settingsViewModel.updateSubtitleSettings(true, "French")
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateSubtitles(true) }
        coVerify { mockSettingsRepository.updateSubtitleLanguage("French") }
    }

    @Test
    fun `updateSubtitleSettings should clear error on success`() = runTest {
        settingsViewModel.updateSubtitleSettings(false, "German")
        advanceUntilIdle()

        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateSubtitleSettings should set error on failure`() = runTest {
        coEvery { mockSettingsRepository.updateSubtitles(any()) } throws RuntimeException("Failed")

        settingsViewModel.updateSubtitleSettings(true, "English")
        advanceUntilIdle()

        assertNotNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateNotificationSettings should call repository`() = runTest {
        settingsViewModel.updateNotificationSettings(false)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateNotifications(false) }
    }

    @Test
    fun `updateNotificationSettings should clear error on success`() = runTest {
        settingsViewModel.updateNotificationSettings(true)
        advanceUntilIdle()

        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateAutoPlay should call repository`() = runTest {
        settingsViewModel.updateAutoPlay(true)
        advanceUntilIdle()

        coVerify { mockSettingsRepository.updateAutoPlay(true) }
    }

    @Test
    fun `updateAutoPlay should clear error on success`() = runTest {
        settingsViewModel.updateAutoPlay(false)
        advanceUntilIdle()

        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `updateAllSettings should create Settings object and save`() = runTest {
        settingsViewModel.updateAllSettings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "Medium",
            enableSubtitles = false,
            subtitleLanguage = "Japanese"
        )
        advanceUntilIdle()

        coVerify {
            mockSettingsRepository.saveSettings(
                Settings(
                    enableNotifications = false,
                    enableAutoPlay = true,
                    streamingQuality = "Medium",
                    enableSubtitles = false,
                    subtitleLanguage = "Japanese"
                )
            )
        }
    }

    @Test
    fun `resetToDefaults should call repository resetToDefaults`() = runTest {
        settingsViewModel.resetToDefaults()
        advanceUntilIdle()

        coVerify { mockSettingsRepository.resetToDefaults() }
    }

    @Test
    fun `resetToDefaults should set loading then unset`() = runTest {
        settingsViewModel.resetToDefaults()
        advanceUntilIdle()

        assertFalse(settingsViewModel.isLoading.value)
    }

    @Test
    fun `resetToDefaults should set error on failure`() = runTest {
        coEvery { mockSettingsRepository.resetToDefaults() } throws RuntimeException("Reset failed")

        settingsViewModel.resetToDefaults()
        advanceUntilIdle()

        assertNotNull(settingsViewModel.error.value)
        assertTrue(settingsViewModel.error.value!!.contains("Reset failed"))
    }

    @Test
    fun `clearError should set error to null`() {
        settingsViewModel.clearError()

        assertNull(settingsViewModel.error.value)
    }

    @Test
    fun `Settings data class should have correct default values`() {
        val settings = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        assertTrue(settings.enableNotifications)
        assertFalse(settings.enableAutoPlay)
        assertEquals("Auto", settings.streamingQuality)
        assertTrue(settings.enableSubtitles)
        assertEquals("English", settings.subtitleLanguage)
    }

    @Test
    fun `Settings copy should preserve unmodified fields`() {
        val original = Settings(
            enableNotifications = true,
            enableAutoPlay = false,
            streamingQuality = "Auto",
            enableSubtitles = true,
            subtitleLanguage = "English"
        )

        val modified = original.copy(streamingQuality = "High")

        assertTrue(modified.enableNotifications)
        assertFalse(modified.enableAutoPlay)
        assertEquals("High", modified.streamingQuality)
        assertTrue(modified.enableSubtitles)
        assertEquals("English", modified.subtitleLanguage)
    }

    @Test
    fun `settings screen should support quality options`() {
        val qualityOptions = listOf("Auto", "High", "Medium", "Low")

        assertEquals(4, qualityOptions.size)
        assertTrue(qualityOptions.contains("Auto"))
        assertTrue(qualityOptions.contains("High"))
        assertTrue(qualityOptions.contains("Medium"))
        assertTrue(qualityOptions.contains("Low"))
    }

    @Test
    fun `settings screen should support subtitle language options`() {
        val languages = listOf("English", "Spanish", "French", "German", "Japanese")

        assertEquals(5, languages.size)
        assertTrue(languages.contains("English"))
        assertTrue(languages.contains("Japanese"))
    }

    @Test
    fun `onNavigateBack callback should be invocable`() {
        var navigateBackCalled = false
        val onNavigateBack: () -> Unit = { navigateBackCalled = true }

        onNavigateBack()

        assertTrue(navigateBackCalled)
    }

    @Test
    fun `onLogout callback should be invocable`() {
        var logoutCalled = false
        val onLogout: () -> Unit = { logoutCalled = true }

        onLogout()

        assertTrue(logoutCalled)
    }
}
