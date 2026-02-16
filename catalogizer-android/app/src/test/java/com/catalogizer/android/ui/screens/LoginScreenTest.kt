package com.catalogizer.android.ui.screens

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.AuthState
import com.catalogizer.android.data.repository.AuthRepository
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
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

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
class LoginScreenTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var authViewModel: AuthViewModel
    private val authStateFlow = MutableStateFlow(AuthState())

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
    fun `AuthViewModel initial state should not be authenticated`() = runTest {
        advanceUntilIdle()
        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `AuthViewModel initial state should not be loading`() = runTest {
        // After init completes, loading should be false
        advanceUntilIdle()
        val state = authViewModel.authState.value
        assertFalse(state.isLoading)
    }

    @Test
    fun `AuthViewModel initial state should have no error`() = runTest {
        advanceUntilIdle()
        val state = authViewModel.authState.value
        assertNull(state.error)
    }

    @Test
    fun `login should set loading state to true initially`() = runTest {
        coEvery { mockAuthRepository.login(any(), any(), any()) } coAnswers {
            kotlinx.coroutines.delay(1000)
            com.catalogizer.android.data.remote.ApiResult.success(mockk(relaxed = true))
        }

        advanceUntilIdle() // complete init checkAuthStatus coroutine

        authViewModel.login("user", "pass")
        // Advance just enough to start the coroutine (sets loading=true) but not past the delay
        advanceTimeBy(1)

        // Should be in loading state
        val state = authViewModel.authState.value
        assertTrue(state.isLoading)
    }

    @Test
    fun `login success should set authenticated state`() = runTest {
        val mockLoginResponse = mockk<com.catalogizer.android.data.models.LoginResponse>(relaxed = true)
        coEvery { mockAuthRepository.login(any(), any(), any()) } returns
            com.catalogizer.android.data.remote.ApiResult.success(mockLoginResponse)

        authViewModel.login("testuser", "testpass")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertTrue(state.isAuthenticated)
        assertFalse(state.isLoading)
    }

    @Test
    fun `login failure should set error state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any(), any()) } returns
            com.catalogizer.android.data.remote.ApiResult.error("Invalid credentials")

        authViewModel.login("wrong", "wrong")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertEquals("Invalid credentials", state.error)
    }

    @Test
    fun `login exception should set error state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any(), any()) } throws RuntimeException("Network error")

        authViewModel.login("user", "pass")
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
        assertEquals("Network error", state.error)
    }

    @Test
    fun `login should pass correct credentials to repository`() = runTest {
        coEvery { mockAuthRepository.login(any(), any(), any()) } returns
            com.catalogizer.android.data.remote.ApiResult.success(mockk(relaxed = true))

        authViewModel.login("myuser", "mypass")
        advanceUntilIdle()

        coVerify { mockAuthRepository.login("myuser", "mypass", any()) }
    }

    @Test
    fun `logout should set unauthenticated state`() = runTest {
        // First login
        coEvery { mockAuthRepository.login(any(), any(), any()) } returns
            com.catalogizer.android.data.remote.ApiResult.success(mockk(relaxed = true))
        authViewModel.login("user", "pass")
        advanceUntilIdle()

        // Then logout
        authViewModel.logout()
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `logout should call repository logout`() = runTest {
        authViewModel.logout()
        advanceUntilIdle()

        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `login error should clear previous error on retry`() = runTest {
        // First attempt fails
        coEvery { mockAuthRepository.login(any(), any(), any()) } returns
            com.catalogizer.android.data.remote.ApiResult.error("Error 1")
        authViewModel.login("user", "pass")
        advanceUntilIdle()
        assertEquals("Error 1", authViewModel.authState.value.error)

        // Second attempt - should clear the error and succeed
        coEvery { mockAuthRepository.login(any(), any(), any()) } returns
            com.catalogizer.android.data.remote.ApiResult.success(mockk(relaxed = true))
        authViewModel.login("user", "pass")
        advanceUntilIdle()

        // After successful login, error should be null and user authenticated
        assertNull(authViewModel.authState.value.error)
        assertTrue(authViewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `AuthState default values should be correct`() {
        val state = AuthState()

        assertFalse(state.isAuthenticated)
        assertFalse(state.isLoading)
        assertNull(state.error)
        assertNull(state.user)
    }

    @Test
    fun `AuthState copy should preserve unmodified fields`() {
        val original = AuthState(isAuthenticated = true, isLoading = false, error = null)
        val copied = original.copy(isLoading = true)

        assertTrue(copied.isAuthenticated)
        assertTrue(copied.isLoading)
        assertNull(copied.error)
    }

    @Test
    fun `checkAuthStatus should update state based on repository`() = runTest {
        coEvery { mockAuthRepository.isAuthenticated() } returns true

        // Re-create ViewModel to trigger init block
        authViewModel = AuthViewModel(mockAuthRepository)
        advanceUntilIdle()

        val state = authViewModel.authState.value
        assertTrue(state.isAuthenticated)
    }
}
