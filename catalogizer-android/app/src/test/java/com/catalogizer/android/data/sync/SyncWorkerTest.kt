package com.catalogizer.android.data.sync

import android.content.Context
import androidx.work.ListenableWorker
import androidx.work.WorkerParameters
import com.catalogizer.android.MainDispatcherRule
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SyncWorkerTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockContext = mockk<Context>(relaxed = true)
    private val mockWorkerParams = mockk<WorkerParameters>(relaxed = true)
    private val mockSyncManager = mockk<SyncManager>(relaxed = true)
    private lateinit var syncWorker: SyncWorker

    @Before
    fun setup() {
        syncWorker = SyncWorker(mockContext, mockWorkerParams, mockSyncManager)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `doWork should return success when sync succeeds`() = runTest {
        val successResult = SyncResult(
            success = true,
            timestamp = System.currentTimeMillis(),
            syncedItems = 10,
            failedItems = 0
        )
        coEvery { mockSyncManager.performManualSync() } returns successResult

        val result = syncWorker.doWork()

        assertEquals(ListenableWorker.Result.success(), result)
        coVerify { mockSyncManager.performManualSync() }
    }

    @Test
    fun `doWork should return retry when sync fails`() = runTest {
        val failResult = SyncResult(
            success = false,
            timestamp = System.currentTimeMillis(),
            syncedItems = 0,
            failedItems = 5,
            errorMessage = "Network unavailable"
        )
        coEvery { mockSyncManager.performManualSync() } returns failResult

        val result = syncWorker.doWork()

        assertEquals(ListenableWorker.Result.retry(), result)
    }

    @Test
    fun `doWork should return failure when exception occurs`() = runTest {
        coEvery { mockSyncManager.performManualSync() } throws RuntimeException("Unexpected error")

        val result = syncWorker.doWork()

        assertEquals(ListenableWorker.Result.failure(), result)
    }

    @Test
    fun `doWork should call performManualSync exactly once`() = runTest {
        val successResult = SyncResult(
            success = true,
            timestamp = System.currentTimeMillis()
        )
        coEvery { mockSyncManager.performManualSync() } returns successResult

        syncWorker.doWork()

        coVerify(exactly = 1) { mockSyncManager.performManualSync() }
    }

    @Test
    fun `doWork should retry when sync partially fails`() = runTest {
        val partialResult = SyncResult(
            success = false,
            timestamp = System.currentTimeMillis(),
            syncedItems = 5,
            failedItems = 3,
            errorMessage = "Partial sync failure"
        )
        coEvery { mockSyncManager.performManualSync() } returns partialResult

        val result = syncWorker.doWork()

        assertEquals(ListenableWorker.Result.retry(), result)
    }

    @Test
    fun `doWork should return success even with zero synced items on success`() = runTest {
        val emptySuccess = SyncResult(
            success = true,
            timestamp = System.currentTimeMillis(),
            syncedItems = 0,
            failedItems = 0
        )
        coEvery { mockSyncManager.performManualSync() } returns emptySuccess

        val result = syncWorker.doWork()

        assertEquals(ListenableWorker.Result.success(), result)
    }
}
