@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.screens.viewer

import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.BrokenImage
import androidx.compose.material.icons.filled.ChevronLeft
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.ZoomIn
import androidx.compose.material.icons.filled.ZoomOutMap
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.input.key.KeyEventType
import androidx.compose.ui.input.key.key
import androidx.compose.ui.input.key.onKeyEvent
import androidx.compose.ui.input.key.type
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.ui.screens.media.LeafActionKind
import com.catalogizer.androidtv.ui.screens.media.leafActionKindFor
import kotlinx.coroutines.flow.first
import kotlinx.serialization.json.jsonPrimitive

/**
 * Three visual states of the image viewer. Modelling them as a sealed type lets
 * the stateless [ImageViewerContent] be rendered deterministically to a PNG on
 * the host (§11.4.170) for every state × theme without any network.
 */
internal sealed interface ImageViewerUiState {
    object Loading : ImageViewerUiState
    data class Error(val message: String) : ImageViewerUiState
    data class Ready(val imageUrl: String) : ImageViewerUiState
}

/**
 * Resolve the absolute image URL from the `stream_url` returned by
 * `GET /api/v1/entities/{id}/stream`, mirroring [MediaPlayerScreen]'s pattern:
 * a server-relative path (starts with `/`) is prefixed with the base URL,
 * an already-absolute URL is used as-is. Pure + internal so it is unit-testable
 * without Compose / a device.
 */
internal fun resolveStreamUrl(streamPath: String?, baseUrl: String): String? {
    if (streamPath.isNullOrBlank()) return null
    return if (streamPath.startsWith("/")) baseUrl.trimEnd('/') + streamPath else streamPath
}

/** Next sibling index with wraparound. Pure for unit testing. */
internal fun nextSiblingIndex(current: Int, size: Int): Int =
    if (size <= 0) 0 else (current + 1).mod(size)

/** Previous sibling index with wraparound. Pure for unit testing. */
internal fun prevSiblingIndex(current: Int, size: Int): Int =
    if (size <= 0) 0 else (current - 1).mod(size)

/**
 * Fit (whole image visible, letterboxed on black) vs Crop (fills the screen,
 * edges cropped) — the two zoom states toggled by D-pad CENTER. Pure mapping.
 */
internal fun imageContentScale(zoomed: Boolean): ContentScale =
    if (zoomed) ContentScale.Crop else ContentScale.Fit

/** "3 / 12" style position label, or null when there is no sibling set. */
internal fun positionLabel(currentIndex: Int, total: Int): String? =
    if (total >= 2 && currentIndex in 0 until total) "${currentIndex + 1} / $total" else null

/**
 * Android-TV image viewer for jpg / png / webp (and other still-image) leaf
 * entities. Resolves the bytes via the SAME `GET /api/v1/entities/{id}/stream`
 * auth pattern the video player uses, then renders them with Coil on a black
 * background.
 *
 * D-pad controls (mirrors [MediaPlayerScreen]'s focus-owning Box + onKeyEvent):
 *  - LEFT / RIGHT : previous / next image within the same parent directory /
 *    collection, when the parent exposes ≥2 image children; otherwise a no-op
 *    (single-image view — see the honest limitation note below).
 *  - CENTER / ENTER : toggle zoom (Fit ↔ Crop).
 *  - BACK : exit.
 *
 * Loading + error states are explicit (§11.4.1 — never a blank screen / crash):
 * a spinner while resolving, a branded broken-image message on failure, and
 * Coil's own loading / error slots while the bytes transfer.
 */
@Composable
fun ImageViewerScreen(
    mediaId: Long,
    onNavigateBack: () -> Unit
) {
    val context = LocalContext.current

    // The currently-shown entity. Starts at [mediaId]; LEFT/RIGHT walk siblings.
    var currentId by remember(mediaId) { mutableStateOf(mediaId) }
    var state by remember { mutableStateOf<ImageViewerUiState>(ImageViewerUiState.Loading) }
    var title by remember { mutableStateOf("") }
    var authToken by remember { mutableStateOf<String?>(null) }
    var zoomed by remember { mutableStateOf(false) }
    // Ordered ids of the image siblings (incl. the entry image) discovered from
    // the parent. Empty ⇒ single-image mode (no prev/next).
    var siblingIds by remember(mediaId) { mutableStateOf<List<Long>>(emptyList()) }

    val focusRequester = remember { androidx.compose.ui.focus.FocusRequester() }

    // Resolve the stream URL + title for the current entity (re-runs on prev/next).
    LaunchedEffect(currentId) {
        state = ImageViewerUiState.Loading
        try {
            val container = DependencyContainer.getInstance(context)
            authToken = container.authRepository.authState.value.token
            val baseUrl = container.getServerUrl().trimEnd('/')

            // Best-effort real title (keeps a sensible default on failure).
            try {
                val item = container.mediaRepository.getMediaById(currentId).first()
                if (item != null && item.title.isNotBlank()) title = item.title
            } catch (_: Exception) { /* keep prior/empty title */ }

            val response = container.api.getEntityStream(currentId)
            if (response.isSuccessful) {
                val streamPath = response.body()?.get("stream_url")?.jsonPrimitive?.content
                val url = resolveStreamUrl(streamPath, baseUrl)
                state = if (url != null) {
                    ImageViewerUiState.Ready(url)
                } else {
                    ImageViewerUiState.Error("No image URL in server response.")
                }
            } else {
                state = ImageViewerUiState.Error(
                    when (response.code()) {
                        404 -> "No file linked to this image. It may not have been scanned yet."
                        401, 403 -> "Authentication required. Please sign in again."
                        500 -> "Server error resolving the image. Storage may be unreachable."
                        else -> "Image unavailable (HTTP ${response.code()})."
                    }
                )
            }
        } catch (e: java.net.ConnectException) {
            state = ImageViewerUiState.Error("Cannot connect to server. Check it is running and reachable.")
        } catch (e: java.net.SocketTimeoutException) {
            state = ImageViewerUiState.Error("Server request timed out. The server may be busy.")
        } catch (e: Exception) {
            state = ImageViewerUiState.Error("Failed to load image: ${e.message}")
        }
    }

    // One-time sibling discovery keyed on the ENTRY image. If its parent exposes
    // ≥2 image children we enable prev/next; otherwise single-image mode. Any
    // failure leaves single-image mode (honest fallback, §11.4.6 — no guessing).
    LaunchedEffect(mediaId) {
        try {
            val container = DependencyContainer.getInstance(context)
            val entry = container.mediaRepository.getMediaById(mediaId).first()
            val parentId = entry?.parentId
            if (parentId != null) {
                val images = container.mediaRepository.getEntityChildren(parentId)
                    .filter { leafActionKindFor(it.mediaType, it.smbPath) == LeafActionKind.IMAGE }
                    .map { it.id }
                if (images.size >= 2 && images.contains(mediaId)) {
                    siblingIds = images
                }
            }
        } catch (_: Exception) { /* single-image mode */ }
    }

    val currentIndex = siblingIds.indexOf(currentId)
    val hasSiblings = siblingIds.size >= 2 && currentIndex >= 0

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .focusRequester(focusRequester)
            .focusable()
            .onKeyEvent { event ->
                if (event.type != KeyEventType.KeyDown) return@onKeyEvent false
                when (event.key) {
                    Key.DirectionCenter, Key.Enter -> {
                        zoomed = !zoomed
                        true
                    }
                    Key.DirectionLeft -> {
                        if (hasSiblings) {
                            currentId = siblingIds[prevSiblingIndex(currentIndex, siblingIds.size)]
                            zoomed = false
                            true
                        } else false
                    }
                    Key.DirectionRight -> {
                        if (hasSiblings) {
                            currentId = siblingIds[nextSiblingIndex(currentIndex, siblingIds.size)]
                            zoomed = false
                            true
                        } else false
                    }
                    Key.Back -> {
                        onNavigateBack()
                        true
                    }
                    else -> false
                }
            }
    ) {
        ImageViewerContent(
            title = title,
            state = state,
            zoomed = zoomed,
            positionLabel = positionLabel(currentIndex, siblingIds.size),
            hasSiblings = hasSiblings,
            authToken = authToken,
            onBack = onNavigateBack,
            onToggleZoom = { zoomed = !zoomed }
        )
    }

    // Own the D-pad focus once composed (mirrors the player's auto-focus).
    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(200)
        try { focusRequester.requestFocus() } catch (_: Exception) {}
    }
}

/**
 * Stateless visuals for the image viewer — rendered directly to a PNG on the
 * host for the §11.4.170 visual proof (every state × {light,dark}) with no
 * network or device. Carries a focusable Back affordance + zoom toggle so the
 * user always has a visible, D-pad-reachable way out and a way to zoom.
 */
@Composable
internal fun ImageViewerContent(
    title: String,
    state: ImageViewerUiState,
    zoomed: Boolean,
    positionLabel: String?,
    hasSiblings: Boolean,
    authToken: String?,
    onBack: () -> Unit,
    onToggleZoom: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black)
    ) {
        when (state) {
            is ImageViewerUiState.Loading -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    CircularProgressIndicator(color = Color.White)
                    Spacer(Modifier.height(12.dp))
                    Text(
                        "Loading image…",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            }
            is ImageViewerUiState.Error -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.BrokenImage,
                        contentDescription = "Error",
                        tint = Color.White,
                        modifier = Modifier.size(48.dp)
                    )
                    Spacer(Modifier.height(16.dp))
                    Text(
                        "Unable to Display Image",
                        style = MaterialTheme.typography.headlineSmall,
                        color = Color.White
                    )
                    Spacer(Modifier.height(8.dp))
                    Text(
                        text = state.message,
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            }
            is ImageViewerUiState.Ready -> {
                val ctx = LocalContext.current
                SubcomposeAsyncImage(
                    model = ImageRequest.Builder(ctx)
                        .data(state.imageUrl)
                        .apply {
                            // /stream requires the same Bearer auth the player uses.
                            if (!authToken.isNullOrBlank()) {
                                addHeader("Authorization", "Bearer $authToken")
                            }
                        }
                        .crossfade(true)
                        .build(),
                    contentDescription = title.ifBlank { "Image" },
                    contentScale = imageContentScale(zoomed),
                    modifier = Modifier.fillMaxSize(),
                    loading = {
                        Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                            CircularProgressIndicator(color = Color.White)
                        }
                    },
                    error = {
                        // §11.4.1 — a Coil decode/transfer failure shows an explicit
                        // message, never a blank black screen.
                        Column(
                            modifier = Modifier.fillMaxSize(),
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center
                        ) {
                            Icon(
                                imageVector = Icons.Default.BrokenImage,
                                contentDescription = "Image failed to load",
                                tint = Color.White,
                                modifier = Modifier.size(48.dp)
                            )
                            Spacer(Modifier.height(12.dp))
                            Text(
                                "Image failed to load.",
                                style = MaterialTheme.typography.bodyLarge,
                                color = Color.White.copy(alpha = 0.7f)
                            )
                        }
                    }
                )
            }
        }

        // Top overlay: Back affordance + title + (optional) position. Always
        // present so the user can read the title and exit regardless of state.
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Brush.verticalGradient(listOf(Color.Black.copy(alpha = 0.85f), Color.Transparent)))
                .padding(horizontal = 24.dp, vertical = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            ViewerIconButton(
                icon = Icons.Default.ArrowBack,
                contentDescription = "Back",
                onClick = onBack
            )
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title.ifBlank { "Image" },
                    style = MaterialTheme.typography.headlineSmall,
                    color = Color.White,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                positionLabel?.let {
                    Text(
                        text = it,
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            }
            // Zoom toggle affordance (visible state cue + clickable for pointer
            // devices; D-pad CENTER toggles it too).
            ViewerIconButton(
                icon = if (zoomed) Icons.Default.ZoomOutMap else Icons.Default.ZoomIn,
                contentDescription = if (zoomed) "Fit to screen" else "Zoom in",
                onClick = onToggleZoom
            )
        }

        // Sibling-navigation hint (only when a prev/next set exists) — bottom bar.
        if (hasSiblings) {
            Row(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .fillMaxWidth()
                    .background(Brush.verticalGradient(listOf(Color.Transparent, Color.Black.copy(alpha = 0.85f))))
                    .padding(horizontal = 24.dp, vertical = 16.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.Center
            ) {
                Icon(Icons.Default.ChevronLeft, contentDescription = null, tint = Color.White, modifier = Modifier.size(22.dp))
                Text(
                    "Use ◀ ▶ to browse",
                    style = MaterialTheme.typography.bodyMedium,
                    color = Color.White.copy(alpha = 0.85f),
                    modifier = Modifier.padding(horizontal = 8.dp)
                )
                Icon(Icons.Default.ChevronRight, contentDescription = null, tint = Color.White, modifier = Modifier.size(22.dp))
            }
        }
    }
}

/**
 * Small circular, focusable icon button for the overlay (mirrors the player's
 * top-bar back button). TV [Surface] is focusable + clickable for D-pad.
 */
@Composable
private fun ViewerIconButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    contentDescription: String,
    onClick: () -> Unit
) {
    Surface(
        onClick = onClick,
        modifier = Modifier.size(40.dp),
        shape = ClickableSurfaceDefaults.shape(shape = CircleShape),
        colors = ClickableSurfaceDefaults.colors(
            containerColor = Color.White.copy(alpha = 0.15f)
        )
    ) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Icon(
                imageVector = icon,
                contentDescription = contentDescription,
                tint = Color.White,
                modifier = Modifier.size(22.dp)
            )
        }
    }
}
