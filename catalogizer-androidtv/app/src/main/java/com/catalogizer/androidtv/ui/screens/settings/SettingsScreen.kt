package com.catalogizer.androidtv.ui.screens.settings

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.unit.dp
import androidx.compose.material3.RadioButton
import androidx.compose.material3.RadioButtonDefaults
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.foundation.lazy.list.TvLazyColumn
import androidx.tv.material3.*
import androidx.lifecycle.Lifecycle
import com.catalogizer.androidtv.data.models.Settings
import com.catalogizer.androidtv.data.tv.LaunchAction
import com.catalogizer.androidtv.ui.viewmodel.SettingsViewModel

// WCAG AA-compliant colors for TV (dark background)
private val TextPrimary = Color(0xFFFFFFFF)       // White — highest contrast for titles/labels
private val TextSecondary = Color(0xFFE0E0E0)      // Near-white — body / description text
private val TextDisabled = Color(0xFF9E9E9E)       // Medium gray — disabled state, still readable
private val FocusBorderColor = Color(0xFF9ECAFF)   // Theme primary — focus ring color

/**
 * Reusable row for a labeled toggle switch with:
 *  - High-contrast label text (HELIX-002, 004, 008, 012)
 *  - Descriptive subtitle (HELIX-001, 007, 010, 016)
 *  - Visible focus border indicator (HELIX-005, 015)
 *  - Consistent alignment (HELIX-003, 006)
 *  - Disabled visual feedback with explanation (HELIX-014)
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun ToggleRow(
    label: String,
    description: String,
    checked: Boolean,
    onCheckedChange: (Boolean) -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    disabledReason: String? = null
) {
    var isFocused by remember { mutableStateOf(false) }

    Row(
        modifier = modifier
            .fillMaxWidth()
            .onFocusChanged { focusState -> isFocused = focusState.isFocused || focusState.hasFocus }
            .then(
                if (isFocused) Modifier.border(
                    BorderStroke(2.dp, FocusBorderColor),
                    shape = RoundedCornerShape(8.dp)
                ) else Modifier
            )
            .selectable(
                selected = checked,
                enabled = enabled,
                onClick = { if (enabled) onCheckedChange(!checked) },
                role = Role.Switch
            )
            .padding(horizontal = 12.dp, vertical = 10.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Column(modifier = Modifier.weight(1f).padding(end = 12.dp)) {
            Text(
                text = label,
                style = MaterialTheme.typography.bodyLarge,  // 18sp — large enough for TV (HELIX-009)
                color = if (enabled) TextPrimary else TextDisabled
            )
            Spacer(modifier = Modifier.height(2.dp))
            Text(
                text = if (!enabled && disabledReason != null) disabledReason else description,
                style = MaterialTheme.typography.bodySmall, // 14sp body description
                color = if (enabled) TextSecondary else TextDisabled
            )
        }
        Switch(
            checked = checked,
            onCheckedChange = if (enabled) onCheckedChange else null,
            colors = SwitchDefaults.colors(
                checkedThumbColor = TextPrimary,
                checkedTrackColor = FocusBorderColor,
                uncheckedThumbColor = TextDisabled,
                uncheckedTrackColor = Color(0xFF616161)
            )
        )
    }
}

/**
 * Section container with a visible section title for context (HELIX-011).
 * Uses a non-interactive Box instead of Card so that inner elements (toggles,
 * radio buttons, chips) are individually focusable via DPAD navigation.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun SettingsSection(
    title: String,
    modifier: Modifier = Modifier,
    content: @Composable ColumnScope.() -> Unit
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .background(
                color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
                shape = RoundedCornerShape(12.dp)
            )
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp)
    ) {
        // Section title — white, bold, clearly visible (HELIX-011)
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium, // 20sp
            color = TextPrimary
        )
        Spacer(modifier = Modifier.height(8.dp))
        content()
    }
}

/**
 * TV settings screen with D-pad-accessible toggle rows, streaming quality selector,
 * subtitle language picker, notification settings, and account logout.
 * All controls feature WCAG AA-compliant contrast and visible focus indicators.
 */
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

    val context = androidx.compose.ui.platform.LocalContext.current

    LaunchedEffect(Unit) {
        settingsViewModel.loadSettings()
        try {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
            settingsViewModel.loadChannelSettings(container.mediaRepository)
        } catch (_: Exception) {}
    }

    LaunchedEffect(settingsState) {
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
            // Header — section title ensures users know what screen they are on (HELIX-011)
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Settings",
                    style = MaterialTheme.typography.headlineMedium,
                    color = TextPrimary
                )
                var backFocused by remember { mutableStateOf(false) }
                Button(
                    onClick = onNavigateBack,
                    modifier = Modifier
                        .onFocusChanged { backFocused = it.isFocused }
                        .then(
                            if (backFocused) Modifier.border(
                                BorderStroke(2.dp, FocusBorderColor),
                                shape = RoundedCornerShape(8.dp)
                            ) else Modifier
                        )
                ) {
                    Text("Back", color = TextPrimary)
                }
            }

            TvLazyColumn(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                // ── Playback Settings ────────────────────────────────────────────
                item {
                    SettingsSection(title = "Playback Settings") {
                        // Auto Play toggle
                        ToggleRow(
                            label = "Auto Play Next Episode",
                            description = "Automatically start the next episode when the current one ends.",
                            checked = enableAutoPlay,
                            onCheckedChange = { enableAutoPlay = it }
                        )

                        Spacer(modifier = Modifier.height(12.dp))

                        // Streaming Quality selector
                        Column(modifier = Modifier.fillMaxWidth()) {
                            Text(
                                text = "Streaming Quality",
                                style = MaterialTheme.typography.bodyLarge, // 18sp (HELIX-009)
                                color = TextPrimary
                            )
                            Spacer(modifier = Modifier.height(2.dp))
                            Text(
                                text = "Choose the video quality for streaming. 'Auto' adjusts based on your connection speed.",
                                style = MaterialTheme.typography.bodySmall,
                                color = TextSecondary
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                listOf("Auto", "High", "Medium", "Low").forEach { quality ->
                                    var chipFocused by remember { mutableStateOf(false) }
                                    val isSelected = streamingQuality == quality
                                    Card(
                                        onClick = {
                                            streamingQuality = quality
                                            settingsViewModel.updateStreamingQuality(quality)
                                        },
                                        modifier = Modifier
                                            .onFocusChanged { chipFocused = it.isFocused }
                                            .then(
                                                if (chipFocused) Modifier.border(
                                                    BorderStroke(2.dp, FocusBorderColor),
                                                    shape = RoundedCornerShape(8.dp)
                                                ) else Modifier
                                            ),
                                        colors = CardDefaults.colors(
                                            containerColor = if (isSelected)
                                                MaterialTheme.colorScheme.primary
                                            else
                                                MaterialTheme.colorScheme.surface
                                        )
                                    ) {
                                        Text(
                                            text = quality,
                                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                                            style = MaterialTheme.typography.bodyMedium, // 16sp (HELIX-009)
                                            color = if (isSelected) TextPrimary else TextSecondary
                                        )
                                    }
                                }
                            }
                        }
                    }
                }

                // ── Subtitle Settings ────────────────────────────────────────────
                item {
                    SettingsSection(title = "Subtitle Settings") {
                        // Enable Subtitles toggle
                        ToggleRow(
                            label = "Enable Subtitles",
                            description = "Show subtitles during video playback when available.",
                            checked = enableSubtitles,
                            onCheckedChange = {
                                enableSubtitles = it
                                settingsViewModel.updateSubtitleSettings(it, subtitleLanguage)
                            }
                        )

                        Spacer(modifier = Modifier.height(12.dp))

                        // Subtitle Language
                        Column(modifier = Modifier.fillMaxWidth()) {
                            Text(
                                text = "Subtitle Language",
                                style = MaterialTheme.typography.bodyLarge, // 18sp (HELIX-009)
                                color = if (enableSubtitles) TextPrimary else TextDisabled
                            )
                            Spacer(modifier = Modifier.height(2.dp))
                            Text(
                                text = "Select the preferred language for subtitles. Only applies when subtitles are enabled.",
                                style = MaterialTheme.typography.bodySmall,
                                color = if (enableSubtitles) TextSecondary else TextDisabled
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            listOf("English", "Spanish", "French", "German", "Japanese").forEach { lang ->
                                var rowFocused by remember { mutableStateOf(false) }
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .onFocusChanged { rowFocused = it.isFocused || it.hasFocus }
                                        .then(
                                            if (rowFocused) Modifier.border(
                                                BorderStroke(2.dp, FocusBorderColor),
                                                shape = RoundedCornerShape(6.dp)
                                            ) else Modifier
                                        )
                                        .selectable(
                                            selected = subtitleLanguage == lang,
                                            enabled = enableSubtitles,
                                            onClick = {
                                                if (enableSubtitles) {
                                                    subtitleLanguage = lang
                                                    settingsViewModel.updateSubtitleSettings(enableSubtitles, lang)
                                                }
                                            },
                                            role = Role.RadioButton
                                        )
                                        .padding(vertical = 6.dp, horizontal = 4.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    RadioButton(
                                        selected = subtitleLanguage == lang,
                                        onClick = null,
                                        colors = RadioButtonDefaults.colors(
                                            selectedColor = FocusBorderColor,
                                            unselectedColor = if (enableSubtitles) TextSecondary else TextDisabled
                                        )
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text(
                                        text = lang,
                                        style = MaterialTheme.typography.bodyMedium, // 16sp (HELIX-009)
                                        color = if (enableSubtitles) TextPrimary else TextDisabled
                                    )
                                }
                            }
                        }
                    }
                }

                // ── Notification Settings ────────────────────────────────────────
                item {
                    SettingsSection(title = "Notifications") {
                        ToggleRow(
                            label = "Enable Notifications",
                            description = "Receive alerts for new content, sync status, and system updates.",
                            checked = enableNotifications,
                            onCheckedChange = {
                                enableNotifications = it
                                settingsViewModel.updateNotificationSettings(it)
                            }
                        )
                    }
                }

                // ── Channel Tap Behavior ────────────────────────────────────────
                item {
                    val activeMediaTypes by settingsViewModel.activeMediaTypes.collectAsStateWithLifecycle()
                    val launchActions by settingsViewModel.channelLaunchActions.collectAsStateWithLifecycle()

                    SettingsSection(title = "Channel Tap Behavior") {
                        Text(
                            text = "Choose what happens when you select an item from a home screen channel.",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary,
                            modifier = Modifier.padding(bottom = 8.dp)
                        )
                        ChannelSettingsSection(
                            activeMediaTypes = activeMediaTypes,
                            launchActions = launchActions,
                            onUpdateAction = { type, action ->
                                settingsViewModel.updateLaunchAction(type, action)
                            }
                        )
                    }
                }

                // ── Account ──────────────────────────────────────────────────────
                item {
                    SettingsSection(title = "Account") {
                        Text(
                            text = "Sign out of your Catalogizer account. Your settings and collection data will remain on the server.",
                            style = MaterialTheme.typography.bodySmall,
                            color = TextSecondary,
                            modifier = Modifier.padding(bottom = 12.dp)
                        )
                        var logoutFocused by remember { mutableStateOf(false) }
                        Button(
                            onClick = onLogout,
                            modifier = Modifier
                                .fillMaxWidth()
                                .onFocusChanged { logoutFocused = it.isFocused }
                                .then(
                                    if (logoutFocused) Modifier.border(
                                        BorderStroke(2.dp, Color(0xFFFFB4AB)), // error-tone focus ring
                                        shape = RoundedCornerShape(8.dp)
                                    ) else Modifier
                                ),
                            colors = ButtonDefaults.colors(
                                containerColor = MaterialTheme.colorScheme.error
                            )
                        ) {
                            Text("Logout", color = TextPrimary)
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
