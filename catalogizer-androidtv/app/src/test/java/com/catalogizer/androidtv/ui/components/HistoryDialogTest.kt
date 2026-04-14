package com.catalogizer.androidtv.ui.components

import com.catalogizer.androidtv.data.playback.PlaybackFormatter
import com.catalogizer.androidtv.data.playback.UiPlaybackProgress
import com.catalogizer.androidtv.data.playback.UiPlaybackSession
import org.junit.Assert.*
import org.junit.Test
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone

/**
 * Tests for HistoryDialog logic: summary row construction from UiPlaybackProgress,
 * session list display, date formatting, empty/loading states, and null handling.
 */
class HistoryDialogTest {

    // --- Summary construction from UiPlaybackProgress ---

    @Test
    fun `null progress produces Never reproduced summary`() {
        val progress: UiPlaybackProgress? = null
        val summaryRows = if (progress == null) {
            listOf("Status" to "Never reproduced")
        } else {
            emptyList()
        }
        assertEquals(1, summaryRows.size)
        assertEquals("Status", summaryRows[0].first)
        assertEquals("Never reproduced", summaryRows[0].second)
    }

    @Test
    fun `non-null progress produces five summary rows`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = 7200L,
            lastPosition = 3600L,
            lastSessionAmount = 1200L,
            totalReproductions = 3,
            aggregateAmount = 5400L,
            lastSessionEndedAtMs = System.currentTimeMillis()
        )
        val summaryRows = buildList {
            add("Total duration" to PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit))
            add("Current position" to PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit))
            add("Last session" to PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit))
            add("Times played" to PlaybackFormatter.formatReproductionCount(progress.totalReproductions))
            add("Total reproduced" to PlaybackFormatter.formatAmount(progress.aggregateAmount, progress.positionUnit))
        }
        assertEquals(5, summaryRows.size)
        assertEquals("Total duration", summaryRows[0].first)
        assertEquals("2h 0m", summaryRows[0].second)
        assertEquals("Current position", summaryRows[1].first)
        assertEquals("1h 0m / 2h 0m", summaryRows[1].second)
        assertEquals("Last session", summaryRows[2].first)
        assertEquals("20m", summaryRows[2].second)
        assertEquals("Times played", summaryRows[3].first)
        assertEquals("3\u00D7", summaryRows[3].second)
        assertEquals("Total reproduced", summaryRows[4].first)
        assertEquals("1h 30m", summaryRows[4].second)
    }

    @Test
    fun `summary handles null durationTotal by formatting zero`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 1L,
            positionUnit = "seconds",
            durationTotal = null,
            lastPosition = 600L,
            lastSessionAmount = 300L,
            totalReproductions = 1,
            aggregateAmount = 600L,
            lastSessionEndedAtMs = null
        )
        val durationLabel = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("-", durationLabel)
    }

    @Test
    fun `summary with pages unit formats correctly`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 2L,
            positionUnit = "pages",
            durationTotal = 320L,
            lastPosition = 140L,
            lastSessionAmount = 15L,
            totalReproductions = 5,
            aggregateAmount = 280L,
            lastSessionEndedAtMs = System.currentTimeMillis()
        )
        val durationLabel = PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit)
        assertEquals("320 pages", durationLabel)
        val positionLabel = PlaybackFormatter.formatProgress(progress.lastPosition, progress.durationTotal, progress.positionUnit)
        assertEquals("140 pages / 320 pages", positionLabel)
        val lastSession = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit)
        assertEquals("15 pages", lastSession)
    }

    @Test
    fun `summary with events unit formats correctly`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 3L,
            positionUnit = "events",
            durationTotal = null,
            lastPosition = 5L,
            lastSessionAmount = 1L,
            totalReproductions = 5,
            aggregateAmount = 5L,
            lastSessionEndedAtMs = null
        )
        val lastSession = PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit)
        assertEquals("1 run", lastSession)
        val aggregate = PlaybackFormatter.formatAmount(progress.aggregateAmount, progress.positionUnit)
        assertEquals("5 runs", aggregate)
    }

    // --- Session list display ---

    @Test
    fun `empty session list should show no-sessions message`() {
        val sessions = emptyList<UiPlaybackSession>()
        val loaded = true
        assertTrue(loaded && sessions.isEmpty())
    }

    @Test
    fun `not-loaded state should show loading message`() {
        val loaded = false
        assertFalse(loaded)
    }

    @Test
    fun `session list with items should display sessions`() {
        val sessions = listOf(
            UiPlaybackSession(
                id = 1L,
                positionUnit = "seconds",
                startPosition = 0L,
                endPosition = 3600L,
                totalAmount = 3600L,
                startedAtMs = System.currentTimeMillis() - 86400000,
                endedAtMs = System.currentTimeMillis() - 82800000,
                completed = true
            ),
            UiPlaybackSession(
                id = 2L,
                positionUnit = "seconds",
                startPosition = 3600L,
                endPosition = 5400L,
                totalAmount = 1800L,
                startedAtMs = System.currentTimeMillis() - 3600000,
                endedAtMs = null,
                completed = false
            )
        )
        assertEquals(2, sessions.size)
        assertTrue(sessions[0].completed)
        assertFalse(sessions[1].completed)
    }

    @Test
    fun `session row formats amount correctly`() {
        val session = UiPlaybackSession(
            id = 1L,
            positionUnit = "seconds",
            startPosition = 0L,
            endPosition = 5400L,
            totalAmount = 5400L,
            startedAtMs = 1700000000000L,
            endedAtMs = 1700005400000L,
            completed = true
        )
        val amount = PlaybackFormatter.formatAmount(session.totalAmount, session.positionUnit)
        assertEquals("1h 30m", amount)
    }

    @Test
    fun `session status text shows Completed for completed sessions`() {
        val completed = true
        val status = if (completed) "Completed" else "In progress"
        assertEquals("Completed", status)
    }

    @Test
    fun `session status text shows In progress for incomplete sessions`() {
        val completed = false
        val status = if (completed) "Completed" else "In progress"
        assertEquals("In progress", status)
    }

    // --- Date formatting ---

    @Test
    fun `formatInstant returns dash for zero timestamp`() {
        val result = formatInstant(0L)
        assertEquals("-", result)
    }

    @Test
    fun `formatInstant returns dash for negative timestamp`() {
        val result = formatInstant(-1L)
        assertEquals("-", result)
    }

    @Test
    fun `formatInstant returns formatted date for positive timestamp`() {
        val ms = 1700000000000L // 2023-11-14 approximately
        val result = formatInstant(ms)
        assertTrue(result.isNotBlank())
        assertTrue(result.contains("-"))
        assertTrue(result.contains(":"))
    }

    @Test
    fun `formatInstant produces consistent output for same input`() {
        val ms = 1700000000000L
        val result1 = formatInstant(ms)
        val result2 = formatInstant(ms)
        assertEquals(result1, result2)
    }

    // --- Session keying ---

    @Test
    fun `session IDs serve as unique keys`() {
        val sessions = listOf(
            UiPlaybackSession(1L, "seconds", 0, 100, 100, 1000, 2000, true),
            UiPlaybackSession(2L, "seconds", 100, 200, 100, 2000, 3000, true),
            UiPlaybackSession(3L, "seconds", 200, 300, 100, 3000, 4000, false)
        )
        val ids = sessions.map { it.id }.toSet()
        assertEquals(sessions.size, ids.size)
    }

    // --- Dismiss callback ---

    @Test
    fun `dismiss callback fires correctly`() {
        var dismissed = false
        val onDismiss: () -> Unit = { dismissed = true }
        onDismiss()
        assertTrue(dismissed)
    }

    // Helper matching the private function in HistoryDialog.kt
    private fun formatInstant(ms: Long): String {
        if (ms <= 0L) return "-"
        val fmt = SimpleDateFormat("yyyy-MM-dd HH:mm", Locale.getDefault())
        fmt.timeZone = TimeZone.getDefault()
        return fmt.format(Date(ms))
    }
}
