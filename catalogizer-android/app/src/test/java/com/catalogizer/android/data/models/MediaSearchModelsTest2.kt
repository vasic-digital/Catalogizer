package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaSearchModelsTest2 {

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
    fun `MediaSearchRequest with all fields`() {
        val request = MediaSearchRequest(
            query = "matrix",
            mediaType = "movie",
            yearMin = 1990,
            yearMax = 2005,
            ratingMin = 7.0,
            quality = "1080p",
            sortBy = "rating",
            sortOrder = "desc",
            limit = 50,
            offset = 10
        )
        assertEquals("matrix", request.query)
        assertEquals("movie", request.mediaType)
        assertEquals(1990, request.yearMin)
        assertEquals(2005, request.yearMax)
        assertEquals(7.0, request.ratingMin!!, 0.001)
        assertEquals("1080p", request.quality)
        assertEquals("rating", request.sortBy)
        assertEquals("desc", request.sortOrder)
        assertEquals(50, request.limit)
        assertEquals(10, request.offset)
    }

    // --- MediaSearchResponse ---

    @Test
    fun `MediaSearchResponse with items`() {
        val items = listOf(
            MediaItem(
                id = 1L, title = "Movie 1", mediaType = "movie",
                directoryPath = "/path", createdAt = "2024-01-01", updatedAt = "2024-01-01"
            )
        )
        val response = MediaSearchResponse(items = items, total = 100, limit = 20, offset = 0)
        assertEquals(1, response.items.size)
        assertEquals(100, response.total)
        assertEquals(20, response.limit)
        assertEquals(0, response.offset)
    }

    @Test
    fun `MediaSearchResponse with empty items`() {
        val response = MediaSearchResponse(items = emptyList(), total = 0, limit = 20, offset = 0)
        assertTrue(response.items.isEmpty())
        assertEquals(0, response.total)
    }

    // --- MediaStats ---

    @Test
    fun `MediaStats with data`() {
        val byType = mapOf("movie" to 100, "tv_show" to 50, "music" to 200)
        val byQuality = mapOf("1080p" to 80, "720p" to 60, "4k" to 10)
        val stats = MediaStats(
            totalItems = 350,
            byType = byType,
            byQuality = byQuality,
            totalSize = 1099511627776L,
            recentAdditions = 15
        )
        assertEquals(350, stats.totalItems)
        assertEquals(3, stats.byType.size)
        assertEquals(100, stats.byType["movie"])
        assertEquals(3, stats.byQuality.size)
        assertEquals(1099511627776L, stats.totalSize)
        assertEquals(15, stats.recentAdditions)
    }

    @Test
    fun `MediaStats with empty maps`() {
        val stats = MediaStats(
            totalItems = 0,
            byType = emptyMap(),
            byQuality = emptyMap(),
            totalSize = 0L,
            recentAdditions = 0
        )
        assertEquals(0, stats.totalItems)
        assertTrue(stats.byType.isEmpty())
        assertTrue(stats.byQuality.isEmpty())
    }
}
