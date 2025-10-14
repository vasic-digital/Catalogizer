package com.catalogizer.android

import android.content.Context
import androidx.work.ListenableWorker
import androidx.work.WorkerFactory
import androidx.work.WorkerParameters
import com.catalogizer.android.data.sync.SyncWorker

class CatalogizerWorkerFactory(private val dependencyContainer: DependencyContainer) : WorkerFactory() {

    override fun createWorker(
        appContext: Context,
        workerClassName: String,
        workerParameters: WorkerParameters
    ): ListenableWorker? {
        return when (workerClassName) {
            SyncWorker::class.java.name -> {
                SyncWorker(
                    appContext,
                    workerParameters,
                    dependencyContainer.syncManager
                )
            }
            else -> null
        }
    }
}