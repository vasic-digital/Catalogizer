package com.catalogizer.androidtv.data.local

import org.junit.Assert.*
import org.junit.Test

class ConvertersTest2 {

    // --- String List ---

    @Test
    fun `fromStringList converts list to JSON`() {
        val result = Converters.fromStringList(listOf("Action", "Drama"))
        assertNotNull(result)
        assertTrue(result!!.contains("Action"))
        assertTrue(result.contains("Drama"))
    }

    @Test
    fun `fromStringList returns null for null input`() {
        assertNull(Converters.fromStringList(null))
    }

    @Test
    fun `fromStringList converts empty list`() {
        val result = Converters.fromStringList(emptyList())
        assertEquals("[]", result)
    }

    @Test
    fun `toStringList converts JSON to list`() {
        val result = Converters.toStringList("""["Action","Drama"]""")
        assertNotNull(result)
        assertEquals(2, result!!.size)
        assertEquals("Action", result[0])
        assertEquals("Drama", result[1])
    }

    @Test
    fun `toStringList returns null for null input`() {
        assertNull(Converters.toStringList(null))
    }

    @Test
    fun `toStringList converts empty JSON array`() {
        val result = Converters.toStringList("[]")
        assertNotNull(result)
        assertTrue(result!!.isEmpty())
    }

    @Test
    fun `string list round trip`() {
        val original = listOf("Sci-Fi", "Action", "Thriller")
        val json = Converters.fromStringList(original)
        val restored = Converters.toStringList(json)
        assertEquals(original, restored)
    }

    // --- String Map ---

    @Test
    fun `fromStringMap converts map to JSON`() {
        val map = mapOf("key1" to "value1", "key2" to "value2")
        val result = Converters.fromStringMap(map)
        assertNotNull(result)
        assertTrue(result!!.contains("key1"))
    }

    @Test
    fun `fromStringMap returns null for null input`() {
        assertNull(Converters.fromStringMap(null))
    }

    @Test
    fun `toStringMap converts JSON to map`() {
        val result = Converters.toStringMap("""{"key1":"value1","key2":"value2"}""")
        assertNotNull(result)
        assertEquals(2, result!!.size)
        assertEquals("value1", result["key1"])
    }

    @Test
    fun `toStringMap returns null for null input`() {
        assertNull(Converters.toStringMap(null))
    }

    @Test
    fun `string map round trip`() {
        val original = mapOf("title" to "Test", "year" to "2024")
        val json = Converters.fromStringMap(original)
        val restored = Converters.toStringMap(json)
        assertEquals(original, restored)
    }

    // --- ExternalMetadata List ---

    @Test
    fun `fromExternalMetadataList returns null for null`() {
        assertNull(Converters.fromExternalMetadataList(null))
    }

    @Test
    fun `fromExternalMetadataList converts empty list`() {
        val result = Converters.fromExternalMetadataList(emptyList())
        assertEquals("[]", result)
    }

    @Test
    fun `toExternalMetadataList returns null for null`() {
        assertNull(Converters.toExternalMetadataList(null))
    }

    @Test
    fun `toExternalMetadataList converts empty JSON array`() {
        val result = Converters.toExternalMetadataList("[]")
        assertNotNull(result)
        assertTrue(result!!.isEmpty())
    }

    // --- MediaVersion List ---

    @Test
    fun `fromMediaVersionList returns null for null`() {
        assertNull(Converters.fromMediaVersionList(null))
    }

    @Test
    fun `fromMediaVersionList converts empty list`() {
        val result = Converters.fromMediaVersionList(emptyList())
        assertEquals("[]", result)
    }

    @Test
    fun `toMediaVersionList returns null for null`() {
        assertNull(Converters.toMediaVersionList(null))
    }

    @Test
    fun `toMediaVersionList converts empty JSON array`() {
        val result = Converters.toMediaVersionList("[]")
        assertNotNull(result)
        assertTrue(result!!.isEmpty())
    }
}
