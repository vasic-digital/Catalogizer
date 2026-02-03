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
 * Instrumentation tests for WatchProgressDao
 */
@RunWith(AndroidJUnit4::class)
class WatchProgressDaoTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var watchProgressDao: WatchProgressDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        watchProgressDao = database.watchProgressDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun createWatchProgress(
        mediaId: Long = 1,
        progress: Double = 0.5,
        lastWatched: Long = System.currentTimeMillis()
    ): WatchProgress {
        return WatchProgress(
            mediaId = mediaId,
            progress = progress,
            lastWatched = lastWatched,
            updatedAt = lastWatched
        )
    }

    @Test
    fun insertAndRetrieveProgress() = runBlocking {
        // Given
        val progress = createWatchProgress(mediaId = 1, progress = 0.5)

        // When
        watchProgressDao.insertOrUpdate(progress)
        val retrieved = watchProgressDao.getProgress(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(1L, retrieved?.mediaId)
        assertEquals(0.5, retrieved?.progress!!, 0.01)
    }

    @Test
    fun getProgressFlow() = runBlocking {
        // Given
        val progress = createWatchProgress(mediaId = 1, progress = 0.75)
        watchProgressDao.insertOrUpdate(progress)

        // When
        val retrieved = watchProgressDao.getProgressFlow(1).first()

        // Then
        assertNotNull(retrieved)
        assertEquals(0.75, retrieved?.progress!!, 0.01)
    }

    @Test
    fun getProgressFlowReturnsNullForNonExistent() = runBlocking {
        // When
        val retrieved = watchProgressDao.getProgressFlow(999).first()

        // Then
        assertNull(retrieved)
    }

    @Test
    fun getAllProgress() = runBlocking {
        // Given
        val progressItems = listOf(
            createWatchProgress(mediaId = 1, progress = 0.2),
            createWatchProgress(mediaId = 2, progress = 0.5),
            createWatchProgress(mediaId = 3, progress = 0.8)
        )
        progressItems.forEach { watchProgressDao.insertOrUpdate(it) }

        // When
        val allProgress = watchProgressDao.getAllProgress().first()

        // Then
        assertEquals(3, allProgress.size)
    }

    @Test
    fun getAllProgressOrderedByLastWatchedDesc() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val progressItems = listOf(
            createWatchProgress(mediaId = 1, lastWatched = baseTime - 2000),
            createWatchProgress(mediaId = 2, lastWatched = baseTime),
            createWatchProgress(mediaId = 3, lastWatched = baseTime - 1000)
        )
        progressItems.forEach { watchProgressDao.insertOrUpdate(it) }

        // When
        val allProgress = watchProgressDao.getAllProgress().first()

        // Then
        assertEquals(3, allProgress.size)
        assertEquals(2L, allProgress[0].mediaId) // Most recent
        assertEquals(3L, allProgress[1].mediaId)
        assertEquals(1L, allProgress[2].mediaId) // Oldest
    }

    @Test
    fun updateProgress() = runBlocking {
        // Given
        val progress = createWatchProgress(mediaId = 1, progress = 0.3)
        watchProgressDao.insertOrUpdate(progress)

        // When
        val updatedProgress = progress.copy(progress = 0.7)
        watchProgressDao.update(updatedProgress)
        val retrieved = watchProgressDao.getProgress(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(0.7, retrieved?.progress!!, 0.01)
    }

    @Test
    fun insertOrUpdateReplacesProgress() = runBlocking {
        // Given
        val progress = createWatchProgress(mediaId = 1, progress = 0.3)
        watchProgressDao.insertOrUpdate(progress)

        // When
        val updatedProgress = createWatchProgress(mediaId = 1, progress = 0.9)
        watchProgressDao.insertOrUpdate(updatedProgress)
        val retrieved = watchProgressDao.getProgress(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(0.9, retrieved?.progress!!, 0.01)
    }

    @Test
    fun deleteByMediaId() = runBlocking {
        // Given
        val progressItems = listOf(
            createWatchProgress(mediaId = 1),
            createWatchProgress(mediaId = 2)
        )
        progressItems.forEach { watchProgressDao.insertOrUpdate(it) }

        // When
        watchProgressDao.deleteByMediaId(1)

        // Then
        assertNull(watchProgressDao.getProgress(1))
        assertNotNull(watchProgressDao.getProgress(2))
    }

    @Test
    fun deleteAll() = runBlocking {
        // Given
        val progressItems = listOf(
            createWatchProgress(mediaId = 1),
            createWatchProgress(mediaId = 2),
            createWatchProgress(mediaId = 3)
        )
        progressItems.forEach { watchProgressDao.insertOrUpdate(it) }

        // When
        watchProgressDao.deleteAll()
        val allProgress = watchProgressDao.getAllProgress().first()

        // Then
        assertTrue(allProgress.isEmpty())
    }

    @Test
    fun deleteOldProgress() = runBlocking {
        // Given
        val baseTime = System.currentTimeMillis()
        val progressItems = listOf(
            createWatchProgress(mediaId = 1, lastWatched = baseTime - 100000),
            createWatchProgress(mediaId = 2, lastWatched = baseTime),
            createWatchProgress(mediaId = 3, lastWatched = baseTime - 200000)
        )
        progressItems.forEach { watchProgressDao.insertOrUpdate(it) }

        // When
        watchProgressDao.deleteOldProgress(baseTime - 50000)
        val remaining = watchProgressDao.getAllProgress().first()

        // Then
        assertEquals(1, remaining.size)
        assertEquals(2L, remaining[0].mediaId)
    }

    @Test
    fun progressRangeValidation() = runBlocking {
        // Test progress at boundaries
        val progressAtZero = createWatchProgress(mediaId = 1, progress = 0.0)
        val progressAtOne = createWatchProgress(mediaId = 2, progress = 1.0)
        val progressAtMid = createWatchProgress(mediaId = 3, progress = 0.5)

        watchProgressDao.insertOrUpdate(progressAtZero)
        watchProgressDao.insertOrUpdate(progressAtOne)
        watchProgressDao.insertOrUpdate(progressAtMid)

        // Then
        val retrievedZero = watchProgressDao.getProgress(1)
        val retrievedOne = watchProgressDao.getProgress(2)
        val retrievedMid = watchProgressDao.getProgress(3)

        assertEquals(0.0, retrievedZero?.progress!!, 0.001)
        assertEquals(1.0, retrievedOne?.progress!!, 0.001)
        assertEquals(0.5, retrievedMid?.progress!!, 0.001)
    }

    @Test
    fun continueWatchingScenario() = runBlocking {
        // Given - simulate continue watching scenario
        val baseTime = System.currentTimeMillis()

        // User watches three items to different progress points
        val progressItems = listOf(
            createWatchProgress(mediaId = 1, progress = 0.25, lastWatched = baseTime - 3000),
            createWatchProgress(mediaId = 2, progress = 0.75, lastWatched = baseTime - 1000),
            createWatchProgress(mediaId = 3, progress = 0.50, lastWatched = baseTime - 2000)
        )
        progressItems.forEach { watchProgressDao.insertOrUpdate(it) }

        // When - retrieve all progress ordered by last watched
        val allProgress = watchProgressDao.getAllProgress().first()

        // Then - should be ordered by most recently watched
        assertEquals(3, allProgress.size)
        assertEquals(2L, allProgress[0].mediaId) // Most recently watched
        assertEquals(0.75, allProgress[0].progress, 0.01)
    }

    @Test
    fun markAsCompleted() = runBlocking {
        // Given - progress at 90%
        val progress = createWatchProgress(mediaId = 1, progress = 0.9)
        watchProgressDao.insertOrUpdate(progress)

        // When - mark as completed (100%)
        val completed = progress.copy(progress = 1.0)
        watchProgressDao.update(completed)
        val retrieved = watchProgressDao.getProgress(1)

        // Then
        assertNotNull(retrieved)
        assertEquals(1.0, retrieved?.progress!!, 0.01)
    }
}
