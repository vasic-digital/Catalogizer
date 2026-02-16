package com.catalogizer.androidtv.utils

import org.junit.Assert.*
import org.junit.Test

class FormatUtilsTest {

    // --- formatDuration Tests ---

    @Test
    fun `formatDuration with zero seconds should return 0 colon 00`() {
        assertEquals("0:00", formatDuration(0L))
    }

    @Test
    fun `formatDuration with seconds only should format correctly`() {
        assertEquals("0:30", formatDuration(30L))
    }

    @Test
    fun `formatDuration with minutes and seconds should format correctly`() {
        assertEquals("5:30", formatDuration(330L))
    }

    @Test
    fun `formatDuration with hours should format correctly`() {
        assertEquals("1:30:00", formatDuration(5400L))
    }

    @Test
    fun `formatDuration with complex time should format correctly`() {
        assertEquals("2:15:45", formatDuration(8145L))
    }

    // --- formatFileSize Tests ---

    @Test
    fun `formatFileSize with zero bytes should return 0 B`() {
        assertEquals("0 B", formatFileSize(0L))
    }

    @Test
    fun `formatFileSize with negative bytes should return 0 B`() {
        assertEquals("0 B", formatFileSize(-100L))
    }

    @Test
    fun `formatFileSize with bytes should format correctly`() {
        assertEquals("500.0 B", formatFileSize(500L))
    }

    @Test
    fun `formatFileSize with kilobytes should format correctly`() {
        assertEquals("1.0 KB", formatFileSize(1024L))
    }

    @Test
    fun `formatFileSize with megabytes should format correctly`() {
        assertEquals("1.0 MB", formatFileSize(1048576L))
    }

    @Test
    fun `formatFileSize with gigabytes should format correctly`() {
        assertEquals("1.0 GB", formatFileSize(1073741824L))
    }

    @Test
    fun `formatFileSize with terabytes should format correctly`() {
        assertEquals("1.0 TB", formatFileSize(1099511627776L))
    }

    // --- formatBitrate Tests ---

    @Test
    fun `formatBitrate with bps should format correctly`() {
        assertEquals("500 bps", formatBitrate(500L))
    }

    @Test
    fun `formatBitrate with Kbps should format correctly`() {
        assertEquals("320 Kbps", formatBitrate(320_000L))
    }

    @Test
    fun `formatBitrate with Mbps should format correctly`() {
        assertEquals("5.0 Mbps", formatBitrate(5_000_000L))
    }

    // --- formatResolution Tests ---

    @Test
    fun `formatResolution with valid dimensions should format correctly`() {
        assertEquals("1920x1080", formatResolution(1920, 1080))
    }

    @Test
    fun `formatResolution with null width should return null`() {
        assertNull(formatResolution(null, 1080))
    }

    @Test
    fun `formatResolution with null height should return null`() {
        assertNull(formatResolution(1920, null))
    }

    @Test
    fun `formatResolution with both null should return null`() {
        assertNull(formatResolution(null, null))
    }

    // --- getQualityFromResolution Tests ---

    @Test
    fun `getQualityFromResolution with 4K should return 4K UHD`() {
        assertEquals("4K UHD", getQualityFromResolution(3840, 2160))
    }

    @Test
    fun `getQualityFromResolution with 1440p should return 2K QHD`() {
        assertEquals("2K QHD", getQualityFromResolution(2560, 1440))
    }

    @Test
    fun `getQualityFromResolution with 1080p should return Full HD`() {
        assertEquals("Full HD", getQualityFromResolution(1920, 1080))
    }

    @Test
    fun `getQualityFromResolution with 720p should return HD`() {
        assertEquals("HD", getQualityFromResolution(1280, 720))
    }

    @Test
    fun `getQualityFromResolution with 480p should return SD`() {
        assertEquals("SD", getQualityFromResolution(640, 480))
    }

    @Test
    fun `getQualityFromResolution with low resolution should return Low Quality`() {
        assertEquals("Low Quality", getQualityFromResolution(320, 240))
    }

    @Test
    fun `getQualityFromResolution with null should return Unknown`() {
        assertEquals("Unknown", getQualityFromResolution(null, null))
    }

    // --- formatMediaType Tests ---

    @Test
    fun `formatMediaType movie should return Movie`() {
        assertEquals("Movie", formatMediaType("movie"))
    }

    @Test
    fun `formatMediaType tv should return TV Show`() {
        assertEquals("TV Show", formatMediaType("tv"))
    }

    @Test
    fun `formatMediaType music should return Music`() {
        assertEquals("Music", formatMediaType("music"))
    }

    @Test
    fun `formatMediaType document should return Document`() {
        assertEquals("Document", formatMediaType("document"))
    }

    @Test
    fun `formatMediaType video should return Video`() {
        assertEquals("Video", formatMediaType("video"))
    }

    @Test
    fun `formatMediaType audio should return Audio`() {
        assertEquals("Audio", formatMediaType("audio"))
    }

    @Test
    fun `formatMediaType image should return Image`() {
        assertEquals("Image", formatMediaType("image"))
    }

    @Test
    fun `formatMediaType should be case insensitive`() {
        assertEquals("Movie", formatMediaType("MOVIE"))
        assertEquals("TV Show", formatMediaType("TV"))
    }

    // --- formatCodec Tests ---

    @Test
    fun `formatCodec h264 should return H dot 264`() {
        assertEquals("H.264", formatCodec("h264"))
    }

    @Test
    fun `formatCodec avc should return H dot 264`() {
        assertEquals("H.264", formatCodec("avc"))
    }

    @Test
    fun `formatCodec h265 should return H dot 265`() {
        assertEquals("H.265", formatCodec("h265"))
    }

    @Test
    fun `formatCodec hevc should return H dot 265`() {
        assertEquals("H.265", formatCodec("hevc"))
    }

    @Test
    fun `formatCodec vp9 should return VP9`() {
        assertEquals("VP9", formatCodec("vp9"))
    }

    @Test
    fun `formatCodec av1 should return AV1`() {
        assertEquals("AV1", formatCodec("av1"))
    }

    @Test
    fun `formatCodec unknown should return uppercase`() {
        assertEquals("UNKNOWN", formatCodec("unknown"))
    }

    // --- formatFrameRate Tests ---

    @Test
    fun `formatFrameRate with integer rate should format without decimal`() {
        assertEquals("30 fps", formatFrameRate(30.0))
    }

    @Test
    fun `formatFrameRate with fractional rate should format with decimal`() {
        assertEquals("24.0 fps", formatFrameRate(23.976))
    }

    @Test
    fun `formatFrameRate with null should return null`() {
        assertNull(formatFrameRate(null))
    }

    // --- formatAudioChannels Tests ---

    @Test
    fun `formatAudioChannels 1 should return Mono`() {
        assertEquals("Mono", formatAudioChannels(1))
    }

    @Test
    fun `formatAudioChannels 2 should return Stereo`() {
        assertEquals("Stereo", formatAudioChannels(2))
    }

    @Test
    fun `formatAudioChannels 6 should return 5 dot 1`() {
        assertEquals("5.1", formatAudioChannels(6))
    }

    @Test
    fun `formatAudioChannels 8 should return 7 dot 1`() {
        assertEquals("7.1", formatAudioChannels(8))
    }

    @Test
    fun `formatAudioChannels with other value should return N channels`() {
        assertEquals("4 channels", formatAudioChannels(4))
    }

    @Test
    fun `formatAudioChannels null should return null`() {
        assertNull(formatAudioChannels(null))
    }

    // --- formatSampleRate Tests ---

    @Test
    fun `formatSampleRate 44100 should return 44 kHz`() {
        assertEquals("44 kHz", formatSampleRate(44100))
    }

    @Test
    fun `formatSampleRate 48000 should return 48 kHz`() {
        assertEquals("48 kHz", formatSampleRate(48000))
    }

    @Test
    fun `formatSampleRate below 1000 should return Hz`() {
        assertEquals("500 Hz", formatSampleRate(500))
    }

    @Test
    fun `formatSampleRate null should return null`() {
        assertNull(formatSampleRate(null))
    }

    // --- formatProgress Tests ---

    @Test
    fun `formatProgress should calculate percentage correctly`() {
        assertEquals("50%", formatProgress(50, 100))
    }

    @Test
    fun `formatProgress with zero total should return 0 percent`() {
        assertEquals("0%", formatProgress(50, 0))
    }

    @Test
    fun `formatProgress with negative total should return 0 percent`() {
        assertEquals("0%", formatProgress(50, -1))
    }

    @Test
    fun `formatProgress complete should return 100 percent`() {
        assertEquals("100%", formatProgress(100, 100))
    }

    // --- formatProgressTime Tests ---

    @Test
    fun `formatProgressTime should format both current and total`() {
        assertEquals("1:30 / 5:00", formatProgressTime(90, 300))
    }

    @Test
    fun `formatProgressTime with hours should format correctly`() {
        assertEquals("1:00:00 / 2:00:00", formatProgressTime(3600, 7200))
    }
}
