package com.catalogizer.androidtv.data.tv

import android.content.Context
import android.util.Log
import androidx.work.*
import com.catalogizer.androidtv.DependencyContainer
import java.util.concurrent.TimeUnit

/**
 * Periodic [CoroutineWorker] that refreshes Android TV home screen channels
 * every [SYNC_INTERVAL_HOURS] hours. Requires network connectivity.
 * Skips execution if the user is not authenticated.
 */
class TvChannelSyncWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    companion object {
        private const val TAG = "TvChannelSync"
        const val WORK_NAME = "tv_channel_sync"
        const val SYNC_INTERVAL_HOURS = 6L

        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .setRequiresBatteryNotLow(true)
                .build()

            val request = PeriodicWorkRequestBuilder<TvChannelSyncWorker>(
                SYNC_INTERVAL_HOURS, TimeUnit.HOURS
            )
                .setConstraints(constraints)
                .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, 30, TimeUnit.MINUTES)
                .build()

            WorkManager.getInstance(context).enqueueUniquePeriodicWork(
                WORK_NAME,
                ExistingPeriodicWorkPolicy.KEEP,
                request
            )
            Log.d(TAG, "Periodic sync enqueued (every ${SYNC_INTERVAL_HOURS}h)")
        }

        fun cancel(context: Context) {
            WorkManager.getInstance(context).cancelUniqueWork(WORK_NAME)
            Log.d(TAG, "Periodic sync cancelled")
        }
    }

    override suspend fun doWork(): Result {
        Log.d(TAG, "Sync worker started")

        val container = DependencyContainer.getInstance(applicationContext)

        try {
            val authState = container.authRepository.authState.value
            if (!authState.isAuthenticated) {
                Log.d(TAG, "User not authenticated, skipping sync")
                return Result.success()
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to check auth state: ${e.message}")
            return Result.retry()
        }

        return try {
            val tvChannelRepo = TvChannelRepository(
                applicationContext,
                container.mediaRepository,
                container.settingsRepository
            )
            tvChannelRepo.refreshAllChannels()

            val watchNextManager = WatchNextManager(applicationContext, container.mediaRepository)
            watchNextManager.refreshWatchNext()

            Log.d(TAG, "Sync worker completed successfully")
            Result.success()
        } catch (e: Exception) {
            Log.e(TAG, "Sync worker failed: ${e.message}")
            Result.retry()
        }
    }
}
