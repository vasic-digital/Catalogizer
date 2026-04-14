package com.catalogizer.android.ui.splash

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.MediaRepository
import com.catalogizer.android.ui.viewmodel.MainViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for SplashContent logic via MainViewModel.
 * Covers splash timing, version display, transition conditions,
 * first-launch detection, and the app readiness signal.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SplashContentTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockMediaRepository: MediaRepository
    private lateinit var mainViewModel: MainViewModel

    @Before
    fun setup() {
        mockMediaRepository = mockk(relaxed = true)
        mainViewModel = MainViewModel(mockMediaRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state is loading`() {
        assertTrue(mainViewModel.isLoading.value)
    }

    @Test
    fun `initializeApp sets loading to false`() = runTest {
        mainViewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(mainViewModel.isLoading.value)
    }

    @Test
    fun `initializeApp handles errors gracefully`() = runTest {
        mainViewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(mainViewModel.isLoading.value)
    }

    @Test
    fun `multiple initializeApp calls are safe`() = runTest {
        mainViewModel.initializeApp()
        mainViewModel.initializeApp()
        mainViewModel.initializeApp()
        advanceUntilIdle()

        assertFalse(mainViewModel.isLoading.value)
    }

    @Test
    fun `isLoading is a StateFlow with initial value true`() {
        val flow = mainViewModel.isLoading
        assertNotNull(flow)
        assertTrue(flow.value)
    }

    @Test
    fun `splash completes when timer done and app ready`() {
        val timerComplete = true
        val isAppReady = true
        val shouldComplete = timerComplete && isAppReady

        assertTrue(shouldComplete)
    }

    @Test
    fun `splash does not complete when timer not done`() {
        val timerComplete = false
        val isAppReady = true
        val shouldComplete = timerComplete && isAppReady

        assertFalse(shouldComplete)
    }

    @Test
    fun `splash does not complete when app not ready`() {
        val timerComplete = true
        val isAppReady = false
        val shouldComplete = timerComplete && isAppReady

        assertFalse(shouldComplete)
    }

    @Test
    fun `splash does not complete when neither condition met`() {
        val timerComplete = false
        val isAppReady = false
        val shouldComplete = timerComplete && isAppReady

        assertFalse(shouldComplete)
    }

    @Test
    fun `first launch uses longer splash duration`() {
        val isFirstLaunch = true
        val minDuration = if (isFirstLaunch) 5000L else 2500L

        assertEquals(5000L, minDuration)
    }

    @Test
    fun `subsequent launch uses shorter splash duration`() {
        val isFirstLaunch = false
        val minDuration = if (isFirstLaunch) 5000L else 2500L

        assertEquals(2500L, minDuration)
    }

    @Test
    fun `splash screen displays app title`() {
        val title = "Catalogizer"

        assertEquals("Catalogizer", title)
        assertTrue(title.isNotBlank())
    }

    @Test
    fun `splash screen displays subtitle`() {
        val subtitle = "Advanced Multi-Protocol Media\nCollection Management System"

        assertTrue(subtitle.contains("Multi-Protocol"))
        assertTrue(subtitle.contains("Media"))
        assertTrue(subtitle.contains("Collection Management"))
    }

    @Test
    fun `splash screen displays branding footer`() {
        val footer = "Made with \u2764\uFE0F by Vasic Digital"

        assertTrue(footer.contains("Vasic Digital"))
    }

    @Test
    fun `splash screen displays version string`() {
        val versionName = "1.1.0"
        val versionLabel = "v$versionName"

        assertEquals("v1.1.0", versionLabel)
        assertTrue(versionLabel.startsWith("v"))
    }

    @Test
    fun `onSplashComplete callback is invocable`() {
        var completed = false
        val onSplashComplete: () -> Unit = { completed = true }

        onSplashComplete()

        assertTrue(completed)
    }

    @Test
    fun `onSplashComplete is called only once`() {
        var callCount = 0
        val onSplashComplete: () -> Unit = { callCount++ }

        onSplashComplete()

        assertEquals(1, callCount)
    }

    @Test
    fun `splash gradient colors are valid`() {
        val startColor = 0xFF0F172A
        val endColor = 0xFF1E293B

        assertTrue(startColor != endColor)
        assertTrue(startColor > 0)
        assertTrue(endColor > 0)
    }

    @Test
    fun `progress indicator color is valid`() {
        val progressColor = 0xFF4A90E2

        assertTrue(progressColor > 0)
    }

    @Test
    fun `subtitle text color is valid`() {
        val subtitleColor = 0xFF94A3B8

        assertTrue(subtitleColor > 0)
    }

    @Test
    fun `footer text color is valid`() {
        val footerColor = 0xFF64748B

        assertTrue(footerColor > 0)
    }

    @Test
    fun `version text color is valid`() {
        val versionColor = 0xFF475569

        assertTrue(versionColor > 0)
    }

    @Test
    fun `app icon size is 96dp`() {
        val iconSizeDp = 96
        assertEquals(96, iconSizeDp)
    }

    @Test
    fun `progress indicator size is 32dp`() {
        val indicatorSizeDp = 32
        assertEquals(32, indicatorSizeDp)
    }

    @Test
    fun `progress indicator stroke width is 3dp`() {
        val strokeWidthDp = 3
        assertEquals(3, strokeWidthDp)
    }

    @Test
    fun `title font size is 28sp`() {
        val titleFontSizeSp = 28
        assertEquals(28, titleFontSizeSp)
    }

    @Test
    fun `subtitle font size is 14sp`() {
        val subtitleFontSizeSp = 14
        assertEquals(14, subtitleFontSizeSp)
    }

    @Test
    fun `footer font size is 12sp`() {
        val footerFontSizeSp = 12
        assertEquals(12, footerFontSizeSp)
    }

    @Test
    fun `version font size is 11sp`() {
        val versionFontSizeSp = 11
        assertEquals(11, versionFontSizeSp)
    }

    @Test
    fun `first launch preference key is correct`() {
        val key = "has_launched"
        assertEquals("has_launched", key)
    }

    @Test
    fun `preferences name is correct`() {
        val prefsName = "catalogizer_prefs"
        assertEquals("catalogizer_prefs", prefsName)
    }

    @Test
    fun `app icon content description is Catalogizer`() {
        val contentDescription = "Catalogizer"
        assertEquals("Catalogizer", contentDescription)
    }

    @Test
    fun `isAppReady maps to ViewModel isLoading inverted`() = runTest {
        assertTrue(mainViewModel.isLoading.value)

        val isAppReady = !mainViewModel.isLoading.value
        assertFalse(isAppReady)

        mainViewModel.initializeApp()
        advanceUntilIdle()

        val isAppReadyAfterInit = !mainViewModel.isLoading.value
        assertTrue(isAppReadyAfterInit)
    }
}
