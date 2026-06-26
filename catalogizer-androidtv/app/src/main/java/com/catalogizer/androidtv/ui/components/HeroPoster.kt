@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.Text
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest
import com.catalogizer.androidtv.R

/**
 * Bounded, branded hero poster for the media-detail screen (§11.4.162 / §11.4.170).
 *
 * Root-cause this composable fixes (FACT, on-device screenshot): the detail
 * screen previously placed a `CoverImage(Modifier.fillMaxWidth(), FillWidth)`
 * with NO height/aspect bound inside the 280/400 dp hero Box. When the cover
 * URL was missing or failed, the backend gray placeholder SVG floated as a
 * full-width-but-short band, rendering a GIANT featureless gray rounded bar
 * across the top of the screen.
 *
 * [HeroPoster] is BOUNDED + BRANDED by construction:
 *  - The caller supplies the hero height via [modifier] (the 280/400 dp Box).
 *    The image fills that box with `Modifier.fillMaxSize()` + `ContentScale.Crop`
 *    (NOT FillWidth), so a real poster always fills the hero without floating
 *    and is never distorted (Crop, not stretch).
 *  - When [coverUrl] is null/blank OR the image fails to load, it renders the
 *    branded OpenDesign fallback — the dark-navy vertical gradient
 *    (`#1A237E -> #0A0A0A`) + the app logo + the [title] centered — NEVER a bare
 *    gray pill / the backend placeholder SVG. This is the §11.4.162 "no bare
 *    gray, no label overlap, light+dark" rule applied to the missing-poster case.
 *
 * Stateless + host-renderable (§11.4.170): the body holds no
 * `DependencyContainer`/network singleton — the caller resolves [coverUrl] (the
 * server-prefix / image-proxy rewrite stays in the screen). [mediaType] is passed
 * for semantics only (accessibility description) and carries no hidden I/O. This
 * lets a Roborazzi/Paparazzi host-render test capture the composable to a PNG
 * with no device/emulator.
 */
@Composable
fun HeroPoster(
    coverUrl: String?,
    title: String,
    mediaType: String? = null,
    modifier: Modifier = Modifier
) {
    Box(modifier = modifier) {
        if (coverUrl.isNullOrBlank()) {
            HeroBrandedFallback(
                title = title,
                mediaType = mediaType,
                modifier = Modifier.fillMaxSize()
            )
        } else {
            val context = LocalContext.current
            SubcomposeAsyncImage(
                model = ImageRequest.Builder(context)
                    .data(coverUrl)
                    .crossfade(true)
                    .build(),
                contentDescription = title,
                modifier = Modifier.fillMaxSize(),
                // Crop fills the bounded hero box edge-to-edge without floating
                // and without distorting the poster (no stretch). This is the
                // load-bearing change from the old unbounded FillWidth.
                contentScale = ContentScale.Crop,
                // Both the still-loading and the failed-to-load states render the
                // branded fallback rather than the backend gray placeholder, so a
                // missing/slow cover is never a giant gray bar.
                loading = {
                    HeroBrandedFallback(
                        title = title,
                        mediaType = mediaType,
                        modifier = Modifier.fillMaxSize()
                    )
                },
                error = {
                    HeroBrandedFallback(
                        title = title,
                        mediaType = mediaType,
                        modifier = Modifier.fillMaxSize()
                    )
                }
            )
        }
    }
}

/**
 * Branded missing-poster fill: OpenDesign dark-navy gradient + app logo + title.
 * Stateless and resource-only (no network), so it renders identically on the
 * host (Roborazzi) and on-device — the §11.4.170 visual-proof anchor.
 */
@Composable
private fun HeroBrandedFallback(
    title: String,
    mediaType: String?,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier.background(
            Brush.verticalGradient(listOf(Color(0xFF1A237E), Color(0xFF0A0A0A)))
        ),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(horizontal = 24.dp)
        ) {
            Image(
                painter = painterResource(R.drawable.app_icon),
                contentDescription = mediaType?.let { "$title ($it)" } ?: title,
                modifier = Modifier.size(96.dp),
                contentScale = ContentScale.Fit
            )
            Spacer(Modifier.height(16.dp))
            Text(
                text = title,
                style = MaterialTheme.typography.titleLarge,
                color = Color.White,
                fontWeight = FontWeight.Bold,
                textAlign = TextAlign.Center,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis
            )
        }
    }
}
