package com.catalogizer.androidtv.ui

import android.net.Uri
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class DeepLinkParserEdgeCaseTest {

    @Test
    fun `parse with null URI returns null mediaId`() {
        val result = DeepLinkParser.parse(null)
        assertNull(result.mediaId)
        assertNull(result.mediaType)
    }

    @Test
    fun `parse with different scheme returns values`() {
        // DeepLinkParser.parse() doesn't validate scheme — it just reads pathSegments and query params
        val uri = Uri.parse("https://catalogizer.app/media/99?type=tv_show")
        val result = DeepLinkParser.parse(uri)
        // Path for https URI: pathSegments would be ["media", "99"], firstOrNull = "media" which is non-numeric -> null
        // Actually for https://catalogizer.app/media/99 the path is /media/99 and segments are ["media", "99"]
        // firstOrNull() = "media" -> toLongOrNull() = null
        assertNull(result.mediaId)
    }

    @Test
    fun `parse with very large mediaId`() {
        val uri = Uri.parse("catalogizer://media/${Long.MAX_VALUE}?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals(Long.MAX_VALUE, result.mediaId)
        assertEquals("movie", result.mediaType)
    }

    @Test
    fun `parse with negative mediaId`() {
        val uri = Uri.parse("catalogizer://media/-42?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals(-42L, result.mediaId)
        assertEquals("movie", result.mediaType)
    }

    @Test
    fun `parse with multiple path segments uses first`() {
        // catalogizer://media/123/extra -> pathSegments = ["123", "extra"]
        // firstOrNull() = "123" -> 123L
        val uri = Uri.parse("catalogizer://media/123/extra?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals(123L, result.mediaId)
        assertEquals("movie", result.mediaType)
    }

    @Test
    fun `isAudioWithoutContext with null mediaType returns false`() {
        assertFalse(DeepLinkParser.isAudioWithoutContext(null, hasExternalMetadata = false))
    }

    @Test
    fun `isAudioWithoutContext with empty string returns false`() {
        assertFalse(DeepLinkParser.isAudioWithoutContext("", hasExternalMetadata = false))
    }

    @Test
    fun `parse extracts type correctly with extra query params`() {
        val uri = Uri.parse("catalogizer://media/77?foo=bar&type=game&extra=ignored")
        val result = DeepLinkParser.parse(uri)
        assertEquals(77L, result.mediaId)
        assertEquals("game", result.mediaType)
    }
}
