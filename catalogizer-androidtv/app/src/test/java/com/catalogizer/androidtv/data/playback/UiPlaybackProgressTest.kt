package com.catalogizer.androidtv.data.playback

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class UiPlaybackProgressTest {

    // ------------------------------------------------------------------
    // UiPlaybackProgress construction and field access
    // ------------------------------------------------------------------

    @Test
    fun `default UiPlaybackProgress holds supplied values`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 42L,
            positionUnit = "seconds",
            durationTotal = 7200L,
            lastPosition = 3600L,
            lastSessionAmount = 1800L,
            totalReproductions = 3L,
            aggregateAmount = 5400L,
            lastSessionEndedAtMs = 1700000000000L
        )
        assertEquals(42L, progress.mediaItemId)
        assertEquals("seconds", progress.positionUnit)
        assertEquals(7200L, progress.durationTotal)
        assertEquals(3600L, progress.lastPosition)
        assertEquals(1800L, progress.lastSessionAmount)
        assertEquals(3L, progress.totalReproductions)
        assertEquals(5400L, progress.aggregateAmount)
        assertEquals(1700000000000L, progress.lastSessionEndedAtMs)
    }

    @Test
    fun `durationTotal can be null for unknown duration`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "pages",
            durationTotal = null,
            lastPosition = 50L,
            lastSessionAmount = 10L,
            totalReproductions = 1L,
            aggregateAmount = 50L,
            lastSessionEndedAtMs = null
        )
        assertNull(progress.durationTotal)
        assertNull(progress.lastSessionEndedAtMs)
    }

    @Test
    fun `positionUnit pages is valid`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "pages",
            durationTotal = 320L,
            lastPosition = 140L,
            lastSessionAmount = 20L,
            totalReproductions = 5L,
            aggregateAmount = 140L,
            lastSessionEndedAtMs = null
        )
        assertEquals("pages", progress.positionUnit)
    }

    @Test
    fun `positionUnit events is valid`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "events",
            durationTotal = null,
            lastPosition = 0L,
            lastSessionAmount = 1L,
            totalReproductions = 7L,
            aggregateAmount = 7L,
            lastSessionEndedAtMs = null
        )
        assertEquals("events", progress.positionUnit)
    }

    @Test
    fun `zero position and amounts are valid`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 99L,
            positionUnit = "seconds",
            durationTotal = 0L,
            lastPosition = 0L,
            lastSessionAmount = 0L,
            totalReproductions = 0L,
            aggregateAmount = 0L,
            lastSessionEndedAtMs = null
        )
        assertEquals(0L, progress.lastPosition)
        assertEquals(0L, progress.totalReproductions)
    }

    @Test
    fun `data class equality works for identical values`() {
        val a = UiPlaybackProgress(1L, "seconds", 100L, 50L, 25L, 2L, 75L, 1000L)
        val b = UiPlaybackProgress(1L, "seconds", 100L, 50L, 25L, 2L, 75L, 1000L)
        assertEquals(a, b)
        assertEquals(a.hashCode(), b.hashCode())
    }

    @Test
    fun `data class inequality when fields differ`() {
        val a = UiPlaybackProgress(1L, "seconds", 100L, 50L, 25L, 2L, 75L, 1000L)
        val b = UiPlaybackProgress(2L, "seconds", 100L, 50L, 25L, 2L, 75L, 1000L)
        assertFalse(a == b)
    }

    @Test
    fun `copy changes only specified fields`() {
        val original = UiPlaybackProgress(1L, "seconds", 100L, 50L, 25L, 2L, 75L, 1000L)
        val copy = original.copy(lastPosition = 75L, totalReproductions = 3L)
        assertEquals(75L, copy.lastPosition)
        assertEquals(3L, copy.totalReproductions)
        assertEquals(original.mediaItemId, copy.mediaItemId)
        assertEquals(original.positionUnit, copy.positionUnit)
    }

    @Test
    fun `large values for epoch milliseconds`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = Long.MAX_VALUE,
            lastPosition = Long.MAX_VALUE - 1,
            lastSessionAmount = Long.MAX_VALUE / 2,
            totalReproductions = 999_999L,
            aggregateAmount = Long.MAX_VALUE,
            lastSessionEndedAtMs = Long.MAX_VALUE
        )
        assertEquals(Long.MAX_VALUE, progress.durationTotal)
        assertEquals(Long.MAX_VALUE, progress.lastSessionEndedAtMs)
    }

    // ------------------------------------------------------------------
    // UiPlaybackSession construction and field access
    // ------------------------------------------------------------------

    @Test
    fun `UiPlaybackSession holds supplied values`() {
        val session = UiPlaybackSession(
            id = 10L,
            positionUnit = "seconds",
            startPosition = 0L,
            endPosition = 3600L,
            totalAmount = 3600L,
            startedAtMs = 1700000000000L,
            endedAtMs = 1700003600000L,
            completed = true
        )
        assertEquals(10L, session.id)
        assertEquals("seconds", session.positionUnit)
        assertEquals(0L, session.startPosition)
        assertEquals(3600L, session.endPosition)
        assertEquals(3600L, session.totalAmount)
        assertEquals(1700000000000L, session.startedAtMs)
        assertEquals(1700003600000L, session.endedAtMs)
        assertTrue(session.completed)
    }

    @Test
    fun `UiPlaybackSession endedAtMs can be null for ongoing session`() {
        val session = UiPlaybackSession(
            id = 11L,
            positionUnit = "pages",
            startPosition = 0L,
            endPosition = 100L,
            totalAmount = 100L,
            startedAtMs = 1700000000000L,
            endedAtMs = null,
            completed = false
        )
        assertNull(session.endedAtMs)
        assertFalse(session.completed)
    }

    @Test
    fun `UiPlaybackSession data class equality`() {
        val a = UiPlaybackSession(1L, "seconds", 0L, 100L, 100L, 1000L, 2000L, false)
        val b = UiPlaybackSession(1L, "seconds", 0L, 100L, 100L, 1000L, 2000L, false)
        assertEquals(a, b)
        assertEquals(a.hashCode(), b.hashCode())
    }

    @Test
    fun `UiPlaybackSession copy changes only specified fields`() {
        val original = UiPlaybackSession(1L, "seconds", 0L, 100L, 100L, 1000L, null, false)
        val copy = original.copy(endPosition = 200L, completed = true, endedAtMs = 3000L)
        assertEquals(200L, copy.endPosition)
        assertTrue(copy.completed)
        assertEquals(3000L, copy.endedAtMs)
        assertEquals(original.id, copy.id)
    }

    @Test
    fun `UiPlaybackSession with zero positions`() {
        val session = UiPlaybackSession(
            id = 0L,
            positionUnit = "events",
            startPosition = 0L,
            endPosition = 0L,
            totalAmount = 0L,
            startedAtMs = 0L,
            endedAtMs = 0L,
            completed = false
        )
        assertEquals(0L, session.id)
        assertEquals(0L, session.startPosition)
        assertEquals(0L, session.endPosition)
    }

    // ------------------------------------------------------------------
    // Percentage / fraction computation (via PlaybackFormatter)
    // ------------------------------------------------------------------

    @Test
    fun `progress fraction computes correctly from UiPlaybackProgress fields`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = 7200L,
            lastPosition = 3600L,
            lastSessionAmount = 1800L,
            totalReproductions = 1L,
            aggregateAmount = 3600L,
            lastSessionEndedAtMs = null
        )
        val fraction = PlaybackFormatter.progressFraction(
            progress.lastPosition,
            progress.durationTotal
        )
        assertEquals(0.5f, fraction, 0.001f)
    }

    @Test
    fun `progress fraction is zero when duration is null`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = null,
            lastPosition = 100L,
            lastSessionAmount = 100L,
            totalReproductions = 1L,
            aggregateAmount = 100L,
            lastSessionEndedAtMs = null
        )
        val fraction = PlaybackFormatter.progressFraction(
            progress.lastPosition,
            progress.durationTotal
        )
        assertEquals(0f, fraction, 0.001f)
    }

    @Test
    fun `progress fraction clamps to 1 when position exceeds duration`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = 100L,
            lastPosition = 200L,
            lastSessionAmount = 200L,
            totalReproductions = 1L,
            aggregateAmount = 200L,
            lastSessionEndedAtMs = null
        )
        val fraction = PlaybackFormatter.progressFraction(
            progress.lastPosition,
            progress.durationTotal
        )
        assertEquals(1f, fraction, 0.001f)
    }
}
