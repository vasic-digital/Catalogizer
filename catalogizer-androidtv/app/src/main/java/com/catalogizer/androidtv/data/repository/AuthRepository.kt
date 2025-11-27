package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.data.models.AuthState
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class AuthRepository(private val context: Context) {
    
    private val _authState = MutableStateFlow<AuthState>(AuthState.Unauthenticated)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()

    suspend fun login(username: String, password: String) {
        try {
            // TODO: Implement actual authentication logic
            // For now, simulate successful login for demo credentials
            if (username == "demo" && password == "demo") {
                _authState.value = AuthState(
                    isAuthenticated = true,
                    token = "demo-token-${System.currentTimeMillis()}",
                    username = username
                )
            } else {
                _authState.value = AuthState(
                    isAuthenticated = false,
                    error = "Invalid credentials. Use demo/demo for testing."
                )
            }
        } catch (e: Exception) {
            _authState.value = AuthState(
                isAuthenticated = false,
                error = "Login failed: ${e.message}"
            )
        }
    }

    suspend fun logout() {
        _authState.value = AuthState.Unauthenticated
    }

    suspend fun refreshToken() {
        // TODO: Implement token refresh logic
        // For now, just keep current authenticated state
        val current = _authState.value
        if (current.isAuthenticated) {
            _authState.value = current.copy(
                token = "refreshed-token-${System.currentTimeMillis()}"
            )
        }
    }

    suspend fun clearError() {
        val current = _authState.value
        if (current.error != null) {
            _authState.value = current.copy(error = null)
        }
    }
}