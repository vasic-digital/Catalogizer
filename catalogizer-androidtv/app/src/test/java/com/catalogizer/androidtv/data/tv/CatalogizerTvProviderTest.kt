package com.catalogizer.androidtv.data.tv

import android.net.Uri
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

/**
 * Tests for [CatalogizerTvContract] helper methods and URI builder
 * logic, and structural validation of the [TvDatabaseHelper] SQL.
 * The [CatalogizerTvContractTest] covers constants; this file
 * focuses on the helper methods' edge cases and the database
 * helper's upgrade path.
 */
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE, sdk = [28])
class CatalogizerTvProviderTest {

    // ------------------------------------------------------------------
    // URI builder round-trip tests
    // ------------------------------------------------------------------

    @Test
    fun `buildMediaUri and getMediaIdFromUri are inverse operations`() {
        for (id in listOf(0L, 1L, 42L, Long.MAX_VALUE)) {
            val uri = CatalogizerTvContract.buildMediaUri(id)
            val extracted = CatalogizerTvContract.getMediaIdFromUri(uri)
            assertEquals("Round-trip failed for id=$id", id, extracted)
        }
    }

    @Test
    fun `buildCategoryUri and getCategoryIdFromUri are inverse operations`() {
        for (id in listOf(0L, 1L, 99L, 1000L)) {
            val uri = CatalogizerTvContract.buildCategoryUri(id)
            val extracted = CatalogizerTvContract.getCategoryIdFromUri(uri)
            assertEquals("Round-trip failed for id=$id", id, extracted)
        }
    }

    @Test
    fun `buildMediaUri produces content scheme`() {
        val uri = CatalogizerTvContract.buildMediaUri(1L)
        assertEquals("content", uri.scheme)
    }

    @Test
    fun `buildMediaUri includes authority`() {
        val uri = CatalogizerTvContract.buildMediaUri(1L)
        assertEquals(CatalogizerTvContract.CONTENT_AUTHORITY, uri.authority)
    }

    @Test
    fun `buildMediaUri includes media path segment`() {
        val uri = CatalogizerTvContract.buildMediaUri(5L)
        val segments = uri.pathSegments
        assertTrue(segments.contains("media"))
    }

    @Test
    fun `buildCategoryUri produces content scheme`() {
        val uri = CatalogizerTvContract.buildCategoryUri(1L)
        assertEquals("content", uri.scheme)
    }

    @Test
    fun `buildCategoryUri includes categories path segment`() {
        val uri = CatalogizerTvContract.buildCategoryUri(5L)
        val segments = uri.pathSegments
        assertTrue(segments.contains("categories"))
    }

    // ------------------------------------------------------------------
    // MediaEntry CONTENT_URI structure
    // ------------------------------------------------------------------

    @Test
    fun `MediaEntry CONTENT_URI is a child of BASE_CONTENT_URI`() {
        val base = CatalogizerTvContract.BASE_CONTENT_URI.toString()
        val media = CatalogizerTvContract.MediaEntry.CONTENT_URI.toString()
        assertTrue(
            "MediaEntry URI should start with base URI",
            media.startsWith(base)
        )
    }

    @Test
    fun `CategoryEntry CONTENT_URI is a child of BASE_CONTENT_URI`() {
        val base = CatalogizerTvContract.BASE_CONTENT_URI.toString()
        val category = CatalogizerTvContract.CategoryEntry.CONTENT_URI.toString()
        assertTrue(
            "CategoryEntry URI should start with base URI",
            category.startsWith(base)
        )
    }

    // ------------------------------------------------------------------
    // CONTENT_TYPE MIME format
    // ------------------------------------------------------------------

    @Test
    fun `MediaEntry CONTENT_TYPE is cursor dir type`() {
        assertTrue(CatalogizerTvContract.MediaEntry.CONTENT_TYPE.startsWith("vnd.android.cursor.dir/"))
    }

    @Test
    fun `MediaEntry CONTENT_ITEM_TYPE is cursor item type`() {
        assertTrue(CatalogizerTvContract.MediaEntry.CONTENT_ITEM_TYPE.startsWith("vnd.android.cursor.item/"))
    }

    @Test
    fun `CategoryEntry CONTENT_TYPE is cursor dir type`() {
        assertTrue(CatalogizerTvContract.CategoryEntry.CONTENT_TYPE.startsWith("vnd.android.cursor.dir/"))
    }

    @Test
    fun `CategoryEntry CONTENT_ITEM_TYPE is cursor item type`() {
        assertTrue(CatalogizerTvContract.CategoryEntry.CONTENT_ITEM_TYPE.startsWith("vnd.android.cursor.item/"))
    }

    @Test
    fun `MediaEntry and CategoryEntry CONTENT_TYPEs are distinct`() {
        assertTrue(
            CatalogizerTvContract.MediaEntry.CONTENT_TYPE !=
                CatalogizerTvContract.CategoryEntry.CONTENT_TYPE
        )
    }

    // ------------------------------------------------------------------
    // TABLE_NAME constants
    // ------------------------------------------------------------------

    @Test
    fun `MediaEntry TABLE_NAME is media`() {
        assertEquals("media", CatalogizerTvContract.MediaEntry.TABLE_NAME)
    }

    @Test
    fun `CategoryEntry TABLE_NAME is categories`() {
        assertEquals("categories", CatalogizerTvContract.CategoryEntry.TABLE_NAME)
    }

    // ------------------------------------------------------------------
    // SQL statements structure
    // ------------------------------------------------------------------

    @Test
    fun `CREATE_MEDIA_TABLE has PRIMARY KEY`() {
        assertTrue(TvDatabaseHelper.CREATE_MEDIA_TABLE.contains("PRIMARY KEY"))
    }

    @Test
    fun `CREATE_MEDIA_TABLE has UNIQUE constraint on media_id`() {
        assertTrue(TvDatabaseHelper.CREATE_MEDIA_TABLE.contains("UNIQUE"))
    }

    @Test
    fun `CREATE_MEDIA_TABLE has NOT NULL on title`() {
        val sql = TvDatabaseHelper.CREATE_MEDIA_TABLE
        // title column should be NOT NULL
        assertTrue(sql.contains("title"))
        assertTrue(sql.contains("NOT NULL"))
    }

    @Test
    fun `CREATE_CATEGORIES_TABLE has PRIMARY KEY`() {
        assertTrue(TvDatabaseHelper.CREATE_CATEGORIES_TABLE.contains("PRIMARY KEY"))
    }

    @Test
    fun `CREATE_CATEGORIES_TABLE has UNIQUE on category_name`() {
        assertTrue(TvDatabaseHelper.CREATE_CATEGORIES_TABLE.contains("UNIQUE"))
    }

    @Test
    fun `DROP statements reference correct table names`() {
        assertTrue(
            TvDatabaseHelper.DROP_MEDIA_TABLE.contains(CatalogizerTvContract.MediaEntry.TABLE_NAME)
        )
        assertTrue(
            TvDatabaseHelper.DROP_CATEGORIES_TABLE.contains(CatalogizerTvContract.CategoryEntry.TABLE_NAME)
        )
    }

    @Test
    fun `DROP statements use IF EXISTS`() {
        assertTrue(TvDatabaseHelper.DROP_MEDIA_TABLE.contains("IF EXISTS"))
        assertTrue(TvDatabaseHelper.DROP_CATEGORIES_TABLE.contains("IF EXISTS"))
    }

    // ------------------------------------------------------------------
    // Database version and name
    // ------------------------------------------------------------------

    @Test
    fun `DATABASE_VERSION is 1`() {
        assertEquals(1, TvDatabaseHelper.DATABASE_VERSION)
    }

    @Test
    fun `DATABASE_NAME ends with dot db`() {
        assertTrue(TvDatabaseHelper.DATABASE_NAME.endsWith(".db"))
    }

    // ------------------------------------------------------------------
    // Column completeness in SQL
    // ------------------------------------------------------------------

    @Test
    fun `CREATE_MEDIA_TABLE contains all defined media columns`() {
        val sql = TvDatabaseHelper.CREATE_MEDIA_TABLE
        val expectedColumns = listOf(
            CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID,
            CatalogizerTvContract.MediaEntry.COLUMN_TITLE,
            CatalogizerTvContract.MediaEntry.COLUMN_DESCRIPTION,
            CatalogizerTvContract.MediaEntry.COLUMN_DURATION,
            CatalogizerTvContract.MediaEntry.COLUMN_THUMBNAIL_URL,
            CatalogizerTvContract.MediaEntry.COLUMN_VIDEO_URL,
            CatalogizerTvContract.MediaEntry.COLUMN_AUDIO_URL,
            CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY,
            CatalogizerTvContract.MediaEntry.COLUMN_CREATED_AT,
            CatalogizerTvContract.MediaEntry.COLUMN_UPDATED_AT,
        )
        for (col in expectedColumns) {
            assertTrue("SQL must contain column '$col'", sql.contains(col))
        }
    }

    @Test
    fun `CREATE_CATEGORIES_TABLE contains all defined category columns`() {
        val sql = TvDatabaseHelper.CREATE_CATEGORIES_TABLE
        val expectedColumns = listOf(
            CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME,
            CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE,
            CatalogizerTvContract.CategoryEntry.COLUMN_THUMBNAIL_URL,
        )
        for (col in expectedColumns) {
            assertTrue("SQL must contain column '$col'", sql.contains(col))
        }
    }

    // ------------------------------------------------------------------
    // TvDatabaseHelper — Robolectric lifecycle
    // ------------------------------------------------------------------

    @Test
    fun `TvDatabaseHelper can be instantiated with Robolectric context`() {
        val context = androidx.test.core.app.ApplicationProvider.getApplicationContext<android.app.Application>()
        val helper = TvDatabaseHelper(context)
        assertNotNull(helper)
        helper.close()
    }

    @Test
    fun `TvDatabaseHelper creates tables on first access`() {
        val context = androidx.test.core.app.ApplicationProvider.getApplicationContext<android.app.Application>()
        val helper = TvDatabaseHelper(context)
        val db = helper.readableDatabase
        assertNotNull(db)
        assertTrue(db.isOpen)
        helper.close()
    }

    @Test
    fun `TvDatabaseHelper writable database is accessible`() {
        val context = androidx.test.core.app.ApplicationProvider.getApplicationContext<android.app.Application>()
        val helper = TvDatabaseHelper(context)
        val db = helper.writableDatabase
        assertNotNull(db)
        assertTrue(db.isOpen)
        helper.close()
    }
}
