package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

class HomeUiStateTest {

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Movie",
        mediaType: String = "movie",
        watchProgress: Double = 0.0
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        directoryPath = "/test",
        createdAt = "2024-01-01",
        updatedAt = "2024-01-01",
        watchProgress = watchProgress
    )

    @Test
    fun `HomeUiState has correct defaults`() {
        val state = HomeUiState()

        assertTrue(state.isLoading)
        assertNull(state.error)
        assertTrue(state.continueWatching.isEmpty())
        assertTrue(state.recentMovies.isEmpty())
        assertTrue(state.recommended.isEmpty())
        assertTrue(state.trending.isEmpty())
        assertTrue(state.topRatedMovies.isEmpty())
        assertTrue(state.topRatedTvShows.isEmpty())
        assertTrue(state.topRatedMusic.isEmpty())
        assertTrue(state.topRatedDocuments.isEmpty())
        assertTrue(state.recentComics.isEmpty())
        assertNull(state.featuredItem)
    }

    @Test
    fun `HomeUiState loading state`() {
        val state = HomeUiState(isLoading = true)
        assertTrue(state.isLoading)
    }

    @Test
    fun `HomeUiState error state`() {
        val state = HomeUiState(error = "Failed to load content")
        assertEquals("Failed to load content", state.error)
    }

    @Test
    fun `HomeUiState with content`() {
        val movies = listOf(createTestMediaItem(1, "Movie 1"), createTestMediaItem(2, "Movie 2"))
        val featured = createTestMediaItem(1, "Movie 1")

        val state = HomeUiState(
            topRatedMovies = movies,
            featuredItem = featured
        )

        assertEquals(2, state.topRatedMovies.size)
        assertEquals("Movie 1", state.featuredItem?.title)
    }

    @Test
    fun `HomeUiState copy updates correctly`() {
        // Default is isLoading=true (TICKET-002 fix — spinner from first frame)
        val initial = HomeUiState()
        val loaded = initial.copy(
            isLoading = false,
            topRatedMovies = listOf(createTestMediaItem())
        )

        assertTrue(initial.isLoading)
        assertFalse(loaded.isLoading)
        assertEquals(1, loaded.topRatedMovies.size)
    }

    @Test
    fun `HomeUiState equality works correctly`() {
        val state1 = HomeUiState(isLoading = true)
        val state2 = HomeUiState(isLoading = true)
        val state3 = HomeUiState(isLoading = false)

        assertEquals(state1, state2)
        assertNotEquals(state1, state3)
    }

    @Test
    fun `HomeUiState with continue watching items`() {
        val continueWatching = listOf(
            createTestMediaItem(1, "Movie 1", watchProgress = 0.5),
            createTestMediaItem(2, "Movie 2", watchProgress = 0.3)
        )

        val state = HomeUiState(continueWatching = continueWatching)

        assertEquals(2, state.continueWatching.size)
        assertEquals(0.5, state.continueWatching[0].watchProgress, 0.01)
    }
}
