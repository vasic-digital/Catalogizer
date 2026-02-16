package com.catalogizer.android

import android.content.Context
import androidx.work.WorkerParameters
import com.catalogizer.android.data.sync.SyncManager
import com.catalogizer.android.data.sync.SyncWorker
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class CatalogizerWorkerFactoryTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockDependencyContainer = mockk<DependencyContainer>(relaxed = true)
    private val mockSyncManager = mockk<SyncManager>(relaxed = true)
    private val mockContext = mockk<Context>(relaxed = true)
    private val mockWorkerParams = mockk<WorkerParameters>(relaxed = true)
    private lateinit var workerFactory: CatalogizerWorkerFactory

    @Before
    fun setup() {
        every { mockDependencyContainer.syncManager } returns mockSyncManager
        workerFactory = CatalogizerWorkerFactory(mockDependencyContainer)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `createWorker should return SyncWorker for SyncWorker class name`() {
        val worker = workerFactory.createWorker(
            mockContext,
            SyncWorker::class.java.name,
            mockWorkerParams
        )

        assertNotNull(worker)
        assertTrue(worker is SyncWorker)
    }

    @Test
    fun `createWorker should return null for unknown worker class name`() {
        val worker = workerFactory.createWorker(
            mockContext,
            "com.catalogizer.android.UnknownWorker",
            mockWorkerParams
        )

        assertNull(worker)
    }

    @Test
    fun `createWorker should return null for empty class name`() {
        val worker = workerFactory.createWorker(
            mockContext,
            "",
            mockWorkerParams
        )

        assertNull(worker)
    }

    @Test
    fun `createWorker should access syncManager from dependency container`() {
        workerFactory.createWorker(
            mockContext,
            SyncWorker::class.java.name,
            mockWorkerParams
        )

        verify { mockDependencyContainer.syncManager }
    }

    @Test
    fun `createWorker should return different instances for each call`() {
        val worker1 = workerFactory.createWorker(
            mockContext,
            SyncWorker::class.java.name,
            mockWorkerParams
        )
        val worker2 = workerFactory.createWorker(
            mockContext,
            SyncWorker::class.java.name,
            mockWorkerParams
        )

        assertNotNull(worker1)
        assertNotNull(worker2)
        assertNotSame(worker1, worker2)
    }
}
