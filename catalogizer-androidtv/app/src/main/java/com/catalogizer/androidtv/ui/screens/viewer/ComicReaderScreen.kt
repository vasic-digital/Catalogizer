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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest
import com.catalogizer.androidtv.DependencyContainer
import kotlinx.coroutines.flow.first
import kotlinx.serialization.json.jsonPrimitive

/**
 * Three visual states of the comic reader. Modelling them as a sealed type lets
 * the stateless [ComicReaderContent] be rendered deterministically to a PNG on
 * the host (§11.4.170) for every state × theme without any network.
 *
 * [Ready] carries only the page COUNT; the current page index is held as
 * separate screen state so paging re-renders the page image (a new Coil request)
 * WITHOUT refetching the `/pages` listing.
 */
internal sealed interface ComicReaderUiState {
    object Loading : ComicReaderUiState
    data class Error(val message: String) : ComicReaderUiState
    data class Ready(val totalPages: Int) : ComicReaderUiState
}

/**
 * Absolute URL of a single comic page image:
 * `<baseUrl>/api/v1/entities/<mediaId>/pages/<pageIndex>` (0-based index, matching
 * catalog-api `comic_pages_handler.go`). Pure + internal so it is unit-testable
 * without Compose / a device (§11.4.6 — URL construction proven, not guessed).
 */
internal fun comicPageUrl(baseUrl: String, mediaId: Long, pageIndex: Int): String =
    "${baseUrl.trimEnd('/')}/api/v1/entities/$mediaId/pages/$pageIndex"

/**
 * Next page index, CLAMPED to the closed range [0, total-1]. Comics read linearly
 * — unlike the image viewer's sibling wraparound, the last page does NOT wrap to
 * the first (you cannot turn past the back cover). Pure for unit testing.
 */
internal fun nextComicPage(current: Int, total: Int): Int =
    if (total <= 0) 0 else (current + 1).coerceIn(0, total - 1)

/** Previous page index, CLAMPED to [0, total-1] (first page does not wrap). */
internal fun prevComicPage(current: Int, total: Int): Int =
    if (total <= 0) 0 else (current - 1).coerceIn(0, total - 1)

/** First page index (always 0 when there is at least one page). */
internal fun firstComicPage(): Int = 0

/** Last page index, [0, total-1]; 0 for an empty/unknown set. */
internal fun lastComicPage(total: Int): Int = if (total <= 0) 0 else total - 1

/** "Page 3 / 12" style indicator, or null when there are no readable pages. */
internal fun comicPageLabel(currentIndex: Int, total: Int): String? =
    if (total >= 1 && currentIndex in 0 until total) "Page ${currentIndex + 1} / $total" else null

/**
 * Honest, user-facing error message for a non-2xx `/pages` response (§11.4.1 — an
 * explicit message, never a blank screen / crash). The `.cbr` case is the
 * forensically-important one: catalog-api returns HTTP 501 for RAR-backed comics
 * because there is no RAR decoder wired yet, so the reader states that plainly
 * rather than showing an empty page. Pure + internal for unit testing.
 */
internal fun comicErrorForHttp(code: Int, ext: String?): String = when (code) {
    501 -> {
        val label = ext?.takeIf { it.isNotBlank() }?.let { ".${it.lowercase()}" }
        if (label != null) "This comic format ($label) is not yet supported."
        else "This comic format is not yet supported."
    }
    400 -> "This file is not a supported comic archive (.cbz)."
    401, 403 -> "Authentication required. Please sign in again."
    404 -> "No file linked to this comic. It may not have been scanned yet."
    500 -> "Server error opening the comic. Storage may be unreachable."
    else -> "Comic unavailable (HTTP $code)."
}

/**
 * Android-TV comic reader for `.cbz` archives (media_type == "comic"). On open it
 * fetches `GET /api/v1/entities/{id}/pages` (Bearer auth) for the page count, then
 * renders the current page via Coil from
 * `GET /api/v1/entities/{id}/pages/{n}` — the SAME per-request Bearer auth the
 * image viewer / player use — on a black background.
 *
 * D-pad controls (mirrors [ImageViewerScreen]'s focus-owning Box + onKeyEvent):
 *  - LEFT / RIGHT : previous / next page, clamped to [0, total-1] (no wrap).
 *  - UP / DOWN    : jump to first / last page.
 *  - CENTER / ENTER : toggle zoom (Fit ↔ Crop).
 *  - BACK : exit.
 *
 * Loading + error states are explicit (§11.4.1 — never a blank screen / crash):
 * a spinner while the page list resolves + while each page image transfers, and
 * an honest message for `.cbr` (HTTP 501), an empty archive, or a transport
 * failure. `.cbr` is NOT silently treated as a blank/working comic.
 */
@Composable
fun ComicReaderScreen(
    mediaId: Long,
    onNavigateBack: () -> Unit
) {
    val context = LocalContext.current

    var state by remember(mediaId) { mutableStateOf<ComicReaderUiState>(ComicReaderUiState.Loading) }
    var currentPage by remember(mediaId) { mutableStateOf(0) }
    var title by remember(mediaId) { mutableStateOf("") }
    var authToken by remember { mutableStateOf<String?>(null) }
    var baseUrl by remember { mutableStateOf("") }
    var zoomed by remember { mutableStateOf(false) }

    val focusRequester = remember { androidx.compose.ui.focus.FocusRequester() }

    // Resolve the page count + title for the comic. Re-runs only when the comic
    // changes (paging does NOT re-run this — it only moves currentPage).
    LaunchedEffect(mediaId) {
        state = ComicReaderUiState.Loading
        currentPage = 0
        zoomed = false
        // Archive extension (for the honest .cbr message) — derived from the
        // item's path, defaulted from the load failure path below.
        var ext: String? = null
        try {
            val container = DependencyContainer.getInstance(context)
            authToken = container.authRepository.authState.value.token
            baseUrl = container.getServerUrl().trimEnd('/')

            // Best-effort real title + archive extension (keeps sensible defaults
            // on failure, §11.4.6 — no guessing about the file).
            try {
                val item = container.mediaRepository.getMediaById(mediaId).first()
                if (item != null) {
                    if (item.title.isNotBlank()) title = item.title
                    ext = item.smbPath?.substringAfterLast('.', "")?.lowercase()?.takeIf { it.isNotBlank() }
                }
            } catch (_: Exception) { /* keep prior/empty title */ }

            val response = container.api.getComicPages(mediaId)
            if (response.isSuccessful) {
                val total = response.body()
                    ?.get("total_pages")
                    ?.jsonPrimitive
                    ?.content
                    ?.toIntOrNull() ?: 0
                state = if (total > 0) {
                    ComicReaderUiState.Ready(total)
                } else {
                    ComicReaderUiState.Error("This comic has no readable pages.")
                }
            } else {
                state = ComicReaderUiState.Error(comicErrorForHttp(response.code(), ext))
            }
        } catch (e: java.net.ConnectException) {
            state = ComicReaderUiState.Error("Cannot connect to server. Check it is running and reachable.")
        } catch (e: java.net.SocketTimeoutException) {
            state = ComicReaderUiState.Error("Server request timed out. The server may be busy.")
        } catch (e: Exception) {
            state = ComicReaderUiState.Error("Failed to load comic: ${e.message}")
        }
    }

    val totalPages = (state as? ComicReaderUiState.Ready)?.totalPages ?: 0
    val hasMultiplePages = totalPages >= 2

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
                        if (totalPages > 0) {
                            currentPage = prevComicPage(currentPage, totalPages)
                            zoomed = false
                            true
                        } else false
                    }
                    Key.DirectionRight -> {
                        if (totalPages > 0) {
                            currentPage = nextComicPage(currentPage, totalPages)
                            zoomed = false
                            true
                        } else false
                    }
                    Key.DirectionUp -> {
                        if (totalPages > 0) {
                            currentPage = firstComicPage()
                            zoomed = false
                            true
                        } else false
                    }
                    Key.DirectionDown -> {
                        if (totalPages > 0) {
                            currentPage = lastComicPage(totalPages)
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
        ComicReaderContent(
            title = title,
            state = state,
            pageImageUrl = if (state is ComicReaderUiState.Ready)
                comicPageUrl(baseUrl, mediaId, currentPage) else null,
            pageLabel = comicPageLabel(currentPage, totalPages),
            zoomed = zoomed,
            hasMultiplePages = hasMultiplePages,
            authToken = authToken,
            onBack = onNavigateBack,
            onToggleZoom = { zoomed = !zoomed }
        )
    }

    // Own the D-pad focus once composed (mirrors the image viewer's auto-focus).
    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(200)
        try { focusRequester.requestFocus() } catch (_: Exception) {}
    }
}

/**
 * Stateless visuals for the comic reader — rendered directly to a PNG on the host
 * for the §11.4.170 visual proof (every state × {light,dark}) with no network or
 * device. Carries a focusable Back affordance + zoom toggle so the user always
 * has a visible, D-pad-reachable way out and a way to zoom, and a page indicator
 * so the reader always knows where they are.
 */
@Composable
internal fun ComicReaderContent(
    title: String,
    state: ComicReaderUiState,
    pageImageUrl: String?,
    pageLabel: String?,
    zoomed: Boolean,
    hasMultiplePages: Boolean,
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
            is ComicReaderUiState.Loading -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    CircularProgressIndicator(color = Color.White)
                    Spacer(Modifier.height(12.dp))
                    Text(
                        "Loading comic…",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            }
            is ComicReaderUiState.Error -> {
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
                        "Unable to Open Comic",
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
            is ComicReaderUiState.Ready -> {
                val ctx = LocalContext.current
                SubcomposeAsyncImage(
                    model = ImageRequest.Builder(ctx)
                        .data(pageImageUrl)
                        .apply {
                            // /pages/{n} requires the same Bearer auth the player uses.
                            if (!authToken.isNullOrBlank()) {
                                addHeader("Authorization", "Bearer $authToken")
                            }
                        }
                        .crossfade(true)
                        .build(),
                    contentDescription = pageLabel ?: title.ifBlank { "Comic page" },
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
                                contentDescription = "Page failed to load",
                                tint = Color.White,
                                modifier = Modifier.size(48.dp)
                            )
                            Spacer(Modifier.height(12.dp))
                            Text(
                                "Page failed to load.",
                                style = MaterialTheme.typography.bodyLarge,
                                color = Color.White.copy(alpha = 0.7f)
                            )
                        }
                    }
                )
            }
        }

        // Top overlay: Back affordance + title + page indicator. Always present so
        // the user can read the title + position and exit regardless of state.
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Brush.verticalGradient(listOf(Color.Black.copy(alpha = 0.85f), Color.Transparent)))
                .padding(horizontal = 24.dp, vertical = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            ComicIconButton(
                icon = Icons.Default.ArrowBack,
                contentDescription = "Back",
                onClick = onBack
            )
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title.ifBlank { "Comic" },
                    style = MaterialTheme.typography.headlineSmall,
                    color = Color.White,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                pageLabel?.let {
                    Text(
                        text = it,
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            }
            // Zoom toggle affordance (visible state cue + clickable for pointer
            // devices; D-pad CENTER toggles it too).
            ComicIconButton(
                icon = if (zoomed) Icons.Default.ZoomOutMap else Icons.Default.ZoomIn,
                contentDescription = if (zoomed) "Fit to screen" else "Zoom in",
                onClick = onToggleZoom
            )
        }

        // Page-turn hint (only when a multi-page comic is loaded) — bottom bar.
        if (hasMultiplePages) {
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
                    "Use ◀ ▶ to turn pages",
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
 * Small circular, focusable icon button for the overlay (mirrors the image
 * viewer's top-bar back button). TV [Surface] is focusable + clickable for D-pad.
 */
@Composable
private fun ComicIconButton(
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
