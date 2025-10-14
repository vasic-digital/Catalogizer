package com.catalogizer.androidtv.ui.screens.media

import androidx.compose.foundation.layout.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun MediaDetailScreen(
    mediaId: Long,
    onNavigateBack: () -> Unit,
    onNavigateToPlayer: (Long) -> Unit
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
                text = "Media Details",
                style = MaterialTheme.typography.headlineLarge
            )

            Text(
                text = "Media ID: $mediaId",
                style = MaterialTheme.typography.bodyLarge
            )

            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Button(onClick = { onNavigateToPlayer(mediaId) }) {
                    Text("Play")
                }

                Button(onClick = onNavigateBack) {
                    Text("Back")
                }
            }
        }
    }
}