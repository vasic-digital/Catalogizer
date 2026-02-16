package com.catalogizer.androidtv.data.tv

import android.content.ContentUris
import android.net.Uri
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class CatalogizerTvContractTest {

    @Test
    fun `CONTENT_AUTHORITY is correct`() {
        assertEquals("com.catalogizer.androidtv.tv", CatalogizerTvContract.CONTENT_AUTHORITY)
    }

    @Test
    fun `BASE_CONTENT_URI has correct format`() {
        val expected = "content://com.catalogizer.androidtv.tv"
        assertEquals(expected, CatalogizerTvContract.BASE_CONTENT_URI.toString())
    }

    @Test
    fun `PATH_MEDIA is correct`() {
        assertEquals("media", CatalogizerTvContract.PATH_MEDIA)
    }

    @Test
    fun `PATH_CATEGORIES is correct`() {
        assertEquals("categories", CatalogizerTvContract.PATH_CATEGORIES)
    }

    @Test
    fun `MediaEntry CONTENT_URI has correct path`() {
        val expected = "content://com.catalogizer.androidtv.tv/media"
        assertEquals(expected, CatalogizerTvContract.MediaEntry.CONTENT_URI.toString())
    }

    @Test
    fun `CategoryEntry CONTENT_URI has correct path`() {
        val expected = "content://com.catalogizer.androidtv.tv/categories"
        assertEquals(expected, CatalogizerTvContract.CategoryEntry.CONTENT_URI.toString())
    }

    @Test
    fun `MediaEntry has correct CONTENT_TYPE`() {
        assertTrue(CatalogizerTvContract.MediaEntry.CONTENT_TYPE.startsWith("vnd.android.cursor.dir/"))
    }

    @Test
    fun `MediaEntry has correct CONTENT_ITEM_TYPE`() {
        assertTrue(CatalogizerTvContract.MediaEntry.CONTENT_ITEM_TYPE.startsWith("vnd.android.cursor.item/"))
    }

    @Test
    fun `MediaEntry column constants are defined`() {
        assertEquals("media_id", CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID)
        assertEquals("title", CatalogizerTvContract.MediaEntry.COLUMN_TITLE)
        assertEquals("description", CatalogizerTvContract.MediaEntry.COLUMN_DESCRIPTION)
        assertEquals("duration", CatalogizerTvContract.MediaEntry.COLUMN_DURATION)
        assertEquals("thumbnail_url", CatalogizerTvContract.MediaEntry.COLUMN_THUMBNAIL_URL)
        assertEquals("video_url", CatalogizerTvContract.MediaEntry.COLUMN_VIDEO_URL)
        assertEquals("audio_url", CatalogizerTvContract.MediaEntry.COLUMN_AUDIO_URL)
        assertEquals("category", CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY)
        assertEquals("created_at", CatalogizerTvContract.MediaEntry.COLUMN_CREATED_AT)
        assertEquals("updated_at", CatalogizerTvContract.MediaEntry.COLUMN_UPDATED_AT)
    }

    @Test
    fun `CategoryEntry column constants are defined`() {
        assertEquals("category_name", CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME)
        assertEquals("category_type", CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE)
        assertEquals("thumbnail_url", CatalogizerTvContract.CategoryEntry.COLUMN_THUMBNAIL_URL)
    }

    @Test
    fun `buildMediaUri creates correct URI`() {
        val uri = CatalogizerTvContract.buildMediaUri(42L)
        assertEquals("content://com.catalogizer.androidtv.tv/media/42", uri.toString())
    }

    @Test
    fun `buildCategoryUri creates correct URI`() {
        val uri = CatalogizerTvContract.buildCategoryUri(10L)
        assertEquals("content://com.catalogizer.androidtv.tv/categories/10", uri.toString())
    }

    @Test
    fun `getMediaIdFromUri extracts correct ID`() {
        val uri = CatalogizerTvContract.buildMediaUri(42L)
        val id = CatalogizerTvContract.getMediaIdFromUri(uri)
        assertEquals(42L, id)
    }

    @Test
    fun `getCategoryIdFromUri extracts correct ID`() {
        val uri = CatalogizerTvContract.buildCategoryUri(10L)
        val id = CatalogizerTvContract.getCategoryIdFromUri(uri)
        assertEquals(10L, id)
    }

    @Test
    fun `TvDatabaseHelper constants are correct`() {
        assertEquals("catalogizer_tv.db", TvDatabaseHelper.DATABASE_NAME)
        assertEquals(1, TvDatabaseHelper.DATABASE_VERSION)
    }

    @Test
    fun `CREATE_MEDIA_TABLE SQL contains required columns`() {
        val sql = TvDatabaseHelper.CREATE_MEDIA_TABLE
        assertTrue(sql.contains(CatalogizerTvContract.MediaEntry.TABLE_NAME))
        assertTrue(sql.contains(CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID))
        assertTrue(sql.contains(CatalogizerTvContract.MediaEntry.COLUMN_TITLE))
    }

    @Test
    fun `CREATE_CATEGORIES_TABLE SQL contains required columns`() {
        val sql = TvDatabaseHelper.CREATE_CATEGORIES_TABLE
        assertTrue(sql.contains(CatalogizerTvContract.CategoryEntry.TABLE_NAME))
        assertTrue(sql.contains(CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME))
    }
}
