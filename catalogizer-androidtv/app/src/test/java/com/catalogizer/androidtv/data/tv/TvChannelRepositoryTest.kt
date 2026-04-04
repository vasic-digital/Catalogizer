package com.catalogizer.androidtv.data.tv

import android.content.Context
import com.catalogizer.androidtv.data.models.MediaItem
import com.catalogizer.androidtv.data.repository.MediaRepository
import com.catalogizer.androidtv.data.repository.SettingsRepository
import io.mockk.*
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
class TvChannelRepositoryTest {

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
    fun `buildDefaultChannelContent combines continue watching, recent, and trending`() = runTest {
        val continueWatching = listOf(
            MediaItem(id = 1L, title = "Continue 1", watchProgress = 0.5, mediaType = "movie")
        )
        val recent = listOf(
            MediaItem(id = 2L, title = "Recent 1", mediaType = "movie")
        )
        val trending = listOf(
            MediaItem(id = 3L, title = "Trending 1", mediaType = "tv_show")
        )
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(continueWatching)
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(recent)
        coEvery { mediaRepository.getTrendingItems(any()) } returns trending

        val items = repo.buildDefaultChannelContent()
        assertTrue(items.isNotEmpty())
        assertTrue(items.size <= 30)
    }

    @Test
    fun `buildDefaultChannelContent caps at 30 items`() = runTest {
        val manyItems = (1L..50L).map {
            MediaItem(id = it, title = "Item $it", mediaType = "movie")
        }
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(emptyList())
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(manyItems)
        coEvery { mediaRepository.getTrendingItems(any()) } returns emptyList()

        val items = repo.buildDefaultChannelContent()
        assertTrue(items.size <= 30)
    }

    @Test
    fun `buildDefaultChannelContent deduplicates items by ID`() = runTest {
        val item = MediaItem(id = 1L, title = "Duplicate", mediaType = "movie", watchProgress = 0.5)
        coEvery { mediaRepository.searchMedia(any()) } returns flowOf(listOf(item))
        coEvery { mediaRepository.browseEntities(any(), any(), any(), any()) } returns flowOf(listOf(item))
        coEvery { mediaRepository.getTrendingItems(any()) } returns listOf(item)

        val items = repo.buildDefaultChannelContent()
        assertEquals(1, items.count { it.id == 1L })
    }

    @Test
    fun `buildCategoryContent fetches items for given type`() = runTest {
        val movies = listOf(
            MediaItem(id = 1L, title = "Movie 1", mediaType = "movie"),
            MediaItem(id = 2L, title = "Movie 2", mediaType = "movie")
        )
        coEvery { mediaRepository.browseEntities("movie", 30, "created", "desc") } returns flowOf(movies)

        val items = repo.buildCategoryContent("movie")
        assertEquals(2, items.size)
        assertEquals("Movie 1", items[0].title)
    }

    @Test
    fun `buildCategoryContent caps at 30 items`() = runTest {
        val manyItems = (1L..50L).map {
            MediaItem(id = it, title = "Movie $it", mediaType = "movie")
        }
        coEvery { mediaRepository.browseEntities("movie", 30, "created", "desc") } returns flowOf(manyItems)

        val items = repo.buildCategoryContent("movie")
        assertTrue(items.size <= 30)
    }

    @Test
    fun `getActiveMediaTypes returns only types with content`() = runTest {
        coEvery { mediaRepository.getEntityStats() } returns Pair(100, mapOf(
            "movie" to 50,
            "tv_show" to 30,
            "music" to 0,
            "game" to 20
        ))

        val types = repo.getActiveMediaTypes()
        assertTrue(types.contains("movie"))
        assertTrue(types.contains("tv_show"))
        assertTrue(types.contains("game"))
        assertFalse(types.contains("music"))
    }

    @Test
    fun `getActiveMediaTypes returns empty list on API failure`() = runTest {
        coEvery { mediaRepository.getEntityStats() } returns Pair(0, emptyMap())

        val types = repo.getActiveMediaTypes()
        assertTrue(types.isEmpty())
    }
}
