package com.catalogizer.androidtv.data.local

import com.catalogizer.androidtv.data.models.ExternalMetadata
import com.catalogizer.androidtv.data.models.MediaVersion
import org.junit.Assert.*
import org.junit.Test

class ConvertersTest {

    @Test
    fun `fromStringList converts list to JSON string`() {
        val list = listOf("action", "drama", "thriller")
        val result = Converters.fromStringList(list)

        assertTrue(result.contains("action"))
        assertTrue(result.contains("drama"))
        assertTrue(result.contains("thriller"))
    }

    @Test
    fun `toStringList converts JSON string to list`() {
        val jsonStr = """["action","drama","thriller"]"""
        val result = Converters.toStringList(jsonStr)

        assertEquals(3, result.size)
        assertEquals("action", result[0])
        assertEquals("drama", result[1])
        assertEquals("thriller", result[2])
    }

    @Test
    fun `String list round-trip preserves data`() {
        val original = listOf("one", "two", "three")
        val json = Converters.fromStringList(original)
        val restored = Converters.toStringList(json)

        assertEquals(original, restored)
    }

    @Test
    fun `fromStringList handles empty list`() {
        val result = Converters.fromStringList(emptyList())
        assertEquals("[]", result)
    }

    @Test
    fun `toStringList handles empty array`() {
        val result = Converters.toStringList("[]")
        assertTrue(result.isEmpty())
    }

    @Test
    fun `ExternalMetadata list round-trip preserves data`() {
        val metadata = listOf(
            ExternalMetadata(
                id = 1L,
                mediaId = 42L,
                provider = "tmdb",
                externalId = "tt1375666",
                title = "Inception",
                posterUrl = "http://img.tmdb.org/poster.jpg",
                genres = listOf("Action", "Sci-Fi"),
                lastUpdated = "2024-01-01"
            )
        )

        val json = Converters.fromExternalMetadataList(metadata)
        val restored = Converters.toExternalMetadataList(json)

        assertEquals(1, restored.size)
        assertEquals("Inception", restored[0].title)
        assertEquals("tmdb", restored[0].provider)
        assertEquals("tt1375666", restored[0].externalId)
    }

    @Test
    fun `MediaVersion list round-trip preserves data`() {
        val versions = listOf(
            MediaVersion(
                id = 1L,
                mediaId = 42L,
                version = "1.0",
                quality = "1080p",
                filePath = "/media/movie.mkv",
                fileSize = 4_000_000_000L,
                codec = "h265"
            )
        )

        val json = Converters.fromMediaVersionList(versions)
        val restored = Converters.toMediaVersionList(json)

        assertEquals(1, restored.size)
        assertEquals("1080p", restored[0].quality)
        assertEquals("h265", restored[0].codec)
        assertEquals(4_000_000_000L, restored[0].fileSize)
    }

    @Test
    fun `empty ExternalMetadata list round-trip`() {
        val json = Converters.fromExternalMetadataList(emptyList())
        val restored = Converters.toExternalMetadataList(json)

        assertTrue(restored.isEmpty())
    }

    @Test
    fun `empty MediaVersion list round-trip`() {
        val json = Converters.fromMediaVersionList(emptyList())
        val restored = Converters.toMediaVersionList(json)

        assertTrue(restored.isEmpty())
    }
}
