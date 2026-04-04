package com.catalogizer.androidtv.ui.screens.settings

import androidx.compose.foundation.BorderStroke
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
import androidx.tv.material3.*
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.tv.LaunchAction

private val TextPrimary = Color(0xFFFFFFFF)
private val TextSecondary = Color(0xFFE0E0E0)
private val FocusBorderColor = Color(0xFF9ECAFF)

/**
 * Settings section that lets users configure per-media-type behavior when clicking
 * a program card on the Android TV home screen channel: open detail or play immediately.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun ChannelSettingsSection(
    activeMediaTypes: List<String>,
    launchActions: Map<String, LaunchAction>,
    onUpdateAction: (String, LaunchAction) -> Unit
) {
    if (activeMediaTypes.isEmpty()) return

    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        for (typeValue in activeMediaTypes) {
            val mediaType = MediaType.fromValue(typeValue)
            val currentAction = launchActions[typeValue] ?: LaunchAction.DETAIL

            Column(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp)) {
                Text(
                    text = mediaType.displayName,
                    style = MaterialTheme.typography.bodyLarge,
                    color = TextPrimary
                )
                Spacer(modifier = Modifier.height(4.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    LaunchAction.values().forEach { action ->
                        val label = when (action) {
                            LaunchAction.DETAIL -> "Detail Screen"
                            LaunchAction.IMMEDIATE_PLAY -> "Play Immediately"
                        }
                        val isSelected = currentAction == action
                        var rowFocused by remember { mutableStateOf(false) }

                        Row(
                            modifier = Modifier
                                .onFocusChanged { rowFocused = it.isFocused || it.hasFocus }
                                .then(
                                    if (rowFocused) Modifier.border(
                                        BorderStroke(2.dp, FocusBorderColor),
                                        shape = RoundedCornerShape(6.dp)
                                    ) else Modifier
                                )
                                .selectable(
                                    selected = isSelected,
                                    onClick = { onUpdateAction(typeValue, action) },
                                    role = Role.RadioButton
                                )
                                .padding(vertical = 4.dp, horizontal = 4.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            RadioButton(
                                selected = isSelected,
                                onClick = null,
                                colors = RadioButtonDefaults.colors(
                                    selectedColor = FocusBorderColor,
                                    unselectedColor = TextSecondary
                                )
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                text = label,
                                style = MaterialTheme.typography.bodyMedium,
                                color = if (isSelected) TextPrimary else TextSecondary
                            )
                        }
                    }
                }
            }
        }
    }
}
