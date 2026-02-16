package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import io.mockk.*
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
        coEvery { mockApi.searchMedia(any()) } returns Response.success(items)

        val result = repository.searchMedia(MediaSearchRequest(query = "test")).first()

        assertEquals(2, result.size)
        assertEquals("Test Movie", result[0].title)
    }

    @Test
    fun `searchMedia returns empty list on API failure`() = runTest {
        coEvery { mockApi.searchMedia(any()) } returns Response.error(
            500,
            okhttp3.ResponseBody.create(null, "Server error")
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
        coEvery { mockApi.getMediaById(1L) } returns Response.success(testItem)

        val result = repository.getMediaById(1L).first()

        assertNotNull(result)
        assertEquals("Test Movie", result?.title)
    }

    @Test
    fun `getMediaById returns null on API failure`() = runTest {
        coEvery { mockApi.getMediaById(1L) } returns Response.error(
            404,
            okhttp3.ResponseBody.create(null, "Not found")
        )

        val result = repository.getMediaById(1L).first()

        assertNull(result)
    }

    @Test
    fun `getMediaById returns null on exception`() = runTest {
        coEvery { mockApi.getMediaById(1L) } throws RuntimeException("Network error")

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
            okhttp3.ResponseBody.create(null, "Server error")
        )

        try {
            repository.updateWatchProgress(1L, 0.5)
            fail("Expected exception")
        } catch (e: Exception) {
            assertTrue(e.message?.contains("Failed to update watch progress") == true)
        }
    }

    @Test
    fun `updateFavoriteStatus calls API`() = runTest {
        coEvery { mockApi.updateFavoriteStatus(any(), any()) } returns Response.success(Unit)

        repository.updateFavoriteStatus(1L, true)

        coVerify { mockApi.updateFavoriteStatus(1L, mapOf("favorite" to true)) }
    }

    @Test
    fun `updateFavoriteStatus throws on API failure`() = runTest {
        coEvery { mockApi.updateFavoriteStatus(any(), any()) } returns Response.error(
            500,
            okhttp3.ResponseBody.create(null, "Server error")
        )

        try {
            repository.updateFavoriteStatus(1L, false)
            fail("Expected exception")
        } catch (e: Exception) {
            assertTrue(e.message?.contains("Failed to update favorite status") == true)
        }
    }

    @Test
    fun `searchMedia passes correct params to API`() = runTest {
        coEvery { mockApi.searchMedia(any()) } returns Response.success(emptyList())

        val request = MediaSearchRequest(
            query = "inception",
            mediaType = "movie",
            limit = 50,
            offset = 10
        )

        repository.searchMedia(request).first()

        coVerify {
            mockApi.searchMedia(match { params ->
                params["q"] == "inception" &&
                    params["media_type"] == "movie" &&
                    params["limit"] == "50" &&
                    params["offset"] == "10"
            })
        }
    }
}
