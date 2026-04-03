package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import com.catalogizer.androidtv.data.models.AuthState

// Focus ring color matches theme primary for WCAG-visible focus indicators (HELIX-005, 015)
private val FocusBorderColor = Color(0xFF9ECAFF)

/**
 * TV top bar composable with app title, optional user info display, and
 * search/settings action buttons with visible focus border indicators.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun TopBar(
    title: String,
    onSearchClick: () -> Unit,
    onSettingsClick: () -> Unit,
    modifier: Modifier = Modifier,
    authState: AuthState? = null
) {
    Row(
        modifier = modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        // App title/logo — full white for maximum contrast on dark background (HELIX-002, 008)
        Text(
            text = title,
            style = MaterialTheme.typography.headlineMedium,
            color = Color.White
        )

        // Action buttons
        Row(
            horizontalArrangement = Arrangement.spacedBy(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // User info (if authenticated)
            authState?.let { state ->
                if (state.isAuthenticated) {
                    Text(
                        text = state.username ?: "User",
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color(0xFFE0E0E0),
                        modifier = Modifier.padding(end = 16.dp)
                    )
                }
            }

            // Search button — with visible focus indicator (HELIX-005, 015)
            var searchFocused by remember { mutableStateOf(false) }
            Surface(
                onClick = onSearchClick,
                modifier = Modifier
                    .onFocusChanged { searchFocused = it.isFocused }
                    .then(
                        if (searchFocused) Modifier.border(
                            BorderStroke(2.dp, FocusBorderColor),
                            shape = RoundedCornerShape(8.dp)
                        ) else Modifier
                    ),
                shape = ClickableSurfaceDefaults.shape(),
                colors = ClickableSurfaceDefaults.colors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Icon(
                    imageVector = Icons.Default.Search,
                    contentDescription = "Search",
                    modifier = Modifier.padding(12.dp),
                    tint = Color.White
                )
            }

            // Settings button — with visible focus indicator (HELIX-005, 015)
            var settingsFocused by remember { mutableStateOf(false) }
            Surface(
                onClick = onSettingsClick,
                modifier = Modifier
                    .onFocusChanged { settingsFocused = it.isFocused }
                    .then(
                        if (settingsFocused) Modifier.border(
                            BorderStroke(2.dp, FocusBorderColor),
                            shape = RoundedCornerShape(8.dp)
                        ) else Modifier
                    ),
                shape = ClickableSurfaceDefaults.shape(),
                colors = ClickableSurfaceDefaults.colors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Icon(
                    imageVector = Icons.Default.Settings,
                    contentDescription = "Settings",
                    modifier = Modifier.padding(12.dp),
                    tint = Color.White
                )
            }
        }
    }
}

/**
 * Network connection status indicator showing a Wi-Fi icon and label
 * with color-coded connected/offline states.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun ConnectionStatus(
    isConnected: Boolean,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier,
        colors = SurfaceDefaults.colors(
            containerColor = if (isConnected)
                MaterialTheme.colorScheme.primaryContainer
            else
                MaterialTheme.colorScheme.errorContainer
        ),
        shape = SurfaceDefaults.shape
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            Icon(
                imageVector = if (isConnected)
                    Icons.Default.Wifi
                else
                    Icons.Default.WifiOff,
                contentDescription = if (isConnected) "Connected" else "Disconnected",
                modifier = Modifier.size(20.dp), // bumped from 16dp for TV readability (HELIX-009)
                tint = if (isConnected)
                    MaterialTheme.colorScheme.onPrimaryContainer
                else
                    MaterialTheme.colorScheme.onError
            )

            Text(
                text = if (isConnected) "Connected" else "Offline",
                style = MaterialTheme.typography.labelMedium, // 14sp — bumped from labelSmall/12sp (HELIX-009)
                color = if (isConnected)
                    MaterialTheme.colorScheme.onPrimaryContainer
                else
                    MaterialTheme.colorScheme.onError
            )
        }
    }
}
