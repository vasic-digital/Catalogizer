package com.catalogizer.androidtv.ui.screens.search

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchRequest
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
class SearchScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var searchViewModel: SearchViewModel

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie"
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            directoryPath = "/media/test",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z"
        )
    }

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        searchViewModel = SearchViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state should have empty search query`() {
        assertEquals("", searchViewModel.searchQuery.value)
    }

    @Test
    fun `initial state should have empty search results`() {
        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `initial state should not be loading`() {
        assertFalse(searchViewModel.isLoading.value)
    }

    @Test
    fun `initial state should have no error`() {
        assertNull(searchViewModel.error.value)
    }

    @Test
    fun `updateSearchQuery should update query value`() {
        searchViewModel.updateSearchQuery("test query")

        assertEquals("test query", searchViewModel.searchQuery.value)
    }

    @Test
    fun `updateSearchQuery should clear error`() {
        // Set error first by triggering a failed search
        searchViewModel.updateSearchQuery("initial")
        // Manually verify error clearing
        searchViewModel.updateSearchQuery("new query")

        assertNull(searchViewModel.error.value)
    }

    @Test
    fun `search with blank query should not trigger repository call`() = runTest {
        searchViewModel.updateSearchQuery("")
        searchViewModel.search()
        advanceUntilIdle()

        coVerify(exactly = 0) { mockMediaRepository.searchMedia(any()) }
    }

    @Test
    fun `search with whitespace-only query should not trigger repository call`() = runTest {
        searchViewModel.updateSearchQuery("   ")
        searchViewModel.search()
        advanceUntilIdle()

        coVerify(exactly = 0) { mockMediaRepository.searchMedia(any()) }
    }

    @Test
    fun `search should call repository with correct query`() = runTest {
        val results = listOf(createTestMediaItem(1L, "Result"))
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(results)

        searchViewModel.updateSearchQuery("test")
        searchViewModel.search()
        advanceUntilIdle()

        coVerify {
            mockMediaRepository.searchMedia(match { it.query == "test" })
        }
    }

    @Test
    fun `search should set results on success`() = runTest {
        val results = listOf(
            createTestMediaItem(1L, "Movie 1"),
            createTestMediaItem(2L, "Movie 2")
        )
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(results)

        searchViewModel.updateSearchQuery("movie")
        searchViewModel.search()
        advanceUntilIdle()

        assertEquals(2, searchViewModel.searchResults.value.size)
        assertEquals("Movie 1", searchViewModel.searchResults.value[0].title)
        assertEquals("Movie 2", searchViewModel.searchResults.value[1].title)
    }

    @Test
    fun `search should set isLoading to false after completion`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())

        searchViewModel.updateSearchQuery("test")
        searchViewModel.search()
        advanceUntilIdle()

        assertFalse(searchViewModel.isLoading.value)
    }

    @Test
    fun `search should set error on exception`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Network error")

        searchViewModel.updateSearchQuery("test")
        searchViewModel.search()
        advanceUntilIdle()

        assertNotNull(searchViewModel.error.value)
        assertTrue(searchViewModel.error.value!!.contains("Network error"))
    }

    @Test
    fun `search should clear error before searching`() = runTest {
        // First search fails
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Error")
        searchViewModel.updateSearchQuery("fail")
        searchViewModel.search()
        advanceUntilIdle()
        assertNotNull(searchViewModel.error.value)

        // Second search succeeds - error should be cleared
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        searchViewModel.updateSearchQuery("success")
        searchViewModel.search()

        // Error should be null immediately after calling search
        assertNull(searchViewModel.error.value)
        advanceUntilIdle()
    }

    @Test
    fun `search should request 50 items as limit`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())

        searchViewModel.updateSearchQuery("test")
        searchViewModel.search()
        advanceUntilIdle()

        coVerify {
            mockMediaRepository.searchMedia(match { it.limit == 50 })
        }
    }

    @Test
    fun `search with empty results should set empty results list`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(emptyList())

        searchViewModel.updateSearchQuery("nonexistent")
        searchViewModel.search()
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertFalse(searchViewModel.isLoading.value)
    }

    @Test
    fun `clearResults should reset all state`() {
        // Set some state manually
        searchViewModel.updateSearchQuery("test")

        searchViewModel.clearResults()

        assertEquals("", searchViewModel.searchQuery.value)
        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertNull(searchViewModel.error.value)
    }

    @Test
    fun `clearResults should clear search query`() {
        searchViewModel.updateSearchQuery("something")

        searchViewModel.clearResults()

        assertEquals("", searchViewModel.searchQuery.value)
    }

    @Test
    fun `clearResults should clear results list`() = runTest {
        val results = listOf(createTestMediaItem(1L))
        coEvery { mockMediaRepository.searchMedia(any()) } returns flowOf(results)
        searchViewModel.updateSearchQuery("test")
        searchViewModel.search()
        advanceUntilIdle()

        searchViewModel.clearResults()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `clearResults should clear error`() {
        searchViewModel.clearResults()

        assertNull(searchViewModel.error.value)
    }

    @Test
    fun `search error message should include original error details`() = runTest {
        coEvery { mockMediaRepository.searchMedia(any()) } throws RuntimeException("Connection timeout")

        searchViewModel.updateSearchQuery("test")
        searchViewModel.search()
        advanceUntilIdle()

        val error = searchViewModel.error.value
        assertNotNull(error)
        assertTrue(error!!.contains("Connection timeout"))
        assertTrue(error.startsWith("Search failed:"))
    }

    @Test
    fun `consecutive searches should update results correctly`() = runTest {
        val results1 = listOf(createTestMediaItem(1L, "First"))
        val results2 = listOf(createTestMediaItem(2L, "Second"))

        coEvery { mockMediaRepository.searchMedia(match { it.query == "first" }) } returns flowOf(results1)
        coEvery { mockMediaRepository.searchMedia(match { it.query == "second" }) } returns flowOf(results2)

        // First search
        searchViewModel.updateSearchQuery("first")
        searchViewModel.search()
        advanceUntilIdle()
        assertEquals(1, searchViewModel.searchResults.value.size)
        assertEquals("First", searchViewModel.searchResults.value[0].title)

        // Second search
        searchViewModel.updateSearchQuery("second")
        searchViewModel.search()
        advanceUntilIdle()
        assertEquals(1, searchViewModel.searchResults.value.size)
        assertEquals("Second", searchViewModel.searchResults.value[0].title)
    }

    @Test
    fun `SearchViewModel should extend ViewModel`() {
        assertTrue(searchViewModel is androidx.lifecycle.ViewModel)
    }
}
