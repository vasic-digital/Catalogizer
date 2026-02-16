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
    fun `LoginResponse deserializes correctly`() {
        val jsonStr = """{
            "token": "jwt-token-123",
            "user_id": 42,
            "username": "admin",
            "expires_at": "2026-01-01T00:00:00Z"
        }"""

        val response = json.decodeFromString<LoginResponse>(jsonStr)

        assertEquals("jwt-token-123", response.token)
        assertEquals(42L, response.userId)
        assertEquals("admin", response.username)
        assertEquals("2026-01-01T00:00:00Z", response.expiresAt)
    }

    @Test
    fun `LoginResponse deserializes without expiresAt`() {
        val jsonStr = """{
            "token": "jwt-token-123",
            "user_id": 42,
            "username": "admin"
        }"""

        val response = json.decodeFromString<LoginResponse>(jsonStr)

        assertEquals("jwt-token-123", response.token)
        assertEquals(42L, response.userId)
        assertEquals("admin", response.username)
        assertNull(response.expiresAt)
    }

    @Test
    fun `LoginResponse serializes correctly`() {
        val response = LoginResponse(
            token = "test-token",
            userId = 1L,
            username = "testuser",
            expiresAt = "2026-12-31T23:59:59Z"
        )

        val jsonStr = json.encodeToString(response)

        assertTrue(jsonStr.contains("\"token\":\"test-token\""))
        assertTrue(jsonStr.contains("\"user_id\":1"))
        assertTrue(jsonStr.contains("\"username\":\"testuser\""))
        assertTrue(jsonStr.contains("\"expires_at\":\"2026-12-31T23:59:59Z\""))
    }

    @Test
    fun `LoginResponse round-trip serialization`() {
        val original = LoginResponse(
            token = "roundtrip-token",
            userId = 99L,
            username = "roundtrip",
            expiresAt = "2026-06-15T12:00:00Z"
        )

        val serialized = json.encodeToString(original)
        val deserialized = json.decodeFromString<LoginResponse>(serialized)

        assertEquals(original, deserialized)
    }

    @Test
    fun `LoginResponse with unknown fields ignores them`() {
        val jsonStr = """{
            "token": "jwt-token",
            "user_id": 1,
            "username": "user",
            "unknown_field": "ignored",
            "another_unknown": 123
        }"""

        val response = json.decodeFromString<LoginResponse>(jsonStr)

        assertEquals("jwt-token", response.token)
        assertEquals(1L, response.userId)
        assertEquals("user", response.username)
    }
}
