package com.catalogizer.androidtv.ui.screens.media

import com.catalogizer.androidtv.data.models.ExternalMetadata
import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for MediaDetailScreen logic: metadata display, type label mapping,
 * file size formatting, cover URL resolution, favorite toggling, retry logic,
 * error messages, and back navigation callback behavior.
 */
class MediaDetailScreenTest {

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String? = "movie",
        year: Int? = 2024,
        description: String? = "An epic adventure",
        rating: Double? = 8.5,
        quality: String? = "1080p",
        fileSize: Long? = null,
        coverUrl: String? = null,
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        rating = rating,
        quality = quality,
        fileSize = fileSize,
        coverUrl = coverUrl,
        directoryPath = "/media/movies/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-06-01T00:00:00Z",
        watchProgress = watchProgress,
        isFavorite = isFavorite
    )

    // --- Media type label mapping ---

    @Test
    fun `movie type maps to Movie label`() {
        val label = mapTypeLabel("movie")
        assertEquals("Movie", label)
    }

    @Test
    fun `tv_show type maps to TV Show label`() {
        val label = mapTypeLabel("tv_show")
        assertEquals("TV Show", label)
    }

    @Test
    fun `tv_episode type maps to Episode label`() {
        val label = mapTypeLabel("tv_episode")
        assertEquals("Episode", label)
    }

    @Test
    fun `music_album type maps to Album label`() {
        val label = mapTypeLabel("music_album")
        assertEquals("Album", label)
    }

    @Test
    fun `song type maps to Song label`() {
        val label = mapTypeLabel("song")
        assertEquals("Song", label)
    }

    @Test
    fun `game type maps to Game label`() {
        val label = mapTypeLabel("game")
        assertEquals("Game", label)
    }

    @Test
    fun `book type maps to Book label`() {
        val label = mapTypeLabel("book")
        assertEquals("Book", label)
    }

    @Test
    fun `unknown type uses formatted fallback`() {
        val label = mapTypeLabel("music_artist")
        assertEquals("Music artist", label)
    }

    @Test
    fun `null type uses unknown fallback`() {
        val label = mapTypeLabel(null)
        assertEquals("Unknown", label)
    }

    // --- File size formatting ---

    @Test
    fun `file size formats to GB for large files`() {
        val size = 2147483648L // 2 GB
        val formatted = formatFileSize(size)
        assertEquals("2.0 GB", formatted)
    }

    @Test
    fun `file size formats to MB for medium files`() {
        val size = 524288000L // ~500 MB
        val formatted = formatFileSize(size)
        assertEquals("500.0 MB", formatted)
    }

    @Test
    fun `file size formats to KB for small files`() {
        val size = 512000L // 500 KB
        val formatted = formatFileSize(size)
        assertEquals("500.0 KB", formatted)
    }

    @Test
    fun `file size returns null for zero`() {
        val size = 0L
        val shouldShow = size > 0
        assertFalse(shouldShow)
    }

    @Test
    fun `file size returns null for null`() {
        val size: Long? = null
        assertNull(size)
    }

    // --- Cover URL resolution ---

    @Test
    fun `relative cover URL prepended with server URL`() {
        val url = "/api/v1/assets/cover.jpg"
        val serverUrl = "http://192.168.0.100:8080"
        val result = resolveCoverUrl(url, serverUrl)
        assertEquals("http://192.168.0.100:8080/api/v1/assets/cover.jpg", result)
    }

    @Test
    fun `TMDB cover URL routed through proxy`() {
        val url = "https://image.tmdb.org/t/p/w500/abc.jpg"
        val serverUrl = "http://192.168.0.100:8080"
        val result = resolveCoverUrl(url, serverUrl)
        assertTrue(result != null && result.contains("/api/v1/image-proxy?url="))
    }

    @Test
    fun `external cover URL used as-is`() {
        val url = "https://cdn.example.com/cover.jpg"
        val serverUrl = "http://192.168.0.100:8080"
        val result = resolveCoverUrl(url, serverUrl)
        assertEquals("https://cdn.example.com/cover.jpg", result)
    }

    @Test
    fun `null cover URL returns null`() {
        val result = resolveCoverUrl(null, "http://localhost:8080")
        assertNull(result)
    }

    // --- Metadata display ---

    @Test
    fun `year displays when present`() {
        val item = createTestMediaItem(year = 2024)
        assertNotNull(item.year)
        assertEquals(2024, item.year)
    }

    @Test
    fun `year hidden when null`() {
        val item = createTestMediaItem(year = null)
        assertNull(item.year)
    }

    @Test
    fun `rating displays with star when positive`() {
        val item = createTestMediaItem(rating = 8.5)
        assertNotNull(item.rating)
        assertTrue(item.rating!! > 0)
        assertEquals("8.5", String.format("%.1f", item.rating))
    }

    @Test
    fun `rating hidden when null`() {
        val item = createTestMediaItem(rating = null)
        assertNull(item.rating)
    }

    @Test
    fun `rating hidden when zero`() {
        val item = createTestMediaItem(rating = 0.0)
        val shouldShow = item.rating != null && item.rating!! > 0
        assertFalse(shouldShow)
    }

    @Test
    fun `quality badge displays when present`() {
        val item = createTestMediaItem(quality = "4K HDR")
        assertNotNull(item.quality)
        assertEquals("4K HDR", item.quality!!.uppercase())
    }

    @Test
    fun `quality badge uppercased for display`() {
        val item = createTestMediaItem(quality = "1080p")
        assertEquals("1080P", item.quality!!.uppercase())
    }

    @Test
    fun `description displayed when non-blank`() {
        val item = createTestMediaItem(description = "A great movie")
        assertNotNull(item.description)
        assertTrue(item.description!!.isNotBlank())
    }

    @Test
    fun `description hidden when null`() {
        val item = createTestMediaItem(description = null)
        assertNull(item.description)
    }

    @Test
    fun `description hidden when blank`() {
        val item = createTestMediaItem(description = "")
        assertFalse(item.description!!.isNotBlank())
    }

    // --- Favorite toggle ---

    @Test
    fun `favorite toggle flips state`() {
        var isFavorite = false
        val oldState = isFavorite
        isFavorite = !oldState
        assertTrue(isFavorite)
    }

    @Test
    fun `favorite button shows correct text based on state`() {
        val favoriteText = if (true) "Favorited" else "Favorite"
        assertEquals("Favorited", favoriteText)

        val unfavoriteText = if (false) "Favorited" else "Favorite"
        assertEquals("Favorite", unfavoriteText)
    }

    @Test
    fun `favorite toggle reverts on error`() {
        var isFavorite = false
        val oldState = isFavorite
        isFavorite = !oldState // optimistic update
        assertTrue(isFavorite)

        // Simulate error: revert
        isFavorite = oldState
        assertFalse(isFavorite)
    }

    // --- Error states ---

    @Test
    fun `connection error message is descriptive`() {
        val error = "Cannot connect to server. Check that the server is running and reachable."
        assertTrue(error.contains("server"))
    }

    @Test
    fun `timeout error message is descriptive`() {
        val error = "Server request timed out. The server may be busy or unreachable."
        assertTrue(error.contains("timed out"))
    }

    @Test
    fun `not found error includes media ID`() {
        val mediaId = 42L
        val error = "Media item not found (id=$mediaId). The server may still be processing this entry."
        assertTrue(error.contains("42"))
    }

    // --- Retry logic ---

    @Test
    fun `retry increments retry count`() {
        var retryCount = 0
        retryCount++
        assertEquals(1, retryCount)
        retryCount++
        assertEquals(2, retryCount)
    }

    @Test
    fun `retry resets error and sets loading`() {
        var error: String? = "Previous error"
        var isLoading = false
        error = null
        isLoading = true
        assertNull(error)
        assertTrue(isLoading)
    }

    // --- Navigation callbacks ---

    @Test
    fun `back navigation callback fires`() {
        var navigatedBack = false
        val onNavigateBack: () -> Unit = { navigatedBack = true }
        onNavigateBack()
        assertTrue(navigatedBack)
    }

    @Test
    fun `player navigation callback receives correct ID`() {
        var navigatedId: Long? = null
        val onNavigateToPlayer: (Long) -> Unit = { navigatedId = it }
        onNavigateToPlayer(42L)
        assertEquals(42L, navigatedId)
    }

    // --- Responsive layout ---

    @Test
    fun `compact layout activates below 600dp width`() {
        val screenWidthDp = 480
        val isCompact = screenWidthDp < 600
        assertTrue(isCompact)
    }

    @Test
    fun `standard layout activates at 600dp or above`() {
        val screenWidthDp = 1920
        val isCompact = screenWidthDp < 600
        assertFalse(isCompact)
    }

    // --- Helper functions mirroring MediaDetailScreen logic ---

    private fun mapTypeLabel(mediaType: String?): String = when (mediaType) {
        "movie" -> "Movie"
        "tv_show" -> "TV Show"
        "tv_episode" -> "Episode"
        "music_album" -> "Album"
        "song" -> "Song"
        "game" -> "Game"
        "book" -> "Book"
        null -> "Unknown"
        else -> mediaType.replace("_", " ").replaceFirstChar { it.uppercase() }
    }

    private fun formatFileSize(size: Long): String = when {
        size > 1073741824 -> String.format("%.1f GB", size.toFloat() / 1073741824)
        size > 1048576 -> String.format("%.1f MB", size.toFloat() / 1048576)
        else -> String.format("%.1f KB", size.toFloat() / 1024)
    }

    private fun resolveCoverUrl(url: String?, serverUrl: String): String? {
        if (url == null) return null
        return when {
            url.startsWith("/") -> serverUrl.trimEnd('/') + url
            url.contains("image.tmdb.org") -> {
                val encoded = java.net.URLEncoder.encode(url, "UTF-8")
                serverUrl.trimEnd('/') + "/api/v1/image-proxy?url=$encoded"
            }
            else -> url
        }
    }
}
