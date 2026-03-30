package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchResponse
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.remote.EntityStatsResponse
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class MediaRepositoryTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockContext: Context
    private lateinit var mockApi: CatalogizerApi
    private lateinit var repository: MediaRepository

    private val testItem = MediaItem(
        id = 1L, title = "Test Movie", mediaType = "movie",
        directoryPath = "/media/test", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @Before
    fun setup() {
        mockContext = mockk(relaxed = true)
        mockApi = mockk(relaxed = true)
        repository = MediaRepository(mockContext, mockApi)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // --- browseEntities ---

    @Test
    fun `browseEntities returns items on success`() = runTest {
        val response = MediaSearchResponse(items = listOf(testItem), total = 1)
        coEvery { mockApi.browseEntities(any(), any()) } returns Response.success(response)

        val result = repository.browseEntities("movie").first()

        assertEquals(1, result.size)
        assertEquals("Test Movie", result[0].title)
    }

    @Test
    fun `browseEntities returns empty on API failure`() = runTest {
        coEvery { mockApi.browseEntities(any(), any()) } returns Response.error(
            500, okhttp3.ResponseBody.create(null, "Error")
        )

        val result = repository.browseEntities("movie").first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `browseEntities returns empty on exception`() = runTest {
        coEvery { mockApi.browseEntities(any(), any()) } throws RuntimeException("Network error")

        val result = repository.browseEntities("movie").first()

        assertTrue(result.isEmpty())
    }

    // --- searchMedia ---

    @Test
    fun `searchMedia returns items on success`() = runTest {
        val response = MediaSearchResponse(items = listOf(testItem), total = 1)
        coEvery { mockApi.searchMedia(any()) } returns Response.success(response)

        val request = com.catalogizer.androidtv.data.models.MediaSearchRequest(query = "test")
        val result = repository.searchMedia(request).first()

        assertEquals(1, result.size)
    }

    @Test
    fun `searchMedia returns empty on failure`() = runTest {
        coEvery { mockApi.searchMedia(any()) } returns Response.error(
            404, okhttp3.ResponseBody.create(null, "Not found")
        )

        val request = com.catalogizer.androidtv.data.models.MediaSearchRequest(query = "test")
        val result = repository.searchMedia(request).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia returns empty on exception`() = runTest {
        coEvery { mockApi.searchMedia(any()) } throws RuntimeException("Timeout")

        val request = com.catalogizer.androidtv.data.models.MediaSearchRequest(query = "test")
        val result = repository.searchMedia(request).first()

        assertTrue(result.isEmpty())
    }

    // --- getMediaById ---

    @Test
    fun `getMediaById tries entity endpoint first`() = runTest {
        coEvery { mockApi.getEntityById(1L) } returns Response.success(testItem)

        val result = repository.getMediaById(1L).first()

        assertNotNull(result)
        assertEquals("Test Movie", result?.title)
    }

    @Test
    fun `getMediaById falls back to media endpoint`() = runTest {
        coEvery { mockApi.getEntityById(1L) } returns Response.error(
            404, okhttp3.ResponseBody.create(null, "Not found")
        )
        coEvery { mockApi.getMediaById(1L) } returns Response.success(testItem)

        val result = repository.getMediaById(1L).first()

        assertNotNull(result)
    }

    @Test
    fun `getMediaById returns null on all failures`() = runTest {
        coEvery { mockApi.getEntityById(1L) } returns Response.error(
            404, okhttp3.ResponseBody.create(null, "Not found")
        )
        coEvery { mockApi.getMediaById(1L) } returns Response.error(
            404, okhttp3.ResponseBody.create(null, "Not found")
        )

        val result = repository.getMediaById(1L).first()

        assertNull(result)
    }

    @Test
    fun `getMediaById returns null on exception`() = runTest {
        coEvery { mockApi.getEntityById(any()) } throws RuntimeException("Error")

        val result = repository.getMediaById(1L).first()

        assertNull(result)
    }

    // --- updateWatchProgress ---

    @Test
    fun `updateWatchProgress calls API`() = runTest {
        coEvery { mockApi.updateWatchProgress(any(), any()) } returns Response.success(Unit)

        repository.updateWatchProgress(1L, 0.5)

        coVerify { mockApi.updateWatchProgress(1L, mapOf("progress" to 0.5)) }
    }

    @Test
    fun `updateWatchProgress throws on failure`() = runTest {
        coEvery { mockApi.updateWatchProgress(any(), any()) } returns Response.error(
            500, okhttp3.ResponseBody.create(null, "Error")
        )

        try {
            repository.updateWatchProgress(1L, 0.5)
            fail("Should have thrown")
        } catch (e: Exception) {
            assertTrue(e.message!!.contains("Failed"))
        }
    }

    // --- checkFavorite ---

    @Test
    fun `checkFavorite returns true when favorited`() = runTest {
        coEvery { mockApi.checkFavorite(any(), any()) } returns Response.success(
            mapOf("is_favorite" to true)
        )

        val result = repository.checkFavorite("movie", 1L)

        assertTrue(result)
    }

    @Test
    fun `checkFavorite returns false when not favorited`() = runTest {
        coEvery { mockApi.checkFavorite(any(), any()) } returns Response.success(
            mapOf("is_favorite" to false)
        )

        val result = repository.checkFavorite("movie", 1L)

        assertFalse(result)
    }

    @Test
    fun `checkFavorite returns false on error`() = runTest {
        coEvery { mockApi.checkFavorite(any(), any()) } throws RuntimeException("Error")

        val result = repository.checkFavorite("movie", 1L)

        assertFalse(result)
    }

    // --- toggleFavorite ---

    @Test
    fun `toggleFavorite adds when not currently favorite`() = runTest {
        coEvery { mockApi.addFavorite(any()) } returns Response.success(mapOf("status" to "ok"))

        val result = repository.toggleFavorite("movie", 1L, false)

        assertTrue(result)
        coVerify { mockApi.addFavorite(any()) }
    }

    @Test
    fun `toggleFavorite removes when currently favorite`() = runTest {
        coEvery { mockApi.removeFavorite(any(), any()) } returns Response.success(mapOf("status" to "ok"))

        val result = repository.toggleFavorite("movie", 1L, true)

        assertFalse(result)
        coVerify { mockApi.removeFavorite("movie", 1L) }
    }

    // --- getEntityStats ---

    @Test
    fun `getEntityStats returns data on success`() = runTest {
        val stats = EntityStatsResponse(totalEntities = 500, byType = mapOf("movie" to 200))
        coEvery { mockApi.getEntityStats() } returns Response.success(stats)

        val result = repository.getEntityStats()

        assertEquals(500, result.first)
        assertEquals(200, result.second["movie"])
    }

    @Test
    fun `getEntityStats returns defaults on failure`() = runTest {
        coEvery { mockApi.getEntityStats() } returns Response.error(
            500, okhttp3.ResponseBody.create(null, "Error")
        )

        val result = repository.getEntityStats()

        assertEquals(0, result.first)
        assertTrue(result.second.isEmpty())
    }

    // --- getCollections ---

    @Test
    fun `getCollections returns null on error`() = runTest {
        coEvery { mockApi.getCollections() } throws RuntimeException("Error")

        val result = repository.getCollections()

        assertNull(result)
    }

    // --- getPlaylists ---

    @Test
    fun `getPlaylists returns null on error`() = runTest {
        coEvery { mockApi.getPlaylists() } throws RuntimeException("Error")

        val result = repository.getPlaylists()

        assertNull(result)
    }

    // --- createPlaylist ---

    @Test
    fun `createPlaylist returns null on error`() = runTest {
        coEvery { mockApi.createPlaylist(any()) } throws RuntimeException("Error")

        val result = repository.createPlaylist("My Playlist")

        assertNull(result)
    }
}
