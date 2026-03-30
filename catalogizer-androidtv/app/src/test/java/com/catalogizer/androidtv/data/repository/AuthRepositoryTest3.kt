package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.remote.LoginResponse
import com.catalogizer.androidtv.data.remote.LoginUser
import io.mockk.*
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response

@OptIn(ExperimentalCoroutinesApi::class)
class AuthRepositoryTest3 {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var mockContext: Context
    private lateinit var mockApi: CatalogizerApi
    private lateinit var repository: AuthRepository

    private val testUser = LoginUser(id = 1L, username = "admin", email = "admin@test.com")
    private val testLoginResponse = LoginResponse(
        user = testUser,
        session_token = "test_token_123",
        refresh_token = "refresh_token_456",
        expires_at = "2026-12-31T23:59:59Z"
    )

    @Before
    fun setup() {
        mockContext = mockk(relaxed = true)
        mockApi = mockk(relaxed = true)
        repository = AuthRepository(mockContext, mockApi)
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `initial auth state is unauthenticated`() {
        val state = repository.authState.value
        assertFalse(state.isAuthenticated)
        assertNull(state.token)
        assertNull(state.username)
    }

    @Test
    fun `login success updates auth state`() = runTest {
        coEvery { mockApi.login(any()) } returns Response.success(testLoginResponse)

        repository.login("admin", "password")

        val state = repository.authState.value
        assertTrue(state.isAuthenticated)
        assertEquals("test_token_123", state.token)
        assertEquals("admin", state.username)
        assertEquals(1L, state.userId)
    }

    @Test
    fun `login failure sets error state`() = runTest {
        coEvery { mockApi.login(any()) } returns Response.error(
            401, okhttp3.ResponseBody.create(null, "Unauthorized")
        )

        repository.login("bad", "creds")

        val state = repository.authState.value
        assertFalse(state.isAuthenticated)
        assertNotNull(state.error)
        assertTrue(state.error!!.contains("Login failed"))
    }

    @Test
    fun `login exception sets error state`() = runTest {
        coEvery { mockApi.login(any()) } throws RuntimeException("Network error")

        repository.login("user", "pass")

        val state = repository.authState.value
        assertFalse(state.isAuthenticated)
        assertTrue(state.error!!.contains("Network error"))
    }

    @Test
    fun `login with null API throws and sets error`() = runTest {
        val repoNoApi = AuthRepository(mockContext, null)

        repoNoApi.login("user", "pass")

        val state = repoNoApi.authState.value
        assertFalse(state.isAuthenticated)
        assertNotNull(state.error)
    }

    @Test
    fun `logout resets to unauthenticated`() = runTest {
        // First login
        coEvery { mockApi.login(any()) } returns Response.success(testLoginResponse)
        repository.login("admin", "pass")
        assertTrue(repository.authState.value.isAuthenticated)

        // Then logout
        repository.logout()

        val state = repository.authState.value
        assertFalse(state.isAuthenticated)
        assertNull(state.token)
    }

    @Test
    fun `clearError removes error from state`() = runTest {
        // Generate an error
        coEvery { mockApi.login(any()) } throws RuntimeException("Error")
        repository.login("u", "p")
        assertNotNull(repository.authState.value.error)

        // Clear it
        repository.clearError()
        assertNull(repository.authState.value.error)
    }

    @Test
    fun `clearError does nothing when no error`() = runTest {
        repository.clearError()
        assertNull(repository.authState.value.error)
    }

    @Test
    fun `isTokenExpired returns true when no expiry set`() {
        assertTrue(repository.isTokenExpired())
    }

    @Test
    fun `isTokenExpired returns true when token is expired`() = runTest {
        // Login with an already-expired token
        val expiredResponse = LoginResponse(
            user = testUser,
            session_token = "token",
            expires_at = "2020-01-01T00:00:00Z"
        )
        coEvery { mockApi.login(any()) } returns Response.success(expiredResponse)
        repository.login("u", "p")

        assertTrue(repository.isTokenExpired())
    }

    @Test
    fun `shouldRefreshToken returns false when not authenticated`() {
        assertFalse(repository.shouldRefreshToken())
    }

    @Test
    fun `setApi updates the API instance`() {
        val newApi = mockk<CatalogizerApi>(relaxed = true)
        repository.setApi(newApi)
        // No exception means success - the API was updated
    }

    @Test
    fun `login sends correct credentials`() = runTest {
        coEvery { mockApi.login(any()) } returns Response.success(testLoginResponse)

        repository.login("myuser", "mypass")

        coVerify {
            mockApi.login(mapOf("username" to "myuser", "password" to "mypass"))
        }
    }

    @Test
    fun `login with empty body response sets error`() = runTest {
        coEvery { mockApi.login(any()) } returns Response.success(null)

        repository.login("user", "pass")

        val state = repository.authState.value
        assertFalse(state.isAuthenticated)
        assertNotNull(state.error)
    }
}
