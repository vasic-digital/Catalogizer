package com.catalogizer.android.data.local

import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.TypeConverter
import androidx.room.TypeConverters
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase
import android.content.Context
import com.catalogizer.android.data.models.ExternalMetadata
import com.catalogizer.android.data.models.MediaItem
import com.catalogizer.android.data.models.MediaVersion
import com.catalogizer.android.data.sync.SyncOperation
import com.catalogizer.android.data.sync.SyncOperationType
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlinx.serialization.decodeFromString

@Database(
    entities = [MediaItem::class, SearchHistory::class, DownloadItem::class, SyncOperation::class, WatchProgress::class, Favorite::class],
    version = 1,
    exportSchema = false
)
@TypeConverters(Converters::class)
abstract class CatalogizerDatabase : RoomDatabase() {
    abstract fun mediaDao(): MediaDao
    abstract fun searchHistoryDao(): SearchHistoryDao
    abstract fun downloadDao(): DownloadDao
    abstract fun syncOperationDao(): SyncOperationDao
    abstract fun watchProgressDao(): WatchProgressDao
    abstract fun favoriteDao(): FavoriteDao

    companion object {
        @Volatile
        private var INSTANCE: CatalogizerDatabase? = null

        // Define migrations here as the schema evolves
        val ALL_MIGRATIONS: Array<Migration> = arrayOf(
            // Example for future use:
            // MIGRATION_1_2
        )

        fun getDatabase(context: Context): CatalogizerDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    CatalogizerDatabase::class.java,
                    "catalogizer_database"
                )
                    .addMigrations(*ALL_MIGRATIONS)
                    .fallbackToDestructiveMigration()
                    .build()
                INSTANCE = instance
                instance
            }
        }
    }
}

class Converters {
    private val json = Json { ignoreUnknownKeys = true }

    @TypeConverter
    fun fromStringList(value: List<String>?): String? {
        return value?.let { json.encodeToString(it) }
    }

    @TypeConverter
    fun toStringList(value: String?): List<String>? {
        return value?.let { json.decodeFromString(it) }
    }

    @TypeConverter
    fun fromExternalMetadataList(value: List<ExternalMetadata>?): String? {
        return value?.let { json.encodeToString(it) }
    }

    @TypeConverter
    fun toExternalMetadataList(value: String?): List<ExternalMetadata>? {
        return value?.let { json.decodeFromString(it) }
    }

    @TypeConverter
    fun fromMediaVersionList(value: List<MediaVersion>?): String? {
        return value?.let { json.encodeToString(it) }
    }

    @TypeConverter
    fun toMediaVersionList(value: String?): List<MediaVersion>? {
        return value?.let { json.decodeFromString(it) }
    }

    @TypeConverter
    fun fromStringMap(value: Map<String, String>?): String? {
        return value?.let { json.encodeToString(it) }
    }

    @TypeConverter
    fun toStringMap(value: String?): Map<String, String>? {
        return value?.let { json.decodeFromString(it) }
    }

    @TypeConverter
    fun fromSyncOperationType(value: SyncOperationType?): String? {
        return value?.name
    }

    @TypeConverter
    fun toSyncOperationType(value: String?): SyncOperationType? {
        return value?.let { SyncOperationType.valueOf(it) }
    }
}