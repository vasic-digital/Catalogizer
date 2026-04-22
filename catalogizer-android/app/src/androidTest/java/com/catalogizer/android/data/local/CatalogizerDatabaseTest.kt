package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * Instrumentation tests for CatalogizerDatabase — verifies the Room
 * database builds cleanly with every DAO accessor returning a
 * non-null instance. Regression guard against Room-generated
 * implementation absences that only surface at runtime (e.g. after
 * adding a new @Entity without rebuilding kapt).
 */
@RunWith(AndroidJUnit4::class)
class CatalogizerDatabaseTest {

    private lateinit var database: CatalogizerDatabase

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
    }

    @After
    fun teardown() {
        database.close()
    }

    @Test
    fun mediaDao_isAvailable() {
        assertNotNull(database.mediaDao())
    }

    @Test
    fun searchHistoryDao_isAvailable() {
        assertNotNull(database.searchHistoryDao())
    }

    @Test
    fun downloadDao_isAvailable() {
        assertNotNull(database.downloadDao())
    }

    @Test
    fun syncOperationDao_isAvailable() {
        assertNotNull(database.syncOperationDao())
    }

    @Test
    fun watchProgressDao_isAvailable() {
        assertNotNull(database.watchProgressDao())
    }

    @Test
    fun favoriteDao_isAvailable() {
        assertNotNull(database.favoriteDao())
    }

    @Test
    fun database_isOpen_afterConstruction() {
        assertTrue(database.isOpen)
    }

    @Test
    fun database_closeThenReopenInstance_buildsIndependently() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        val second = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        try {
            assertNotSame(database, second)
            assertTrue(second.isOpen)
        } finally {
            second.close()
        }
    }
}
