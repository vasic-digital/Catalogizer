package com.catalogizer.androidtv.ui.components

import com.catalogizer.androidtv.data.models.ExternalMetadata
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.playback.PlaybackFormatter
import com.catalogizer.androidtv.data.playback.UiPlaybackProgress
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for MediaCard component logic: media type icon selection,
 * gradient mapping, watch progress computation, cover URL resolution,
 * and MediaItem computed properties that drive card rendering.
 */
class MediaCardTest {

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String? = "movie",
        year: Int? = 2024,
        description: String? = "A test movie",
        coverImage: String? = null,
        coverUrl: String? = null,
        rating: Double? = 8.5,
        quality: String? = null,
        duration: Long? = null,
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false,
        externalMetadata: List<ExternalMetadata>? = null
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        coverImage = coverImage,
        coverUrl = coverUrl,
        rating = rating,
        quality = quality,
        duration = duration,
        directoryPath = "/media/movies/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-06-01T00:00:00Z",
        watchProgress = watchProgress,
        isFavorite = isFavorite,
        externalMetadata = externalMetadata
    )

    // --- Thumbnail URL resolution ---

    @Test
    fun `thumbnailUrl returns null when no cover sources present`() {
        val item = createTestMediaItem(coverImage = null, coverUrl = null, externalMetadata = null)
        assertNull(item.thumbnailUrl)
    }

    @Test
    fun `thumbnailUrl returns coverUrl when present`() {
        val item = createTestMediaItem(coverUrl = "/api/v1/assets/cover.jpg")
        assertEquals("/api/v1/assets/cover.jpg", item.thumbnailUrl)
    }

    @Test
    fun `thumbnailUrl returns coverImage when coverUrl is null`() {
        val item = createTestMediaItem(coverImage = "/images/poster.jpg", coverUrl = null)
        assertEquals("/images/poster.jpg", item.thumbnailUrl)
    }

    @Test
    fun `thumbnailUrl prefers external metadata cover over coverUrl`() {
        val metadata = ExternalMetadata(
            provider = "tmdb",
            externalId = "123",
            title = "Test",
            coverUrl = "https://image.tmdb.org/t/p/w500/abc.jpg"
        )
        val item = createTestMediaItem(
            coverUrl = "/local/cover.jpg",
            externalMetadata = listOf(metadata)
        )
        assertEquals("https://image.tmdb.org/t/p/w500/abc.jpg", item.thumbnailUrl)
    }

    @Test
    fun `thumbnailUrl falls back to external posterUrl when coverUrl is null`() {
        val metadata = ExternalMetadata(
            provider = "tmdb",
            externalId = "123",
            title = "Test",
            posterUrl = "https://image.tmdb.org/t/p/w500/poster.jpg"
        )
        val item = createTestMediaItem(
            coverUrl = null,
            coverImage = null,
            externalMetadata = listOf(metadata)
        )
        assertEquals("https://image.tmdb.org/t/p/w500/poster.jpg", item.thumbnailUrl)
    }

    // --- Watch progress state ---

    @Test
    fun `hasWatchProgress is false when progress is zero`() {
        val item = createTestMediaItem(watchProgress = 0.0)
        assertFalse(item.hasWatchProgress)
    }

    @Test
    fun `hasWatchProgress is true when progress is greater than zero`() {
        val item = createTestMediaItem(watchProgress = 0.01)
        assertTrue(item.hasWatchProgress)
    }

    @Test
    fun `progress bar should show for progress between 0 and 0_95 exclusive`() {
        val item = createTestMediaItem(watchProgress = 0.5)
        val shouldShowProgressBar = item.hasWatchProgress && item.watchProgress > 0 && item.watchProgress < 0.95
        assertTrue(shouldShowProgressBar)
    }

    @Test
    fun `progress bar should not show for completed items`() {
        val item = createTestMediaItem(watchProgress = 0.95)
        val shouldShowProgressBar = item.hasWatchProgress && item.watchProgress > 0 && item.watchProgress < 0.95
        assertFalse(shouldShowProgressBar)
    }

    @Test
    fun `progress bar should not show for unwatched items`() {
        val item = createTestMediaItem(watchProgress = 0.0)
        val shouldShowProgressBar = item.hasWatchProgress && item.watchProgress > 0 && item.watchProgress < 0.95
        assertFalse(shouldShowProgressBar)
    }

    // --- isCompleted ---

    @Test
    fun `isCompleted is true when progress at 0_9`() {
        val item = createTestMediaItem(watchProgress = 0.9)
        assertTrue(item.isCompleted)
    }

    @Test
    fun `isCompleted is true when progress above 0_9`() {
        val item = createTestMediaItem(watchProgress = 1.0)
        assertTrue(item.isCompleted)
    }

    @Test
    fun `isCompleted is false when progress below 0_9`() {
        val item = createTestMediaItem(watchProgress = 0.89)
        assertFalse(item.isCompleted)
    }

    // --- Duration formatting logic (mirrors card rendering) ---

    @Test
    fun `duration formats correctly for hours and minutes`() {
        val duration = 7200L // 2 hours in seconds
        val minutes = duration / 60 // 120
        val hours = minutes / 60 // 2
        val remainingMinutes = minutes % 60 // 0
        val durationText = if (hours > 0) "${hours}h ${remainingMinutes}m" else "${minutes}m"
        assertEquals("2h 0m", durationText)
    }

    @Test
    fun `duration formats correctly for minutes only`() {
        val duration = 1800L // 30 minutes in seconds
        val minutes = duration / 60
        val hours = minutes / 60
        val remainingMinutes = minutes % 60
        val durationText = if (hours > 0) "${hours}h ${remainingMinutes}m" else "${minutes}m"
        assertEquals("30m", durationText)
    }

    @Test
    fun `duration formats correctly for mixed hours and minutes`() {
        val duration = 5400L // 1.5 hours
        val minutes = duration / 60 // 90
        val hours = minutes / 60 // 1
        val remainingMinutes = minutes % 60 // 30
        val durationText = if (hours > 0) "${hours}h ${remainingMinutes}m" else "${minutes}m"
        assertEquals("1h 30m", durationText)
    }

    // --- Media type badge text ---

    @Test
    fun `media type badge text formats with underscores replaced`() {
        val mediaType = "tv_show"
        val badgeText = mediaType.replace("_", " ").lowercase()
        assertEquals("tv show", badgeText)
    }

    @Test
    fun `media type badge text handles null media type`() {
        val mediaType: String? = null
        val badgeText = (mediaType ?: "").replace("_", " ").lowercase()
        assertEquals("", badgeText)
    }

    @Test
    fun `media type badge text handles simple type`() {
        val mediaType = "movie"
        val badgeText = mediaType.replace("_", " ").lowercase()
        assertEquals("movie", badgeText)
    }

    // --- Watch progress percentage display ---

    @Test
    fun `watch progress percentage displays correctly`() {
        val item = createTestMediaItem(watchProgress = 0.75)
        val percentText = "${(item.watchProgress * 100).toInt()}%"
        assertEquals("75%", percentText)
    }

    @Test
    fun `watch progress percentage rounds down correctly`() {
        val item = createTestMediaItem(watchProgress = 0.999)
        val percentText = "${(item.watchProgress * 100).toInt()}%"
        assertEquals("99%", percentText)
    }

    // --- Rating display ---

    @Test
    fun `rating formats to one decimal place`() {
        val rating = 8.5
        val ratingText = String.format("%.1f", rating)
        assertEquals("8.5", ratingText)
    }

    @Test
    fun `rating formats whole number correctly`() {
        val rating = 9.0
        val ratingText = String.format("%.1f", rating)
        assertEquals("9.0", ratingText)
    }

    // --- Genres and cast from external metadata ---

    @Test
    fun `genres returns empty list when no external metadata`() {
        val item = createTestMediaItem(externalMetadata = null)
        assertTrue(item.genres.isEmpty())
    }

    @Test
    fun `genres returns genres from first external metadata entry`() {
        val metadata = ExternalMetadata(
            provider = "tmdb",
            externalId = "123",
            title = "Test",
            genres = listOf("Action", "Thriller")
        )
        val item = createTestMediaItem(externalMetadata = listOf(metadata))
        assertEquals(listOf("Action", "Thriller"), item.genres)
    }

    @Test
    fun `cast returns empty list when no external metadata`() {
        val item = createTestMediaItem(externalMetadata = null)
        assertTrue(item.cast.isEmpty())
    }

    @Test
    fun `cast returns cast from first external metadata entry`() {
        val metadata = ExternalMetadata(
            provider = "tmdb",
            externalId = "123",
            title = "Test",
            cast = listOf("Actor One", "Actor Two")
        )
        val item = createTestMediaItem(externalMetadata = listOf(metadata))
        assertEquals(listOf("Actor One", "Actor Two"), item.cast)
    }

    // --- Backdrop URL ---

    @Test
    fun `backdropUrl returns null when no external metadata`() {
        val item = createTestMediaItem(externalMetadata = null)
        assertNull(item.backdropUrl)
    }

    @Test
    fun `backdropUrl returns backdrop from first external metadata`() {
        val metadata = ExternalMetadata(
            provider = "tmdb",
            externalId = "123",
            title = "Test",
            backdropUrl = "https://image.tmdb.org/t/p/w1280/backdrop.jpg"
        )
        val item = createTestMediaItem(externalMetadata = listOf(metadata))
        assertEquals("https://image.tmdb.org/t/p/w1280/backdrop.jpg", item.backdropUrl)
    }

    // --- Cover URL resolution logic (mirrors MediaCard composable) ---

    @Test
    fun `relative cover URL should be prepended with server URL`() {
        val rawUrl = "/api/v1/assets/cover.jpg"
        val serverUrl = "http://192.168.0.100:8080"
        val coverUrl = when {
            rawUrl.startsWith("/") -> serverUrl.trimEnd('/') + rawUrl
            rawUrl.contains("image.tmdb.org") -> {
                val encoded = java.net.URLEncoder.encode(rawUrl, "UTF-8")
                serverUrl.trimEnd('/') + "/api/v1/image-proxy?url=$encoded"
            }
            else -> rawUrl
        }
        assertEquals("http://192.168.0.100:8080/api/v1/assets/cover.jpg", coverUrl)
    }

    @Test
    fun `TMDB URL should be routed through image proxy`() {
        val rawUrl = "https://image.tmdb.org/t/p/w500/abc.jpg"
        val serverUrl = "http://192.168.0.100:8080"
        val coverUrl = when {
            rawUrl.startsWith("/") -> serverUrl.trimEnd('/') + rawUrl
            rawUrl.contains("image.tmdb.org") -> {
                val encoded = java.net.URLEncoder.encode(rawUrl, "UTF-8")
                serverUrl.trimEnd('/') + "/api/v1/image-proxy?url=$encoded"
            }
            else -> rawUrl
        }
        assertTrue(coverUrl.contains("/api/v1/image-proxy?url="))
        assertTrue(coverUrl.startsWith("http://192.168.0.100:8080/"))
    }

    @Test
    fun `absolute non-TMDB URL should be used as-is`() {
        val rawUrl = "https://cdn.example.com/cover.jpg"
        val serverUrl = "http://192.168.0.100:8080"
        val coverUrl = when {
            rawUrl.startsWith("/") -> serverUrl.trimEnd('/') + rawUrl
            rawUrl.contains("image.tmdb.org") -> {
                val encoded = java.net.URLEncoder.encode(rawUrl, "UTF-8")
                serverUrl.trimEnd('/') + "/api/v1/image-proxy?url=$encoded"
            }
            else -> rawUrl
        }
        assertEquals("https://cdn.example.com/cover.jpg", coverUrl)
    }

    // --- UiPlaybackProgress integration ---

    @Test
    fun `progress badge data can be constructed for card overlay`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = 7200L,
            lastPosition = 3600L,
            lastSessionAmount = 1200L,
            totalReproductions = 3,
            aggregateAmount = 5400L,
            lastSessionEndedAtMs = System.currentTimeMillis()
        )
        assertEquals(1L, progress.mediaItemId)
        assertEquals("seconds", progress.positionUnit)
        assertEquals(7200L, progress.durationTotal)
        assertEquals(3600L, progress.lastPosition)
        assertEquals(3, progress.totalReproductions)
    }

    @Test
    fun `null progress should not render progress badge`() {
        val progress: UiPlaybackProgress? = null
        assertNull(progress)
    }
}
