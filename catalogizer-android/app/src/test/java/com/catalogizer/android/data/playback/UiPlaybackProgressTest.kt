package com.catalogizer.android.data.playback

import org.junit.Assert.*
import org.junit.Test

class UiPlaybackProgressTest {

    // --- UiPlaybackProgress ---

    @Test
    fun `UiPlaybackProgress stores all fields`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 42L,
            positionUnit = "seconds",
            durationTotal = 7200L,
            lastPosition = 3600L,
            lastSessionAmount = 1800L,
            totalReproductions = 5L,
            aggregateAmount = 9000L,
            lastSessionEndedAtMs = 1700000000000L
        )

        assertEquals(42L, progress.mediaItemId)
        assertEquals("seconds", progress.positionUnit)
        assertEquals(7200L, progress.durationTotal)
        assertEquals(3600L, progress.lastPosition)
        assertEquals(1800L, progress.lastSessionAmount)
        assertEquals(5L, progress.totalReproductions)
        assertEquals(9000L, progress.aggregateAmount)
        assertEquals(1700000000000L, progress.lastSessionEndedAtMs)
    }

    @Test
    fun `UiPlaybackProgress with null durationTotal`() {
        val progress = createProgress(durationTotal = null)
        assertNull(progress.durationTotal)
    }

    @Test
    fun `UiPlaybackProgress with null lastSessionEndedAtMs`() {
        val progress = createProgress(lastSessionEndedAtMs = null)
        assertNull(progress.lastSessionEndedAtMs)
    }

    @Test
    fun `UiPlaybackProgress with pages unit`() {
        val progress = createProgress(positionUnit = "pages")
        assertEquals("pages", progress.positionUnit)
    }

    @Test
    fun `UiPlaybackProgress with events unit`() {
        val progress = createProgress(positionUnit = "events")
        assertEquals("events", progress.positionUnit)
    }

    @Test
    fun `UiPlaybackProgress with zero values`() {
        val progress = createProgress(
            lastPosition = 0L,
            lastSessionAmount = 0L,
            totalReproductions = 0L,
            aggregateAmount = 0L
        )

        assertEquals(0L, progress.lastPosition)
        assertEquals(0L, progress.lastSessionAmount)
        assertEquals(0L, progress.totalReproductions)
        assertEquals(0L, progress.aggregateAmount)
    }

    @Test
    fun `UiPlaybackProgress equality`() {
        val p1 = createProgress(mediaItemId = 1L)
        val p2 = createProgress(mediaItemId = 1L)
        assertEquals(p1, p2)
    }

    @Test
    fun `UiPlaybackProgress inequality on different mediaItemId`() {
        val p1 = createProgress(mediaItemId = 1L)
        val p2 = createProgress(mediaItemId = 2L)
        assertNotEquals(p1, p2)
    }

    @Test
    fun `UiPlaybackProgress inequality on different lastPosition`() {
        val p1 = createProgress(lastPosition = 100L)
        val p2 = createProgress(lastPosition = 200L)
        assertNotEquals(p1, p2)
    }

    @Test
    fun `UiPlaybackProgress copy updates single field`() {
        val original = createProgress(lastPosition = 100L)
        val updated = original.copy(lastPosition = 200L)

        assertEquals(200L, updated.lastPosition)
        assertEquals(original.mediaItemId, updated.mediaItemId)
        assertEquals(original.positionUnit, updated.positionUnit)
        assertEquals(original.durationTotal, updated.durationTotal)
    }

    @Test
    fun `UiPlaybackProgress hashCode consistent with equality`() {
        val p1 = createProgress()
        val p2 = createProgress()
        assertEquals(p1.hashCode(), p2.hashCode())
    }

    @Test
    fun `UiPlaybackProgress can compute progress fraction via PlaybackFormatter`() {
        val progress = createProgress(lastPosition = 1800L, durationTotal = 3600L)
        val fraction = PlaybackFormatter.progressFraction(progress.lastPosition, progress.durationTotal)
        assertEquals(0.5f, fraction, 0.001f)
    }

    @Test
    fun `UiPlaybackProgress can format position via PlaybackFormatter`() {
        val progress = createProgress(lastPosition = 5400L, positionUnit = "seconds")
        val formatted = PlaybackFormatter.formatAmount(progress.lastPosition, progress.positionUnit)
        assertEquals("1h 30m", formatted)
    }

    @Test
    fun `UiPlaybackProgress can format reproduction count via PlaybackFormatter`() {
        val progress = createProgress(totalReproductions = 7L)
        val formatted = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)
        assertEquals("7\u00D7", formatted)
    }

    // --- UiPlaybackSession ---

    @Test
    fun `UiPlaybackSession stores all fields`() {
        val session = UiPlaybackSession(
            id = 1L,
            positionUnit = "seconds",
            startPosition = 0L,
            endPosition = 1800L,
            totalAmount = 1800L,
            startedAtMs = 1700000000000L,
            endedAtMs = 1700001800000L,
            completed = false
        )

        assertEquals(1L, session.id)
        assertEquals("seconds", session.positionUnit)
        assertEquals(0L, session.startPosition)
        assertEquals(1800L, session.endPosition)
        assertEquals(1800L, session.totalAmount)
        assertEquals(1700000000000L, session.startedAtMs)
        assertEquals(1700001800000L, session.endedAtMs)
        assertFalse(session.completed)
    }

    @Test
    fun `UiPlaybackSession completed session`() {
        val session = createSession(completed = true)
        assertTrue(session.completed)
    }

    @Test
    fun `UiPlaybackSession with null endedAtMs represents active session`() {
        val session = createSession(endedAtMs = null)
        assertNull(session.endedAtMs)
    }

    @Test
    fun `UiPlaybackSession equality`() {
        val s1 = createSession(id = 1L)
        val s2 = createSession(id = 1L)
        assertEquals(s1, s2)
    }

    @Test
    fun `UiPlaybackSession inequality on different id`() {
        val s1 = createSession(id = 1L)
        val s2 = createSession(id = 2L)
        assertNotEquals(s1, s2)
    }

    @Test
    fun `UiPlaybackSession copy updates completed flag`() {
        val original = createSession(completed = false)
        val updated = original.copy(completed = true, endedAtMs = 1700001800000L)

        assertTrue(updated.completed)
        assertEquals(1700001800000L, updated.endedAtMs)
        assertEquals(original.id, updated.id)
        assertEquals(original.startPosition, updated.startPosition)
    }

    @Test
    fun `UiPlaybackSession hashCode consistent with equality`() {
        val s1 = createSession()
        val s2 = createSession()
        assertEquals(s1.hashCode(), s2.hashCode())
    }

    @Test
    fun `UiPlaybackSession duration can be computed`() {
        val session = createSession(startPosition = 100L, endPosition = 500L)
        val duration = session.endPosition - session.startPosition
        assertEquals(400L, duration)
    }

    @Test
    fun `UiPlaybackSession with pages position unit`() {
        val session = createSession(positionUnit = "pages", startPosition = 1L, endPosition = 50L)
        assertEquals("pages", session.positionUnit)
        assertEquals(49L, session.endPosition - session.startPosition)
    }

    // --- Helpers ---

    private fun createProgress(
        mediaItemId: Long = 1L,
        positionUnit: String = "seconds",
        durationTotal: Long? = 3600L,
        lastPosition: Long = 1800L,
        lastSessionAmount: Long = 900L,
        totalReproductions: Long = 3L,
        aggregateAmount: Long = 5400L,
        lastSessionEndedAtMs: Long? = 1700000000000L
    ): UiPlaybackProgress = UiPlaybackProgress(
        mediaItemId = mediaItemId,
        positionUnit = positionUnit,
        durationTotal = durationTotal,
        lastPosition = lastPosition,
        lastSessionAmount = lastSessionAmount,
        totalReproductions = totalReproductions,
        aggregateAmount = aggregateAmount,
        lastSessionEndedAtMs = lastSessionEndedAtMs
    )

    private fun createSession(
        id: Long = 1L,
        positionUnit: String = "seconds",
        startPosition: Long = 0L,
        endPosition: Long = 1800L,
        totalAmount: Long = 1800L,
        startedAtMs: Long = 1700000000000L,
        endedAtMs: Long? = 1700001800000L,
        completed: Boolean = false
    ): UiPlaybackSession = UiPlaybackSession(
        id = id,
        positionUnit = positionUnit,
        startPosition = startPosition,
        endPosition = endPosition,
        totalAmount = totalAmount,
        startedAtMs = startedAtMs,
        endedAtMs = endedAtMs,
        completed = completed
    )
}
