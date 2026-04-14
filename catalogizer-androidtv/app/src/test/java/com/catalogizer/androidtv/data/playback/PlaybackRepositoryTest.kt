package com.catalogizer.androidtv.data.playback

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import io.mockk.clearAllMocks
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonArray
import kotlinx.serialization.json.buildJsonObject
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class PlaybackRepositoryTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var api: CatalogizerApi
    private lateinit var repository: PlaybackRepository

    @Before
    fun setup() {
        api = mockk()
        repository = PlaybackRepository(api)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // ------------------------------------------------------------------
    // getProgress — success cases
    // ------------------------------------------------------------------

    @Test
    fun `getProgress returns UiPlaybackProgress on successful response`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("seconds"))
            put("duration_total", JsonPrimitive(7200L))
            put("last_position", JsonPrimitive(3600L))
            put("last_session_amount", JsonPrimitive(1800L))
            put("total_reproductions", JsonPrimitive(3L))
            put("aggregate_amount", JsonPrimitive(5400L))
            put("last_session_ended_at", JsonPrimitive("2026-01-15T10:30:00Z"))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(42L) } returns Response.success(body)

        val result = repository.getProgress(42L)

        assertNotNull(result)
        assertEquals(42L, result!!.mediaItemId)
        assertEquals("seconds", result.positionUnit)
        assertEquals(7200L, result.durationTotal)
        assertEquals(3600L, result.lastPosition)
        assertEquals(1800L, result.lastSessionAmount)
        assertEquals(3L, result.totalReproductions)
        assertEquals(5400L, result.aggregateAmount)
        assertNotNull(result.lastSessionEndedAtMs)
    }

    @Test
    fun `getProgress returns progress with null duration when field missing`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("pages"))
            put("last_position", JsonPrimitive(50L))
            put("last_session_amount", JsonPrimitive(10L))
            put("total_reproductions", JsonPrimitive(1L))
            put("aggregate_amount", JsonPrimitive(50L))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNull(result!!.durationTotal)
        assertEquals("pages", result.positionUnit)
        assertEquals(50L, result.lastPosition)
    }

    @Test
    fun `getProgress defaults position_unit to seconds when missing`() = runTest {
        val progressJson = buildJsonObject {
            put("last_position", JsonPrimitive(100L))
            put("last_session_amount", JsonPrimitive(100L))
            put("total_reproductions", JsonPrimitive(1L))
            put("aggregate_amount", JsonPrimitive(100L))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertEquals("seconds", result!!.positionUnit)
    }

    @Test
    fun `getProgress defaults numeric fields to zero when missing`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("seconds"))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertEquals(0L, result!!.lastPosition)
        assertEquals(0L, result.lastSessionAmount)
        assertEquals(0L, result.totalReproductions)
        assertEquals(0L, result.aggregateAmount)
    }

    // ------------------------------------------------------------------
    // getProgress — null / missing progress cases
    // ------------------------------------------------------------------

    @Test
    fun `getProgress returns null when progress is json null`() = runTest {
        val body = buildJsonObject {
            put("progress", JsonNull)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null when progress key is absent`() = runTest {
        val body = buildJsonObject {
            put("other_field", JsonPrimitive("value"))
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null when response body is null`() = runTest {
        coEvery { api.getEntityProgress(1L) } returns Response.success(null)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    // ------------------------------------------------------------------
    // getProgress — error cases
    // ------------------------------------------------------------------

    @Test
    fun `getProgress returns null on HTTP 404`() = runTest {
        coEvery { api.getEntityProgress(999L) } returns Response.error(
            404, "Not found".toResponseBody(null)
        )

        val result = repository.getProgress(999L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null on HTTP 500`() = runTest {
        coEvery { api.getEntityProgress(1L) } returns Response.error(
            500, "Server error".toResponseBody(null)
        )

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null on network exception`() = runTest {
        coEvery { api.getEntityProgress(1L) } throws RuntimeException("Connection refused")

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null on JSON parse exception`() = runTest {
        // Body contains progress but it is a primitive, not an object
        val body = buildJsonObject {
            put("progress", JsonPrimitive("not-an-object"))
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    // ------------------------------------------------------------------
    // getProgress — ISO 8601 timestamp parsing
    // ------------------------------------------------------------------

    @Test
    fun `getProgress parses ISO 8601 timestamp with offset`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("seconds"))
            put("last_position", JsonPrimitive(0L))
            put("last_session_amount", JsonPrimitive(0L))
            put("total_reproductions", JsonPrimitive(0L))
            put("aggregate_amount", JsonPrimitive(0L))
            put("last_session_ended_at", JsonPrimitive("2026-03-15T14:30:00+02:00"))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNotNull(result!!.lastSessionEndedAtMs)
        assertTrue(result.lastSessionEndedAtMs!! > 0L)
    }

    @Test
    fun `getProgress returns null timestamp for invalid date string`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("seconds"))
            put("last_position", JsonPrimitive(0L))
            put("last_session_amount", JsonPrimitive(0L))
            put("total_reproductions", JsonPrimitive(0L))
            put("aggregate_amount", JsonPrimitive(0L))
            put("last_session_ended_at", JsonPrimitive("not-a-date"))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNull(result!!.lastSessionEndedAtMs)
    }

    @Test
    fun `getProgress returns null timestamp for empty date string`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("seconds"))
            put("last_position", JsonPrimitive(0L))
            put("last_session_amount", JsonPrimitive(0L))
            put("total_reproductions", JsonPrimitive(0L))
            put("aggregate_amount", JsonPrimitive(0L))
            put("last_session_ended_at", JsonPrimitive(""))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNull(result!!.lastSessionEndedAtMs)
    }

    @Test
    fun `getProgress returns null timestamp for literal null string`() = runTest {
        val progressJson = buildJsonObject {
            put("position_unit", JsonPrimitive("seconds"))
            put("last_position", JsonPrimitive(0L))
            put("last_session_amount", JsonPrimitive(0L))
            put("total_reproductions", JsonPrimitive(0L))
            put("aggregate_amount", JsonPrimitive(0L))
            put("last_session_ended_at", JsonPrimitive("null"))
        }
        val body = buildJsonObject {
            put("progress", progressJson)
        }
        coEvery { api.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNull(result!!.lastSessionEndedAtMs)
    }

    // ------------------------------------------------------------------
    // getProgress — API call verification
    // ------------------------------------------------------------------

    @Test
    fun `getProgress calls api with correct media item id`() = runTest {
        coEvery { api.getEntityProgress(any()) } returns Response.success(
            buildJsonObject { put("progress", JsonNull) }
        )

        repository.getProgress(77L)

        coVerify(exactly = 1) { api.getEntityProgress(77L) }
    }

    // ------------------------------------------------------------------
    // listHistory — success cases
    // ------------------------------------------------------------------

    @Test
    fun `listHistory returns sessions from valid response`() = runTest {
        val sessionsArray = buildJsonArray {
            add(buildJsonObject {
                put("id", JsonPrimitive(1L))
                put("position_unit", JsonPrimitive("seconds"))
                put("start_position", JsonPrimitive(0L))
                put("end_position", JsonPrimitive(3600L))
                put("total_amount", JsonPrimitive(3600L))
                put("started_at", JsonPrimitive("2026-01-10T08:00:00Z"))
                put("ended_at", JsonPrimitive("2026-01-10T09:00:00Z"))
                put("completed", JsonPrimitive(true))
            })
            add(buildJsonObject {
                put("id", JsonPrimitive(2L))
                put("position_unit", JsonPrimitive("seconds"))
                put("start_position", JsonPrimitive(0L))
                put("end_position", JsonPrimitive(1800L))
                put("total_amount", JsonPrimitive(1800L))
                put("started_at", JsonPrimitive("2026-01-11T10:00:00Z"))
                put("ended_at", JsonPrimitive("2026-01-11T10:30:00Z"))
                put("completed", JsonPrimitive(false))
            })
        }
        val body = buildJsonObject {
            put("sessions", sessionsArray)
        }
        coEvery { api.getEntityHistory(42L, 50) } returns Response.success(body)

        val result = repository.listHistory(42L)

        assertEquals(2, result.size)
        assertEquals(1L, result[0].id)
        assertTrue(result[0].completed)
        assertEquals(2L, result[1].id)
        assertEquals(false, result[1].completed)
        assertEquals(0L, result[0].startPosition)
        assertEquals(3600L, result[0].endPosition)
    }

    @Test
    fun `listHistory passes custom limit to API`() = runTest {
        val body = buildJsonObject {
            put("sessions", buildJsonArray { })
        }
        coEvery { api.getEntityHistory(1L, 10) } returns Response.success(body)

        repository.listHistory(1L, limit = 10)

        coVerify(exactly = 1) { api.getEntityHistory(1L, 10) }
    }

    @Test
    fun `listHistory defaults missing session fields to zero or false`() = runTest {
        val sessionsArray = buildJsonArray {
            add(buildJsonObject {
                put("started_at", JsonPrimitive("2026-01-10T08:00:00Z"))
            })
        }
        val body = buildJsonObject {
            put("sessions", sessionsArray)
        }
        coEvery { api.getEntityHistory(1L, 50) } returns Response.success(body)

        val result = repository.listHistory(1L)

        assertEquals(1, result.size)
        assertEquals(0L, result[0].id)
        assertEquals("seconds", result[0].positionUnit)
        assertEquals(0L, result[0].startPosition)
        assertEquals(0L, result[0].endPosition)
        assertEquals(0L, result[0].totalAmount)
        assertEquals(false, result[0].completed)
    }

    // ------------------------------------------------------------------
    // listHistory — empty / missing cases
    // ------------------------------------------------------------------

    @Test
    fun `listHistory returns empty list when sessions key is absent`() = runTest {
        val body = buildJsonObject {
            put("other", JsonPrimitive("value"))
        }
        coEvery { api.getEntityHistory(1L, 50) } returns Response.success(body)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list when response body is null`() = runTest {
        coEvery { api.getEntityHistory(1L, 50) } returns Response.success(null)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list on empty sessions array`() = runTest {
        val body = buildJsonObject {
            put("sessions", buildJsonArray { })
        }
        coEvery { api.getEntityHistory(1L, 50) } returns Response.success(body)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    // ------------------------------------------------------------------
    // listHistory — error cases
    // ------------------------------------------------------------------

    @Test
    fun `listHistory returns empty list on HTTP error`() = runTest {
        coEvery { api.getEntityHistory(1L, 50) } returns Response.error(
            500, "Server error".toResponseBody(null)
        )

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list on network exception`() = runTest {
        coEvery { api.getEntityHistory(1L, 50) } throws RuntimeException("Timeout")

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list on JSON parse exception`() = runTest {
        val body = buildJsonObject {
            put("sessions", JsonPrimitive("not-an-array"))
        }
        coEvery { api.getEntityHistory(1L, 50) } returns Response.success(body)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }
}
