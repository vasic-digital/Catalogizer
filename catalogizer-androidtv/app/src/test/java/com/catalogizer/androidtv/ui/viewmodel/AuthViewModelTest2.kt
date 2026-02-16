package com.catalogizer.androidtv.ui.viewmodel

import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
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

@OptIn(ExperimentalCoroutinesApi::class)
class AuthViewModelTest2 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private val mockAuthRepository = mockk<AuthRepository>(relaxed = true)
    private val authStateFlow = MutableStateFlow(AuthState.Unauthenticated)

    @Before
    fun setup() {
        every { mockAuthRepository.authState } returns authStateFlow
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial auth state is unauthenticated`() = runTest {
        val viewModel = AuthViewModel(mockAuthRepository)
        advanceUntilIdle()

        assertFalse(viewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `login calls repository login`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } just Runs

        val viewModel = AuthViewModel(mockAuthRepository)
        viewModel.login("admin", "password")
        advanceUntilIdle()

        coVerify { mockAuthRepository.login("admin", "password") }
    }

    @Test
    fun `logout calls repository logout`() = runTest {
        coEvery { mockAuthRepository.logout() } just Runs

        val viewModel = AuthViewModel(mockAuthRepository)
        viewModel.logout()
        advanceUntilIdle()

        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `refreshToken calls repository refreshToken`() = runTest {
        coEvery { mockAuthRepository.refreshToken() } just Runs

        val viewModel = AuthViewModel(mockAuthRepository)
        viewModel.refreshToken()
        advanceUntilIdle()

        coVerify { mockAuthRepository.refreshToken() }
    }

    @Test
    fun `clearError calls repository clearError`() = runTest {
        coEvery { mockAuthRepository.clearError() } just Runs

        val viewModel = AuthViewModel(mockAuthRepository)
        viewModel.clearError()
        advanceUntilIdle()

        coVerify { mockAuthRepository.clearError() }
    }

    @Test
    fun `auth state reflects repository state changes`() = runTest {
        val viewModel = AuthViewModel(mockAuthRepository)

        // Start collecting to activate the WhileSubscribed upstream
        val collected = mutableListOf<AuthState>()
        val job = launch {
            viewModel.authState.collect { collected.add(it) }
        }
        advanceUntilIdle()

        // Simulate repository updating auth state
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            username = "admin",
            token = "tok"
        )
        advanceUntilIdle()

        assertTrue(viewModel.authState.value.isAuthenticated)
        assertEquals("admin", viewModel.authState.value.username)

        job.cancel()
    }

    @Test
    fun `login handles exception gracefully`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException("Network error")

        val viewModel = AuthViewModel(mockAuthRepository)
        // Should not throw
        viewModel.login("admin", "pass")
        advanceUntilIdle()

        // ViewModel catches the exception and the auth state remains
        assertNotNull(viewModel.authState)
    }
}
