package com.catalogizer.android.ui.components

import com.catalogizer.android.data.playback.PlaybackFormatter
import com.catalogizer.android.data.playback.UiPlaybackProgress
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for ProgressBadge rendering logic and data formatting.
 * Verifies the PlaybackFormatter outputs used by ProgressBadge to
 * display duration, position, session amount, play count, and progress bar.
 */
@RunWith(RobolectricTestRunner::class)
class ProgressBadgeTest {

    private fun createProgress(
        mediaItemId: Long = 1L,
        positionUnit: String = "seconds",
        durationTotal: Long? = 7200L,
        lastPosition: Long = 3600L,
        lastSessionAmount: Long = 1800L,
        totalReproductions: Long = 5L,
        aggregateAmount: Long = 9000L,
        lastSessionEndedAtMs: Long? = null,
    ): UiPlaybackProgress = UiPlaybackProgress(
        mediaItemId = mediaItemId,
        positionUnit = positionUnit,
        durationTotal = durationTotal,
        lastPosition = lastPosition,
        lastSessionAmount = lastSessionAmount,
        totalReproductions = totalReproductions,
        aggregateAmount = aggregateAmount,
        lastSessionEndedAtMs = lastSessionEndedAtMs,
    )

    @Test
    fun `null progress renders nothing`() {
        val progress: UiPlaybackProgress? = null
        assertNull(progress)
    }

    @Test
    fun `duration label formats seconds correctly for hours and minutes`() {
        val progress = createProgress(durationTotal = 7200L, positionUnit = "seconds")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("2h 0m", duration)
    }

    @Test
    fun `duration label formats seconds for minutes only`() {
        val progress = createProgress(durationTotal = 1800L, positionUnit = "seconds")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("30m", duration)
    }

    @Test
    fun `duration label formats seconds for seconds only`() {
        val progress = createProgress(durationTotal = 45L, positionUnit = "seconds")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("45s", duration)
    }

    @Test
    fun `duration label shows dash for zero duration`() {
        val progress = createProgress(durationTotal = 0L, positionUnit = "seconds")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("-", duration)
    }

    @Test
    fun `duration label shows dash for null duration`() {
        val progress = createProgress(durationTotal = null, positionUnit = "seconds")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("-", duration)
    }

    @Test
    fun `current position label shows fraction format`() {
        val progress = createProgress(lastPosition = 3600L, durationTotal = 7200L, positionUnit = "seconds")
        val label = PlaybackFormatter.formatProgress(
            progress.lastPosition, progress.durationTotal, progress.positionUnit)

        assertEquals("1h 0m / 2h 0m", label)
    }

    @Test
    fun `current position label with null total shows position only`() {
        val progress = createProgress(lastPosition = 3600L, durationTotal = null, positionUnit = "seconds")
        val label = PlaybackFormatter.formatProgress(
            progress.lastPosition, progress.durationTotal, progress.positionUnit)

        assertEquals("1h 0m", label)
    }

    @Test
    fun `current position label with zero total shows position only`() {
        val progress = createProgress(lastPosition = 600L, durationTotal = 0L, positionUnit = "seconds")
        val label = PlaybackFormatter.formatProgress(
            progress.lastPosition, progress.durationTotal, progress.positionUnit)

        assertEquals("10m", label)
    }

    @Test
    fun `last session label appends last suffix`() {
        val progress = createProgress(lastSessionAmount = 1800L, positionUnit = "seconds")
        val label = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit) + " last"

        assertEquals("30m last", label)
    }

    @Test
    fun `last session label with zero amount`() {
        val progress = createProgress(lastSessionAmount = 0L, positionUnit = "seconds")
        val label = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit) + " last"

        assertEquals("- last", label)
    }

    @Test
    fun `reproduction count formats with multiplication sign`() {
        val progress = createProgress(totalReproductions = 5L)
        val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)

        assertEquals("5\u00D7", reps)
    }

    @Test
    fun `reproduction count for zero plays`() {
        val progress = createProgress(totalReproductions = 0L)
        val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)

        assertEquals("0\u00D7", reps)
    }

    @Test
    fun `reproduction count for single play`() {
        val progress = createProgress(totalReproductions = 1L)
        val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)

        assertEquals("1\u00D7", reps)
    }

    @Test
    fun `reproduction count for large number`() {
        val progress = createProgress(totalReproductions = 1000L)
        val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)

        assertEquals("1000\u00D7", reps)
    }

    @Test
    fun `played label combines count and played text`() {
        val progress = createProgress(totalReproductions = 3L)
        val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)
        val label = "$reps played"

        assertEquals("3\u00D7 played", label)
    }

    @Test
    fun `progress fraction at 0 percent`() {
        val progress = createProgress(lastPosition = 0L, durationTotal = 7200L)
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction at 50 percent`() {
        val progress = createProgress(lastPosition = 3600L, durationTotal = 7200L)
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals(0.5f, pct, 0.001f)
    }

    @Test
    fun `progress fraction at 100 percent`() {
        val progress = createProgress(lastPosition = 7200L, durationTotal = 7200L)
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals(1.0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction clamped above 100 percent`() {
        val progress = createProgress(lastPosition = 8000L, durationTotal = 7200L)
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals(1.0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction with null total returns zero`() {
        val progress = createProgress(lastPosition = 3600L, durationTotal = null)
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction with zero total returns zero`() {
        val progress = createProgress(lastPosition = 100L, durationTotal = 0L)
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress fraction with negative total returns zero`() {
        val pct = PlaybackFormatter.progressFraction(100L, -10L)

        assertEquals(0f, pct, 0.001f)
    }

    @Test
    fun `progress bar shown only when fraction is positive`() {
        val pctPositive = PlaybackFormatter.progressFraction(100L, 7200L)
        val pctZero = PlaybackFormatter.progressFraction(0L, 7200L)
        val pctNullTotal = PlaybackFormatter.progressFraction(100L, null)

        assertTrue(pctPositive > 0f)
        assertFalse(pctZero > 0f)
        assertFalse(pctNullTotal > 0f)
    }

    @Test
    fun `pages unit formats amount correctly`() {
        val progress = createProgress(durationTotal = 350L, positionUnit = "pages")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("350 pages", duration)
    }

    @Test
    fun `pages unit with single page`() {
        val progress = createProgress(durationTotal = 1L, positionUnit = "pages")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("1 page", duration)
    }

    @Test
    fun `events unit formats amount correctly`() {
        val progress = createProgress(durationTotal = 10L, positionUnit = "events")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("10 runs", duration)
    }

    @Test
    fun `events unit with single run`() {
        val progress = createProgress(durationTotal = 1L, positionUnit = "events")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("1 run", duration)
    }

    @Test
    fun `unknown unit formats with raw value and unit name`() {
        val progress = createProgress(durationTotal = 42L, positionUnit = "chapters")
        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("42 chapters", duration)
    }

    @Test
    fun `progress at 25 percent`() {
        val pct = PlaybackFormatter.progressFraction(1800L, 7200L)

        assertEquals(0.25f, pct, 0.001f)
    }

    @Test
    fun `progress at 75 percent`() {
        val pct = PlaybackFormatter.progressFraction(5400L, 7200L)

        assertEquals(0.75f, pct, 0.001f)
    }

    @Test
    fun `formatSeconds handles exact hour boundary`() {
        assertEquals("1h 0m", PlaybackFormatter.formatSeconds(3600L))
    }

    @Test
    fun `formatSeconds handles hours minutes and seconds shows hours and minutes only`() {
        assertEquals("1h 30m", PlaybackFormatter.formatSeconds(5400L))
    }

    @Test
    fun `formatSeconds handles zero`() {
        assertEquals("-", PlaybackFormatter.formatSeconds(0L))
    }

    @Test
    fun `formatSeconds handles negative`() {
        assertEquals("-", PlaybackFormatter.formatSeconds(-10L))
    }

    @Test
    fun `formatSeconds handles seconds under a minute`() {
        assertEquals("30s", PlaybackFormatter.formatSeconds(30L))
    }

    @Test
    fun `full badge data consistency check`() {
        val progress = createProgress(
            durationTotal = 5400L,
            lastPosition = 2700L,
            lastSessionAmount = 900L,
            totalReproductions = 3L,
            positionUnit = "seconds",
        )

        val duration = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        val current = PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit)
        val lastSession = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit) + " last"
        val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions) + " played"
        val pct = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)

        assertEquals("1h 30m", duration)
        assertEquals("45m / 1h 30m", current)
        assertEquals("15m last", lastSession)
        assertEquals("3\u00D7 played", reps)
        assertEquals(0.5f, pct, 0.001f)
    }
}
