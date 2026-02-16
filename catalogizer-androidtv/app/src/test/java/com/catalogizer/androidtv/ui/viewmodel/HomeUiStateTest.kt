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

        assertFalse(state.isLoading)
        assertNull(state.error)
        assertTrue(state.continueWatching.isEmpty())
        assertTrue(state.recentlyAdded.isEmpty())
        assertTrue(state.movies.isEmpty())
        assertTrue(state.tvShows.isEmpty())
        assertTrue(state.music.isEmpty())
        assertTrue(state.documents.isEmpty())
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
        val tvShows = listOf(createTestMediaItem(3, "Show 1", "tv_show"))
        val featured = createTestMediaItem(1, "Movie 1")

        val state = HomeUiState(
            movies = movies,
            tvShows = tvShows,
            featuredItem = featured
        )

        assertEquals(2, state.movies.size)
        assertEquals(1, state.tvShows.size)
        assertEquals("Movie 1", state.featuredItem?.title)
    }

    @Test
    fun `HomeUiState copy updates correctly`() {
        val initial = HomeUiState()
        val loading = initial.copy(isLoading = true)
        val loaded = loading.copy(
            isLoading = false,
            movies = listOf(createTestMediaItem())
        )

        assertFalse(initial.isLoading)
        assertTrue(loading.isLoading)
        assertFalse(loaded.isLoading)
        assertEquals(1, loaded.movies.size)
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
