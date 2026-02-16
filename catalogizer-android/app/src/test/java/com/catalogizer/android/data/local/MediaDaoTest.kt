package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.MediaItem
import io.mockk.clearAllMocks
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
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
class MediaDaoTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var database: CatalogizerDatabase
    private lateinit var mediaDao: MediaDao

    private fun createMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie",
        year: Int? = 2024,
        description: String? = "A test media item",
        rating: Double? = 8.5,
        quality: String? = "1080p",
        fileSize: Long? = 1024000L,
        directoryPath: String = "/media/movies",
        isFavorite: Boolean = false,
        watchProgress: Double = 0.0,
        isDownloaded: Boolean = false,
        createdAt: String = "2024-01-01T00:00:00Z",
        updatedAt: String = "2024-01-01T00:00:00Z"
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            year = year,
            description = description,
            rating = rating,
            quality = quality,
            fileSize = fileSize,
            directoryPath = directoryPath,
            createdAt = createdAt,
            updatedAt = updatedAt,
            isFavorite = isFavorite,
            watchProgress = watchProgress,
            isDownloaded = isDownloaded
        )
    }

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        )
            .allowMainThreadQueries()
            .build()
        mediaDao = database.mediaDao()
    }

    @After
    fun tearDown() {
        database.close()
        clearAllMocks()
    }

    // --- Insert Tests ---

    @Test
    fun `insertMedia should store media item in database`() = runTest {
        val item = createMediaItem(id = 1L, title = "Inception")

        mediaDao.insertMedia(item)

        val retrieved = mediaDao.getMediaById(1L)
        assertNotNull(retrieved)
        assertEquals("Inception", retrieved?.title)
    }

    @Test
    fun `insertMedia with REPLACE strategy should update existing item`() = runTest {
        val original = createMediaItem(id = 1L, title = "Original")
        val updated = createMediaItem(id = 1L, title = "Updated")

        mediaDao.insertMedia(original)
        mediaDao.insertMedia(updated)

        val retrieved = mediaDao.getMediaById(1L)
        assertEquals("Updated", retrieved?.title)
    }

    @Test
    fun `insertAllMedia should store multiple items`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Movie 1"),
            createMediaItem(id = 2L, title = "Movie 2"),
            createMediaItem(id = 3L, title = "Movie 3")
        )

        mediaDao.insertAllMedia(items)

        val count = mediaDao.getCachedItemsCount()
        assertEquals(3, count)
    }

    // --- Query Tests ---

    @Test
    fun `getMediaById should return null for non-existent id`() = runTest {
        val result = mediaDao.getMediaById(999L)

        assertNull(result)
    }

    @Test
    fun `getMediaById should return correct item`() = runTest {
        val item = createMediaItem(id = 42L, title = "The Matrix")
        mediaDao.insertMedia(item)

        val result = mediaDao.getMediaById(42L)

        assertNotNull(result)
        assertEquals(42L, result?.id)
        assertEquals("The Matrix", result?.title)
        assertEquals("movie", result?.mediaType)
    }

    @Test
    fun `getMediaByIdFlow should emit item changes`() = runTest {
        val item = createMediaItem(id = 1L, title = "Flow Test")
        mediaDao.insertMedia(item)

        val result = mediaDao.getMediaByIdFlow(1L).first()

        assertNotNull(result)
        assertEquals("Flow Test", result?.title)
    }

    @Test
    fun `getMediaByIdFlow should emit null for missing item`() = runTest {
        val result = mediaDao.getMediaByIdFlow(999L).first()

        assertNull(result)
    }

    @Test
    fun `getRecentlyAdded should return items ordered by createdAt descending`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Old", createdAt = "2024-01-01T00:00:00Z"),
            createMediaItem(id = 2L, title = "New", createdAt = "2024-06-01T00:00:00Z"),
            createMediaItem(id = 3L, title = "Newest", createdAt = "2024-12-01T00:00:00Z")
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.getRecentlyAdded(10).first()

        assertEquals(3, result.size)
        assertEquals("Newest", result[0].title)
        assertEquals("New", result[1].title)
        assertEquals("Old", result[2].title)
    }

    @Test
    fun `getRecentlyAdded should respect limit parameter`() = runTest {
        val items = (1..20).map { i ->
            createMediaItem(id = i.toLong(), title = "Item $i", createdAt = "2024-01-${String.format("%02d", i)}T00:00:00Z")
        }
        mediaDao.insertAllMedia(items)

        val result = mediaDao.getRecentlyAdded(5).first()

        assertEquals(5, result.size)
    }

    @Test
    fun `getTopRated should return items ordered by rating descending`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Low", rating = 5.0),
            createMediaItem(id = 2L, title = "High", rating = 9.5),
            createMediaItem(id = 3L, title = "Medium", rating = 7.0)
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.getTopRated(10).first()

        assertEquals(3, result.size)
        assertEquals("High", result[0].title)
        assertEquals("Medium", result[1].title)
        assertEquals("Low", result[2].title)
    }

    @Test
    fun `getAllMediaTypes should return distinct media types`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, mediaType = "movie"),
            createMediaItem(id = 2L, mediaType = "music"),
            createMediaItem(id = 3L, mediaType = "movie"),
            createMediaItem(id = 4L, mediaType = "tv_show")
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.getAllMediaTypes().first()

        assertEquals(3, result.size)
        assertTrue(result.contains("movie"))
        assertTrue(result.contains("music"))
        assertTrue(result.contains("tv_show"))
    }

    @Test
    fun `getTotalCount should return correct count`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L),
            createMediaItem(id = 2L),
            createMediaItem(id = 3L)
        )
        mediaDao.insertAllMedia(items)

        val count = mediaDao.getTotalCount().first()

        assertEquals(3, count)
    }

    @Test
    fun `getTotalCount should return zero for empty database`() = runTest {
        val count = mediaDao.getTotalCount().first()

        assertEquals(0, count)
    }

    @Test
    fun `getCountByType should return correct count per type`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, mediaType = "movie"),
            createMediaItem(id = 2L, mediaType = "movie"),
            createMediaItem(id = 3L, mediaType = "music")
        )
        mediaDao.insertAllMedia(items)

        val movieCount = mediaDao.getCountByType("movie").first()
        val musicCount = mediaDao.getCountByType("music").first()

        assertEquals(2, movieCount)
        assertEquals(1, musicCount)
    }

    // --- Search Tests ---

    @Test
    fun `searchCached should find items by title`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Inception"),
            createMediaItem(id = 2L, title = "Interstellar"),
            createMediaItem(id = 3L, title = "The Matrix")
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.searchCached("Inter")

        assertEquals(1, result.size)
        assertEquals("Interstellar", result[0].title)
    }

    @Test
    fun `searchCached should find items by description`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Movie", description = "A thriller about dreams"),
            createMediaItem(id = 2L, title = "Show", description = "A comedy series")
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.searchCached("thriller")

        assertEquals(1, result.size)
        assertEquals("Movie", result[0].title)
    }

    @Test
    fun `searchCached should return empty list for no matches`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Inception"),
            createMediaItem(id = 2L, title = "The Matrix")
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.searchCached("XYZ_NOT_FOUND")

        assertTrue(result.isEmpty())
    }

    @Test
    fun `getByType should return items of specified type`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, title = "Movie 1", mediaType = "movie"),
            createMediaItem(id = 2L, title = "Song 1", mediaType = "music"),
            createMediaItem(id = 3L, title = "Movie 2", mediaType = "movie")
        )
        mediaDao.insertAllMedia(items)

        val movies = mediaDao.getByType("movie")

        assertEquals(2, movies.size)
        assertTrue(movies.all { it.mediaType == "movie" })
    }

    // --- Update Tests ---

    @Test
    fun `updateMedia should modify existing item`() = runTest {
        val original = createMediaItem(id = 1L, title = "Original Title")
        mediaDao.insertMedia(original)

        val updated = original.copy(title = "Updated Title")
        mediaDao.updateMedia(updated)

        val result = mediaDao.getMediaById(1L)
        assertEquals("Updated Title", result?.title)
    }

    @Test
    fun `updateFavoriteStatus should toggle favorite`() = runTest {
        val item = createMediaItem(id = 1L, isFavorite = false)
        mediaDao.insertMedia(item)

        mediaDao.updateFavoriteStatus(1L, true)

        val result = mediaDao.getMediaById(1L)
        assertTrue(result?.isFavorite == true)
    }

    @Test
    fun `updateFavoriteStatus should unset favorite`() = runTest {
        val item = createMediaItem(id = 1L, isFavorite = true)
        mediaDao.insertMedia(item)

        mediaDao.updateFavoriteStatus(1L, false)

        val result = mediaDao.getMediaById(1L)
        assertFalse(result?.isFavorite == true)
    }

    @Test
    fun `updateWatchProgress should update progress and lastWatched`() = runTest {
        val item = createMediaItem(id = 1L, watchProgress = 0.0)
        mediaDao.insertMedia(item)

        mediaDao.updateWatchProgress(1L, 0.5, "2024-06-15T12:00:00Z")

        val result = mediaDao.getMediaById(1L)
        assertEquals(0.5, result?.watchProgress ?: 0.0, 0.001)
        assertEquals("2024-06-15T12:00:00Z", result?.lastWatched)
    }

    @Test
    fun `updateDownloadStatus should set download status`() = runTest {
        val item = createMediaItem(id = 1L, isDownloaded = false)
        mediaDao.insertMedia(item)

        mediaDao.updateDownloadStatus(1L, true)

        val result = mediaDao.getMediaById(1L)
        assertTrue(result?.isDownloaded == true)
    }

    @Test
    fun `updateRating should set new rating`() = runTest {
        val item = createMediaItem(id = 1L, rating = 5.0)
        mediaDao.insertMedia(item)

        mediaDao.updateRating(1L, 9.0)

        val result = mediaDao.getMediaById(1L)
        assertEquals(9.0, result?.rating ?: 0.0, 0.001)
    }

    // --- Delete Tests ---

    @Test
    fun `deleteMedia should remove item from database`() = runTest {
        val item = createMediaItem(id = 1L)
        mediaDao.insertMedia(item)

        mediaDao.deleteMedia(item)

        val result = mediaDao.getMediaById(1L)
        assertNull(result)
    }

    @Test
    fun `deleteMediaById should remove item by id`() = runTest {
        val item = createMediaItem(id = 1L)
        mediaDao.insertMedia(item)

        mediaDao.deleteMediaById(1L)

        val result = mediaDao.getMediaById(1L)
        assertNull(result)
    }

    @Test
    fun `deleteById should remove item by id`() = runTest {
        val item = createMediaItem(id = 1L)
        mediaDao.insertMedia(item)

        mediaDao.deleteById(1L)

        val result = mediaDao.getMediaById(1L)
        assertNull(result)
    }

    @Test
    fun `deleteAllMedia should remove all items`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L),
            createMediaItem(id = 2L),
            createMediaItem(id = 3L)
        )
        mediaDao.insertAllMedia(items)

        mediaDao.deleteAllMedia()

        val count = mediaDao.getCachedItemsCount()
        assertEquals(0, count)
    }

    @Test
    fun `deleteOldMedia should remove items older than timestamp`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, updatedAt = "2023-01-01T00:00:00Z"),
            createMediaItem(id = 2L, updatedAt = "2024-06-01T00:00:00Z"),
            createMediaItem(id = 3L, updatedAt = "2025-01-01T00:00:00Z")
        )
        mediaDao.insertAllMedia(items)

        mediaDao.deleteOldMedia("2024-01-01T00:00:00Z")

        val remaining = mediaDao.getAllCached()
        assertEquals(2, remaining.size)
    }

    // --- Transaction Tests ---

    @Test
    fun `refreshMedia should delete all and insert new items`() = runTest {
        val oldItems = listOf(
            createMediaItem(id = 1L, title = "Old 1"),
            createMediaItem(id = 2L, title = "Old 2")
        )
        mediaDao.insertAllMedia(oldItems)

        val newItems = listOf(
            createMediaItem(id = 3L, title = "New 1"),
            createMediaItem(id = 4L, title = "New 2"),
            createMediaItem(id = 5L, title = "New 3")
        )
        mediaDao.refreshMedia(newItems)

        val allItems = mediaDao.getAllCached()
        assertEquals(3, allItems.size)
        assertNull(mediaDao.getMediaById(1L))
        assertNull(mediaDao.getMediaById(2L))
        assertNotNull(mediaDao.getMediaById(3L))
    }

    @Test
    fun `insertOrUpdate should insert new item`() = runTest {
        val item = createMediaItem(id = 1L, title = "New Item")

        mediaDao.insertOrUpdate(item)

        val result = mediaDao.getMediaById(1L)
        assertNotNull(result)
        assertEquals("New Item", result?.title)
    }

    @Test
    fun `insertOrUpdate should update existing item`() = runTest {
        val original = createMediaItem(id = 1L, title = "Original")
        mediaDao.insertMedia(original)

        val updated = original.copy(title = "Updated")
        mediaDao.insertOrUpdate(updated)

        val result = mediaDao.getMediaById(1L)
        assertEquals("Updated", result?.title)
    }

    // --- Cached Data Tests ---

    @Test
    fun `getAllCached should return all items`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L),
            createMediaItem(id = 2L),
            createMediaItem(id = 3L)
        )
        mediaDao.insertAllMedia(items)

        val result = mediaDao.getAllCached()

        assertEquals(3, result.size)
    }

    @Test
    fun `getCachedItemsCount should return correct count`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L),
            createMediaItem(id = 2L)
        )
        mediaDao.insertAllMedia(items)

        val count = mediaDao.getCachedItemsCount()

        assertEquals(2, count)
    }

    @Test
    fun `getById should return item by id`() = runTest {
        val item = createMediaItem(id = 1L, title = "Test")
        mediaDao.insertMedia(item)

        val result = mediaDao.getById(1L)

        assertNotNull(result)
        assertEquals("Test", result?.title)
    }

    @Test
    fun `getById should return null for missing item`() = runTest {
        val result = mediaDao.getById(999L)

        assertNull(result)
    }

    @Test
    fun `getTotalDownloadSize should return sum of downloaded file sizes`() = runTest {
        val items = listOf(
            createMediaItem(id = 1L, fileSize = 1000L, isDownloaded = true),
            createMediaItem(id = 2L, fileSize = 2000L, isDownloaded = true),
            createMediaItem(id = 3L, fileSize = 3000L, isDownloaded = false)
        )
        mediaDao.insertAllMedia(items)

        val totalSize = mediaDao.getTotalDownloadSize()

        assertEquals(3000L, totalSize)
    }

    // --- Media Item Properties Tests ---

    @Test
    fun `inserted item should preserve all fields`() = runTest {
        val item = createMediaItem(
            id = 1L,
            title = "Test Movie",
            mediaType = "movie",
            year = 2024,
            description = "Test description",
            rating = 8.5,
            quality = "4k",
            fileSize = 5000000L,
            directoryPath = "/media/test",
            isFavorite = true,
            watchProgress = 0.75,
            isDownloaded = true
        )
        mediaDao.insertMedia(item)

        val result = mediaDao.getMediaById(1L)

        assertNotNull(result)
        assertEquals("Test Movie", result?.title)
        assertEquals("movie", result?.mediaType)
        assertEquals(2024, result?.year)
        assertEquals("Test description", result?.description)
        assertEquals(8.5, result?.rating ?: 0.0, 0.001)
        assertEquals("4k", result?.quality)
        assertEquals(5000000L, result?.fileSize)
        assertEquals("/media/test", result?.directoryPath)
        assertTrue(result?.isFavorite == true)
        assertEquals(0.75, result?.watchProgress ?: 0.0, 0.001)
        assertTrue(result?.isDownloaded == true)
    }
}
