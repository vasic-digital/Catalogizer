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
 * ViewModel tests for SearchViewModel
 */
@ExperimentalCoroutinesApi
class SearchViewModelTest {

    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: SearchViewModel

    private fun createMediaItem(
        id: Long,
        title: String,
        mediaType: String = "movie",
        description: String? = null
    ): MediaItem = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        description = description,
        directoryPath = "/media/$id",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z"
    )

    @Before
    fun setup() {
        Dispatchers.setMain(StandardTestDispatcher())
        mockMediaRepository = mockk(relaxed = true)
        viewModel = SearchViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
        clearAllMocks()
    }

    @Test
    fun `initial state should have empty query, empty results, and not searching`() = runTest {
        // When
        advanceUntilIdle()

        // Then
        assertEquals("", viewModel.query.value)
        assertEquals(emptyList<MediaItem>(), viewModel.searchResults.value)
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search with query updates results`() = runTest {
        // Given
        val mediaItems = listOf(
            createMediaItem(1, "The Matrix", "movie"),
            createMediaItem(2, "Matrix Reloaded", "movie"),
            createMediaItem(3, "Unrelated Film", "movie")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(mediaItems)

        // When
        viewModel.search("Matrix")
        advanceUntilIdle()

        // Then
        assertEquals("Matrix", viewModel.query.value)
        assertFalse(viewModel.isSearching.value)
        assertEquals(2, viewModel.searchResults.value.size)
        assertTrue(viewModel.searchResults.value.all { it.title.contains("Matrix") })

        // Verify repository was called
        coVerify { mockMediaRepository.getRecentMedia(50) }
    }

    @Test
    fun `search with blank query clears results`() = runTest {
        // Given - first perform a search to have results
        val mediaItems = listOf(
            createMediaItem(1, "The Matrix", "movie")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(mediaItems)

        viewModel.search("Matrix")
        advanceUntilIdle()

        // Verify we have results
        assertEquals(1, viewModel.searchResults.value.size)

        // When - search with blank query
        viewModel.search("")
        advanceUntilIdle()

        // Then
        assertEquals("", viewModel.query.value)
        assertEquals(emptyList<MediaItem>(), viewModel.searchResults.value)
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search debounces and cancels previous job`() = runTest {
        // Given
        val mediaItems = listOf(
            createMediaItem(1, "Final Search Result", "movie")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(mediaItems)

        // When - rapidly search multiple queries without advancing time
        viewModel.search("first")
        viewModel.search("second")
        viewModel.search("Final")
        advanceUntilIdle()

        // Then - only the last query should be the active one
        assertEquals("Final", viewModel.query.value)
        assertFalse(viewModel.isSearching.value)

        // The repository should have been called at most once (for the final query after debounce)
        coVerify(atMost = 1) { mockMediaRepository.getRecentMedia(50) }
    }

    @Test
    fun `search handles errors gracefully`() = runTest {
        // Given
        coEvery { mockMediaRepository.getRecentMedia(50) } throws RuntimeException("Network error")

        // When
        viewModel.search("test")
        advanceUntilIdle()

        // Then - should have empty results and not be searching
        assertEquals("test", viewModel.query.value)
        assertEquals(emptyList<MediaItem>(), viewModel.searchResults.value)
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search filters results by title match`() = runTest {
        // Given
        val mediaItems = listOf(
            createMediaItem(1, "Action Movie", "action"),
            createMediaItem(2, "Comedy Show", "comedy"),
            createMediaItem(3, "Action Hero", "action", description = "A great film")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(mediaItems)

        // When
        viewModel.search("Action")
        advanceUntilIdle()

        // Then
        assertEquals(2, viewModel.searchResults.value.size)
        assertTrue(viewModel.searchResults.value.any { it.title == "Action Movie" })
        assertTrue(viewModel.searchResults.value.any { it.title == "Action Hero" })
    }

    @Test
    fun `search filters results by description match`() = runTest {
        // Given
        val mediaItems = listOf(
            createMediaItem(1, "Some Film", "movie", description = "A thrilling adventure"),
            createMediaItem(2, "Another Film", "movie", description = "A boring documentary")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(mediaItems)

        // When
        viewModel.search("thrilling")
        advanceUntilIdle()

        // Then
        assertEquals(1, viewModel.searchResults.value.size)
        assertEquals("Some Film", viewModel.searchResults.value[0].title)
    }

    @Test
    fun `search filters results by mediaType match`() = runTest {
        // Given
        val mediaItems = listOf(
            createMediaItem(1, "Some Film", "movie"),
            createMediaItem(2, "A Song", "music")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(mediaItems)

        // When
        viewModel.search("music")
        advanceUntilIdle()

        // Then
        assertEquals(1, viewModel.searchResults.value.size)
        assertEquals("A Song", viewModel.searchResults.value[0].title)
    }
}
