package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class AuthStateTest2 {

    @Test
    fun `default AuthState is unauthenticated`() {
        val state = AuthState()
        assertFalse(state.isAuthenticated)
        assertNull(state.username)
        assertNull(state.token)
        assertNull(state.userId)
        assertNull(state.expiresAt)
        assertNull(state.error)
        assertFalse(state.isLoading)
    }

    @Test
    fun `Unauthenticated companion matches default`() {
        val state = AuthState.Unauthenticated
        assertFalse(state.isAuthenticated)
        assertNull(state.token)
    }

    @Test
    fun `authenticated state with all fields`() {
        val state = AuthState(
            isAuthenticated = true,
            username = "admin",
            token = "jwt_token_123",
            userId = 42L,
            expiresAt = 1735689600000L,
            error = null,
            isLoading = false
        )
        assertTrue(state.isAuthenticated)
        assertEquals("admin", state.username)
        assertEquals("jwt_token_123", state.token)
        assertEquals(42L, state.userId)
        assertEquals(1735689600000L, state.expiresAt)
        assertNull(state.error)
        assertFalse(state.isLoading)
    }

    @Test
    fun `error state`() {
        val state = AuthState(
            isAuthenticated = false,
            error = "Invalid credentials"
        )
        assertFalse(state.isAuthenticated)
        assertEquals("Invalid credentials", state.error)
    }

    @Test
    fun `loading state`() {
        val state = AuthState(isLoading = true)
        assertTrue(state.isLoading)
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `copy preserves unchanged fields`() {
        val original = AuthState(
            isAuthenticated = true,
            username = "admin",
            token = "token123",
            userId = 1L
        )
        val updated = original.copy(error = "Session expired")
        assertTrue(updated.isAuthenticated)
        assertEquals("admin", updated.username)
        assertEquals("token123", updated.token)
        assertEquals("Session expired", updated.error)
    }

    @Test
    fun `copy can transition from authenticated to unauthenticated`() {
        val authenticated = AuthState(
            isAuthenticated = true,
            username = "user",
            token = "token"
        )
        val loggedOut = authenticated.copy(
            isAuthenticated = false,
            token = null,
            username = null
        )
        assertFalse(loggedOut.isAuthenticated)
        assertNull(loggedOut.token)
        assertNull(loggedOut.username)
    }

    @Test
    fun `equality check`() {
        val state1 = AuthState(isAuthenticated = true, username = "admin", token = "t1")
        val state2 = AuthState(isAuthenticated = true, username = "admin", token = "t1")
        assertEquals(state1, state2)
    }

    @Test
    fun `inequality check`() {
        val state1 = AuthState(isAuthenticated = true, token = "t1")
        val state2 = AuthState(isAuthenticated = true, token = "t2")
        assertNotEquals(state1, state2)
    }
}
