package com.catalogizer.androidtv.data.tv

import android.content.ContentUris
import android.net.Uri
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@RunWith(RobolectricTestRunner::class)
class CatalogizerTvContractTest2 {

    @Test
    fun `content authority is correct`() {
        assertEquals("com.catalogizer.androidtv.tv", CatalogizerTvContract.CONTENT_AUTHORITY)
    }

    @Test
    fun `base content URI is correct`() {
        assertEquals(
            "content://com.catalogizer.androidtv.tv",
            CatalogizerTvContract.BASE_CONTENT_URI.toString()
        )
    }

    @Test
    fun `media path constant`() {
        assertEquals("media", CatalogizerTvContract.PATH_MEDIA)
    }

    @Test
    fun `categories path constant`() {
        assertEquals("categories", CatalogizerTvContract.PATH_CATEGORIES)
    }

    // --- MediaEntry ---

    @Test
    fun `media entry table name`() {
        assertEquals("media", CatalogizerTvContract.MediaEntry.TABLE_NAME)
    }

    @Test
    fun `media entry content URI`() {
        assertEquals(
            "content://com.catalogizer.androidtv.tv/media",
            CatalogizerTvContract.MediaEntry.CONTENT_URI.toString()
        )
    }

    @Test
    fun `media entry content type`() {
        assertTrue(CatalogizerTvContract.MediaEntry.CONTENT_TYPE.startsWith("vnd.android.cursor.dir/"))
    }

    @Test
    fun `media entry content item type`() {
        assertTrue(CatalogizerTvContract.MediaEntry.CONTENT_ITEM_TYPE.startsWith("vnd.android.cursor.item/"))
    }

    @Test
    fun `media entry column constants`() {
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

    // --- CategoryEntry ---

    @Test
    fun `category entry table name`() {
        assertEquals("categories", CatalogizerTvContract.CategoryEntry.TABLE_NAME)
    }

    @Test
    fun `category entry content URI`() {
        assertEquals(
            "content://com.catalogizer.androidtv.tv/categories",
            CatalogizerTvContract.CategoryEntry.CONTENT_URI.toString()
        )
    }

    @Test
    fun `category entry column constants`() {
        assertEquals("category_name", CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME)
        assertEquals("category_type", CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE)
        assertEquals("thumbnail_url", CatalogizerTvContract.CategoryEntry.COLUMN_THUMBNAIL_URL)
    }

    // --- URI builders ---

    @Test
    fun `buildMediaUri appends ID`() {
        val uri = CatalogizerTvContract.buildMediaUri(42L)
        assertEquals(
            "content://com.catalogizer.androidtv.tv/media/42",
            uri.toString()
        )
    }

    @Test
    fun `buildCategoryUri appends ID`() {
        val uri = CatalogizerTvContract.buildCategoryUri(7L)
        assertEquals(
            "content://com.catalogizer.androidtv.tv/categories/7",
            uri.toString()
        )
    }

    @Test
    fun `getMediaIdFromUri extracts ID`() {
        val uri = ContentUris.withAppendedId(CatalogizerTvContract.MediaEntry.CONTENT_URI, 99L)
        assertEquals(99L, CatalogizerTvContract.getMediaIdFromUri(uri))
    }

    @Test
    fun `getCategoryIdFromUri extracts ID`() {
        val uri = ContentUris.withAppendedId(CatalogizerTvContract.CategoryEntry.CONTENT_URI, 5L)
        assertEquals(5L, CatalogizerTvContract.getCategoryIdFromUri(uri))
    }
}
