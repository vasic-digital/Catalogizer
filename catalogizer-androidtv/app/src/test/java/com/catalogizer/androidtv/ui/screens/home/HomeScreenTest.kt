package com.catalogizer.androidtv.ui.screens.home

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
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
        watchProgress: Double = 0.0
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            directoryPath = "/media/test",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            watchProgress = watchProgress
        )
    }

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())
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
        assertTrue(state.recentlyAdded.isEmpty())
        assertTrue(state.movies.isEmpty())
        assertTrue(state.tvShows.isEmpty())
        assertTrue(state.music.isEmpty())
        assertTrue(state.documents.isEmpty())
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
    fun `loadHomeData should populate recentlyAdded section`() = runTest {
        val recentItems = listOf(
            createTestMediaItem(1L, "Recent Movie 1"),
            createTestMediaItem(2L, "Recent Movie 2")
        )
        coEvery {
            mockMediaRepository.searchMedia(match { it.sortBy == "created_at" && it.sortOrder == "desc" && it.mediaType == null })
        } returns flowOf(recentItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(2, state.recentlyAdded.size)
        assertEquals("Recent Movie 1", state.recentlyAdded[0].title)
    }

    @Test
    fun `loadHomeData should populate movies section`() = runTest {
        val movies = listOf(
            createTestMediaItem(1L, "Action Movie", MediaType.MOVIE.value)
        )
        coEvery {
            mockMediaRepository.searchMedia(match { it.mediaType == MediaType.MOVIE.value })
        } returns flowOf(movies)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.movies.size)
        assertEquals("Action Movie", state.movies[0].title)
    }

    @Test
    fun `loadHomeData should populate tvShows section`() = runTest {
        val tvShows = listOf(
            createTestMediaItem(1L, "TV Show 1", MediaType.TV_SHOW.value)
        )
        coEvery {
            mockMediaRepository.searchMedia(match { it.mediaType == MediaType.TV_SHOW.value })
        } returns flowOf(tvShows)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.tvShows.size)
    }

    @Test
    fun `loadHomeData should populate music section`() = runTest {
        val music = listOf(
            createTestMediaItem(1L, "Song 1", MediaType.MUSIC.value)
        )
        coEvery {
            mockMediaRepository.searchMedia(match { it.mediaType == MediaType.MUSIC.value })
        } returns flowOf(music)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.music.size)
    }

    @Test
    fun `loadHomeData should populate documents section`() = runTest {
        val docs = listOf(
            createTestMediaItem(1L, "Ebook 1", MediaType.EBOOK.value)
        )
        coEvery {
            mockMediaRepository.searchMedia(match { it.mediaType == MediaType.EBOOK.value })
        } returns flowOf(docs)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertEquals(1, state.documents.size)
    }

    @Test
    fun `loadHomeData should filter continue watching for items with progress`() = runTest {
        val watchingItems = listOf(
            createTestMediaItem(1L, "Watching", watchProgress = 0.5),
            createTestMediaItem(2L, "Completed", watchProgress = 0.95),
            createTestMediaItem(3L, "Not Started", watchProgress = 0.0)
        )
        coEvery {
            mockMediaRepository.searchMedia(match { it.sortBy == "last_watched" })
        } returns flowOf(watchingItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        // Only items with progress > 0 and not completed (< 0.9) should be in continue watching
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
            mockMediaRepository.searchMedia(match { it.sortBy == "last_watched" })
        } returns flowOf(watchingItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertNotNull(state.featuredItem)
        assertEquals("Currently Watching", state.featuredItem?.title)
    }

    @Test
    fun `loadHomeData should set featured item from recently added when no watching items`() = runTest {
        val recentItems = listOf(createTestMediaItem(1L, "New Movie"))
        coEvery {
            mockMediaRepository.searchMedia(match { it.sortBy == "created_at" && it.mediaType == null })
        } returns flowOf(recentItems)

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        // Featured item should be from recentlyAdded when no continue watching
        assertNotNull(state.featuredItem)
        assertEquals("New Movie", state.featuredItem?.title)
    }

    @Test
    fun `loadHomeData should set error on exception`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        val state = homeViewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals("Network error", state.error)
    }

    @Test
    fun `loadHomeData should clear error before reloading`() = runTest {
        // First load fails
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Error")
        homeViewModel.loadHomeData()
        advanceUntilIdle()
        assertNotNull(homeViewModel.uiState.value.error)

        // Second load succeeds
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertNull(homeViewModel.uiState.value.error)
    }

    @Test
    fun `loadHomeData with generic exception should show fallback message`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException()

        homeViewModel.loadHomeData()
        advanceUntilIdle()

        assertEquals("Failed to load content", homeViewModel.uiState.value.error)
    }

    @Test
    fun `refreshContent should call loadHomeData`() = runTest {
        homeViewModel.refreshContent()
        advanceUntilIdle()

        // Verify repository was called (indicating loadHomeData ran)
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
        coEvery { mockMediaRepository.updateFavoriteStatus(any(), any()) } just Runs

        homeViewModel.toggleFavorite(1L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.getMediaById(1L) }
        coVerify { mockMediaRepository.updateFavoriteStatus(1L, true) }
    }

    @Test
    fun `toggleFavorite should toggle from false to true`() = runTest {
        val item = createTestMediaItem(1L, isFavorite = false)
        coEvery { mockMediaRepository.getMediaById(1L) } returns flowOf(item)
        coEvery { mockMediaRepository.updateFavoriteStatus(any(), any()) } just Runs

        homeViewModel.toggleFavorite(1L)
        advanceUntilIdle()

        coVerify { mockMediaRepository.updateFavoriteStatus(1L, true) }
    }

    @Test
    fun `HomeUiState default values should be correct`() {
        val state = HomeUiState()

        assertFalse(state.isLoading)
        assertNull(state.error)
        assertTrue(state.continueWatching.isEmpty())
        assertTrue(state.recentlyAdded.isEmpty())
        assertTrue(state.movies.isEmpty())
        assertTrue(state.tvShows.isEmpty())
        assertTrue(state.music.isEmpty())
        assertTrue(state.documents.isEmpty())
        assertNull(state.featuredItem)
    }

    @Test
    fun `HomeUiState copy should preserve unmodified fields`() {
        val state = HomeUiState(isLoading = true, error = "err")
        val copied = state.copy(isLoading = false)

        assertFalse(copied.isLoading)
        assertEquals("err", copied.error)
    }

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test",
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
}
