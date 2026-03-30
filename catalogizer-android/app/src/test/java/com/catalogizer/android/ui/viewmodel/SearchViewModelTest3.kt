package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SearchViewModelTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: SearchViewModel

    private val testMovieItem = MediaItem(
        id = 1L, title = "The Matrix", mediaType = "movie",
        description = "A sci-fi classic",
        directoryPath = "/media/movies", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    private val testMusicItem = MediaItem(
        id = 2L, title = "Dark Side of the Moon", mediaType = "music",
        description = "Pink Floyd album",
        directoryPath = "/media/music", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(
            listOf(testMovieItem, testMusicItem)
        )
        viewModel = SearchViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial query is empty`() {
        assertEquals("", viewModel.query.value)
    }

    @Test
    fun `initial search results are empty`() {
        assertTrue(viewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `initial isSearching is false`() {
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search updates query value`() {
        viewModel.search("matrix")
        assertEquals("matrix", viewModel.query.value)
    }

    @Test
    fun `blank query clears search results`() = runTest {
        viewModel.search("")
        advanceUntilIdle()

        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search filters results by title`() = runTest {
        viewModel.search("Matrix")
        // Wait for debounce (300ms) + execution
        advanceTimeBy(500)
        advanceUntilIdle()

        val results = viewModel.searchResults.value
        assertTrue(results.any { it.title.contains("Matrix", ignoreCase = true) })
    }

    @Test
    fun `search filters results by media type`() = runTest {
        viewModel.search("music")
        advanceTimeBy(500)
        advanceUntilIdle()

        val results = viewModel.searchResults.value
        assertTrue(results.any { it.mediaType.contains("music", ignoreCase = true) })
    }

    @Test
    fun `search filters results by description`() = runTest {
        viewModel.search("sci-fi")
        advanceTimeBy(500)
        advanceUntilIdle()

        val results = viewModel.searchResults.value
        assertTrue(results.any { it.description?.contains("sci-fi", ignoreCase = true) == true })
    }

    @Test
    fun `search with no matching results returns empty list`() = runTest {
        viewModel.search("nonexistent_xyz_123")
        advanceTimeBy(500)
        advanceUntilIdle()

        assertTrue(viewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `search handles API exception gracefully`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(50) } throws RuntimeException("Network error")

        viewModel.search("test")
        advanceTimeBy(500)
        advanceUntilIdle()

        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isSearching.value)
    }
}
