package com.catalogizer.android.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.repository.MediaRepository
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.MainDispatcherRule
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.Assert.*

/**
 * ViewModel tests for HomeViewModel
 */
@ExperimentalCoroutinesApi
class HomeViewModelTest {

    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: HomeViewModel

    private fun createMediaItem(
        id: Long,
        title: String,
        mediaType: String = "movie"
    ): MediaItem = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        directoryPath = "/media/$id",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z"
    )

    @Before
    fun setup() {
        Dispatchers.setMain(StandardTestDispatcher())
        mockMediaRepository = mockk(relaxed = true)
        viewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
        clearAllMocks()
    }

    @Test
    fun `initial state should have empty lists and loading true`() = runTest {
        // Then - check initial state before any data loading
        assertEquals(emptyList<MediaItem>(), viewModel.recentMedia.value)
        assertEquals(emptyList<MediaItem>(), viewModel.favoriteMedia.value)
        assertTrue(viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `loadHomeData populates recentMedia and favoriteMedia`() = runTest {
        // Given
        val recentItems = listOf(
            createMediaItem(1, "Recent Movie 1"),
            createMediaItem(2, "Recent Movie 2")
        )
        val favoriteItems = listOf(
            createMediaItem(3, "Popular Movie 1"),
            createMediaItem(4, "Popular Movie 2"),
            createMediaItem(5, "Popular Movie 3")
        )
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(favoriteItems)

        // When
        viewModel.loadHomeData()
        advanceUntilIdle()

        // Then
        assertEquals(2, viewModel.recentMedia.value.size)
        assertEquals("Recent Movie 1", viewModel.recentMedia.value[0].title)
        assertEquals("Recent Movie 2", viewModel.recentMedia.value[1].title)

        assertEquals(3, viewModel.favoriteMedia.value.size)
        assertEquals("Popular Movie 1", viewModel.favoriteMedia.value[0].title)

        assertFalse(viewModel.isLoading.value)
        assertNull(viewModel.error.value)

        // Verify repository calls
        coVerify { mockMediaRepository.getRecentMedia(20) }
        coVerify { mockMediaRepository.getPopularMedia(20) }
    }

    @Test
    fun `loadHomeData handles errors`() = runTest {
        // Given
        coEvery { mockMediaRepository.getRecentMedia(20) } throws RuntimeException("Network unavailable")

        // When
        viewModel.loadHomeData()
        advanceUntilIdle()

        // Then
        assertNotNull(viewModel.error.value)
        assertEquals("Network unavailable", viewModel.error.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loading state transitions correctly during loadHomeData`() = runTest {
        // Given
        val recentItems = listOf(createMediaItem(1, "Movie"))
        val favoriteItems = listOf(createMediaItem(2, "Popular"))
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(favoriteItems)

        // Initial state
        assertTrue(viewModel.isLoading.value)

        // When
        viewModel.loadHomeData()
        advanceUntilIdle()

        // Then - after completion, loading should be false
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData clears previous error on new load`() = runTest {
        // Given - first load fails
        coEvery { mockMediaRepository.getRecentMedia(20) } throws RuntimeException("First error")
        viewModel.loadHomeData()
        advanceUntilIdle()
        assertNotNull(viewModel.error.value)

        // Given - second load succeeds
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(emptyList())
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.success(emptyList())

        // When
        viewModel.loadHomeData()
        advanceUntilIdle()

        // Then - error should be cleared
        assertNull(viewModel.error.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `loadHomeData handles partial success when only recent media succeeds`() = runTest {
        // Given - recent succeeds but popular returns error result
        val recentItems = listOf(createMediaItem(1, "Recent Movie"))
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult.success(recentItems)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult.error("Server error")

        // When
        viewModel.loadHomeData()
        advanceUntilIdle()

        // Then - recent should be populated, favorites should remain empty
        assertEquals(1, viewModel.recentMedia.value.size)
        assertEquals(emptyList<MediaItem>(), viewModel.favoriteMedia.value)
        assertFalse(viewModel.isLoading.value)
        assertNull(viewModel.error.value) // no exception thrown, just unsuccessful result
    }

    @Test
    fun `loadHomeData handles null data in successful result`() = runTest {
        // Given - results are successful but data is null
        coEvery { mockMediaRepository.getRecentMedia(20) } returns ApiResult(data = null, error = null, isSuccess = false)
        coEvery { mockMediaRepository.getPopularMedia(20) } returns ApiResult(data = null, error = null, isSuccess = false)

        // When
        viewModel.loadHomeData()
        advanceUntilIdle()

        // Then - lists should remain empty
        assertEquals(emptyList<MediaItem>(), viewModel.recentMedia.value)
        assertEquals(emptyList<MediaItem>(), viewModel.favoriteMedia.value)
        assertFalse(viewModel.isLoading.value)
    }
}
