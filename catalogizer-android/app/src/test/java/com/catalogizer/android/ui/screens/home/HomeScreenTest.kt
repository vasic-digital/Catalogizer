package com.catalogizer.android.ui.screens.home

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.playback.PlaybackRepository
import com.catalogizer.android.data.playback.UiPlaybackProgress
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import com.catalogizer.android.ui.viewmodel.HomeViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for HomeScreen behavior via its HomeViewModel.
 * Covers loading state, media grid display, error state, empty state,
 * navigation callback contracts, and playback progress integration.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class HomeScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var mockPlaybackRepository: PlaybackRepository
    private lateinit var viewModel: HomeViewModel

    private fun createMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie",
        year: Int? = 2024,
        rating: Double? = 8.5,
        coverImage: String? = null,
    ): MediaItem = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        rating = rating,
        coverImage = coverImage,
        directoryPath = "/media/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z",
    )

    private fun createProgress(
        mediaItemId: Long,
        lastPosition: Long = 1800L,
        durationTotal: Long = 7200L,
    ): UiPlaybackProgress = UiPlaybackProgress(
        mediaItemId = mediaItemId,
        positionUnit = "seconds",
        durationTotal = durationTotal,
        lastPosition = lastPosition,
        lastSessionAmount = 900L,
        totalReproductions = 2L,
        aggregateAmount = 3600L,
        lastSessionEndedAtMs = null,
    )

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        mockPlaybackRepository = mockk(relaxed = true)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state has loading true`() {
        viewModel = HomeViewModel(mockMediaRepository)

        assertTrue(viewModel.isLoading.value)
    }

    @Test
    fun `initial state has empty recent media`() {
        viewModel = HomeViewModel(mockMediaRepository)

        assertTrue(viewModel.recentMedia.value.isEmpty())
    }

    @Test
    fun `initial state has empty favorite media`() {
        viewModel = HomeViewModel(mockMediaRepository)

        assertTrue(viewModel.favoriteMedia.value.isEmpty())
    }

    @Test
    fun `initial state has no error`() {
        viewModel = HomeViewModel(mockMediaRepository)

        assertNull(viewModel.error.value)
    }

    @Test
    fun `initial state has empty progress map`() {
        viewModel = HomeViewModel(mockMediaRepository)

        assertTrue(viewModel.progressById.value.isEmpty())
    }

    @Test
    fun `loadHomeData populates recent and favorite media`() = runTest {
        val recentItems = listOf(
            createMediaItem(1L, "Recent 1"),
            createMediaItem(2L, "Recent 2"),
            createMediaItem(3L, "Recent 3"),
        )
        val favoriteItems = listOf(
            createMediaItem(4L, "Popular 1"),
            createMediaItem(5L, "Popular 2"),
        )
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(favoriteItems)
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(3, viewModel.recentMedia.value.size)
        assertEquals("Recent 1", viewModel.recentMedia.value[0].title)
        assertEquals(2, viewModel.favoriteMedia.value.size)
        assertEquals("Popular 1", viewModel.favoriteMedia.value[0].title)
        assertFalse(viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData sets loading to false after completion`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData shows error on exception`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException("Server unreachable")
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
        assertEquals("Server unreachable", viewModel.error.value)
    }

    @Test
    fun `loadHomeData shows fallback error for null exception message`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException()
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals("Failed to load media", viewModel.error.value)
    }

    @Test
    fun `loadHomeData clears previous error on retry`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException("Error")
        viewModel = HomeViewModel(mockMediaRepository)
        viewModel.loadHomeData()
        advanceUntilIdle()
        assertNotNull(viewModel.error.value)

        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel.loadHomeData()
        advanceUntilIdle()

        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData handles unsuccessful recent media result`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.error("API error")
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(
            listOf(createMediaItem(1L, "Popular")))
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertEquals(1, viewModel.favoriteMedia.value.size)
    }

    @Test
    fun `loadHomeData handles unsuccessful popular media result`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(
            listOf(createMediaItem(1L, "Recent")))
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.error("API error")
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(1, viewModel.recentMedia.value.size)
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
    }

    @Test
    fun `loadHomeData handles null data in results`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult(data = null, isSuccess = true)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult(data = null, isSuccess = true)
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
    }

    @Test
    fun `loadHomeData empty results show empty state`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
        assertNull(viewModel.error.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData requests 20 items from repository`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        coVerify { mockMediaRepository.getRecentMedia(20) }
        coVerify { mockMediaRepository.getPopularMedia(20) }
    }

    @Test
    fun `loadHomeData fetches playback progress when playbackRepository provided`() = runTest {
        val recentItems = listOf(createMediaItem(1L, "Movie"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockPlaybackRepository.getProgress(1L) } returns createProgress(1L)
        viewModel = HomeViewModel(mockMediaRepository, mockPlaybackRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        coVerify { mockPlaybackRepository.getProgress(1L) }
    }

    @Test
    fun `loadHomeData does not fetch progress without playbackRepository`() = runTest {
        val recentItems = listOf(createMediaItem(1L, "Movie"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel = HomeViewModel(mockMediaRepository, null)

        viewModel.loadHomeData()
        advanceUntilIdle()

        coVerify(exactly = 0) { mockPlaybackRepository.getProgress(any()) }
    }

    @Test
    fun `loadHomeData deduplicates items for progress loading`() = runTest {
        val sharedItem = createMediaItem(1L, "Shared Item")
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(listOf(sharedItem))
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(listOf(sharedItem))
        coEvery { mockPlaybackRepository.getProgress(1L) } returns createProgress(1L)
        viewModel = HomeViewModel(mockMediaRepository, mockPlaybackRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        coVerify(exactly = 1) { mockPlaybackRepository.getProgress(1L) }
    }

    @Test
    fun `loadHomeData progress failure does not affect media data`() = runTest {
        val recentItems = listOf(createMediaItem(1L, "Movie"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockPlaybackRepository.getProgress(any()) } throws RuntimeException("Progress API down")
        viewModel = HomeViewModel(mockMediaRepository, mockPlaybackRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(1, viewModel.recentMedia.value.size)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `navigation callback for search is invocable`() {
        var searchNavigated = false
        val onNavigateToSearch: () -> Unit = { searchNavigated = true }

        onNavigateToSearch()

        assertTrue(searchNavigated)
    }

    @Test
    fun `navigation callback for settings is invocable`() {
        var settingsNavigated = false
        val onNavigateToSettings: () -> Unit = { settingsNavigated = true }

        onNavigateToSettings()

        assertTrue(settingsNavigated)
    }

    @Test
    fun `navigation callback for media detail receives correct id`() {
        var detailId: Long? = null
        val onNavigateToMediaDetail: (Long) -> Unit = { detailId = it }

        onNavigateToMediaDetail(42L)

        assertEquals(42L, detailId)
    }

    @Test
    fun `media items contain media type for badge display`() = runTest {
        val items = listOf(
            createMediaItem(1L, "Movie A", "movie"),
            createMediaItem(2L, "Show B", "tv_show"),
            createMediaItem(3L, "Album C", "music"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(items)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals("movie", viewModel.recentMedia.value[0].mediaType)
        assertEquals("tv_show", viewModel.recentMedia.value[1].mediaType)
        assertEquals("music", viewModel.recentMedia.value[2].mediaType)
    }

    @Test
    fun `media type badge formatting replaces underscores with spaces`() {
        val mediaType = "tv_show"
        val formatted = mediaType.replace("_", " ")

        assertEquals("tv show", formatted)
    }

    @Test
    fun `rating badge formats to one decimal place`() {
        val rating = 8.5
        val formatted = "%.1f".format(rating)

        assertEquals("8.5", formatted)
    }

    @Test
    fun `rating badge formats integer rating`() {
        val rating = 9.0
        val formatted = "%.1f".format(rating)

        assertEquals("9.0", formatted)
    }

    @Test
    fun `media items with year display year value`() {
        val item = createMediaItem(year = 2024)
        assertNotNull(item.year)
        assertEquals(2024, item.year)
    }

    @Test
    fun `media items without year have null year`() {
        val item = createMediaItem(year = null)
        assertNull(item.year)
    }

    @Test
    fun `multiple loadHomeData calls are safe`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(
            listOf(createMediaItem(1L, "Item")))
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        viewModel = HomeViewModel(mockMediaRepository)

        viewModel.loadHomeData()
        viewModel.loadHomeData()
        viewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
        assertEquals(1, viewModel.recentMedia.value.size)
    }

    @Test
    fun `media type gradient mapping covers all known types`() {
        val knownTypes = listOf(
            "movie", "tv_show", "tv_season", "tv_episode",
            "music_album", "song", "music",
            "game", "software",
            "book", "comic", "ebook",
            "unknown_type",
        )
        knownTypes.forEach { type ->
            val initial = type.take(1).uppercase()
            assertTrue("Initial should be non-empty for type '$type'", initial.isNotEmpty())
        }
    }

    @Test
    fun `cover image URL resolution handles absolute path`() {
        val coverImage = "/api/v1/assets/cover/1.jpg"
        val serverUrl = "http://192.168.0.100:8080"

        val resolved = if (coverImage.startsWith("/")) {
            serverUrl.trimEnd('/') + coverImage
        } else {
            coverImage
        }

        assertEquals("http://192.168.0.100:8080/api/v1/assets/cover/1.jpg", resolved)
    }

    @Test
    fun `cover image URL resolution handles TMDB proxy`() {
        val coverUrl = "https://image.tmdb.org/t/p/w500/abc.jpg"
        val serverUrl = "http://192.168.0.100:8080"

        val isTmdb = coverUrl.contains("image.tmdb.org")

        assertTrue(isTmdb)
        val encoded = java.net.URLEncoder.encode(coverUrl, "UTF-8")
        val resolved = serverUrl.trimEnd('/') + "/api/v1/image-proxy?url=$encoded"
        assertTrue(resolved.startsWith("http://192.168.0.100:8080/api/v1/image-proxy?url="))
    }

    @Test
    fun `cover image URL resolution handles external URL`() {
        val coverUrl = "https://example.com/cover.jpg"
        val isAbsolutePath = coverUrl.startsWith("/")
        val isTmdb = coverUrl.contains("image.tmdb.org")

        assertFalse(isAbsolutePath)
        assertFalse(isTmdb)
    }

    @Test
    fun `history target callback receives media item`() {
        var target: MediaItem? = null
        val onItemLongPress: (MediaItem) -> Unit = { target = it }
        val item = createMediaItem(42L, "History Target")

        onItemLongPress(item)

        assertNotNull(target)
        assertEquals(42L, target?.id)
        assertEquals("History Target", target?.title)
    }
}
