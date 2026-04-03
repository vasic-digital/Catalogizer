package com.catalogizer.android.data.sync

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters

/**
 * [CoroutineWorker] that delegates background data synchronization to [SyncManager].
 * Scheduled periodically by [WorkManager] and retries on sync failure.
 */
class SyncWorker(
    context: Context,
    workerParams: WorkerParameters,
    private val syncManager: SyncManager
) : CoroutineWorker(context, workerParams) {

    override suspend fun doWork(): Result {
        return try {
            val syncResult = syncManager.performManualSync()

            if (syncResult.success) {
                Result.success()
            } else {
                Result.retry()
            }
        } catch (e: Exception) {
            Result.failure()
        }
    }
}