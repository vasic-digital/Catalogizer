package com.catalogizer.android.data.local

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Entity(tableName = "watch_progress")
data class WatchProgress(
    @PrimaryKey
    @ColumnInfo(name = "media_id")
    val mediaId: Long,
    val progress: Double,
    @ColumnInfo(name = "last_watched")
    val lastWatched: Long = System.currentTimeMillis(),
    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis()
)

@Dao
interface WatchProgressDao {
    
    @Query("SELECT * FROM watch_progress WHERE media_id = :mediaId")
    suspend fun getProgress(mediaId: Long): WatchProgress?
    
    @Query("SELECT * FROM watch_progress WHERE media_id = :mediaId")
    fun getProgressFlow(mediaId: Long): Flow<WatchProgress?>
    
    @Query("SELECT * FROM watch_progress ORDER BY last_watched DESC")
    fun getAllProgress(): Flow<List<WatchProgress>>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertOrUpdate(progress: WatchProgress)
    
    @Update
    suspend fun update(progress: WatchProgress)
    
    @Query("DELETE FROM watch_progress WHERE media_id = :mediaId")
    suspend fun deleteByMediaId(mediaId: Long)
    
    @Query("DELETE FROM watch_progress")
    suspend fun deleteAll()
    
    @Query("DELETE FROM watch_progress WHERE last_watched < :cutoffTime")
    suspend fun deleteOldProgress(cutoffTime: Long)
}