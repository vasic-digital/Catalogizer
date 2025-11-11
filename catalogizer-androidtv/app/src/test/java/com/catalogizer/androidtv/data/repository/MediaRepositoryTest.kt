package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import io.mockk.*
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.After
import org.junit.Before
import org.junit.Test
import retrofit2.Response
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class MediaRepositoryTest {

    private lateinit var repository: MediaRepository
    private lateinit var mockApi: CatalogizerApi

    // Test data
    private val testMediaItem = MediaItem(
        id = 123L,
        title = "Test Movie",
        mediaType = "movie",
        year = 2024,
        rating = 8.5,
        quality = "1080p",
        directoryPath = "/movies/test.mp4",
        storageRootName = "main_storage",
        fileSize = 1024000000L,
        duration = 7200,
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z"
    )

    private val testMediaList = listOf(
        testMediaItem,
        MediaItem(
            id = 124L,
            title = "Test Movie 2",
            mediaType = "movie",
            year = 2024,
            rating = 7.5,
            quality = "720p",
            directoryPath = "/movies/test2.mp4",
            storageRootName = "main_storage",
            fileSize = 512000000L,
            duration = 5400,
            createdAt = "2024-01-02T00:00:00Z",
            updatedAt = "2024-01-02T00:00:00Z"
        ),
        MediaItem(
            id = 125L,
            title = "Test TV Show",
            mediaType = "tv_show",
            year = 2023,
            rating = 9.0,
            quality = "1080p",
            directoryPath = "/tv/test.mp4",
            storageRootName = "main_storage",
            fileSize = 768000000L,
            duration = 3600,
            createdAt = "2024-01-03T00:00:00Z",
            updatedAt = "2024-01-03T00:00:00Z"
        )
    )

    @Before
    fun setup() {
        mockApi = mockk()
        repository = MediaRepository(mockApi)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // searchMedia Tests

    @Test
    fun `searchMedia returns list of media items successfully`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(
            query = "Test",
            limit = 20,
            offset = 0
        )
        coEvery { mockApi.searchMedia(any()) } returns Response.success(testMediaList)

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertEquals(3, result.size)
        assertEquals("Test Movie", result[0].title)
        assertEquals("Test Movie 2", result[1].title)
        assertEquals("Test TV Show", result[2].title)
    }

    @Test
    fun `searchMedia with all filter parameters`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(
            query = "Matrix",
            mediaType = "movie",
            yearMin = 1999,
            yearMax = 1999,
            ratingMin = 8.0,
            quality = "1080p",
            sortBy = "rating",
            sortOrder = "desc",
            limit = 10,
            offset = 0
        )
        coEvery { mockApi.searchMedia(any()) } returns Response.success(listOf(testMediaItem))

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertEquals(1, result.size)
        coVerify {
            mockApi.searchMedia(match {
                it["query"] == "Matrix" &&
                it["media_type"] == "movie" &&
                it["year_min"] == "1999" &&
                it["year_max"] == "1999" &&
                it["rating_min"] == "8.0" &&
                it["quality"] == "1080p" &&
                it["sort_by"] == "rating" &&
                it["sort_order"] == "desc" &&
                it["limit"] == "10" &&
                it["offset"] == "0"
            })
        }
    }

    @Test
    fun `searchMedia with null optional parameters`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(
            query = null,
            mediaType = null,
            yearMin = null,
            yearMax = null,
            ratingMin = null,
            quality = null,
            sortBy = null,
            sortOrder = null,
            limit = 20,
            offset = 0
        )
        coEvery { mockApi.searchMedia(any()) } returns Response.success(testMediaList)

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertEquals(3, result.size)
        coVerify {
            mockApi.searchMedia(match {
                it.size == 2 && // Only limit and offset should be present
                it["limit"] == "20" &&
                it["offset"] == "0"
            })
        }
    }

    @Test
    fun `searchMedia returns empty list on API error`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(query = "Test")
        coEvery { mockApi.searchMedia(any()) } returns Response.error(
            404,
            "Not found".toResponseBody()
        )

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia returns empty list on null body`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(query = "Test")
        coEvery { mockApi.searchMedia(any()) } returns Response.success(null)

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia returns empty list on exception`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(query = "Test")
        coEvery { mockApi.searchMedia(any()) } throws Exception("Network error")

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertTrue(result.isEmpty())
    }

    @Test
    fun `searchMedia handles 500 server error`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(query = "Test")
        coEvery { mockApi.searchMedia(any()) } returns Response.error(
            500,
            "Internal server error".toResponseBody()
        )

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertTrue(result.isEmpty())
    }

    // getMediaById Tests

    @Test
    fun `getMediaById returns media item successfully`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(mapOf("id" to "123")) } returns Response.success(listOf(testMediaItem))

        // When
        val result = repository.getMediaById(123L).first()

        // Then
        assertNotNull(result)
        assertEquals(123L, result.id)
        assertEquals("Test Movie", result.title)
    }

    @Test
    fun `getMediaById returns first item when multiple items returned`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(mapOf("id" to "123")) } returns Response.success(testMediaList)

        // When
        val result = repository.getMediaById(123L).first()

        // Then
        assertNotNull(result)
        assertEquals(123L, result.id)
        assertEquals("Test Movie", result.title)
    }

    @Test
    fun `getMediaById returns null when media not found`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(mapOf("id" to "999")) } returns Response.success(emptyList())

        // When
        val result = repository.getMediaById(999L).first()

        // Then
        assertNull(result)
    }

    @Test
    fun `getMediaById returns null when API returns null body`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(mapOf("id" to "123")) } returns Response.success(null)

        // When
        val result = repository.getMediaById(123L).first()

        // Then
        assertNull(result)
    }

    @Test
    fun `getMediaById returns null on API error`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(mapOf("id" to "123")) } returns Response.error(
            404,
            "Not found".toResponseBody()
        )

        // When
        val result = repository.getMediaById(123L).first()

        // Then
        assertNull(result)
    }

    @Test
    fun `getMediaById returns null on exception`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(mapOf("id" to "123")) } throws Exception("Network error")

        // When
        val result = repository.getMediaById(123L).first()

        // Then
        assertNull(result)
    }

    @Test
    fun `getMediaById calls API with correct ID parameter`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(any()) } returns Response.success(listOf(testMediaItem))

        // When
        repository.getMediaById(123L).first()

        // Then
        coVerify { mockApi.searchMedia(mapOf("id" to "123")) }
    }

    // updateWatchProgress Tests

    @Test
    fun `updateWatchProgress calls API successfully`() = runTest {
        // Given
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateWatchProgress(123L, 0.5)

        // Then
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to 0.5)) }
    }

    @Test
    fun `updateWatchProgress with zero progress`() = runTest {
        // Given
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateWatchProgress(123L, 0.0)

        // Then
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to 0.0)) }
    }

    @Test
    fun `updateWatchProgress with full progress`() = runTest {
        // Given
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateWatchProgress(123L, 1.0)

        // Then
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to 1.0)) }
    }

    @Test
    fun `updateWatchProgress handles API error gracefully`() = runTest {
        // Given
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.error(
            500,
            "Server error".toResponseBody()
        )

        // When
        repository.updateWatchProgress(123L, 0.5)

        // Then
        // Should not throw exception
        coVerify { mockApi.updateWatchProgress(123L, any()) }
    }

    @Test
    fun `updateWatchProgress handles exception gracefully`() = runTest {
        // Given
        coEvery { mockApi.updateWatchProgress(123L, any()) } throws Exception("Network error")

        // When
        repository.updateWatchProgress(123L, 0.5)

        // Then
        // Should not throw exception
        coVerify { mockApi.updateWatchProgress(123L, any()) }
    }

    @Test
    fun `updateWatchProgress with decimal progress value`() = runTest {
        // Given
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateWatchProgress(123L, 0.7532)

        // Then
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to 0.7532)) }
    }

    // updateFavoriteStatus Tests

    @Test
    fun `updateFavoriteStatus adds to favorites successfully`() = runTest {
        // Given
        coEvery { mockApi.updateFavoriteStatus(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateFavoriteStatus(123L, true)

        // Then
        coVerify { mockApi.updateFavoriteStatus(123L, mapOf("is_favorite" to true)) }
    }

    @Test
    fun `updateFavoriteStatus removes from favorites successfully`() = runTest {
        // Given
        coEvery { mockApi.updateFavoriteStatus(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateFavoriteStatus(123L, false)

        // Then
        coVerify { mockApi.updateFavoriteStatus(123L, mapOf("is_favorite" to false)) }
    }

    @Test
    fun `updateFavoriteStatus handles API error gracefully`() = runTest {
        // Given
        coEvery { mockApi.updateFavoriteStatus(123L, any()) } returns Response.error(
            500,
            "Server error".toResponseBody()
        )

        // When
        repository.updateFavoriteStatus(123L, true)

        // Then
        // Should not throw exception
        coVerify { mockApi.updateFavoriteStatus(123L, any()) }
    }

    @Test
    fun `updateFavoriteStatus handles exception gracefully`() = runTest {
        // Given
        coEvery { mockApi.updateFavoriteStatus(123L, any()) } throws Exception("Network error")

        // When
        repository.updateFavoriteStatus(123L, true)

        // Then
        // Should not throw exception
        coVerify { mockApi.updateFavoriteStatus(123L, any()) }
    }

    @Test
    fun `updateFavoriteStatus calls API with correct parameters`() = runTest {
        // Given
        coEvery { mockApi.updateFavoriteStatus(456L, any()) } returns Response.success(Unit)

        // When
        repository.updateFavoriteStatus(456L, true)

        // Then
        coVerify { mockApi.updateFavoriteStatus(456L, mapOf("is_favorite" to true)) }
    }

    // Integration-style Tests

    @Test
    fun `searchMedia preserves media item data integrity`() = runTest {
        // Given
        val detailedMediaItem = MediaItem(
            id = 789L,
            title = "Detailed Movie",
            mediaType = "movie",
            year = 2024,
            rating = 9.2,
            quality = "4k",
            directoryPath = "/movies/detailed.mkv",
            storageRootName = "premium_storage",
            fileSize = 25600000000L, // 25 GB
            duration = 10800, // 3 hours
            description = "A very detailed movie for testing",
            coverImage = "https://example.com/cover.jpg",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-15T10:00:00Z"
        )
        val searchRequest = MediaSearchRequest(query = "Detailed")
        coEvery { mockApi.searchMedia(any()) } returns Response.success(listOf(detailedMediaItem))

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        assertEquals(1, result.size)
        val item = result[0]
        assertEquals(789L, item.id)
        assertEquals("Detailed Movie", item.title)
        assertEquals("movie", item.mediaType)
        assertEquals(2024, item.year)
        assertEquals(9.2, item.rating)
        assertEquals("4k", item.quality)
        assertEquals("/movies/detailed.mkv", item.directoryPath)
        assertEquals("premium_storage", item.storageRootName)
        assertEquals(25600000000L, item.fileSize)
        assertEquals(10800, item.duration)
        assertEquals("A very detailed movie for testing", item.description)
        assertEquals("https://example.com/cover.jpg", item.coverImage)
    }

    @Test
    fun `multiple sequential operations work correctly`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(any()) } returns Response.success(testMediaList)
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)
        coEvery { mockApi.updateFavoriteStatus(123L, any()) } returns Response.success(Unit)

        // When
        val searchResult = repository.searchMedia(MediaSearchRequest(query = "Test")).first()
        repository.updateWatchProgress(123L, 0.5)
        repository.updateFavoriteStatus(123L, true)

        // Then
        assertEquals(3, searchResult.size)
        coVerify { mockApi.searchMedia(any()) }
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to 0.5)) }
        coVerify { mockApi.updateFavoriteStatus(123L, mapOf("is_favorite" to true)) }
    }

    @Test
    fun `repository handles different media types`() = runTest {
        // Given
        val movieRequest = MediaSearchRequest(mediaType = "movie")
        val tvRequest = MediaSearchRequest(mediaType = "tv_show")
        val musicRequest = MediaSearchRequest(mediaType = "music")

        coEvery { mockApi.searchMedia(match { it["media_type"] == "movie" }) } returns
            Response.success(listOf(testMediaItem.copy(mediaType = "movie")))
        coEvery { mockApi.searchMedia(match { it["media_type"] == "tv_show" }) } returns
            Response.success(listOf(testMediaItem.copy(mediaType = "tv_show")))
        coEvery { mockApi.searchMedia(match { it["media_type"] == "music" }) } returns
            Response.success(listOf(testMediaItem.copy(mediaType = "music")))

        // When
        val movieResult = repository.searchMedia(movieRequest).first()
        val tvResult = repository.searchMedia(tvRequest).first()
        val musicResult = repository.searchMedia(musicRequest).first()

        // Then
        assertEquals("movie", movieResult[0].mediaType)
        assertEquals("tv_show", tvResult[0].mediaType)
        assertEquals("music", musicResult[0].mediaType)
    }

    // Edge Case Tests

    @Test
    fun `searchMedia with offset parameter`() = runTest {
        // Given
        val searchRequest = MediaSearchRequest(offset = 100, limit = 20)
        coEvery { mockApi.searchMedia(any()) } returns Response.success(testMediaList)

        // When
        val result = repository.searchMedia(searchRequest).first()

        // Then
        coVerify {
            mockApi.searchMedia(match {
                it["offset"] == "100" && it["limit"] == "20"
            })
        }
    }

    @Test
    fun `updateWatchProgress with negative progress is accepted by repository`() = runTest {
        // Given
        // Note: Validation should happen at ViewModel/UI level, not repository
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateWatchProgress(123L, -0.5)

        // Then
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to -0.5)) }
    }

    @Test
    fun `updateWatchProgress with progress greater than 1 is accepted by repository`() = runTest {
        // Given
        // Note: Validation should happen at ViewModel/UI level, not repository
        coEvery { mockApi.updateWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        repository.updateWatchProgress(123L, 1.5)

        // Then
        coVerify { mockApi.updateWatchProgress(123L, mapOf("progress" to 1.5)) }
    }

    @Test
    fun `getMediaById with very large ID`() = runTest {
        // Given
        val largeId = Long.MAX_VALUE
        coEvery { mockApi.searchMedia(mapOf("id" to largeId.toString())) } returns
            Response.success(listOf(testMediaItem.copy(id = largeId)))

        // When
        val result = repository.getMediaById(largeId).first()

        // Then
        assertNotNull(result)
        assertEquals(largeId, result.id)
    }
}
