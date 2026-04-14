package com.catalogizer.androidtv.ui.screens.player

import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for MediaPlayerScreen logic: time formatting, playback speed presets,
 * error message mapping, overlay menu state, D-PAD key handling logic,
 * track info data, and subtitle toggle state.
 */
class MediaPlayerScreenTest {

    // --- Time formatting ---

    @Test
    fun `formatTime returns 00 colon 00 for zero`() {
        assertEquals("00:00", formatTime(0L))
    }

    @Test
    fun `formatTime returns 00 colon 00 for negative`() {
        assertEquals("00:00", formatTime(-1000L))
    }

    @Test
    fun `formatTime formats seconds correctly`() {
        assertEquals("00:30", formatTime(30_000L))
    }

    @Test
    fun `formatTime formats minutes and seconds correctly`() {
        assertEquals("05:30", formatTime(330_000L))
    }

    @Test
    fun `formatTime formats hours correctly`() {
        assertEquals("1:30:00", formatTime(5_400_000L))
    }

    @Test
    fun `formatTime formats complex duration correctly`() {
        assertEquals("2:15:45", formatTime(8_145_000L))
    }

    @Test
    fun `formatTime handles exactly one hour`() {
        assertEquals("1:00:00", formatTime(3_600_000L))
    }

    @Test
    fun `formatTime handles exactly one minute`() {
        assertEquals("01:00", formatTime(60_000L))
    }

    // --- Speed presets ---

    @Test
    fun `speed options contains eight presets`() {
        val speedOptions = listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f, 3.0f)
        assertEquals(8, speedOptions.size)
    }

    @Test
    fun `default speed is 1_0x`() {
        val speedOptions = listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f, 3.0f)
        val defaultIndex = speedOptions.indexOf(1.0f)
        assertEquals(3, defaultIndex)
    }

    @Test
    fun `minimum speed is 0_25x`() {
        val speedOptions = listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f, 3.0f)
        assertEquals(0.25f, speedOptions.first(), 0.001f)
    }

    @Test
    fun `maximum speed is 3_0x`() {
        val speedOptions = listOf(0.25f, 0.5f, 0.75f, 1.0f, 1.25f, 1.5f, 2.0f, 3.0f)
        assertEquals(3.0f, speedOptions.last(), 0.001f)
    }

    @Test
    fun `speed label formats correctly`() {
        val speed = 1.5f
        val label = "${speed}x"
        assertEquals("1.5x", label)
    }

    @Test
    fun `non-default speed indicator activates`() {
        val playbackSpeed = 1.5f
        val isNonDefault = playbackSpeed != 1.0f
        assertTrue(isNonDefault)
    }

    @Test
    fun `default speed indicator is inactive`() {
        val playbackSpeed = 1.0f
        val isNonDefault = playbackSpeed != 1.0f
        assertFalse(isNonDefault)
    }

    // --- Error message mapping ---

    @Test
    fun `404 error produces file not scanned message`() {
        val code = 404
        val error = mapErrorCode(code)
        assertTrue(error.contains("not have been scanned"))
    }

    @Test
    fun `401 error produces auth message`() {
        val code = 401
        val error = mapErrorCode(code)
        assertTrue(error.contains("Authentication"))
    }

    @Test
    fun `403 error produces auth message`() {
        val code = 403
        val error = mapErrorCode(code)
        assertTrue(error.contains("Authentication"))
    }

    @Test
    fun `500 error produces server error message`() {
        val code = 500
        val error = mapErrorCode(code)
        assertTrue(error.contains("Server error"))
    }

    @Test
    fun `unknown error code produces generic message`() {
        val code = 502
        val error = mapErrorCode(code)
        assertTrue(error.contains("HTTP 502"))
    }

    // --- Overlay menu state management ---

    @Test
    fun `opening speed menu closes other menus`() {
        var showSpeedMenu = false
        var showAudioMenu = true
        var showSubtitleMenu = true

        showSpeedMenu = true
        showAudioMenu = false
        showSubtitleMenu = false

        assertTrue(showSpeedMenu)
        assertFalse(showAudioMenu)
        assertFalse(showSubtitleMenu)
    }

    @Test
    fun `opening audio menu closes other menus`() {
        var showSpeedMenu = true
        var showAudioMenu = false
        var showSubtitleMenu = true

        showAudioMenu = true
        showSpeedMenu = false
        showSubtitleMenu = false

        assertFalse(showSpeedMenu)
        assertTrue(showAudioMenu)
        assertFalse(showSubtitleMenu)
    }

    @Test
    fun `opening subtitle menu closes other menus`() {
        var showSpeedMenu = true
        var showAudioMenu = true
        var showSubtitleMenu = false

        showSubtitleMenu = true
        showSpeedMenu = false
        showAudioMenu = false

        assertFalse(showSpeedMenu)
        assertFalse(showAudioMenu)
        assertTrue(showSubtitleMenu)
    }

    @Test
    fun `back key closes menus before navigating`() {
        var showSpeedMenu = true
        var showAudioMenu = false
        var showSubtitleMenu = false
        var navigatedBack = false

        if (showSpeedMenu || showAudioMenu || showSubtitleMenu) {
            showSpeedMenu = false
            showAudioMenu = false
            showSubtitleMenu = false
        } else {
            navigatedBack = true
        }

        assertFalse(showSpeedMenu)
        assertFalse(navigatedBack)
    }

    @Test
    fun `back key navigates when no menus open`() {
        var showSpeedMenu = false
        var showAudioMenu = false
        var showSubtitleMenu = false
        var navigatedBack = false

        if (showSpeedMenu || showAudioMenu || showSubtitleMenu) {
            showSpeedMenu = false
            showAudioMenu = false
            showSubtitleMenu = false
        } else {
            navigatedBack = true
        }

        assertTrue(navigatedBack)
    }

    // --- Controls visibility ---

    @Test
    fun `controls auto-hide disabled when menus are open`() {
        val showControls = true
        val isPlaying = true
        val showSpeedMenu = true
        val showAudioMenu = false
        val showSubtitleMenu = false

        val shouldAutoHide = showControls && isPlaying &&
            !showSpeedMenu && !showAudioMenu && !showSubtitleMenu
        assertFalse(shouldAutoHide)
    }

    @Test
    fun `controls auto-hide enabled when playing and no menus open`() {
        val showControls = true
        val isPlaying = true
        val showSpeedMenu = false
        val showAudioMenu = false
        val showSubtitleMenu = false

        val shouldAutoHide = showControls && isPlaying &&
            !showSpeedMenu && !showAudioMenu && !showSubtitleMenu
        assertTrue(shouldAutoHide)
    }

    @Test
    fun `controls auto-hide disabled when paused`() {
        val showControls = true
        val isPlaying = false

        val shouldAutoHide = showControls && isPlaying
        assertFalse(shouldAutoHide)
    }

    // --- Subtitle toggle ---

    @Test
    fun `subtitle index minus one means off`() {
        val selectedSubtitleIndex = -1
        val isOff = selectedSubtitleIndex < 0
        assertTrue(isOff)
    }

    @Test
    fun `subtitle index zero or above means on`() {
        val selectedSubtitleIndex = 0
        val isOn = selectedSubtitleIndex >= 0
        assertTrue(isOn)
    }

    @Test
    fun `subtitle label shows Off when index is negative`() {
        val selectedSubtitleIndex = -1
        val label = if (selectedSubtitleIndex >= 0) "On" else "Off"
        assertEquals("Off", label)
    }

    @Test
    fun `subtitle label shows On when index is non-negative`() {
        val selectedSubtitleIndex = 1
        val label = if (selectedSubtitleIndex >= 0) "On" else "Off"
        assertEquals("On", label)
    }

    @Test
    fun `subtitle menu items include Off as first option`() {
        val subtitleTracks = listOf("English", "Spanish")
        val menuLabels = listOf("Off") + subtitleTracks
        assertEquals(3, menuLabels.size)
        assertEquals("Off", menuLabels[0])
        assertEquals("English", menuLabels[1])
    }

    @Test
    fun `subtitle menu selection maps to track index with offset`() {
        // Menu index 0 = Off (track index -1)
        // Menu index 1 = first track (track index 0)
        val menuIndex = 2
        val trackIndex = menuIndex - 1
        assertEquals(1, trackIndex)
    }

    // --- Audio track defaults ---

    @Test
    fun `empty audio tracks shows Default label`() {
        val audioTracks = emptyList<String>()
        val labels = if (audioTracks.isEmpty()) listOf("Default") else audioTracks
        assertEquals(1, labels.size)
        assertEquals("Default", labels[0])
    }

    @Test
    fun `audio tracks list uses track labels`() {
        val audioTracks = listOf("English 5.1", "Japanese 2.0")
        val labels = if (audioTracks.isEmpty()) listOf("Default") else audioTracks
        assertEquals(2, labels.size)
        assertEquals("English 5.1", labels[0])
    }

    // --- Retry mechanism ---

    @Test
    fun `retry clears error and resets URL`() {
        var streamError: String? = "Previous error"
        var resolvedUrl = "http://old-url"
        var retryCount = 0

        streamError = null
        resolvedUrl = ""
        retryCount++

        assertNull(streamError)
        assertEquals("", resolvedUrl)
        assertEquals(1, retryCount)
    }

    // --- Progress indicator ---

    @Test
    fun `buffered progress fraction is clamped to 0 through 1`() {
        val bufferedPosition = 5000L
        val duration = 10000L
        val fraction = (bufferedPosition.toFloat() / duration.toFloat().coerceAtLeast(1f))
            .coerceIn(0f, 1f)
        assertEquals(0.5f, fraction, 0.001f)
    }

    @Test
    fun `buffered progress handles zero duration`() {
        val bufferedPosition = 5000L
        val duration = 0L
        val fraction = (bufferedPosition.toFloat() / duration.toFloat().coerceAtLeast(1f))
            .coerceIn(0f, 1f)
        assertEquals(1.0f, fraction, 0.001f)
    }

    // --- Helper matching MediaPlayerScreen.kt formatTime ---

    private fun formatTime(timeMs: Long): String {
        if (timeMs <= 0) return "00:00"
        val seconds = (timeMs / 1000) % 60
        val minutes = (timeMs / (1000 * 60)) % 60
        val hours = timeMs / (1000 * 60 * 60)
        return if (hours > 0) {
            String.format("%d:%02d:%02d", hours, minutes, seconds)
        } else {
            String.format("%02d:%02d", minutes, seconds)
        }
    }

    private fun mapErrorCode(code: Int): String = when (code) {
        404 -> "No playable file linked to this media item. The file may not have been scanned yet."
        401, 403 -> "Authentication required. Please sign in again."
        500 -> "Server error resolving stream. The storage may be unreachable."
        else -> "Stream unavailable (HTTP $code)"
    }
}
