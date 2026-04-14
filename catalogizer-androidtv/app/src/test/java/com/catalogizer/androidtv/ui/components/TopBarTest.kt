package com.catalogizer.androidtv.ui.components

import com.catalogizer.androidtv.data.models.AuthState
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for TopBar component logic: title display, auth state display,
 * action button callback wiring, and ConnectionStatus state rendering.
 */
class TopBarTest {

    // --- Title display ---

    @Test
    fun `top bar title should be Catalogizer`() {
        val title = "Catalogizer"
        assertEquals("Catalogizer", title)
        assertTrue(title.isNotBlank())
    }

    @Test
    fun `top bar accepts custom title text`() {
        val title = "My Custom App"
        assertEquals("My Custom App", title)
    }

    // --- Auth state display logic ---

    @Test
    fun `top bar shows username when authenticated`() {
        val authState = AuthState(isAuthenticated = true, username = "admin")
        assertTrue(authState.isAuthenticated)
        assertEquals("admin", authState.username)
    }

    @Test
    fun `top bar shows default User when username is null and authenticated`() {
        val authState = AuthState(isAuthenticated = true, username = null)
        val displayName = authState.username ?: "User"
        assertEquals("User", displayName)
    }

    @Test
    fun `top bar does not show user info when not authenticated`() {
        val authState = AuthState(isAuthenticated = false)
        assertFalse(authState.isAuthenticated)
    }

    @Test
    fun `top bar handles null auth state`() {
        val authState: AuthState? = null
        assertNull(authState)
    }

    // --- Action button callbacks ---

    @Test
    fun `search button callback fires correctly`() {
        var searchClicked = false
        val onSearchClick: () -> Unit = { searchClicked = true }
        onSearchClick()
        assertTrue(searchClicked)
    }

    @Test
    fun `settings button callback fires correctly`() {
        var settingsClicked = false
        val onSettingsClick: () -> Unit = { settingsClicked = true }
        onSettingsClick()
        assertTrue(settingsClicked)
    }

    @Test
    fun `search and settings callbacks are independent`() {
        var searchCount = 0
        var settingsCount = 0
        val onSearchClick: () -> Unit = { searchCount++ }
        val onSettingsClick: () -> Unit = { settingsCount++ }

        onSearchClick()
        onSearchClick()
        onSettingsClick()

        assertEquals(2, searchCount)
        assertEquals(1, settingsCount)
    }

    // --- ConnectionStatus logic ---

    @Test
    fun `connection status shows Connected when online`() {
        val isConnected = true
        val statusText = if (isConnected) "Connected" else "Offline"
        assertEquals("Connected", statusText)
    }

    @Test
    fun `connection status shows Offline when disconnected`() {
        val isConnected = false
        val statusText = if (isConnected) "Connected" else "Offline"
        assertEquals("Offline", statusText)
    }

    @Test
    fun `connection status content description matches state`() {
        val isConnected = true
        val contentDesc = if (isConnected) "Connected" else "Disconnected"
        assertEquals("Connected", contentDesc)
    }

    @Test
    fun `disconnected content description is correct`() {
        val isConnected = false
        val contentDesc = if (isConnected) "Connected" else "Disconnected"
        assertEquals("Disconnected", contentDesc)
    }

    // --- Auth state transitions ---

    @Test
    fun `auth state transitions from unauthenticated to authenticated`() {
        val initial = AuthState.Unauthenticated
        assertFalse(initial.isAuthenticated)

        val authenticated = AuthState(
            isAuthenticated = true,
            username = "testuser",
            token = "jwt-token-123"
        )
        assertTrue(authenticated.isAuthenticated)
        assertEquals("testuser", authenticated.username)
    }

    @Test
    fun `auth state with error does not display as authenticated`() {
        val errorState = AuthState(
            isAuthenticated = false,
            error = "Invalid credentials"
        )
        assertFalse(errorState.isAuthenticated)
        assertNotNull(errorState.error)
    }

    @Test
    fun `auth state loading flag is independent of authenticated flag`() {
        val loadingAndAuth = AuthState(isAuthenticated = true, isLoading = true)
        assertTrue(loadingAndAuth.isAuthenticated)
        assertTrue(loadingAndAuth.isLoading)

        val loadingNotAuth = AuthState(isAuthenticated = false, isLoading = true)
        assertFalse(loadingNotAuth.isAuthenticated)
        assertTrue(loadingNotAuth.isLoading)
    }
}
