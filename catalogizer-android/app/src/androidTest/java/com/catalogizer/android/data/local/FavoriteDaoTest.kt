package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * Instrumentation tests for FavoriteDao
 */
@RunWith(AndroidJUnit4::class)
class FavoriteDaoTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var favoriteDao: FavoriteDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        favoriteDao = database.favoriteDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun createFavorite(
        mediaId: Long = 1,
        createdAt: Long = System.currentTimeMillis()
    ): Favorite {
        return Favorite(
            mediaId = mediaId,
            createdAt = createdAt,
            updatedAt = createdAt
        )
    }

    @Test
    fun insertAndRetrieveFavorite() = runBlocking {
        // Given
        val favorite = createFavorite(mediaId = 1)

        // When
        favoriteDao.insertOrUpdate(favorite)
        val retrieved = favoriteDao.getFavorite(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(1L, retrieved?.mediaId)
    }

    @Test
    fun getAllFavorites() = runBlocking {
        // Given
        val favorites = listOf(
            createFavorite(mediaId = 1),
            createFavorite(mediaId = 2),
            createFavorite(mediaId = 3)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        // When
        val allFavorites = favoriteDao.getAllFavorites().first()

        // Then
        assertEquals(3, allFavorites.size)
    }

    @Test
    fun getAllFavoritesOrderedByUpdatedAtDesc() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val favorites = listOf(
            Favorite(mediaId = 1, createdAt = baseTime - 2000, updatedAt = baseTime - 2000),
            Favorite(mediaId = 2, createdAt = baseTime, updatedAt = baseTime),
            Favorite(mediaId = 3, createdAt = baseTime - 1000, updatedAt = baseTime - 1000)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        // When
        val allFavorites = favoriteDao.getAllFavorites().first()

        // Then
        assertEquals(3, allFavorites.size)
        assertEquals(2L, allFavorites[0].mediaId) // Most recent
        assertEquals(3L, allFavorites[1].mediaId)
        assertEquals(1L, allFavorites[2].mediaId) // Oldest
    }

    @Test
    fun getFavoriteFlow() = runBlocking {
        // Given
        val favorite = createFavorite(mediaId = 1)
        favoriteDao.insertOrUpdate(favorite)

        // When
        val retrieved = favoriteDao.getFavoriteFlow(1).first()

        // Then
        assertNotNull(retrieved)
        assertEquals(1L, retrieved?.mediaId)
    }

    @Test
    fun getFavoriteFlowReturnsNullForNonExistent() = runBlocking {
        // When
        val retrieved = favoriteDao.getFavoriteFlow(999).first()

        // Then
        assertNull(retrieved)
    }

    @Test
    fun deleteFavorite() = runBlocking {
        // Given
        val favorite = createFavorite(mediaId = 1)
        favoriteDao.insertOrUpdate(favorite)

        // When
        favoriteDao.delete(favorite)
        val retrieved = favoriteDao.getFavorite(1)

        // Then
        assertNull(retrieved)
    }

    @Test
    fun deleteByMediaId() = runBlocking {
        // Given
        val favorites = listOf(
            createFavorite(mediaId = 1),
            createFavorite(mediaId = 2)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        // When
        favoriteDao.deleteByMediaId(1)

        // Then
        assertNull(favoriteDao.getFavorite(1))
        assertNotNull(favoriteDao.getFavorite(2))
    }

    @Test
    fun deleteAll() = runBlocking {
        // Given
        val favorites = listOf(
            createFavorite(mediaId = 1),
            createFavorite(mediaId = 2),
            createFavorite(mediaId = 3)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        // When
        favoriteDao.deleteAll()
        val count = favoriteDao.getFavoritesCount()

        // Then
        assertEquals(0, count)
    }

    @Test
    fun getFavoritesCount() = runBlocking {
        // Given
        val favorites = listOf(
            createFavorite(mediaId = 1),
            createFavorite(mediaId = 2),
            createFavorite(mediaId = 3)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        // When
        val count = favoriteDao.getFavoritesCount()

        // Then
        assertEquals(3, count)
    }

    @Test
    fun getFavoritesCountFlow() = runBlocking {
        // Given
        val favorites = listOf(
            createFavorite(mediaId = 1),
            createFavorite(mediaId = 2)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        // When
        val count = favoriteDao.getFavoritesCountFlow().first()

        // Then
        assertEquals(2, count)
    }

    @Test
    fun insertOrUpdateReplacesFavorite() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val favorite = Favorite(mediaId = 1, createdAt = baseTime, updatedAt = baseTime)
        favoriteDao.insertOrUpdate(favorite)

        // When
        val updatedFavorite = Favorite(mediaId = 1, createdAt = baseTime, updatedAt = baseTime + 1000)
        favoriteDao.insertOrUpdate(updatedFavorite)
        val retrieved = favoriteDao.getFavorite(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(baseTime + 1000, retrieved!!.updatedAt)
    }

    @Test
    fun duplicateFavoriteNotInserted() = runBlocking {
        // Given
        val favorite = createFavorite(mediaId = 1)
        favoriteDao.insertOrUpdate(favorite)
        favoriteDao.insertOrUpdate(favorite) // Insert again

        // When
        val count = favoriteDao.getFavoritesCount()

        // Then
        assertEquals(1, count) // Should still be 1 due to REPLACE strategy
    }
}
