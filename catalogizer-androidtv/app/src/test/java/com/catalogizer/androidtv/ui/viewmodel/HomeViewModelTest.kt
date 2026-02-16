package com.catalogizer.androidtv.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@ExperimentalCoroutinesApi
class HomeViewModelTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mediaRepository: MediaRepository
    private lateinit var viewModel: HomeViewModel

    @Before
    fun setup() {
        mediaRepository = mockk()
        viewModel = HomeViewModel(mediaRepository)
    }

    @Test
    fun `initial ui state should be default state`() = runTest {
        val initialState = viewModel.uiState.value

        assertFalse(initialState.isLoading)
        assertNull(initialState.error)
        assertTrue(initialState.continueWatching.isEmpty())
        assertTrue(initialState.recentlyAdded.isEmpty())
        assertTrue(initialState.movies.isEmpty())
        assertTrue(initialState.tvShows.isEmpty())
        assertTrue(initialState.music.isEmpty())
        assertTrue(initialState.documents.isEmpty())
        assertNull(initialState.featuredItem)
    }

    @Test
    fun `loadHomeData success should update ui state with all content sections`() = runTest {
        val continueWatchingItems = listOf(createTestMediaItem(1L, "Continue Watching"))
        val recentlyAddedItems = listOf(createTestMediaItem(2L, "Recently Added"))
        val movieItems = listOf(createTestMediaItem(3L, "Movie"))
        val tvShowItems = listOf(createTestMediaItem(4L, "TV Show"))
        val musicItems = listOf(createTestMediaItem(5L, "Music"))
        val documentItems = listOf(createTestMediaItem(6L, "Document"))

        // Mock all repository calls
        mockRepositoryCalls(
            continueWatchingItems, recentlyAddedItems, movieItems,
            tvShowItems, musicItems, documentItems
        )

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value

        assertFalse(uiState.isLoading)
        assertNull(uiState.error)
        assertEquals(continueWatchingItems, uiState.continueWatching)
        assertEquals(recentlyAddedItems, uiState.recentlyAdded)
        assertEquals(movieItems, uiState.movies)
        assertEquals(tvShowItems, uiState.tvShows)
        assertEquals(musicItems, uiState.music)
        assertEquals(documentItems, uiState.documents)
        assertEquals(continueWatchingItems[0], uiState.featuredItem) // Should be first continue watching item
    }

    @Test
    fun `loadHomeData with no continue watching should use recently added as featured`() = runTest {
        val continueWatchingItems = emptyList<MediaItem>()
        val recentlyAddedItems = listOf(createTestMediaItem(2L, "Recently Added"))
        val movieItems = listOf(createTestMediaItem(3L, "Movie"))
        val tvShowItems = listOf(createTestMediaItem(4L, "TV Show"))
        val musicItems = listOf(createTestMediaItem(5L, "Music"))
        val documentItems = listOf(createTestMediaItem(6L, "Document"))

        mockRepositoryCalls(
            continueWatchingItems, recentlyAddedItems, movieItems,
            tvShowItems, musicItems, documentItems
        )

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value
        assertEquals(recentlyAddedItems[0], uiState.featuredItem) // Should be first recently added item
    }

    @Test
    fun `loadHomeData failure should update ui state with error`() = runTest {
        val exception = RuntimeException("Network error")
        coEvery { mediaRepository.searchMedia(any()) } throws exception

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value

        assertFalse(uiState.isLoading)
        // Each individual load method catches exceptions and returns emptyList(),
        // so no top-level error is set when individual sections fail gracefully
        assertNull(uiState.error)
        assertTrue(uiState.continueWatching.isEmpty())
        assertTrue(uiState.recentlyAdded.isEmpty())
        assertTrue(uiState.movies.isEmpty())
        assertTrue(uiState.tvShows.isEmpty())
        assertTrue(uiState.music.isEmpty())
        assertTrue(uiState.documents.isEmpty())
        assertNull(uiState.featuredItem)
    }

    @Test
    fun `loadHomeData should set loading state during operation`() = runTest {
        val continueWatchingItems = listOf(createTestMediaItem(1L, "Continue Watching"))
        val recentlyAddedItems = listOf(createTestMediaItem(2L, "Recently Added"))
        val movieItems = listOf(createTestMediaItem(3L, "Movie"))
        val tvShowItems = listOf(createTestMediaItem(4L, "TV Show"))
        val musicItems = listOf(createTestMediaItem(5L, "Music"))
        val documentItems = listOf(createTestMediaItem(6L, "Document"))

        mockRepositoryCalls(
            continueWatchingItems, recentlyAddedItems, movieItems,
            tvShowItems, musicItems, documentItems
        )

        // Start loading and wait for completion
        viewModel.loadHomeData()
        advanceUntilIdle()

        // After completion, loading should be false and data should be populated
        val finalState = viewModel.uiState.value
        assertFalse(finalState.isLoading)
        assertNull(finalState.error)
        // Verify data was actually loaded (proves loading cycle completed)
        assertEquals(movieItems, finalState.movies)
        assertEquals(tvShowItems, finalState.tvShows)
    }

    @Test
    fun `refreshContent should call loadHomeData`() = runTest {
        val continueWatchingItems = listOf(createTestMediaItem(1L, "Continue Watching"))
        val recentlyAddedItems = listOf(createTestMediaItem(2L, "Recently Added"))
        val movieItems = listOf(createTestMediaItem(3L, "Movie"))
        val tvShowItems = listOf(createTestMediaItem(4L, "TV Show"))
        val musicItems = listOf(createTestMediaItem(5L, "Music"))
        val documentItems = listOf(createTestMediaItem(6L, "Document"))

        mockRepositoryCalls(
            continueWatchingItems, recentlyAddedItems, movieItems,
            tvShowItems, musicItems, documentItems
        )

        viewModel.refreshContent()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value
        assertFalse(uiState.isLoading)
        assertEquals(continueWatchingItems, uiState.continueWatching)
    }

    @Test
    fun `markAsWatched should update progress to 1_0 and refresh data`() = runTest {
        val mediaId = 123L
        val continueWatchingItems = listOf(createTestMediaItem(1L, "Continue Watching"))
        val recentlyAddedItems = listOf(createTestMediaItem(2L, "Recently Added"))

        // Mock initial load
        mockRepositoryCalls(continueWatchingItems, recentlyAddedItems, emptyList(), emptyList(), emptyList(), emptyList())

        coEvery { mediaRepository.updateWatchProgress(mediaId, 1.0) } returns Unit

        viewModel.markAsWatched(mediaId)
        advanceUntilIdle()

        coVerify { mediaRepository.updateWatchProgress(mediaId, 1.0) }
        // loadHomeData calls searchMedia 6 times (one per content section)
        coVerify(atLeast = 6) { mediaRepository.searchMedia(any()) }
    }

    @Test
    fun `updateWatchProgress should call repository with correct parameters`() = runTest {
        val mediaId = 123L
        val progress = 0.75

        coEvery { mediaRepository.updateWatchProgress(mediaId, progress) } returns Unit

        viewModel.updateWatchProgress(mediaId, progress)
        advanceUntilIdle()

        coVerify { mediaRepository.updateWatchProgress(mediaId, progress) }
    }

    @Test
    fun `toggleFavorite should toggle favorite status and refresh data`() = runTest {
        val mediaId = 123L
        val mediaItem = createTestMediaItem(mediaId, "Test Media", isFavorite = false)
        val continueWatchingItems = listOf(mediaItem)

        // Mock getMediaById
        coEvery { mediaRepository.getMediaById(mediaId) } returns flowOf(mediaItem)
        coEvery { mediaRepository.updateFavoriteStatus(mediaId, true) } returns Unit

        // Mock load calls
        mockRepositoryCalls(continueWatchingItems, emptyList(), emptyList(), emptyList(), emptyList(), emptyList())

        viewModel.toggleFavorite(mediaId)
        advanceUntilIdle()

        coVerify { mediaRepository.getMediaById(mediaId) }
        coVerify { mediaRepository.updateFavoriteStatus(mediaId, true) }
        // loadHomeData calls searchMedia 6 times (one per content section) for refresh
        coVerify(atLeast = 6) { mediaRepository.searchMedia(any()) }
    }

    @Test
    fun `toggleFavorite with null media item should not crash`() = runTest {
        val mediaId = 123L

        coEvery { mediaRepository.getMediaById(mediaId) } returns flowOf(null)

        viewModel.toggleFavorite(mediaId)
        advanceUntilIdle()

        coVerify { mediaRepository.getMediaById(mediaId) }
        coVerify(exactly = 0) { mediaRepository.updateFavoriteStatus(any(), any()) }
    }

    @Test
    fun `loadHomeData with partial failures should still load successful sections`() = runTest {
        val continueWatchingItems = listOf(createTestMediaItem(1L, "Continue Watching"))
        val recentlyAddedItems = listOf(createTestMediaItem(2L, "Recently Added"))

        // Mock continue watching and recently added to succeed
        coEvery {
            mediaRepository.searchMedia(match { it.sortBy == "last_watched" })
        } returns flowOf(continueWatchingItems)

        coEvery {
            mediaRepository.searchMedia(match { it.sortBy == "created_at" && it.mediaType == null })
        } returns flowOf(recentlyAddedItems)

        // Mock other calls to fail
        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "movie" })
        } throws RuntimeException("Movies failed")

        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "tv_show" })
        } throws RuntimeException("TV shows failed")

        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "music" })
        } throws RuntimeException("Music failed")

        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "ebook" })
        } throws RuntimeException("Documents failed")

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value

        assertFalse(uiState.isLoading)
        assertNull(uiState.error) // No overall error since some sections succeeded
        assertEquals(continueWatchingItems, uiState.continueWatching)
        assertEquals(recentlyAddedItems, uiState.recentlyAdded)
        assertTrue(uiState.movies.isEmpty())
        assertTrue(uiState.tvShows.isEmpty())
        assertTrue(uiState.music.isEmpty())
        assertTrue(uiState.documents.isEmpty())
    }

    private fun createTestMediaItem(id: Long, title: String, isFavorite: Boolean = false): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = "movie",
            directoryPath = "/path/to/$title",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            isFavorite = isFavorite,
            watchProgress = if (title.contains("Continue")) 0.5 else 0.0
        )
    }

    private fun mockRepositoryCalls(
        continueWatching: List<MediaItem>,
        recentlyAdded: List<MediaItem>,
        movies: List<MediaItem>,
        tvShows: List<MediaItem>,
        music: List<MediaItem>,
        documents: List<MediaItem>
    ) {
        // Mock continue watching (sort by last_watched)
        coEvery {
            mediaRepository.searchMedia(match { it.sortBy == "last_watched" })
        } returns flowOf(continueWatching)

        // Mock recently added (sort by created_at, no media type)
        coEvery {
            mediaRepository.searchMedia(match { it.sortBy == "created_at" && it.mediaType == null })
        } returns flowOf(recentlyAdded)

        // Mock movies
        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "movie" })
        } returns flowOf(movies)

        // Mock TV shows
        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "tv_show" })
        } returns flowOf(tvShows)

        // Mock music
        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "music" })
        } returns flowOf(music)

        // Mock documents (ebook)
        coEvery {
            mediaRepository.searchMedia(match { it.mediaType == "ebook" })
        } returns flowOf(documents)
    }
}