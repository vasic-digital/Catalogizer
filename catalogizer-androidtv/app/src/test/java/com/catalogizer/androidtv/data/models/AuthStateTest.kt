package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class AuthStateTest {

    @Test
    fun `AuthState has correct defaults`() {
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
    fun `AuthState Unauthenticated is default state`() {
        val state = AuthState.Unauthenticated

        assertFalse(state.isAuthenticated)
        assertNull(state.username)
        assertNull(state.token)
    }

    @Test
    fun `AuthState authenticated state`() {
        val state = AuthState(
            isAuthenticated = true,
            username = "admin",
            token = "jwt-token-123",
            userId = 42L,
            expiresAt = System.currentTimeMillis() + 3600000
        )

        assertTrue(state.isAuthenticated)
        assertEquals("admin", state.username)
        assertEquals("jwt-token-123", state.token)
        assertEquals(42L, state.userId)
        assertNotNull(state.expiresAt)
    }

    @Test
    fun `AuthState error state`() {
        val state = AuthState(
            isAuthenticated = false,
            error = "Invalid credentials"
        )

        assertFalse(state.isAuthenticated)
        assertEquals("Invalid credentials", state.error)
    }

    @Test
    fun `AuthState loading state`() {
        val state = AuthState(isLoading = true)

        assertTrue(state.isLoading)
        assertFalse(state.isAuthenticated)
    }

    @Test
    fun `AuthState copy updates correctly`() {
        val initial = AuthState.Unauthenticated
        val loading = initial.copy(isLoading = true)
        val authenticated = loading.copy(
            isAuthenticated = true,
            isLoading = false,
            username = "user",
            token = "token"
        )
        val withError = initial.copy(error = "Failed")

        assertFalse(initial.isLoading)
        assertTrue(loading.isLoading)
        assertTrue(authenticated.isAuthenticated)
        assertEquals("user", authenticated.username)
        assertFalse(authenticated.isLoading)
        assertEquals("Failed", withError.error)
    }

    @Test
    fun `AuthState equality works correctly`() {
        val state1 = AuthState(isAuthenticated = true, username = "admin", token = "tok")
        val state2 = AuthState(isAuthenticated = true, username = "admin", token = "tok")
        val state3 = AuthState(isAuthenticated = true, username = "other", token = "tok")

        assertEquals(state1, state2)
        assertNotEquals(state1, state3)
    }
}
