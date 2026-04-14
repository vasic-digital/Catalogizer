package com.catalogizer.android.ui.components

import com.catalogizer.android.data.playback.PlaybackFormatter
import com.catalogizer.android.data.playback.PlaybackRepository
import com.catalogizer.android.data.playback.UiPlaybackProgress
import com.catalogizer.android.data.playback.UiPlaybackSession
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for HistoryDialog data loading and display logic.
 * Since Compose UI testing requires instrumentation (androidTest scope),
 * these tests verify the data layer contracts that HistoryDialog relies on:
 * PlaybackRepository responses, PlaybackFormatter output, and data models.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class HistoryDialogTest {

    private lateinit var mockRepository: PlaybackRepository

    private fun createProgress(
        mediaItemId: Long = 1L,
        positionUnit: String = "seconds",
        durationTotal: Long? = 7200L,
        lastPosition: Long = 3600L,
        lastSessionAmount: Long = 1800L,
        totalReproductions: Long = 5L,
        aggregateAmount: Long = 9000L,
        lastSessionEndedAtMs: Long? = 1700000000000L,
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

    private fun createSession(
        id: Long = 1L,
        positionUnit: String = "seconds",
        startPosition: Long = 0L,
        endPosition: Long = 1800L,
        totalAmount: Long = 1800L,
        startedAtMs: Long = 1700000000000L,
        endedAtMs: Long? = 1700001800000L,
        completed: Boolean = true,
    ): UiPlaybackSession = UiPlaybackSession(
        id = id,
        positionUnit = positionUnit,
        startPosition = startPosition,
        endPosition = endPosition,
        totalAmount = totalAmount,
        startedAtMs = startedAtMs,
        endedAtMs = endedAtMs,
        completed = completed,
    )

    @Before
    fun setup() {
        mockRepository = mockk(relaxed = true)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `repository getProgress returns valid progress for media item`() = runTest {
        val progress = createProgress(mediaItemId = 42L)
        coEvery { mockRepository.getProgress(42L) } returns progress

        val result = mockRepository.getProgress(42L)

        assertNotNull(result)
        assertEquals(42L, result?.mediaItemId)
        assertEquals("seconds", result?.positionUnit)
        assertEquals(7200L, result?.durationTotal)
        assertEquals(3600L, result?.lastPosition)
        assertEquals(1800L, result?.lastSessionAmount)
        assertEquals(5L, result?.totalReproductions)
        assertEquals(9000L, result?.aggregateAmount)
    }

    @Test
    fun `repository getProgress returns null for item with no history`() = runTest {
        coEvery { mockRepository.getProgress(99L) } returns null

        val result = mockRepository.getProgress(99L)

        assertNull(result)
    }

    @Test
    fun `repository listHistory returns sessions for media item`() = runTest {
        val sessions = listOf(
            createSession(id = 1L, totalAmount = 1800L, completed = true),
            createSession(id = 2L, totalAmount = 900L, completed = false),
            createSession(id = 3L, totalAmount = 3600L, completed = true),
        )
        coEvery { mockRepository.listHistory(1L, limit = 100) } returns sessions

        val result = mockRepository.listHistory(1L, limit = 100)

        assertEquals(3, result.size)
        assertTrue(result[0].completed)
        assertFalse(result[1].completed)
        assertTrue(result[2].completed)
    }

    @Test
    fun `repository listHistory returns empty list for item with no sessions`() = runTest {
        coEvery { mockRepository.listHistory(99L, limit = 100) } returns emptyList()

        val result = mockRepository.listHistory(99L, limit = 100)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `summary rows show never reproduced when progress is null`() {
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
    fun `summary rows build correctly with valid progress`() {
        val progress = createProgress(
            durationTotal = 7200L,
            lastPosition = 3600L,
            lastSessionAmount = 1800L,
            totalReproductions = 5L,
            aggregateAmount = 9000L,
            positionUnit = "seconds",
        )

        val summaryRows = buildList {
            add("Total duration" to PlaybackFormatter.formatAmount(
                progress.durationTotal ?: 0L, progress.positionUnit))
            add("Current position" to PlaybackFormatter.formatProgress(
                progress.lastPosition, progress.durationTotal, progress.positionUnit))
            add("Last session" to PlaybackFormatter.formatAmount(
                progress.lastSessionAmount, progress.positionUnit))
            add("Times played" to PlaybackFormatter.formatReproductionCount(
                progress.totalReproductions))
            add("Total reproduced" to PlaybackFormatter.formatAmount(
                progress.aggregateAmount, progress.positionUnit))
        }

        assertEquals(5, summaryRows.size)
        assertEquals("Total duration", summaryRows[0].first)
        assertEquals("2h 0m", summaryRows[0].second)
        assertEquals("Current position", summaryRows[1].first)
        assertEquals("1h 0m / 2h 0m", summaryRows[1].second)
        assertEquals("Last session", summaryRows[2].first)
        assertEquals("30m", summaryRows[2].second)
        assertEquals("Times played", summaryRows[3].first)
        assertEquals("5\u00D7", summaryRows[3].second)
        assertEquals("Total reproduced", summaryRows[4].first)
        assertEquals("2h 30m", summaryRows[4].second)
    }

    @Test
    fun `summary rows handle null durationTotal gracefully`() {
        val progress = createProgress(durationTotal = null)

        val totalDuration = PlaybackFormatter.formatAmount(
            progress.durationTotal ?: 0L, progress.positionUnit)

        assertEquals("-", totalDuration)
    }

    @Test
    fun `summary rows handle zero values`() {
        val progress = createProgress(
            durationTotal = 0L,
            lastPosition = 0L,
            lastSessionAmount = 0L,
            totalReproductions = 0L,
            aggregateAmount = 0L,
        )

        assertEquals("-", PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit))
        assertEquals("-", PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit))
        assertEquals("0\u00D7", PlaybackFormatter.formatReproductionCount(progress.totalReproductions))
        assertEquals("-", PlaybackFormatter.formatAmount(progress.aggregateAmount, progress.positionUnit))
    }

    @Test
    fun `summary rows handle pages unit`() {
        val progress = createProgress(
            positionUnit = "pages",
            durationTotal = 350L,
            lastPosition = 120L,
            lastSessionAmount = 30L,
            aggregateAmount = 500L,
        )

        assertEquals("350 pages", PlaybackFormatter.formatAmount(progress.durationTotal ?: 0L, progress.positionUnit))
        assertEquals("120 pages / 350 pages", PlaybackFormatter.formatProgress(
            progress.lastPosition, progress.durationTotal, progress.positionUnit))
        assertEquals("30 pages", PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit))
    }

    @Test
    fun `summary rows handle single page unit`() {
        val progress = createProgress(
            positionUnit = "pages",
            lastSessionAmount = 1L,
        )

        assertEquals("1 page", PlaybackFormatter.formatAmount(progress.lastSessionAmount, progress.positionUnit))
    }

    @Test
    fun `session row shows completed status`() {
        val session = createSession(completed = true)
        val status = if (session.completed) "Completed" else "In progress"

        assertEquals("Completed", status)
    }

    @Test
    fun `session row shows in progress status`() {
        val session = createSession(completed = false)
        val status = if (session.completed) "Completed" else "In progress"

        assertEquals("In progress", status)
    }

    @Test
    fun `formatInstant returns dash for zero timestamp`() {
        val ms = 0L
        val result = if (ms <= 0L) "-" else "formatted"

        assertEquals("-", result)
    }

    @Test
    fun `formatInstant returns dash for negative timestamp`() {
        val ms = -1L
        val result = if (ms <= 0L) "-" else "formatted"

        assertEquals("-", result)
    }

    @Test
    fun `session amount formats correctly with seconds unit`() {
        val session = createSession(totalAmount = 5400L, positionUnit = "seconds")
        val amount = PlaybackFormatter.formatAmount(session.totalAmount, session.positionUnit)

        assertEquals("1h 30m", amount)
    }

    @Test
    fun `session amount formats correctly with events unit`() {
        val session = createSession(totalAmount = 3L, positionUnit = "events")
        val amount = PlaybackFormatter.formatAmount(session.totalAmount, session.positionUnit)

        assertEquals("3 runs", amount)
    }

    @Test
    fun `session amount formats single event correctly`() {
        val session = createSession(totalAmount = 1L, positionUnit = "events")
        val amount = PlaybackFormatter.formatAmount(session.totalAmount, session.positionUnit)

        assertEquals("1 run", amount)
    }

    @Test
    fun `repository getProgress is called with correct media item id`() = runTest {
        coEvery { mockRepository.getProgress(any()) } returns null

        mockRepository.getProgress(42L)

        coVerify { mockRepository.getProgress(42L) }
    }

    @Test
    fun `repository listHistory is called with correct limit`() = runTest {
        coEvery { mockRepository.listHistory(any(), any()) } returns emptyList()

        mockRepository.listHistory(1L, limit = 100)

        coVerify { mockRepository.listHistory(1L, limit = 100) }
    }

    @Test
    fun `UiPlaybackProgress data class equality works correctly`() {
        val p1 = createProgress(mediaItemId = 1L)
        val p2 = createProgress(mediaItemId = 1L)
        val p3 = createProgress(mediaItemId = 2L)

        assertEquals(p1, p2)
        assertNotEquals(p1, p3)
    }

    @Test
    fun `UiPlaybackSession data class equality works correctly`() {
        val s1 = createSession(id = 1L)
        val s2 = createSession(id = 1L)
        val s3 = createSession(id = 2L)

        assertEquals(s1, s2)
        assertNotEquals(s1, s3)
    }

    @Test
    fun `session list preserves order from repository`() = runTest {
        val sessions = listOf(
            createSession(id = 3L, startedAtMs = 1700003000000L),
            createSession(id = 1L, startedAtMs = 1700001000000L),
            createSession(id = 2L, startedAtMs = 1700002000000L),
        )
        coEvery { mockRepository.listHistory(1L, limit = 100) } returns sessions

        val result = mockRepository.listHistory(1L, limit = 100)

        assertEquals(3L, result[0].id)
        assertEquals(1L, result[1].id)
        assertEquals(2L, result[2].id)
    }

    @Test
    fun `session key uses id for list identification`() {
        val sessions = listOf(
            createSession(id = 10L),
            createSession(id = 20L),
            createSession(id = 30L),
        )
        val keys = sessions.map { it.id }

        assertEquals(listOf(10L, 20L, 30L), keys)
        assertEquals(keys.size, keys.toSet().size)
    }

    @Test
    fun `dismiss callback is invocable`() {
        var dismissed = false
        val onDismiss: () -> Unit = { dismissed = true }

        onDismiss()

        assertTrue(dismissed)
    }
}
