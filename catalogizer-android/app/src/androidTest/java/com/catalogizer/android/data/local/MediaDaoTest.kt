package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.catalogizer.android.data.models.MediaItem
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * Instrumentation tests for MediaDao
 */
@RunWith(AndroidJUnit4::class)
class MediaDaoTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var mediaDao: MediaDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        mediaDao = database.mediaDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun createTestMediaItem(
        id: Long = 1,
        title: String = "Test Movie",
        mediaType: String = "movie",
        isFavorite: Boolean = false,
        watchProgress: Double = 0.0,
        isDownloaded: Boolean = false,
        rating: Double? = null
    ): MediaItem {
        return MediaItem(
            id = id,
            title = title,
            mediaType = mediaType,
            year = 2024,
            description = "Test description",
            coverImage = "https://example.com/cover.jpg",
            rating = rating,
            quality = "1080p",
            fileSize = 1000000L,
            duration = 120,
            directoryPath = "/media/movies",
            smbPath = "smb://server/movies",
            createdAt = "2024-01-01T00:00:00Z",
            updatedAt = "2024-01-01T00:00:00Z",
            isFavorite = isFavorite,
            watchProgress = watchProgress,
            isDownloaded = isDownloaded
        )
    }

    @Test
    fun insertAndRetrieveMediaItem() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem()

        // When
        mediaDao.insertMedia(mediaItem)
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(mediaItem.title, retrieved?.title)
        assertEquals(mediaItem.mediaType, retrieved?.mediaType)
    }

    @Test
    fun insertMultipleMediaItems() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, title = "Movie 1"),
            createTestMediaItem(id = 2, title = "Movie 2"),
            createTestMediaItem(id = 3, title = "Movie 3")
        )

        // When
        mediaDao.insertAllMedia(mediaItems)
        val count = mediaDao.getCachedItemsCount()

        // Then
        assertEquals(3, count)
    }

    @Test
    fun updateFavoriteStatus() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem(isFavorite = false)
        mediaDao.insertMedia(mediaItem)

        // When
        mediaDao.updateFavoriteStatus(1, true)
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNotNull(retrieved)
        assertTrue(retrieved!!.isFavorite)
    }

    @Test
    fun updateWatchProgress() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem()
        mediaDao.insertMedia(mediaItem)

        // When
        mediaDao.updateWatchProgress(1, 0.5, "2024-01-02T00:00:00Z")
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(0.5, retrieved!!.watchProgress, 0.01)
        assertEquals("2024-01-02T00:00:00Z", retrieved.lastWatched)
    }

    @Test
    fun updateDownloadStatus() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem(isDownloaded = false)
        mediaDao.insertMedia(mediaItem)

        // When
        mediaDao.updateDownloadStatus(1, true)
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNotNull(retrieved)
        assertTrue(retrieved!!.isDownloaded)
    }

    @Test
    fun deleteMediaItem() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem()
        mediaDao.insertMedia(mediaItem)

        // When
        mediaDao.deleteMediaById(1)
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNull(retrieved)
    }

    @Test
    fun deleteAllMedia() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1),
            createTestMediaItem(id = 2),
            createTestMediaItem(id = 3)
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        mediaDao.deleteAllMedia()
        val count = mediaDao.getCachedItemsCount()

        // Then
        assertEquals(0, count)
    }

    @Test
    fun searchMedia() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, title = "The Matrix"),
            createTestMediaItem(id = 2, title = "Inception"),
            createTestMediaItem(id = 3, title = "The Matrix Reloaded")
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val results = mediaDao.searchCached("Matrix")

        // Then
        assertEquals(2, results.size)
        assertTrue(results.all { it.title.contains("Matrix") })
    }

    @Test
    fun getMediaByType() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, mediaType = "movie"),
            createTestMediaItem(id = 2, mediaType = "tv_show"),
            createTestMediaItem(id = 3, mediaType = "movie")
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val movies = mediaDao.getByType("movie")

        // Then
        assertEquals(2, movies.size)
        assertTrue(movies.all { it.mediaType == "movie" })
    }

    @Test
    fun getRecentlyAdded() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, title = "Movie 1"),
            createTestMediaItem(id = 2, title = "Movie 2"),
            createTestMediaItem(id = 3, title = "Movie 3"),
            createTestMediaItem(id = 4, title = "Movie 4"),
            createTestMediaItem(id = 5, title = "Movie 5")
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val recent = mediaDao.getRecentlyAdded(3).first()

        // Then
        assertEquals(3, recent.size)
    }

    @Test
    fun getTopRated() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, title = "Low Rated", rating = 5.0),
            createTestMediaItem(id = 2, title = "High Rated", rating = 9.0),
            createTestMediaItem(id = 3, title = "Medium Rated", rating = 7.0),
            createTestMediaItem(id = 4, title = "No Rating", rating = null)
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val topRated = mediaDao.getTopRated(2).first()

        // Then
        assertEquals(2, topRated.size)
        assertEquals("High Rated", topRated[0].title)
        assertEquals("Medium Rated", topRated[1].title)
    }

    @Test
    fun getAllMediaTypes() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, mediaType = "movie"),
            createTestMediaItem(id = 2, mediaType = "tv_show"),
            createTestMediaItem(id = 3, mediaType = "movie"),
            createTestMediaItem(id = 4, mediaType = "documentary")
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val types = mediaDao.getAllMediaTypes().first()

        // Then
        assertEquals(3, types.size)
        assertTrue(types.contains("movie"))
        assertTrue(types.contains("tv_show"))
        assertTrue(types.contains("documentary"))
    }

    @Test
    fun getTotalCount() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1),
            createTestMediaItem(id = 2),
            createTestMediaItem(id = 3)
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val count = mediaDao.getTotalCount().first()

        // Then
        assertEquals(3, count)
    }

    @Test
    fun getCountByType() = runBlocking {
        // Given
        val mediaItems = listOf(
            createTestMediaItem(id = 1, mediaType = "movie"),
            createTestMediaItem(id = 2, mediaType = "tv_show"),
            createTestMediaItem(id = 3, mediaType = "movie")
        )
        mediaDao.insertAllMedia(mediaItems)

        // When
        val movieCount = mediaDao.getCountByType("movie").first()

        // Then
        assertEquals(2, movieCount)
    }

    @Test
    fun updateRating() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem(rating = null)
        mediaDao.insertMedia(mediaItem)

        // When
        mediaDao.updateRating(1, 8.5)
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(8.5, retrieved!!.rating!!, 0.01)
    }

    @Test
    fun insertOrUpdateExistingItem() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem(title = "Original Title")
        mediaDao.insertMedia(mediaItem)

        // When
        val updatedItem = mediaItem.copy(title = "Updated Title")
        mediaDao.insertOrUpdate(updatedItem)
        val retrieved = mediaDao.getMediaById(1)

        // Then
        assertNotNull(retrieved)
        assertEquals("Updated Title", retrieved!!.title)
    }

    @Test
    fun refreshMediaReplacesAll() = runBlocking {
        // Given
        val oldItems = listOf(
            createTestMediaItem(id = 1, title = "Old 1"),
            createTestMediaItem(id = 2, title = "Old 2")
        )
        mediaDao.insertAllMedia(oldItems)

        // When
        val newItems = listOf(
            createTestMediaItem(id = 3, title = "New 1"),
            createTestMediaItem(id = 4, title = "New 2"),
            createTestMediaItem(id = 5, title = "New 3")
        )
        mediaDao.refreshMedia(newItems)

        // Then
        val count = mediaDao.getCachedItemsCount()
        assertEquals(3, count)

        val old1 = mediaDao.getMediaById(1)
        assertNull(old1)

        val new1 = mediaDao.getMediaById(3)
        assertNotNull(new1)
    }

    @Test
    fun getMediaByIdFlow() = runBlocking {
        // Given
        val mediaItem = createTestMediaItem()
        mediaDao.insertMedia(mediaItem)

        // When
        val retrieved = mediaDao.getMediaByIdFlow(1).first()

        // Then
        assertNotNull(retrieved)
        assertEquals(mediaItem.title, retrieved?.title)
    }
}
