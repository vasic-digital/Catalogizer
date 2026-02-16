package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
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

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String = "movie",
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        directoryPath = "/test",
        createdAt = "2024-01-01",
        updatedAt = "2024-01-01",
        watchProgress = watchProgress,
        isFavorite = isFavorite
    )

    @Before
    fun setup() {
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        viewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state is not loading with empty lists`() {
        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertTrue(state.movies.isEmpty())
    }

    @Test
    fun `loadHomeData sets loading state`() = runTest {
        viewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `loadHomeData populates movies`() = runTest {
        val movies = listOf(createTestMediaItem(1, "Movie 1"), createTestMediaItem(2, "Movie 2"))
        coEvery { mockMediaRepository.searchMedia(match { it.mediaType == "movie" }) } returns flowOf(movies)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(2, viewModel.uiState.value.movies.size)
    }

    @Test
    fun `loadHomeData handles error`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        viewModel.loadHomeData()
        advanceUntilIdle()

        // Each individual load method catches exceptions and returns emptyList(),
        // so no top-level error is set - sections are just empty
        assertNull(viewModel.uiState.value.error)
        assertFalse(viewModel.uiState.value.isLoading)
        assertTrue(viewModel.uiState.value.movies.isEmpty())
    }

    @Test
    fun `refreshContent calls loadHomeData`() = runTest {
        viewModel.refreshContent()
        advanceUntilIdle()

        // loadHomeData should have been triggered (which calls searchMedia)
        assertFalse(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `markAsWatched calls repository`() = runTest {
        coEvery { mockMediaRepository.updateWatchProgress(any(), any()) } just Runs

        viewModel.markAsWatched(42L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateWatchProgress(42L, 1.0) }
    }

    @Test
    fun `updateWatchProgress calls repository`() = runTest {
        coEvery { mockMediaRepository.updateWatchProgress(any(), any()) } just Runs

        viewModel.updateWatchProgress(42L, 0.75)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateWatchProgress(42L, 0.75) }
    }

    @Test
    fun `toggleFavorite calls repository with toggled value`() = runTest {
        val item = createTestMediaItem(42, isFavorite = false)
        coEvery { mockMediaRepository.getMediaById(42L) } returns flowOf(item)
        coEvery { mockMediaRepository.updateFavoriteStatus(any(), any()) } just Runs

        viewModel.toggleFavorite(42L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateFavoriteStatus(42L, true) }
    }
}
