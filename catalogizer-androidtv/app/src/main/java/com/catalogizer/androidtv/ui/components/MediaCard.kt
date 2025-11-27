package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import coil.compose.AsyncImage
import coil.request.ImageRequest
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.utils.formatDuration
import com.catalogizer.androidtv.utils.formatFileSize

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun MediaCard(
    mediaItem: MediaItem,
    onClick: () -> Unit,
    onFocus: () -> Unit,
    modifier: Modifier = Modifier
) {
    var isFocused by remember { mutableStateOf(false) }
    val context = LocalContext.current

    Card(
        onClick = onClick,
        modifier = modifier
            .onFocusChanged {
                isFocused = it.isFocused
                if (it.isFocused) onFocus()
            },
        colors = CardDefaults.colors(
            containerColor = if (isFocused)
                MaterialTheme.colorScheme.primaryContainer
            else
                MaterialTheme.colorScheme.surface
        ),
        shape = CardDefaults.shape(RoundedCornerShape(8.dp)),
        scale = CardDefaults.scale(
            scale = if (isFocused) 1.1f else 1.0f
        ),
        border = CardDefaults.border(
            focusedBorder = Border(
                border = BorderStroke(
                    width = 3.dp,
                    color = MaterialTheme.colorScheme.primary
                ),
                shape = RoundedCornerShape(8.dp)
            )
        )
    ) {
        Column {
            // Media poster/thumbnail
            AsyncImage(
                model = ImageRequest.Builder(context)
                    .data(mediaItem.posterUrl ?: mediaItem.thumbnailUrl)
                    .crossfade(true)
                    .build(),
                contentDescription = mediaItem.title,
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(2f / 3f)
                    .clip(RoundedCornerShape(topStart = 8.dp, topEnd = 8.dp)),
                contentScale = ContentScale.Crop
            )

            // Media info
            Column(
                modifier = Modifier.padding(12.dp)
            ) {
                Text(
                    text = mediaItem.title,
                    style = MaterialTheme.typography.titleSmall,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    color = if (isFocused)
                        MaterialTheme.colorScheme.onPrimaryContainer
                    else
                        MaterialTheme.colorScheme.onSurface
                )

                Spacer(modifier = Modifier.height(4.dp))

                if (mediaItem.year != null) {
                    Text(
                        text = mediaItem.year.toString(),
                        style = MaterialTheme.typography.bodySmall,
                        color = if (isFocused)
                            MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                        else
                            MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                    )
                }

                if (mediaItem.duration != null && mediaItem.duration > 0) {
                    Text(
                        text = formatDuration(mediaItem.duration),
                        style = MaterialTheme.typography.bodySmall,
                        color = if (isFocused)
                            MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                        else
                            MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                    )
                }

                if (mediaItem.fileSize != null) {
                    Text(
                        text = formatFileSize(mediaItem.fileSize),
                        style = MaterialTheme.typography.bodySmall,
                        color = if (isFocused)
                            MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                        else
                            MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                    )
                }

                // Media type indicator
                Surface(
                    modifier = Modifier
                        .padding(top = 8.dp)
                        .clip(RoundedCornerShape(4.dp)),
                    color = when (mediaItem.mediaType.lowercase()) {
                        "movie" -> Color(0xFF4CAF50)
                        "tv" -> Color(0xFF2196F3)
                        "music" -> Color(0xFFFF9800)
                        "document" -> Color(0xFF9C27B0)
                        else -> MaterialTheme.colorScheme.secondary
                    }
                ) {
                    Text(
                        text = mediaItem.mediaType.uppercase(),
                        style = MaterialTheme.typography.labelSmall,
                        color = Color.White,
                        modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                    )
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
    onFocus: () -> Unit,
    modifier: Modifier = Modifier
) {
    var isFocused by remember { mutableStateOf(false) }
    val context = LocalContext.current

    Card(
        onClick = onClick,
        modifier = modifier
            .height(120.dp)
            .onFocusChanged {
                isFocused = it.isFocused
                if (it.isFocused) onFocus()
            },
        colors = CardDefaults.colors(
            containerColor = if (isFocused)
                MaterialTheme.colorScheme.primaryContainer
            else
                MaterialTheme.colorScheme.surface
        ),
        shape = CardDefaults.shape(RoundedCornerShape(8.dp)),
        scale = CardDefaults.scale(
            scale = if (isFocused) 1.05f else 1.0f
        ),
        border = CardDefaults.border(
            focusedBorder = Border(
                border = BorderStroke(
                    width = 2.dp,
                    color = MaterialTheme.colorScheme.primary
                ),
                shape = RoundedCornerShape(8.dp)
            )
        )
    ) {
        Row(
            modifier = Modifier.fillMaxSize()
        ) {
            // Thumbnail
            AsyncImage(
                model = ImageRequest.Builder(context)
                    .data(mediaItem.thumbnailUrl ?: mediaItem.posterUrl)
                    .crossfade(true)
                    .build(),
                contentDescription = mediaItem.title,
                modifier = Modifier
                    .width(80.dp)
                    .fillMaxHeight()
                    .clip(RoundedCornerShape(topStart = 8.dp, bottomStart = 8.dp)),
                contentScale = ContentScale.Crop
            )

            // Content
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(12.dp),
                verticalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text(
                        text = mediaItem.title,
                        style = MaterialTheme.typography.titleSmall,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                        color = if (isFocused)
                            MaterialTheme.colorScheme.onPrimaryContainer
                        else
                            MaterialTheme.colorScheme.onSurface
                    )

                    if (mediaItem.year != null) {
                        Text(
                            text = mediaItem.year.toString(),
                            style = MaterialTheme.typography.bodySmall,
                            color = if (isFocused)
                                MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                            else
                                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    if (mediaItem.duration != null && mediaItem.duration > 0) {
                        Text(
                            text = formatDuration(mediaItem.duration),
                            style = MaterialTheme.typography.bodySmall,
                            color = if (isFocused)
                                MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                            else
                                MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f)
                        )
                    }

                    Surface(
                        containerColor = when (mediaItem.mediaType.lowercase()) {
                            "movie" -> Color(0xFF4CAF50)
                            "tv" -> Color(0xFF2196F3)
                            "music" -> Color(0xFFFF9800)
                            "document" -> Color(0xFF9C27B0)
                            else -> MaterialTheme.colorScheme.secondary
                        },
                        shape = RoundedCornerShape(4.dp)
                    ) {
                        Text(
                            text = mediaItem.mediaType.uppercase(),
                            style = MaterialTheme.typography.labelSmall,
                            color = Color.White,
                            modifier = Modifier.padding(horizontal = 4.dp, vertical = 2.dp)
                        )
                    }
                }
            }
        }
    }
}