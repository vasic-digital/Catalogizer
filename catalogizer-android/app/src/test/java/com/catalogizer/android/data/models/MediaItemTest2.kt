package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaItemTest2 {

    private fun createMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String = "movie",
        year: Int? = 2024,
        description: String? = "A test movie",
        coverImage: String? = null,
        rating: Double? = 8.5,
        quality: String? = "1080p",
        fileSize: Long? = 1073741824L,
        duration: Int? = 7200,
        directoryPath: String = "/media/movies/test",
        smbPath: String? = null,
        createdAt: String = "2024-01-01T00:00:00Z",
        updatedAt: String = "2024-01-01T00:00:00Z",
        isFavorite: Boolean = false,
        watchProgress: Double = 0.0,
        lastWatched: String? = null,
        isDownloaded: Boolean = false
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        coverImage = coverImage,
        rating = rating,
        quality = quality,
        fileSize = fileSize,
        duration = duration,
        directoryPath = directoryPath,
        smbPath = smbPath,
        createdAt = createdAt,
        updatedAt = updatedAt,
        isFavorite = isFavorite,
        watchProgress = watchProgress,
        lastWatched = lastWatched,
        isDownloaded = isDownloaded
    )

    @Test
    fun `MediaItem defaults are correct`() {
        val item = createMediaItem()
        assertEquals(1L, item.id)
        assertEquals("Test Movie", item.title)
        assertEquals("movie", item.mediaType)
        assertFalse(item.isFavorite)
        assertEquals(0.0, item.watchProgress, 0.001)
        assertFalse(item.isDownloaded)
    }

    @Test
    fun `MediaItem with null optional fields`() {
        val item = createMediaItem(
            year = null,
            description = null,
            coverImage = null,
            rating = null,
            quality = null,
            fileSize = null,
            duration = null,
            smbPath = null,
            lastWatched = null
        )
        assertNull(item.year)
        assertNull(item.description)
        assertNull(item.coverImage)
        assertNull(item.rating)
        assertNull(item.quality)
        assertNull(item.fileSize)
        assertNull(item.duration)
        assertNull(item.smbPath)
        assertNull(item.lastWatched)
    }

    @Test
    fun `MediaItem copy updates specific fields`() {
        val original = createMediaItem()
        val updated = original.copy(
            isFavorite = true,
            watchProgress = 0.75,
            rating = 9.0
        )
        assertTrue(updated.isFavorite)
        assertEquals(0.75, updated.watchProgress, 0.001)
        assertEquals(9.0, updated.rating!!, 0.001)
        // Unchanged fields
        assertEquals(original.id, updated.id)
        assertEquals(original.title, updated.title)
    }

    @Test
    fun `MediaItem equality check`() {
        val item1 = createMediaItem(id = 1L)
        val item2 = createMediaItem(id = 1L)
        assertEquals(item1, item2)
    }

    @Test
    fun `MediaItem inequality when different ids`() {
        val item1 = createMediaItem(id = 1L)
        val item2 = createMediaItem(id = 2L)
        assertNotEquals(item1, item2)
    }

    @Test
    fun `MediaItem with SMB path`() {
        val item = createMediaItem(smbPath = "smb://server/share/movies/test.mkv")
        assertEquals("smb://server/share/movies/test.mkv", item.smbPath)
    }

    @Test
    fun `MediaItem with external metadata`() {
        val metadata = listOf(
            ExternalMetadata(
                id = 1L, mediaId = 1L, provider = "tmdb",
                externalId = "tt0133093", title = "The Matrix",
                lastUpdated = "2024-01-01"
            )
        )
        val item = createMediaItem().copy(externalMetadata = metadata)
        assertNotNull(item.externalMetadata)
        assertEquals(1, item.externalMetadata!!.size)
        assertEquals("tmdb", item.externalMetadata!![0].provider)
    }

    @Test
    fun `MediaItem with versions`() {
        val versions = listOf(
            MediaVersion(
                id = 1L, mediaId = 1L, version = "1.0",
                quality = "1080p", filePath = "/path/to/file.mkv",
                fileSize = 1073741824L
            )
        )
        val item = createMediaItem().copy(versions = versions)
        assertNotNull(item.versions)
        assertEquals(1, item.versions!!.size)
        assertEquals("1080p", item.versions!![0].quality)
    }

    @Test
    fun `MediaItem with download status`() {
        val item = createMediaItem(isDownloaded = true)
        assertTrue(item.isDownloaded)
    }

    @Test
    fun `MediaItem with watch progress and last watched`() {
        val item = createMediaItem(
            watchProgress = 0.5,
            lastWatched = "2024-06-15T10:30:00Z"
        )
        assertEquals(0.5, item.watchProgress, 0.001)
        assertEquals("2024-06-15T10:30:00Z", item.lastWatched)
    }
}
