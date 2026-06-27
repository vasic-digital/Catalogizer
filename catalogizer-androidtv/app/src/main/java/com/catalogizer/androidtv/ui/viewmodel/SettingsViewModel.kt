package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import com.catalogizer.androidtv.data.tv.LaunchAction
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/**
 * Manages TV settings state including playback, subtitle, notification, and auto-play
 * preferences. Observes [SettingsRepository.settingsFlow] and persists changes on update.
 */
class SettingsViewModel(
    private val settingsRepository: SettingsRepository
) : ViewModel() {

    private val _settingsState = MutableStateFlow<Settings?>(null)
    val settingsState: StateFlow<Settings?> = _settingsState

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error

    private val _channelLaunchActions = MutableStateFlow<Map<String, LaunchAction>>(emptyMap())
    val channelLaunchActions: StateFlow<Map<String, LaunchAction>> = _channelLaunchActions.asStateFlow()

    private val _activeMediaTypes = MutableStateFlow<List<String>>(emptyList())
    val activeMediaTypes: StateFlow<List<String>> = _activeMediaTypes.asStateFlow()

    fun loadChannelSettings(mediaRepository: MediaRepository) {
        viewModelScope.launch {
            try {
                val actions = settingsRepository.getAllLaunchActions()
                _channelLaunchActions.value = actions

                val (_, byType) = mediaRepository.getEntityStats()
                _activeMediaTypes.value = byType.filter { it.value > 0 }.keys.toList()
            } catch (e: Exception) {
                // Defaults will be used
            }
        }
    }

    fun updateLaunchAction(mediaType: String, action: LaunchAction) {
        viewModelScope.launch {
            settingsRepository.saveLaunchAction(mediaType, action)
            _channelLaunchActions.update { it + (mediaType to action) }
        }
    }

    init {
        // Observe settings changes from DataStore
        viewModelScope.launch {
            settingsRepository.settingsFlow
                .catch { e ->
                    _error.value = "Failed to load settings: ${e.message}"
                }
                .collect { settings ->
                    _settingsState.value = settings
                }
        }
    }

    fun loadSettings() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                _settingsState.value = settingsRepository.getSettingsAsync()
            } catch (e: Exception) {
                _error.value = "Failed to load settings: ${e.message}"
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun updateSettings(settings: Settings) {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                settingsRepository.saveSettings(settings)
                _settingsState.value = settings
                _error.value = null
            } catch (e: Exception) {
                _error.value = "Failed to save settings: ${e.message}"
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun updateStreamingQuality(quality: String) {
        viewModelScope.launch {
            try {
                settingsRepository.updateStreamingQuality(quality)
                _error.value = null
            } catch (e: Exception) {
                _error.value = "Failed to update streaming quality: ${e.message}"
            }
        }
    }

    fun updateSubtitleSettings(
        enableSubtitles: Boolean,
        subtitleLanguage: String
    ) {
        viewModelScope.launch {
            try {
                settingsRepository.updateSubtitles(enableSubtitles)
                settingsRepository.updateSubtitleLanguage(subtitleLanguage)
                _error.value = null
            } catch (e: Exception) {
                _error.value = "Failed to update subtitle settings: ${e.message}"
            }
        }
    }

    fun updateNotificationSettings(enableNotifications: Boolean) {
        viewModelScope.launch {
            try {
                settingsRepository.updateNotifications(enableNotifications)
                _error.value = null
            } catch (e: Exception) {
                _error.value = "Failed to update notification settings: ${e.message}"
            }
        }
    }

    fun updateAutoPlay(enableAutoPlay: Boolean) {
        viewModelScope.launch {
            try {
                settingsRepository.updateAutoPlay(enableAutoPlay)
                _error.value = null
            } catch (e: Exception) {
                _error.value = "Failed to update auto-play settings: ${e.message}"
            }
        }
    }

    fun updateAllSettings(
        enableNotifications: Boolean,
        enableAutoPlay: Boolean,
        streamingQuality: String,
        enableSubtitles: Boolean,
        subtitleLanguage: String
    ) {
        val settings = Settings(
            enableNotifications = enableNotifications,
            enableAutoPlay = enableAutoPlay,
            streamingQuality = streamingQuality,
            enableSubtitles = enableSubtitles,
            subtitleLanguage = subtitleLanguage
        )
        updateSettings(settings)
    }

    fun resetToDefaults() {
        viewModelScope.launch {
            _isLoading.value = true
            try {
                settingsRepository.resetToDefaults()
                _error.value = null
            } catch (e: Exception) {
                _error.value = "Failed to reset settings: ${e.message}"
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun clearError() {
        _error.value = null
    }

    /** Trigger a test crash via Crashlytics to verify the crash pipeline. */
    fun triggerTestCrash() {
        // Crashlytics captures exactly this: an unhandled RuntimeException
        // wrapping the intentional-crash context so the stack trace is
        // clearly identifiable in the Firebase console.
        throw RuntimeException(
            "CRASHLYTICS_TEST_CRASH — triggered from Settings > Test Crash. " +
            "This is an intentional crash to verify Firebase Crashlytics is working."
        )
    }

    /** Enable or disable Crashlytics automatic crash reporting. */
    fun toggleCrashlytics(enabled: Boolean) {
        com.google.firebase.crashlytics.FirebaseCrashlytics.getInstance()
            .setCrashlyticsCollectionEnabled(enabled)
        android.util.Log.i("SettingsViewModel", "Crashlytics collection: $enabled")
    }
}
