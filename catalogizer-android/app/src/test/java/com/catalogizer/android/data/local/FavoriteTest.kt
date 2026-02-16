package com.catalogizer.android.data.local

import org.junit.Assert.*
import org.junit.Test

class FavoriteTest {

    @Test
    fun `Favorite has correct defaults`() {
        val favorite = Favorite(mediaId = 42L)

        assertEquals(42L, favorite.mediaId)
        assertTrue(favorite.createdAt > 0)
        assertTrue(favorite.updatedAt > 0)
    }

    @Test
    fun `Favorite with custom timestamps`() {
        val favorite = Favorite(
            mediaId = 1L,
            createdAt = 1000L,
            updatedAt = 2000L
        )

        assertEquals(1L, favorite.mediaId)
        assertEquals(1000L, favorite.createdAt)
        assertEquals(2000L, favorite.updatedAt)
    }

    @Test
    fun `Favorite equality works correctly`() {
        val ts = System.currentTimeMillis()
        val f1 = Favorite(mediaId = 1L, createdAt = ts, updatedAt = ts)
        val f2 = Favorite(mediaId = 1L, createdAt = ts, updatedAt = ts)
        val f3 = Favorite(mediaId = 2L, createdAt = ts, updatedAt = ts)

        assertEquals(f1, f2)
        assertNotEquals(f1, f3)
    }

    @Test
    fun `Favorite copy updates correctly`() {
        val original = Favorite(mediaId = 1L, createdAt = 1000L, updatedAt = 1000L)
        val updated = original.copy(updatedAt = 2000L)

        assertEquals(1000L, original.updatedAt)
        assertEquals(2000L, updated.updatedAt)
        assertEquals(original.mediaId, updated.mediaId)
    }

    @Test
    fun `Favorite uses mediaId as primary key`() {
        val favorite = Favorite(mediaId = 99L)
        assertEquals(99L, favorite.mediaId)
    }
}
