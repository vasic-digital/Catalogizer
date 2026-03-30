package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

class HomeUiStateTest2 {

    private val testItem = MediaItem(
        id = 1L, title = "Movie", directoryPath = "/path",
        createdAt = "2024-01-01", updatedAt = "2024-01-01"
    )

    @Test
    fun `default HomeUiState has correct defaults`() {
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
        assertEquals(0, state.totalEntities)
        assertTrue(state.statsByType.isEmpty())
    }

    @Test
    fun `HomeUiState with loading state`() {
        val state = HomeUiState(isLoading = true)
        assertTrue(state.isLoading)
    }

    @Test
    fun `HomeUiState with error`() {
        val state = HomeUiState(error = "Network error")
        assertEquals("Network error", state.error)
    }

    @Test
    fun `HomeUiState with movies`() {
        val items = listOf(testItem, testItem.copy(id = 2L))
        val state = HomeUiState(movies = items)
        assertEquals(2, state.movies.size)
    }

    @Test
    fun `HomeUiState with featured item`() {
        val state = HomeUiState(featuredItem = testItem)
        assertNotNull(state.featuredItem)
        assertEquals("Movie", state.featuredItem?.title)
    }

    @Test
    fun `HomeUiState with stats`() {
        val stats = mapOf("movie" to 100, "tv_show" to 50)
        val state = HomeUiState(totalEntities = 150, statsByType = stats)
        assertEquals(150, state.totalEntities)
        assertEquals(100, state.statsByType["movie"])
    }

    @Test
    fun `HomeUiState copy updates loading`() {
        val initial = HomeUiState()
        val loading = initial.copy(isLoading = true)
        val done = loading.copy(isLoading = false, movies = listOf(testItem))

        assertFalse(initial.isLoading)
        assertTrue(loading.isLoading)
        assertFalse(done.isLoading)
        assertEquals(1, done.movies.size)
    }

    @Test
    fun `HomeUiState category-specific rails default to empty`() {
        val state = HomeUiState()
        assertTrue(state.recentMovies.isEmpty())
        assertTrue(state.recentMusicAlbums.isEmpty())
        assertTrue(state.recentTvShows.isEmpty())
        assertTrue(state.recentGames.isEmpty())
        assertTrue(state.recentConcerts.isEmpty())
        assertTrue(state.recentBooks.isEmpty())
        assertTrue(state.recentSoftware.isEmpty())
    }

    @Test
    fun `HomeUiState recently played rails default to empty`() {
        val state = HomeUiState()
        assertTrue(state.playedMovies.isEmpty())
        assertTrue(state.playedMusicAlbums.isEmpty())
        assertTrue(state.playedTvShows.isEmpty())
        assertTrue(state.playedGames.isEmpty())
        assertTrue(state.playedConcerts.isEmpty())
    }

    // --- ContentRail ---

    @Test
    fun `ContentRail has correct defaults`() {
        val rail = ContentRail(title = "Recently Added")
        assertEquals("Recently Added", rail.title)
        assertTrue(rail.items.isEmpty())
    }

    @Test
    fun `ContentRail with items`() {
        val rail = ContentRail(
            title = "Movies",
            items = listOf(testItem, testItem.copy(id = 2L))
        )
        assertEquals("Movies", rail.title)
        assertEquals(2, rail.items.size)
    }
}
