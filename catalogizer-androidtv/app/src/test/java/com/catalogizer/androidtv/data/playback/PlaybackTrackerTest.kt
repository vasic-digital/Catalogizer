package com.catalogizer.androidtv.data.playback

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import io.mockk.clearAllMocks
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class PlaybackTrackerTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var api: CatalogizerApi
    private lateinit var testScope: TestScope
    private lateinit var tracker: PlaybackTracker

    @Before
    fun setup() {
        api = mockk(relaxed = true)
        testScope = TestScope()
        tracker = PlaybackTracker(api, testScope)
    }

    @After
    fun tearDown() {
        tracker.stopProgressTicker()
        clearAllMocks()
    }

    // ------------------------------------------------------------------
    // start — success cases
    // ------------------------------------------------------------------

    @Test
    fun `start returns session id on success`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(123L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)

        val id = tracker.start(mediaItemId = 1L, fileId = 10L)

        assertEquals(123L, id)
        assertTrue(tracker.isActive())
    }

    @Test
    fun `start sends correct body with all parameters`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(1L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)

        tracker.start(
            mediaItemId = 42L,
            fileId = 7L,
            positionUnit = "pages",
            startPosition = 100L
        )

        coVerify {
            api.startPlaybackSession(match { body ->
                body["media_item_id"] == 42L &&
                body["file_id"] == 7L &&
                body["position_unit"] == "pages" &&
                body["start_position"] == 100L
            })
        }
    }

    @Test
    fun `start omits file_id when null`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(1L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)

        tracker.start(mediaItemId = 42L, fileId = null)

        coVerify {
            api.startPlaybackSession(match { body ->
                !body.containsKey("file_id")
            })
        }
    }

    @Test
    fun `start uses default positionUnit and startPosition`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(1L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)

        tracker.start(mediaItemId = 1L, fileId = null)

        coVerify {
            api.startPlaybackSession(match { body ->
                body["position_unit"] == "seconds" &&
                body["start_position"] == 0L
            })
        }
    }

    // ------------------------------------------------------------------
    // start — failure cases
    // ------------------------------------------------------------------

    @Test
    fun `start returns 0 on HTTP error`() = runTest {
        coEvery { api.startPlaybackSession(any()) } returns Response.error(
            500, "Server error".toResponseBody(null)
        )

        val id = tracker.start(mediaItemId = 1L, fileId = null)

        assertEquals(0L, id)
        assertFalse(tracker.isActive())
    }

    @Test
    fun `start returns 0 on network exception`() = runTest {
        coEvery { api.startPlaybackSession(any()) } throws RuntimeException("No route to host")

        val id = tracker.start(mediaItemId = 1L, fileId = null)

        assertEquals(0L, id)
        assertFalse(tracker.isActive())
    }

    @Test
    fun `start returns 0 when response body is null`() = runTest {
        coEvery { api.startPlaybackSession(any()) } returns Response.success(null)

        val id = tracker.start(mediaItemId = 1L, fileId = null)

        assertEquals(0L, id)
    }

    @Test
    fun `start returns 0 when session_id key is missing`() = runTest {
        val responseBody = buildJsonObject {
            put("other_key", JsonPrimitive("value"))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)

        val id = tracker.start(mediaItemId = 1L, fileId = null)

        assertEquals(0L, id)
    }

    // ------------------------------------------------------------------
    // isActive
    // ------------------------------------------------------------------

    @Test
    fun `isActive is false before start`() {
        assertFalse(tracker.isActive())
    }

    @Test
    fun `isActive is true after successful start`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(5L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)

        tracker.start(mediaItemId = 1L, fileId = null)

        assertTrue(tracker.isActive())
    }

    @Test
    fun `isActive is false after failed start`() = runTest {
        coEvery { api.startPlaybackSession(any()) } throws RuntimeException("fail")

        tracker.start(mediaItemId = 1L, fileId = null)

        assertFalse(tracker.isActive())
    }

    // ------------------------------------------------------------------
    // end — success cases
    // ------------------------------------------------------------------

    @Test
    fun `end sends correct body and clears session`() = runTest {
        // Start a session first
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.endPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        assertTrue(tracker.isActive())

        tracker.end(endPosition = 3600L, totalAmount = 3600L, completed = true)

        assertFalse(tracker.isActive())
        coVerify {
            api.endPlaybackSession(match { body ->
                body["session_id"] == 10L &&
                body["end_position"] == 3600L &&
                body["total_amount"] == 3600L &&
                body["completed"] == true
            })
        }
    }

    @Test
    fun `end is no-op when no session is active`() = runTest {
        assertFalse(tracker.isActive())

        tracker.end(endPosition = 100L, totalAmount = 100L, completed = false)

        coVerify(exactly = 0) { api.endPlaybackSession(any()) }
    }

    @Test
    fun `end can be called multiple times safely`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.endPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.end(endPosition = 100L, totalAmount = 100L, completed = true)
        tracker.end(endPosition = 200L, totalAmount = 200L, completed = false)

        // Second call is a no-op because sessionId was zeroed
        coVerify(exactly = 1) { api.endPlaybackSession(any()) }
    }

    // ------------------------------------------------------------------
    // end — error cases
    // ------------------------------------------------------------------

    @Test
    fun `end clears session even when API call throws`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.endPlaybackSession(any()) } throws RuntimeException("Network error")

        tracker.start(mediaItemId = 1L, fileId = null)
        assertTrue(tracker.isActive())

        tracker.end(endPosition = 100L, totalAmount = 100L, completed = false)

        // Session is cleared even on exception (finally block)
        assertFalse(tracker.isActive())
    }

    @Test
    fun `end clears session on HTTP error`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.endPlaybackSession(any()) } returns Response.error(
            500, "error".toResponseBody(null)
        )

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.end(endPosition = 100L, totalAmount = 100L, completed = false)

        assertFalse(tracker.isActive())
    }

    // ------------------------------------------------------------------
    // end — stops progress ticker
    // ------------------------------------------------------------------

    @Test
    fun `end stops progress ticker`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.endPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 500L },
            getTotalAmount = { 500L }
        )

        tracker.end(endPosition = 500L, totalAmount = 500L, completed = true)

        // After end, advancing time should NOT produce progress calls
        testScope.advanceTimeBy(5000L)
        coVerify(exactly = 0) { api.progressPlaybackSession(any()) }
    }

    // ------------------------------------------------------------------
    // startProgressTicker
    // ------------------------------------------------------------------

    @Test
    fun `startProgressTicker sends progress after interval`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 500L },
            getTotalAmount = { 500L }
        )

        testScope.advanceTimeBy(1100L)

        coVerify(atLeast = 1) {
            api.progressPlaybackSession(match { body ->
                body["session_id"] == 10L &&
                body["end_position"] == 500L &&
                body["total_amount"] == 500L
            })
        }
    }

    @Test
    fun `startProgressTicker sends multiple ticks over time`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 500L },
            getTotalAmount = { 500L }
        )

        testScope.advanceTimeBy(3100L)

        coVerify(atLeast = 3) { api.progressPlaybackSession(any()) }
    }

    @Test
    fun `startProgressTicker reads live position from getter`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)

        var position = 100L
        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { position },
            getTotalAmount = { position }
        )

        testScope.advanceTimeBy(1100L)
        position = 200L
        testScope.advanceTimeBy(1100L)

        coVerify {
            api.progressPlaybackSession(match { it["end_position"] == 100L })
        }
        coVerify {
            api.progressPlaybackSession(match { it["end_position"] == 200L })
        }
    }

    @Test
    fun `startProgressTicker survives API exception and retries next tick`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        // First call throws, second succeeds
        coEvery { api.progressPlaybackSession(any()) } throws RuntimeException("Network")
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 100L },
            getTotalAmount = { 100L }
        )

        testScope.advanceTimeBy(2100L)

        // Should have attempted at least 2 ticks (first fails, second retries)
        coVerify(atLeast = 2) { api.progressPlaybackSession(any()) }
    }

    @Test
    fun `startProgressTicker replaces previous ticker`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)

        // Start first ticker
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 100L },
            getTotalAmount = { 100L }
        )

        // Replace with second ticker
        tracker.startProgressTicker(
            intervalMs = 2000L,
            getCurrentPosition = { 200L },
            getTotalAmount = { 200L }
        )

        testScope.advanceTimeBy(2100L)

        // Should only see calls from the second ticker (position 200)
        coVerify {
            api.progressPlaybackSession(match { it["end_position"] == 200L })
        }
    }

    // ------------------------------------------------------------------
    // stopProgressTicker
    // ------------------------------------------------------------------

    @Test
    fun `stopProgressTicker stops sending progress`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(10L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)

        tracker.start(mediaItemId = 1L, fileId = null)
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 100L },
            getTotalAmount = { 100L }
        )

        testScope.advanceTimeBy(1100L)
        tracker.stopProgressTicker()
        testScope.advanceTimeBy(5000L)

        // Only the initial tick before stop should have been called
        coVerify(atMost = 1) { api.progressPlaybackSession(any()) }
    }

    @Test
    fun `stopProgressTicker is safe to call when no ticker is running`() {
        tracker.stopProgressTicker()
        // Should not throw
    }

    @Test
    fun `stopProgressTicker is safe to call multiple times`() {
        tracker.stopProgressTicker()
        tracker.stopProgressTicker()
        tracker.stopProgressTicker()
    }

    // ------------------------------------------------------------------
    // Full lifecycle: start -> tick -> end
    // ------------------------------------------------------------------

    @Test
    fun `full lifecycle start tick end works correctly`() = runTest {
        val responseBody = buildJsonObject {
            put("session_id", JsonPrimitive(42L))
        }
        coEvery { api.startPlaybackSession(any()) } returns Response.success(responseBody)
        coEvery { api.progressPlaybackSession(any()) } returns Response.success(Unit)
        coEvery { api.endPlaybackSession(any()) } returns Response.success(Unit)

        // Start
        val id = tracker.start(mediaItemId = 5L, fileId = 10L, positionUnit = "seconds", startPosition = 0L)
        assertEquals(42L, id)
        assertTrue(tracker.isActive())

        // Tick
        tracker.startProgressTicker(
            intervalMs = 1000L,
            getCurrentPosition = { 1500L },
            getTotalAmount = { 1500L }
        )
        testScope.advanceTimeBy(1100L)

        // End
        tracker.end(endPosition = 3000L, totalAmount = 3000L, completed = true)
        assertFalse(tracker.isActive())

        coVerify(exactly = 1) { api.startPlaybackSession(any()) }
        coVerify(atLeast = 1) { api.progressPlaybackSession(any()) }
        coVerify(exactly = 1) { api.endPlaybackSession(any()) }
    }
}
