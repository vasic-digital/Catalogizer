@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.BorderStroke
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.tv.material3.*
import com.catalogizer.androidtv.data.models.MediaItem

@Composable
fun MediaCard(
    mediaItem: MediaItem,
    onClick: () -> Unit,
    onFocus: () -> Unit,
    modifier: Modifier = Modifier,
    isFocused: Boolean = false
) {
    Card(
        onClick = onClick,
        modifier = modifier.aspectRatio(2f/3f),
        scale = CardDefaults.scale(
            scale = if (isFocused) 1.05f else 1.0f
        ),
        border = CardDefaults.border(
            focusedBorder = Border(
                border = androidx.compose.foundation.BorderStroke(
                    width = 2.dp,
                    color = MaterialTheme.colorScheme.primary
                ),
                shape = RoundedCornerShape(8.dp)
            )
        )
    ) {
        Column {
            // Thumbnail placeholder
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(16f/9f)
                    .background(MaterialTheme.colorScheme.surfaceVariant)
                    .clip(RoundedCornerShape(topStart = 8.dp, topEnd = 8.dp)),
                contentAlignment = Alignment.Center
            ) {
                if (mediaItem.thumbnailUrl != null) {
                    // TODO: Load actual thumbnail using Coil
                    // AsyncImage(
                    //     model = mediaItem.thumbnailUrl,
                    //     contentDescription = mediaItem.title,
                    //     modifier = Modifier.fillMaxSize(),
                    //     contentScale = ContentScale.Crop
                    // )
                }
                
                // Play button overlay
                Box(
                    modifier = Modifier
                        .background(
                            color = Color.Black.copy(alpha = 0.6f),
                            shape = RoundedCornerShape(50)
                        )
                ) {
                    Icon(
                        imageVector = Icons.Default.PlayArrow,
                        contentDescription = "Play",
                        modifier = Modifier.size(32.dp),
                        tint = Color.White
                    )
                }
            }

            // Content info
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(12.dp)
            ) {
                // Title
                Text(
                    text = mediaItem.title,
                    style = MaterialTheme.typography.titleSmall.copy(
                        fontWeight = FontWeight.Bold
                    ),
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.fillMaxWidth()
                )

                Spacer(modifier = Modifier.height(4.dp))

                // Metadata row
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Year and duration
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        mediaItem.year?.let { year ->
                            Text(
                                text = year.toString(),
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                            )
                        }

                        mediaItem.duration?.let { duration ->
                            val minutes = duration / 60
                            val hours = minutes / 60
                            val remainingMinutes = minutes % 60
                            
                            val durationText = when {
                                hours > 0 -> "${hours}h ${remainingMinutes}m"
                                else -> "${minutes}m"
                            }

                            if (mediaItem.year != null) {
                                Text(
                                    text = " • $durationText",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                                )
                            } else {
                                Text(
                                    text = durationText,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                                )
                            }
                        }
                    }

                        // Watch progress indicator
                        if (mediaItem.hasWatchProgress) {
                            Box(
                                modifier = Modifier.background(
                                    color = MaterialTheme.colorScheme.primary,
                                    shape = RoundedCornerShape(2.dp)
                                )
                            ) {
                                Text(
                                    text = "${(mediaItem.watchProgress * 100).toInt()}%",
                                    style = MaterialTheme.typography.labelSmall.copy(
                                        fontSize = 10.sp
                                    ),
                                    modifier = Modifier.padding(horizontal = 4.dp, vertical = 2.dp)
                                )
                            }
                        }
                }

                    // Quality indicator
                    mediaItem.quality?.let { quality ->
                        Spacer(modifier = Modifier.height(4.dp))
                        
                        Box(
                            modifier = Modifier.background(
                                color = MaterialTheme.colorScheme.secondary,
                                shape = RoundedCornerShape(4.dp)
                            )
                        ) {
                            Text(
                                text = quality,
                                style = MaterialTheme.typography.labelSmall.copy(
                                    fontSize = 10.sp
                                ),
                                modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                                color = MaterialTheme.colorScheme.onSecondary
                            )
                        }
                    }

                // Rating
                mediaItem.rating?.let { rating ->
                    Spacer(modifier = Modifier.height(4.dp))
                    
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        // Star icon would go here
                        Text(
                            text = String.format("%.1f", rating),
                            style = MaterialTheme.typography.bodySmall.copy(
                                fontWeight = FontWeight.Medium
                            ),
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f)
                        )
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun CompactMediaCard(
    mediaItem: MediaItem,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    @OptIn(ExperimentalTvMaterial3Api::class)
    Surface(
        onClick = onClick,
        modifier = modifier.fillMaxWidth(),
        shape = ClickableSurfaceDefaults.shape(),
        colors = ClickableSurfaceDefaults.colors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        border = ClickableSurfaceDefaults.border(
            border = androidx.tv.material3.Border(
                androidx.compose.foundation.BorderStroke(1.dp, androidx.tv.material3.MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f))
            )
        )
    ) {
        Row(
            modifier = Modifier.padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
        // Thumbnail
        @OptIn(ExperimentalTvMaterial3Api::class)
        Surface(
            modifier = Modifier,
            shape = NonInteractiveSurfaceDefaults.shape,
            colors = NonInteractiveSurfaceDefaults.colors(
                containerColor = androidx.tv.material3.MaterialTheme.colorScheme.surfaceVariant
            ),
            border = androidx.tv.material3.Border(
                androidx.compose.foundation.BorderStroke(1.dp, androidx.tv.material3.MaterialTheme.colorScheme.onSurface.copy(alpha = 0.2f))
            )
        ) {
            Box(
                modifier = Modifier.size(80.dp, 60.dp),
                contentAlignment = Alignment.Center
            ) {
                // TODO: Load actual thumbnail
                Box(
                    modifier = Modifier.background(
                        color = Color.Black.copy(alpha = 0.6f),
                        shape = RoundedCornerShape(50)
                    )
                ) {
                    Icon(
                        imageVector = Icons.Default.PlayArrow,
                        contentDescription = "Play",
                        modifier = Modifier.size(24.dp),
                        tint = Color.White
                    )
                }
            }
        }

        Spacer(modifier = Modifier.width(12.dp))

        // Info
        Column(
            modifier = Modifier.weight(1f)
        ) {
            Text(
                text = mediaItem.title,
                style = MaterialTheme.typography.titleSmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )

            Spacer(modifier = Modifier.height(4.dp))

            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                mediaItem.year?.let { year ->
                    Text(
                        text = year.toString(),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                    )
                }

                mediaItem.quality?.let { quality ->
                    if (mediaItem.year != null) {
                        Text(
                            text = " • $quality",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    } else {
                        Text(
                            text = quality,
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }
                }
            }

            mediaItem.rating?.let { rating ->
                Text(
                    text = String.format("%.1f", rating),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.primary
                )
            }
        }  // Close Column
        }  // Close Row
    }  // Close Surface
}