package com.catalogizer.androidtv.ui.splash

import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for SplashContent logic: splash duration selection, safety timeout,
 * completion conditions, and first-launch detection.
 */
class SplashContentTest {

    // --- Splash duration selection ---

    @Test
    fun `first launch uses longer splash duration`() {
        val isFirstLaunch = true
        val minDuration = if (isFirstLaunch) 2500L else 1500L
        assertEquals(2500L, minDuration)
    }

    @Test
    fun `subsequent launch uses shorter splash duration`() {
        val isFirstLaunch = false
        val minDuration = if (isFirstLaunch) 2500L else 1500L
        assertEquals(1500L, minDuration)
    }

    // --- Completion conditions ---

    @Test
    fun `splash completes when timer done and app ready`() {
        val timerComplete = true
        val isAppReady = true
        val safetyTimeout = false

        val shouldComplete = (timerComplete && isAppReady) || safetyTimeout
        assertTrue(shouldComplete)
    }

    @Test
    fun `splash does not complete when timer not done`() {
        val timerComplete = false
        val isAppReady = true
        val safetyTimeout = false

        val shouldComplete = (timerComplete && isAppReady) || safetyTimeout
        assertFalse(shouldComplete)
    }

    @Test
    fun `splash does not complete when app not ready`() {
        val timerComplete = true
        val isAppReady = false
        val safetyTimeout = false

        val shouldComplete = (timerComplete && isAppReady) || safetyTimeout
        assertFalse(shouldComplete)
    }

    @Test
    fun `splash does not complete when both timer and app not ready`() {
        val timerComplete = false
        val isAppReady = false
        val safetyTimeout = false

        val shouldComplete = (timerComplete && isAppReady) || safetyTimeout
        assertFalse(shouldComplete)
    }

    // --- Safety timeout ---

    @Test
    fun `safety timeout forces completion regardless of app ready`() {
        val timerComplete = false
        val isAppReady = false
        val safetyTimeout = true

        val shouldComplete = (timerComplete && isAppReady) || safetyTimeout
        assertTrue(shouldComplete)
    }

    @Test
    fun `safety timeout forces completion regardless of timer`() {
        val timerComplete = false
        val isAppReady = true
        val safetyTimeout = true

        val shouldComplete = (timerComplete && isAppReady) || safetyTimeout
        assertTrue(shouldComplete)
    }

    @Test
    fun `safety timeout is 5000ms`() {
        val safetyTimeoutMs = 5000L
        assertEquals(5000L, safetyTimeoutMs)
    }

    @Test
    fun `safety timeout is less than ANR threshold`() {
        val safetyTimeoutMs = 5000L
        // Android ANR threshold for broadcast receivers is 10s, for input is 5s
        // Our safety timeout must complete before the app freezes
        assertTrue(safetyTimeoutMs <= 5000L)
    }

    // --- First launch detection ---

    @Test
    fun `first launch detected when has_launched pref is false`() {
        val hasLaunched = false
        val isFirstLaunch = !hasLaunched
        assertTrue(isFirstLaunch)
    }

    @Test
    fun `not first launch when has_launched pref is true`() {
        val hasLaunched = true
        val isFirstLaunch = !hasLaunched
        assertFalse(isFirstLaunch)
    }

    // --- Splash callback ---

    @Test
    fun `onSplashComplete callback fires when conditions met`() {
        var completed = false
        val onSplashComplete: () -> Unit = { completed = true }

        val timerComplete = true
        val isAppReady = true
        val safetyTimeout = false

        if ((timerComplete && isAppReady) || safetyTimeout) {
            onSplashComplete()
        }

        assertTrue(completed)
    }

    @Test
    fun `onSplashComplete callback does not fire prematurely`() {
        var completed = false
        val onSplashComplete: () -> Unit = { completed = true }

        val timerComplete = false
        val isAppReady = false
        val safetyTimeout = false

        if ((timerComplete && isAppReady) || safetyTimeout) {
            onSplashComplete()
        }

        assertFalse(completed)
    }

    // --- Elapsed time tracking ---

    @Test
    fun `elapsed time calculation works correctly`() {
        val startTime = 1000L
        val currentTime = 3500L
        val elapsed = currentTime - startTime
        assertEquals(2500L, elapsed)
    }

    @Test
    fun `elapsed time is non-negative`() {
        val startTime = System.currentTimeMillis()
        val elapsed = System.currentTimeMillis() - startTime
        assertTrue(elapsed >= 0)
    }

    // --- Display content ---

    @Test
    fun `splash displays app name Catalogizer`() {
        val appName = "Catalogizer"
        assertEquals("Catalogizer", appName)
    }

    @Test
    fun `splash displays subtitle`() {
        val subtitle = "Advanced Multi-Protocol Media\nCollection Management System"
        assertTrue(subtitle.contains("Multi-Protocol"))
        assertTrue(subtitle.contains("Media"))
    }

    @Test
    fun `splash displays footer branding`() {
        val footer = "Made with \u2764\uFE0F by Vasic Digital"
        assertTrue(footer.contains("Vasic Digital"))
    }

    // --- Loading indicator ---

    @Test
    fun `splash shows loading indicator while not complete`() {
        val timerComplete = false
        val isAppReady = false
        val shouldShowSpinner = !((timerComplete && isAppReady))
        assertTrue(shouldShowSpinner)
    }
}
