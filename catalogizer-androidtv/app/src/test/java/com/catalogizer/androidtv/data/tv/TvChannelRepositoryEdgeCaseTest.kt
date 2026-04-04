package com.catalogizer.androidtv.data.tv

import android.content.Context
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.RuntimeEnvironment
import org.robolectric.annotation.Config

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE, sdk = [28])
class TvChannelRepositoryEdgeCaseTest {

    private lateinit var context: Context
    private lateinit var mediaRepository: MediaRepository
    private lateinit var settingsRepository: SettingsRepository
    private lateinit var repo: TvChannelRepository

    @Before
    fun setup() {
        context = RuntimeEnvironment.getApplication()
        mediaRepository = mockk(relaxed = true)
        settingsRepository = mockk(relaxed = true)
        repo = TvChannelRepository(context, mediaRepository, settingsRepository)
    }

    @Test
    fun `buildDefaultChannelContent returns empty when all API calls fail`() = runTest {
        coEvery { mediaRepository.searchMedia(any()) } throws RuntimeException("Network error")
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } throws RuntimeException("Network error")
        coEvery { mediaRepository.getTrendingItems(any()) } throws RuntimeException("Network error")

        val items = repo.buildDefaultChannelContent()

        assertTrue(items.isEmpty())
    }

    @Test
    fun `buildDefaultChannelContent excludes items with progress below 5 percent`() = runTest {
        val belowThreshold = listOf(
            MediaItem(id = 1L, title = "Barely Started", watchProgress = 0.01, mediaType = "movie"),
            MediaItem(id = 2L, title = "Not Started", watchProgress = 0.0, mediaType = "movie"),
            MediaItem(id = 3L, title = "Just Under", watchProgress = 0.04, mediaType = "movie")
        )
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(belowThreshold)
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(emptyList())
        coEvery { mediaRepository.getTrendingItems(any()) } returns emptyList()

        val items = repo.buildDefaultChannelContent()

        // None of the continue-watching items should appear (all below 5%)
        // They would be filtered before being added from searchMedia
        // The browseEntities/trending calls return empty, so total should be 0
        assertTrue(items.none { it.id == 1L || it.id == 2L || it.id == 3L })
    }

    @Test
    fun `buildDefaultChannelContent excludes items with progress above 90 percent`() = runTest {
        val aboveThreshold = listOf(
            MediaItem(id = 1L, title = "Nearly Done", watchProgress = 0.91, mediaType = "movie"),
            MediaItem(id = 2L, title = "Completed", watchProgress = 1.0, mediaType = "movie"),
            MediaItem(id = 3L, title = "Exactly 90", watchProgress = 0.90, mediaType = "movie")
        )
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(aboveThreshold)
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(emptyList())
        coEvery { mediaRepository.getTrendingItems(any()) } returns emptyList()

        val items = repo.buildDefaultChannelContent()

        // All are filtered: >= 0.9 is not < 0.9
        assertTrue(items.none { it.id == 1L || it.id == 2L || it.id == 3L })
    }

    @Test
    fun `buildCategoryContent returns empty on API failure`() = runTest {
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } throws RuntimeException("API failure")

        val items = repo.buildCategoryContent("movie")

        assertTrue(items.isEmpty())
    }

    @Test
    fun `buildCategoryContent returns empty list for unknown media type`() = runTest {
        coEvery { mediaRepository.browseEntities("unknown_type", any(), any(), any()) } returns flowOf(emptyList())

        val items = repo.buildCategoryContent("unknown_type")

        assertTrue(items.isEmpty())
    }

    @Test
    fun `getActiveMediaTypes handles partial API response`() = runTest {
        coEvery { mediaRepository.getEntityStats() } returns Pair(
            75,
            mapOf(
                "movie" to 50,
                "tv_show" to 25,
                "music" to 0
            )
        )

        val types = repo.getActiveMediaTypes()

        assertTrue(types.contains("movie"))
        assertTrue(types.contains("tv_show"))
        assertFalse(types.contains("music"))
        assertEquals(2, types.size)
    }

    @Test
    fun `MAX_PROGRAMS_PER_CHANNEL is 30`() {
        assertEquals(30, TvChannelRepository.MAX_PROGRAMS_PER_CHANNEL)
    }

    @Test
    fun `DEFAULT_CHANNEL_KEY is default`() {
        assertEquals("default", TvChannelRepository.DEFAULT_CHANNEL_KEY)
    }

    @Test
    fun `DEFAULT_CHANNEL_DISPLAY_NAME is Catalogizer Picks`() {
        assertEquals("Catalogizer Picks", TvChannelRepository.DEFAULT_CHANNEL_DISPLAY_NAME)
    }
}
