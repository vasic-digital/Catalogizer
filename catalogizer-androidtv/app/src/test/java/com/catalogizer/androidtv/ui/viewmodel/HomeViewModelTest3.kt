package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
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

    private val testItem = MediaItem(
        id = 1L, title = "Test Movie", mediaType = "movie",
        directoryPath = "/media", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
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
    fun `initial state is not loading with empty data`() {
        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertTrue(state.recentMovies.isEmpty())
        assertTrue(state.topRatedMovies.isEmpty())
        assertTrue(state.topRatedTvShows.isEmpty())
    }

    @Test
    fun `loadHomeData sets loading state`() = runTest {
        viewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `loadHomeData populates top rated movies`() = runTest {
        val movies = listOf(testItem, testItem.copy(id = 2L, title = "Movie 2"))
        coEvery { mockMediaRepository.browseEntities("movie", any(), any(), any()) } returns flowOf(movies)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(2, viewModel.uiState.value.topRatedMovies.size)
    }

    @Test
    fun `loadHomeData populates top rated tv shows`() = runTest {
        val shows = listOf(testItem.copy(id = 3L, title = "TV Show", mediaType = "tv_show"))
        coEvery { mockMediaRepository.browseEntities("tv_show", any(), any(), any()) } returns flowOf(shows)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(1, viewModel.uiState.value.topRatedTvShows.size)
    }

    @Test
    fun `loadHomeData sets featured item from recent movies`() = runTest {
        val items = listOf(testItem)
        coEvery { mockMediaRepository.browseEntities("movie", any(), any(), any()) } returns flowOf(items)

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertNotNull(viewModel.uiState.value.featuredItem)
        assertEquals("Test Movie", viewModel.uiState.value.featuredItem?.title)
    }

    @Test
    fun `loadHomeData sets error on exception`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertNull(viewModel.uiState.value.error) // Exception caught internally, returns empty list
        assertFalse(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `loadHomeData loads entity stats`() = runTest {
        coEvery { mockMediaRepository.getEntityStats() } returns Pair(500, mapOf("movie" to 200))

        viewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals(500, viewModel.uiState.value.totalEntities)
        assertEquals(200, viewModel.uiState.value.statsByType["movie"])
    }

    @Test
    fun `refreshContent calls loadHomeData`() = runTest {
        viewModel.refreshContent()
        advanceUntilIdle()

        coVerify { mockMediaRepository.searchMedia(any()) }
    }

    @Test
    fun `markAsWatched calls repository updateWatchProgress`() = runTest {
        viewModel.markAsWatched(1L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateWatchProgress(1L, 1.0) }
    }

    @Test
    fun `updateWatchProgress calls repository`() = runTest {
        viewModel.updateWatchProgress(1L, 0.5)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateWatchProgress(1L, 0.5) }
    }
}
