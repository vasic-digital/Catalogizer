package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.launch

class SettingsViewModel(
    private val settingsRepository: SettingsRepository
) : ViewModel() {

    private val _settingsState = MutableStateFlow<Settings?>(null)
    val settingsState: StateFlow<Settings?> = _settingsState

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error

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
}
