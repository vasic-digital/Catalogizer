package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SearchViewModelTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockMediaRepository = mockk<MediaRepository>(relaxed = true)
    private lateinit var viewModel: SearchViewModel

    private val testMediaItems = listOf(
        MediaItem(id = 1, title = "Inception", mediaType = "movie", directoryPath = "/m/1", createdAt = "2024-01-01", updatedAt = "2024-01-01", description = "Dream movie"),
        MediaItem(id = 2, title = "Interstellar", mediaType = "movie", directoryPath = "/m/2", createdAt = "2024-01-01", updatedAt = "2024-01-01", description = "Space movie"),
        MediaItem(id = 3, title = "Dark Knight", mediaType = "movie", directoryPath = "/m/3", createdAt = "2024-01-01", updatedAt = "2024-01-01", description = "Batman")
    )

    @Before
    fun setup() {
        viewModel = SearchViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state has empty query and results`() {
        assertEquals("", viewModel.query.value)
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search with blank query clears results`() = runTest {
        viewModel.search("")
        advanceUntilIdle()

        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search with blank query after spaces clears results`() = runTest {
        viewModel.search("   ")
        advanceUntilIdle()

        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search updates query value`() = runTest {
        viewModel.search("inception")

        assertEquals("inception", viewModel.query.value)
    }

    @Test
    fun `search filters results by title`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(testMediaItems)

        viewModel.search("Inception")
        advanceTimeBy(400) // past debounce
        advanceUntilIdle()

        val results = viewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("Inception", results[0].title)
    }

    @Test
    fun `search filters results by media type`() = runTest {
        val items = testMediaItems + listOf(
            MediaItem(id = 4, title = "Song", mediaType = "music", directoryPath = "/m/4", createdAt = "2024-01-01", updatedAt = "2024-01-01")
        )
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(items)

        viewModel.search("music")
        advanceTimeBy(400)
        advanceUntilIdle()

        val results = viewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("Song", results[0].title)
    }

    @Test
    fun `search filters results by description`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(testMediaItems)

        viewModel.search("Dream")
        advanceTimeBy(400)
        advanceUntilIdle()

        val results = viewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("Inception", results[0].title)
    }

    @Test
    fun `search handles API failure gracefully`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.error("Network error")

        viewModel.search("test")
        advanceTimeBy(400)
        advanceUntilIdle()

        // On failure, searchResults should be empty (the filter on null data returns empty)
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isSearching.value)
    }

    @Test
    fun `search is case insensitive`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(50) } returns ApiResult.success(testMediaItems)

        viewModel.search("inception")
        advanceTimeBy(400)
        advanceUntilIdle()

        assertEquals(1, viewModel.searchResults.value.size)
        assertEquals("Inception", viewModel.searchResults.value[0].title)
    }
}
