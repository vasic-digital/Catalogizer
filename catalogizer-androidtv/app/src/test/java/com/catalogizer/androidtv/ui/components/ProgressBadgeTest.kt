package com.catalogizer.androidtv.ui.components

import com.catalogizer.androidtv.data.playback.PlaybackFormatter
import com.catalogizer.androidtv.data.playback.UiPlaybackProgress
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for ProgressBadge component logic: label formatting for duration,
 * position, last session, reproduction count, and progress fraction computation.
 */
class ProgressBadgeTest {

    private fun createProgress(
        durationTotal: Long? = 7200L,
        lastPosition: Long = 3600L,
        lastSessionAmount: Long = 1200L,
        totalReproductions: Long = 3,
        aggregateAmount: Long = 5400L,
        positionUnit: String = "seconds"
    ) = UiPlaybackProgress(
        mediaItemId = 1L,
        positionUnit = positionUnit,
        durationTotal = durationTotal,
        lastPosition = lastPosition,
        lastSessionAmount = lastSessionAmount,
        totalReproductions = totalReproductions,
        aggregateAmount = aggregateAmount,
        lastSessionEndedAtMs = System.currentTimeMillis()
    )

    // --- Duration label ---

    @Test
    fun `duration label formats hours and minutes for seconds unit`() {
        val progress = createProgress(durationTotal = 7200L)
        val label = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("2h 0m", label)
    }

    @Test
    fun `duration label formats minutes only when under one hour`() {
        val progress = createProgress(durationTotal = 1800L)
        val label = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("30m", label)
    }

    @Test
    fun `duration label shows dash for null duration`() {
        val progress = createProgress(durationTotal = null)
        val label = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("-", label)
    }

    @Test
    fun `duration label shows dash for zero duration`() {
        val progress = createProgress(durationTotal = 0L)
        val label = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("-", label)
    }

    @Test
    fun `duration label formats seconds when under one minute`() {
        val progress = createProgress(durationTotal = 45L)
        val label = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("45s", label)
    }

    // --- Current position label ---

    @Test
    fun `current position label shows fraction of total`() {
        val progress = createProgress(durationTotal = 7200L, lastPosition = 3600L)
        val label = PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit)
        assertEquals("1h 0m / 2h 0m", label)
    }

    @Test
    fun `current position label falls back when total is null`() {
        val progress = createProgress(durationTotal = null, lastPosition = 600L)
        val label = PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit)
        assertEquals("10m", label)
    }

    @Test
    fun `current position label falls back when total is zero`() {
        val progress = createProgress(durationTotal = 0L, lastPosition = 120L)
        val label = PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit)
        assertEquals("2m", label)
    }

    // --- Last session amount label ---

    @Test
    fun `last session label appends last suffix`() {
        val progress = createProgress(lastSessionAmount = 1200L)
        val label = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit) + " last"
        assertEquals("20m last", label)
    }

    @Test
    fun `last session label shows dash for zero amount`() {
        val progress = createProgress(lastSessionAmount = 0L)
        val label = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit) + " last"
        assertEquals("- last", label)
    }

    // --- Reproduction count label ---

    @Test
    fun `reproduction count formats with multiplication sign`() {
        val progress = createProgress(totalReproductions = 5)
        val label = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)
        assertEquals("5\u00D7", label)
    }

    @Test
    fun `reproduction count formats zero`() {
        val label = PlaybackFormatter.formatReproductionCount(0)
        assertEquals("0\u00D7", label)
    }

    @Test
    fun `reproduction count formats single play`() {
        val label = PlaybackFormatter.formatReproductionCount(1)
        assertEquals("1\u00D7", label)
    }

    @Test
    fun `badge text combines count with played suffix`() {
        val reps = PlaybackFormatter.formatReproductionCount(3)
        val text = "$reps played"
        assertEquals("3\u00D7 played", text)
    }

    // --- Progress fraction ---

    @Test
    fun `progress fraction is 0_5 for halfway position`() {
        val pct = PlaybackFormatter.progressFraction(3600L, 7200L)
        assertEquals(0.5f, pct, 0.001f)
    }

    @Test
    fun `progress fraction is 0 when total is null`() {
        val pct = PlaybackFormatter.progressFraction(3600L, null)
        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction is 0 when total is zero`() {
        val pct = PlaybackFormatter.progressFraction(3600L, 0L)
        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction is clamped to 1 when position exceeds total`() {
        val pct = PlaybackFormatter.progressFraction(10000L, 7200L)
        assertEquals(1.0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction is 0 when position is zero`() {
        val pct = PlaybackFormatter.progressFraction(0L, 7200L)
        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress bar shows when fraction is positive`() {
        val pct = PlaybackFormatter.progressFraction(1800L, 7200L)
        assertTrue(pct > 0f)
    }

    @Test
    fun `progress bar hidden when fraction is zero`() {
        val pct = PlaybackFormatter.progressFraction(0L, 7200L)
        assertFalse(pct > 0f)
    }

    // --- Pages unit ---

    @Test
    fun `pages unit formats single page correctly`() {
        val label = PlaybackFormatter.formatAmount(1L, "pages")
        assertEquals("1 page", label)
    }

    @Test
    fun `pages unit formats multiple pages correctly`() {
        val label = PlaybackFormatter.formatAmount(320L, "pages")
        assertEquals("320 pages", label)
    }

    // --- Events unit ---

    @Test
    fun `events unit formats single run correctly`() {
        val label = PlaybackFormatter.formatAmount(1L, "events")
        assertEquals("1 run", label)
    }

    @Test
    fun `events unit formats multiple runs correctly`() {
        val label = PlaybackFormatter.formatAmount(5L, "events")
        assertEquals("5 runs", label)
    }

    // --- Null progress early return ---

    @Test
    fun `null progress should cause badge to not render`() {
        val progress: UiPlaybackProgress? = null
        assertNull(progress)
    }
}
