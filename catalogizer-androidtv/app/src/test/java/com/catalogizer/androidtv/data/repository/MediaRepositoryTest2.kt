package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.models.MediaSearchResponse
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import io.mockk.*
import okhttp3.ResponseBody.Companion.toResponseBody
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class MediaRepositoryTest2 {

    private val mockContext = mockk<android.content.Context>(relaxed = true)
    private val mockApi = mockk<CatalogizerApi>(relaxed = true)
    private lateinit var repository: MediaRepository

    private val testItem = MediaItem(
        id = 1L,
        title = "Test Movie",
        mediaType = "movie",
        directoryPath = "/movies/test",
        createdAt = "2024-01-01",
        updatedAt = "2024-01-01"
    )

    @Before
    fun setup() {
        repository = MediaRepository(mockContext, mockApi)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `searchMedia returns items on success`() = runTest {
        val items = listOf(testItem, testItem.copy(id = 2, title = "Movie 2"))
        val searchResponse = MediaSearchResponse(items = items, total = items.size, limit = 20, offset = 0)
        coEvery { mockApi.searchMedia(any()) } returns Response.success(searchResponse)

        val result = repository.searchMedia(MediaSearchRequest(query = "test")).first()

        assertEquals(2, result.size)
        assertEquals("Test Movie", result[0].title)
    }

    @Test
    fun `searchMedia returns empty list on API failure`() = runTest {
        coEvery { mockApi.searchMedia(any()) } returns Response.error<MediaSearchResponse>(
            500,
            "Server error".toResponseBody(null)
        )

        val result = repository.searchMedia(MediaSearchRequest(query = "test")).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia returns empty list on exception`() = runTest {
        coEvery { mockApi.searchMedia(any()) } throws RuntimeException("Network error")

        val result = repository.searchMedia(MediaSearchRequest(query = "test")).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `getMediaById returns item on success`() = runTest {
        coEvery { mockApi.getEntityById(1L) } returns Response.success(testItem)

        val result = repository.getMediaById(1L).first()

        assertNotNull(result)
        assertEquals("Test Movie", result?.title)
    }

    @Test
    fun `getMediaById returns null on API failure`() = runTest {
        coEvery { mockApi.getEntityById(1L) } returns Response.error(
            404,
            "Not found".toResponseBody(null)
        )
        coEvery { mockApi.getMediaById(1L) } returns Response.error(
            404,
            "Not found".toResponseBody(null)
        )

        val result = repository.getMediaById(1L).first()

        assertNull(result)
    }

    @Test
    fun `getMediaById returns null on exception`() = runTest {
        coEvery { mockApi.getEntityById(1L) } throws RuntimeException("Network error")

        val result = repository.getMediaById(1L).first()

        assertNull(result)
    }

    @Test
    fun `updateWatchProgress calls API`() = runTest {
        coEvery { mockApi.updateWatchProgress(any(), any()) } returns Response.success(Unit)

        repository.updateWatchProgress(1L, 0.5)

        coVerify { mockApi.updateWatchProgress(1L, mapOf("progress" to 0.5)) }
    }

    @Test
    fun `updateWatchProgress throws on API failure`() = runTest {
        coEvery { mockApi.updateWatchProgress(any(), any()) } returns Response.error(
            500,
            "Server error".toResponseBody(null)
        )

        try {
            repository.updateWatchProgress(1L, 0.5)
            fail("Expected exception")
        } catch (e: Exception) {
            assertTrue(e.message?.contains("Failed to update watch progress") == true)
        }
    }

    @Test
    fun `favorites calls API`() = runTest {
        coEvery { mockApi.addFavorite(any()) } returns Response.success(mapOf("status" to "ok"))

        repository.addFavorite("movie", 1L)

        coVerify { mockApi.addFavorite(any()) }
    }

    @Test
    fun `favorites throws on API failure`() = runTest {
        coEvery { mockApi.removeFavorite(any(), any()) } returns Response.error<Map<String, String>>(
            500,
            "Server error".toResponseBody(null)
        )

        try {
            repository.removeFavorite("movie", 1L)
            fail("Expected exception")
        } catch (e: Exception) {
            assertTrue(e.message?.contains("Failed to remove favorite") == true)
        }
    }

    @Test
    fun `searchMedia passes correct params to API`() = runTest {
        val searchResponse = MediaSearchResponse(items = emptyList(), total = 0, limit = 50, offset = 10)
        coEvery { mockApi.searchMedia(any()) } returns Response.success(searchResponse)

        val request = MediaSearchRequest(
            query = "inception",
            mediaType = "movie",
            limit = 50,
            offset = 10
        )

        repository.searchMedia(request).first()

        coVerify {
            mockApi.searchMedia(match { params ->
                params["query"] == "inception" &&
                    params["media_types"] == "movie" &&
                    params["limit"] == "50" &&
                    params["offset"] == "10"
            })
        }
    }
}
