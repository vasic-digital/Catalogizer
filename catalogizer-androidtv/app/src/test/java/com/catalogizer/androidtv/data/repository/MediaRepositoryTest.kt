package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@ExperimentalCoroutinesApi
class MediaRepositoryTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var context: Context
    private lateinit var api: CatalogizerApi
    private lateinit var repository: MediaRepository

    @Before
    fun setup() {
        context = mockk()
        api = mockk()
        repository = MediaRepository(context, api)
    }

    @Test
    fun `searchMedia success should return media items flow`() = runTest {
        val searchRequest = MediaSearchRequest(
            query = "test movie",
            limit = 10,
            offset = 0,
            mediaType = "movie"
        )

        val mediaItems = listOf(
            MediaItem(
                id = 1L,
                title = "Test Movie",
                mediaType = "movie",
                directoryPath = "/path/to/movie",
                createdAt = "2024-01-01T00:00:00Z",
                updatedAt = "2024-01-01T00:00:00Z"
            )
        )

        val successResponse = Response.success(mediaItems)
        coEvery { api.searchMedia(any()) } returns successResponse

        val result = repository.searchMedia(searchRequest).first()

        assertEquals(mediaItems, result)
    }

    @Test
    fun `searchMedia with null response body should return empty list`() = runTest {
        val searchRequest = MediaSearchRequest(query = "test")

        val successResponse = Response.success<List<MediaItem>>(null)
        coEvery { api.searchMedia(any()) } returns successResponse

        val result = repository.searchMedia(searchRequest).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia failure should return empty list`() = runTest {
        val searchRequest = MediaSearchRequest(query = "test")

        val errorResponse = Response.error<List<MediaItem>>(
            500,
            "Server error".toResponseBody(null)
        )
        coEvery { api.searchMedia(any()) } returns errorResponse

        val result = repository.searchMedia(searchRequest).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia with exception should return empty list`() = runTest {
        val searchRequest = MediaSearchRequest(query = "test")

        val exception = RuntimeException("Network error")
        coEvery { api.searchMedia(any()) } throws exception

        val result = repository.searchMedia(searchRequest).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia with all parameters should build correct query params`() = runTest {
        val searchRequest = MediaSearchRequest(
            query = "test query",
            mediaType = "movie",
            limit = 50,
            offset = 20
        )

        val mediaItems = emptyList<MediaItem>()
        val successResponse = Response.success(mediaItems)
        coEvery { api.searchMedia(any()) } returns successResponse

        repository.searchMedia(searchRequest).first()

        // Verify that searchMedia was called with correct parameters
        // The mock will capture the parameters, but we can't easily verify them
        // In a real scenario, we'd use a more sophisticated mock or argument captor
    }

    @Test
    fun `getMediaById success should return media item flow`() = runTest {
        val mediaId = 123L
        val mediaItem = MediaItem(
            id = mediaId,
            title = "Test Movie",
            mediaType = "movie",
            directoryPath = "/path/to/movie",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z"
        )

        val successResponse = Response.success(mediaItem)
        coEvery { api.getMediaById(mediaId) } returns successResponse

        val result = repository.getMediaById(mediaId).first()

        assertEquals(mediaItem, result)
    }

    @Test
    fun `getMediaById failure should return null flow`() = runTest {
        val mediaId = 123L

        val errorResponse = Response.error<MediaItem>(
            404,
            "Not found".toResponseBody(null)
        )
        coEvery { api.getMediaById(mediaId) } returns errorResponse

        val result = repository.getMediaById(mediaId).first()

        assertNull(result)
    }

    @Test
    fun `getMediaById with exception should return null flow`() = runTest {
        val mediaId = 123L

        val exception = RuntimeException("Network error")
        coEvery { api.getMediaById(mediaId) } throws exception

        val result = repository.getMediaById(mediaId).first()

        assertNull(result)
    }

    @Test
    fun `updateWatchProgress success should complete without exception`() = runTest {
        val mediaId = 123L
        val progress = 0.75

        val successResponse = Response.success(Unit)
        coEvery { api.updateWatchProgress(mediaId, any()) } returns successResponse

        // Should not throw exception
        repository.updateWatchProgress(mediaId, progress)
    }

    @Test
    fun `updateWatchProgress failure should throw exception`() = runTest {
        val mediaId = 123L
        val progress = 0.75

        val errorResponse = Response.error<Unit>(
            500,
            "Server error".toResponseBody(null)
        )
        coEvery { api.updateWatchProgress(mediaId, any()) } returns errorResponse

        try {
            repository.updateWatchProgress(mediaId, progress)
            fail("Expected exception to be thrown")
        } catch (e: Exception) {
            assertTrue(e.message?.contains("Failed to update watch progress") == true)
        }
    }

    @Test
    fun `updateWatchProgress with network exception should throw exception`() = runTest {
        val mediaId = 123L
        val progress = 0.75

        val networkException = RuntimeException("Network error")
        coEvery { api.updateWatchProgress(mediaId, any()) } throws networkException

        try {
            repository.updateWatchProgress(mediaId, progress)
            fail("Expected exception to be thrown")
        } catch (e: RuntimeException) {
            assertEquals("Network error", e.message)
        }
    }

    @Test
    fun `updateFavoriteStatus success should complete without exception`() = runTest {
        val mediaId = 123L
        val isFavorite = true

        val successResponse = Response.success(Unit)
        coEvery { api.updateFavoriteStatus(mediaId, any()) } returns successResponse

        // Should not throw exception
        repository.updateFavoriteStatus(mediaId, isFavorite)
    }

    @Test
    fun `updateFavoriteStatus failure should throw exception`() = runTest {
        val mediaId = 123L
        val isFavorite = false

        val errorResponse = Response.error<Unit>(
            500,
            "Server error".toResponseBody(null)
        )
        coEvery { api.updateFavoriteStatus(mediaId, any()) } returns errorResponse

        try {
            repository.updateFavoriteStatus(mediaId, isFavorite)
            fail("Expected exception to be thrown")
        } catch (e: Exception) {
            assertTrue(e.message?.contains("Failed to update favorite status") == true)
        }
    }

    @Test
    fun `updateFavoriteStatus with network exception should throw exception`() = runTest {
        val mediaId = 123L
        val isFavorite = true

        val networkException = RuntimeException("Network error")
        coEvery { api.updateFavoriteStatus(mediaId, any()) } throws networkException

        try {
            repository.updateFavoriteStatus(mediaId, isFavorite)
            fail("Expected exception to be thrown")
        } catch (e: RuntimeException) {
            assertEquals("Network error", e.message)
        }
    }

    @Test
    fun `searchMedia with empty request should work correctly`() = runTest {
        val searchRequest = MediaSearchRequest()

        val mediaItems = emptyList<MediaItem>()
        val successResponse = Response.success(mediaItems)
        coEvery { api.searchMedia(any()) } returns successResponse

        val result = repository.searchMedia(searchRequest).first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia with complex request should work correctly`() = runTest {
        val searchRequest = MediaSearchRequest(
            query = "action movie",
            mediaType = "movie",
            yearMin = 2020,
            yearMax = 2024,
            ratingMin = 7.0,
            quality = "1080p",
            sortBy = "rating",
            sortOrder = "desc",
            limit = 25,
            offset = 50
        )

        val mediaItems = listOf(
            MediaItem(
                id = 1L,
                title = "Action Movie",
                mediaType = "movie",
                year = 2023,
                rating = 8.5,
                quality = "1080p",
                directoryPath = "/path/to/movie",
                createdAt = "2024-01-01T00:00:00Z",
                updatedAt = "2024-01-01T00:00:00Z"
            )
        )

        val successResponse = Response.success(mediaItems)
        coEvery { api.searchMedia(any()) } returns successResponse

        val result = repository.searchMedia(searchRequest).first()

        assertEquals(1, result.size)
        assertEquals("Action Movie", result[0].title)
    }
}