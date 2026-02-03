package com.catalogizer.androidtv.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import com.catalogizer.androidtv.data.models.Settings
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder

@OptIn(ExperimentalCoroutinesApi::class)
class SettingsRepositoryTest {

    @get:Rule
    val temporaryFolder = TemporaryFolder()

    private val testDispatcher = StandardTestDispatcher()
    private val testScope = TestScope(testDispatcher)

    private lateinit var dataStore: DataStore<Preferences>
    private lateinit var repository: SettingsRepository
    private lateinit var dataStoreScope: CoroutineScope

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        dataStoreScope = CoroutineScope(testDispatcher + Job())
        dataStore = PreferenceDataStoreFactory.create(
            scope = dataStoreScope,
            produceFile = { temporaryFolder.newFile("test_settings.preferences_pb") }
        )
        repository = SettingsRepository(dataStore)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
        dataStoreScope.cancel()
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
    fun `saveSettings updates current settings`() = testScope.runTest {
        val newSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "High",
            enableSubtitles = false,
            subtitleLanguage = "Spanish"
        )

        repository.saveSettings(newSettings)
        val retrievedSettings = repository.getSettingsAsync()

        assertEquals(newSettings, retrievedSettings)
        assertFalse(retrievedSettings.enableNotifications)
        assertTrue(retrievedSettings.enableAutoPlay)
        assertEquals("High", retrievedSettings.streamingQuality)
        assertFalse(retrievedSettings.enableSubtitles)
        assertEquals("Spanish", retrievedSettings.subtitleLanguage)
    }

    @Test
    fun `settingsFlow emits updated settings`() = testScope.runTest {
        val newSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "4K",
            enableSubtitles = false,
            subtitleLanguage = "French"
        )

        repository.saveSettings(newSettings)

        val emittedSettings = repository.settingsFlow.first()
        assertEquals(newSettings, emittedSettings)
    }

    @Test
    fun `updateNotifications persists change`() = testScope.runTest {
        repository.updateNotifications(false)

        val settings = repository.getSettingsAsync()
        assertFalse(settings.enableNotifications)
    }

    @Test
    fun `updateAutoPlay persists change`() = testScope.runTest {
        repository.updateAutoPlay(true)

        val settings = repository.getSettingsAsync()
        assertTrue(settings.enableAutoPlay)
    }

    @Test
    fun `updateStreamingQuality persists change`() = testScope.runTest {
        repository.updateStreamingQuality("4K")

        val settings = repository.getSettingsAsync()
        assertEquals("4K", settings.streamingQuality)
    }

    @Test
    fun `updateSubtitles persists change`() = testScope.runTest {
        repository.updateSubtitles(false)

        val settings = repository.getSettingsAsync()
        assertFalse(settings.enableSubtitles)
    }

    @Test
    fun `updateSubtitleLanguage persists change`() = testScope.runTest {
        repository.updateSubtitleLanguage("Japanese")

        val settings = repository.getSettingsAsync()
        assertEquals("Japanese", settings.subtitleLanguage)
    }

    @Test
    fun `resetToDefaults restores default settings`() = testScope.runTest {
        // First change settings
        val customSettings = Settings(
            enableNotifications = false,
            enableAutoPlay = true,
            streamingQuality = "4K",
            enableSubtitles = false,
            subtitleLanguage = "German"
        )
        repository.saveSettings(customSettings)

        // Verify custom settings
        var settings = repository.getSettingsAsync()
        assertEquals(customSettings, settings)

        // Reset to defaults
        repository.resetToDefaults()

        // Verify defaults restored
        settings = repository.getSettingsAsync()
        assertTrue(settings.enableNotifications)
        assertFalse(settings.enableAutoPlay)
        assertEquals("Auto", settings.streamingQuality)
        assertTrue(settings.enableSubtitles)
        assertEquals("English", settings.subtitleLanguage)
    }

    @Test
    fun `multiple sequential updates persist correctly`() = testScope.runTest {
        repository.updateNotifications(false)
        repository.updateAutoPlay(true)
        repository.updateStreamingQuality("High")

        val settings = repository.getSettingsAsync()
        assertFalse(settings.enableNotifications)
        assertTrue(settings.enableAutoPlay)
        assertEquals("High", settings.streamingQuality)
        // Other settings should retain defaults
        assertTrue(settings.enableSubtitles)
        assertEquals("English", settings.subtitleLanguage)
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
