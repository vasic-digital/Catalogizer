package com.catalogizer.android.ui.viewmodel

import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.models.AuthState
import com.catalogizer.android.data.models.LoginResponse
import com.catalogizer.android.data.models.User
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.repository.AuthRepository
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
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
    private lateinit var viewModel: AuthViewModel

    private val testUser = User(
        id = 1L,
        username = "admin",
        email = "admin@test.com",
        firstName = "Admin",
        lastName = "User",
        role = "admin",
        isActive = true,
        createdAt = "2025-01-01",
        updatedAt = "2025-01-01"
    )

    @Before
    fun setup() {
        mockAuthRepository = mockk(relaxed = true)
        coEvery { mockAuthRepository.isAuthenticated() } returns false
        viewModel = AuthViewModel(mockAuthRepository)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial auth state defaults to not authenticated`() = runTest {
        advanceUntilIdle()
        assertFalse(viewModel.authState.value.isAuthenticated)
    }

    @Test
    fun `init checks authentication status`() = runTest {
        advanceUntilIdle()
        coVerify { mockAuthRepository.isAuthenticated() }
    }

    @Test
    fun `init sets authenticated true when repository says authenticated`() = runTest {
        clearAllMocks()
        val repo = mockk<AuthRepository>(relaxed = true)
        coEvery { repo.isAuthenticated() } returns true
        val vm = AuthViewModel(repo)
        advanceUntilIdle()

        assertTrue(vm.authState.value.isAuthenticated)
    }

    @Test
    fun `login success sets authenticated state`() = runTest {
        val loginResponse = LoginResponse(
            token = "token123",
            refreshToken = "refresh123",
            user = testUser,
            expiresIn = 3600
        )
        coEvery { mockAuthRepository.login("admin", "pass") } returns ApiResult.success(loginResponse)

        viewModel.login("admin", "pass")
        advanceUntilIdle()

        assertTrue(viewModel.authState.value.isAuthenticated)
        assertNull(viewModel.authState.value.error)
        assertFalse(viewModel.authState.value.isLoading)
    }

    @Test
    fun `login failure sets error state`() = runTest {
        coEvery { mockAuthRepository.login("bad", "creds") } returns ApiResult.error("Invalid credentials")

        viewModel.login("bad", "creds")
        advanceUntilIdle()

        assertFalse(viewModel.authState.value.isAuthenticated)
        assertEquals("Invalid credentials", viewModel.authState.value.error)
        assertFalse(viewModel.authState.value.isLoading)
    }

    @Test
    fun `login exception sets error state`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException("Network error")

        viewModel.login("user", "pass")
        advanceUntilIdle()

        assertFalse(viewModel.authState.value.isAuthenticated)
        assertEquals("Network error", viewModel.authState.value.error)
        assertFalse(viewModel.authState.value.isLoading)
    }

    @Test
    fun `logout resets auth state`() = runTest {
        // First login
        val loginResponse = LoginResponse(
            token = "token", refreshToken = "refresh", user = testUser, expiresIn = 3600
        )
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(loginResponse)
        viewModel.login("admin", "pass")
        advanceUntilIdle()
        assertTrue(viewModel.authState.value.isAuthenticated)

        // Then logout
        viewModel.logout()
        advanceUntilIdle()

        assertFalse(viewModel.authState.value.isAuthenticated)
        coVerify { mockAuthRepository.logout() }
    }

    @Test
    fun `login sets loading state during request`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } coAnswers {
            // Simulate delay in response
            ApiResult.success(
                LoginResponse(token = "t", refreshToken = "r", user = testUser, expiresIn = 3600)
            )
        }

        // Before login
        advanceUntilIdle()
        assertFalse(viewModel.authState.value.isLoading)

        viewModel.login("user", "pass")
        advanceUntilIdle()

        // After completion, loading should be false
        assertFalse(viewModel.authState.value.isLoading)
    }

    @Test
    fun `authState is a StateFlow`() {
        val state = viewModel.authState.value
        assertNotNull(state)
        assertFalse(state.isLoading)
    }
}
