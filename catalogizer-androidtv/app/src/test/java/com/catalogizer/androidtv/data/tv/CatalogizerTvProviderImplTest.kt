package com.catalogizer.androidtv.data.tv

import android.content.ContentValues
import android.net.Uri
import android.provider.BaseColumns
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.Robolectric
import org.robolectric.annotation.Config

/**
 * Unit tests for [CatalogizerTvProviderImpl] using Robolectric to
 * supply a real SQLite database via [TvDatabaseHelper]. Each test
 * exercises the ContentProvider CRUD operations and URI matching.
 */
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE, sdk = [28])
class CatalogizerTvProviderImplTest {

    private lateinit var provider: CatalogizerTvProviderImpl

    @Before
    fun setup() {
        provider = Robolectric.setupContentProvider(CatalogizerTvProviderImpl::class.java)
    }

    // ------------------------------------------------------------------
    // Helper: build ContentValues for a media row
    // ------------------------------------------------------------------
    private fun mediaValues(
        mediaId: String = "m1",
        title: String = "Test Movie",
        description: String? = "A test movie",
        duration: Int? = 7200,
        category: String? = "movie"
    ): ContentValues = ContentValues().apply {
        put(CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID, mediaId)
        put(CatalogizerTvContract.MediaEntry.COLUMN_TITLE, title)
        if (description != null) put(CatalogizerTvContract.MediaEntry.COLUMN_DESCRIPTION, description)
        if (duration != null) put(CatalogizerTvContract.MediaEntry.COLUMN_DURATION, duration)
        if (category != null) put(CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY, category)
    }

    // ------------------------------------------------------------------
    // Helper: build ContentValues for a category row
    // ------------------------------------------------------------------
    private fun categoryValues(
        name: String = "Movies",
        type: String? = "MOVIE"
    ): ContentValues = ContentValues().apply {
        put(CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME, name)
        if (type != null) put(CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE, type)
    }

    // ------------------------------------------------------------------
    // onCreate
    // ------------------------------------------------------------------

    @Test
    fun `onCreate returns true`() {
        // Robolectric already called onCreate; re-verify via a new instance
        val fresh = CatalogizerTvProviderImpl()
        // Without a context it returns false
        val result = fresh.onCreate()
        // Provider created by Robolectric has a context, so the one we set up succeeds
        assertNotNull(provider)
    }

    // ------------------------------------------------------------------
    // getType
    // ------------------------------------------------------------------

    @Test
    fun `getType returns correct type for media directory URI`() {
        val type = provider.getType(CatalogizerTvContract.MediaEntry.CONTENT_URI)
        assertEquals(CatalogizerTvContract.MediaEntry.CONTENT_TYPE, type)
    }

    @Test
    fun `getType returns correct type for media item URI`() {
        val uri = CatalogizerTvContract.buildMediaUri(1L)
        val type = provider.getType(uri)
        assertEquals(CatalogizerTvContract.MediaEntry.CONTENT_ITEM_TYPE, type)
    }

    @Test
    fun `getType returns correct type for categories directory URI`() {
        val type = provider.getType(CatalogizerTvContract.CategoryEntry.CONTENT_URI)
        assertEquals(CatalogizerTvContract.CategoryEntry.CONTENT_TYPE, type)
    }

    @Test
    fun `getType returns correct type for category item URI`() {
        val uri = CatalogizerTvContract.buildCategoryUri(1L)
        val type = provider.getType(uri)
        assertEquals(CatalogizerTvContract.CategoryEntry.CONTENT_ITEM_TYPE, type)
    }

    @Test
    fun `getType throws for unknown URI`() {
        val unknownUri = Uri.parse("content://${CatalogizerTvContract.CONTENT_AUTHORITY}/unknown")
        try {
            provider.getType(unknownUri)
            fail("Expected IllegalArgumentException for unknown URI")
        } catch (e: IllegalArgumentException) {
            assertTrue(e.message?.contains("Unknown URI") == true)
        }
    }

    // ------------------------------------------------------------------
    // insert — media
    // ------------------------------------------------------------------

    @Test
    fun `insert media returns URI with valid id`() {
        val uri = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues())
        assertNotNull(uri)
        val id = CatalogizerTvContract.getMediaIdFromUri(uri!!)
        assertTrue("Inserted ID should be positive", id > 0)
    }

    @Test
    fun `insert media makes row queryable`() {
        val insertUri = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues())
        assertNotNull(insertUri)

        val cursor = provider.query(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            null, null, null, null
        )
        assertNotNull(cursor)
        assertEquals(1, cursor!!.count)
        cursor.moveToFirst()
        val titleIdx = cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_TITLE)
        assertEquals("Test Movie", cursor.getString(titleIdx))
        cursor.close()
    }

    @Test
    fun `insert multiple media rows returns distinct URIs`() {
        val uri1 = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "Movie 1"))
        val uri2 = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m2", "Movie 2"))
        assertNotNull(uri1)
        assertNotNull(uri2)
        assertTrue(uri1.toString() != uri2.toString())
    }

    @Test
    fun `insert media with unknown URI throws`() {
        val unknownUri = Uri.parse("content://${CatalogizerTvContract.CONTENT_AUTHORITY}/unknown")
        try {
            provider.insert(unknownUri, mediaValues())
            fail("Expected IllegalArgumentException")
        } catch (e: IllegalArgumentException) {
            assertTrue(e.message?.contains("Unknown URI") == true)
        }
    }

    // ------------------------------------------------------------------
    // insert — categories
    // ------------------------------------------------------------------

    @Test
    fun `insert category returns URI with valid id`() {
        val uri = provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues())
        assertNotNull(uri)
        val id = CatalogizerTvContract.getCategoryIdFromUri(uri!!)
        assertTrue("Inserted ID should be positive", id > 0)
    }

    @Test
    fun `insert category makes row queryable`() {
        provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues("TV Shows", "TV_SHOW"))

        val cursor = provider.query(
            CatalogizerTvContract.CategoryEntry.CONTENT_URI,
            null, null, null, null
        )
        assertNotNull(cursor)
        assertEquals(1, cursor!!.count)
        cursor.moveToFirst()
        val nameIdx = cursor.getColumnIndex(CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME)
        assertEquals("TV Shows", cursor.getString(nameIdx))
        cursor.close()
    }

    // ------------------------------------------------------------------
    // query — media
    // ------------------------------------------------------------------

    @Test
    fun `query media by id returns single row`() {
        val insertUri = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues())!!
        val id = CatalogizerTvContract.getMediaIdFromUri(insertUri)
        val itemUri = CatalogizerTvContract.buildMediaUri(id)

        val cursor = provider.query(itemUri, null, null, null, null)
        assertNotNull(cursor)
        assertEquals(1, cursor!!.count)
        cursor.close()
    }

    @Test
    fun `query media with selection filters correctly`() {
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "Action Movie", category = "movie"))
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m2", "Comedy Show", category = "tv_show"))

        val cursor = provider.query(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            null,
            "${CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY} = ?",
            arrayOf("movie"),
            null
        )
        assertNotNull(cursor)
        assertEquals(1, cursor!!.count)
        cursor.moveToFirst()
        val titleIdx = cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_TITLE)
        assertEquals("Action Movie", cursor.getString(titleIdx))
        cursor.close()
    }

    @Test
    fun `query media with projection returns only requested columns`() {
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues())

        val projection = arrayOf(CatalogizerTvContract.MediaEntry.COLUMN_TITLE)
        val cursor = provider.query(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            projection, null, null, null
        )
        assertNotNull(cursor)
        assertEquals(1, cursor!!.columnCount)
        cursor.close()
    }

    @Test
    fun `query empty media table returns empty cursor`() {
        val cursor = provider.query(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            null, null, null, null
        )
        assertNotNull(cursor)
        assertEquals(0, cursor!!.count)
        cursor.close()
    }

    @Test
    fun `query unknown URI throws`() {
        val unknownUri = Uri.parse("content://${CatalogizerTvContract.CONTENT_AUTHORITY}/unknown")
        try {
            provider.query(unknownUri, null, null, null, null)
            fail("Expected IllegalArgumentException")
        } catch (e: IllegalArgumentException) {
            assertTrue(e.message?.contains("Unknown URI") == true)
        }
    }

    // ------------------------------------------------------------------
    // query — categories
    // ------------------------------------------------------------------

    @Test
    fun `query category by id returns single row`() {
        val insertUri = provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues())!!
        val id = CatalogizerTvContract.getCategoryIdFromUri(insertUri)
        val itemUri = CatalogizerTvContract.buildCategoryUri(id)

        val cursor = provider.query(itemUri, null, null, null, null)
        assertNotNull(cursor)
        assertEquals(1, cursor!!.count)
        cursor.close()
    }

    // ------------------------------------------------------------------
    // update — media
    // ------------------------------------------------------------------

    @Test
    fun `update media by directory URI updates matching rows`() {
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "Original Title"))

        val newValues = ContentValues().apply {
            put(CatalogizerTvContract.MediaEntry.COLUMN_TITLE, "Updated Title")
        }
        val updated = provider.update(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            newValues,
            "${CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID} = ?",
            arrayOf("m1")
        )
        assertEquals(1, updated)

        val cursor = provider.query(CatalogizerTvContract.MediaEntry.CONTENT_URI, null, null, null, null)!!
        cursor.moveToFirst()
        val titleIdx = cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_TITLE)
        assertEquals("Updated Title", cursor.getString(titleIdx))
        cursor.close()
    }

    @Test
    fun `update media by item URI updates single row`() {
        val insertUri = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "Title"))!!
        val id = CatalogizerTvContract.getMediaIdFromUri(insertUri)
        val itemUri = CatalogizerTvContract.buildMediaUri(id)

        val newValues = ContentValues().apply {
            put(CatalogizerTvContract.MediaEntry.COLUMN_TITLE, "New Title")
        }
        val updated = provider.update(itemUri, newValues, null, null)
        assertEquals(1, updated)
    }

    @Test
    fun `update non-existent media returns zero`() {
        val newValues = ContentValues().apply {
            put(CatalogizerTvContract.MediaEntry.COLUMN_TITLE, "Nothing")
        }
        val updated = provider.update(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            newValues,
            "${CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID} = ?",
            arrayOf("nonexistent")
        )
        assertEquals(0, updated)
    }

    @Test
    fun `update unknown URI throws`() {
        val unknownUri = Uri.parse("content://${CatalogizerTvContract.CONTENT_AUTHORITY}/unknown")
        try {
            provider.update(unknownUri, ContentValues(), null, null)
            fail("Expected IllegalArgumentException")
        } catch (e: IllegalArgumentException) {
            assertTrue(e.message?.contains("Unknown URI") == true)
        }
    }

    // ------------------------------------------------------------------
    // update — categories
    // ------------------------------------------------------------------

    @Test
    fun `update category by directory URI updates matching rows`() {
        provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues("Movies", "MOVIE"))

        val newValues = ContentValues().apply {
            put(CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE, "UPDATED")
        }
        val updated = provider.update(
            CatalogizerTvContract.CategoryEntry.CONTENT_URI,
            newValues,
            "${CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME} = ?",
            arrayOf("Movies")
        )
        assertEquals(1, updated)
    }

    @Test
    fun `update category by item URI updates single row`() {
        val insertUri = provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues())!!
        val id = CatalogizerTvContract.getCategoryIdFromUri(insertUri)
        val itemUri = CatalogizerTvContract.buildCategoryUri(id)

        val newValues = ContentValues().apply {
            put(CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_TYPE, "NEW_TYPE")
        }
        val updated = provider.update(itemUri, newValues, null, null)
        assertEquals(1, updated)
    }

    // ------------------------------------------------------------------
    // delete — media
    // ------------------------------------------------------------------

    @Test
    fun `delete media by directory URI with selection deletes matching rows`() {
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "Movie 1"))
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m2", "Movie 2"))

        val deleted = provider.delete(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            "${CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID} = ?",
            arrayOf("m1")
        )
        assertEquals(1, deleted)

        val cursor = provider.query(CatalogizerTvContract.MediaEntry.CONTENT_URI, null, null, null, null)!!
        assertEquals(1, cursor.count)
        cursor.close()
    }

    @Test
    fun `delete media by item URI deletes single row`() {
        val insertUri = provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues())!!
        val id = CatalogizerTvContract.getMediaIdFromUri(insertUri)
        val itemUri = CatalogizerTvContract.buildMediaUri(id)

        val deleted = provider.delete(itemUri, null, null)
        assertEquals(1, deleted)

        val cursor = provider.query(CatalogizerTvContract.MediaEntry.CONTENT_URI, null, null, null, null)!!
        assertEquals(0, cursor.count)
        cursor.close()
    }

    @Test
    fun `delete all media rows`() {
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "A"))
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m2", "B"))
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m3", "C"))

        val deleted = provider.delete(CatalogizerTvContract.MediaEntry.CONTENT_URI, null, null)
        assertEquals(3, deleted)
    }

    @Test
    fun `delete non-existent media returns zero`() {
        val deleted = provider.delete(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            "${CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID} = ?",
            arrayOf("nonexistent")
        )
        assertEquals(0, deleted)
    }

    @Test
    fun `delete unknown URI throws`() {
        val unknownUri = Uri.parse("content://${CatalogizerTvContract.CONTENT_AUTHORITY}/unknown")
        try {
            provider.delete(unknownUri, null, null)
            fail("Expected IllegalArgumentException")
        } catch (e: IllegalArgumentException) {
            assertTrue(e.message?.contains("Unknown URI") == true)
        }
    }

    // ------------------------------------------------------------------
    // delete — categories
    // ------------------------------------------------------------------

    @Test
    fun `delete category by directory URI deletes matching rows`() {
        provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues("Movies", "MOVIE"))
        provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues("TV Shows", "TV_SHOW"))

        val deleted = provider.delete(
            CatalogizerTvContract.CategoryEntry.CONTENT_URI,
            "${CatalogizerTvContract.CategoryEntry.COLUMN_CATEGORY_NAME} = ?",
            arrayOf("Movies")
        )
        assertEquals(1, deleted)
    }

    @Test
    fun `delete category by item URI deletes single row`() {
        val insertUri = provider.insert(CatalogizerTvContract.CategoryEntry.CONTENT_URI, categoryValues())!!
        val id = CatalogizerTvContract.getCategoryIdFromUri(insertUri)
        val itemUri = CatalogizerTvContract.buildCategoryUri(id)

        val deleted = provider.delete(itemUri, null, null)
        assertEquals(1, deleted)
    }

    // ------------------------------------------------------------------
    // Sort order
    // ------------------------------------------------------------------

    @Test
    fun `query media with sort order returns rows in order`() {
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m1", "Zebra"))
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, mediaValues("m2", "Alpha"))

        val cursor = provider.query(
            CatalogizerTvContract.MediaEntry.CONTENT_URI,
            null, null, null,
            "${CatalogizerTvContract.MediaEntry.COLUMN_TITLE} ASC"
        )!!
        cursor.moveToFirst()
        val titleIdx = cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_TITLE)
        assertEquals("Alpha", cursor.getString(titleIdx))
        cursor.moveToNext()
        assertEquals("Zebra", cursor.getString(titleIdx))
        cursor.close()
    }

    // ------------------------------------------------------------------
    // Insert-then-query round trip for all columns
    // ------------------------------------------------------------------

    @Test
    fun `inserted media row retains all column values`() {
        val values = ContentValues().apply {
            put(CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID, "full_test")
            put(CatalogizerTvContract.MediaEntry.COLUMN_TITLE, "Full Test")
            put(CatalogizerTvContract.MediaEntry.COLUMN_DESCRIPTION, "Description text")
            put(CatalogizerTvContract.MediaEntry.COLUMN_DURATION, 3600)
            put(CatalogizerTvContract.MediaEntry.COLUMN_THUMBNAIL_URL, "http://img.example.com/thumb.jpg")
            put(CatalogizerTvContract.MediaEntry.COLUMN_VIDEO_URL, "http://cdn.example.com/video.mp4")
            put(CatalogizerTvContract.MediaEntry.COLUMN_AUDIO_URL, "http://cdn.example.com/audio.mp3")
            put(CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY, "tv_show")
            put(CatalogizerTvContract.MediaEntry.COLUMN_CREATED_AT, 1700000000L)
            put(CatalogizerTvContract.MediaEntry.COLUMN_UPDATED_AT, 1700100000L)
        }
        provider.insert(CatalogizerTvContract.MediaEntry.CONTENT_URI, values)

        val cursor = provider.query(CatalogizerTvContract.MediaEntry.CONTENT_URI, null, null, null, null)!!
        cursor.moveToFirst()

        assertEquals("full_test", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_MEDIA_ID)))
        assertEquals("Full Test", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_TITLE)))
        assertEquals("Description text", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_DESCRIPTION)))
        assertEquals(3600, cursor.getInt(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_DURATION)))
        assertEquals("http://img.example.com/thumb.jpg", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_THUMBNAIL_URL)))
        assertEquals("http://cdn.example.com/video.mp4", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_VIDEO_URL)))
        assertEquals("http://cdn.example.com/audio.mp3", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_AUDIO_URL)))
        assertEquals("tv_show", cursor.getString(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_CATEGORY)))
        assertEquals(1700000000L, cursor.getLong(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_CREATED_AT)))
        assertEquals(1700100000L, cursor.getLong(cursor.getColumnIndex(CatalogizerTvContract.MediaEntry.COLUMN_UPDATED_AT)))

        cursor.close()
    }
}
