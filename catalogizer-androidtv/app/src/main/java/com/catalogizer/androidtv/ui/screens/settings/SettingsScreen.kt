package com.catalogizer.androidtv.ui.screens.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.selection.selectable
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.foundation.lazy.list.TvLazyColumn
import androidx.tv.foundation.lazy.list.items
import androidx.tv.material3.*
import androidx.lifecycle.Lifecycle
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.ui.viewmodel.SettingsViewModel

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun SettingsScreen(
    settingsViewModel: SettingsViewModel = androidx.lifecycle.viewmodel.compose.viewModel(),
    onNavigateBack: () -> Unit,
    onLogout: () -> Unit
) {
    val settingsState: Settings? by settingsViewModel.settingsState.collectAsStateWithLifecycle()
    
    // Settings values
    var enableNotifications by remember { mutableStateOf(true) }
    var enableAutoPlay by remember { mutableStateOf(false) }
    var streamingQuality by remember { mutableStateOf("Auto") }
    var enableSubtitles by remember { mutableStateOf(true) }
    var subtitleLanguage by remember { mutableStateOf("English") }

    LaunchedEffect(Unit) {
        // Load settings from ViewModel
        settingsViewModel.loadSettings()
    }

    LaunchedEffect(settingsState) {
        // Update local state when ViewModel state changes
        settingsState?.let { settings ->
            enableNotifications = settings.enableNotifications
            enableAutoPlay = settings.enableAutoPlay
            streamingQuality = settings.streamingQuality
            enableSubtitles = settings.enableSubtitles
            subtitleLanguage = settings.subtitleLanguage
        }
    }

    Box(modifier = Modifier.fillMaxSize()) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            verticalArrangement = Arrangement.spacedBy(24.dp)
        ) {
            // Header
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Settings",
                    style = MaterialTheme.typography.headlineMedium
                )
                Button(onClick = onNavigateBack) {
                    Text("Back")
                }
            }

            // Settings sections
            TvLazyColumn(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                // Playback Settings
                item {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        onClick = {} // Empty onClick for compatibility
                    ) {
                        Column(
                            modifier = Modifier.padding(16.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            Text(
                                text = "Playback Settings",
                                style = MaterialTheme.typography.titleMedium
                            )

                            // Auto Play
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .selectable(
                                        selected = enableAutoPlay,
                                        onClick = { enableAutoPlay = !enableAutoPlay },
                                        role = Role.Switch
                                    ),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "Auto Play Next Episode",
                                    modifier = Modifier.weight(1f)
                                )
                                Switch(
                                    checked = enableAutoPlay,
                                    onCheckedChange = { enableAutoPlay = it }
                                )
                            }

                            // Streaming Quality
                            Column(
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text(
                                    text = "Streaming Quality",
                                    style = MaterialTheme.typography.bodyMedium
                                )
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    listOf("Auto", "High", "Medium", "Low").forEach { quality ->
                                        // Use TV-compatible chip implementation
                                        Card(
                                            onClick = { 
                                                streamingQuality = quality
                                                settingsViewModel.updateStreamingQuality(quality)
                                            },
                                            colors = CardDefaults.colors(
                                                containerColor = if (streamingQuality == quality) 
                                                    MaterialTheme.colorScheme.primary 
                                                else 
                                                    MaterialTheme.colorScheme.surface
                                            )
                                        ) {
                                            Text(
                                                text = quality,
                                                modifier = Modifier.padding(8.dp),
                                                color = if (streamingQuality == quality) 
                                                    MaterialTheme.colorScheme.onPrimary 
                                                else 
                                                    MaterialTheme.colorScheme.onSurface
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                // Subtitle Settings
                item {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        onClick = {} // Empty onClick for compatibility
                    ) {
                        Column(
                            modifier = Modifier.padding(16.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            Text(
                                text = "Subtitle Settings",
                                style = MaterialTheme.typography.titleMedium
                            )

                            // Enable Subtitles
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .selectable(
                                        selected = enableSubtitles,
                                        onClick = { enableSubtitles = !enableSubtitles },
                                        role = Role.Switch
                                    ),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "Enable Subtitles",
                                    modifier = Modifier.weight(1f)
                                )
                                Switch(
                                    checked = enableSubtitles,
                                    onCheckedChange = { 
                                        enableSubtitles = it
                                        settingsViewModel.updateSubtitleSettings(it, subtitleLanguage)
                                    }
                                )
                            }

                            // Subtitle Language
                            Column(
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text(
                                    text = "Subtitle Language",
                                    style = MaterialTheme.typography.bodyMedium
                                )
                                listOf("English", "Spanish", "French", "German", "Japanese").forEach { lang ->
                                    Row(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .selectable(
                                                selected = subtitleLanguage == lang,
                                                onClick = { 
                                                    subtitleLanguage = lang
                                                    settingsViewModel.updateSubtitleSettings(enableSubtitles, lang)
                                                },
                                                role = Role.RadioButton
                                            ),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        RadioButton(
                                            selected = subtitleLanguage == lang,
                                            onClick = null // Handled by Row's selectable
                                        )
                                        Spacer(modifier = Modifier.width(8.dp))
                                        Text(lang)
                                    }
                                }
                            }
                        }
                    }
                }

                // Notification Settings
                item {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        onClick = {} // Empty onClick for compatibility
                    ) {
                        Column(
                            modifier = Modifier.padding(16.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            Text(
                                text = "Notifications",
                                style = MaterialTheme.typography.titleMedium
                            )

                            // Enable Notifications
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .selectable(
                                        selected = enableNotifications,
                                        onClick = { 
                                            enableNotifications = !enableNotifications
                                            settingsViewModel.updateNotificationSettings(enableNotifications)
                                        },
                                        role = Role.Switch
                                    ),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "Enable Notifications",
                                    modifier = Modifier.weight(1f)
                                )
                                Switch(
                                    checked = enableNotifications,
                                    onCheckedChange = { enableNotifications = it }
                                )
                            }
                        }
                    }
                }

                // Account Actions
                item {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        onClick = {} // Empty onClick for compatibility
                    ) {
                        Column(
                            modifier = Modifier.padding(16.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            Text(
                                text = "Account",
                                style = MaterialTheme.typography.titleMedium
                            )

                            Button(
                                onClick = onLogout,
                                modifier = Modifier.fillMaxWidth(),
                                colors = ButtonDefaults.colors(
                                    containerColor = MaterialTheme.colorScheme.error
                                )
                            ) {
                                Text("Logout")
                            }
                        }
                    }
                }
            }
        }
    }

    // Save settings when they change
    LaunchedEffect(enableNotifications, enableAutoPlay, streamingQuality, enableSubtitles, subtitleLanguage) {
        settingsViewModel.updateAllSettings(
            enableNotifications = enableNotifications,
            enableAutoPlay = enableAutoPlay,
            streamingQuality = streamingQuality,
            enableSubtitles = enableSubtitles,
            subtitleLanguage = subtitleLanguage
        )
    }
}

// Simple data class for settings
data class UserSettings(
    val enableNotifications: Boolean = true,
    val enableAutoPlay: Boolean = false,
    val streamingQuality: String = "Auto",
    val enableSubtitles: Boolean = true,
    val subtitleLanguage: String = "English"
)

// Simple ViewModel for settings
class SettingsViewModel : androidx.lifecycle.ViewModel() {
    private val _settingsState = mutableStateOf<UserSettings?>(null)
    val settingsState = _settingsState

    fun loadSettings() {
        // TODO: Load from SharedPreferences or repository
        _settingsState.value = UserSettings()
    }

    fun updateAllSettings(
        enableNotifications: Boolean,
        enableAutoPlay: Boolean,
        streamingQuality: String,
        enableSubtitles: Boolean,
        subtitleLanguage: String
    ) {
        val newSettings = UserSettings(
            enableNotifications = enableNotifications,
            enableAutoPlay = enableAutoPlay,
            streamingQuality = streamingQuality,
            enableSubtitles = enableSubtitles,
            subtitleLanguage = subtitleLanguage
        )
        _settingsState.value = newSettings
        
        // TODO: Save to SharedPreferences or repository
    }

    fun updateNotificationSettings(enabled: Boolean) {
        val current = _settingsState.value ?: return
        _settingsState.value = current.copy(enableNotifications = enabled)
        // TODO: Save to SharedPreferences
    }

    fun updateStreamingQuality(quality: String) {
        val current = _settingsState.value ?: return
        _settingsState.value = current.copy(streamingQuality = quality)
        // TODO: Save to SharedPreferences
    }

    fun updateSubtitleSettings(enabled: Boolean, language: String) {
        val current = _settingsState.value ?: return
        _settingsState.value = current.copy(
            enableSubtitles = enabled,
            subtitleLanguage = language
        )
        // TODO: Save to SharedPreferences
    }
}