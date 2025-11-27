package com.catalogizer.androidtv.data.tv

import android.content.ContentProvider
import android.content.ContentUris
import android.content.ContentValues
import android.content.Context
import android.content.UriMatcher
import android.database.Cursor
import android.database.SQLException
import android.net.Uri
import android.provider.BaseColumns

class CatalogizerTvProvider : ContentProvider() {
    
    companion object {
        // URI matcher codes
        private const val MEDIA = 100
        private const val MEDIA_ID = 101
        private const val CATEGORIES = 200
        private const val CATEGORY_ID = 201
        
        private val uriMatcher = UriMatcher(UriMatcher.NO_MATCH).apply {
            addURI(
                CatalogizerTvContract.CONTENT_AUTHORITY,
                CatalogizerTvContract.PATH_MEDIA,
                MEDIA
            )
            addURI(
                CatalogizerTvContract.CONTENT_AUTHORITY,
                "${CatalogizerTvContract.PATH_MEDIA}/#",
                MEDIA_ID
            )
            addURI(
                CatalogizerTvContract.CONTENT_AUTHORITY,
                CatalogizerTvContract.PATH_CATEGORIES,
                CATEGORIES
            )
            addURI(
                CatalogizerTvContract.CONTENT_AUTHORITY,
                "${CatalogizerTvContract.PATH_CATEGORIES}/#",
                CATEGORY_ID
            )
        }
    }
    
    private lateinit var dbHelper: TvDatabaseHelper
    
    override fun onCreate(): Boolean {
        dbHelper = TvDatabaseHelper(requireContext())
        return true
    }
    
    override fun query(
        uri: Uri,
        projection: Array<out String>?,
        selection: String?,
        selectionArgs: Array<out String>?,
        sortOrder: String?
    ): Cursor? {
        val database = dbHelper.readableDatabase
        return when (uriMatcher.match(uri)) {
            MEDIA -> {
                database.query(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    projection,
                    selection,
                    selectionArgs,
                    null,
                    null,
                    sortOrder
                )
            }
            MEDIA_ID -> {
                database.query(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    projection,
                    "${BaseColumns._ID} = ?",
                    arrayOf(uri.lastPathSegment),
                    null,
                    null,
                    sortOrder
                )
            }
            CATEGORIES -> {
                database.query(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    projection,
                    selection,
                    selectionArgs,
                    null,
                    null,
                    sortOrder
                )
            }
            CATEGORY_ID -> {
                database.query(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    projection,
                    "${BaseColumns._ID} = ?",
                    arrayOf(uri.lastPathSegment),
                    null,
                    null,
                    sortOrder
                )
            }
            else -> throw IllegalArgumentException("Unknown URI: $uri")
        }
    }
    
    override fun insert(uri: Uri, values: ContentValues?): Uri? {
        val database = dbHelper.writableDatabase
        
        return when (uriMatcher.match(uri)) {
            MEDIA -> {
                val id = database.insert(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    null,
                    values
                )
                if (id > 0) {
                    val returnUri = ContentUris.withAppendedId(
                        CatalogizerTvContract.MediaEntry.CONTENT_URI,
                        id
                    )
                    context?.contentResolver?.notifyChange(returnUri, null)
                    returnUri
                } else {
                    throw SQLException("Failed to insert row into $uri")
                }
            }
            CATEGORIES -> {
                val id = database.insert(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    null,
                    values
                )
                if (id > 0) {
                    val returnUri = ContentUris.withAppendedId(
                        CatalogizerTvContract.CategoryEntry.CONTENT_URI,
                        id
                    )
                    context?.contentResolver?.notifyChange(returnUri, null)
                    returnUri
                } else {
                    throw SQLException("Failed to insert row into $uri")
                }
            }
            else -> throw IllegalArgumentException("Unknown URI: $uri")
        }
    }
    
    override fun update(
        uri: Uri,
        values: ContentValues?,
        selection: String?,
        selectionArgs: Array<out String>?
    ): Int {
        val database = dbHelper.writableDatabase
        
        val rowsUpdated = when (uriMatcher.match(uri)) {
            MEDIA -> {
                database.update(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    values,
                    selection,
                    selectionArgs
                )
            }
            MEDIA_ID -> {
                database.update(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    values,
                    "${BaseColumns._ID} = ?",
                    arrayOf(uri.lastPathSegment)
                )
            }
            CATEGORIES -> {
                database.update(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    values,
                    selection,
                    selectionArgs
                )
            }
            CATEGORY_ID -> {
                database.update(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    values,
                    "${BaseColumns._ID} = ?",
                    arrayOf(uri.lastPathSegment)
                )
            }
            else -> throw IllegalArgumentException("Unknown URI: $uri")
        }
        
        if (rowsUpdated > 0) {
            context?.contentResolver?.notifyChange(uri, null)
        }
        
        return rowsUpdated
    }
    
    override fun delete(
        uri: Uri,
        selection: String?,
        selectionArgs: Array<out String>?
    ): Int {
        val database = dbHelper.writableDatabase
        
        val rowsDeleted = when (uriMatcher.match(uri)) {
            MEDIA -> {
                database.delete(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    selection,
                    selectionArgs
                )
            }
            MEDIA_ID -> {
                database.delete(
                    CatalogizerTvContract.MediaEntry.TABLE_NAME,
                    "${BaseColumns._ID} = ?",
                    arrayOf(uri.lastPathSegment)
                )
            }
            CATEGORIES -> {
                database.delete(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    selection,
                    selectionArgs
                )
            }
            CATEGORY_ID -> {
                database.delete(
                    CatalogizerTvContract.CategoryEntry.TABLE_NAME,
                    "${BaseColumns._ID} = ?",
                    arrayOf(uri.lastPathSegment)
                )
            }
            else -> throw IllegalArgumentException("Unknown URI: $uri")
        }
        
        if (rowsDeleted > 0) {
            context?.contentResolver?.notifyChange(uri, null)
        }
        
        return rowsDeleted
    }
    
    override fun getType(uri: Uri): String? {
        return when (uriMatcher.match(uri)) {
            MEDIA -> CatalogizerTvContract.MediaEntry.CONTENT_TYPE
            MEDIA_ID -> CatalogizerTvContract.MediaEntry.CONTENT_ITEM_TYPE
            CATEGORIES -> CatalogizerTvContract.CategoryEntry.CONTENT_TYPE
            CATEGORY_ID -> CatalogizerTvContract.CategoryEntry.CONTENT_ITEM_TYPE
            else -> throw IllegalArgumentException("Unknown URI: $uri")
        }
    }
}