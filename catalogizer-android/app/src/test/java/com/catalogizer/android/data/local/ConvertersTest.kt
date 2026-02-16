package com.catalogizer.android.data.local

import com.catalogizer.android.data.models.ExternalMetadata
import com.catalogizer.android.data.models.MediaVersion
import com.catalogizer.android.data.sync.SyncOperationType
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class ConvertersTest {

    private lateinit var converters: Converters

    @Before
    fun setup() {
        converters = Converters()
    }

    // --- String List Converters ---

    @Test
    fun `fromStringList converts list to JSON string`() {
        val list = listOf("action", "drama", "thriller")
        val result = converters.fromStringList(list)

        assertNotNull(result)
        assertTrue(result!!.contains("action"))
        assertTrue(result.contains("drama"))
        assertTrue(result.contains("thriller"))
    }

    @Test
    fun `toStringList converts JSON string to list`() {
        val jsonStr = """["action","drama","thriller"]"""
        val result = converters.toStringList(jsonStr)

        assertNotNull(result)
        assertEquals(3, result?.size)
        assertEquals("action", result?.get(0))
        assertEquals("drama", result?.get(1))
        assertEquals("thriller", result?.get(2))
    }

    @Test
    fun `fromStringList handles null input`() {
        val result = converters.fromStringList(null)
        assertNull(result)
    }

    @Test
    fun `toStringList handles null input`() {
        val result = converters.toStringList(null)
        assertNull(result)
    }

    @Test
    fun `String list round-trip preserves data`() {
        val original = listOf("one", "two", "three")
        val json = converters.fromStringList(original)
        val restored = converters.toStringList(json)

        assertEquals(original, restored)
    }

    @Test
    fun `fromStringList handles empty list`() {
        val result = converters.fromStringList(emptyList())
        assertNotNull(result)
        assertEquals("[]", result)
    }

    @Test
    fun `toStringList handles empty array`() {
        val result = converters.toStringList("[]")
        assertNotNull(result)
        assertTrue(result!!.isEmpty())
    }

    // --- String Map Converters ---

    @Test
    fun `fromStringMap converts map to JSON string`() {
        val map = mapOf("key1" to "value1", "key2" to "value2")
        val result = converters.fromStringMap(map)

        assertNotNull(result)
        assertTrue(result!!.contains("key1"))
        assertTrue(result.contains("value1"))
    }

    @Test
    fun `toStringMap converts JSON string to map`() {
        val jsonStr = """{"key1":"value1","key2":"value2"}"""
        val result = converters.toStringMap(jsonStr)

        assertNotNull(result)
        assertEquals(2, result?.size)
        assertEquals("value1", result?.get("key1"))
        assertEquals("value2", result?.get("key2"))
    }

    @Test
    fun `fromStringMap handles null input`() {
        val result = converters.fromStringMap(null)
        assertNull(result)
    }

    @Test
    fun `toStringMap handles null input`() {
        val result = converters.toStringMap(null)
        assertNull(result)
    }

    @Test
    fun `String map round-trip preserves data`() {
        val original = mapOf("a" to "1", "b" to "2")
        val json = converters.fromStringMap(original)
        val restored = converters.toStringMap(json)

        assertEquals(original, restored)
    }

    // --- SyncOperationType Converters ---

    @Test
    fun `fromSyncOperationType converts enum to string`() {
        assertEquals("UPDATE_PROGRESS", converters.fromSyncOperationType(SyncOperationType.UPDATE_PROGRESS))
        assertEquals("TOGGLE_FAVORITE", converters.fromSyncOperationType(SyncOperationType.TOGGLE_FAVORITE))
        assertEquals("UPLOAD_RATING", converters.fromSyncOperationType(SyncOperationType.UPLOAD_RATING))
        assertEquals("UPDATE_METADATA", converters.fromSyncOperationType(SyncOperationType.UPDATE_METADATA))
        assertEquals("DELETE_MEDIA", converters.fromSyncOperationType(SyncOperationType.DELETE_MEDIA))
    }

    @Test
    fun `toSyncOperationType converts string to enum`() {
        assertEquals(SyncOperationType.UPDATE_PROGRESS, converters.toSyncOperationType("UPDATE_PROGRESS"))
        assertEquals(SyncOperationType.TOGGLE_FAVORITE, converters.toSyncOperationType("TOGGLE_FAVORITE"))
        assertEquals(SyncOperationType.UPLOAD_RATING, converters.toSyncOperationType("UPLOAD_RATING"))
        assertEquals(SyncOperationType.UPDATE_METADATA, converters.toSyncOperationType("UPDATE_METADATA"))
        assertEquals(SyncOperationType.DELETE_MEDIA, converters.toSyncOperationType("DELETE_MEDIA"))
    }

    @Test
    fun `fromSyncOperationType handles null input`() {
        val result = converters.fromSyncOperationType(null)
        assertNull(result)
    }

    @Test
    fun `toSyncOperationType handles null input`() {
        val result = converters.toSyncOperationType(null)
        assertNull(result)
    }

    @Test
    fun `SyncOperationType round-trip preserves data`() {
        for (type in SyncOperationType.values()) {
            val str = converters.fromSyncOperationType(type)
            val restored = converters.toSyncOperationType(str)
            assertEquals(type, restored)
        }
    }
}
