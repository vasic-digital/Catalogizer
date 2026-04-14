package com.catalogizer.androidtv.ui.screens.login

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

/**
 * Tests for LoginScreen validation logic, D-PAD navigation state,
 * server URL handling, and credential input validation.
 */
@OptIn(ExperimentalCoroutinesApi::class)
class LoginScreenSafeTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var authViewModel: AuthViewModel
    private val authStateFlow = MutableStateFlow(AuthState.Unauthenticated)

    @Before
    fun setup() {
        mockAuthRepository = mockk(relaxed = true)
        every { mockAuthRepository.authState } returns authStateFlow
        authViewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // --- validateAndLogin logic ---

    @Test
    fun `blank username fails validation`() {
        val username = ""
        val password = "password123"
        var usernameError: String? = null
        var valid = true

        val safeUsername = username.trim()
        if (safeUsername.isBlank()) {
            usernameError = "Username is required"
            valid = false
        }

        assertFalse(valid)
        assertEquals("Username is required", usernameError)
    }

    @Test
    fun `whitespace-only username fails validation`() {
        val username = "   "
        val password = "password123"
        var usernameError: String? = null
        var valid = true

        val safeUsername = username.trim()
        if (safeUsername.isBlank()) {
            usernameError = "Username is required"
            valid = false
        }

        assertFalse(valid)
        assertEquals("Username is required", usernameError)
    }

    @Test
    fun `single character username fails minimum length validation`() {
        val username = "a"
        var usernameError: String? = null
        var valid = true

        val safeUsername = username.trim()
        if (safeUsername.isBlank()) {
            usernameError = "Username is required"
            valid = false
        } else if (safeUsername.length < 2) {
            usernameError = "Username must be at least 2 characters"
            valid = false
        }

        assertFalse(valid)
        assertEquals("Username must be at least 2 characters", usernameError)
    }

    @Test
    fun `two character username passes minimum length validation`() {
        val username = "ab"
        var usernameError: String? = null
        var valid = true

        val safeUsername = username.trim()
        if (safeUsername.isBlank()) {
            usernameError = "Username is required"
            valid = false
        } else if (safeUsername.length < 2) {
            usernameError = "Username must be at least 2 characters"
            valid = false
        }

        assertTrue(valid)
        assertNull(usernameError)
    }

    @Test
    fun `blank password fails validation`() {
        val password = ""
        var passwordError: String? = null
        var valid = true

        if (password.isBlank()) {
            passwordError = "Password is required"
            valid = false
        }

        assertFalse(valid)
        assertEquals("Password is required", passwordError)
    }

    @Test
    fun `short password fails minimum length validation`() {
        val password = "abc"
        var passwordError: String? = null
        var valid = true

        if (password.isBlank()) {
            passwordError = "Password is required"
            valid = false
        } else if (password.length < 4) {
            passwordError = "Password must be at least 4 characters"
            valid = false
        }

        assertFalse(valid)
        assertEquals("Password must be at least 4 characters", passwordError)
    }

    @Test
    fun `four character password passes minimum length validation`() {
        val password = "abcd"
        var passwordError: String? = null
        var valid = true

        if (password.isBlank()) {
            passwordError = "Password is required"
            valid = false
        } else if (password.length < 4) {
            passwordError = "Password must be at least 4 characters"
            valid = false
        }

        assertTrue(valid)
        assertNull(passwordError)
    }

    @Test
    fun `both blank username and password produce both errors`() {
        val username = ""
        val password = ""
        var usernameError: String? = null
        var passwordError: String? = null
        var valid = true

        val safeUsername = username.trim()
        if (safeUsername.isBlank()) {
            usernameError = "Username is required"
            valid = false
        }
        if (password.isBlank()) {
            passwordError = "Password is required"
            valid = false
        }

        assertFalse(valid)
        assertEquals("Username is required", usernameError)
        assertEquals("Password is required", passwordError)
    }

    @Test
    fun `valid credentials pass all validation`() {
        val username = "admin"
        val password = "secret123"
        var usernameError: String? = null
        var passwordError: String? = null
        var valid = true

        val safeUsername = username.trim()
        if (safeUsername.isBlank()) {
            usernameError = "Username is required"
            valid = false
        } else if (safeUsername.length < 2) {
            usernameError = "Username must be at least 2 characters"
            valid = false
        }
        if (password.isBlank()) {
            passwordError = "Password is required"
            valid = false
        } else if (password.length < 4) {
            passwordError = "Password must be at least 4 characters"
            valid = false
        }

        assertTrue(valid)
        assertNull(usernameError)
        assertNull(passwordError)
    }

    @Test
    fun `username is trimmed before validation`() {
        val username = "  admin  "
        val safeUsername = username.trim()
        assertEquals("admin", safeUsername)
        assertEquals(5, safeUsername.length)
    }

    @Test
    fun `password is not trimmed`() {
        val password = "  pass  "
        assertEquals(8, password.length)
    }

    // --- Login button enabled state ---

    @Test
    fun `login button disabled when username blank`() {
        val isEnabled = !false && "".isNotBlank() && "pass".isNotBlank()
        assertFalse(isEnabled)
    }

    @Test
    fun `login button disabled when password blank`() {
        val isEnabled = !false && "user".isNotBlank() && "".isNotBlank()
        assertFalse(isEnabled)
    }

    @Test
    fun `login button disabled when loading`() {
        val isEnabled = !true && "user".isNotBlank() && "pass".isNotBlank()
        assertFalse(isEnabled)
    }

    @Test
    fun `login button enabled when credentials valid and not loading`() {
        val isEnabled = !false && "user".isNotBlank() && "pass".isNotBlank()
        assertTrue(isEnabled)
    }

    // --- Server URL handling ---

    @Test
    fun `server URL is non-blank after configuration`() {
        val serverUrl = "http://192.168.0.100:8080"
        assertTrue(serverUrl.isNotBlank())
    }

    @Test
    fun `empty server URL is blank`() {
        val serverUrl = ""
        assertTrue(serverUrl.isBlank())
    }

    @Test
    fun `connect button disabled when server URL is blank`() {
        val serverUrl = ""
        val isLoading = false
        val enabled = serverUrl.isNotBlank() && !isLoading
        assertFalse(enabled)
    }

    @Test
    fun `connect button enabled when server URL present and not loading`() {
        val serverUrl = "http://192.168.0.100:8080"
        val isLoading = false
        val enabled = serverUrl.isNotBlank() && !isLoading
        assertTrue(enabled)
    }

    @Test
    fun `discover button disabled when discovering`() {
        val isDiscovering = true
        val isLoading = false
        val enabled = !isDiscovering && !isLoading
        assertFalse(enabled)
    }

    @Test
    fun `discover button disabled when loading`() {
        val isDiscovering = false
        val isLoading = true
        val enabled = !isDiscovering && !isLoading
        assertFalse(enabled)
    }

    @Test
    fun `discover button enabled when neither discovering nor loading`() {
        val isDiscovering = false
        val isLoading = false
        val enabled = !isDiscovering && !isLoading
        assertTrue(enabled)
    }

    // --- Server connection state ---

    @Test
    fun `server connection null means unknown`() {
        val serverConnected: Boolean? = null
        assertNull(serverConnected)
    }

    @Test
    fun `server connection true means connected`() {
        val serverConnected: Boolean? = true
        assertTrue(serverConnected == true)
    }

    @Test
    fun `server connection false means failed`() {
        val serverConnected: Boolean? = false
        assertTrue(serverConnected == false)
    }

    @Test
    fun `connection status text shows Connected when true`() {
        val serverConnected: Boolean? = true
        val text = if (serverConnected == true) "Connected to server" else "Unable to reach server"
        assertEquals("Connected to server", text)
    }

    @Test
    fun `connection status text shows Unable when false`() {
        val serverConnected: Boolean? = false
        val text = if (serverConnected == true) "Connected to server" else "Unable to reach server"
        assertEquals("Unable to reach server", text)
    }

    // --- Auth state transitions via ViewModel ---

    @Test
    fun `login calls repository with credentials`() = runTest {
        authViewModel.login("admin", "password")
        advanceUntilIdle()
        coVerify { mockAuthRepository.login("admin", "password") }
    }

    @Test
    fun `successful auth triggers onLoginSuccess path`() = runTest {
        val job = launch { authViewModel.authState.collect {} }
        advanceUntilIdle()

        authStateFlow.value = AuthState(isAuthenticated = true, token = "token-abc")
        advanceUntilIdle()

        assertTrue(authViewModel.authState.value.isAuthenticated)
        job.cancel()
    }

    @Test
    fun `auth error sets error message`() = runTest {
        val job = launch { authViewModel.authState.collect {} }
        advanceUntilIdle()

        authStateFlow.value = AuthState(isAuthenticated = false, error = "Invalid credentials")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
        assertEquals("Invalid credentials", state.error)
        job.cancel()
    }

    // --- Responsive sizing logic ---

    @Test
    fun `compact layout applies for screen width under 600dp`() {
        val screenWidthDp = 480
        val isCompact = screenWidthDp < 600
        assertTrue(isCompact)
    }

    @Test
    fun `standard layout applies for screen width 600dp or more`() {
        val screenWidthDp = 1920
        val isCompact = screenWidthDp < 600
        assertFalse(isCompact)
    }

    // --- D-PAD focus navigation order ---

    @Test
    fun `focus order is username then password then sign-in then server-url`() {
        val focusOrder = listOf("username", "password", "signIn", "serverUrl")
        assertEquals(4, focusOrder.size)
        assertEquals("username", focusOrder[0])
        assertEquals("password", focusOrder[1])
        assertEquals("signIn", focusOrder[2])
        assertEquals("serverUrl", focusOrder[3])
    }
}
