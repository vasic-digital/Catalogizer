package com.catalogizer.androidtv.data.tv

import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

class WatchNextManagerTest {

    @Test
    fun `filterContinueWatching includes items between 5% and 90%`() {
        val items = listOf(
            MediaItem(id = 1L, title = "Too Early", watchProgress = 0.02),
            MediaItem(id = 2L, title = "In Progress", watchProgress = 0.5),
            MediaItem(id = 3L, title = "Almost Done", watchProgress = 0.85),
            MediaItem(id = 4L, title = "Completed", watchProgress = 0.95)
        )
        val filtered = WatchNextManager.filterContinueWatching(items)
        assertEquals(2, filtered.size)
        assertEquals(2L, filtered[0].id)
        assertEquals(3L, filtered[1].id)
    }

    @Test
    fun `filterContinueWatching returns empty for no in-progress items`() {
        val items = listOf(
            MediaItem(id = 1L, title = "Not Started", watchProgress = 0.0),
            MediaItem(id = 2L, title = "Done", watchProgress = 1.0)
        )
        val filtered = WatchNextManager.filterContinueWatching(items)
        assertTrue(filtered.isEmpty())
    }

    @Test
    fun `filterContinueWatching boundary at exactly 5% is included`() {
        val item = MediaItem(id = 1L, title = "At 5%", watchProgress = 0.05)
        val filtered = WatchNextManager.filterContinueWatching(listOf(item))
        assertEquals(1, filtered.size)
    }

    @Test
    fun `filterContinueWatching boundary at exactly 90% is excluded`() {
        val item = MediaItem(id = 1L, title = "At 90%", watchProgress = 0.90)
        val filtered = WatchNextManager.filterContinueWatching(listOf(item))
        assertTrue(filtered.isEmpty())
    }

    @Test
    fun `isStale returns true for items older than 30 days`() {
        val thirtyOneDaysAgo = System.currentTimeMillis() - (31L * 24 * 60 * 60 * 1000)
        assertTrue(WatchNextManager.isStale(thirtyOneDaysAgo))
    }

    @Test
    fun `isStale returns false for recent items`() {
        val oneDayAgo = System.currentTimeMillis() - (1L * 24 * 60 * 60 * 1000)
        assertFalse(WatchNextManager.isStale(oneDayAgo))
    }

    @Test
    fun `isStale returns false for zero timestamp`() {
        assertFalse(WatchNextManager.isStale(0L))
    }

    @Test
    fun `isCompleted returns true for progress above 90%`() {
        assertTrue(WatchNextManager.isCompleted(0.91))
        assertTrue(WatchNextManager.isCompleted(1.0))
    }

    @Test
    fun `isCompleted returns false for progress at or below 90%`() {
        assertFalse(WatchNextManager.isCompleted(0.90))
        assertFalse(WatchNextManager.isCompleted(0.5))
        assertFalse(WatchNextManager.isCompleted(0.0))
    }

    @Test
    fun `isTvSeriesType returns true for tv_show and tv_episode`() {
        assertTrue(WatchNextManager.isTvSeriesType("tv_show"))
        assertTrue(WatchNextManager.isTvSeriesType("tv_episode"))
    }

    @Test
    fun `isTvSeriesType returns false for other types`() {
        assertFalse(WatchNextManager.isTvSeriesType("movie"))
        assertFalse(WatchNextManager.isTvSeriesType("music"))
        assertFalse(WatchNextManager.isTvSeriesType(null))
    }
}
