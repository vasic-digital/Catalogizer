package com.catalogizer.androidtv.utils

import org.junit.Assert.*
import org.junit.Test

class FormatUtilsTest3 {

    // --- formatCodec edge cases ---

    @Test
    fun `formatCodec avc maps to H_264`() {
        assertEquals("H.264", formatCodec("avc"))
    }

    @Test
    fun `formatCodec h265 maps to H_265`() {
        assertEquals("H.265", formatCodec("h265"))
    }

    @Test
    fun `formatCodec vp9`() {
        assertEquals("VP9", formatCodec("vp9"))
    }

    @Test
    fun `formatCodec av1`() {
        assertEquals("AV1", formatCodec("av1"))
    }

    @Test
    fun `formatCodec mp3`() {
        assertEquals("MP3", formatCodec("mp3"))
    }

    @Test
    fun `formatCodec opus`() {
        assertEquals("Opus", formatCodec("opus"))
    }

    @Test
    fun `formatCodec case insensitive`() {
        assertEquals("H.264", formatCodec("H264"))
    }

    // --- formatMediaType edge cases ---

    @Test
    fun `formatMediaType document`() {
        assertEquals("Document", formatMediaType("document"))
    }

    @Test
    fun `formatMediaType video`() {
        assertEquals("Video", formatMediaType("video"))
    }

    @Test
    fun `formatMediaType audio`() {
        assertEquals("Audio", formatMediaType("audio"))
    }

    @Test
    fun `formatMediaType image`() {
        assertEquals("Image", formatMediaType("image"))
    }

    // --- formatRelativeTime ---

    @Test
    fun `formatRelativeTime just now`() {
        val now = System.currentTimeMillis()
        assertEquals("Just now", formatRelativeTime(now))
    }

    @Test
    fun `formatRelativeTime minutes ago`() {
        val fiveMinAgo = System.currentTimeMillis() - (5 * 60 * 1000)
        assertEquals("5 minutes ago", formatRelativeTime(fiveMinAgo))
    }

    @Test
    fun `formatRelativeTime 1 minute ago`() {
        val oneMinAgo = System.currentTimeMillis() - (61 * 1000)
        assertEquals("1 minute ago", formatRelativeTime(oneMinAgo))
    }

    @Test
    fun `formatRelativeTime hours ago`() {
        val threeHoursAgo = System.currentTimeMillis() - (3 * 60 * 60 * 1000)
        assertEquals("3 hours ago", formatRelativeTime(threeHoursAgo))
    }

    @Test
    fun `formatRelativeTime 1 hour ago`() {
        val oneHourAgo = System.currentTimeMillis() - (61 * 60 * 1000)
        assertEquals("1 hour ago", formatRelativeTime(oneHourAgo))
    }

    @Test
    fun `formatRelativeTime days ago`() {
        val twoDaysAgo = System.currentTimeMillis() - (2 * 24 * 60 * 60 * 1000L)
        assertEquals("2 days ago", formatRelativeTime(twoDaysAgo))
    }

    @Test
    fun `formatRelativeTime 1 day ago`() {
        val oneDayAgo = System.currentTimeMillis() - (25 * 60 * 60 * 1000L)
        assertEquals("1 day ago", formatRelativeTime(oneDayAgo))
    }

    @Test
    fun `formatRelativeTime weeks ago`() {
        val twoWeeksAgo = System.currentTimeMillis() - (14 * 24 * 60 * 60 * 1000L)
        assertEquals("2 weeks ago", formatRelativeTime(twoWeeksAgo))
    }

    @Test
    fun `formatRelativeTime 1 year ago`() {
        val oneYearAgo = System.currentTimeMillis() - (365 * 24 * 60 * 60 * 1000L)
        assertEquals("1 year ago", formatRelativeTime(oneYearAgo))
    }

    // --- formatFrameRate ---

    @Test
    fun `formatFrameRate 60 fps integer`() {
        assertEquals("60 fps", formatFrameRate(60.0))
    }

    @Test
    fun `formatFrameRate 29_97 fps fractional`() {
        val result = formatFrameRate(29.97)
        assertNotNull(result)
        assertTrue(result!!.contains("29"))
        assertTrue(result.contains("fps"))
    }

    // --- formatAudioChannels additional ---

    @Test
    fun `formatAudioChannels 3 channels`() {
        assertEquals("3 channels", formatAudioChannels(3))
    }

    // --- formatSampleRate additional ---

    @Test
    fun `formatSampleRate 48kHz`() {
        assertEquals("48 kHz", formatSampleRate(48000))
    }

    @Test
    fun `formatSampleRate 96kHz`() {
        assertEquals("96 kHz", formatSampleRate(96000))
    }

    // --- formatProgress edge cases ---

    @Test
    fun `formatProgress negative total`() {
        assertEquals("0%", formatProgress(50, -10))
    }

    @Test
    fun `formatProgress zero current`() {
        assertEquals("0%", formatProgress(0, 100))
    }

    // --- formatBitrate edge cases ---

    @Test
    fun `formatBitrate exactly 1 Mbps`() {
        assertEquals("1.0 Mbps", formatBitrate(1_000_000))
    }

    @Test
    fun `formatBitrate exactly 1 Kbps`() {
        assertEquals("1 Kbps", formatBitrate(1_000))
    }

    @Test
    fun `formatBitrate zero`() {
        assertEquals("0 bps", formatBitrate(0))
    }
}
