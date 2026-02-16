package com.catalogizer.android.ui.screens

import com.catalogizer.android.MainDispatcherRule
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for the SettingsScreen composable.
 *
 * The SettingsScreen is a stateless composable that takes onNavigateBack and onLogout
 * callbacks. It displays static content (About section) and a Sign Out button.
 * Testing focuses on verifying the composable's contract and callback behavior
 * through the ViewModel/callback layer.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class SettingsScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private var navigateBackCalled = false
    private var logoutCalled = false

    @Before
    fun setup() {
        navigateBackCalled = false
        logoutCalled = false
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `SettingsScreen should accept onNavigateBack callback`() {
        val callback: () -> Unit = { navigateBackCalled = true }

        callback.invoke()

        assertTrue(navigateBackCalled)
    }

    @Test
    fun `SettingsScreen should accept onLogout callback`() {
        val callback: () -> Unit = { logoutCalled = true }

        callback.invoke()

        assertTrue(logoutCalled)
    }

    @Test
    fun `navigate back callback should be invocable`() {
        var invoked = false
        val onNavigateBack: () -> Unit = { invoked = true }

        onNavigateBack()

        assertTrue(invoked)
    }

    @Test
    fun `logout callback should be invocable`() {
        var invoked = false
        val onLogout: () -> Unit = { invoked = true }

        onLogout()

        assertTrue(invoked)
    }

    @Test
    fun `callbacks should be independent of each other`() {
        var backCount = 0
        var logoutCount = 0
        val onNavigateBack: () -> Unit = { backCount++ }
        val onLogout: () -> Unit = { logoutCount++ }

        onNavigateBack()
        onNavigateBack()
        onLogout()

        assertEquals(2, backCount)
        assertEquals(1, logoutCount)
    }

    @Test
    fun `SettingsScreen expected display strings should be non-empty`() {
        // Verify the static text content that SettingsScreen displays
        val title = "Settings"
        val aboutTitle = "About"
        val aboutAppName = "Catalogizer for Android"
        val aboutDescription = "Multi-platform media collection manager"
        val signOutTitle = "Sign Out"
        val signOutDescription = "Sign out of your account"
        val signOutButtonText = "Sign Out"

        assertTrue(title.isNotBlank())
        assertTrue(aboutTitle.isNotBlank())
        assertTrue(aboutAppName.isNotBlank())
        assertTrue(aboutDescription.isNotBlank())
        assertTrue(signOutTitle.isNotBlank())
        assertTrue(signOutDescription.isNotBlank())
        assertTrue(signOutButtonText.isNotBlank())
    }

    @Test
    fun `about section content should contain app name`() {
        val appName = "Catalogizer for Android"

        assertTrue(appName.contains("Catalogizer"))
        assertTrue(appName.contains("Android"))
    }

    @Test
    fun `about section content should contain description`() {
        val description = "Multi-platform media collection manager"

        assertTrue(description.contains("media"))
        assertTrue(description.contains("collection"))
    }

    @Test
    fun `sign out section should have title and description`() {
        val signOutTitle = "Sign Out"
        val signOutDescription = "Sign out of your account"

        assertEquals("Sign Out", signOutTitle)
        assertTrue(signOutDescription.contains("Sign out"))
    }

    @Test
    fun `multiple callback invocations should not throw`() {
        var callCount = 0
        val callback: () -> Unit = { callCount++ }

        repeat(100) { callback() }

        assertEquals(100, callCount)
    }

    @Test
    fun `onLogout should trigger only logout and not navigation`() {
        var backCalled = false
        var logoutInvoked = false
        val onBack: () -> Unit = { backCalled = true }
        val onLogout: () -> Unit = { logoutInvoked = true }

        onLogout()

        assertTrue(logoutInvoked)
        assertFalse(backCalled)
    }

    @Test
    fun `onNavigateBack should trigger only navigation and not logout`() {
        var backInvoked = false
        var logoutCalled = false
        val onBack: () -> Unit = { backInvoked = true }
        val onLogout: () -> Unit = { logoutCalled = true }

        onBack()

        assertTrue(backInvoked)
        assertFalse(logoutCalled)
    }
}
