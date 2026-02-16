package com.catalogizer.android.data.remote

import com.catalogizer.android.data.models.LoginRequest
import com.catalogizer.android.data.models.LoginResponse
import com.catalogizer.android.data.models.User
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit

/**
 * Tests that Retrofit correctly deserializes API responses using kotlinx.serialization.
 *
 * This test class prevents the "java.lang.Class cannot be cast to
 * java.lang.reflect.ParameterizedType" error that occurs when GsonConverterFactory
 * is used with @Serializable data classes. The converter MUST be
 * retrofit2-kotlinx-serialization-converter.
 */
class RetrofitSerializationTest {

    private lateinit var mockWebServer: MockWebServer
    private lateinit var api: CatalogizerApi

    @Before
    fun setup() {
        mockWebServer = MockWebServer()
        mockWebServer.start()

        val json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
        }

        val contentType = "application/json".toMediaType()

        api = Retrofit.Builder()
            .baseUrl(mockWebServer.url("/"))
            .addConverterFactory(json.asConverterFactory(contentType))
            .build()
            .create(CatalogizerApi::class.java)
    }

    @After
    fun tearDown() {
        mockWebServer.shutdown()
    }

    @Test
    fun `login endpoint deserializes LoginResponse correctly`() = runBlocking {
        val responseJson = """{
            "user": {
                "id": 1,
                "username": "admin",
                "email": "admin@example.com",
                "first_name": "John",
                "last_name": "Doe",
                "role": "admin",
                "is_active": true,
                "created_at": "2025-01-01T00:00:00Z",
                "updated_at": "2026-01-01T00:00:00Z"
            },
            "token": "eyJhbGciOiJIUzI1NiJ9.test_jwt_token",
            "refresh_token": "refresh_abc123",
            "expires_in": 3600
        }"""

        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody(responseJson)
        )

        val response = api.login(LoginRequest("admin", "password"))

        assertTrue(response.isSuccessful)
        val body = response.body()
        assertNotNull(body)
        assertEquals("eyJhbGciOiJIUzI1NiJ9.test_jwt_token", body!!.token)
        assertEquals("refresh_abc123", body.refreshToken)
        assertEquals(3600L, body.expiresIn)
        assertEquals(1L, body.user.id)
        assertEquals("admin", body.user.username)
        assertEquals("John", body.user.firstName)
        assertEquals("Doe", body.user.lastName)

        // Verify the request was sent correctly
        val recordedRequest = mockWebServer.takeRequest()
        assertEquals("POST", recordedRequest.method)
        assertTrue(recordedRequest.path!!.contains("auth/login"))
        val requestBody = recordedRequest.body.readUtf8()
        assertTrue(requestBody.contains("\"username\":\"admin\""))
        assertTrue(requestBody.contains("\"password\":\"password\""))
    }

    @Test
    fun `login endpoint handles error response`() = runBlocking {
        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(401)
                .setHeader("Content-Type", "application/json")
                .setBody("""{"error": "Invalid credentials"}""")
        )

        val response = api.login(LoginRequest("wrong", "wrong"))

        assertFalse(response.isSuccessful)
        assertEquals(401, response.code())
    }

    @Test
    fun `login endpoint handles response with unknown fields`() = runBlocking {
        val responseJson = """{
            "user": {
                "id": 1,
                "username": "admin",
                "email": "admin@example.com",
                "first_name": "John",
                "last_name": "Doe",
                "role": "admin",
                "is_active": true,
                "created_at": "2025-01-01T00:00:00Z",
                "updated_at": "2026-01-01T00:00:00Z",
                "avatar_url": "https://example.com/avatar.png"
            },
            "token": "jwt_token",
            "refresh_token": "refresh_token",
            "expires_in": 7200,
            "session_id": "sess_123",
            "mfa_required": false
        }"""

        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody(responseJson)
        )

        val response = api.login(LoginRequest("admin", "pass"))

        assertTrue(response.isSuccessful)
        val body = response.body()!!
        assertEquals("jwt_token", body.token)
        assertEquals(7200L, body.expiresIn)
    }

    @Test
    fun `getAuthStatus deserializes correctly`() = runBlocking {
        val responseJson = """{
            "authenticated": true,
            "user": {
                "id": 1,
                "username": "admin",
                "email": "admin@test.com",
                "first_name": "A",
                "last_name": "B",
                "role": "admin",
                "is_active": true,
                "created_at": "2025-01-01",
                "updated_at": "2025-01-01"
            },
            "permissions": ["read:media", "write:media"]
        }"""

        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody(responseJson)
        )

        val response = api.getAuthStatus()

        assertTrue(response.isSuccessful)
        val body = response.body()!!
        assertTrue(body.authenticated)
        assertEquals("admin", body.user?.username)
        assertEquals(2, body.permissions?.size)
    }

    @Test
    fun `getProfile deserializes User correctly`() = runBlocking {
        val responseJson = """{
            "id": 42,
            "username": "testuser",
            "email": "test@example.com",
            "first_name": "Test",
            "last_name": "User",
            "role": "user",
            "is_active": true,
            "last_login": "2026-02-16T10:00:00Z",
            "created_at": "2025-06-01T00:00:00Z",
            "updated_at": "2026-02-16T10:00:00Z",
            "permissions": ["read:media"]
        }"""

        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody(responseJson)
        )

        val response = api.getProfile()

        assertTrue(response.isSuccessful)
        val user = response.body()!!
        assertEquals(42L, user.id)
        assertEquals("testuser", user.username)
        assertEquals("Test", user.firstName)
        assertEquals("User", user.lastName)
        assertEquals("Test User", user.fullName)
        assertFalse(user.isAdmin)
    }

    @Test
    fun `register endpoint serializes and deserializes correctly`() = runBlocking {
        val responseJson = """{
            "id": 5,
            "username": "newuser",
            "email": "new@example.com",
            "first_name": "New",
            "last_name": "User",
            "role": "user",
            "is_active": true,
            "created_at": "2026-02-16T12:00:00Z",
            "updated_at": "2026-02-16T12:00:00Z"
        }"""

        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(201)
                .setHeader("Content-Type", "application/json")
                .setBody(responseJson)
        )

        val response = api.register(
            com.catalogizer.android.data.models.RegisterRequest(
                username = "newuser",
                email = "new@example.com",
                password = "secure123",
                firstName = "New",
                lastName = "User"
            )
        )

        assertTrue(response.isSuccessful)
        val user = response.body()!!
        assertEquals("newuser", user.username)
        assertEquals("New", user.firstName)

        // Verify request body uses snake_case
        val requestBody = mockWebServer.takeRequest().body.readUtf8()
        assertTrue(requestBody.contains("\"first_name\":\"New\""))
        assertTrue(requestBody.contains("\"last_name\":\"User\""))
    }

    @Test
    fun `toApiResult extension handles successful response`() = runBlocking {
        val responseJson = """{
            "user": {
                "id": 1,
                "username": "admin",
                "email": "admin@test.com",
                "first_name": "A",
                "last_name": "B",
                "role": "admin",
                "is_active": true,
                "created_at": "2025-01-01",
                "updated_at": "2025-01-01"
            },
            "token": "token",
            "refresh_token": "refresh",
            "expires_in": 3600
        }"""

        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(200)
                .setHeader("Content-Type", "application/json")
                .setBody(responseJson)
        )

        val result = api.login(LoginRequest("admin", "pass")).toApiResult()

        assertTrue(result.isSuccess)
        assertNotNull(result.data)
        assertEquals("token", result.data!!.token)
        assertNull(result.error)
    }

    @Test
    fun `toApiResult extension handles error response`() = runBlocking {
        mockWebServer.enqueue(
            MockResponse()
                .setResponseCode(401)
                .setBody("""{"error":"Invalid credentials"}""")
        )

        val result = api.login(LoginRequest("bad", "bad")).toApiResult()

        assertFalse(result.isSuccess)
        assertNotNull(result.error)
        assertNull(result.data)
    }
}
