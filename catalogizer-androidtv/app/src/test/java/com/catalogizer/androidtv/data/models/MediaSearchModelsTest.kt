package com.catalogizer.androidtv.data.models

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
    fun `MediaSearchRequest serializes correctly`() {
        val request = MediaSearchRequest(
            query = "inception",
            mediaType = "movie",
            yearMin = 2010,
            sortBy = "rating"
        )

        val jsonStr = json.encodeToString(request)

        // kotlinx.serialization uses Kotlin property names (camelCase) since
        // the class uses Gson @SerializedName annotations (not kotlinx @SerialName).
        assertTrue(jsonStr.contains("\"mediaType\":\"movie\""))
        assertTrue(jsonStr.contains("\"yearMin\":2010"))
        assertTrue(jsonStr.contains("\"sortBy\":\"rating\""))
    }

    @Test
    fun `MediaSearchResponse deserializes correctly`() {
        val jsonStr = """{
            "items": [],
            "total": 0,
            "limit": 20,
            "offset": 0
        }"""

        val response = json.decodeFromString<MediaSearchResponse>(jsonStr)

        assertTrue(response.items.isEmpty())
        assertEquals(0, response.total)
        assertEquals(20, response.limit)
    }

    @Test
    fun `MediaStats deserializes correctly`() {
        // kotlinx.serialization uses Kotlin property names (camelCase) since
        // the class uses Gson @SerializedName annotations (not kotlinx @SerialName).
        val jsonStr = """{
            "totalItems": 500,
            "byType": {"movie": 200, "music": 300},
            "byQuality": {"1080p": 300, "720p": 200},
            "totalSize": 5000000000000,
            "recentAdditions": 25
        }"""

        val stats = json.decodeFromString<MediaStats>(jsonStr)

        assertEquals(500, stats.totalItems)
        assertEquals(200, stats.byType["movie"])
        assertEquals(5000000000000L, stats.totalSize)
    }

    @Test
    fun `MediaSearchRequest round-trip`() {
        val original = MediaSearchRequest(
            query = "test",
            mediaType = "movie",
            limit = 30
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<MediaSearchRequest>(serialized)

        assertEquals(original, deserialized)
    }
}
