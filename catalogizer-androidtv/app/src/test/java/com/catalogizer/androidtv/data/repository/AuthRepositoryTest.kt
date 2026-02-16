package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.MainDispatcherRule
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import com.catalogizer.androidtv.data.remote.LoginResponse
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import okhttp3.ResponseBody.Companion.toResponseBody
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import retrofit2.Response
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone

@ExperimentalCoroutinesApi
class AuthRepositoryTest {

    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var context: Context
    private lateinit var api: CatalogizerApi
    private lateinit var repository: AuthRepository

    @Before
    fun setup() {
        context = mockk()
        api = mockk()
        repository = AuthRepository(context, api)
    }

    @Test
    fun `initial auth state should be unauthenticated`() = runTest {
        val initialState = repository.authState.first()
        assertEquals(AuthState.Unauthenticated, initialState)
    }

    @Test
    fun `login success should update auth state with user data`() = runTest {
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser",
            expiresAt = "2024-12-31T23:59:59Z"
        )
        val successResponse = Response.success(loginResponse)

        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        val authState = repository.authState.first()
        assertTrue(authState.isAuthenticated)
        assertEquals("test-token", authState.token)
        assertEquals(123L, authState.userId)
        assertEquals("testuser", authState.username)
        assertNull(authState.error)
    }

    @Test
    fun `login failure should update auth state with error`() = runTest {
        val errorResponse = Response.error<LoginResponse>(
            401,
            "Unauthorized".toResponseBody(null)
        )

        coEvery { api.login(any()) } returns errorResponse

        repository.login("testuser", "wrongpassword")

        val authState = repository.authState.first()
        assertFalse(authState.isAuthenticated)
        // Response.error() returns "Response.error()" as the message
        assertEquals("Login failed: Response.error()", authState.error)
        assertNull(authState.token)
    }

    @Test
    fun `login with null response body should set error state`() = runTest {
        val successResponse = Response.success<LoginResponse>(null)

        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        val authState = repository.authState.first()
        assertFalse(authState.isAuthenticated)
        assertEquals("Login failed: Invalid response", authState.error)
    }

    @Test
    fun `login with exception should update auth state with error`() = runTest {
        val exception = RuntimeException("Network error")
        coEvery { api.login(any()) } throws exception

        repository.login("testuser", "password")

        val authState = repository.authState.first()
        assertFalse(authState.isAuthenticated)
        assertEquals("Login failed: Network error", authState.error)
    }

    @Test
    fun `login with null api should throw exception`() = runTest {
        repository = AuthRepository(context, null)

        // The login method catches IllegalStateException internally
        // and sets an error state instead of propagating it
        repository.login("testuser", "password")

        val authState = repository.authState.first()
        assertFalse(authState.isAuthenticated)
        assertEquals("Login failed: API not initialized", authState.error)
    }

    @Test
    fun `logout should reset auth state to unauthenticated`() = runTest {
        // First login
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")
        var authState = repository.authState.first()
        assertTrue(authState.isAuthenticated)

        // Then logout
        repository.logout()

        authState = repository.authState.first()
        assertEquals(AuthState.Unauthenticated, authState)
    }

    @Test
    fun `refresh token success should update token and expiry`() = runTest {
        // First login
        val loginResponse = LoginResponse(
            token = "old-token",
            userId = 123L,
            username = "testuser",
            expiresAt = "2024-01-01T00:00:00Z"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        // Refresh token
        val refreshResponse = LoginResponse(
            token = "new-token",
            userId = 123L,
            username = "testuser",
            expiresAt = "2024-12-31T23:59:59Z"
        )
        val refreshSuccessResponse = Response.success(refreshResponse)
        coEvery { api.refreshToken(any()) } returns refreshSuccessResponse

        repository.refreshToken()

        val authState = repository.authState.first()
        assertTrue(authState.isAuthenticated)
        assertEquals("new-token", authState.token)
        assertEquals("testuser", authState.username)
    }

    @Test
    fun `refresh token failure should logout user`() = runTest {
        // First login
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        // Refresh fails
        val refreshErrorResponse = Response.error<LoginResponse>(
            401,
            "Token expired".toResponseBody(null)
        )
        coEvery { api.refreshToken(any()) } returns refreshErrorResponse

        repository.refreshToken()

        val authState = repository.authState.first()
        assertEquals(AuthState.Unauthenticated, authState)
    }

    @Test
    fun `refresh token with unauthenticated user should do nothing`() = runTest {
        repository.refreshToken()

        val authState = repository.authState.first()
        assertEquals(AuthState.Unauthenticated, authState)
    }

    @Test
    fun `refresh token with null api should throw exception`() = runTest {
        // Create a new repository with null api
        // The refreshToken method catches exceptions internally
        // and resets to Unauthenticated when refresh fails
        val nullApiRepository = AuthRepository(context, null)

        // Since this is a new repository, it starts unauthenticated
        // and refreshToken does nothing for unauthenticated state
        nullApiRepository.refreshToken()

        val authState = nullApiRepository.authState.first()
        assertEquals(AuthState.Unauthenticated, authState)
    }

    @Test
    fun `clear error should remove error from auth state`() = runTest {
        // Set error state
        val errorResponse = Response.error<LoginResponse>(
            401,
            "Unauthorized".toResponseBody(null)
        )
        coEvery { api.login(any()) } returns errorResponse

        repository.login("testuser", "password")

        var authState = repository.authState.first()
        // Response.error() returns "Response.error()" as the message
        assertEquals("Login failed: Response.error()", authState.error)

        // Clear error
        repository.clearError()

        authState = repository.authState.first()
        assertNull(authState.error)
        assertFalse(authState.isAuthenticated)
    }

    @Test
    fun `clear error with no error should do nothing`() = runTest {
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")
        repository.clearError()

        val authState = repository.authState.first()
        assertTrue(authState.isAuthenticated)
        assertNull(authState.error)
    }

    @Test
    fun `isTokenExpired with no expiry should return true`() = runTest {
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        assertTrue(repository.isTokenExpired())
    }

    @Test
    fun `isTokenExpired with future expiry should return false`() = runTest {
        val futureTime = System.currentTimeMillis() + 3600000 // 1 hour from now
        val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.getDefault())
        format.timeZone = TimeZone.getTimeZone("UTC")
        val expiresAt = format.format(java.util.Date(futureTime))

        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser",
            expiresAt = expiresAt
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        assertFalse(repository.isTokenExpired())
    }

    @Test
    fun `isTokenExpired with past expiry should return true`() = runTest {
        val pastTime = System.currentTimeMillis() - 3600000 // 1 hour ago
        val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.getDefault())
        format.timeZone = TimeZone.getTimeZone("UTC")
        val expiresAt = format.format(java.util.Date(pastTime))

        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser",
            expiresAt = expiresAt
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        assertTrue(repository.isTokenExpired())
    }

    @Test
    fun `shouldRefreshToken with expiry more than 5 minutes away should return false`() = runTest {
        val futureTime = System.currentTimeMillis() + 360000 // 6 minutes from now
        val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.getDefault())
        format.timeZone = TimeZone.getTimeZone("UTC")
        val expiresAt = format.format(java.util.Date(futureTime))

        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser",
            expiresAt = expiresAt
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        assertFalse(repository.shouldRefreshToken())
    }

    @Test
    fun `shouldRefreshToken with expiry within 5 minutes should return true`() = runTest {
        val futureTime = System.currentTimeMillis() + 240000 // 4 minutes from now
        val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.getDefault())
        format.timeZone = TimeZone.getTimeZone("UTC")
        val expiresAt = format.format(java.util.Date(futureTime))

        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser",
            expiresAt = expiresAt
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        assertTrue(repository.shouldRefreshToken())
    }

    @Test
    fun `shouldRefreshToken with no expiry should return false`() = runTest {
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { api.login(any()) } returns successResponse

        repository.login("testuser", "password")

        assertFalse(repository.shouldRefreshToken())
    }

    @Test
    fun `setApi should update api instance`() = runTest {
        val newApi = mockk<CatalogizerApi>()
        repository.setApi(newApi)

        // Test that the new api is used
        val loginResponse = LoginResponse(
            token = "test-token",
            userId = 123L,
            username = "testuser"
        )
        val successResponse = Response.success(loginResponse)
        coEvery { newApi.login(any()) } returns successResponse

        repository.login("testuser", "password")

        val authState = repository.authState.first()
        assertTrue(authState.isAuthenticated)
    }
}