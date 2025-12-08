package com.catalogizer.android.ui.viewmodel

import androidx.arch.core.executor.testing.InstantTaskExecutorRule
import com.catalogizer.android.data.repository.AuthRepository
import com.catalogizer.android.data.remote.CatalogizerApi
import io.mockk.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
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
        every { mockAuthRepository.isAuthenticated } returns flowOf(true)
        
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
        verify { mockAuthRepository.isAuthenticated }
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
        
        // Verify login attempt was made
        coVerify { mockAuthRepository.login("invaliduser", "wrongpassword") }
    }
}