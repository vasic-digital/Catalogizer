package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.repository.SettingsRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

class SettingsViewModel(
    private val settingsRepository: SettingsRepository = SettingsRepository()
) : ViewModel() {
    
    private val _settingsState = MutableStateFlow<Settings?>(null)
    val settingsState: StateFlow<Settings?> = _settingsState

    fun loadSettings() {
        viewModelScope.launch {
            _settingsState.value = settingsRepository.getSettings()
        }
    }

    fun updateSettings(settings: Settings) {
        viewModelScope.launch {
            settingsRepository.saveSettings(settings)
            _settingsState.value = settings
        }
    }
    
    fun updateStreamingQuality(quality: String) {
        val currentSettings = _settingsState.value ?: return
        val updatedSettings = currentSettings.copy(streamingQuality = quality)
        updateSettings(updatedSettings)
    }
    
    fun updateSubtitleSettings(
        enableSubtitles: Boolean,
        subtitleLanguage: String
    ) {
        val currentSettings = _settingsState.value ?: return
        val updatedSettings = currentSettings.copy(
            enableSubtitles = enableSubtitles,
            subtitleLanguage = subtitleLanguage
        )
        updateSettings(updatedSettings)
    }
    
    fun updateNotificationSettings(enableNotifications: Boolean) {
        val currentSettings = _settingsState.value ?: return
        val updatedSettings = currentSettings.copy(
            enableNotifications = enableNotifications
        )
        updateSettings(updatedSettings)
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
}