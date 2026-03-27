package com.catalogizer.androidtv.data.remote

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class LoginResponseTest {

    private lateinit var json: Json

    @Before
    fun setup() {
        json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            isLenient = true
        }
    }

    @Test
    fun `LoginResponse constructs correctly with new API shape`() {
        val user = LoginUser(id = 42L, username = "admin", email = "admin@test.com")
        val response = LoginResponse(
            user = user,
            session_token = "jwt-token-123",
            expires_at = "2026-01-01T00:00:00Z"
        )

        assertEquals("jwt-token-123", response.token)
        assertEquals(42L, response.userId)
        assertEquals("admin", response.username)
        assertEquals("2026-01-01T00:00:00Z", response.expiresAt)
        assertEquals("jwt-token-123", response.session_token)
    }

    @Test
    fun `LoginResponse constructs without expiresAt`() {
        val user = LoginUser(id = 42L, username = "admin")
        val response = LoginResponse(
            user = user,
            session_token = "jwt-token-123"
        )

        assertEquals("jwt-token-123", response.token)
        assertEquals(42L, response.userId)
        assertEquals("admin", response.username)
        assertNull(response.expiresAt)
    }

    @Test
    fun `LoginResponse convenience properties map correctly`() {
        val user = LoginUser(id = 1L, username = "testuser")
        val response = LoginResponse(
            user = user,
            session_token = "test-token",
            expires_at = "2026-12-31T23:59:59Z"
        )

        // Verify convenience properties
        assertEquals("test-token", response.token)
        assertEquals(1L, response.userId)
        assertEquals("testuser", response.username)
        assertEquals("2026-12-31T23:59:59Z", response.expiresAt)
    }

    @Test
    fun `LoginResponse round-trip equality`() {
        val user = LoginUser(id = 99L, username = "roundtrip")
        val original = LoginResponse(
            user = user,
            session_token = "roundtrip-token",
            expires_at = "2026-06-15T12:00:00Z"
        )

        val copy = original.copy()

        assertEquals(original, copy)
        assertEquals(original.token, copy.token)
        assertEquals(original.userId, copy.userId)
        assertEquals(original.username, copy.username)
        assertEquals(original.expiresAt, copy.expiresAt)
    }

    @Test
    fun `LoginUser serializes and deserializes correctly`() {
        val user = LoginUser(id = 1L, username = "user", email = "user@test.com", display_name = "Test User")

        val serialized = json.encodeToString(user)
        val deserialized = json.decodeFromString<LoginUser>(serialized)

        assertEquals(user, deserialized)
        assertEquals(1L, deserialized.id)
        assertEquals("user", deserialized.username)
        assertEquals("user@test.com", deserialized.email)
        assertEquals("Test User", deserialized.display_name)
    }

    @Test
    fun `LoginUser deserializes with unknown fields`() {
        val jsonStr = """{
            "id": 1,
            "username": "user",
            "unknown_field": "ignored",
            "another_unknown": 123
        }"""

        val user = json.decodeFromString<LoginUser>(jsonStr)

        assertEquals(1L, user.id)
        assertEquals("user", user.username)
        assertNull(user.email)
        assertNull(user.display_name)
    }

    @Test
    fun `LoginResponse optional fields default to null`() {
        val user = LoginUser(id = 5L, username = "minimal")
        val response = LoginResponse(
            user = user,
            session_token = "tok"
        )

        assertNull(response.refresh_token)
        assertNull(response.expires_at)
        assertNull(response.expiresAt)
    }
}
