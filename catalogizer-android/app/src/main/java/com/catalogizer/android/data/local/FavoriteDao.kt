package com.catalogizer.android.data.local

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Entity(tableName = "favorites")
data class Favorite(
    @PrimaryKey
    @ColumnInfo(name = "media_id")
    val mediaId: Long,
    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis(),
    @ColumnInfo(name = "updated_at")
    val updatedAt: Long = System.currentTimeMillis()
)

@Dao
interface FavoriteDao {
    
    @Query("SELECT * FROM favorites ORDER BY updated_at DESC")
    fun getAllFavorites(): Flow<List<Favorite>>
    
    @Query("SELECT * FROM favorites WHERE media_id = :mediaId")
    suspend fun getFavorite(mediaId: Long): Favorite?
    
    @Query("SELECT * FROM favorites WHERE media_id = :mediaId")
    fun getFavoriteFlow(mediaId: Long): Flow<Favorite?>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertOrUpdate(favorite: Favorite)
    
    @Delete
    suspend fun delete(favorite: Favorite)
    
    @Query("DELETE FROM favorites WHERE media_id = :mediaId")
    suspend fun deleteByMediaId(mediaId: Long)
    
    @Query("DELETE FROM favorites")
    suspend fun deleteAll()
    
    @Query("SELECT COUNT(*) FROM favorites")
    suspend fun getFavoritesCount(): Int
    
    @Query("SELECT COUNT(*) FROM favorites")
    fun getFavoritesCountFlow(): Flow<Int>
}