package com.catalogizer.androidtv.data.remote

import org.junit.Assert.*
import org.junit.Test

class LoginResponseTest2 {

    @Test
    fun `LoginUser holds correct data`() {
        val user = LoginUser(
            id = 1L,
            username = "admin",
            email = "admin@test.com",
            display_name = "Admin User"
        )
        assertEquals(1L, user.id)
        assertEquals("admin", user.username)
        assertEquals("admin@test.com", user.email)
        assertEquals("Admin User", user.display_name)
    }

    @Test
    fun `LoginUser with null optional fields`() {
        val user = LoginUser(id = 1L, username = "user")
        assertNull(user.email)
        assertNull(user.display_name)
    }

    @Test
    fun `LoginResponse token convenience property`() {
        val user = LoginUser(id = 1L, username = "admin")
        val response = LoginResponse(
            user = user,
            session_token = "jwt_token_123",
            refresh_token = "refresh_456",
            expires_at = "2025-12-31T23:59:59Z"
        )
        assertEquals("jwt_token_123", response.token)
        assertEquals("jwt_token_123", response.session_token)
    }

    @Test
    fun `LoginResponse userId convenience property`() {
        val user = LoginUser(id = 42L, username = "admin")
        val response = LoginResponse(
            user = user,
            session_token = "token"
        )
        assertEquals(42L, response.userId)
    }

    @Test
    fun `LoginResponse username convenience property`() {
        val user = LoginUser(id = 1L, username = "testuser")
        val response = LoginResponse(
            user = user,
            session_token = "token"
        )
        assertEquals("testuser", response.username)
    }

    @Test
    fun `LoginResponse expiresAt convenience property`() {
        val user = LoginUser(id = 1L, username = "admin")
        val response = LoginResponse(
            user = user,
            session_token = "token",
            expires_at = "2025-12-31T23:59:59Z"
        )
        assertEquals("2025-12-31T23:59:59Z", response.expiresAt)
    }

    @Test
    fun `LoginResponse with null optional fields`() {
        val user = LoginUser(id = 1L, username = "admin")
        val response = LoginResponse(
            user = user,
            session_token = "token"
        )
        assertNull(response.refresh_token)
        assertNull(response.expires_at)
        assertNull(response.expiresAt)
    }

    @Test
    fun `EntityStatsResponse holds correct data`() {
        val stats = EntityStatsResponse(
            totalEntities = 500,
            byType = mapOf("movie" to 200, "tv_show" to 150, "music" to 150)
        )
        assertEquals(500, stats.totalEntities)
        assertEquals(3, stats.byType.size)
        assertEquals(200, stats.byType["movie"])
    }

    @Test
    fun `EntityStatsResponse defaults`() {
        val stats = EntityStatsResponse()
        assertEquals(0, stats.totalEntities)
        assertTrue(stats.byType.isEmpty())
    }
}
