package com.catalogizer.android.data.models

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

/**
 * Tests for JSON serialization/deserialization of Auth models.
 *
 * These tests prevent regressions like the "java.lang.Class cannot be cast to
 * java.lang.reflect.ParameterizedType" error that occurs when Gson is used
 * with @Serializable (kotlinx.serialization) data classes.
 *
 * The models MUST use kotlinx.serialization (not Gson) because:
 * 1. All data classes are annotated with @Serializable
 * 2. Field mappings use @SerialName (not Gson's @SerializedName)
 * 3. The Retrofit converter must be kotlinx-serialization-converter (not GsonConverterFactory)
 */
class AuthSerializationTest {

    private lateinit var json: Json

    @Before
    fun setup() {
        json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
        }
    }

    // ---- LoginRequest ----

    @Test
    fun `LoginRequest serializes to correct JSON`() {
        val request = LoginRequest(username = "admin", password = "secret123")
        val jsonStr = json.encodeToString(request)

        assertTrue(jsonStr.contains("\"username\":\"admin\""))
        assertTrue(jsonStr.contains("\"password\":\"secret123\""))
    }

    @Test
    fun `LoginRequest deserializes from JSON`() {
        val jsonStr = """{"username":"admin","password":"secret123"}"""
        val request = json.decodeFromString<LoginRequest>(jsonStr)

        assertEquals("admin", request.username)
        assertEquals("secret123", request.password)
    }

    // ---- User ----

    @Test
    fun `User deserializes with snake_case field mapping`() {
        val jsonStr = """{
            "id": 1,
            "username": "admin",
            "email": "admin@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "role": "admin",
            "is_active": true,
            "last_login": "2026-01-01T00:00:00Z",
            "created_at": "2025-01-01T00:00:00Z",
            "updated_at": "2026-01-01T00:00:00Z",
            "permissions": ["read:media", "write:media"]
        }"""

        val user = json.decodeFromString<User>(jsonStr)

        assertEquals(1L, user.id)
        assertEquals("admin", user.username)
        assertEquals("admin@example.com", user.email)
        assertEquals("John", user.firstName)
        assertEquals("Doe", user.lastName)
        assertEquals("admin", user.role)
        assertTrue(user.isActive)
        assertEquals("2026-01-01T00:00:00Z", user.lastLogin)
        assertEquals("2025-01-01T00:00:00Z", user.createdAt)
        assertEquals("2026-01-01T00:00:00Z", user.updatedAt)
        assertEquals(listOf("read:media", "write:media"), user.permissions)
        assertEquals("John Doe", user.fullName)
        assertTrue(user.isAdmin)
    }

    @Test
    fun `User deserializes with optional fields missing`() {
        val jsonStr = """{
            "id": 2,
            "username": "viewer",
            "email": "viewer@example.com",
            "first_name": "Jane",
            "last_name": "Smith",
            "role": "user",
            "is_active": true,
            "created_at": "2025-06-01T00:00:00Z",
            "updated_at": "2025-06-01T00:00:00Z"
        }"""

        val user = json.decodeFromString<User>(jsonStr)

        assertEquals(2L, user.id)
        assertEquals("viewer", user.username)
        assertNull(user.lastLogin)
        assertNull(user.permissions)
        assertFalse(user.isAdmin)
    }

    @Test
    fun `User serializes with snake_case field names`() {
        val user = User(
            id = 1,
            username = "admin",
            email = "admin@test.com",
            firstName = "John",
            lastName = "Doe",
            role = "admin",
            isActive = true,
            createdAt = "2025-01-01",
            updatedAt = "2025-01-01"
        )

        val jsonStr = json.encodeToString(user)

        assertTrue(jsonStr.contains("\"first_name\":\"John\""))
        assertTrue(jsonStr.contains("\"last_name\":\"Doe\""))
        assertTrue(jsonStr.contains("\"is_active\":true"))
        assertTrue(jsonStr.contains("\"created_at\":\"2025-01-01\""))
        assertTrue(jsonStr.contains("\"updated_at\":\"2025-01-01\""))
    }

    // ---- LoginResponse ----

    @Test
    fun `LoginResponse deserializes complete API response`() {
        val jsonStr = """{
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
            "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
            "refresh_token": "refresh_abc123",
            "expires_in": 3600
        }"""

        val response = json.decodeFromString<LoginResponse>(jsonStr)

        assertEquals("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test", response.token)
        assertEquals("refresh_abc123", response.refreshToken)
        assertEquals(3600L, response.expiresIn)
        assertNotNull(response.user)
        assertEquals(1L, response.user.id)
        assertEquals("admin", response.user.username)
        assertEquals("John", response.user.firstName)
        assertEquals("Doe", response.user.lastName)
    }

    @Test
    fun `LoginResponse deserializes with unknown fields present`() {
        val jsonStr = """{
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
                "some_future_field": "ignored"
            },
            "token": "jwt_token",
            "refresh_token": "refresh_token",
            "expires_in": 7200,
            "session_id": "extra_field_ignored"
        }"""

        val response = json.decodeFromString<LoginResponse>(jsonStr)

        assertEquals("jwt_token", response.token)
        assertEquals("refresh_token", response.refreshToken)
        assertEquals(7200L, response.expiresIn)
        assertEquals("admin", response.user.username)
    }

    @Test
    fun `LoginResponse serializes correctly`() {
        val user = User(
            id = 1, username = "admin", email = "a@b.com",
            firstName = "A", lastName = "B", role = "admin",
            isActive = true, createdAt = "2025-01-01", updatedAt = "2025-01-01"
        )
        val response = LoginResponse(
            user = user, token = "tok", refreshToken = "ref", expiresIn = 3600
        )

        val jsonStr = json.encodeToString(response)

        assertTrue(jsonStr.contains("\"refresh_token\":\"ref\""))
        assertTrue(jsonStr.contains("\"expires_in\":3600"))
        assertTrue(jsonStr.contains("\"token\":\"tok\""))
    }

    // ---- RegisterRequest ----

    @Test
    fun `RegisterRequest serializes with snake_case`() {
        val request = RegisterRequest(
            username = "newuser",
            email = "new@test.com",
            password = "pass123",
            firstName = "New",
            lastName = "User"
        )

        val jsonStr = json.encodeToString(request)

        assertTrue(jsonStr.contains("\"first_name\":\"New\""))
        assertTrue(jsonStr.contains("\"last_name\":\"User\""))
    }

    // ---- AuthStatus ----

    @Test
    fun `AuthStatus deserializes authenticated response`() {
        val jsonStr = """{
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

        val status = json.decodeFromString<AuthStatus>(jsonStr)

        assertTrue(status.authenticated)
        assertNotNull(status.user)
        assertEquals("admin", status.user?.username)
        assertEquals(listOf("read:media", "write:media"), status.permissions)
        assertNull(status.error)
    }

    @Test
    fun `AuthStatus deserializes unauthenticated response`() {
        val jsonStr = """{"authenticated": false, "error": "Token expired"}"""

        val status = json.decodeFromString<AuthStatus>(jsonStr)

        assertFalse(status.authenticated)
        assertNull(status.user)
        assertEquals("Token expired", status.error)
    }

    // ---- ChangePasswordRequest ----

    @Test
    fun `ChangePasswordRequest serializes with snake_case`() {
        val request = ChangePasswordRequest(
            currentPassword = "old123",
            newPassword = "new456"
        )

        val jsonStr = json.encodeToString(request)

        assertTrue(jsonStr.contains("\"current_password\":\"old123\""))
        assertTrue(jsonStr.contains("\"new_password\":\"new456\""))
    }

    // ---- UpdateProfileRequest ----

    @Test
    fun `UpdateProfileRequest serializes with snake_case and nulls`() {
        val request = UpdateProfileRequest(firstName = "Updated", lastName = null, email = null)
        val jsonStr = json.encodeToString(request)

        assertTrue(jsonStr.contains("\"first_name\":\"Updated\""))
    }

    // ---- ErrorResponse ----

    @Test
    fun `ErrorResponse deserializes error from API`() {
        val jsonStr = """{
            "error": "Invalid credentials",
            "code": 401,
            "details": {"field": "password", "reason": "incorrect"}
        }"""

        val error = json.decodeFromString<ErrorResponse>(jsonStr)

        assertEquals("Invalid credentials", error.error)
        assertEquals(401, error.code)
        assertEquals("password", error.details?.get("field"))
        assertEquals("incorrect", error.details?.get("reason"))
    }

    @Test
    fun `ErrorResponse deserializes minimal error`() {
        val jsonStr = """{"error": "Server error"}"""

        val error = json.decodeFromString<ErrorResponse>(jsonStr)

        assertEquals("Server error", error.error)
        assertNull(error.code)
        assertNull(error.details)
    }

    // ---- Round-trip tests ----

    @Test
    fun `LoginRequest survives round-trip serialization`() {
        val original = LoginRequest("user", "pass")
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<LoginRequest>(serialized)

        assertEquals(original, deserialized)
    }

    @Test
    fun `User survives round-trip serialization`() {
        val original = User(
            id = 42, username = "test", email = "t@t.com",
            firstName = "Test", lastName = "User", role = "user",
            isActive = true, lastLogin = "2026-01-01",
            createdAt = "2025-01-01", updatedAt = "2026-01-01",
            permissions = listOf("read:media")
        )
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<User>(serialized)

        assertEquals(original, deserialized)
    }

    @Test
    fun `LoginResponse survives round-trip serialization`() {
        val user = User(
            id = 1, username = "admin", email = "a@b.com",
            firstName = "A", lastName = "B", role = "admin",
            isActive = true, createdAt = "2025-01-01", updatedAt = "2025-01-01"
        )
        val original = LoginResponse(
            user = user, token = "jwt_token",
            refreshToken = "refresh_token", expiresIn = 3600
        )
        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<LoginResponse>(serialized)

        assertEquals(original, deserialized)
    }
}
