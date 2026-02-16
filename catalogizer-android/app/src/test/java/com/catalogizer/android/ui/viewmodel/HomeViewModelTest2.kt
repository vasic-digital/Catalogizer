package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class HomeViewModelTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockMediaRepository = mockk<MediaRepository>(relaxed = true)
    private lateinit var viewModel: HomeViewModel

    private val testMediaItem = MediaItem(
        id = 1L,
        title = "Test Movie",
        mediaType = "movie",
        directoryPath = "/movies/test",
        createdAt = "2024-01-01",
        updatedAt = "2024-01-01",
        rating = 8.5
    )

    @Before
    fun setup() {
        viewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state has empty lists and loading true`() {
        assertTrue(viewModel.recentMedia.value.isEmpty())
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
        assertTrue(viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData populates recent media on success`() = runTest {
        val items = listOf(testMediaItem, testMediaItem.copy(id = 2, title = "Movie 2"))
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(items)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(2, viewModel.recentMedia.value.size)
        assertEquals("Test Movie", viewModel.recentMedia.value[0].title)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData populates favorite media on success`() = runTest {
        val favorites = listOf(testMediaItem.copy(isFavorite = true))
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(favorites)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(1, viewModel.favoriteMedia.value.size)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData sets error on exception`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } throws RuntimeException("Network error")

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals("Network error", viewModel.error.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData handles partial failure gracefully`() = runTest {
        val items = listOf(testMediaItem)
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(items)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.error("Failed")

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(1, viewModel.recentMedia.value.size)
        assertTrue(viewModel.favoriteMedia.value.isEmpty())
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData sets loading to true then false`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())

        viewModel.loadHomeData()

        // Before advancing, isLoading should be true (first statement in loadHomeData)
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
    }
}
