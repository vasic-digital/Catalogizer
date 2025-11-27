@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.tv.material3.*
import com.catalogizer.androidtv.data.models.MediaItem

@Composable
fun MediaCarousel(
    title: String,
    items: List<MediaItem>,
    onItemClick: (MediaItem) -> Unit,
    modifier: Modifier = Modifier
) {
    Column(modifier = modifier) {
        // Section header
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 48.dp, vertical = 16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.headlineMedium.copy(
                    fontWeight = FontWeight.Bold
                )
            )

            // See all button
            Surface(
                onClick = { /* Navigate to see all */ },
                shape = ClickableSurfaceDefaults.shape(),
                colors = ClickableSurfaceDefaults.colors(
                    containerColor = MaterialTheme.colorScheme.secondary
                )
            ) {
                Text(
                    text = "See All",
                    style = MaterialTheme.typography.labelMedium.copy(
                        color = MaterialTheme.colorScheme.onSecondary
                    ),
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                )
            }
        }

        // Media items carousel
        androidx.tv.foundation.lazy.list.TvLazyRow(
            contentPadding = PaddingValues(horizontal = 48.dp),
            horizontalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            items(items.size) { index ->
                val mediaItem = items[index]
                var isFocused by remember { mutableStateOf(false) }

                MediaCard(
                    mediaItem = mediaItem,
                    onClick = { onItemClick(mediaItem) },
                    onFocus = { isFocused = !isFocused },
                    isFocused = isFocused,
                    modifier = Modifier
                        .width(200.dp)
                        .onFocusChanged { focusState ->
                            isFocused = focusState.isFocused
                        }
                )
            }
        }
    }
}

@Composable
fun FeaturedMediaCard(
    mediaItem: MediaItem,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    var infoFocused by remember { mutableStateOf(false) }
    var playFocused by remember { mutableStateOf(false) }

    Card(
        onClick = onClick,
        modifier = modifier.aspectRatio(16f/9f),
        scale = CardDefaults.scale(scale = 1.0f)
    ) {
        Box(
            modifier = Modifier.fillMaxSize()
        ) {
            // Background gradient
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Color.Black.copy(alpha = 0.3f)
                    )
            )

            // Content
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(48.dp),
                verticalArrangement = Arrangement.Bottom,
                horizontalAlignment = Alignment.Start
            ) {
                // Title
                Text(
                    text = mediaItem.title,
                    style = MaterialTheme.typography.headlineLarge.copy(
                        fontWeight = FontWeight.Bold
                    ),
                    color = Color.White,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis
                )

                Spacer(modifier = Modifier.height(8.dp))

                // Description
                mediaItem.description?.let { description ->
                    Text(
                        text = description,
                        style = MaterialTheme.typography.bodyLarge.copy(
                            color = Color.White.copy(alpha = 0.9f)
                        ),
                        maxLines = 3,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.widthIn(max = 600.dp)
                    )
                }

                Spacer(modifier = Modifier.height(16.dp))

                // Metadata and buttons row
                Row(
                    horizontalArrangement = Arrangement.spacedBy(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Year and rating
                    Column {
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            mediaItem.year?.let { year ->
                                Text(
                                    text = year.toString(),
                                    style = MaterialTheme.typography.titleMedium.copy(
                                        color = Color.White,
                                        fontWeight = FontWeight.Medium
                                    )
                                )
                            }

                            mediaItem.rating?.let { rating ->
                                Text(
                                    text = String.format("%.1f", rating),
                                    style = MaterialTheme.typography.titleMedium.copy(
                                        color = Color.White,
                                        fontWeight = FontWeight.Medium
                                    )
                                )
                            }
                        }

                        mediaItem.quality?.let { quality ->
                            Text(
                                text = quality,
                                style = MaterialTheme.typography.bodyMedium.copy(
                                    color = Color.White.copy(alpha = 0.8f)
                                )
                            )
                        }
                    }

                    // Action buttons
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(12.dp)
                    ) {
                        // Info button
                        @OptIn(ExperimentalTvMaterial3Api::class)
                        Surface(
                            onClick = { /* Show details */ },
                            shape = ClickableSurfaceDefaults.shape(),
                            colors = ClickableSurfaceDefaults.colors(
                                containerColor = Color.Black.copy(alpha = 0.5f)
                            ),
                            scale = ClickableSurfaceDefaults.scale(
                                scale = if (infoFocused) 1.1f else 1.0f
                            ),
                            border = ClickableSurfaceDefaults.border(
                                border = androidx.tv.material3.Border(
                                    BorderStroke(1.dp, if (infoFocused) Color.White else Color.Transparent)
                                )
                            )
                        ) {
                            @OptIn(ExperimentalTvMaterial3Api::class)
                            Icon(
                                imageVector = Icons.Default.Info,
                                contentDescription = "Info",
                                modifier = Modifier.size(32.dp),
                                tint = Color.White
                            )
                        }

                        // Play button
                        @OptIn(ExperimentalTvMaterial3Api::class)
                        Surface(
                            onClick = onClick,
                            shape = ClickableSurfaceDefaults.shape(),
                            colors = ClickableSurfaceDefaults.colors(
                                containerColor = MaterialTheme.colorScheme.primary
                            ),
                            scale = ClickableSurfaceDefaults.scale(
                                scale = if (playFocused) 1.1f else 1.0f
                            ),
                            border = ClickableSurfaceDefaults.border(
                                border = androidx.tv.material3.Border(
                                    BorderStroke(2.dp, if (playFocused) Color.White else Color.Transparent)
                                )
                            )
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.PlayArrow,
                                    contentDescription = "Play",
                                    tint = Color.White
                                )
                                Text(
                                    text = "Play",
                                    style = MaterialTheme.typography.labelLarge.copy(
                                        color = Color.White,
                                        fontWeight = FontWeight.Bold
                                    ),
                                    textAlign = TextAlign.Center
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}