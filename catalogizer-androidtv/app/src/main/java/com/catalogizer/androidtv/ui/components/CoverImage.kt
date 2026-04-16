@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import android.net.Uri
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest

/**
 * Robust cover image composable with automatic fallback from proxy URLs
 * to direct CDN URLs. This fixes environments where the API host cannot
 * reach external CDNs (e.g. DNS override, firewall) while the Android TV
 * device can reach them directly.
 *
 * If the supplied [url] is an `/api/v1/image-proxy?url=...` link, the
 * actual image URL is extracted and tried first. On failure it falls
 * back to the proxy URL so both paths are covered.
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

    val (primaryUrl, fallbackUrl) = remember(url) {
        val proxyParam = url
            ?.takeIf { it.contains("/api/v1/image-proxy") }
            ?.let {
                try {
                    Uri.parse(it).getQueryParameter("url")
                } catch (_: Exception) {
                    null
                }
            }
            ?.let {
                try {
                    java.net.URLDecoder.decode(it, "UTF-8")
                } catch (_: Exception) {
                    it
                }
            }

        if (proxyParam != null && proxyParam.startsWith("http")) {
            proxyParam to url
        } else {
            url to null
        }
    }

    var currentUrl by remember(primaryUrl) { mutableStateOf(primaryUrl) }

    val request = ImageRequest.Builder(context)
        .data(currentUrl)
        .crossfade(true)
        .listener(
            onError = { _, _ ->
                if (fallbackUrl != null && currentUrl != fallbackUrl) {
                    currentUrl = fallbackUrl
                }
            }
        )
        .build()

    val fallbackPlaceholderUrl = remember(context, mediaType) {
        val container = com.catalogizer.androidtv.DependencyContainer.getInstance(context)
        container.getServerUrl().trimEnd('/') + "/api/v1/cover/placeholder/${mediaType ?: "movie"}"
    }

    var showFallback by remember { mutableStateOf(false) }

    if (showFallback) {
        SubcomposeAsyncImage(
            model = ImageRequest.Builder(context)
                .data(fallbackPlaceholderUrl)
                .crossfade(true)
                .build(),
            contentDescription = contentDescription,
            modifier = modifier,
            contentScale = ContentScale.Crop,
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
    } else {
        SubcomposeAsyncImage(
            model = request,
            contentDescription = contentDescription,
            modifier = modifier,
            contentScale = contentScale,
            error = {
                showFallback = true
            }
        )
    }
}
