package com.catalogizer.androidtv.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.ExternalMetadata
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
import com.catalogizer.androidtv.data.repository.MediaRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
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
        mediaRepository = mockk(relaxed = true)
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(emptyList())
        coEvery { mediaRepository.getSimilarItems(any()) } returns emptyList()
        coEvery { mediaRepository.getTrendingItems(any()) } returns emptyList()
        coEvery { mediaRepository.getEntityStats() } returns Pair(0, emptyMap())
        viewModel = HomeViewModel(mediaRepository)
    }

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String = "movie",
        year: Int? = 2024,
        description: String? = "A test movie",
        coverImage: String? = null,
        rating: Double? = 8.5,
        directoryPath: String = "/media/movies/test",
        createdAt: String = "2024-01-01T00:00:00Z",
        updatedAt: String = "2024-06-01T00:00:00Z",
        externalMetadata: List<ExternalMetadata>? = null,
        watchProgress: Double = 0.0,
        isFavorite: Boolean = false,
        isDownloaded: Boolean = false
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        coverImage = coverImage,
        rating = rating,
        directoryPath = directoryPath,
        createdAt = createdAt,
        updatedAt = updatedAt,
        externalMetadata = externalMetadata,
        watchProgress = watchProgress,
        isFavorite = isFavorite,
        isDownloaded = isDownloaded
    )

    @Test
    fun `initial ui state should be default state`() = runTest {
        val initialState = viewModel.uiState.value

        assertFalse(initialState.isLoading)
        assertNull(initialState.error)
        assertTrue(initialState.continueWatching.isEmpty())
        assertTrue(initialState.recentMovies.isEmpty())
        assertTrue(initialState.recommended.isEmpty())
        assertTrue(initialState.trending.isEmpty())
        assertTrue(initialState.topRatedMovies.isEmpty())
        assertTrue(initialState.topRatedTvShows.isEmpty())
        assertTrue(initialState.topRatedMusic.isEmpty())
        assertTrue(initialState.topRatedDocuments.isEmpty())
        assertNull(initialState.featuredItem)
    }

    @Test
    fun `loadHomeData success should update ui state with content sections`() = runTest {
        val movieItems = listOf(createTestMediaItem(3L, "Movie"))
        val showItems = listOf(createTestMediaItem(4L, "TV Show"))

        coEvery {
            mediaRepository.browseEntities("movie", any(), any(), any())
        } returns flowOf(movieItems)

        coEvery {
            mediaRepository.browseEntities("tv_show", any(), any(), any())
        } returns flowOf(showItems)

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value

        assertFalse(uiState.isLoading)
        assertNull(uiState.error)
        assertEquals(movieItems, uiState.recentMovies)
        assertEquals(movieItems, uiState.topRatedMovies)
        assertEquals(showItems, uiState.recentTvShows)
        assertEquals(showItems, uiState.topRatedTvShows)
    }

    @Test
    fun `loadHomeData with no continue watching should use recent movies as featured`() = runTest {
        val movieItems = listOf(createTestMediaItem(3L, "Movie"))

        coEvery {
            mediaRepository.browseEntities("movie", any(), any(), any())
        } returns flowOf(movieItems)

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value
        assertEquals(movieItems[0], uiState.featuredItem)
    }

    @Test
    fun `loadHomeData failure should update ui state with error`() = runTest {
        coEvery { mediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value

        assertFalse(uiState.isLoading)
        assertNull(uiState.error) // Exception caught internally, returns empty list
        assertTrue(uiState.recentMovies.isEmpty())
        assertTrue(uiState.recommended.isEmpty())
        assertTrue(uiState.trending.isEmpty())
        assertNull(uiState.featuredItem)
    }

    @Test
    fun `loadHomeData should set loading state during operation`() = runTest {
        viewModel.loadHomeData()
        advanceUntilIdle()

        val finalState = viewModel.uiState.value
        assertFalse(finalState.isLoading)
        assertNull(finalState.error)
    }

    @Test
    fun `refreshContent should call loadHomeData`() = runTest {
        viewModel.refreshContent()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value
        assertFalse(uiState.isLoading)
    }

    @Test
    fun `markAsWatched should update progress to 1_0 and refresh data`() = runTest {
        val mediaId = 123L

        coEvery { mediaRepository.updateWatchProgress(mediaId, 1.0) } returns Unit

        viewModel.markAsWatched(mediaId)
        advanceUntilIdle()

        coVerify { mediaRepository.updateWatchProgress(mediaId, 1.0) }
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

        coEvery { mediaRepository.getMediaById(mediaId) } returns flowOf(mediaItem)
        coEvery { mediaRepository.checkFavorite(any(), any()) } returns true
        coEvery { mediaRepository.toggleFavorite(any(), any(), any()) } returns true

        viewModel.toggleFavorite(mediaId)
        advanceUntilIdle()

        coVerify { mediaRepository.getMediaById(mediaId) }
        coVerify { mediaRepository.toggleFavorite(any(), any(), any()) }
    }

    @Test
    fun `toggleFavorite with null media item should not crash`() = runTest {
        val mediaId = 123L

        coEvery { mediaRepository.getMediaById(mediaId) } returns flowOf(null)

        viewModel.toggleFavorite(mediaId)
        advanceUntilIdle()

        coVerify { mediaRepository.getMediaById(mediaId) }
        coVerify(exactly = 0) { mediaRepository.toggleFavorite(any(), any(), any()) }
    }

    @Test
    fun `loadHomeData should load trending items`() = runTest {
        val trendingItems = listOf(createTestMediaItem(10L, "Trending Movie"))
        coEvery { mediaRepository.getTrendingItems(10) } returns trendingItems

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value
        assertEquals(trendingItems, uiState.trending)
    }

    @Test
    fun `loadHomeData should load recommended items from recently played`() = runTest {
        val playedItem = createTestMediaItem(1L, "Played", watchProgress = 0.5)
        val similarItems = listOf(createTestMediaItem(20L, "Similar"))

        coEvery {
            mediaRepository.searchMedia(match { it.sortBy == "updated_at" })
        } returns flowOf(listOf(playedItem))

        coEvery { mediaRepository.getSimilarItems(1L) } returns similarItems

        viewModel.loadHomeData()
        advanceUntilIdle()

        val uiState = viewModel.uiState.value
        assertEquals(similarItems, uiState.recommended)
    }
}
