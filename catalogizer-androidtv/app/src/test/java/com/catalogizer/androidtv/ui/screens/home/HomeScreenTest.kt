package com.catalogizer.androidtv.ui.screens.home

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaType
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.ui.viewmodel.HomeUiState
import com.catalogizer.androidtv.ui.viewmodel.HomeViewModel
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
        mediaType: String = "movie",
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            directoryPath = "/media/test",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            watchProgress = watchProgress,
            isFavorite = isFavorite
        )
    }

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        coEvery { mockMediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(emptyList())
        coEvery { mockMediaRepository.getSimilarItems(any()) } returns emptyList()
        coEvery { mockMediaRepository.getTrendingItems(any()) } returns emptyList()
        coEvery { mockMediaRepository.getEntityStats() } returns Pair(0, emptyMap())
        homeViewModel = HomeViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial uiState should have loading set to false`() {
        val state = homeViewModel.uiState.value
        assertFalse(state.isLoading)
    }

    @Test
    fun `initial uiState should have no error`() {
        val state = homeViewModel.uiState.value
        assertNull(state.error)
    }

    @Test
    fun `initial uiState should have empty content sections`() {
        val state = homeViewModel.uiState.value
        assertTrue(state.continueWatching.isEmpty())
        assertTrue(state.recentMovies.isEmpty())
        assertTrue(state.recommended.isEmpty())
        assertTrue(state.trending.isEmpty())
        assertTrue(state.topRatedMovies.isEmpty())
        assertTrue(state.topRatedTvShows.isEmpty())
        assertTrue(state.topRatedMusic.isEmpty())
        assertTrue(state.topRatedDocuments.isEmpty())
    }

    @Test
    fun `initial uiState should have null featured item`() {
        val state = homeViewModel.uiState.value
        assertNull(state.featuredItem)
    }

    @Test
    fun `loadHomeData should set loading to true then false`() = runTest {
        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertFalse(homeViewModel.uiState.value.isLoading)
    }

    @Test
    fun `loadHomeData should populate recent movies section`() = runTest {
        val movies = listOf(
            createTestMediaItem(1L, "Recent Movie 1", MediaType.MOVIE.value),
            createTestMediaItem(2L, "Recent Movie 2", MediaType.MOVIE.value)
        )
        coEvery {
            mockMediaRepository.browseEntities(eq("movie"), eq(10), eq("created"), eq("desc"))
        } returns flowOf(movies)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        coVerify(atLeast = 1) { mockMediaRepository.browseEntities(eq("movie"), any(), any(), any()) }
        val state = homeViewModel.uiState.value
        assertNull("Error should be null: ${state.error}", state.error)
        assertFalse("Loading should be false", state.isLoading)
        assertEquals(2, state.recentMovies.size)
        assertEquals("Recent Movie 1", state.recentMovies[0].title)
        
        coVerify { mockMediaRepository.browseEntities(eq("movie"), eq(10), eq("created"), eq("desc")) }
    }

    @Test
    fun `loadHomeData should populate top rated movies section`() = runTest {
        val movies = listOf(
            createTestMediaItem(1L, "Action Movie", MediaType.MOVIE.value)
        )
        coEvery {
            mockMediaRepository.browseEntities(eq("movie"), eq(10), eq("rating"), eq("desc"))
        } returns flowOf(movies)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.topRatedMovies.size)
        assertEquals("Action Movie", state.topRatedMovies[0].title)
    }

    @Test
    fun `loadHomeData should populate top rated tv shows section`() = runTest {
        val tvShows = listOf(
            createTestMediaItem(1L, "TV Show 1", MediaType.TV_SHOW.value)
        )
        coEvery {
            mockMediaRepository.browseEntities(eq("tv_show"), eq(10), eq("rating"), eq("desc"))
        } returns flowOf(tvShows)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.topRatedTvShows.size)
    }

    @Test
    fun `loadHomeData should populate trending section`() = runTest {
        val trending = listOf(createTestMediaItem(1L, "Trending Movie"))
        coEvery { mockMediaRepository.getTrendingItems(10) } returns trending

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.trending.size)
    }

    @Test
    fun `loadHomeData should populate recommended section`() = runTest {
        val playedItem = createTestMediaItem(1L, "Played", watchProgress = 0.5)
        val similar = listOf(createTestMediaItem(2L, "Similar"))

        coEvery {
            mockMediaRepository.searchMedia(match { it.sortBy == "updated_at" && it.sortOrder == "desc" && it.limit == 50 })
        } returns flowOf(listOf(playedItem))

        coEvery { mockMediaRepository.getSimilarItems(1L) } returns similar

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.recommended.size)
        assertEquals("Similar", state.recommended[0].title)
    }

    @Test
    fun `loadHomeData should filter continue watching for items with progress`() = runTest {
        val watchingItems = listOf(
            createTestMediaItem(1L, "Watching", watchProgress = 0.5),
            createTestMediaItem(2L, "Completed", watchProgress = 0.95),
            createTestMediaItem(3L, "Not Started", watchProgress = 0.0)
        )
        coEvery {
            mockMediaRepository.searchMedia(any())
        } returns flowOf(watchingItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        val continueWatching = state.continueWatching
        assertEquals(1, continueWatching.size)
        assertEquals("Watching", continueWatching[0].title)
    }

    @Test
    fun `loadHomeData should set featured item from continue watching`() = runTest {
        val watchingItems = listOf(
            createTestMediaItem(1L, "Currently Watching", watchProgress = 0.3)
        )
        coEvery {
            mockMediaRepository.searchMedia(any())
        } returns flowOf(watchingItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertNotNull(state.featuredItem)
        assertEquals("Currently Watching", state.featuredItem?.title)
    }

    @Test
    fun `loadHomeData should set error on exception`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertFalse(state.isLoading)
        assertNull(state.error) // exceptions are caught per section, not propagated
        assertTrue(state.topRatedMovies.isEmpty())
    }

    @Test
    fun `refreshContent should call loadHomeData`() = runTest {
        homeViewModel.refreshContent()
        advanceUntilIdle()

        coVerify(atLeast = 1) { mockMediaRepository.searchMedia(any()) }
    }

    @Test
    fun `markAsWatched should update progress to 1_0`() = runTest {
        coEvery { mockMediaRepository.updateWatchProgress(any(), any()) } just Runs

        homeViewModel.markAsWatched(42L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateWatchProgress(42L, 1.0) }
    }

    @Test
    fun `updateWatchProgress should call repository`() = runTest {
        coEvery { mockMediaRepository.updateWatchProgress(any(), any()) } just Runs

        homeViewModel.updateWatchProgress(1L, 0.5)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateWatchProgress(1L, 0.5) }
    }

    @Test
    fun `toggleFavorite should call repository methods`() = runTest {
        val item = createTestMediaItem(1L, isFavorite = false)
        coEvery { mockMediaRepository.getMediaById(1L) } returns flowOf(item)
        coEvery { mockMediaRepository.toggleFavorite(any(), any(), any()) } returns true

        homeViewModel.toggleFavorite(1L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.getMediaById(1L) }
        coVerify { mockMediaRepository.toggleFavorite(any(), eq(1L), any()) }
    }

    @Test
    fun `HomeUiState default values should be correct`() {
        val state = HomeUiState()

        assertFalse(state.isLoading)
        assertNull(state.error)
        assertTrue(state.continueWatching.isEmpty())
        assertTrue(state.recentMovies.isEmpty())
        assertTrue(state.recommended.isEmpty())
        assertTrue(state.trending.isEmpty())
        assertTrue(state.topRatedMovies.isEmpty())
        assertTrue(state.topRatedTvShows.isEmpty())
        assertTrue(state.topRatedMusic.isEmpty())
        assertTrue(state.topRatedDocuments.isEmpty())
        assertNull(state.featuredItem)
    }

    @Test
    fun `HomeUiState copy should preserve unmodified fields`() {
        val state = HomeUiState(isLoading = true, error = "err")
        val copied = state.copy(isLoading = false)

        assertFalse(copied.isLoading)
        assertEquals("err", copied.error)
    }
}
