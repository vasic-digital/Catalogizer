package com.catalogizer.android.data.local

import org.junit.Assert.*
import org.junit.Test

class WatchProgressDaoTest {

    @Test
    fun `WatchProgress has correct defaults`() {
        val progress = WatchProgress(
            mediaId = 1L,
            progress = 0.5
        )

        assertEquals(1L, progress.mediaId)
        assertEquals(0.5, progress.progress, 0.01)
        assertTrue(progress.lastWatched > 0)
        assertTrue(progress.updatedAt > 0)
    }

    @Test
    fun `WatchProgress equality works correctly`() {
        val ts = System.currentTimeMillis()
        val p1 = WatchProgress(mediaId = 1L, progress = 0.5, lastWatched = ts, updatedAt = ts)
        val p2 = WatchProgress(mediaId = 1L, progress = 0.5, lastWatched = ts, updatedAt = ts)
        val p3 = WatchProgress(mediaId = 2L, progress = 0.5, lastWatched = ts, updatedAt = ts)

        assertEquals(p1, p2)
        assertNotEquals(p1, p3)
    }

    @Test
    fun `WatchProgress copy updates correctly`() {
        val original = WatchProgress(mediaId = 1L, progress = 0.3)
        val updated = original.copy(progress = 0.8)

        assertEquals(0.3, original.progress, 0.01)
        assertEquals(0.8, updated.progress, 0.01)
        assertEquals(original.mediaId, updated.mediaId)
    }

    @Test
    fun `WatchProgress progress ranges`() {
        val zero = WatchProgress(mediaId = 1L, progress = 0.0)
        val half = WatchProgress(mediaId = 2L, progress = 0.5)
        val full = WatchProgress(mediaId = 3L, progress = 1.0)

        assertEquals(0.0, zero.progress, 0.01)
        assertEquals(0.5, half.progress, 0.01)
        assertEquals(1.0, full.progress, 0.01)
    }

    @Test
    fun `WatchProgress uses mediaId as primary key`() {
        val p1 = WatchProgress(mediaId = 42L, progress = 0.75)
        assertEquals(42L, p1.mediaId)
    }
}
