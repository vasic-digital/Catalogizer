@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Image
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.tv.material3.ExperimentalTvMaterial3Api
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest

/**
 * Branded-neutral placeholder scrim (§11.4.162) used by the loading/error
 * fallbacks. Translucent dark verticalgradient: lets the parent's brand
 * gradient tint through while keeping the centered glyph legible, instead of a
 * bare gray placeholder bar.
 */
private val NeutralPlaceholderScrim: Brush = Brush.verticalGradient(
    listOf(Color.Black.copy(alpha = 0.25f), Color.Black.copy(alpha = 0.55f))
)

/**
 * Cover image composable that always uses the backend-provided URL.
 *
 * Per the image policy, all images MUST be served via the backend proxy;
 * client apps never communicate directly with external CDNs. If the backend
 * URL is missing or fails to load, a backend placeholder SVG is requested.
 */
@Composable
fun CoverImage(
    url: String?,
    contentDescription: String?,
    modifier: Modifier = Modifier,
    contentScale: ContentScale = ContentScale.Crop,
    mediaType: String? = "movie"
) {
    val context = LocalContext.current

    val fallbackUrl = remember(context, mediaType) {
        val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
        container.getServerUrl().trimEnd('/') + "/api/v1/cover/placeholder/${mediaType ?: "movie"}"
    }

    var currentUrl by remember(url) { mutableStateOf(if (url.isNullOrBlank()) fallbackUrl else url) }
    var useFallback by remember(url) { mutableStateOf(false) }

    val modelUrl = if (useFallback) fallbackUrl else currentUrl

    SubcomposeAsyncImage(
        model = ImageRequest.Builder(context)
            .data(modelUrl)
            .crossfade(true)
            .build(),
        contentDescription = contentDescription,
        modifier = modifier,
        contentScale = contentScale,
        loading = {
            // Branded-neutral placeholder fill (§11.4.162) — a translucent dark
            // scrim that lets the parent's brand gradient tint through, with the
            // image glyph centered. NOT a bare gray bar / opaque gray pill.
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(NeutralPlaceholderScrim),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.Image,
                    contentDescription = "Loading cover image",
                    tint = Color.White.copy(alpha = 0.85f),
                    modifier = Modifier.padding(12.dp)
                )
            }
        },
        error = {
            if (!useFallback && modelUrl != fallbackUrl) {
                // Trigger recomposition with fallback URL on next frame
                useFallback = true
            }
            // Same branded-neutral scrim as the loading state so a missing /
            // failed cover reads as an intentional placeholder (§11.4.162), never
            // a bare gray placeholder bar.
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(NeutralPlaceholderScrim),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.Image,
                    contentDescription = "Cover image unavailable",
                    tint = Color.White.copy(alpha = 0.9f),
                    modifier = Modifier.padding(12.dp)
                )
            }
        }
    )
}
