package com.catalogizer.androidtv.data.models

/**
 * Immutable state holder for authentication status, credentials, and error information.
 * [Unauthenticated] provides a default unauthenticated instance.
 */
data class AuthState(
    val isAuthenticated: Boolean = false,
    val username: String? = null,
    val token: String? = null,
    val userId: Long? = null,
    val expiresAt: Long? = null, // Unix timestamp in milliseconds
    val error: String? = null,
    val isLoading: Boolean = false
) {
    companion object {
        val Unauthenticated = AuthState()
    }
}