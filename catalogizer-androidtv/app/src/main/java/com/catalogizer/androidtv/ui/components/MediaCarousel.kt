package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.BorderStroke
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.tv.foundation.PivotOffsets
import androidx.tv.foundation.lazy.list.TvLazyRow
import androidx.tv.foundation.lazy.list.items
import androidx.tv.material3.*
import coil.compose.AsyncImage
import coil.request.ImageRequest
import com.catalogizer.androidtv.data.models.MediaItem

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun MediaCarousel(
    mediaItems: List<MediaItem>,
    onItemClick: (MediaItem) -> Unit,
    onItemFocus: (MediaItem) -> Unit,
    modifier: Modifier = Modifier
) {
    if (mediaItems.isEmpty()) return

    var focusedItem by remember { mutableStateOf<MediaItem?>(null) }

    Column(modifier = modifier) {
        // Main featured item display
        focusedItem?.let { item ->
            FeaturedMediaDisplay(
                mediaItem = item,
                onPlayClick = { onItemClick(item) },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(400.dp)
                    .padding(bottom = 24.dp)
            )
        }

        // Carousel row
        TvLazyRow(
            contentPadding = PaddingValues(horizontal = 48.dp),
            horizontalArrangement = Arrangement.spacedBy(16.dp),
            pivotOffsets = PivotOffsets(parentFraction = 0.07f)
        ) {
            items(mediaItems) { item ->
                CarouselCard(
                    mediaItem = item,
                    isSelected = item == focusedItem,
                    onClick = { onItemClick(item) },
                    onFocus = {
                        focusedItem = item
                        onItemFocus(item)
                    },
                    modifier = Modifier.width(160.dp)
                )
            }
        }
    }

    // Set initial focused item
    LaunchedEffect(mediaItems) {
        if (focusedItem == null && mediaItems.isNotEmpty()) {
            focusedItem = mediaItems.first()
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun FeaturedMediaDisplay(
    mediaItem: MediaItem,
    onPlayClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current

    Box(modifier = modifier) {
        // Background image
        AsyncImage(
            model = ImageRequest.Builder(context)
                .data(mediaItem.backdropUrl ?: mediaItem.posterUrl)
                .crossfade(true)
                .build(),
            contentDescription = null,
            modifier = Modifier
                .fillMaxSize()
                .clip(RoundedCornerShape(16.dp)),
            contentScale = ContentScale.Crop
        )

        // Gradient overlay
        Box(
            modifier = Modifier
                .fillMaxSize()
                .clip(RoundedCornerShape(16.dp))
                .background(
                    Brush.horizontalGradient(
                        colors = listOf(
                            Color.Black.copy(alpha = 0.8f),
                            Color.Black.copy(alpha = 0.6f),
                            Color.Transparent
                        ),
                        startX = 0f,
                        endX = 1000f
                    )
                )
        )

        // Content overlay
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(48.dp),
            verticalArrangement = Arrangement.Bottom
        ) {
            Text(
                text = mediaItem.title,
                style = MaterialTheme.typography.displaySmall,
                color = Color.White,
                modifier = Modifier.padding(bottom = 8.dp)
            )

            if (mediaItem.year != null) {
                Text(
                    text = mediaItem.year.toString(),
                    style = MaterialTheme.typography.titleMedium,
                    color = Color.White.copy(alpha = 0.8f),
                    modifier = Modifier.padding(bottom = 8.dp)
                )
            }

            if (!mediaItem.description.isNullOrBlank()) {
                Text(
                    text = mediaItem.description,
                    style = MaterialTheme.typography.bodyLarge,
                    color = Color.White.copy(alpha = 0.9f),
                    maxLines = 3,
                    modifier = Modifier
                        .widthIn(max = 600.dp)
                        .padding(bottom = 24.dp)
                )
            }

            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                var playFocused by remember { mutableStateOf(false) }
                var infoFocused by remember { mutableStateOf(false) }

                Button(
                    onClick = onPlayClick,
                    modifier = Modifier.onFocusChanged { playFocused = it.isFocused },
                    colors = ButtonDefaults.colors(
                        containerColor = if (playFocused)
                            MaterialTheme.colorScheme.primary
                        else
                            Color.White,
                        contentColor = if (playFocused)
                            MaterialTheme.colorScheme.onPrimary
                        else
                            Color.Black
                    ),
                    scale = ButtonDefaults.scale(
                        scale = if (playFocused) 1.1f else 1.0f
                    )
                ) {
                    Text("Play")
                }

                OutlinedButton(
                    onClick = { /* Navigate to details */ },
                    modifier = Modifier.onFocusChanged { infoFocused = it.isFocused },
                    colors = ButtonDefaults.colors(
                        containerColor = if (infoFocused)
                            Color.White.copy(alpha = 0.2f)
                        else
                            Color.Transparent,
                        contentColor = Color.White
                    ),
                    border = ButtonDefaults.border(
                        border = BorderStroke(
                            width = 1.dp,
                            color = Color.White
                        )
                    ),
                    scale = ButtonDefaults.scale(
                        scale = if (infoFocused) 1.1f else 1.0f
                    )
                ) {
                    Text("More Info")
                }
            }
        }
    }
}

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
private fun CarouselCard(
    mediaItem: MediaItem,
    isSelected: Boolean,
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
            containerColor = Color.Transparent
        ),
        scale = CardDefaults.scale(
            scale = when {
                isSelected -> 1.2f
                isFocused -> 1.1f
                else -> 1.0f
            }
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
            AsyncImage(
                model = ImageRequest.Builder(context)
                    .data(mediaItem.posterUrl ?: mediaItem.thumbnailUrl)
                    .crossfade(true)
                    .build(),
                contentDescription = mediaItem.title,
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(2f / 3f)
                    .clip(RoundedCornerShape(8.dp)),
                contentScale = ContentScale.Crop
            )

            if (isSelected) {
                Text(
                    text = mediaItem.title,
                    style = MaterialTheme.typography.titleSmall,
                    color = MaterialTheme.colorScheme.onSurface,
                    textAlign = TextAlign.Center,
                    maxLines = 2,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp)
                )
            }
        }
    }
}