@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Image
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.tv.material3.ExperimentalTvMaterial3Api
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest

/**
 * Cover image composable that always uses the backend-provided URL.
 *
 * Per the image policy, all images MUST be served via the backend proxy;
 * client apps never communicate directly with external CDNs. If the backend
 * URL is missing, a backend placeholder SVG is requested instead.
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

    val modelUrl = remember(url, context, mediaType) {
        url ?: run {
            val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
            container.getServerUrl().trimEnd('/') + "/api/v1/cover/placeholder/${mediaType ?: "movie"}"
        }
    }

    SubcomposeAsyncImage(
        model = ImageRequest.Builder(context)
            .data(modelUrl)
            .crossfade(true)
            .build(),
        contentDescription = contentDescription,
        modifier = modifier,
        contentScale = contentScale,
        loading = {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.Image,
                    contentDescription = "Loading",
                    tint = Color.White.copy(alpha = 0.3f)
                )
            }
        },
        error = {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.Image,
                    contentDescription = "Image unavailable",
                    tint = Color.White.copy(alpha = 0.5f)
                )
            }
        }
    )
}
