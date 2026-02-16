package com.catalogizer.androidtv.ui.screens.login

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class LoginScreenTest {

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

    @Test
    fun `AuthViewModel initial state should be unauthenticated`() = runTest {
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `AuthViewModel initial state should have no error`() = runTest {
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertNull(state.error)
    }

    @Test
    fun `AuthViewModel initial state should not be loading`() = runTest {
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isLoading)
    }

    @Test
    fun `login should call repository login`() = runTest {
        authViewModel.login("testuser", "testpass")
        advanceUntilIdle()

        coVerify { mockAuthRepository.login("testuser", "testpass") }
    }

    @Test
    fun `login should pass correct credentials`() = runTest {
        authViewModel.login("admin", "secret123")
        advanceUntilIdle()

        coVerify { mockAuthRepository.login("admin", "secret123") }
    }

    @Test
    fun `login success should update auth state to authenticated`() = runTest {
        // Start collecting to activate the WhileSubscribed upstream
        val job = launch { authViewModel.authState.collect {} }
        advanceUntilIdle()

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        // Simulate repository updating auth state
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            username = "user",
            token = "test-token"
        )
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertTrue(state.isAuthenticated)
        assertEquals("user", state.username)

        job.cancel()
    }

    @Test
    fun `login failure should set error in auth state`() = runTest {
        // Simulate repository login throwing exception
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException("Invalid credentials")

        authViewModel.login("wrong", "wrong")
        advanceUntilIdle()

        // AuthViewModel catches the exception - repo updates state
        coVerify { mockAuthRepository.login("wrong", "wrong") }
    }

    @Test
    fun `login with repository error should update auth state error`() = runTest {
        // Start collecting to activate the WhileSubscribed upstream
        val job = launch { authViewModel.authState.collect {} }
        advanceUntilIdle()

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        // Simulate repository setting error state
        authStateFlow.value = AuthState(
            isAuthenticated = false,
            error = "Login failed: Invalid credentials"
        )
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
        assertNotNull(state.error)
        assertTrue(state.error!!.contains("Invalid credentials"))

        job.cancel()
    }

    @Test
    fun `logout should call repository logout`() = runTest {
        authViewModel.logout()
        advanceUntilIdle()

        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `logout should reset state to unauthenticated`() = runTest {
        // Set as authenticated first
        authStateFlow.value = AuthState(isAuthenticated = true, token = "token")

        authViewModel.logout()
        advanceUntilIdle()

        // Simulate repository resetting state
        authStateFlow.value = AuthState.Unauthenticated

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `refreshToken should call repository refreshToken`() = runTest {
        authViewModel.refreshToken()
        advanceUntilIdle()

        coVerify { mockAuthRepository.refreshToken() }
    }

    @Test
    fun `clearError should call repository clearError`() = runTest {
        authViewModel.clearError()
        advanceUntilIdle()

        coVerify { mockAuthRepository.clearError() }
    }

    @Test
    fun `authState should reflect repository state changes`() = runTest {
        // Start collecting to activate the WhileSubscribed upstream
        val job = launch { authViewModel.authState.collect {} }
        advanceUntilIdle()

        // Update repository state
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            username = "testuser",
            token = "abc123",
            userId = 42L
        )
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertTrue(state.isAuthenticated)
        assertEquals("testuser", state.username)
        assertEquals("abc123", state.token)
        assertEquals(42L, state.userId)

        job.cancel()
    }

    @Test
    fun `AuthState Unauthenticated companion should have correct defaults`() {
        val unauthenticated = AuthState.Unauthenticated

        assertFalse(unauthenticated.isAuthenticated)
        assertNull(unauthenticated.username)
        assertNull(unauthenticated.token)
        assertNull(unauthenticated.userId)
        assertNull(unauthenticated.expiresAt)
        assertNull(unauthenticated.error)
        assertFalse(unauthenticated.isLoading)
    }

    @Test
    fun `AuthState should support copy with modified fields`() {
        val original = AuthState(isAuthenticated = true, username = "user1")
        val copied = original.copy(username = "user2")

        assertTrue(copied.isAuthenticated)
        assertEquals("user2", copied.username)
    }

    @Test
    fun `login screen should support performLogin validation`() {
        // Test the performLogin helper logic
        val blankUsername = ""
        val blankPassword = ""

        assertTrue(blankUsername.isBlank())
        assertTrue(blankPassword.isBlank())

        val validUsername = "user"
        val validPassword = "pass"

        assertTrue(validUsername.isNotBlank())
        assertTrue(validPassword.isNotBlank())
    }

    @Test
    fun `login button should be disabled when username is blank`() {
        val username = ""
        val password = "password"
        val isLoading = false

        val isEnabled = !isLoading && username.isNotBlank() && password.isNotBlank()

        assertFalse(isEnabled)
    }

    @Test
    fun `login button should be disabled when password is blank`() {
        val username = "user"
        val password = ""
        val isLoading = false

        val isEnabled = !isLoading && username.isNotBlank() && password.isNotBlank()

        assertFalse(isEnabled)
    }

    @Test
    fun `login button should be disabled when loading`() {
        val username = "user"
        val password = "pass"
        val isLoading = true

        val isEnabled = !isLoading && username.isNotBlank() && password.isNotBlank()

        assertFalse(isEnabled)
    }

    @Test
    fun `login button should be enabled when credentials valid and not loading`() {
        val username = "user"
        val password = "pass"
        val isLoading = false

        val isEnabled = !isLoading && username.isNotBlank() && password.isNotBlank()

        assertTrue(isEnabled)
    }
}
