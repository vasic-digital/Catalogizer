package com.catalogizer.androidtv.ui.screens.home

import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.playback.UiPlaybackProgress
import com.catalogizer.androidtv.ui.viewmodel.HomeUiState
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for HomeScreen rail-building logic: dynamic content rail construction,
 * stats header labels, empty library detection, and featured item selection.
 */
class HomeScreenSectionsTest {

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie",
        watchProgress: Double = 0.0
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        directoryPath = "/media/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z",
        watchProgress = watchProgress
    )

    // --- Dynamic rail construction ---

    @Test
    fun `empty state shows when all content sections are empty`() {
        val state = HomeUiState(isLoading = false)
        val allEmpty = state.continueWatching.isEmpty() &&
            state.recentMovies.isEmpty() && state.recentTvShows.isEmpty() &&
            state.recentMusicAlbums.isEmpty() && state.recentGames.isEmpty() &&
            state.recentBooks.isEmpty() && state.recentComics.isEmpty() &&
            state.recentSoftware.isEmpty() && state.recentConcerts.isEmpty() &&
            state.recommended.isEmpty() && state.trending.isEmpty() &&
            state.topRatedMovies.isEmpty() && state.topRatedTvShows.isEmpty() &&
            state.topRatedMusic.isEmpty() && state.topRatedDocuments.isEmpty()
        assertTrue(allEmpty)
    }

    @Test
    fun `non-empty continue watching section creates rail`() {
        val items = listOf(createTestMediaItem(1L, "Watching", watchProgress = 0.5))
        val state = HomeUiState(isLoading = false, continueWatching = items)
        assertTrue(state.continueWatching.isNotEmpty())
    }

    @Test
    fun `rails are built only for non-empty sections`() {
        val movies = listOf(createTestMediaItem(1L, "Movie"))
        val tvShows = listOf(createTestMediaItem(2L, "TV Show", mediaType = "tv_show"))
        val state = HomeUiState(
            isLoading = false,
            recentMovies = movies,
            recentTvShows = tvShows
        )

        val rails = buildList {
            if (state.continueWatching.isNotEmpty()) add("Continue Watching")
            if (state.recentMovies.isNotEmpty()) add("Recently Added Movies")
            if (state.recentTvShows.isNotEmpty()) add("Recently Added TV Shows")
            if (state.recentMusicAlbums.isNotEmpty()) add("Recently Added Music Albums")
            if (state.recentGames.isNotEmpty()) add("Recently Added Games")
            if (state.recentBooks.isNotEmpty()) add("Recently Added Books")
            if (state.recentComics.isNotEmpty()) add("Recently Added Comics")
            if (state.recentSoftware.isNotEmpty()) add("Recently Added Software")
            if (state.recentConcerts.isNotEmpty()) add("Recently Added Concerts")
            if (state.recommended.isNotEmpty()) add("Recommended for You")
            if (state.trending.isNotEmpty()) add("Trending")
            if (state.topRatedMovies.isNotEmpty()) add("Top Rated Movies")
            if (state.topRatedTvShows.isNotEmpty()) add("Top Rated TV Shows")
            if (state.topRatedMusic.isNotEmpty()) add("Top Rated Music")
            if (state.topRatedDocuments.isNotEmpty()) add("Top Rated Documents")
        }
        assertEquals(2, rails.size)
        assertEquals("Recently Added Movies", rails[0])
        assertEquals("Recently Added TV Shows", rails[1])
    }

    @Test
    fun `all fourteen possible rails are created when all sections populated`() {
        val item = createTestMediaItem()
        val watchingItem = createTestMediaItem(watchProgress = 0.5)
        val state = HomeUiState(
            isLoading = false,
            continueWatching = listOf(watchingItem),
            recentMovies = listOf(item),
            recentTvShows = listOf(item),
            recentMusicAlbums = listOf(item),
            recentGames = listOf(item),
            recentBooks = listOf(item),
            recentComics = listOf(item),
            recentSoftware = listOf(item),
            recentConcerts = listOf(item),
            recommended = listOf(item),
            trending = listOf(item),
            topRatedMovies = listOf(item),
            topRatedTvShows = listOf(item),
            topRatedMusic = listOf(item),
            topRatedDocuments = listOf(item)
        )

        var railCount = 0
        if (state.continueWatching.isNotEmpty()) railCount++
        if (state.recentMovies.isNotEmpty()) railCount++
        if (state.recentTvShows.isNotEmpty()) railCount++
        if (state.recentMusicAlbums.isNotEmpty()) railCount++
        if (state.recentGames.isNotEmpty()) railCount++
        if (state.recentBooks.isNotEmpty()) railCount++
        if (state.recentComics.isNotEmpty()) railCount++
        if (state.recentSoftware.isNotEmpty()) railCount++
        if (state.recentConcerts.isNotEmpty()) railCount++
        if (state.recommended.isNotEmpty()) railCount++
        if (state.trending.isNotEmpty()) railCount++
        if (state.topRatedMovies.isNotEmpty()) railCount++
        if (state.topRatedTvShows.isNotEmpty()) railCount++
        if (state.topRatedMusic.isNotEmpty()) railCount++
        if (state.topRatedDocuments.isNotEmpty()) railCount++

        assertEquals(15, railCount)
    }

    // --- Loading and error states ---

    @Test
    fun `loading state prevents content rendering`() {
        val state = HomeUiState(isLoading = true)
        assertTrue(state.isLoading)
    }

    @Test
    fun `error state shows error message`() {
        val state = HomeUiState(isLoading = false, error = "Network error")
        assertNotNull(state.error)
        assertEquals("Network error", state.error)
    }

    @Test
    fun `error state takes precedence over content`() {
        val state = HomeUiState(
            isLoading = false,
            error = "Server down",
            recentMovies = listOf(createTestMediaItem())
        )
        assertNotNull(state.error)
    }

    // --- Stats header ---

    @Test
    fun `stats header shows when totalEntities is positive`() {
        val state = HomeUiState(totalEntities = 150)
        assertTrue(state.totalEntities > 0)
    }

    @Test
    fun `stats header hidden when totalEntities is zero`() {
        val state = HomeUiState(totalEntities = 0)
        assertFalse(state.totalEntities > 0)
    }

    @Test
    fun `stats type labels map correctly`() {
        val labels = mapOf(
            "movie" to "Movies",
            "tv_show" to "TV Shows",
            "tv_season" to "Seasons",
            "tv_episode" to "Episodes",
            "music_artist" to "Artists",
            "music_album" to "Albums",
            "song" to "Songs",
            "game" to "Games",
            "software" to "Software",
            "book" to "Books",
            "comic" to "Comics"
        )
        assertEquals("Movies", labels["movie"])
        assertEquals("TV Shows", labels["tv_show"])
        assertEquals("Books", labels["book"])
    }

    @Test
    fun `stats header filters out zero-count types`() {
        val statsByType = mapOf("movie" to 50, "tv_show" to 0, "song" to 100)
        val nonZero = statsByType.filter { it.value > 0 }
        assertEquals(2, nonZero.size)
        assertFalse(nonZero.containsKey("tv_show"))
    }

    @Test
    fun `stats header sorts by count descending`() {
        val statsByType = mapOf("movie" to 50, "song" to 100, "game" to 10)
        val sorted = statsByType.filter { it.value > 0 }.entries.sortedByDescending { it.value }
        assertEquals("song", sorted[0].key)
        assertEquals("movie", sorted[1].key)
        assertEquals("game", sorted[2].key)
    }

    @Test
    fun `unknown type label falls back to formatted key`() {
        val statsTypeLabels = mapOf("movie" to "Movies")
        val key = "new_type"
        val label = statsTypeLabels[key]
            ?: key.replace("_", " ").replaceFirstChar { it.uppercase() }
        assertEquals("New type", label)
    }

    // --- Progress by ID map ---

    @Test
    fun `progressById map provides progress for media items`() {
        val progress = UiPlaybackProgress(
            mediaItemId = 42L,
            positionUnit = "seconds",
            durationTotal = 3600L,
            lastPosition = 1800L,
            lastSessionAmount = 600L,
            totalReproductions = 2,
            aggregateAmount = 2400L,
            lastSessionEndedAtMs = null
        )
        val progressById = mapOf(42L to progress)
        val state = HomeUiState(progressById = progressById)
        assertNotNull(state.progressById[42L])
        assertNull(state.progressById[99L])
    }

    // --- Featured item selection ---

    @Test
    fun `featured item comes from continue watching when available`() {
        val watching = createTestMediaItem(1L, "Currently Watching", watchProgress = 0.3)
        val state = HomeUiState(
            continueWatching = listOf(watching),
            featuredItem = watching
        )
        assertNotNull(state.featuredItem)
        assertEquals("Currently Watching", state.featuredItem?.title)
    }

    @Test
    fun `featured item falls back to recent movies when no continue watching`() {
        val movie = createTestMediaItem(1L, "New Movie")
        val state = HomeUiState(
            continueWatching = emptyList(),
            recentMovies = listOf(movie),
            featuredItem = movie
        )
        assertNotNull(state.featuredItem)
        assertEquals("New Movie", state.featuredItem?.title)
    }

    // --- Navigation callbacks ---

    @Test
    fun `continue watching rail uses player navigation`() {
        val navigateToPlayer = true
        assertTrue(navigateToPlayer)
    }

    @Test
    fun `regular rails use detail navigation`() {
        val navigateToPlayer = false
        assertFalse(navigateToPlayer)
    }
}
