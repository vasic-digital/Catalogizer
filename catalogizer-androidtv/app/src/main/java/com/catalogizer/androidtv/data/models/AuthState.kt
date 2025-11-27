package com.catalogizer.androidtv.data.models

data class AuthState(
    val isAuthenticated: Boolean = false,
    val username: String? = null,
    val token: String? = null,
    val error: String? = null,
    val isLoading: Boolean = false
) {
    companion object {
        val Unauthenticated = AuthState()
    }
}