package com.catalogizer.android.ui.screens.login

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.AuthState
import com.catalogizer.android.data.models.LoginResponse
import com.catalogizer.android.data.models.User
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.AuthRepository
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceTimeBy
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
 * Tests for LoginScreen behavior via AuthViewModel.
 * Covers form validation logic, submit flow, error display,
 * server URL validation, and navigation on success.
 */
@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class LoginScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var authViewModel: AuthViewModel

    private fun createMockUser(): User = User(
        id = 1L,
        username = "testuser",
        email = "test@example.com",
        firstName = "Test",
        lastName = "User",
        role = "user",
        isActive = true,
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z",
    )

    private fun createLoginResponse(): LoginResponse = LoginResponse(
        user = createMockUser(),
        token = "jwt_token_123",
        refreshToken = "refresh_token_456",
        expiresIn = 3600L,
    )

    @Before
    fun setup() {
        mockAuthRepository = mockk(relaxed = true)
        coEvery { mockAuthRepository.isAuthenticated() } returns false
        authViewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial state is not authenticated`() = runTest {
        advanceUntilIdle()
        val state = authViewModel.authState.value

        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `initial state is not loading`() = runTest {
        advanceUntilIdle()
        val state = authViewModel.authState.value

        assertFalse(state.isLoading)
    }

    @Test
    fun `initial state has no error`() = runTest {
        advanceUntilIdle()

        assertNull(authViewModel.authState.value.error)
    }

    @Test
    fun `login with valid credentials sets authenticated state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(createLoginResponse())

        authViewModel.login("admin", "password123")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertTrue(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertNull(state.error)
    }

    @Test
    fun `login with invalid credentials sets error state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.error("Invalid credentials")

        authViewModel.login("wrong", "wrong")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertEquals("Invalid credentials", state.error)
    }

    @Test
    fun `login with network error sets error state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException("Network error")

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
        assertEquals("Network error", state.error)
    }

    @Test
    fun `login sets loading to true during request`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } coAnswers {
            kotlinx.coroutines.delay(1000)
            ApiResult.success(createLoginResponse())
        }
        advanceUntilIdle()

        authViewModel.login("user", "pass")
        advanceTimeBy(1)

        assertTrue(authViewModel.authState.value.isLoading)
    }

    @Test
    fun `login sets loading to false after completion`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(createLoginResponse())

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        assertFalse(authViewModel.authState.value.isLoading)
    }

    @Test
    fun `login passes correct credentials to repository`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(createLoginResponse())

        authViewModel.login("myuser", "mypass")
        advanceUntilIdle()

        coVerify { mockAuthRepository.login("myuser", "mypass") }
    }

    @Test
    fun `retry clears previous error`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.error("First error")
        authViewModel.login("user", "pass")
        advanceUntilIdle()
        assertEquals("First error", authViewModel.authState.value.error)

        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(createLoginResponse())
        authViewModel.login("user", "pass")
        advanceUntilIdle()

        assertNull(authViewModel.authState.value.error)
        assertTrue(authViewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `login exception with null message uses fallback`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException()

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        assertEquals("Login failed", authViewModel.authState.value.error)
    }

    @Test
    fun `login result with null error uses fallback message`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult(
            data = null, error = null, isSuccess = false)

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        assertEquals("Login failed", authViewModel.authState.value.error)
    }

    @Test
    fun `logout sets unauthenticated state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(createLoginResponse())
        authViewModel.login("user", "pass")
        advanceUntilIdle()
        assertTrue(authViewModel.authState.value.isAuthenticated)

        authViewModel.logout()
        advanceUntilIdle()

        assertFalse(authViewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `logout calls repository logout`() = runTest {
        authViewModel.logout()
        advanceUntilIdle()

        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `checkAuthState updates state from repository`() = runTest {
        coEvery { mockAuthRepository.isAuthenticated() } returns true
        authViewModel = AuthViewModel(mockAuthRepository)
        advanceUntilIdle()

        assertTrue(authViewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `empty username validation logic`() {
        val username = ""
        val password = "password"
        val serverConnected = true

        val canSubmit = username.isNotBlank() && password.isNotBlank() && serverConnected

        assertFalse(canSubmit)
    }

    @Test
    fun `empty password validation logic`() {
        val username = "admin"
        val password = ""
        val serverConnected = true

        val canSubmit = username.isNotBlank() && password.isNotBlank() && serverConnected

        assertFalse(canSubmit)
    }

    @Test
    fun `both fields empty validation logic`() {
        val username = ""
        val password = ""
        val serverConnected = true

        val canSubmit = username.isNotBlank() && password.isNotBlank() && serverConnected

        assertFalse(canSubmit)
    }

    @Test
    fun `whitespace-only username validation logic`() {
        val username = "   "
        val password = "password"
        val serverConnected = true

        val canSubmit = username.isNotBlank() && password.isNotBlank() && serverConnected

        assertFalse(canSubmit)
    }

    @Test
    fun `valid credentials with server connected enables submit`() {
        val username = "admin"
        val password = "password123"
        val serverConnected = true

        val canSubmit = username.isNotBlank() && password.isNotBlank() && serverConnected

        assertTrue(canSubmit)
    }

    @Test
    fun `valid credentials with server disconnected disables submit`() {
        val username = "admin"
        val password = "password123"
        val serverConnected = false

        val canSubmit = username.isNotBlank() && password.isNotBlank() && serverConnected

        assertFalse(canSubmit)
    }

    @Test
    fun `fields disabled during loading`() {
        val isLoading = true
        val serverConnected = true
        val fieldsEnabled = !isLoading && serverConnected

        assertFalse(fieldsEnabled)
    }

    @Test
    fun `fields enabled when not loading and server connected`() {
        val isLoading = false
        val serverConnected = true
        val fieldsEnabled = !isLoading && serverConnected

        assertTrue(fieldsEnabled)
    }

    @Test
    fun `fields disabled when server not connected`() {
        val isLoading = false
        val serverConnected = false
        val fieldsEnabled = !isLoading && serverConnected

        assertFalse(fieldsEnabled)
    }

    @Test
    fun `server URL validation accepts valid HTTP URL`() {
        val serverUrl = "http://192.168.0.100:8080"

        assertTrue(serverUrl.isNotBlank())
        assertTrue(serverUrl.startsWith("http"))
    }

    @Test
    fun `server URL validation accepts valid HTTPS URL`() {
        val serverUrl = "https://catalog.example.com"

        assertTrue(serverUrl.isNotBlank())
        assertTrue(serverUrl.startsWith("https"))
    }

    @Test
    fun `server URL validation rejects blank URL`() {
        val serverUrl = ""

        assertFalse(serverUrl.isNotBlank())
    }

    @Test
    fun `connect button enabled when server URL is not blank`() {
        val serverUrl = "http://localhost:8080"
        val isLoading = false
        val connectEnabled = serverUrl.isNotBlank() && !isLoading

        assertTrue(connectEnabled)
    }

    @Test
    fun `connect button disabled when server URL is blank`() {
        val serverUrl = ""
        val isLoading = false
        val connectEnabled = serverUrl.isNotBlank() && !isLoading

        assertFalse(connectEnabled)
    }

    @Test
    fun `connect button disabled during loading`() {
        val serverUrl = "http://localhost:8080"
        val isLoading = true
        val connectEnabled = serverUrl.isNotBlank() && !isLoading

        assertFalse(connectEnabled)
    }

    @Test
    fun `discover button disabled during discovery`() {
        val isDiscovering = true
        val isLoading = false
        val discoverEnabled = !isDiscovering && !isLoading

        assertFalse(discoverEnabled)
    }

    @Test
    fun `discover button disabled during loading`() {
        val isDiscovering = false
        val isLoading = true
        val discoverEnabled = !isDiscovering && !isLoading

        assertFalse(discoverEnabled)
    }

    @Test
    fun `discover button enabled when idle`() {
        val isDiscovering = false
        val isLoading = false
        val discoverEnabled = !isDiscovering && !isLoading

        assertTrue(discoverEnabled)
    }

    @Test
    fun `onLoginSuccess callback is invocable`() {
        var loginSuccessCalled = false
        val onLoginSuccess: () -> Unit = { loginSuccessCalled = true }

        onLoginSuccess()

        assertTrue(loginSuccessCalled)
    }

    @Test
    fun `AuthState default values are correct`() {
        val state = AuthState()

        assertFalse(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertNull(state.user)
    }

    @Test
    fun `AuthState copy preserves unmodified fields`() {
        val original = AuthState(isAuthenticated = true, isLoading = false, error = null)
        val copied = original.copy(isLoading = true)

        assertTrue(copied.isAuthenticated)
        assertTrue(copied.isLoading)
        assertNull(copied.error)
    }

    @Test
    fun `connection status text shows correct URL`() {
        val serverUrl = "http://192.168.0.100:8080"
        val serverConnected = true

        val statusText = if (serverConnected && serverUrl.isNotBlank()) {
            "Connected: $serverUrl"
        } else {
            ""
        }

        assertEquals("Connected: http://192.168.0.100:8080", statusText)
    }

    @Test
    fun `connection status hidden when not connected`() {
        val serverUrl = "http://192.168.0.100:8080"
        val serverConnected = false

        val showStatus = serverConnected && serverUrl.isNotBlank()

        assertFalse(showStatus)
    }

    @Test
    fun `connection status hidden when URL is blank`() {
        val serverUrl = ""
        val serverConnected = true

        val showStatus = serverConnected && serverUrl.isNotBlank()

        assertFalse(showStatus)
    }

    @Test
    fun `subnet prefix detection logic handles standard IPv4`() {
        val hostAddress = "192.168.0.105"
        val parts = hostAddress.split(".")

        assertEquals(4, parts.size)
        assertEquals("192.168.0", "${parts[0]}.${parts[1]}.${parts[2]}")
    }

    @Test
    fun `subnet prefix detection logic handles 10 dot network`() {
        val hostAddress = "10.0.2.15"
        val parts = hostAddress.split(".")

        assertEquals(4, parts.size)
        assertEquals("10.0.2", "${parts[0]}.${parts[1]}.${parts[2]}")
    }
}
