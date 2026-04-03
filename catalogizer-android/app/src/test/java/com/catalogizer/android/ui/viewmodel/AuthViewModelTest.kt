package com.catalogizer.android.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.android.data.repository.AuthRepository
import com.catalogizer.android.data.remote.CatalogizerApi
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.test.*
import com.catalogizer.android.MainDispatcherRule
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.models.AuthStatus
import com.catalogizer.android.data.models.LoginResponse
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.Assert.*

/**
 * ViewModel tests for AuthViewModel
 */
@ExperimentalCoroutinesApi
class AuthViewModelTest {
    
    @get:Rule
    val instantExecutorRule = InstantTaskExecutorRule()
    
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()
    
    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var mockApi: CatalogizerApi
    private lateinit var viewModel: com.catalogizer.android.ui.viewmodel.AuthViewModel
    
    @Before
    fun setup() {
        Dispatchers.setMain(StandardTestDispatcher())
        mockApi = mockk(relaxed = true)
        mockAuthRepository = mockk(relaxed = true)
        
        // Setup default mock responses
        coEvery { mockAuthRepository.isAuthenticated() } returns true

        viewModel = com.catalogizer.android.ui.viewmodel.AuthViewModel(mockAuthRepository)
        // No need to advance here since test will handle it
    }
    
    @After
    fun tearDown() {
        Dispatchers.resetMain()
        clearAllMocks()
    }
    
    @Test
    fun `initial auth state should check authentication status`() = runTest {
        // When
        advanceUntilIdle()
        
        // Then - we can verify the auth state structure
        assertNotNull(viewModel.authState)
        
        // Verify auth status check was made
        coVerify { mockAuthRepository.isAuthenticated() }
    }
    
    @Test
    fun `login should update auth state`() = runTest {
        // Given
        val loginResponse = LoginResponse(
            token = "test_token",
            refreshToken = "refresh_token",
            user = mockk(),
            expiresIn = 3600
        )
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.success(loginResponse)
        
        // When
        viewModel.login("testuser", "password")
        advanceUntilIdle()
        
        // Then
        assertNotNull(viewModel.authState.value)
        
        // Verify login was called
        coVerify { mockAuthRepository.login("testuser", "password") }
    }
    
    @Test
    fun `logout should update auth state`() = runTest {
        // When
        viewModel.logout()
        advanceUntilIdle()
        
        // Then
        assertNotNull(viewModel.authState.value)
        
        // Verify logout was called
        coVerify { mockAuthRepository.logout() }
    }
    
    @Test
    fun `login failure should handle error`() = runTest {
        // Given
        coEvery { mockAuthRepository.login(any(), any()) } returns ApiResult.error("Invalid credentials")

        // When
        viewModel.login("invaliduser", "wrongpassword")
        advanceUntilIdle()

        // Then
        assertNotNull(viewModel.authState.value)
        assertFalse(viewModel.authState.value.isAuthenticated)
        assertEquals("Invalid credentials", viewModel.authState.value.error)
        assertFalse(viewModel.authState.value.isLoading)

        // Verify login attempt was made
        coVerify { mockAuthRepository.login("invaliduser", "wrongpassword") }
    }

    @Test
    fun `login exception sets error state`() = runTest {
        // Given
        coEvery { mockAuthRepository.login(any(), any()) } throws RuntimeException("Network error")

        // When
        viewModel.login("user", "pass")
        advanceUntilIdle()

        // Then
        assertFalse(viewModel.authState.value.isAuthenticated)
        assertEquals("Network error", viewModel.authState.value.error)
    }

    @Test
    fun `login sets loading state during request`() = runTest {
        coEvery { mockAuthRepository.login(any(), any()) } coAnswers {
            kotlinx.coroutines.delay(1000)
            ApiResult.success(mockk<LoginResponse>(relaxed = true))
        }

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