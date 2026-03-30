package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaItemTest2 {

    private fun createItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String? = "movie",
        year: Int? = 2024,
        description: String? = "A test description",
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false,
        coverUrl: String? = null,
        coverImage: String? = null,
        externalMetadata: List<ExternalMetadata>? = null
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        directoryPath = "/media/test",
        createdAt = "2024-01-01",
        updatedAt = "2024-01-01",
        watchProgress = watchProgress,
        isFavorite = isFavorite,
        coverUrl = coverUrl,
        coverImage = coverImage,
        externalMetadata = externalMetadata
    )

    @Test
    fun `hasWatchProgress is false when progress is zero`() {
        val item = createItem(watchProgress = 0.0)
        assertFalse(item.hasWatchProgress)
    }

    @Test
    fun `hasWatchProgress is true when progress is greater than zero`() {
        val item = createItem(watchProgress = 0.1)
        assertTrue(item.hasWatchProgress)
    }

    @Test
    fun `isCompleted is true when progress is 0_9 or more`() {
        val item = createItem(watchProgress = 0.9)
        assertTrue(item.isCompleted)
    }

    @Test
    fun `isCompleted is true when progress is 1_0`() {
        val item = createItem(watchProgress = 1.0)
        assertTrue(item.isCompleted)
    }

    @Test
    fun `isCompleted is false when progress is below 0_9`() {
        val item = createItem(watchProgress = 0.89)
        assertFalse(item.isCompleted)
    }

    @Test
    fun `posterUrl returns coverUrl from external metadata`() {
        val metadata = listOf(
            ExternalMetadata(
                id = 1L, provider = "tmdb", externalId = "123",
                title = "Test", posterUrl = "http://tmdb.org/poster.jpg",
                lastUpdated = "2024-01-01"
            )
        )
        val item = createItem(externalMetadata = metadata)
        assertEquals("http://tmdb.org/poster.jpg", item.posterUrl)
    }

    @Test
    fun `posterUrl falls back to coverUrl when no external metadata`() {
        val item = createItem(coverUrl = "http://local/cover.jpg")
        assertEquals("http://local/cover.jpg", item.posterUrl)
    }

    @Test
    fun `posterUrl falls back to coverImage when no coverUrl`() {
        val item = createItem(coverImage = "http://local/image.jpg")
        assertEquals("http://local/image.jpg", item.posterUrl)
    }

    @Test
    fun `posterUrl is null when no images available`() {
        val item = createItem()
        assertNull(item.posterUrl)
    }

    @Test
    fun `backdropUrl from external metadata`() {
        val metadata = listOf(
            ExternalMetadata(
                id = 1L, provider = "tmdb", externalId = "123",
                title = "Test", backdropUrl = "http://tmdb.org/backdrop.jpg",
                lastUpdated = "2024-01-01"
            )
        )
        val item = createItem(externalMetadata = metadata)
        assertEquals("http://tmdb.org/backdrop.jpg", item.backdropUrl)
    }

    @Test
    fun `backdropUrl is null without external metadata`() {
        val item = createItem()
        assertNull(item.backdropUrl)
    }

    @Test
    fun `genres from external metadata`() {
        val metadata = listOf(
            ExternalMetadata(
                id = 1L, provider = "tmdb", externalId = "123",
                title = "Test", genres = listOf("Action", "Sci-Fi"),
                lastUpdated = "2024-01-01"
            )
        )
        val item = createItem(externalMetadata = metadata)
        assertEquals(2, item.genres.size)
        assertTrue(item.genres.contains("Action"))
    }

    @Test
    fun `genres empty without external metadata`() {
        val item = createItem()
        assertTrue(item.genres.isEmpty())
    }

    @Test
    fun `cast from external metadata`() {
        val metadata = listOf(
            ExternalMetadata(
                id = 1L, provider = "tmdb", externalId = "123",
                title = "Test", cast = listOf("Actor A", "Actor B"),
                lastUpdated = "2024-01-01"
            )
        )
        val item = createItem(externalMetadata = metadata)
        assertEquals(2, item.cast.size)
    }

    @Test
    fun `cast empty without external metadata`() {
        val item = createItem()
        assertTrue(item.cast.isEmpty())
    }

    @Test
    fun `thumbnailUrl matches posterUrl`() {
        val item = createItem(coverUrl = "http://local/cover.jpg")
        assertEquals(item.posterUrl, item.thumbnailUrl)
    }

    @Test
    fun `default values are correct`() {
        val item = MediaItem()
        assertEquals(0L, item.id)
        assertEquals("", item.title)
        assertNull(item.mediaType)
        assertEquals("", item.directoryPath)
        assertFalse(item.isFavorite)
        assertEquals(0.0, item.watchProgress, 0.001)
        assertFalse(item.isDownloaded)
    }
}
