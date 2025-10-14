package com.catalogizer.android.data.sync

import androidx.room.Entity
import androidx.room.PrimaryKey
import kotlinx.serialization.Serializable

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

@Serializable
enum class SyncOperationType {
    UPDATE_PROGRESS,
    TOGGLE_FAVORITE,
    UPLOAD_RATING,
    UPDATE_METADATA,
    DELETE_MEDIA
}