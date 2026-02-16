package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import io.mockk.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class AuthInterceptorTest {

    private lateinit var mockWebServer: MockWebServer
    private lateinit var mockAuthRepository: AuthRepository
    private lateinit var interceptor: AuthInterceptor
    private lateinit var client: OkHttpClient
    private val authStateFlow = MutableStateFlow(AuthState.Unauthenticated)

    @Before
    fun setup() {
        mockWebServer = MockWebServer()
        mockWebServer.start()

        mockAuthRepository = mockk(relaxed = true)
        every { mockAuthRepository.authState } returns authStateFlow.asStateFlow()
        every { mockAuthRepository.shouldRefreshToken() } returns false

        interceptor = AuthInterceptor(mockAuthRepository)
        client = OkHttpClient.Builder()
            .addInterceptor(interceptor)
            .build()
    }

    @After
    fun tearDown() {
        mockWebServer.shutdown()
        clearAllMocks()
    }

    @Test
    fun `should add authorization header when authenticated`() {
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            token = "test-token-123",
            username = "testuser",
            userId = 1L
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .build()
        client.newCall(request).execute()

        val recordedRequest = mockWebServer.takeRequest()
        assertEquals("Bearer test-token-123", recordedRequest.getHeader("Authorization"))
    }

    @Test
    fun `should not add authorization header when not authenticated`() {
        authStateFlow.value = AuthState.Unauthenticated
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .build()
        client.newCall(request).execute()

        val recordedRequest = mockWebServer.takeRequest()
        assertNull(recordedRequest.getHeader("Authorization"))
    }

    @Test
    fun `should skip auth for login endpoint`() {
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            token = "test-token",
            username = "testuser",
            userId = 1L
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/auth/login"))
            .build()
        client.newCall(request).execute()

        val recordedRequest = mockWebServer.takeRequest()
        assertNull(recordedRequest.getHeader("Authorization"))
    }

    @Test
    fun `should not refresh token when not needed`() {
        every { mockAuthRepository.shouldRefreshToken() } returns false
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            token = "valid-token",
            username = "testuser",
            userId = 1L
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .build()
        client.newCall(request).execute()

        coVerify(exactly = 0) { mockAuthRepository.refreshToken() }
    }

    @Test
    fun `should refresh token when needed`() {
        every { mockAuthRepository.shouldRefreshToken() } returns true
        coEvery { mockAuthRepository.refreshToken() } just Runs
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            token = "refreshed-token",
            username = "testuser",
            userId = 1L
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .build()
        client.newCall(request).execute()

        coVerify { mockAuthRepository.refreshToken() }
    }

    @Test
    fun `should not add auth header when token is null`() {
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            token = null,
            username = "testuser",
            userId = 1L
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .build()
        client.newCall(request).execute()

        val recordedRequest = mockWebServer.takeRequest()
        assertNull(recordedRequest.getHeader("Authorization"))
    }

    @Test
    fun `should preserve original request headers`() {
        authStateFlow.value = AuthState(
            isAuthenticated = true,
            token = "test-token",
            username = "testuser",
            userId = 1L
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .addHeader("Content-Type", "application/json")
            .addHeader("Accept", "application/json")
            .build()
        client.newCall(request).execute()

        val recordedRequest = mockWebServer.takeRequest()
        assertEquals("application/json", recordedRequest.getHeader("Content-Type"))
        assertEquals("application/json", recordedRequest.getHeader("Accept"))
        assertEquals("Bearer test-token", recordedRequest.getHeader("Authorization"))
    }

    @Test
    fun `should proceed with request even if auth state is error`() {
        authStateFlow.value = AuthState(
            isAuthenticated = false,
            error = "Session expired"
        )
        mockWebServer.enqueue(MockResponse().setResponseCode(200))

        val request = Request.Builder()
            .url(mockWebServer.url("/api/v1/media"))
            .build()
        val response = client.newCall(request).execute()

        assertTrue(response.isSuccessful)
        val recordedRequest = mockWebServer.takeRequest()
        assertNull(recordedRequest.getHeader("Authorization"))
    }
}
