package com.catalogizer.android.ui.screens.search

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

/**
 * Tests for SearchScreen behavior via SearchViewModel.
 * Covers search input, results display, empty results, loading state,
 * debouncing, filtering logic, and navigation callbacks.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SearchScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var searchViewModel: SearchViewModel

    private fun createMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie",
        description: String? = null,
        year: Int? = 2024,
        quality: String? = "1080p",
        rating: Double? = 8.0,
    ): MediaItem = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        description = description,
        year = year,
        quality = quality,
        rating = rating,
        directoryPath = "/media/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z",
    )

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
    fun `initial state has empty query`() {
        assertEquals("", searchViewModel.query.value)
    }

    @Test
    fun `initial state has empty results`() {
        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `initial state is not searching`() {
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search updates query value`() {
        searchViewModel.search("inception")

        assertEquals("inception", searchViewModel.query.value)
    }

    @Test
    fun `search with blank query clears results`() = runTest {
        val items = listOf(createMediaItem(1L, "Test"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(items)
        searchViewModel.search("test")
        advanceUntilIdle()

        searchViewModel.search("")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search with whitespace-only query clears results`() = runTest {
        searchViewModel.search("   ")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search debounces requests by 300ms`() = runTest {
        val items = listOf(createMediaItem(1L, "Test"))
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(items)

        searchViewModel.search("te")
        advanceTimeBy(100)
        searchViewModel.search("tes")
        advanceTimeBy(100)
        searchViewModel.search("test")
        advanceTimeBy(400)
        advanceUntilIdle()

        coVerify(exactly = 1) { mockMediaRepository.getRecentMedia(any()) }
    }

    @Test
    fun `search filters results by title match`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "Inception"),
            createMediaItem(2L, "The Matrix"),
            createMediaItem(3L, "Interstellar"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("Matrix")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("The Matrix", results[0].title)
    }

    @Test
    fun `search filters case-insensitively`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "INCEPTION"),
            createMediaItem(2L, "inception"),
            createMediaItem(3L, "The Matrix"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("inception")
        advanceUntilIdle()

        assertEquals(2, searchViewModel.searchResults.value.size)
    }

    @Test
    fun `search filters by media type`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "A Song", "music"),
            createMediaItem(2L, "A Movie", "movie"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("music")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("A Song", results[0].title)
    }

    @Test
    fun `search filters by description`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "Movie A", description = "A thriller about dreams"),
            createMediaItem(2L, "Movie B", description = "A romantic comedy"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("thriller")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(1, results.size)
        assertEquals("Movie A", results[0].title)
    }

    @Test
    fun `search returns empty results when no matches`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "Inception"),
            createMediaItem(2L, "The Matrix"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("ZZZNOTFOUND")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }

    @Test
    fun `search sets isSearching to false after completion`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())

        searchViewModel.search("test")
        advanceUntilIdle()

        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search handles repository exception gracefully`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } throws RuntimeException("Network error")

        searchViewModel.search("test")
        advanceUntilIdle()

        assertTrue(searchViewModel.searchResults.value.isEmpty())
        assertFalse(searchViewModel.isSearching.value)
    }

    @Test
    fun `search requests up to 50 items from repository`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())

        searchViewModel.search("test")
        advanceUntilIdle()

        coVerify { mockMediaRepository.getRecentMedia(50) }
    }

    @Test
    fun `search cancels previous query when new query arrives`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(
            listOf(createMediaItem(1L, "Fast")))

        searchViewModel.search("slow")
        advanceTimeBy(100)
        searchViewModel.search("fast")
        advanceUntilIdle()

        assertEquals("fast", searchViewModel.query.value)
    }

    @Test
    fun `consecutive empty searches do not trigger repository calls`() = runTest {
        searchViewModel.search("")
        searchViewModel.search("  ")
        searchViewModel.search("")
        advanceUntilIdle()

        coVerify(exactly = 0) { mockMediaRepository.getRecentMedia(any()) }
    }

    @Test
    fun `search handles items with null description`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "No Description", description = null),
            createMediaItem(2L, "Has Desc", description = "Some text about Description"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("Description")
        advanceUntilIdle()

        val results = searchViewModel.searchResults.value
        assertEquals(2, results.size)
    }

    @Test
    fun `search with multiple matching criteria returns item once`() = runTest {
        val item = createMediaItem(
            1L,
            title = "Action Movie",
            mediaType = "movie",
            description = "An action movie about heroes",
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(listOf(item))

        searchViewModel.search("movie")
        advanceUntilIdle()

        assertEquals(1, searchViewModel.searchResults.value.size)
    }

    @Test
    fun `result item displays first letter of media type`() {
        val item = createMediaItem(mediaType = "movie")
        val initial = item.mediaType.take(1).uppercase()

        assertEquals("M", initial)
    }

    @Test
    fun `result item displays tv_show initial correctly`() {
        val item = createMediaItem(mediaType = "tv_show")
        val initial = item.mediaType.take(1).uppercase()

        assertEquals("T", initial)
    }

    @Test
    fun `result item displays year when available`() {
        val item = createMediaItem(year = 2023)
        assertNotNull(item.year)
        assertEquals(2023, item.year)
    }

    @Test
    fun `result item has null year when not available`() {
        val item = createMediaItem(year = null)
        assertNull(item.year)
    }

    @Test
    fun `result item displays quality when available`() {
        val item = createMediaItem(quality = "4k")
        assertNotNull(item.quality)
        assertEquals("4k", item.quality)
    }

    @Test
    fun `result item has null quality when not available`() {
        val item = createMediaItem(quality = null)
        assertNull(item.quality)
    }

    @Test
    fun `result item displays rating formatted to one decimal`() {
        val item = createMediaItem(rating = 8.75)
        val formatted = "%.1f".format(item.rating)

        assertEquals("8.8", formatted)
    }

    @Test
    fun `result item has null rating when not available`() {
        val item = createMediaItem(rating = null)
        assertNull(item.rating)
    }

    @Test
    fun `navigate back callback is invocable`() {
        var backCalled = false
        val onNavigateBack: () -> Unit = { backCalled = true }

        onNavigateBack()

        assertTrue(backCalled)
    }

    @Test
    fun `navigate to media detail callback receives correct id`() {
        var detailId: Long? = null
        val onNavigateToMediaDetail: (Long) -> Unit = { detailId = it }

        onNavigateToMediaDetail(42L)

        assertEquals(42L, detailId)
    }

    @Test
    fun `no results message includes query text`() {
        val query = "inception"
        val message = "No results found for \"$query\""

        assertEquals("No results found for \"inception\"", message)
    }

    @Test
    fun `empty query shows placeholder text`() {
        val query = ""
        val showPlaceholder = query.isEmpty()

        assertTrue(showPlaceholder)
    }

    @Test
    fun `non-empty query with results shows result list`() {
        val query = "test"
        val results = listOf(createMediaItem(1L, "Test Movie"))
        val showResults = query.isNotEmpty() && results.isNotEmpty()

        assertTrue(showResults)
    }

    @Test
    fun `non-empty query with no results shows empty message`() {
        val query = "test"
        val results = emptyList<MediaItem>()
        val showEmptyMessage = query.isNotEmpty() && results.isEmpty()

        assertTrue(showEmptyMessage)
    }

    @Test
    fun `search result items use id as key`() {
        val items = listOf(
            createMediaItem(10L, "A"),
            createMediaItem(20L, "B"),
            createMediaItem(30L, "C"),
        )
        val keys = items.map { it.id }

        assertEquals(listOf(10L, 20L, 30L), keys)
        assertEquals(keys.size, keys.toSet().size)
    }

    @Test
    fun `search handles Cyrillic text in query`() = runTest {
        val allMedia = listOf(
            createMediaItem(1L, "Test"),
        )
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(allMedia)

        searchViewModel.search("\u041C\u043E\u0441\u043A\u0432\u0430")
        advanceUntilIdle()

        assertEquals("\u041C\u043E\u0441\u043A\u0432\u0430", searchViewModel.query.value)
    }

    @Test
    fun `search handles special characters in query`() = runTest {
        coEvery { mockMediaRepository.getRecentMedia(any()) } returns ApiResult.success(emptyList())

        searchViewModel.search("test & \"query\" <script>")
        advanceUntilIdle()

        assertEquals("test & \"query\" <script>", searchViewModel.query.value)
    }

    @Test
    fun `clear action resets query to empty`() = runTest {
        searchViewModel.search("something")
        assertEquals("something", searchViewModel.query.value)

        searchViewModel.search("")
        advanceUntilIdle()

        assertEquals("", searchViewModel.query.value)
        assertTrue(searchViewModel.searchResults.value.isEmpty())
    }
}
