package com.catalogizer.android.data.models

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
        quality: String? = "1080p",
        fileSize: Long? = 4_000_000_000L,
        duration: Int? = 7200,
        directoryPath: String = "/media/movies/test",
        smbPath: String? = null,
        createdAt: String = "2024-01-01T00:00:00Z",
        updatedAt: String = "2024-06-01T00:00:00Z",
        isFavorite: Boolean = false,
        watchProgress: Double = 0.0,
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
        isDownloaded = isDownloaded
    )

    @Before
    fun setup() {
        json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
        }
    }

    @Test
    fun `MediaItem serializes with snake_case field names`() {
        val item = createTestMediaItem()
        val jsonStr = json.encodeToString(item)

        assertTrue(jsonStr.contains("\"media_type\":\"movie\""))
        assertTrue(jsonStr.contains("\"directory_path\":\"/media/movies/test\""))
        assertTrue(jsonStr.contains("\"created_at\":\"2024-01-01T00:00:00Z\""))
        assertTrue(jsonStr.contains("\"updated_at\":\"2024-06-01T00:00:00Z\""))
    }

    @Test
    fun `MediaItem deserializes from snake_case JSON`() {
        val jsonStr = """{
            "id": 42,
            "title": "Inception",
            "media_type": "movie",
            "year": 2010,
            "description": "A mind-bending thriller",
            "cover_image": "http://example.com/inception.jpg",
            "rating": 8.8,
            "quality": "4k",
            "file_size": 8000000000,
            "duration": 8880,
            "directory_path": "/media/movies/inception",
            "smb_path": "//server/movies/inception",
            "created_at": "2024-01-15T00:00:00Z",
            "updated_at": "2024-06-15T00:00:00Z",
            "isFavorite": true,
            "watchProgress": 0.75,
            "isDownloaded": true
        }"""

        val item = json.decodeFromString<MediaItem>(jsonStr)

        assertEquals(42L, item.id)
        assertEquals("Inception", item.title)
        assertEquals("movie", item.mediaType)
        assertEquals(2010, item.year)
        assertEquals("A mind-bending thriller", item.description)
        assertEquals("http://example.com/inception.jpg", item.coverImage)
        assertEquals(8.8, item.rating!!, 0.01)
        assertEquals("4k", item.quality)
        assertEquals(8000000000L, item.fileSize)
        assertEquals(8880, item.duration)
        assertEquals("/media/movies/inception", item.directoryPath)
        assertEquals("//server/movies/inception", item.smbPath)
        assertTrue(item.isFavorite)
        assertEquals(0.75, item.watchProgress, 0.01)
        assertTrue(item.isDownloaded)
    }

    @Test
    fun `MediaItem deserializes with optional fields missing`() {
        val jsonStr = """{
            "id": 1,
            "title": "Simple Item",
            "media_type": "music",
            "directory_path": "/media/music",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        }"""

        val item = json.decodeFromString<MediaItem>(jsonStr)

        assertEquals(1L, item.id)
        assertEquals("Simple Item", item.title)
        assertEquals("music", item.mediaType)
        assertNull(item.year)
        assertNull(item.description)
        assertNull(item.coverImage)
        assertNull(item.rating)
        assertNull(item.quality)
        assertNull(item.fileSize)
        assertNull(item.duration)
        assertNull(item.smbPath)
        assertFalse(item.isFavorite)
        assertEquals(0.0, item.watchProgress, 0.01)
        assertFalse(item.isDownloaded)
    }

    @Test
    fun `MediaItem survives round-trip serialization`() {
        val original = createTestMediaItem(
            id = 99,
            title = "Round Trip",
            isFavorite = true,
            watchProgress = 0.5,
            isDownloaded = true
        )
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MediaItem>(serialized)

        assertEquals(original.id, deserialized.id)
        assertEquals(original.title, deserialized.title)
        assertEquals(original.mediaType, deserialized.mediaType)
        assertEquals(original.year, deserialized.year)
        assertEquals(original.description, deserialized.description)
        assertEquals(original.rating, deserialized.rating)
        assertEquals(original.quality, deserialized.quality)
        assertEquals(original.fileSize, deserialized.fileSize)
        assertEquals(original.duration, deserialized.duration)
        assertEquals(original.directoryPath, deserialized.directoryPath)
        assertEquals(original.isFavorite, deserialized.isFavorite)
        assertEquals(original.watchProgress, deserialized.watchProgress, 0.01)
        assertEquals(original.isDownloaded, deserialized.isDownloaded)
    }

    @Test
    fun `MediaItem copy creates independent instance`() {
        val original = createTestMediaItem()
        val copy = original.copy(title = "Modified Title", isFavorite = true)

        assertNotEquals(original.title, copy.title)
        assertEquals("Modified Title", copy.title)
        assertTrue(copy.isFavorite)
        assertFalse(original.isFavorite)
        assertEquals(original.id, copy.id)
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
