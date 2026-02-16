package com.catalogizer.androidtv.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@ExperimentalCoroutinesApi
class AuthViewModelTest {

    @get:Rule
    val instantTaskExecutorRule = InstantTaskExecutorRule()

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var authRepository: AuthRepository
    private lateinit var viewModel: AuthViewModel
    private val mockAuthState = MutableStateFlow(AuthState.Unauthenticated)

    @Before
    fun setup() {
        authRepository = mockk()
        every { authRepository.authState } returns mockAuthState
        viewModel = AuthViewModel(authRepository)
    }

    @Test
    fun `authState exposes repository authState with correct initial value`() = runTest {
        advanceUntilIdle()
        val initialState = viewModel.authState.value
        assertEquals(AuthState.Unauthenticated, initialState)
    }

    @Test
    fun `login calls repository login with correct parameters`() = runTest {
        coEvery { authRepository.login(any(), any()) } returns Unit

        viewModel.login("testuser", "password123")

        advanceUntilIdle()

        coVerify { authRepository.login("testuser", "password123") }
    }

    @Test
    fun `login handles exceptions gracefully`() = runTest {
        val exception = RuntimeException("Login failed")
        coEvery { authRepository.login(any(), any()) } throws exception

        // Should not throw exception
        viewModel.login("testuser", "password123")

        advanceUntilIdle()

        coVerify { authRepository.login("testuser", "password123") }
    }

    @Test
    fun `logout calls repository logout`() = runTest {
        coEvery { authRepository.logout() } returns Unit

        viewModel.logout()

        advanceUntilIdle()

        coVerify { authRepository.logout() }
    }

    @Test
    fun `refreshToken calls repository refreshToken`() = runTest {
        coEvery { authRepository.refreshToken() } returns Unit

        viewModel.refreshToken()

        advanceUntilIdle()

        coVerify { authRepository.refreshToken() }
    }

    @Test
    fun `clearError calls repository clearError`() = runTest {
        coEvery { authRepository.clearError() } returns Unit

        viewModel.clearError()

        advanceUntilIdle()

        coVerify { authRepository.clearError() }
    }

    @Test
    fun `authState reflects repository state changes`() = runTest {
        // Start collecting to activate the WhileSubscribed upstream
        val job = launch { viewModel.authState.collect {} }
        advanceUntilIdle()

        // Initial state
        var state = viewModel.authState.value
        assertEquals(AuthState.Unauthenticated, state)

        // Update repository state
        val authenticatedState = AuthState(
            isAuthenticated = true,
            token = "test-token",
            username = "testuser",
            userId = 123L
        )
        mockAuthState.value = authenticatedState
        advanceUntilIdle()

        // ViewModel should reflect the change
        state = viewModel.authState.value
        assertEquals(authenticatedState, state)
        assertTrue(state.isAuthenticated)
        assertEquals("test-token", state.token)
        assertEquals("testuser", state.username)
        assertEquals(123L, state.userId)

        job.cancel()
    }

    @Test
    fun `authState reflects error states from repository`() = runTest {
        // Start collecting to activate the WhileSubscribed upstream
        val job = launch { viewModel.authState.collect {} }
        advanceUntilIdle()

        // Set error state in repository
        val errorState = AuthState(
            isAuthenticated = false,
            error = "Invalid credentials"
        )
        mockAuthState.value = errorState
        advanceUntilIdle()

        val state = viewModel.authState.value
        assertEquals(errorState, state)
        assertFalse(state.isAuthenticated)
        assertEquals("Invalid credentials", state.error)

        job.cancel()
    }

    @Test
    fun `multiple login calls work correctly`() = runTest {
        coEvery { authRepository.login(any(), any()) } returns Unit

        viewModel.login("user1", "pass1")
        viewModel.login("user2", "pass2")

        advanceUntilIdle()

        coVerify { authRepository.login("user1", "pass1") }
        coVerify { authRepository.login("user2", "pass2") }
    }

    @Test
    fun `logout followed by login works correctly`() = runTest {
        coEvery { authRepository.logout() } returns Unit
        coEvery { authRepository.login(any(), any()) } returns Unit

        viewModel.logout()
        advanceUntilIdle()

        viewModel.login("testuser", "password")
        advanceUntilIdle()

        coVerify { authRepository.logout() }
        coVerify { authRepository.login("testuser", "password") }
    }

    @Test
    fun `multiple login calls are handled correctly`() = runTest {
        coEvery { authRepository.login(any(), any()) } returns Unit

        // Test multiple concurrent login calls
        viewModel.login("user1", "pass1")
        viewModel.login("user2", "pass2")

        advanceUntilIdle()

        // Both calls should have been made to the repository
        coVerify { authRepository.login("user1", "pass1") }
        coVerify { authRepository.login("user2", "pass2") }
    }
}