package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaSearchResponseTest {

    private fun createItem(id: Long, title: String) = MediaItem(
        id = id, title = title, directoryPath = "/path",
        createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @Test
    fun `allItems returns items when items is non-empty`() {
        val response = MediaSearchResponse(
            items = listOf(createItem(1, "Movie A"), createItem(2, "Movie B")),
            mediaItems = listOf(createItem(3, "Movie C")),
            total = 3
        )
        assertEquals(2, response.allItems.size)
        assertEquals("Movie A", response.allItems[0].title)
    }

    @Test
    fun `allItems falls back to mediaItems when items is null`() {
        val response = MediaSearchResponse(
            items = null,
            mediaItems = listOf(createItem(1, "Media Item")),
            total = 1
        )
        assertEquals(1, response.allItems.size)
        assertEquals("Media Item", response.allItems[0].title)
    }

    @Test
    fun `allItems falls back to mediaItems when items is empty`() {
        val response = MediaSearchResponse(
            items = emptyList(),
            mediaItems = listOf(createItem(1, "Fallback")),
            total = 1
        )
        assertEquals(1, response.allItems.size)
        assertEquals("Fallback", response.allItems[0].title)
    }

    @Test
    fun `allItems returns empty when both are null`() {
        val response = MediaSearchResponse(
            items = null,
            mediaItems = null,
            total = 0
        )
        assertTrue(response.allItems.isEmpty())
    }

    @Test
    fun `allItems returns empty when both are empty`() {
        val response = MediaSearchResponse(
            items = emptyList(),
            mediaItems = emptyList(),
            total = 0
        )
        assertTrue(response.allItems.isEmpty())
    }

    @Test
    fun `defaults are correct`() {
        val response = MediaSearchResponse()
        assertNull(response.items)
        assertNull(response.mediaItems)
        assertEquals(0, response.total)
        assertEquals(20, response.limit)
        assertEquals(0, response.offset)
    }

    @Test
    fun `total and pagination fields`() {
        val response = MediaSearchResponse(
            items = listOf(createItem(1, "A")),
            total = 500,
            limit = 50,
            offset = 100
        )
        assertEquals(500, response.total)
        assertEquals(50, response.limit)
        assertEquals(100, response.offset)
    }
}
