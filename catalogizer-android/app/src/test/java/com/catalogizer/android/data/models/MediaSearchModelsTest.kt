package com.catalogizer.android.data.models

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class MediaSearchModelsTest {

    private lateinit var json: Json

    @Before
    fun setup() {
        json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
        }
    }

    // --- MediaSearchRequest ---

    @Test
    fun `MediaSearchRequest has correct defaults`() {
        val request = MediaSearchRequest()

        assertNull(request.query)
        assertNull(request.mediaType)
        assertNull(request.yearMin)
        assertNull(request.yearMax)
        assertNull(request.ratingMin)
        assertNull(request.quality)
        assertNull(request.sortBy)
        assertNull(request.sortOrder)
        assertEquals(20, request.limit)
        assertEquals(0, request.offset)
    }

    @Test
    fun `MediaSearchRequest serializes with snake_case`() {
        val request = MediaSearchRequest(
            query = "inception",
            mediaType = "movie",
            yearMin = 2010,
            yearMax = 2024,
            ratingMin = 7.0,
            sortBy = "rating",
            sortOrder = "desc",
            limit = 50,
            offset = 10
        )

        val jsonStr = json.encodeToString(request)

        assertTrue(jsonStr.contains("\"media_type\":\"movie\""))
        assertTrue(jsonStr.contains("\"year_min\":2010"))
        assertTrue(jsonStr.contains("\"year_max\":2024"))
        assertTrue(jsonStr.contains("\"rating_min\":7.0"))
        assertTrue(jsonStr.contains("\"sort_by\":\"rating\""))
        assertTrue(jsonStr.contains("\"sort_order\":\"desc\""))
    }

    @Test
    fun `MediaSearchRequest survives round-trip serialization`() {
        val original = MediaSearchRequest(
            query = "test",
            mediaType = "movie",
            yearMin = 2000,
            yearMax = 2024,
            ratingMin = 5.0,
            quality = "1080p",
            sortBy = "title",
            sortOrder = "asc",
            limit = 30,
            offset = 5
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MediaSearchRequest>(serialized)

        assertEquals(original, deserialized)
    }

    // --- MediaSearchResponse ---

    @Test
    fun `MediaSearchResponse deserializes correctly`() {
        val jsonStr = """{
            "items": [
                {
                    "id": 1,
                    "title": "Movie 1",
                    "media_type": "movie",
                    "directory_path": "/movies/1",
                    "created_at": "2024-01-01",
                    "updated_at": "2024-01-01"
                }
            ],
            "total": 100,
            "limit": 20,
            "offset": 0
        }"""

        val response = json.decodeFromString<MediaSearchResponse>(jsonStr)

        assertEquals(1, response.items.size)
        assertEquals("Movie 1", response.items[0].title)
        assertEquals(100, response.total)
        assertEquals(20, response.limit)
        assertEquals(0, response.offset)
    }

    @Test
    fun `MediaSearchResponse with empty items`() {
        val jsonStr = """{
            "items": [],
            "total": 0,
            "limit": 20,
            "offset": 0
        }"""

        val response = json.decodeFromString<MediaSearchResponse>(jsonStr)

        assertTrue(response.items.isEmpty())
        assertEquals(0, response.total)
    }

    // --- MediaStats ---

    @Test
    fun `MediaStats deserializes correctly`() {
        val jsonStr = """{
            "total_items": 500,
            "by_type": {"movie": 200, "tv_show": 150, "music": 100, "other": 50},
            "by_quality": {"1080p": 300, "720p": 150, "4k": 50},
            "total_size": 5000000000000,
            "recent_additions": 25
        }"""

        val stats = json.decodeFromString<MediaStats>(jsonStr)

        assertEquals(500, stats.totalItems)
        assertEquals(200, stats.byType["movie"])
        assertEquals(150, stats.byType["tv_show"])
        assertEquals(300, stats.byQuality["1080p"])
        assertEquals(5000000000000L, stats.totalSize)
        assertEquals(25, stats.recentAdditions)
    }

    @Test
    fun `MediaStats survives round-trip serialization`() {
        val original = MediaStats(
            totalItems = 100,
            byType = mapOf("movie" to 50, "music" to 50),
            byQuality = mapOf("1080p" to 80, "720p" to 20),
            totalSize = 1000000000L,
            recentAdditions = 10
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MediaStats>(serialized)

        assertEquals(original, deserialized)
    }

    // --- ExternalMetadata ---

    @Test
    fun `ExternalMetadata deserializes correctly`() {
        val jsonStr = """{
            "id": 1,
            "media_id": 42,
            "provider": "tmdb",
            "external_id": "tt1375666",
            "title": "Inception",
            "description": "A thief who steals secrets",
            "year": 2010,
            "rating": 8.8,
            "poster_url": "http://image.tmdb.org/poster.jpg",
            "backdrop_url": "http://image.tmdb.org/backdrop.jpg",
            "genres": ["Action", "Sci-Fi", "Thriller"],
            "cast": ["Leonardo DiCaprio", "Tom Hardy"],
            "crew": ["Christopher Nolan"],
            "last_updated": "2024-01-01T00:00:00Z"
        }"""

        val metadata = json.decodeFromString<ExternalMetadata>(jsonStr)

        assertEquals(1L, metadata.id)
        assertEquals(42L, metadata.mediaId)
        assertEquals("tmdb", metadata.provider)
        assertEquals("tt1375666", metadata.externalId)
        assertEquals("Inception", metadata.title)
        assertEquals(2010, metadata.year)
        assertEquals(8.8, metadata.rating!!, 0.01)
        assertEquals(3, metadata.genres?.size)
        assertTrue(metadata.genres?.contains("Action") == true)
        assertEquals(2, metadata.cast?.size)
    }

    // --- MediaVersion ---

    @Test
    fun `MediaVersion deserializes correctly`() {
        val jsonStr = """{
            "id": 1,
            "media_id": 42,
            "version": "1.0",
            "quality": "1080p",
            "file_path": "/media/movie.mkv",
            "file_size": 4000000000,
            "codec": "h265",
            "resolution": "1920x1080",
            "bitrate": 5000000,
            "language": "en"
        }"""

        val version = json.decodeFromString<MediaVersion>(jsonStr)

        assertEquals(1L, version.id)
        assertEquals(42L, version.mediaId)
        assertEquals("1.0", version.version)
        assertEquals("1080p", version.quality)
        assertEquals("/media/movie.mkv", version.filePath)
        assertEquals(4000000000L, version.fileSize)
        assertEquals("h265", version.codec)
        assertEquals("1920x1080", version.resolution)
        assertEquals(5000000L, version.bitrate)
        assertEquals("en", version.language)
    }

    @Test
    fun `MediaVersion with optional fields missing`() {
        val jsonStr = """{
            "id": 1,
            "media_id": 42,
            "version": "1.0",
            "quality": "720p",
            "file_path": "/media/movie.mp4",
            "file_size": 2000000000
        }"""

        val version = json.decodeFromString<MediaVersion>(jsonStr)

        assertNull(version.codec)
        assertNull(version.resolution)
        assertNull(version.bitrate)
        assertNull(version.language)
    }
}
