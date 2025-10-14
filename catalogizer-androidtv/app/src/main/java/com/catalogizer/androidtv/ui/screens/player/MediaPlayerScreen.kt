package com.catalogizer.androidtv.ui.screens.player

import androidx.compose.foundation.layout.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun MediaPlayerScreen(
    mediaId: Long,
    onNavigateBack: () -> Unit
) {
    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp),
            modifier = Modifier.padding(48.dp)
        ) {
            Text(
                text = "Media Player",
                style = MaterialTheme.typography.headlineLarge
            )

            Text(
                text = "Playing Media ID: $mediaId",
                style = MaterialTheme.typography.bodyLarge
            )

            Text(
                text = "Player implementation coming soon",
                style = MaterialTheme.typography.bodyMedium
            )

            Button(onClick = onNavigateBack) {
                Text("Back")
            }
        }
    }
}