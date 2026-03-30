package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class PlaybackProgressTest {

    @Test
    fun `progressPercentage is correct for half watched`() {
        val progress = PlaybackProgress(
            mediaId = 1L,
            position = 5000L,
            duration = 10000L
        )
        assertEquals(0.5, progress.progressPercentage, 0.001)
    }

    @Test
    fun `progressPercentage is zero when position is zero`() {
        val progress = PlaybackProgress(
            mediaId = 1L,
            position = 0L,
            duration = 10000L
        )
        assertEquals(0.0, progress.progressPercentage, 0.001)
    }

    @Test
    fun `progressPercentage is 1_0 when fully watched`() {
        val progress = PlaybackProgress(
            mediaId = 1L,
            position = 10000L,
            duration = 10000L
        )
        assertEquals(1.0, progress.progressPercentage, 0.001)
    }

    @Test
    fun `progressPercentage is zero when duration is zero`() {
        val progress = PlaybackProgress(
            mediaId = 1L,
            position = 5000L,
            duration = 0L
        )
        assertEquals(0.0, progress.progressPercentage, 0.001)
    }

    @Test
    fun `progressPercentage with large values`() {
        val progress = PlaybackProgress(
            mediaId = 1L,
            position = 7200000L,  // 2 hours in ms
            duration = 9000000L   // 2.5 hours in ms
        )
        assertEquals(0.8, progress.progressPercentage, 0.001)
    }

    @Test
    fun `fields are correct`() {
        val progress = PlaybackProgress(
            mediaId = 42L,
            position = 1000L,
            duration = 5000L,
            timestamp = 1234567890L
        )
        assertEquals(42L, progress.mediaId)
        assertEquals(1000L, progress.position)
        assertEquals(5000L, progress.duration)
        assertEquals(1234567890L, progress.timestamp)
    }

    @Test
    fun `timestamp defaults to current time`() {
        val before = System.currentTimeMillis()
        val progress = PlaybackProgress(mediaId = 1L, position = 0L, duration = 1000L)
        val after = System.currentTimeMillis()

        assertTrue(progress.timestamp >= before)
        assertTrue(progress.timestamp <= after)
    }
}
