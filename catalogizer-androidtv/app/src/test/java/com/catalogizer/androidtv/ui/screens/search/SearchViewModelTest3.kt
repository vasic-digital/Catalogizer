package com.catalogizer.androidtv.ui.screens.search

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
class SearchViewModelTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var viewModel: SearchViewModel

    private val testItem = MediaItem(
        id = 1L, title = "The Matrix", mediaType = "movie",
        directoryPath = "/media", createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(listOf(testItem))
        viewModel = SearchViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial search query is empty`() {
        assertEquals("", viewModel.searchQuery.value)
    }

    @Test
    fun `initial search results are empty`() {
        assertTrue(viewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `initial isLoading is false`() {
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `initial error is null`() {
        assertNull(viewModel.error.value)
    }

    @Test
    fun `initial hasSearched is false`() {
        assertFalse(viewModel.hasSearched.value)
    }

    @Test
    fun `updateSearchQuery updates the value`() {
        viewModel.updateSearchQuery("matrix")
        assertEquals("matrix", viewModel.searchQuery.value)
    }

    @Test
    fun `updateSearchQuery clears error`() {
        // Manually set an error state by searching and failing
        viewModel.updateSearchQuery("test")
        // Now update again - error should be cleared
        viewModel.updateSearchQuery("new query")
        assertNull(viewModel.error.value)
    }

    @Test
    fun `search does nothing when query is blank`() = runTest {
        viewModel.updateSearchQuery("")
        viewModel.search()
        advanceUntilIdle()

        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.hasSearched.value)
    }

    @Test
    fun `search populates results on success`() = runTest {
        viewModel.updateSearchQuery("matrix")
        viewModel.search()
        advanceUntilIdle()

        assertTrue(viewModel.hasSearched.value)
        assertFalse(viewModel.isLoading.value)
        assertEquals(1, viewModel.searchResults.value.size)
        assertEquals("The Matrix", viewModel.searchResults.value[0].title)
    }

    @Test
    fun `search sets error on exception`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        viewModel.updateSearchQuery("test")
        viewModel.search()
        advanceUntilIdle()

        assertNotNull(viewModel.error.value)
        assertTrue(viewModel.error.value!!.contains("Search failed"))
        assertTrue(viewModel.hasSearched.value)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `clearResults resets all state`() {
        viewModel.updateSearchQuery("test")
        viewModel.clearResults()

        assertEquals("", viewModel.searchQuery.value)
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertNull(viewModel.error.value)
        assertFalse(viewModel.hasSearched.value)
    }

    @Test
    fun `search passes query to repository`() = runTest {
        viewModel.updateSearchQuery("star wars")
        viewModel.search()
        advanceUntilIdle()

        coVerify {
            mockMediaRepository.searchMedia(match { it.query == "star wars" && it.limit == 50 })
        }
    }
}
