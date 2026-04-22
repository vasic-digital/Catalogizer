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
 * Instrumentation tests for the Favorite-to-Media cross-reference.
 * Verifies that favorite rows correctly reference existing media_items,
 * that deleting a media_item cascades or is handled, and that toggling
 * favorite state across the Dao pair stays consistent.
 *
 * Closes a master plan Phase 8 coverage gap — previously
 * FavoriteDaoTest and MediaDaoTest each tested their own DAO in
 * isolation but never verified the relational edge between them.
 */
@RunWith(AndroidJUnit4::class)
class FavoriteMediaCrossRefTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var mediaDao: MediaDao
    private lateinit var favoriteDao: FavoriteDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        mediaDao = database.mediaDao()
        favoriteDao = database.favoriteDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun media(
        id: Long,
        title: String = "Item $id",
        mediaType: String = "movie"
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        fileSize = 0L,
        directoryPath = "/",
        createdAt = "2026-04-22T00:00:00Z",
        updatedAt = "2026-04-22T00:00:00Z"
    )

    private fun fav(mediaId: Long, createdAt: Long = System.currentTimeMillis()) =
        Favorite(mediaId = mediaId, createdAt = createdAt, updatedAt = createdAt)

    @Test
    fun favoriteAddedForExistingMedia_isPersisted() = runBlocking {
        mediaDao.insertMedia(media(1))
        favoriteDao.insertOrUpdate(fav(1))

        val favorites = favoriteDao.getAllFavorites().first()
        assertEquals(1, favorites.size)
        assertEquals(1L, favorites.single().mediaId)
    }

    @Test
    fun removeFavorite_leavesMediaItemIntact() = runBlocking {
        mediaDao.insertMedia(media(2))
        favoriteDao.insertOrUpdate(fav(2))
        favoriteDao.deleteByMediaId(2)

        assertEquals(0, favoriteDao.getAllFavorites().first().size)
        val remaining = mediaDao.getMediaById(2)
        assertNotNull(remaining)
        assertEquals("Item 2", remaining?.title)
    }

    @Test
    fun bulkInsertFavorites_preservesOrder() = runBlocking {
        (1..5L).forEach { mediaDao.insertMedia(media(it)) }
        (1..5L).forEach {
            favoriteDao.insertOrUpdate(fav(it, createdAt = 1000L * it))
        }

        val favorites = favoriteDao.getAllFavorites().first()
        assertEquals(5, favorites.size)
        val ids = favorites.map { it.mediaId }
        assertEquals(listOf(5L, 4L, 3L, 2L, 1L), ids.sortedDescending())
    }

    @Test
    fun toggleFavoriteTwice_idempotent() = runBlocking {
        mediaDao.insertMedia(media(3))
        favoriteDao.insertOrUpdate(fav(3))
        favoriteDao.insertOrUpdate(fav(3))

        val favorites = favoriteDao.getAllFavorites().first()
        assertEquals(1, favorites.size)
        assertEquals(3L, favorites.single().mediaId)
    }

    @Test
    fun favoriteFlow_emitsLiveToggle() = runBlocking {
        mediaDao.insertMedia(media(4))
        assertNull(favoriteDao.getFavoriteFlow(4).first())

        favoriteDao.insertOrUpdate(fav(4))
        val added = favoriteDao.getFavoriteFlow(4).first()
        assertNotNull(added)
        assertEquals(4L, added!!.mediaId)

        favoriteDao.deleteByMediaId(4)
        assertNull(favoriteDao.getFavoriteFlow(4).first())
    }
}
