package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import com.catalogizer.android.MainDispatcherRule
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
class FavoriteDaoTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var database: CatalogizerDatabase
    private lateinit var favoriteDao: FavoriteDao

    private fun createFavorite(
        mediaId: Long = 1L,
        createdAt: Long = System.currentTimeMillis(),
        updatedAt: Long = System.currentTimeMillis()
    ): Favorite {
        return Favorite(
            mediaId = mediaId,
            createdAt = createdAt,
            updatedAt = updatedAt
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
        favoriteDao = database.favoriteDao()
    }

    @After
    fun tearDown() {
        database.close()
        clearAllMocks()
    }

    // --- Insert Tests ---

    @Test
    fun `insertOrUpdate should add new favorite`() = runTest {
        val favorite = createFavorite(mediaId = 1L)

        favoriteDao.insertOrUpdate(favorite)

        val result = favoriteDao.getFavorite(1L)
        assertNotNull(result)
        assertEquals(1L, result?.mediaId)
    }

    @Test
    fun `insertOrUpdate should update existing favorite`() = runTest {
        val original = createFavorite(mediaId = 1L, updatedAt = 1000L)
        favoriteDao.insertOrUpdate(original)

        val updated = createFavorite(mediaId = 1L, updatedAt = 2000L)
        favoriteDao.insertOrUpdate(updated)

        val result = favoriteDao.getFavorite(1L)
        assertNotNull(result)
        assertEquals(2000L, result?.updatedAt)
    }

    @Test
    fun `insertOrUpdate should handle multiple favorites`() = runTest {
        val favorites = listOf(
            createFavorite(mediaId = 1L),
            createFavorite(mediaId = 2L),
            createFavorite(mediaId = 3L)
        )

        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        val count = favoriteDao.getFavoritesCount()
        assertEquals(3, count)
    }

    // --- Query Tests ---

    @Test
    fun `getFavorite should return null for non-existent mediaId`() = runTest {
        val result = favoriteDao.getFavorite(999L)

        assertNull(result)
    }

    @Test
    fun `getFavorite should return correct favorite`() = runTest {
        val favorite = createFavorite(mediaId = 42L)
        favoriteDao.insertOrUpdate(favorite)

        val result = favoriteDao.getFavorite(42L)

        assertNotNull(result)
        assertEquals(42L, result?.mediaId)
    }

    @Test
    fun `getFavoriteFlow should emit favorite changes`() = runTest {
        val favorite = createFavorite(mediaId = 1L)
        favoriteDao.insertOrUpdate(favorite)

        val result = favoriteDao.getFavoriteFlow(1L).first()

        assertNotNull(result)
        assertEquals(1L, result?.mediaId)
    }

    @Test
    fun `getFavoriteFlow should emit null for missing favorite`() = runTest {
        val result = favoriteDao.getFavoriteFlow(999L).first()

        assertNull(result)
    }

    @Test
    fun `getAllFavorites should return all favorites ordered by updatedAt descending`() = runTest {
        val favorites = listOf(
            createFavorite(mediaId = 1L, updatedAt = 1000L),
            createFavorite(mediaId = 2L, updatedAt = 3000L),
            createFavorite(mediaId = 3L, updatedAt = 2000L)
        )
        favorites.forEach { favoriteDao.insertOrUpdate(it) }

        val result = favoriteDao.getAllFavorites().first()

        assertEquals(3, result.size)
        assertEquals(2L, result[0].mediaId) // updatedAt = 3000L
        assertEquals(3L, result[1].mediaId) // updatedAt = 2000L
        assertEquals(1L, result[2].mediaId) // updatedAt = 1000L
    }

    @Test
    fun `getAllFavorites should return empty list when no favorites`() = runTest {
        val result = favoriteDao.getAllFavorites().first()

        assertTrue(result.isEmpty())
    }

    @Test
    fun `getFavoritesCount should return correct count`() = runTest {
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 1L))
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 2L))
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 3L))

        val count = favoriteDao.getFavoritesCount()

        assertEquals(3, count)
    }

    @Test
    fun `getFavoritesCount should return zero when empty`() = runTest {
        val count = favoriteDao.getFavoritesCount()

        assertEquals(0, count)
    }

    @Test
    fun `getFavoritesCountFlow should emit count changes`() = runTest {
        val initialCount = favoriteDao.getFavoritesCountFlow().first()
        assertEquals(0, initialCount)

        favoriteDao.insertOrUpdate(createFavorite(mediaId = 1L))

        val updatedCount = favoriteDao.getFavoritesCountFlow().first()
        assertEquals(1, updatedCount)
    }

    // --- Delete Tests ---

    @Test
    fun `delete should remove favorite from database`() = runTest {
        val favorite = createFavorite(mediaId = 1L)
        favoriteDao.insertOrUpdate(favorite)

        favoriteDao.delete(favorite)

        val result = favoriteDao.getFavorite(1L)
        assertNull(result)
    }

    @Test
    fun `deleteByMediaId should remove favorite by media id`() = runTest {
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 1L))
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 2L))

        favoriteDao.deleteByMediaId(1L)

        assertNull(favoriteDao.getFavorite(1L))
        assertNotNull(favoriteDao.getFavorite(2L))
    }

    @Test
    fun `deleteByMediaId should handle non-existent id gracefully`() = runTest {
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 1L))

        favoriteDao.deleteByMediaId(999L) // Non-existent ID

        // Should not affect existing favorites
        assertEquals(1, favoriteDao.getFavoritesCount())
    }

    @Test
    fun `deleteAll should remove all favorites`() = runTest {
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 1L))
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 2L))
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 3L))

        favoriteDao.deleteAll()

        val count = favoriteDao.getFavoritesCount()
        assertEquals(0, count)
    }

    @Test
    fun `deleteAll on empty database should not throw`() = runTest {
        favoriteDao.deleteAll()

        val count = favoriteDao.getFavoritesCount()
        assertEquals(0, count)
    }

    // --- Favorite Lifecycle Tests ---

    @Test
    fun `add and then remove favorite should work correctly`() = runTest {
        val favorite = createFavorite(mediaId = 1L)

        // Add
        favoriteDao.insertOrUpdate(favorite)
        assertNotNull(favoriteDao.getFavorite(1L))

        // Remove
        favoriteDao.deleteByMediaId(1L)
        assertNull(favoriteDao.getFavorite(1L))
    }

    @Test
    fun `re-adding a deleted favorite should work`() = runTest {
        val favorite = createFavorite(mediaId = 1L, updatedAt = 1000L)

        favoriteDao.insertOrUpdate(favorite)
        favoriteDao.deleteByMediaId(1L)

        val newFavorite = createFavorite(mediaId = 1L, updatedAt = 2000L)
        favoriteDao.insertOrUpdate(newFavorite)

        val result = favoriteDao.getFavorite(1L)
        assertNotNull(result)
        assertEquals(2000L, result?.updatedAt)
    }

    @Test
    fun `favorite should preserve createdAt and updatedAt`() = runTest {
        val favorite = Favorite(
            mediaId = 1L,
            createdAt = 1000000L,
            updatedAt = 2000000L
        )
        favoriteDao.insertOrUpdate(favorite)

        val result = favoriteDao.getFavorite(1L)

        assertNotNull(result)
        assertEquals(1000000L, result?.createdAt)
        assertEquals(2000000L, result?.updatedAt)
    }

    @Test
    fun `multiple operations should maintain data integrity`() = runTest {
        // Insert several favorites
        (1L..5L).forEach {
            favoriteDao.insertOrUpdate(createFavorite(mediaId = it))
        }
        assertEquals(5, favoriteDao.getFavoritesCount())

        // Delete some
        favoriteDao.deleteByMediaId(2L)
        favoriteDao.deleteByMediaId(4L)
        assertEquals(3, favoriteDao.getFavoritesCount())

        // Re-add one
        favoriteDao.insertOrUpdate(createFavorite(mediaId = 2L))
        assertEquals(4, favoriteDao.getFavoritesCount())

        // Verify specific items
        assertNotNull(favoriteDao.getFavorite(1L))
        assertNotNull(favoriteDao.getFavorite(2L))
        assertNotNull(favoriteDao.getFavorite(3L))
        assertNull(favoriteDao.getFavorite(4L))
        assertNotNull(favoriteDao.getFavorite(5L))
    }
}
