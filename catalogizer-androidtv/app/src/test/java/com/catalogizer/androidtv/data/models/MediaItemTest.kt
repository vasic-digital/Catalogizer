package com.catalogizer.androidtv.data.models

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class MediaItemTest {

    private lateinit var json: Json

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String = "movie",
        year: Int? = 2024,
        description: String? = "A test movie",
        coverImage: String? = null,
        rating: Double? = 8.5,
        directoryPath: String = "/media/movies/test",
        createdAt: String = "2024-01-01T00:00:00Z",
        updatedAt: String = "2024-06-01T00:00:00Z",
        externalMetadata: List<ExternalMetadata>? = null,
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false,
        isDownloaded: Boolean = false
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        coverImage = coverImage,
        rating = rating,
        directoryPath = directoryPath,
        createdAt = createdAt,
        updatedAt = updatedAt,
        externalMetadata = externalMetadata,
        watchProgress = watchProgress,
        isFavorite = isFavorite,
        isDownloaded = isDownloaded
    )

    private fun createTestMetadata() = ExternalMetadata(
        id = 1L,
        mediaId = 1L,
        provider = "tmdb",
        externalId = "tt1375666",
        title = "Inception",
        posterUrl = "http://img.tmdb.org/poster.jpg",
        backdropUrl = "http://img.tmdb.org/backdrop.jpg",
        genres = listOf("Action", "Sci-Fi"),
        cast = listOf("Leonardo DiCaprio", "Tom Hardy"),
        lastUpdated = "2024-01-01"
    )

    @Before
    fun setup() {
        json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
        }
    }

    // --- Computed properties ---

    @Test
    fun `posterUrl returns external metadata poster when available`() {
        val metadata = createTestMetadata()
        val item = createTestMediaItem(
            externalMetadata = listOf(metadata),
            coverImage = "http://local/cover.jpg"
        )

        assertEquals("http://img.tmdb.org/poster.jpg", item.posterUrl)
    }

    @Test
    fun `posterUrl returns coverImage when no metadata`() {
        val item = createTestMediaItem(coverImage = "http://local/cover.jpg")

        assertEquals("http://local/cover.jpg", item.posterUrl)
    }

    @Test
    fun `posterUrl returns null when no metadata and no cover`() {
        val item = createTestMediaItem(coverImage = null, externalMetadata = null)

        assertNull(item.posterUrl)
    }

    @Test
    fun `backdropUrl returns external metadata backdrop`() {
        val metadata = createTestMetadata()
        val item = createTestMediaItem(externalMetadata = listOf(metadata))

        assertEquals("http://img.tmdb.org/backdrop.jpg", item.backdropUrl)
    }

    @Test
    fun `backdropUrl returns null when no metadata`() {
        val item = createTestMediaItem()

        assertNull(item.backdropUrl)
    }

    @Test
    fun `thumbnailUrl returns posterUrl`() {
        val item = createTestMediaItem(coverImage = "http://local/cover.jpg")

        assertEquals(item.posterUrl, item.thumbnailUrl)
    }

    @Test
    fun `genres returns genres from external metadata`() {
        val metadata = createTestMetadata()
        val item = createTestMediaItem(externalMetadata = listOf(metadata))

        assertEquals(listOf("Action", "Sci-Fi"), item.genres)
    }

    @Test
    fun `genres returns empty list when no metadata`() {
        val item = createTestMediaItem()

        assertTrue(item.genres.isEmpty())
    }

    @Test
    fun `cast returns cast from external metadata`() {
        val metadata = createTestMetadata()
        val item = createTestMediaItem(externalMetadata = listOf(metadata))

        assertEquals(listOf("Leonardo DiCaprio", "Tom Hardy"), item.cast)
    }

    @Test
    fun `cast returns empty list when no metadata`() {
        val item = createTestMediaItem()

        assertTrue(item.cast.isEmpty())
    }

    @Test
    fun `hasWatchProgress returns true when progress greater than zero`() {
        val item = createTestMediaItem(watchProgress = 0.5)

        assertTrue(item.hasWatchProgress)
    }

    @Test
    fun `hasWatchProgress returns false when progress is zero`() {
        val item = createTestMediaItem(watchProgress = 0.0)

        assertFalse(item.hasWatchProgress)
    }

    @Test
    fun `isCompleted returns true when progress at 90 percent or more`() {
        assertTrue(createTestMediaItem(watchProgress = 0.9).isCompleted)
        assertTrue(createTestMediaItem(watchProgress = 0.95).isCompleted)
        assertTrue(createTestMediaItem(watchProgress = 1.0).isCompleted)
    }

    @Test
    fun `isCompleted returns false when progress below 90 percent`() {
        assertFalse(createTestMediaItem(watchProgress = 0.0).isCompleted)
        assertFalse(createTestMediaItem(watchProgress = 0.5).isCompleted)
        assertFalse(createTestMediaItem(watchProgress = 0.89).isCompleted)
    }

    // --- Serialization ---

    @Test
    fun `MediaItem serializes with snake_case`() {
        val item = createTestMediaItem()
        val jsonStr = json.encodeToString(item)

        assertTrue(jsonStr.contains("\"media_type\":\"movie\""))
        assertTrue(jsonStr.contains("\"directory_path\":\"/media/movies/test\""))
    }

    @Test
    fun `MediaItem round-trip serialization`() {
        val original = createTestMediaItem(id = 99, title = "Round Trip", isFavorite = true)
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MediaItem>(serialized)

        assertEquals(original.id, deserialized.id)
        assertEquals(original.title, deserialized.title)
        assertEquals(original.mediaType, deserialized.mediaType)
        assertEquals(original.isFavorite, deserialized.isFavorite)
    }

    @Test
    fun `MediaItem equality works correctly`() {
        val item1 = createTestMediaItem(id = 1)
        val item2 = createTestMediaItem(id = 1)
        val item3 = createTestMediaItem(id = 2)

        assertEquals(item1, item2)
        assertNotEquals(item1, item3)
    }
}
