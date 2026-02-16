package com.catalogizer.android.ui.screens

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
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

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class HomeScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var homeViewModel: HomeViewModel

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie"
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            directoryPath = "/media/test",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z"
        )
    }

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        homeViewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state should have loading set to true`() {
        val isLoading = homeViewModel.isLoading.value
        assertTrue(isLoading)
    }

    @Test
    fun `initial state should have empty recent media`() {
        val recentMedia = homeViewModel.recentMedia.value
        assertTrue(recentMedia.isEmpty())
    }

    @Test
    fun `initial state should have empty favorite media`() {
        val favoriteMedia = homeViewModel.favoriteMedia.value
        assertTrue(favoriteMedia.isEmpty())
    }

    @Test
    fun `initial state should have no error`() {
        val error = homeViewModel.error.value
        assertNull(error)
    }

    @Test
    fun `loadHomeData should set loading to true then false`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(homeViewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData should populate recentMedia on success`() = runTest {
        val testMedia = listOf(
            createTestMediaItem(1L, "Movie 1"),
            createTestMediaItem(2L, "Movie 2"),
            createTestMediaItem(3L, "Movie 3")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(testMedia)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val recentMedia = homeViewModel.recentMedia.value
        assertEquals(3, recentMedia.size)
        assertEquals("Movie 1", recentMedia[0].title)
        assertEquals("Movie 2", recentMedia[1].title)
        assertEquals("Movie 3", recentMedia[2].title)
    }

    @Test
    fun `loadHomeData should populate favoriteMedia on success`() = runTest {
        val testFavorites = listOf(
            createTestMediaItem(1L, "Favorite 1"),
            createTestMediaItem(2L, "Favorite 2")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(testFavorites)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val favoriteMedia = homeViewModel.favoriteMedia.value
        assertEquals(2, favoriteMedia.size)
        assertEquals("Favorite 1", favoriteMedia[0].title)
    }

    @Test
    fun `loadHomeData should request 20 recent items`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        coVerify { mockMediaRepository.getRecentMedia(20) }
    }

    @Test
    fun `loadHomeData should request 20 popular items`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        coVerify { mockMediaRepository.getPopularMedia(20) }
    }

    @Test
    fun `loadHomeData should set error on exception`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException("Network error")

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(homeViewModel.isLoading.value)
        assertEquals("Network error", homeViewModel.error.value)
    }

    @Test
    fun `loadHomeData should clear error before loading`() = runTest {
        // First call fails
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException("Error")
        homeViewModel.loadHomeData()
        advanceUntilIdle()
        assertNotNull(homeViewModel.error.value)

        // Second call succeeds
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())
        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertNull(homeViewModel.error.value)
    }

    @Test
    fun `loadHomeData should handle unsuccessful recent media result`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.error("API error")
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        // recentMedia should remain empty when API returns error
        assertTrue(homeViewModel.recentMedia.value.isEmpty())
        assertFalse(homeViewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData should handle unsuccessful popular media result`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.error("API error")

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        // favoriteMedia should remain empty when API returns error
        assertTrue(homeViewModel.favoriteMedia.value.isEmpty())
        assertFalse(homeViewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData should handle null data in recent result`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult(data = null, isSuccess = true)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(emptyList())

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(homeViewModel.recentMedia.value.isEmpty())
    }

    @Test
    fun `loadHomeData should handle null data in popular result`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult(data = null, isSuccess = true)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(homeViewModel.favoriteMedia.value.isEmpty())
    }

    @Test
    fun `loadHomeData with generic exception message should show fallback`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException()

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals("Failed to load media", homeViewModel.error.value)
    }

    @Test
    fun `loadHomeData should handle both recent and popular data simultaneously`() = runTest {
        val recentItems = listOf(createTestMediaItem(1L, "Recent"))
        val popularItems = listOf(createTestMediaItem(2L, "Popular"))

        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(any()) } returns ApiResult.success(popularItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(1, homeViewModel.recentMedia.value.size)
        assertEquals("Recent", homeViewModel.recentMedia.value[0].title)
        assertEquals(1, homeViewModel.favoriteMedia.value.size)
        assertEquals("Popular", homeViewModel.favoriteMedia.value[0].title)
    }
}
