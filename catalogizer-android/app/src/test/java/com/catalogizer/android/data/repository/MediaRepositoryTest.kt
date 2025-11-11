package com.catalogizer.android.data.repository

import androidx.paging.PagingSource
import com.catalogizer.android.data.local.MediaDao
import com.catalogizer.android.data.models.*
import com.catalogizer.android.data.remote.CatalogizerApi
import com.catalogizer.android.data.remote.ApiResult
import io.mockk.*
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Before
import org.junit.Test
import retrofit2.Response
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class MediaRepositoryTest {

    private lateinit var repository: MediaRepository
    private lateinit var mockApi: CatalogizerApi
    private lateinit var mockMediaDao: MediaDao

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
        isFavorite = false,
        watchProgress = 0.0,
        lastWatched = null,
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z"
    )

    private val testMediaStats = MediaStats(
        totalItems = 150,
        byType = mapOf("movie" to 80, "tv_show" to 50, "music" to 20),
        byQuality = mapOf("1080p" to 100, "720p" to 40, "4k" to 10),
        totalSize = 1073741824000L,
        recentAdditions = 25
    )

    @Before
    fun setup() {
        mockApi = mockk()
        mockMediaDao = mockk()
        repository = MediaRepository(mockApi, mockMediaDao)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // getMediaById Tests

    @Test
    fun `getMediaById returns local data when not forcing refresh`() = runTest {
        // Given
        coEvery { mockMediaDao.getMediaById(123L) } returns testMediaItem

        // When
        val result = repository.getMediaById(123L, forceRefresh = false)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(testMediaItem, result.data)
        coVerify(exactly = 0) { mockApi.getMediaById(any()) }
    }

    @Test
    fun `getMediaById fetches from API when local data not found`() = runTest {
        // Given
        coEvery { mockMediaDao.getMediaById(123L) } returns null
        coEvery { mockApi.getMediaById(123L) } returns Response.success(testMediaItem)
        coEvery { mockMediaDao.insertMedia(testMediaItem) } just Runs

        // When
        val result = repository.getMediaById(123L, forceRefresh = false)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(testMediaItem, result.data)
        coVerify { mockApi.getMediaById(123L) }
        coVerify { mockMediaDao.insertMedia(testMediaItem) }
    }

    @Test
    fun `getMediaById forces refresh when flag is true`() = runTest {
        // Given
        coEvery { mockApi.getMediaById(123L) } returns Response.success(testMediaItem)
        coEvery { mockMediaDao.insertMedia(testMediaItem) } just Runs

        // When
        val result = repository.getMediaById(123L, forceRefresh = true)

        // Then
        assertTrue(result.isSuccess)
        coVerify(exactly = 0) { mockMediaDao.getMediaById(any()) }
        coVerify { mockApi.getMediaById(123L) }
    }

    @Test
    fun `getMediaById handles API error`() = runTest {
        // Given
        coEvery { mockMediaDao.getMediaById(123L) } returns null
        coEvery { mockApi.getMediaById(123L) } returns Response.error(404, mockk(relaxed = true))

        // When
        val result = repository.getMediaById(123L)

        // Then
        assertFalse(result.isSuccess)
        assertNotNull(result.error)
    }

    @Test
    fun `getMediaById handles exception`() = runTest {
        // Given
        coEvery { mockMediaDao.getMediaById(123L) } returns null
        coEvery { mockApi.getMediaById(123L) } throws Exception("Network error")

        // When
        val result = repository.getMediaById(123L)

        // Then
        assertFalse(result.isSuccess)
        assertEquals("Network error", result.error)
    }

    // getMediaStats Tests

    @Test
    fun `getMediaStats returns stats successfully`() = runTest {
        // Given
        coEvery { mockApi.getMediaStats() } returns Response.success(testMediaStats)

        // When
        val result = repository.getMediaStats()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(testMediaStats, result.data)
        assertEquals(150, result.data?.totalItems)
    }

    @Test
    fun `getMediaStats handles API error`() = runTest {
        // Given
        coEvery { mockApi.getMediaStats() } returns Response.error(500, mockk(relaxed = true))

        // When
        val result = repository.getMediaStats()

        // Then
        assertFalse(result.isSuccess)
    }

    // getRecentMedia Tests

    @Test
    fun `getRecentMedia returns API data and caches locally`() = runTest {
        // Given
        val recentItems = listOf(testMediaItem)
        coEvery { mockApi.getRecentMedia(10) } returns Response.success(recentItems)
        coEvery { mockMediaDao.insertAllMedia(recentItems) } just Runs

        // When
        val result = repository.getRecentMedia(10)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(recentItems, result.data)
        coVerify { mockMediaDao.insertAllMedia(recentItems) }
    }

    @Test
    fun `getRecentMedia falls back to local data on API error`() = runTest {
        // Given
        val localItems = listOf(testMediaItem)
        coEvery { mockApi.getRecentMedia(10) } throws Exception("Network error")
        coEvery { mockMediaDao.getRecentlyAdded(10) } returns flowOf(localItems)

        // When
        val result = repository.getRecentMedia(10)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(localItems, result.data)
    }

    // getPopularMedia Tests

    @Test
    fun `getPopularMedia returns API data and caches locally`() = runTest {
        // Given
        val popularItems = listOf(testMediaItem)
        coEvery { mockApi.getPopularMedia(10) } returns Response.success(popularItems)
        coEvery { mockMediaDao.insertAllMedia(popularItems) } just Runs

        // When
        val result = repository.getPopularMedia(10)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(popularItems, result.data)
        coVerify { mockMediaDao.insertAllMedia(popularItems) }
    }

    @Test
    fun `getPopularMedia falls back to local data on API error`() = runTest {
        // Given
        val localItems = listOf(testMediaItem)
        coEvery { mockApi.getPopularMedia(10) } throws Exception("Network error")
        coEvery { mockMediaDao.getTopRated(10) } returns flowOf(localItems)

        // When
        val result = repository.getPopularMedia(10)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(localItems, result.data)
    }

    // toggleFavorite Tests

    @Test
    fun `toggleFavorite adds to favorites successfully`() = runTest {
        // Given
        val nonFavoriteMedia = testMediaItem.copy(isFavorite = false)
        coEvery { mockMediaDao.getMediaById(123L) } returns nonFavoriteMedia
        coEvery { mockMediaDao.updateFavoriteStatus(123L, true) } just Runs
        coEvery { mockApi.addToFavorites(123L) } returns Response.success(Unit)

        // When
        val result = repository.toggleFavorite(123L)

        // Then
        assertTrue(result.isSuccess)
        coVerify { mockMediaDao.updateFavoriteStatus(123L, true) }
        coVerify { mockApi.addToFavorites(123L) }
    }

    @Test
    fun `toggleFavorite removes from favorites successfully`() = runTest {
        // Given
        val favoriteMedia = testMediaItem.copy(isFavorite = true)
        coEvery { mockMediaDao.getMediaById(123L) } returns favoriteMedia
        coEvery { mockMediaDao.updateFavoriteStatus(123L, false) } just Runs
        coEvery { mockApi.removeFromFavorites(123L) } returns Response.success(Unit)

        // When
        val result = repository.toggleFavorite(123L)

        // Then
        assertTrue(result.isSuccess)
        coVerify { mockMediaDao.updateFavoriteStatus(123L, false) }
        coVerify { mockApi.removeFromFavorites(123L) }
    }

    @Test
    fun `toggleFavorite reverts local state on API failure`() = runTest {
        // Given
        val nonFavoriteMedia = testMediaItem.copy(isFavorite = false)
        coEvery { mockMediaDao.getMediaById(123L) } returns nonFavoriteMedia
        coEvery { mockMediaDao.updateFavoriteStatus(any(), any()) } just Runs
        coEvery { mockApi.addToFavorites(123L) } returns Response.error(500, mockk(relaxed = true))

        // When
        val result = repository.toggleFavorite(123L)

        // Then
        assertFalse(result.isSuccess)
        // Verify state was set to true, then reverted to false
        coVerify(exactly = 2) { mockMediaDao.updateFavoriteStatus(123L, any()) }
    }

    @Test
    fun `toggleFavorite handles exception`() = runTest {
        // Given
        coEvery { mockMediaDao.getMediaById(123L) } throws Exception("Database error")

        // When
        val result = repository.toggleFavorite(123L)

        // Then
        assertFalse(result.isSuccess)
        assertEquals("Database error", result.error)
    }

    // updateWatchProgress Tests

    @Test
    fun `updateWatchProgress updates local and syncs with server`() = runTest {
        // Given
        coEvery { mockMediaDao.updateWatchProgress(123L, 0.5, any()) } just Runs
        coEvery { mockApi.updateUserWatchProgress(123L, any()) } returns Response.success(Unit)

        // When
        val result = repository.updateWatchProgress(123L, 0.5, 3600L)

        // Then
        assertTrue(result.isSuccess)
        coVerify { mockMediaDao.updateWatchProgress(123L, 0.5, any()) }
        coVerify { mockApi.updateUserWatchProgress(123L, any()) }
    }

    @Test
    fun `updateWatchProgress handles API error`() = runTest {
        // Given
        coEvery { mockMediaDao.updateWatchProgress(any(), any(), any()) } just Runs
        coEvery { mockApi.updateUserWatchProgress(any(), any()) } returns Response.error(500, mockk(relaxed = true))

        // When
        val result = repository.updateWatchProgress(123L, 0.5)

        // Then
        assertFalse(result.isSuccess)
        // Local update should still happen
        coVerify { mockMediaDao.updateWatchProgress(123L, 0.5, any()) }
    }

    @Test
    fun `updateWatchProgress handles exception`() = runTest {
        // Given
        coEvery { mockMediaDao.updateWatchProgress(any(), any(), any()) } throws Exception("Database error")

        // When
        val result = repository.updateWatchProgress(123L, 0.5)

        // Then
        assertFalse(result.isSuccess)
        assertEquals("Database error", result.error)
    }

    // refreshMetadata Tests

    @Test
    fun `refreshMetadata returns metadata successfully`() = runTest {
        // Given
        val metadata = mapOf("title" to "Updated Title", "year" to "2024")
        coEvery { mockApi.refreshMetadata(123L) } returns Response.success(metadata)

        // When
        val result = repository.refreshMetadata(123L)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(metadata, result.data)
    }

    // getExternalMetadata Tests

    @Test
    fun `getExternalMetadata returns external data successfully`() = runTest {
        // Given
        val externalMetadata = listOf(
            ExternalMetadata(
                provider = "tmdb",
                title = "Test Movie",
                description = "A test movie",
                posterUrl = "https://example.com/poster.jpg",
                backdropUrl = "https://example.com/backdrop.jpg",
                genres = listOf("Action", "Adventure"),
                cast = listOf("Actor 1", "Actor 2")
            )
        )
        coEvery { mockApi.getExternalMetadata(123L) } returns Response.success(externalMetadata)

        // When
        val result = repository.getExternalMetadata(123L)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(externalMetadata, result.data)
    }

    // refreshAllMedia Tests

    @Test
    fun `refreshAllMedia syncs all media successfully`() = runTest {
        // Given
        val searchResponse = MediaSearchResponse(
            items = listOf(testMediaItem),
            total = 1,
            limit = 1000,
            offset = 0
        )
        coEvery { mockApi.searchMedia(limit = 1000) } returns Response.success(searchResponse)
        coEvery { mockMediaDao.refreshMedia(any()) } just Runs

        // When
        val result = repository.refreshAllMedia()

        // Then
        assertTrue(result.isSuccess)
        coVerify { mockMediaDao.refreshMedia(searchResponse.items) }
    }

    @Test
    fun `refreshAllMedia handles API error`() = runTest {
        // Given
        coEvery { mockApi.searchMedia(limit = 1000) } returns Response.error(500, mockk(relaxed = true))

        // When
        val result = repository.refreshAllMedia()

        // Then
        assertFalse(result.isSuccess)
    }

    // clearCache Tests

    @Test
    fun `clearCache deletes all media`() = runTest {
        // Given
        coEvery { mockMediaDao.deleteAllMedia() } just Runs

        // When
        repository.clearCache()

        // Then
        coVerify { mockMediaDao.deleteAllMedia() }
    }

    // getStreamUrl Tests

    @Test
    fun `getStreamUrl returns URL successfully`() = runTest {
        // Given
        val urlMap = mapOf("url" to "https://stream.example.com/video.m3u8")
        coEvery { mockApi.getStreamUrl(123L) } returns Response.success(urlMap)

        // When
        val result = repository.getStreamUrl(123L)

        // Then
        assertTrue(result.isSuccess)
        assertEquals("https://stream.example.com/video.m3u8", result.data)
    }

    @Test
    fun `getStreamUrl handles stream_url key`() = runTest {
        // Given
        val urlMap = mapOf("stream_url" to "https://stream.example.com/video.m3u8")
        coEvery { mockApi.getStreamUrl(123L) } returns Response.success(urlMap)

        // When
        val result = repository.getStreamUrl(123L)

        // Then
        assertTrue(result.isSuccess)
        assertEquals("https://stream.example.com/video.m3u8", result.data)
    }

    @Test
    fun `getStreamUrl handles missing URL in response`() = runTest {
        // Given
        val urlMap = mapOf("other_field" to "value")
        coEvery { mockApi.getStreamUrl(123L) } returns Response.success(urlMap)

        // When
        val result = repository.getStreamUrl(123L)

        // Then
        assertFalse(result.isSuccess)
        assertEquals("Stream URL not found in response", result.error)
    }

    // getDownloadUrl Tests

    @Test
    fun `getDownloadUrl returns URL successfully`() = runTest {
        // Given
        val urlMap = mapOf("url" to "https://download.example.com/file.mp4")
        coEvery { mockApi.getDownloadUrl(123L) } returns Response.success(urlMap)

        // When
        val result = repository.getDownloadUrl(123L)

        // Then
        assertTrue(result.isSuccess)
        assertEquals("https://download.example.com/file.mp4", result.data)
    }

    @Test
    fun `getDownloadUrl handles download_url key`() = runTest {
        // Given
        val urlMap = mapOf("download_url" to "https://download.example.com/file.mp4")
        coEvery { mockApi.getDownloadUrl(123L) } returns Response.success(urlMap)

        // When
        val result = repository.getDownloadUrl(123L)

        // Then
        assertTrue(result.isSuccess)
        assertEquals("https://download.example.com/file.mp4", result.data)
    }

    // Offline Support Tests

    @Test
    fun `getAllMediaTypes returns flow of media types`() = runTest {
        // Given
        val mediaTypes = listOf("movie", "tv_show", "music")
        every { mockMediaDao.getAllMediaTypes() } returns flowOf(mediaTypes)

        // When
        val result = repository.getAllMediaTypes().first()

        // Then
        assertEquals(mediaTypes, result)
    }

    @Test
    fun `getTotalCount returns flow of count`() = runTest {
        // Given
        every { mockMediaDao.getTotalCount() } returns flowOf(150)

        // When
        val result = repository.getTotalCount().first()

        // Then
        assertEquals(150, result)
    }

    @Test
    fun `getCountByType returns flow of type count`() = runTest {
        // Given
        every { mockMediaDao.getCountByType("movie") } returns flowOf(80)

        // When
        val result = repository.getCountByType("movie").first()

        // Then
        assertEquals(80, result)
    }

    // getContinueWatching Tests

    @Test
    fun `getContinueWatching returns API data and updates local`() = runTest {
        // Given
        val continueWatchingItems = listOf(
            testMediaItem.copy(watchProgress = 0.5, lastWatched = "2024-01-15T10:00:00Z")
        )
        coEvery { mockApi.getContinueWatching() } returns Response.success(continueWatchingItems)
        coEvery { mockMediaDao.updateWatchProgress(any(), any(), any()) } just Runs

        // When
        val result = repository.getContinueWatching()

        // Then
        assertTrue(result.isSuccess)
        assertEquals(continueWatchingItems, result.data)
        coVerify { mockMediaDao.updateWatchProgress(123L, 0.5, "2024-01-15T10:00:00Z") }
    }

    @Test
    fun `getContinueWatching falls back to local on API error`() = runTest {
        // Given
        coEvery { mockApi.getContinueWatching() } throws Exception("Network error")
        every { mockMediaDao.getContinueWatchingPaging() } returns mockk()

        // When
        val result = repository.getContinueWatching()

        // Then
        assertFalse(result.isSuccess)
        assertEquals("Network error", result.error)
    }

    // getMediaByIdFlow Tests

    @Test
    fun `getMediaByIdFlow returns flow of media item`() = runTest {
        // Given
        every { mockMediaDao.getMediaByIdFlow(123L) } returns flowOf(testMediaItem)

        // When
        val result = repository.getMediaByIdFlow(123L).first()

        // Then
        assertEquals(testMediaItem, result)
    }

    @Test
    fun `getMediaByIdFlow returns null for non-existent media`() = runTest {
        // Given
        every { mockMediaDao.getMediaByIdFlow(999L) } returns flowOf(null)

        // When
        val result = repository.getMediaByIdFlow(999L).first()

        // Then
        assertNull(result)
    }

    // Paging Tests

    @Test
    fun `getMediaPaging returns paging flow`() {
        // Given
        val searchRequest = MediaSearchRequest(query = "test")

        // When
        val result = repository.getMediaPaging(searchRequest)

        // Then
        assertNotNull(result)
    }

    @Test
    fun `getMediaByTypePaging returns paging flow for type`() {
        // Given
        val mockPagingSource = mockk<PagingSource<Int, MediaItem>>()
        every { mockMediaDao.getMediaByTypePaging("movie") } returns mockPagingSource

        // When
        val result = repository.getMediaByTypePaging("movie")

        // Then
        assertNotNull(result)
    }

    @Test
    fun `searchMediaPaging returns paging flow for query`() {
        // Given
        val mockPagingSource = mockk<PagingSource<Int, MediaItem>>()
        every { mockMediaDao.searchMediaPaging("test") } returns mockPagingSource

        // When
        val result = repository.searchMediaPaging("test")

        // Then
        assertNotNull(result)
    }

    @Test
    fun `getFavoritesPaging returns paging flow`() {
        // Given
        val mockPagingSource = mockk<PagingSource<Int, MediaItem>>()
        every { mockMediaDao.getFavoritesPaging() } returns mockPagingSource

        // When
        val result = repository.getFavoritesPaging()

        // Then
        assertNotNull(result)
    }

    @Test
    fun `getContinueWatchingPaging returns paging flow`() {
        // Given
        val mockPagingSource = mockk<PagingSource<Int, MediaItem>>()
        every { mockMediaDao.getContinueWatchingPaging() } returns mockPagingSource

        // When
        val result = repository.getContinueWatchingPaging()

        // Then
        assertNotNull(result)
    }
}
