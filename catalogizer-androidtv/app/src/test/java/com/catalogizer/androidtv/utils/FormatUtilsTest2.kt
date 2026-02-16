package com.catalogizer.androidtv.utils

import org.junit.Assert.*
import org.junit.Test

class FormatUtilsTest2 {

    // --- formatDuration ---

    @Test
    fun `formatDuration formats hours minutes seconds`() {
        assertEquals("1:30:45", formatDuration(5445))
    }

    @Test
    fun `formatDuration formats minutes and seconds only`() {
        assertEquals("5:30", formatDuration(330))
    }

    @Test
    fun `formatDuration formats zero duration`() {
        assertEquals("0:00", formatDuration(0))
    }

    @Test
    fun `formatDuration formats single digit seconds with padding`() {
        assertEquals("1:05", formatDuration(65))
    }

    @Test
    fun `formatDuration formats exactly one hour`() {
        assertEquals("1:00:00", formatDuration(3600))
    }

    // --- formatFileSize ---

    @Test
    fun `formatFileSize formats bytes`() {
        assertEquals("0 B", formatFileSize(0))
    }

    @Test
    fun `formatFileSize formats kilobytes`() {
        val result = formatFileSize(1024)
        assertTrue(result.contains("KB"))
    }

    @Test
    fun `formatFileSize formats megabytes`() {
        val result = formatFileSize(1_048_576)
        assertTrue(result.contains("MB"))
    }

    @Test
    fun `formatFileSize formats gigabytes`() {
        val result = formatFileSize(1_073_741_824)
        assertTrue(result.contains("GB"))
    }

    @Test
    fun `formatFileSize formats terabytes`() {
        val result = formatFileSize(1_099_511_627_776)
        assertTrue(result.contains("TB"))
    }

    @Test
    fun `formatFileSize handles negative values`() {
        assertEquals("0 B", formatFileSize(-1))
    }

    // --- formatBitrate ---

    @Test
    fun `formatBitrate formats megabits`() {
        val result = formatBitrate(5_000_000)
        assertTrue(result.contains("Mbps"))
    }

    @Test
    fun `formatBitrate formats kilobits`() {
        val result = formatBitrate(128_000)
        assertTrue(result.contains("Kbps"))
    }

    @Test
    fun `formatBitrate formats bits`() {
        val result = formatBitrate(500)
        assertTrue(result.contains("bps"))
        assertFalse(result.contains("Kbps"))
    }

    // --- formatResolution ---

    @Test
    fun `formatResolution returns formatted string`() {
        assertEquals("1920x1080", formatResolution(1920, 1080))
    }

    @Test
    fun `formatResolution returns null when width is null`() {
        assertNull(formatResolution(null, 1080))
    }

    @Test
    fun `formatResolution returns null when height is null`() {
        assertNull(formatResolution(1920, null))
    }

    @Test
    fun `formatResolution returns null when both are null`() {
        assertNull(formatResolution(null, null))
    }

    // --- getQualityFromResolution ---

    @Test
    fun `getQualityFromResolution returns 4K UHD for 2160p`() {
        assertEquals("4K UHD", getQualityFromResolution(3840, 2160))
    }

    @Test
    fun `getQualityFromResolution returns 2K QHD for 1440p`() {
        assertEquals("2K QHD", getQualityFromResolution(2560, 1440))
    }

    @Test
    fun `getQualityFromResolution returns Full HD for 1080p`() {
        assertEquals("Full HD", getQualityFromResolution(1920, 1080))
    }

    @Test
    fun `getQualityFromResolution returns HD for 720p`() {
        assertEquals("HD", getQualityFromResolution(1280, 720))
    }

    @Test
    fun `getQualityFromResolution returns SD for 480p`() {
        assertEquals("SD", getQualityFromResolution(854, 480))
    }

    @Test
    fun `getQualityFromResolution returns Low Quality for below 480p`() {
        assertEquals("Low Quality", getQualityFromResolution(320, 240))
    }

    @Test
    fun `getQualityFromResolution returns Unknown for null values`() {
        assertEquals("Unknown", getQualityFromResolution(null, null))
    }

    // --- formatMediaType ---

    @Test
    fun `formatMediaType formats known types`() {
        assertEquals("Movie", formatMediaType("movie"))
        assertEquals("TV Show", formatMediaType("tv"))
        assertEquals("Music", formatMediaType("music"))
        assertEquals("Document", formatMediaType("document"))
        assertEquals("Video", formatMediaType("video"))
        assertEquals("Audio", formatMediaType("audio"))
        assertEquals("Image", formatMediaType("image"))
    }

    @Test
    fun `formatMediaType capitalizes unknown types`() {
        val result = formatMediaType("podcast")
        assertEquals("Podcast", result)
    }

    // --- formatCodec ---

    @Test
    fun `formatCodec formats known codecs`() {
        assertEquals("H.264", formatCodec("h264"))
        assertEquals("H.264", formatCodec("avc"))
        assertEquals("H.265", formatCodec("h265"))
        assertEquals("H.265", formatCodec("hevc"))
        assertEquals("VP9", formatCodec("vp9"))
        assertEquals("AV1", formatCodec("av1"))
        assertEquals("MP3", formatCodec("mp3"))
        assertEquals("AAC", formatCodec("aac"))
        assertEquals("FLAC", formatCodec("flac"))
        assertEquals("Opus", formatCodec("opus"))
    }

    @Test
    fun `formatCodec uppercases unknown codecs`() {
        assertEquals("WEBM", formatCodec("webm"))
    }

    // --- formatFrameRate ---

    @Test
    fun `formatFrameRate formats integer frame rate`() {
        assertEquals("30 fps", formatFrameRate(30.0))
        assertEquals("60 fps", formatFrameRate(60.0))
    }

    @Test
    fun `formatFrameRate formats decimal frame rate`() {
        assertEquals("24.0 fps".substring(0, 4), formatFrameRate(23.976)?.substring(0, 4))
    }

    @Test
    fun `formatFrameRate returns null for null input`() {
        assertNull(formatFrameRate(null))
    }

    // --- formatAudioChannels ---

    @Test
    fun `formatAudioChannels formats known channel counts`() {
        assertEquals("Mono", formatAudioChannels(1))
        assertEquals("Stereo", formatAudioChannels(2))
        assertEquals("5.1", formatAudioChannels(6))
        assertEquals("7.1", formatAudioChannels(8))
    }

    @Test
    fun `formatAudioChannels formats unknown channel count`() {
        assertEquals("4 channels", formatAudioChannels(4))
    }

    @Test
    fun `formatAudioChannels returns null for null input`() {
        assertNull(formatAudioChannels(null))
    }

    // --- formatSampleRate ---

    @Test
    fun `formatSampleRate formats kHz`() {
        assertEquals("44 kHz", formatSampleRate(44100))
        assertEquals("48 kHz", formatSampleRate(48000))
        assertEquals("96 kHz", formatSampleRate(96000))
    }

    @Test
    fun `formatSampleRate formats Hz for low values`() {
        assertEquals("500 Hz", formatSampleRate(500))
    }

    @Test
    fun `formatSampleRate returns null for null input`() {
        assertNull(formatSampleRate(null))
    }

    // --- formatProgress ---

    @Test
    fun `formatProgress calculates percentage`() {
        assertEquals("50%", formatProgress(50, 100))
        assertEquals("100%", formatProgress(100, 100))
        assertEquals("0%", formatProgress(0, 100))
    }

    @Test
    fun `formatProgress handles zero total`() {
        assertEquals("0%", formatProgress(50, 0))
    }

    @Test
    fun `formatProgress handles negative total`() {
        assertEquals("0%", formatProgress(50, -1))
    }

    // --- formatProgressTime ---

    @Test
    fun `formatProgressTime formats current and total`() {
        val result = formatProgressTime(300, 600)
        assertTrue(result.contains("/"))
        assertTrue(result.contains("5:00"))
        assertTrue(result.contains("10:00"))
    }
}
