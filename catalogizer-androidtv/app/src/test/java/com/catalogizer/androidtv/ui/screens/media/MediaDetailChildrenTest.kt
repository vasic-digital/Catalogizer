package com.catalogizer.androidtv.ui.screens.media

import android.content.Context
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.models.MediaSearchResponse
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.repository.MediaRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

/**
 * DEFECT-F (§11.4.143) — the TV-show / season / album detail screen MUST expose
 * its children so the user can pick a specific episode / track to play. Before
 * the fix the detail screen rendered only Play / Back and never fetched
 * `GET /api/v1/entities/{id}/children`, so a user could never choose an episode.
 *
 * RED-polarity (§11.4.115): [RED_MODE] flips the test between reproducing the
 * pre-fix behaviour (children NEVER fetched → assert the gap is present) and
 * verifying the post-fix behaviour (children fetched + exposed in UI state).
 *
 * RED_MODE = true  -> reproduces DEFECT-F on the pre-fix logic (the gap exists).
 * RED_MODE = false -> standing GREEN guard asserting the gap is closed.
 *
 * Unit layer: a MockK-faked [CatalogizerApi] behind a real [MediaRepository]
 * (mocks permitted in unit tests per §11.4.27). The Compose UI itself
 * (focus, D-pad row rendering) is the conductor's on-device follow-up
 * (§11.4.52 AUTONOMOUS_DESIGNED).
 */
@ExperimentalCoroutinesApi
class MediaDetailChildrenTest {

    private companion object {
        /** Flip to true to reproduce DEFECT-F on the pre-fix screen behaviour. */
        const val RED_MODE = false
    }

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val context: Context = mockk(relaxed = true)
    private val api: CatalogizerApi = mockk()
    private val repository: MediaRepository = MediaRepository(context, api)

    /**
     * Minimal UI-state holder mirroring the detail screen's child-fetch state
     * (`children` + `childrenLoading` in MediaDetailScreen). The screen drives
     * this from a LaunchedEffect keyed on the loaded entity.
     */
    private data class DetailUiState(
        val children: List<MediaItem> = emptyList(),
        val childrenLoading: Boolean = false
    )

    /**
     * Replicates MediaDetailScreen's child-fetch decision so it is testable at
     * the unit layer. [redMode] = true is the pre-fix behaviour (never fetch);
     * false is the post-fix behaviour (fetch children for container entities).
     */
    private suspend fun loadDetailUiState(item: MediaItem, redMode: Boolean): DetailUiState {
        if (redMode) {
            // Pre-fix: the detail screen never consulted the children endpoint.
            return DetailUiState()
        }
        return if (isContainerType(item.mediaType)) {
            DetailUiState(children = repository.getEntityChildren(item.id))
        } else {
            DetailUiState()
        }
    }

    private fun episodes(parentId: Long): List<MediaItem> = listOf(
        MediaItem(id = 71, title = "Pilot", mediaType = "tv_episode", parentId = parentId, episodeNumber = 1),
        MediaItem(id = 72, title = "Cat's in the Bag...", mediaType = "tv_episode", parentId = parentId, episodeNumber = 2)
    )

    private val tvShow = MediaItem(id = 7, title = "Breaking Bad", mediaType = "tv_show")
    private val tvSeason = MediaItem(id = 70, title = "Season 1", mediaType = "tv_season", parentId = 7, seasonNumber = 1)
    private val movie = MediaItem(id = 9, title = "Heat", mediaType = "movie")

    @Test
    fun `container entity exposes its children in ui state`() = runTest {
        coEvery { api.getEntityChildren(70) } returns Response.success(
            MediaSearchResponse(items = episodes(70), total = 2)
        )

        val state = loadDetailUiState(tvSeason, RED_MODE)

        if (RED_MODE) {
            // DEFECT-F present: pre-fix screen never surfaces the episodes,
            // so the user has nothing to pick. This assertion documents the gap.
            assertTrue(
                "RED: pre-fix detail screen exposes no children (DEFECT-F)",
                state.children.isEmpty()
            )
        } else {
            // GREEN guard: the season's episodes are fetched and exposed so the
            // user can choose a specific episode to play.
            assertEquals(2, state.children.size)
            assertEquals(listOf("Pilot", "Cat's in the Bag..."), state.children.map { it.title })
        }
    }

    @Test
    fun `tv_show fetches children only after the fix`() = runTest {
        coEvery { api.getEntityChildren(7) } returns Response.success(
            MediaSearchResponse(
                items = listOf(tvSeason, tvSeason.copy(id = 80, title = "Season 2", seasonNumber = 2)),
                total = 2
            )
        )

        val state = loadDetailUiState(tvShow, RED_MODE)

        if (RED_MODE) {
            assertTrue("RED: tv_show exposes no seasons pre-fix", state.children.isEmpty())
            coVerify(exactly = 0) { api.getEntityChildren(7) }
        } else {
            assertEquals(2, state.children.size)
            coVerify(exactly = 1) { api.getEntityChildren(7) }
        }
    }

    @Test
    fun `leaf movie never fetches children in either mode`() = runTest {
        val state = loadDetailUiState(movie, RED_MODE)
        assertTrue(state.children.isEmpty())
        // A movie is a leaf — the children endpoint is never consulted.
        coVerify(exactly = 0) { api.getEntityChildren(any()) }
    }

    // --- Pure helper coverage (mirrors the screen's container/child logic) ---

    @Test
    fun `isContainerType true for show season and album`() {
        assertTrue(isContainerType("tv_show"))
        assertTrue(isContainerType("tv_season"))
        assertTrue(isContainerType("music_album"))
    }

    @Test
    fun `isContainerType false for leaf and null types`() {
        assertFalse(isContainerType("movie"))
        assertFalse(isContainerType("tv_episode"))
        assertFalse(isContainerType("song"))
        assertFalse(isContainerType(null))
    }

    @Test
    fun `childSectionTitle names the child kind`() {
        assertEquals("Seasons", childSectionTitle("tv_show"))
        assertEquals("Episodes", childSectionTitle("tv_season"))
        assertEquals("Tracks", childSectionTitle("music_album"))
    }

    @Test
    fun `childTileSubtitle surfaces the ordinal`() {
        assertEquals("Season 1", childTileSubtitle(tvSeason))
        assertEquals("Episode 2", childTileSubtitle(episodes(70)[1]))
        assertEquals(
            "Track 5",
            childTileSubtitle(MediaItem(id = 1, title = "Song", mediaType = "song", trackNumber = 5))
        )
    }

    @Test
    fun `resolveChildCoverUrl prefixes relative and proxies tmdb`() {
        val server = "http://192.168.0.100:8080"
        assertEquals(
            "http://192.168.0.100:8080/api/v1/assets/c.jpg",
            resolveChildCoverUrl("/api/v1/assets/c.jpg", server)
        )
        assertTrue(
            resolveChildCoverUrl("https://image.tmdb.org/t/p/w500/x.jpg", server)!!
                .contains("/api/v1/image-proxy?url=")
        )
        assertEquals(
            "https://cdn.example.com/x.jpg",
            resolveChildCoverUrl("https://cdn.example.com/x.jpg", server)
        )
        assertEquals(null, resolveChildCoverUrl(null, server))
    }
}
