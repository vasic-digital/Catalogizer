package com.catalogizer.android.data.playback

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.remote.CatalogizerApi
import android.util.Log
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class PlaybackRepositoryTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockApi = mockk<CatalogizerApi>()
    private lateinit var repository: PlaybackRepository

    @Before
    fun setup() {
        mockkStatic(Log::class)
        every { Log.w(any<String>(), any<String>()) } returns 0
        every { Log.w(any<String>(), any<Throwable>()) } returns 0
        repository = PlaybackRepository(mockApi)
    }

    @After
    fun tearDown() {
        clearAllMocks()
        unmockkStatic(Log::class)
    }

    // --- getProgress ---

    @Test
    fun `getProgress returns UiPlaybackProgress on success`() = runTest {
        val progressJson = buildJsonObject {
            put("progress", buildJsonObject {
                put("position_unit", JsonPrimitive("seconds"))
                put("duration_total", JsonPrimitive(3600L))
                put("last_position", JsonPrimitive(1800L))
                put("last_session_amount", JsonPrimitive(900L))
                put("total_reproductions", JsonPrimitive(3L))
                put("aggregate_amount", JsonPrimitive(5400L))
                put("last_session_ended_at", JsonPrimitive("2026-01-15T10:30:00Z"))
            })
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(progressJson)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertEquals(1L, result!!.mediaItemId)
        assertEquals("seconds", result.positionUnit)
        assertEquals(3600L, result.durationTotal)
        assertEquals(1800L, result.lastPosition)
        assertEquals(900L, result.lastSessionAmount)
        assertEquals(3L, result.totalReproductions)
        assertEquals(5400L, result.aggregateAmount)
        assertNotNull(result.lastSessionEndedAtMs)
    }

    @Test
    fun `getProgress returns null on HTTP failure`() = runTest {
        val errorBody = okhttp3.ResponseBody.create(null, "Not found")
        coEvery { mockApi.getEntityProgress(999L) } returns Response.error(404, errorBody)

        val result = repository.getProgress(999L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null when body is null`() = runTest {
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(null)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null when progress field is missing`() = runTest {
        val body = buildJsonObject {
            put("other_field", JsonPrimitive("value"))
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null when progress field is literal null`() = runTest {
        val body = buildJsonObject {
            put("progress", JsonNull)
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(body)

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress returns null on exception`() = runTest {
        coEvery { mockApi.getEntityProgress(1L) } throws RuntimeException("Network error")

        val result = repository.getProgress(1L)

        assertNull(result)
    }

    @Test
    fun `getProgress uses default values for missing optional fields`() = runTest {
        val progressJson = buildJsonObject {
            put("progress", buildJsonObject {
                // Only position_unit, no other fields
                put("position_unit", JsonPrimitive("pages"))
            })
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(progressJson)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertEquals("pages", result!!.positionUnit)
        assertNull(result.durationTotal)
        assertEquals(0L, result.lastPosition)
        assertEquals(0L, result.lastSessionAmount)
        assertEquals(0L, result.totalReproductions)
        assertEquals(0L, result.aggregateAmount)
        assertNull(result.lastSessionEndedAtMs)
    }

    @Test
    fun `getProgress defaults positionUnit to seconds when missing`() = runTest {
        val progressJson = buildJsonObject {
            put("progress", buildJsonObject {
                put("last_position", JsonPrimitive(100L))
            })
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(progressJson)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertEquals("seconds", result!!.positionUnit)
    }

    @Test
    fun `getProgress handles invalid ISO8601 date gracefully`() = runTest {
        val progressJson = buildJsonObject {
            put("progress", buildJsonObject {
                put("position_unit", JsonPrimitive("seconds"))
                put("last_session_ended_at", JsonPrimitive("not-a-date"))
            })
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(progressJson)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNull(result!!.lastSessionEndedAtMs)
    }

    @Test
    fun `getProgress handles null string in date field`() = runTest {
        val progressJson = buildJsonObject {
            put("progress", buildJsonObject {
                put("position_unit", JsonPrimitive("seconds"))
                put("last_session_ended_at", JsonPrimitive("null"))
            })
        }
        coEvery { mockApi.getEntityProgress(1L) } returns Response.success(progressJson)

        val result = repository.getProgress(1L)

        assertNotNull(result)
        assertNull(result!!.lastSessionEndedAtMs)
    }

    // --- listHistory ---

    @Test
    fun `listHistory returns sessions on success`() = runTest {
        val historyJson = buildJsonObject {
            put("sessions", buildJsonArray {
                add(buildJsonObject {
                    put("id", JsonPrimitive(1L))
                    put("position_unit", JsonPrimitive("seconds"))
                    put("start_position", JsonPrimitive(0L))
                    put("end_position", JsonPrimitive(1800L))
                    put("total_amount", JsonPrimitive(1800L))
                    put("started_at", JsonPrimitive("2026-01-15T10:00:00Z"))
                    put("ended_at", JsonPrimitive("2026-01-15T10:30:00Z"))
                    put("completed", JsonPrimitive(false))
                })
                add(buildJsonObject {
                    put("id", JsonPrimitive(2L))
                    put("position_unit", JsonPrimitive("seconds"))
                    put("start_position", JsonPrimitive(1800L))
                    put("end_position", JsonPrimitive(3600L))
                    put("total_amount", JsonPrimitive(1800L))
                    put("started_at", JsonPrimitive("2026-01-15T11:00:00Z"))
                    put("ended_at", JsonPrimitive("2026-01-15T11:30:00Z"))
                    put("completed", JsonPrimitive(true))
                })
            })
        }
        coEvery { mockApi.getEntityHistory(1L, 50) } returns Response.success(historyJson)

        val result = repository.listHistory(1L)

        assertEquals(2, result.size)
        assertEquals(1L, result[0].id)
        assertEquals(0L, result[0].startPosition)
        assertEquals(1800L, result[0].endPosition)
        assertFalse(result[0].completed)
        assertEquals(2L, result[1].id)
        assertTrue(result[1].completed)
    }

    @Test
    fun `listHistory returns empty list on HTTP failure`() = runTest {
        val errorBody = okhttp3.ResponseBody.create(null, "Error")
        coEvery { mockApi.getEntityHistory(1L, 50) } returns Response.error(500, errorBody)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list when body is null`() = runTest {
        coEvery { mockApi.getEntityHistory(1L, 50) } returns Response.success(null)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list when sessions field is missing`() = runTest {
        val body = buildJsonObject {
            put("other", JsonPrimitive("value"))
        }
        coEvery { mockApi.getEntityHistory(1L, 50) } returns Response.success(body)

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory returns empty list on exception`() = runTest {
        coEvery { mockApi.getEntityHistory(1L, 50) } throws RuntimeException("Network error")

        val result = repository.listHistory(1L)

        assertTrue(result.isEmpty())
    }

    @Test
    fun `listHistory passes custom limit`() = runTest {
        val body = buildJsonObject {
            put("sessions", buildJsonArray {})
        }
        coEvery { mockApi.getEntityHistory(1L, 10) } returns Response.success(body)

        repository.listHistory(1L, limit = 10)

        coVerify { mockApi.getEntityHistory(1L, 10) }
    }

    @Test
    fun `listHistory handles session with missing optional fields`() = runTest {
        val historyJson = buildJsonObject {
            put("sessions", buildJsonArray {
                add(buildJsonObject {
                    // Minimal session with only started_at
                    put("started_at", JsonPrimitive("2026-01-15T10:00:00Z"))
                })
            })
        }
        coEvery { mockApi.getEntityHistory(1L, 50) } returns Response.success(historyJson)

        val result = repository.listHistory(1L)

        assertEquals(1, result.size)
        assertEquals(0L, result[0].id)
        assertEquals("seconds", result[0].positionUnit)
        assertEquals(0L, result[0].startPosition)
        assertEquals(0L, result[0].endPosition)
        assertEquals(0L, result[0].totalAmount)
        assertFalse(result[0].completed)
    }

    @Test
    fun `listHistory handles session with null ended_at`() = runTest {
        val historyJson = buildJsonObject {
            put("sessions", buildJsonArray {
                add(buildJsonObject {
                    put("id", JsonPrimitive(1L))
                    put("started_at", JsonPrimitive("2026-01-15T10:00:00Z"))
                    // ended_at deliberately omitted (active session)
                })
            })
        }
        coEvery { mockApi.getEntityHistory(1L, 50) } returns Response.success(historyJson)

        val result = repository.listHistory(1L)

        assertEquals(1, result.size)
        assertNull(result[0].endedAtMs)
    }
}
