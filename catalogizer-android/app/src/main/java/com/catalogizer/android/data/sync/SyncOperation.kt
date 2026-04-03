package com.catalogizer.android.data.sync

import androidx.room.Entity
import androidx.room.PrimaryKey
import kotlinx.serialization.Serializable

/**
 * Room entity representing a queued offline operation pending synchronization
 * with the remote server. Tracks retry count and max retries for fault tolerance.
 */
@Entity(tableName = "sync_operations")
@Serializable
data class SyncOperation(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    val type: SyncOperationType,
    val mediaId: Long,
    val data: String?, // JSON data for the operation
    val timestamp: Long,
    val retryCount: Int = 0,
    val maxRetries: Int = 3
)

/**
 * Enumeration of supported offline sync operation types.
 */
@Serializable
enum class SyncOperationType {
    UPDATE_PROGRESS,
    TOGGLE_FAVORITE,
    UPLOAD_RATING,
    UPDATE_METADATA,
    DELETE_MEDIA
}