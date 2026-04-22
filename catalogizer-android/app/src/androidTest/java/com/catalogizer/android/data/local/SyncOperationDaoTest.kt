package com.catalogizer.android.data.local

import android.content.Context
import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.catalogizer.android.data.sync.SyncOperation
import com.catalogizer.android.data.sync.SyncOperationType
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith

/**
 * Instrumentation tests for SyncOperationDao — the offline operation
 * queue DAO that powers sync-on-reconnect. Closes the Master Plan
 * Phase 8 instrumented-test gap: prior to this file the DAO had no
 * dedicated coverage, so retry-count / max-retry / failed-operation-
 * cleanup behaviour was only exercised via higher-level integration
 * tests.
 */
@RunWith(AndroidJUnit4::class)
class SyncOperationDaoTest {

    private lateinit var database: CatalogizerDatabase
    private lateinit var dao: SyncOperationDao

    @Before
    fun setup() {
        val context = ApplicationProvider.getApplicationContext<Context>()
        database = Room.inMemoryDatabaseBuilder(
            context,
            CatalogizerDatabase::class.java
        ).allowMainThreadQueries().build()
        dao = database.syncOperationDao()
    }

    @After
    fun teardown() {
        database.close()
    }

    private fun op(
        id: Long = 0,
        type: SyncOperationType = SyncOperationType.UPDATE_PROGRESS,
        mediaId: Long = 100,
        data: String? = """{"position":42}""",
        timestamp: Long = System.currentTimeMillis(),
        retryCount: Int = 0,
        maxRetries: Int = 3
    ) = SyncOperation(id, type, mediaId, data, timestamp, retryCount, maxRetries)

    @Test
    fun insert_andGetAll_returnsOperationsInTimestampOrder() = runBlocking {
        val earlier = op(mediaId = 1, timestamp = 1000L)
        val later = op(mediaId = 2, timestamp = 2000L)
        dao.insertOperation(later)
        dao.insertOperation(earlier)

        val all = dao.getAllOperations()

        assertEquals(2, all.size)
        assertEquals(1L, all[0].mediaId)
        assertEquals(2L, all[1].mediaId)
    }

    @Test
    fun getPendingOperations_excludesFailedAtOrAboveMaxRetries() = runBlocking {
        dao.insertOperation(op(mediaId = 10, retryCount = 0, maxRetries = 3))
        dao.insertOperation(op(mediaId = 11, retryCount = 2, maxRetries = 3))
        dao.insertOperation(op(mediaId = 12, retryCount = 3, maxRetries = 3))
        dao.insertOperation(op(mediaId = 13, retryCount = 5, maxRetries = 3))

        val pending = dao.getPendingOperations()

        assertEquals(2, pending.size)
        val pendingMediaIds = pending.map { it.mediaId }.toSet()
        assertTrue(pendingMediaIds.containsAll(setOf(10L, 11L)))
        assertFalse(pendingMediaIds.contains(12L))
        assertFalse(pendingMediaIds.contains(13L))
    }

    @Test
    fun getPendingOperationsCountFlow_emitsCurrentCount() = runBlocking {
        assertEquals(0, dao.getPendingOperationsCountFlow().first())

        dao.insertOperation(op(mediaId = 20, retryCount = 0))
        dao.insertOperation(op(mediaId = 21, retryCount = 0))
        dao.insertOperation(op(mediaId = 22, retryCount = 3, maxRetries = 3))

        assertEquals(2, dao.getPendingOperationsCountFlow().first())
    }

    @Test
    fun updateRetryCount_incrementsStoredValue() = runBlocking {
        val inserted = op(mediaId = 30, retryCount = 0)
        val id = dao.insertOperation(inserted)

        dao.updateRetryCount(id, 2)

        val fetched = dao.getAllOperations().single()
        assertEquals(2, fetched.retryCount)
    }

    @Test
    fun resetRetryCount_zeroesEveryRow() = runBlocking {
        dao.insertOperation(op(mediaId = 40, retryCount = 1))
        dao.insertOperation(op(mediaId = 41, retryCount = 3, maxRetries = 5))

        dao.resetRetryCount()

        dao.getAllOperations().forEach { assertEquals(0, it.retryCount) }
    }

    @Test
    fun deleteFailedOperations_removesAtOrAboveGivenThreshold() = runBlocking {
        dao.insertOperation(op(mediaId = 50, retryCount = 0))
        dao.insertOperation(op(mediaId = 51, retryCount = 3, maxRetries = 3))
        dao.insertOperation(op(mediaId = 52, retryCount = 5, maxRetries = 3))

        dao.deleteFailedOperations(3)

        val remaining = dao.getAllOperations()
        assertEquals(1, remaining.size)
        assertEquals(50L, remaining.single().mediaId)
    }

    @Test
    fun deleteOperationsByMediaAndType_removesExactMatches() = runBlocking {
        dao.insertOperation(op(mediaId = 60, type = SyncOperationType.UPDATE_PROGRESS))
        dao.insertOperation(op(mediaId = 60, type = SyncOperationType.TOGGLE_FAVORITE))
        dao.insertOperation(op(mediaId = 61, type = SyncOperationType.UPDATE_PROGRESS))

        dao.deleteOperationsByMediaAndType(60, SyncOperationType.UPDATE_PROGRESS)

        val remaining = dao.getAllOperations().map { it.mediaId to it.type }.toSet()
        assertEquals(
            setOf(
                60L to SyncOperationType.TOGGLE_FAVORITE,
                61L to SyncOperationType.UPDATE_PROGRESS
            ),
            remaining
        )
    }

    @Test
    fun getOperationByMediaAndType_returnsNullWhenAbsent() = runBlocking {
        dao.insertOperation(op(mediaId = 70, type = SyncOperationType.UPDATE_PROGRESS))

        val result = dao.getOperationByMediaAndType(70, SyncOperationType.DELETE_MEDIA)

        assertNull(result)
    }
}
