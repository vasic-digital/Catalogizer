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
        title: String = "Test",
        mediaType: String = "movie",
        isFavorite: Boolean = false,
        watchProgress: Double = 0.0
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            directoryPath = "/media/test",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            isFavorite = isFavorite,
            watchProgress = watchProgress
        )
    }

    @Before
    fun setup() {
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        coEvery { mockMediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(emptyList())
        coEvery { mockMediaRepository.getSimilarItems(any()) } returns emptyList()
        coEvery { mockMediaRepository.getTrendingItems(any()) } returns emptyList()
        coEvery { mockMediaRepository.getEntityStats() } returns Pair(0, emptyMap())
        viewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state has empty fields`() {
        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertTrue(state.continueWatching.isEmpty())
        assertTrue(state.recentMovies.isEmpty())
        assertTrue(state.topRatedMovies.isEmpty())
        assertTrue(state.recommended.isEmpty())
        assertTrue(state.trending.isEmpty())
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
        coEvery { mockMediaRepository.toggleFavorite(any(), any(), any()) } returns true

        viewModel.toggleFavorite(42L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.toggleFavorite(any(), eq(42L), any()) }
    }
}
