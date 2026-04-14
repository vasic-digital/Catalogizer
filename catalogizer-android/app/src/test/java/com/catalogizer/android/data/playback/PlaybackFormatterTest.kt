package com.catalogizer.android.data.playback

import org.junit.Assert.*
import org.junit.Test

class PlaybackFormatterTest {

    // --- formatSeconds ---

    @Test
    fun `formatSeconds returns dash for zero`() {
        assertEquals("-", PlaybackFormatter.formatSeconds(0L))
    }

    @Test
    fun `formatSeconds returns dash for negative`() {
        assertEquals("-", PlaybackFormatter.formatSeconds(-1L))
        assertEquals("-", PlaybackFormatter.formatSeconds(-100L))
    }

    @Test
    fun `formatSeconds formats seconds only`() {
        assertEquals("1s", PlaybackFormatter.formatSeconds(1L))
        assertEquals("30s", PlaybackFormatter.formatSeconds(30L))
        assertEquals("59s", PlaybackFormatter.formatSeconds(59L))
    }

    @Test
    fun `formatSeconds formats minutes and omits seconds`() {
        assertEquals("1m", PlaybackFormatter.formatSeconds(60L))
        assertEquals("2m", PlaybackFormatter.formatSeconds(120L))
        assertEquals("59m", PlaybackFormatter.formatSeconds(3540L))
    }

    @Test
    fun `formatSeconds formats minutes at boundary`() {
        assertEquals("1m", PlaybackFormatter.formatSeconds(61L))
        assertEquals("1m", PlaybackFormatter.formatSeconds(90L))
        assertEquals("1m", PlaybackFormatter.formatSeconds(119L))
    }

    @Test
    fun `formatSeconds formats hours and minutes`() {
        assertEquals("1h 0m", PlaybackFormatter.formatSeconds(3600L))
        assertEquals("1h 30m", PlaybackFormatter.formatSeconds(5400L))
        assertEquals("2h 0m", PlaybackFormatter.formatSeconds(7200L))
    }

    @Test
    fun `formatSeconds formats large duration`() {
        assertEquals("10h 0m", PlaybackFormatter.formatSeconds(36000L))
        assertEquals("24h 0m", PlaybackFormatter.formatSeconds(86400L))
        assertEquals("100h 0m", PlaybackFormatter.formatSeconds(360000L))
    }

    @Test
    fun `formatSeconds formats 1h 1m`() {
        assertEquals("1h 1m", PlaybackFormatter.formatSeconds(3661L))
    }

    @Test
    fun `formatSeconds formats 2h 59m`() {
        assertEquals("2h 59m", PlaybackFormatter.formatSeconds(10740L))
    }

    // --- formatAmount ---

    @Test
    fun `formatAmount returns dash for zero value`() {
        assertEquals("-", PlaybackFormatter.formatAmount(0L, "seconds"))
        assertEquals("-", PlaybackFormatter.formatAmount(0L, "pages"))
        assertEquals("-", PlaybackFormatter.formatAmount(0L, "events"))
    }

    @Test
    fun `formatAmount returns dash for negative value`() {
        assertEquals("-", PlaybackFormatter.formatAmount(-5L, "seconds"))
        assertEquals("-", PlaybackFormatter.formatAmount(-1L, "pages"))
    }

    @Test
    fun `formatAmount formats seconds via formatSeconds`() {
        assertEquals("30s", PlaybackFormatter.formatAmount(30L, "seconds"))
        assertEquals("1h 30m", PlaybackFormatter.formatAmount(5400L, "seconds"))
    }

    @Test
    fun `formatAmount formats singular page`() {
        assertEquals("1 page", PlaybackFormatter.formatAmount(1L, "pages"))
    }

    @Test
    fun `formatAmount formats plural pages`() {
        assertEquals("2 pages", PlaybackFormatter.formatAmount(2L, "pages"))
        assertEquals("100 pages", PlaybackFormatter.formatAmount(100L, "pages"))
    }

    @Test
    fun `formatAmount formats singular run`() {
        assertEquals("1 run", PlaybackFormatter.formatAmount(1L, "events"))
    }

    @Test
    fun `formatAmount formats plural runs`() {
        assertEquals("2 runs", PlaybackFormatter.formatAmount(2L, "events"))
        assertEquals("50 runs", PlaybackFormatter.formatAmount(50L, "events"))
    }

    @Test
    fun `formatAmount formats unknown unit with raw value`() {
        assertEquals("5 chapters", PlaybackFormatter.formatAmount(5L, "chapters"))
        assertEquals("1 tracks", PlaybackFormatter.formatAmount(1L, "tracks"))
    }

    // --- formatProgress ---

    @Test
    fun `formatProgress with null total falls back to formatAmount`() {
        assertEquals("30s", PlaybackFormatter.formatProgress(30L, null, "seconds"))
    }

    @Test
    fun `formatProgress with zero total falls back to formatAmount`() {
        assertEquals("30s", PlaybackFormatter.formatProgress(30L, 0L, "seconds"))
    }

    @Test
    fun `formatProgress with negative total falls back to formatAmount`() {
        assertEquals("30s", PlaybackFormatter.formatProgress(30L, -1L, "seconds"))
    }

    @Test
    fun `formatProgress formats current and total seconds`() {
        assertEquals("30s / 1m", PlaybackFormatter.formatProgress(30L, 60L, "seconds"))
    }

    @Test
    fun `formatProgress formats current and total pages`() {
        assertEquals("50 pages / 200 pages", PlaybackFormatter.formatProgress(50L, 200L, "pages"))
    }

    @Test
    fun `formatProgress formats current and total events`() {
        assertEquals("3 runs / 10 runs", PlaybackFormatter.formatProgress(3L, 10L, "events"))
    }

    @Test
    fun `formatProgress with zero current and valid total`() {
        assertEquals("- / 1h 0m", PlaybackFormatter.formatProgress(0L, 3600L, "seconds"))
    }

    @Test
    fun `formatProgress with equal current and total`() {
        assertEquals("100 pages / 100 pages", PlaybackFormatter.formatProgress(100L, 100L, "pages"))
    }

    // --- progressFraction ---

    @Test
    fun `progressFraction returns 0 for null total`() {
        assertEquals(0f, PlaybackFormatter.progressFraction(50L, null), 0.001f)
    }

    @Test
    fun `progressFraction returns 0 for zero total`() {
        assertEquals(0f, PlaybackFormatter.progressFraction(50L, 0L), 0.001f)
    }

    @Test
    fun `progressFraction returns 0 for negative total`() {
        assertEquals(0f, PlaybackFormatter.progressFraction(50L, -10L), 0.001f)
    }

    @Test
    fun `progressFraction calculates correct percentage`() {
        assertEquals(0.5f, PlaybackFormatter.progressFraction(50L, 100L), 0.001f)
        assertEquals(0.25f, PlaybackFormatter.progressFraction(25L, 100L), 0.001f)
        assertEquals(0.75f, PlaybackFormatter.progressFraction(75L, 100L), 0.001f)
    }

    @Test
    fun `progressFraction returns 1 for complete`() {
        assertEquals(1f, PlaybackFormatter.progressFraction(100L, 100L), 0.001f)
    }

    @Test
    fun `progressFraction clamps to 1 when current exceeds total`() {
        assertEquals(1f, PlaybackFormatter.progressFraction(150L, 100L), 0.001f)
        assertEquals(1f, PlaybackFormatter.progressFraction(200L, 100L), 0.001f)
    }

    @Test
    fun `progressFraction returns 0 for zero current`() {
        assertEquals(0f, PlaybackFormatter.progressFraction(0L, 100L), 0.001f)
    }

    @Test
    fun `progressFraction handles small fractions`() {
        assertEquals(0.01f, PlaybackFormatter.progressFraction(1L, 100L), 0.001f)
    }

    @Test
    fun `progressFraction handles large values without overflow`() {
        val current = 500_000_000L
        val total = 1_000_000_000L
        assertEquals(0.5f, PlaybackFormatter.progressFraction(current, total), 0.001f)
    }

    // --- formatReproductionCount ---

    @Test
    fun `formatReproductionCount formats zero`() {
        assertEquals("0\u00D7", PlaybackFormatter.formatReproductionCount(0L))
    }

    @Test
    fun `formatReproductionCount formats single`() {
        assertEquals("1\u00D7", PlaybackFormatter.formatReproductionCount(1L))
    }

    @Test
    fun `formatReproductionCount formats multiple`() {
        assertEquals("5\u00D7", PlaybackFormatter.formatReproductionCount(5L))
        assertEquals("100\u00D7", PlaybackFormatter.formatReproductionCount(100L))
    }
}
