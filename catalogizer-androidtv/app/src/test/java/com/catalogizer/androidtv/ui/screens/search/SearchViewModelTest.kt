package com.catalogizer.androidtv.ui.screens.search

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
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SearchViewModelTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mediaRepository: MediaRepository
    private lateinit var viewModel: SearchViewModel

    private val sampleMediaItems = listOf(
        MediaItem(
            id = 1,
            title = "Test Movie",
            mediaType = "movie",
            year = 2024,
            directoryPath = "/movies",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            fileSize = 1000000000L,
            description = "A test movie"
        ),
        MediaItem(
            id = 2,
            title = "Test Series",
            mediaType = "series",
            year = 2023,
            directoryPath = "/series",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            fileSize = 2000000000L,
            description = "A test series"
        )
    )

    @Before
    fun setup() {
        mediaRepository = mockk()
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(sampleMediaItems)
        viewModel = SearchViewModel(mediaRepository)
    }

    @Test
    fun `initial state should have empty results and not loading`() {
        assertEquals("", viewModel.searchQuery.value)
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `updateSearchQuery should update query and clear error`() {
        viewModel.updateSearchQuery("test query")

        assertEquals("test query", viewModel.searchQuery.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `search with blank query should not call repository`() = runTest {
        viewModel.updateSearchQuery("")

        viewModel.search()
        advanceUntilIdle()

        coVerify(exactly = 0) { mediaRepository.searchMedia(any()) }
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `search with whitespace only query should not call repository`() = runTest {
        viewModel.updateSearchQuery("   ")

        viewModel.search()
        advanceUntilIdle()

        coVerify(exactly = 0) { mediaRepository.searchMedia(any()) }
    }

    @Test
    fun `search should call repository and update results`() = runTest {
        viewModel.updateSearchQuery("test")

        viewModel.search()
        advanceUntilIdle()

        coVerify {
            mediaRepository.searchMedia(match {
                it.query == "test" && it.limit == 50
            })
        }
        assertEquals(sampleMediaItems, viewModel.searchResults.value)
        assertFalse(viewModel.isLoading.value)
        assertNull(viewModel.error.value)
    }

    @Test
    fun `search should handle repository errors`() = runTest {
        coEvery { mediaRepository.searchMedia(any()) } throws Exception("Network error")

        viewModel.updateSearchQuery("test")
        viewModel.search()
        advanceUntilIdle()

        assertTrue(viewModel.error.value?.contains("Search failed") == true)
        assertTrue(viewModel.error.value?.contains("Network error") == true)
        assertFalse(viewModel.isLoading.value)
    }

    @Test
    fun `clearResults should reset all state`() {
        viewModel.updateSearchQuery("test")

        viewModel.clearResults()

        assertEquals("", viewModel.searchQuery.value)
        assertTrue(viewModel.searchResults.value.isEmpty())
        assertNull(viewModel.error.value)
    }

    @Test
    fun `multiple searches should update results correctly`() = runTest {
        val firstResults = listOf(sampleMediaItems[0])
        val secondResults = listOf(sampleMediaItems[1])

        coEvery { mediaRepository.searchMedia(match { it.query == "first" }) } returns flowOf(firstResults)
        coEvery { mediaRepository.searchMedia(match { it.query == "second" }) } returns flowOf(secondResults)

        // First search
        viewModel.updateSearchQuery("first")
        viewModel.search()
        advanceUntilIdle()

        assertEquals(firstResults, viewModel.searchResults.value)

        // Second search
        viewModel.updateSearchQuery("second")
        viewModel.search()
        advanceUntilIdle()

        assertEquals(secondResults, viewModel.searchResults.value)
    }

    @Test
    fun `search should create correct MediaSearchRequest`() = runTest {
        viewModel.updateSearchQuery("movie title")

        viewModel.search()
        advanceUntilIdle()

        coVerify {
            mediaRepository.searchMedia(
                MediaSearchRequest(
                    query = "movie title",
                    limit = 50
                )
            )
        }
    }
}
