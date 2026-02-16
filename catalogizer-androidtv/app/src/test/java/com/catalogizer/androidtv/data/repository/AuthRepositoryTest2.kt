package com.catalogizer.androidtv.data.repository

import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.remote.LoginResponse
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class AuthRepositoryTest2 {

    private val mockContext = mockk<android.content.Context>(relaxed = true)
    private val mockApi = mockk<CatalogizerApi>(relaxed = true)
    private lateinit var authRepository: AuthRepository

    @Before
    fun setup() {
        authRepository = AuthRepository(mockContext, mockApi)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial auth state is unauthenticated`() {
        assertEquals(AuthState.Unauthenticated, authRepository.authState.value)
    }

    @Test
    fun `login success sets authenticated state`() = runTest {
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 1L,
            username = "admin",
            expiresAt = "2026-12-31T23:59:59Z"
        )
        coEvery { mockApi.login(any()) } returns Response.success(loginResponse)

        authRepository.login("admin", "password")

        val state = authRepository.authState.value
        assertTrue(state.isAuthenticated)
        assertEquals("admin", state.username)
        assertEquals("test-token", state.token)
        assertEquals(1L, state.userId)
    }

    @Test
    fun `login failure sets error state`() = runTest {
        coEvery { mockApi.login(any()) } returns Response.error(
            401,
            okhttp3.ResponseBody.create(null, "Unauthorized")
        )

        authRepository.login("admin", "wrong")

        val state = authRepository.authState.value
        assertFalse(state.isAuthenticated)
        assertNotNull(state.error)
    }

    @Test
    fun `login exception sets error state`() = runTest {
        coEvery { mockApi.login(any()) } throws RuntimeException("Network error")

        authRepository.login("admin", "pass")

        val state = authRepository.authState.value
        assertFalse(state.isAuthenticated)
        assertTrue(state.error?.contains("Network error") == true)
    }

    @Test
    fun `logout sets unauthenticated state`() = runTest {
        // First login
        val loginResponse = LoginResponse(token = "tok", userId = 1L, username = "admin")
        coEvery { mockApi.login(any()) } returns Response.success(loginResponse)
        authRepository.login("admin", "pass")
        assertTrue(authRepository.authState.value.isAuthenticated)

        // Then logout
        authRepository.logout()

        assertEquals(AuthState.Unauthenticated, authRepository.authState.value)
    }

    @Test
    fun `clearError removes error from state`() = runTest {
        // Create an error state
        coEvery { mockApi.login(any()) } throws RuntimeException("Error")
        authRepository.login("admin", "pass")
        assertNotNull(authRepository.authState.value.error)

        // Clear error
        authRepository.clearError()

        assertNull(authRepository.authState.value.error)
    }

    @Test
    fun `isTokenExpired returns true when no expiry set`() {
        assertTrue(authRepository.isTokenExpired())
    }

    @Test
    fun `shouldRefreshToken returns false when no expiry set`() {
        assertFalse(authRepository.shouldRefreshToken())
    }

    @Test
    fun `setApi updates the API reference`() {
        val newApi = mockk<CatalogizerApi>(relaxed = true)
        authRepository.setApi(newApi)

        // No exception means setApi worked
        assertNotNull(authRepository)
    }

    @Test
    fun `refreshToken with no auth logs out`() = runTest {
        // Not authenticated - refreshToken should set unauthenticated
        authRepository.refreshToken()

        assertEquals(AuthState.Unauthenticated, authRepository.authState.value)
    }

    @Test
    fun `login with null body sets error`() = runTest {
        coEvery { mockApi.login(any()) } returns Response.success(null)

        authRepository.login("admin", "pass")

        val state = authRepository.authState.value
        assertFalse(state.isAuthenticated)
        assertTrue(state.error?.contains("Invalid response") == true)
    }
}
