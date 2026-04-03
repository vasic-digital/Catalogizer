package com.catalogizer.android.data.local

import androidx.room.*
import com.catalogizer.android.data.sync.SyncOperation
import com.catalogizer.android.data.sync.SyncOperationType
import kotlinx.coroutines.flow.Flow

/**
 * Data access object for [SyncOperation] entities managing the queue of
 * pending offline operations awaiting server synchronization.
 */
@Dao
interface SyncOperationDao {

    @Query("SELECT * FROM sync_operations ORDER BY timestamp ASC")
    suspend fun getAllOperations(): List<SyncOperation>

    @Query("SELECT * FROM sync_operations WHERE retryCount < maxRetries ORDER BY timestamp ASC")
    suspend fun getPendingOperations(): List<SyncOperation>

    @Query("SELECT COUNT(*) FROM sync_operations WHERE retryCount < maxRetries")
    suspend fun getPendingOperationsCount(): Int

    @Query("SELECT COUNT(*) FROM sync_operations WHERE retryCount < maxRetries")
    fun getPendingOperationsCountFlow(): Flow<Int>

    @Query("SELECT * FROM sync_operations WHERE mediaId = :mediaId AND type = :type LIMIT 1")
    suspend fun getOperationByMediaAndType(mediaId: Long, type: SyncOperationType): SyncOperation?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertOperation(operation: SyncOperation): Long

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertOperations(operations: List<SyncOperation>): List<Long>

    @Update
    suspend fun updateOperation(operation: SyncOperation)

    @Query("UPDATE sync_operations SET retryCount = :retryCount WHERE id = :operationId")
    suspend fun updateRetryCount(operationId: Long, retryCount: Int)

    @Query("UPDATE sync_operations SET retryCount = 0")
    suspend fun resetRetryCount()

    @Delete
    suspend fun deleteOperation(operation: SyncOperation)

    @Query("DELETE FROM sync_operations WHERE id = :operationId")
    suspend fun deleteOperation(operationId: Long)

    @Query("DELETE FROM sync_operations WHERE retryCount >= :maxRetries")
    suspend fun deleteFailedOperations(maxRetries: Int)

    @Query("DELETE FROM sync_operations WHERE mediaId = :mediaId AND type = :type")
    suspend fun deleteOperationsByMediaAndType(mediaId: Long, type: SyncOperationType)

    @Query("DELETE FROM sync_operations")
    suspend fun deleteAllOperations()

    @Query("SELECT * FROM sync_operations WHERE retryCount >= maxRetries")
    suspend fun getFailedOperations(): List<SyncOperation>

    @Query("SELECT COUNT(*) FROM sync_operations WHERE retryCount >= maxRetries")
    suspend fun getFailedOperationsCount(): Int

    // Get operations by type
    @Query("SELECT * FROM sync_operations WHERE type = :type ORDER BY timestamp ASC")
    suspend fun getOperationsByType(type: SyncOperationType): List<SyncOperation>

    // Get operations for specific media item
    @Query("SELECT * FROM sync_operations WHERE mediaId = :mediaId ORDER BY timestamp ASC")
    suspend fun getOperationsForMedia(mediaId: Long): List<SyncOperation>

    // Clean up old successful operations (older than 30 days)
    @Query("DELETE FROM sync_operations WHERE timestamp < :cutoffTime AND retryCount >= maxRetries")
    suspend fun cleanupOldOperations(cutoffTime: Long)
}