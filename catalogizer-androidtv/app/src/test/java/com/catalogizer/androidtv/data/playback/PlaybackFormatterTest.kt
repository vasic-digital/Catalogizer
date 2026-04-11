package com.catalogizer.androidtv.data.playback

import org.junit.Assert.assertEquals
import org.junit.Test

class PlaybackFormatterTest {

    @Test
    fun `formatAmount seconds renders hours and minutes`() {
        assertEquals("1h 23m", PlaybackFormatter.formatAmount(4980L, "seconds"))
        assertEquals("23m", PlaybackFormatter.formatAmount(1380L, "seconds"))
        assertEquals("45s", PlaybackFormatter.formatAmount(45L, "seconds"))
    }

    @Test
    fun `formatAmount pages uses singular for 1`() {
        assertEquals("1 page", PlaybackFormatter.formatAmount(1L, "pages"))
        assertEquals("120 pages", PlaybackFormatter.formatAmount(120L, "pages"))
    }

    @Test
    fun `formatAmount events uses singular for 1`() {
        assertEquals("1 run", PlaybackFormatter.formatAmount(1L, "events"))
        assertEquals("3 runs", PlaybackFormatter.formatAmount(3L, "events"))
    }

    @Test
    fun `formatAmount zero or negative returns placeholder`() {
        assertEquals("-", PlaybackFormatter.formatAmount(0L, "seconds"))
        assertEquals("-", PlaybackFormatter.formatAmount(-1L, "pages"))
    }

    @Test
    fun `formatProgress without total falls back to current`() {
        assertEquals("23m", PlaybackFormatter.formatProgress(1380L, null, "seconds"))
        assertEquals("23m", PlaybackFormatter.formatProgress(1380L, 0L, "seconds"))
    }

    @Test
    fun `formatProgress with total renders slash`() {
        assertEquals("30m / 2h 0m", PlaybackFormatter.formatProgress(1800L, 7200L, "seconds"))
        assertEquals("140 pages / 320 pages", PlaybackFormatter.formatProgress(140L, 320L, "pages"))
    }

    @Test
    fun `progressFraction clamps and handles null total`() {
        assertEquals(0f, PlaybackFormatter.progressFraction(100L, null), 0.001f)
        assertEquals(0f, PlaybackFormatter.progressFraction(100L, 0L), 0.001f)
        assertEquals(0.5f, PlaybackFormatter.progressFraction(50L, 100L), 0.001f)
        assertEquals(1f, PlaybackFormatter.progressFraction(200L, 100L), 0.001f)
    }

    @Test
    fun `formatReproductionCount appends multiplier`() {
        assertEquals("0×", PlaybackFormatter.formatReproductionCount(0L))
        assertEquals("3×", PlaybackFormatter.formatReproductionCount(3L))
    }
}
