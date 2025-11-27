package com.catalogizer.androidtv.data.tv

import android.content.ContentUris
import android.content.Context
import android.database.Cursor
import android.net.Uri
import android.provider.BaseColumns

/**
 * TV Content Provider for Android TV integration
 * Provides access to media metadata for the TV home screen
 */
object CatalogizerTvContract {
    
    const val CONTENT_AUTHORITY = "com.catalogizer.androidtv.tv"
    val BASE_CONTENT_URI: Uri = Uri.parse("content://$CONTENT_AUTHORITY")
    
    // Path constants
    const val PATH_MEDIA = "media"
    const val PATH_CATEGORIES = "categories"
    
    // Media table constants
    object MediaEntry : BaseColumns {
        const val TABLE_NAME = "media"
        
        val CONTENT_URI: Uri = 
            Uri.withAppendedPath(BASE_CONTENT_URI, PATH_MEDIA)
        
        const val CONTENT_TYPE = 
            "vnd.android.cursor.dir/vnd.$CONTENT_AUTHORITY.$PATH_MEDIA"
        const val CONTENT_ITEM_TYPE = 
            "vnd.android.cursor.item/vnd.$CONTENT_AUTHORITY.$PATH_MEDIA"
        
        // Media columns
        const val COLUMN_MEDIA_ID = "media_id"
        const val COLUMN_TITLE = "title"
        const val COLUMN_DESCRIPTION = "description"
        const val COLUMN_DURATION = "duration"
        const val COLUMN_THUMBNAIL_URL = "thumbnail_url"
        const val COLUMN_VIDEO_URL = "video_url"
        const val COLUMN_AUDIO_URL = "audio_url"
        const val COLUMN_CATEGORY = "category"
        const val COLUMN_CREATED_AT = "created_at"
        const val COLUMN_UPDATED_AT = "updated_at"
    }
    
    // Categories table constants
    object CategoryEntry : BaseColumns {
        const val TABLE_NAME = "categories"
        
        val CONTENT_URI: Uri = 
            Uri.withAppendedPath(BASE_CONTENT_URI, PATH_CATEGORIES)
        
        const val CONTENT_TYPE = 
            "vnd.android.cursor.dir/vnd.$CONTENT_AUTHORITY.$PATH_CATEGORIES"
        const val CONTENT_ITEM_TYPE = 
            "vnd.android.cursor.item/vnd.$CONTENT_AUTHORITY.$PATH_CATEGORIES"
        
        // Category columns
        const val COLUMN_CATEGORY_NAME = "category_name"
        const val COLUMN_CATEGORY_TYPE = "category_type" // MOVIE, TV_SHOW, MUSIC, etc.
        const val COLUMN_THUMBNAIL_URL = "thumbnail_url"
    }
    
    // Helper methods
    fun buildMediaUri(id: Long): Uri {
        return ContentUris.withAppendedId(MediaEntry.CONTENT_URI, id)
    }
    
    fun buildCategoryUri(id: Long): Uri {
        return ContentUris.withAppendedId(CategoryEntry.CONTENT_URI, id)
    }
    
    fun getMediaIdFromUri(uri: Uri): Long {
        return ContentUris.parseId(uri)
    }
    
    fun getCategoryIdFromUri(uri: Uri): Long {
        return ContentUris.parseId(uri)
    }
}

/**
 * Database helper for the TV provider
 * Manages the SQLite database that stores media metadata
 */
class TvDatabaseHelper(context: Context) : android.database.sqlite.SQLiteOpenHelper(
    context,
    DATABASE_NAME,
    null,
    DATABASE_VERSION
) {
    companion object {
        const val DATABASE_NAME = "catalogizer_tv.db"
        const val DATABASE_VERSION = 1
        
        // SQL statements for table creation
        const val CREATE_MEDIA_TABLE = """
            CREATE TABLE ${CatalogizerTvContract.MediaEntry.TABLE_NAME} (
                ${BaseColumns._ID} INTEGER PRIMARY KEY AUTOINCREMENT,
                ${CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID} TEXT NOT NULL UNIQUE,
                ${CatalogizerTvContract.MediaEntry.COLUMN_TITLE} TEXT NOT NULL,
                ${CatalogizerTvContract.MediaEntry.COLUMN_DESCRIPTION} TEXT,
                ${CatalogizerTvContract.MediaEntry.COLUMN_DURATION} INTEGER,
                ${CatalogizerTvContract.MediaEntry.COLUMN_THUMBNAIL_URL} TEXT,
                ${CatalogizerTvContract.MediaEntry.COLUMN_VIDEO_URL} TEXT,
                ${CatalogizerTvContract.MediaEntry.COLUMN_AUDIO_URL} TEXT,
                ${CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY} TEXT,
                ${CatalogizerTvContract.MediaEntry.COLUMN_CREATED_AT} INTEGER,
                ${CatalogizerTvContract.MediaEntry.COLUMN_UPDATED_AT} INTEGER
            )
        """
        
        const val CREATE_CATEGORIES_TABLE = """
            CREATE TABLE ${CatalogizerTvContract.CategoryEntry.TABLE_NAME} (
                ${BaseColumns._ID} INTEGER PRIMARY KEY AUTOINCREMENT,
                ${CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME} TEXT NOT NULL UNIQUE,
                ${CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE} TEXT,
                ${CatalogizerTvContract.CategoryEntry.COLUMN_THUMBNAIL_URL} TEXT
            )
        """
        
        const val DROP_MEDIA_TABLE = 
            "DROP TABLE IF EXISTS ${CatalogizerTvContract.MediaEntry.TABLE_NAME}"
        const val DROP_CATEGORIES_TABLE = 
            "DROP TABLE IF EXISTS ${CatalogizerTvContract.CategoryEntry.TABLE_NAME}"
    }
    
    override fun onCreate(db: android.database.sqlite.SQLiteDatabase) {
        db.execSQL(CREATE_MEDIA_TABLE)
        db.execSQL(CREATE_CATEGORIES_TABLE)
    }
    
    override fun onUpgrade(
        db: android.database.sqlite.SQLiteDatabase,
        oldVersion: Int,
        newVersion: Int
    ) {
        // Handle database upgrades as needed
        db.execSQL(DROP_MEDIA_TABLE)
        db.execSQL(DROP_CATEGORIES_TABLE)
        onCreate(db)
    }
}