package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class AuthViewModelTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var authStateFlow: MutableStateFlow<AuthState>
    private lateinit var viewModel: AuthViewModel

    @Before
    fun setup() {
        authStateFlow = MutableStateFlow(AuthState.Unauthenticated)
        mockAuthRepository = mockk(relaxed = true)
        every { mockAuthRepository.authState } returns authStateFlow
        viewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial auth state is unauthenticated`() = runTest {
        advanceUntilIdle()
        assertFalse(viewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `login calls repository login`() = runTest {
        viewModel.login("admin", "password")
        advanceUntilIdle()

        coVerify { mockAuthRepository.login("admin", "password") }
    }

    @Test
    fun `logout calls repository logout`() = runTest {
        viewModel.logout()
        advanceUntilIdle()

        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `refreshToken calls repository refreshToken`() = runTest {
        viewModel.refreshToken()
        advanceUntilIdle()

        coVerify { mockAuthRepository.refreshToken() }
    }

    @Test
    fun `clearError calls repository clearError`() = runTest {
        viewModel.clearError()
        advanceUntilIdle()

        coVerify { mockAuthRepository.clearError() }
    }

    @Test
    fun `auth state reflects repository state changes`() = runTest {
        // stateIn with WhileSubscribed needs an active collector to propagate upstream changes
        val collectJob = launch(UnconfinedTestDispatcher(testScheduler)) {
            viewModel.authState.collect {}
        }

        advanceUntilIdle()
        assertFalse(viewModel.authState.value.isAuthenticated)

        // Simulate login in repository
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            username = "admin",
            token = "token123"
        )
        advanceUntilIdle()

        assertTrue(viewModel.authState.value.isAuthenticated)
        assertEquals("admin", viewModel.authState.value.username)

        collectJob.cancel()
    }

    @Test
    fun `auth state reflects logout`() = runTest {
        val collectJob = launch(UnconfinedTestDispatcher(testScheduler)) {
            viewModel.authState.collect {}
        }

        // Start authenticated
        authStateFlow.value = AuthState(isAuthenticated = true, token = "token")
        advanceUntilIdle()
        assertTrue(viewModel.authState.value.isAuthenticated)

        // Simulate logout
        authStateFlow.value = AuthState.Unauthenticated
        advanceUntilIdle()
        assertFalse(viewModel.authState.value.isAuthenticated)

        collectJob.cancel()
    }

    @Test
    fun `auth state reflects error`() = runTest {
        val collectJob = launch(UnconfinedTestDispatcher(testScheduler)) {
            viewModel.authState.collect {}
        }

        authStateFlow.value = AuthState(error = "Invalid credentials")
        advanceUntilIdle()

        assertEquals("Invalid credentials", viewModel.authState.value.error)

        collectJob.cancel()
    }

    @Test
    fun `login exception does not crash`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException("Network error")

        // Should not throw
        viewModel.login("user", "pass")
        advanceUntilIdle()
    }
}
