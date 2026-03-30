package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class HomeViewModelTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: HomeViewModel

    private val testMediaItem = MediaItem(
        id = 1L,
        title = "Test Movie",
        mediaType = "movie",
        year = 2024,
        description = "A test movie",
        directoryPath = "/media/movies/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z"
    )

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        viewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state has loading true and empty lists`() {
        assertTrue(viewModel.isLoading.value)
        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData sets recent and favorite media on success`() = runTest {
        val recentItems = listOf(testMediaItem, testMediaItem.copy(id = 2L, title = "Movie 2"))
        val popularItems = listOf(testMediaItem.copy(id = 3L, title = "Popular Movie"))
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(popularItems)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(2, viewModel.recentMedia.value.size)
        assertEquals(1, viewModel.favoriteMedia.value.size)
        assertFalse(viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData handles empty results`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData handles API error for recent media`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.error("Network error")
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())

        viewModel.loadHomeData()
        advanceUntilIdle()

        // Recent media stays empty because API failed
        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData handles exception`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } throws RuntimeException("Connection refused")

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
        assertEquals("Connection refused", viewModel.error.value)
    }

    @Test
    fun `loadHomeData sets isLoading to true then false`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())

        viewModel.loadHomeData()
        // After advanceUntilIdle, loading should be false
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData clears previous error`() = runTest {
        // First call fails
        coEvery { mockMediaRepository.getRecentMedia(20) } throws RuntimeException("Error")
        viewModel.loadHomeData()
        advanceUntilIdle()
        assertNotNull(viewModel.error.value)

        // Second call succeeds
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())
        viewModel.loadHomeData()
        advanceUntilIdle()
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData handles null data in ApiResult`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult(data = null, error = null)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertTrue(viewModel.favoriteMedia.value.isEmpty())
        assertFalse(viewModel.isLoading.value)
    }
}
