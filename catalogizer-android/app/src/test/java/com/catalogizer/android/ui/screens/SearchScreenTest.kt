package com.catalogizer.android.ui.screens

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import com.catalogizer.android.ui.viewmodel.SearchViewModel
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
        mediaType: String = "movie",
        description: String? = null
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            description = description,
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
    fun `initial state should have empty query`() {
        assertEquals("", searchViewModel.query.value)
    }

    @Test
    fun `initial state should have empty search results`() {
        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `initial state should not be searching`() {
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search should update query value`() = runTest {
        searchViewModel.search("test query")

        assertEquals("test query", searchViewModel.query.value)
    }

    @Test
    fun `search with blank query should clear results`() = runTest {
        // First populate results
        val testData = listOf(createTestMediaItem(1L, "Item"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(testData)
        searchViewModel.search("item")
        advanceUntilIdle()

        // Then clear with blank query
        searchViewModel.search("")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search with blank query should set isSearching to false`() = runTest {
        searchViewModel.search("   ")

        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search should debounce requests by 300ms`() = runTest {
        val testData = listOf(createTestMediaItem(1L, "Test"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(testData)

        searchViewModel.search("te")
        advanceTimeBy(100)
        searchViewModel.search("tes")
        advanceTimeBy(100)
        searchViewModel.search("test")
        advanceTimeBy(400) // Total > 300ms since last search
        advanceUntilIdle()

        // Repository should only be called once (for the final query after debounce)
        coVerify(exactly = 1) { mockMediaRepository.getRecentMedia(any()) }
    }

    @Test
    fun `search should filter results by title match`() = runTest {
        val allMedia = listOf(
            createTestMediaItem(1L, "Inception", "movie"),
            createTestMediaItem(2L, "The Matrix", "movie"),
            createTestMediaItem(3L, "Interstellar", "movie")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("Matrix")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("The Matrix", results[0].title)
    }

    @Test
    fun `search should filter case-insensitively`() = runTest {
        val allMedia = listOf(
            createTestMediaItem(1L, "INCEPTION"),
            createTestMediaItem(2L, "inception"),
            createTestMediaItem(3L, "The Matrix")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("inception")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(2, results.size)
    }

    @Test
    fun `search should filter results by mediaType match`() = runTest {
        val allMedia = listOf(
            createTestMediaItem(1L, "A Song", "music"),
            createTestMediaItem(2L, "A Movie", "movie")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("music")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("A Song", results[0].title)
    }

    @Test
    fun `search should filter results by description match`() = runTest {
        val allMedia = listOf(
            createTestMediaItem(1L, "Movie A", "movie", description = "A thriller about dreams"),
            createTestMediaItem(2L, "Movie B", "movie", description = "A romantic comedy")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("thriller")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("Movie A", results[0].title)
    }

    @Test
    fun `search should return empty results when no matches`() = runTest {
        val allMedia = listOf(
            createTestMediaItem(1L, "Inception"),
            createTestMediaItem(2L, "The Matrix")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("XYZ_NOTFOUND")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `search should set isSearching to false after completion`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())

        searchViewModel.search("test")
        advanceUntilIdle()

        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search should handle repository exception gracefully`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException("Network error")

        searchViewModel.search("test")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search should request up to 50 items from repository`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())

        searchViewModel.search("test")
        advanceUntilIdle()

        coVerify { mockMediaRepository.getRecentMedia(50) }
    }

    @Test
    fun `search should cancel previous search when new query arrives`() = runTest {
        val slowData = listOf(createTestMediaItem(1L, "Slow"))
        val fastData = listOf(createTestMediaItem(2L, "Fast"))

        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(fastData)

        searchViewModel.search("slow")
        advanceTimeBy(100)  // Before debounce completes
        searchViewModel.search("fast")
        advanceUntilIdle()

        // Only the final search query should be active
        assertEquals("fast", searchViewModel.query.value)
    }

    @Test
    fun `consecutive empty searches should not trigger repository calls`() = runTest {
        searchViewModel.search("")
        searchViewModel.search("  ")
        searchViewModel.search("")
        advanceUntilIdle()

        coVerify(exactly = 0) { mockMediaRepository.getRecentMedia(any()) }
    }

    @Test
    fun `search results should handle items with null description`() = runTest {
        val allMedia = listOf(
            createTestMediaItem(1L, "No Description", description = null),
            createTestMediaItem(2L, "Has Description", description = "Some text")
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("Description")
        advanceUntilIdle()

        // Should only match the one with description containing "Description" in title
        val results = searchViewModel.searchResults.value
        assertEquals(2, results.size)
    }

    @Test
    fun `search with multiple matching criteria should return item once`() = runTest {
        val item = createTestMediaItem(
            1L,
            title = "Action Movie",
            mediaType = "movie",
            description = "An action movie about heroes"
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(listOf(item))

        searchViewModel.search("movie")
        advanceUntilIdle()

        // Item matches in title, mediaType, and description but should appear only once
        assertEquals(1, searchViewModel.searchResults.value.size)
    }
}
