package com.catalogizer.androidtv.ui.screens.search

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
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SearchViewModelTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockMediaRepository = mockk<MediaRepository>(relaxed = true)
    private lateinit var viewModel: SearchViewModel

    private val testItems = listOf(
        MediaItem(id = 1, title = "Inception", mediaType = "movie", directoryPath = "/m/1", createdAt = "2024-01-01", updatedAt = "2024-01-01"),
        MediaItem(id = 2, title = "Interstellar", mediaType = "movie", directoryPath = "/m/2", createdAt = "2024-01-01", updatedAt = "2024-01-01")
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
    fun `initial state has empty query`() {
        assertEquals("", viewModel.searchQuery.value)
    }

    @Test
    fun `initial state has empty results`() {
        assertTrue(viewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `initial state is not loading`() {
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `initial state has no error`() {
        assertNull(viewModel.error.value)
    }

    @Test
    fun `updateSearchQuery updates query state`() {
        viewModel.updateSearchQuery("inception")
        assertEquals("inception", viewModel.searchQuery.value)
    }

    @Test
    fun `updateSearchQuery clears error`() {
        viewModel.updateSearchQuery("test")
        assertNull(viewModel.error.value)
    }

    @Test
    fun `search with blank query does not search`() = runTest {
        viewModel.updateSearchQuery("")
        viewModel.search()
        advanceUntilIdle()

        assertFalse(viewModel.isLoading.value)
        coVerify(exactly = 0) { mockMediaRepository.searchMedia(any()) }
    }

    @Test
    fun `search with query calls repository`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(testItems)

        viewModel.updateSearchQuery("inception")
        viewModel.search()
        advanceUntilIdle()

        assertEquals(2, viewModel.searchResults.value.size)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `search failure sets error`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        viewModel.updateSearchQuery("inception")
        viewModel.search()
        advanceUntilIdle()

        assertNotNull(viewModel.error.value)
        assertTrue(viewModel.error.value?.contains("Search failed") == true)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `clearResults resets all state`() {
        viewModel.updateSearchQuery("inception")
        viewModel.clearResults()

        assertEquals("", viewModel.searchQuery.value)
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertNull(viewModel.error.value)
    }
}
