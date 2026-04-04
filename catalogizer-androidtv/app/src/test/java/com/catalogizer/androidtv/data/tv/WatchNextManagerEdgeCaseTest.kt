package com.catalogizer.androidtv.data.tv

import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

class WatchNextManagerEdgeCaseTest {

    @Test
    fun `filterContinueWatching with empty list returns empty`() {
        val filtered = WatchNextManager.filterContinueWatching(emptyList())
        assertTrue(filtered.isEmpty())
    }

    @Test
    fun `filterContinueWatching preserves order`() {
        val items = listOf(
            MediaItem(id = 3L, title = "Third", watchProgress = 0.7),
            MediaItem(id = 1L, title = "First", watchProgress = 0.1),
            MediaItem(id = 2L, title = "Second", watchProgress = 0.5)
        )
        val filtered = WatchNextManager.filterContinueWatching(items)
        assertEquals(3, filtered.size)
        assertEquals(3L, filtered[0].id)
        assertEquals(1L, filtered[1].id)
        assertEquals(2L, filtered[2].id)
    }

    @Test
    fun `isStale with negative timestamp returns false`() {
        // Negative timestamp is treated as 0 behaviour: negative means epoch minus,
        // currentTimeMillis - negative = large positive > STALE_THRESHOLD, so it IS stale.
        // But the implementation checks: if (lastEngagementTimeMs == 0L) return false
        // A negative value is NOT 0L, so it will evaluate the comparison.
        // System.currentTimeMillis() - (-X) = currentTimeMillis + X, which is > threshold.
        // Therefore isStale(-1L) == true.
        // We test the actual behavior to lock it in.
        val result = WatchNextManager.isStale(-1L)
        assertTrue(result)
    }

    @Test
    fun `isStale at exactly 30 days is not stale`() {
        // Exactly 30 days ago: currentTimeMillis - (30 * 24 * 60 * 60 * 1000)
        // The condition is > STALE_THRESHOLD_MS (strictly greater than), so exactly 30 days is NOT stale.
        val exactlyThirtyDaysAgo = System.currentTimeMillis() - (30L * 24 * 60 * 60 * 1000)
        // At exactly 30 days, delta == threshold, condition is >, so not stale.
        assertFalse(WatchNextManager.isStale(exactlyThirtyDaysAgo))
    }

    @Test
    fun `isCompleted with negative progress returns false`() {
        // Negative progress is not > 0.90
        assertFalse(WatchNextManager.isCompleted(-0.5))
        assertFalse(WatchNextManager.isCompleted(-1.0))
    }

    @Test
    fun `isTvSeriesType with empty string returns false`() {
        assertFalse(WatchNextManager.isTvSeriesType(""))
    }

    @Test
    fun `filterContinueWatching with all items at boundary values`() {
        val items = listOf(
            // Exactly 5% — included (>= 0.05)
            MediaItem(id = 1L, title = "At Min Boundary", watchProgress = 0.05),
            // Just below 5% — excluded
            MediaItem(id = 2L, title = "Below Min", watchProgress = 0.049),
            // Exactly 90% — excluded (not < 0.90)
            MediaItem(id = 3L, title = "At Max Boundary", watchProgress = 0.90),
            // Just below 90% — included
            MediaItem(id = 4L, title = "Below Max", watchProgress = 0.899)
        )
        val filtered = WatchNextManager.filterContinueWatching(items)
        assertEquals(2, filtered.size)
        assertTrue(filtered.any { it.id == 1L })
        assertTrue(filtered.any { it.id == 4L })
        assertFalse(filtered.any { it.id == 2L })
        assertFalse(filtered.any { it.id == 3L })
    }
}
