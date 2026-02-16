package com.catalogizer.android.data.local

import org.junit.Assert.*
import org.junit.Test

class SearchHistoryTest {

    @Test
    fun `SearchHistory has correct defaults`() {
        val history = SearchHistory(query = "inception")

        assertEquals(0L, history.id)
        assertEquals("inception", history.query)
        assertTrue(history.timestamp > 0)
        assertEquals(0, history.resultsCount)
    }

    @Test
    fun `SearchHistory with all fields`() {
        val history = SearchHistory(
            id = 1L,
            query = "inception",
            timestamp = 1000L,
            resultsCount = 42
        )

        assertEquals(1L, history.id)
        assertEquals("inception", history.query)
        assertEquals(1000L, history.timestamp)
        assertEquals(42, history.resultsCount)
    }

    @Test
    fun `SearchHistory equality works correctly`() {
        val h1 = SearchHistory(id = 1, query = "test", timestamp = 1000L, resultsCount = 5)
        val h2 = SearchHistory(id = 1, query = "test", timestamp = 1000L, resultsCount = 5)
        val h3 = SearchHistory(id = 2, query = "test", timestamp = 1000L, resultsCount = 5)

        assertEquals(h1, h2)
        assertNotEquals(h1, h3)
    }

    @Test
    fun `SearchHistory copy updates correctly`() {
        val original = SearchHistory(query = "test")
        val updated = original.copy(resultsCount = 10)

        assertEquals(0, original.resultsCount)
        assertEquals(10, updated.resultsCount)
        assertEquals(original.query, updated.query)
    }

    @Test
    fun `SearchHistory handles empty query`() {
        val history = SearchHistory(query = "")
        assertEquals("", history.query)
    }
}
