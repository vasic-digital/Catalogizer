package com.catalogizer.android.data.local

import androidx.paging.PagingSource
import androidx.room.*
import com.catalogizer.android.data.models.MediaItem
import kotlinx.coroutines.flow.Flow

@Dao
interface MediaDao {

    @Query("SELECT * FROM media_items ORDER BY updated_at DESC")
    fun getAllMediaPaging(): PagingSource<Int, MediaItem>

    @Query("SELECT * FROM media_items WHERE media_type = :mediaType ORDER BY updated_at DESC")
    fun getMediaByTypePaging(mediaType: String): PagingSource<Int, MediaItem>

    @Query("""
        SELECT * FROM media_items
        WHERE title LIKE '%' || :query || '%'
        OR description LIKE '%' || :query || '%'
        ORDER BY updated_at DESC
    """)
    fun searchMediaPaging(query: String): PagingSource<Int, MediaItem>

    @Query("SELECT * FROM media_items WHERE id = :id")
    suspend fun getMediaById(id: Long): MediaItem?

    @Query("SELECT * FROM media_items WHERE id = :id")
    fun getMediaByIdFlow(id: Long): Flow<MediaItem?>

    @Query("SELECT * FROM media_items WHERE is_favorite = 1 ORDER BY updated_at DESC")
    fun getFavoritesPaging(): PagingSource<Int, MediaItem>

    @Query("SELECT * FROM media_items WHERE is_downloaded = 1 ORDER BY updated_at DESC")
    fun getDownloadedPaging(): PagingSource<Int, MediaItem>

    @Query("SELECT * FROM media_items WHERE watch_progress > 0 AND watch_progress < 1 ORDER BY last_watched DESC")
    fun getContinueWatchingPaging(): PagingSource<Int, MediaItem>

    @Query("SELECT * FROM media_items ORDER BY created_at DESC LIMIT :limit")
    fun getRecentlyAdded(limit: Int = 10): Flow<List<MediaItem>>

    @Query("SELECT * FROM media_items WHERE rating IS NOT NULL ORDER BY rating DESC LIMIT :limit")
    fun getTopRated(limit: Int = 10): Flow<List<MediaItem>>

    @Query("SELECT DISTINCT media_type FROM media_items ORDER BY media_type")
    fun getAllMediaTypes(): Flow<List<String>>

    @Query("SELECT COUNT(*) FROM media_items")
    fun getTotalCount(): Flow<Int>

    @Query("SELECT COUNT(*) FROM media_items WHERE media_type = :mediaType")
    fun getCountByType(mediaType: String): Flow<Int>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertMedia(mediaItem: MediaItem)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAllMedia(mediaItems: List<MediaItem>)

    @Update
    suspend fun updateMedia(mediaItem: MediaItem)

    @Query("UPDATE media_items SET is_favorite = :isFavorite WHERE id = :id")
    suspend fun updateFavoriteStatus(id: Long, isFavorite: Boolean)

    @Query("UPDATE media_items SET watch_progress = :progress, last_watched = :lastWatched WHERE id = :id")
    suspend fun updateWatchProgress(id: Long, progress: Double, lastWatched: String)

    @Query("UPDATE media_items SET is_downloaded = :isDownloaded WHERE id = :id")
    suspend fun updateDownloadStatus(id: Long, isDownloaded: Boolean)

    @Delete
    suspend fun deleteMedia(mediaItem: MediaItem)

    @Query("DELETE FROM media_items WHERE id = :id")
    suspend fun deleteMediaById(id: Long)

    @Query("DELETE FROM media_items WHERE id = :id")
    suspend fun deleteById(id: Long)

    @Query("DELETE FROM media_items")
    suspend fun deleteAllMedia()

    @Query("DELETE FROM media_items WHERE updated_at < :timestamp")
    suspend fun deleteOldMedia(timestamp: String)

    // Cached data methods for offline repository
    @Query("SELECT * FROM media_items ORDER BY updated_at DESC")
    suspend fun getAllCached(): List<MediaItem>

    @Query("SELECT * FROM media_items WHERE id = :id")
    suspend fun getById(id: Long): MediaItem?

    @Query("SELECT * FROM media_items WHERE media_type = :type ORDER BY updated_at DESC")
    suspend fun getByType(type: String): List<MediaItem>

    @Query("SELECT * FROM media_items WHERE title LIKE '%' || :query || '%' OR description LIKE '%' || :query || '%' ORDER BY updated_at DESC")
    suspend fun searchCached(query: String): List<MediaItem>

    @Query("SELECT COUNT(*) FROM media_items")
    suspend fun getCachedItemsCount(): Int

    @Query("UPDATE media_items SET rating = :rating WHERE id = :id")
    suspend fun updateRating(id: Long, rating: Double)

    @Query("SELECT SUM(file_size) FROM media_items WHERE is_downloaded = 1")
    suspend fun getTotalDownloadSize(): Long?

    @Query("DELETE FROM media_items WHERE updated_at < :timestamp")
    suspend fun deleteOldCachedItems(timestamp: Long)

    @Transaction
    suspend fun refreshMedia(mediaItems: List<MediaItem>) {
        deleteAllMedia()
        insertAllMedia(mediaItems)
    }

    // Insert or update media item
    @Transaction
    suspend fun insertOrUpdate(mediaItem: MediaItem) {
        val existing = getById(mediaItem.id)
        if (existing != null) {
            updateMedia(mediaItem)
        } else {
            insertMedia(mediaItem)
        }
    }
}

@Entity(tableName = "search_history")
data class SearchHistory(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    val query: String,
    val timestamp: Long = System.currentTimeMillis(),
    val resultsCount: Int = 0
)

@Dao
interface SearchHistoryDao {

    @Query("SELECT * FROM search_history ORDER BY timestamp DESC LIMIT :limit")
    fun getRecentSearches(limit: Int = 10): Flow<List<SearchHistory>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertSearch(searchHistory: SearchHistory)

    @Query("DELETE FROM search_history WHERE query = :query")
    suspend fun deleteSearch(query: String)

    @Query("DELETE FROM search_history")
    suspend fun clearHistory()

    @Query("DELETE FROM search_history WHERE timestamp < :timestamp")
    suspend fun deleteOldSearches(timestamp: Long)
}

@Entity(tableName = "download_items")
data class DownloadItem(
    @PrimaryKey
    @ColumnInfo(name = "media_id")
    val mediaId: Long,
    val title: String,
    val coverImage: String?,
    val downloadUrl: String,
    val localPath: String?,
    val progress: Float = 0f,
    val status: DownloadStatus = DownloadStatus.PENDING,
    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis(),
    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis()
)

enum class DownloadStatus {
    PENDING, DOWNLOADING, COMPLETED, FAILED, PAUSED, CANCELLED
}

@Dao
interface DownloadDao {

    @Query("SELECT * FROM download_items ORDER BY created_at DESC")
    fun getAllDownloads(): Flow<List<DownloadItem>>

    @Query("SELECT * FROM download_items WHERE status = :status")
    fun getDownloadsByStatus(status: DownloadStatus): Flow<List<DownloadItem>>

    @Query("SELECT * FROM download_items WHERE media_id = :mediaId")
    suspend fun getDownloadByMediaId(mediaId: Long): DownloadItem?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertDownload(downloadItem: DownloadItem)

    @Update
    suspend fun updateDownload(downloadItem: DownloadItem)

    @Query("UPDATE download_items SET progress = :progress, status = :status, updated_at = :updatedAt WHERE media_id = :mediaId")
    suspend fun updateDownloadProgress(mediaId: Long, progress: Float, status: DownloadStatus, updatedAt: Long = System.currentTimeMillis())

    @Delete
    suspend fun deleteDownload(downloadItem: DownloadItem)

    @Query("DELETE FROM download_items WHERE media_id = :mediaId")
    suspend fun deleteDownloadByMediaId(mediaId: Long)

    @Query("DELETE FROM download_items WHERE status = :status")
    suspend fun deleteDownloadsByStatus(status: DownloadStatus)
}