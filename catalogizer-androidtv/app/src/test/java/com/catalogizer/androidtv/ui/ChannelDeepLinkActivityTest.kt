package com.catalogizer.androidtv.ui

import android.net.Uri
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class ChannelDeepLinkActivityTest {

    @Test
    fun `parseDeepLink extracts mediaId from URI`() {
        val uri = Uri.parse("catalogizer://media/42?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals(42L, result.mediaId)
    }

    @Test
    fun `parseDeepLink extracts mediaType from URI`() {
        val uri = Uri.parse("catalogizer://media/42?type=movie")
        val result = DeepLinkParser.parse(uri)
        assertEquals("movie", result.mediaType)
    }

    @Test
    fun `parseDeepLink handles missing type parameter`() {
        val uri = Uri.parse("catalogizer://media/42")
        val result = DeepLinkParser.parse(uri)
        assertEquals(42L, result.mediaId)
        assertNull(result.mediaType)
    }

    @Test
    fun `parseDeepLink returns null mediaId for invalid path`() {
        val uri = Uri.parse("catalogizer://media/invalid")
        val result = DeepLinkParser.parse(uri)
        assertNull(result.mediaId)
    }

    @Test
    fun `parseDeepLink returns null mediaId for empty path`() {
        val uri = Uri.parse("catalogizer://media")
        val result = DeepLinkParser.parse(uri)
        assertNull(result.mediaId)
    }

    @Test
    fun `isAudioWithoutContext returns true for music without metadata`() {
        assertTrue(DeepLinkParser.isAudioWithoutContext("music", hasExternalMetadata = false))
    }

    @Test
    fun `isAudioWithoutContext returns false for music with metadata`() {
        assertFalse(DeepLinkParser.isAudioWithoutContext("music", hasExternalMetadata = true))
    }

    @Test
    fun `isAudioWithoutContext returns false for non-audio types`() {
        assertFalse(DeepLinkParser.isAudioWithoutContext("movie", hasExternalMetadata = false))
    }

    @Test
    fun `isAudioWithoutContext returns true for audiobook and podcast without metadata`() {
        assertTrue(DeepLinkParser.isAudioWithoutContext("audiobook", hasExternalMetadata = false))
        assertTrue(DeepLinkParser.isAudioWithoutContext("podcast", hasExternalMetadata = false))
    }
}
