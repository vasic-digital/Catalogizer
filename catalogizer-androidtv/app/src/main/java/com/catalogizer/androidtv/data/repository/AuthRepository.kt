package com.catalogizer.androidtv.data.repository

import android.content.Context
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone

class AuthRepository(private val context: Context, private var api: CatalogizerApi?) {

    fun setApi(api: CatalogizerApi) {
        this.api = api
    }
    
    private val _authState = MutableStateFlow<AuthState>(AuthState.Unauthenticated)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()

    suspend fun login(username: String, password: String) {
        try {
            val credentials = mapOf("username" to username, "password" to password)
            val response = api?.login(credentials) ?: throw IllegalStateException("API not initialized")

            if (response.isSuccessful) {
                response.body()?.let { loginResponse ->
                    _authState.value = AuthState(
                        isAuthenticated = true,
                        token = loginResponse.token,
                        username = loginResponse.username,
                        userId = loginResponse.userId,
                        expiresAt = loginResponse.expiresAt?.let { parseExpiresAt(it) }
                    )
                } ?: run {
                    _authState.value = AuthState(
                        isAuthenticated = false,
                        error = "Login failed: Invalid response"
                    )
                }
            } else {
                _authState.value = AuthState(
                    isAuthenticated = false,
                    error = "Login failed: ${response.message()}"
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
        try {
            val current = _authState.value
            if (current.isAuthenticated && current.token != null) {
                val tokenBody = mapOf("token" to current.token)
                val response = api?.refreshToken(tokenBody) ?: throw IllegalStateException("API not initialized")

                if (response.isSuccessful) {
                    response.body()?.let { loginResponse ->
                        _authState.value = current.copy(
                            token = loginResponse.token,
                            expiresAt = loginResponse.expiresAt?.let { parseExpiresAt(it) }
                        )
                    }
                } else {
                    // If refresh fails, logout user
                    _authState.value = AuthState.Unauthenticated
                }
            }
        } catch (e: Exception) {
            // If refresh fails, logout user
            _authState.value = AuthState.Unauthenticated
        }
    }

    suspend fun clearError() {
        val current = _authState.value
        if (current.error != null) {
            _authState.value = current.copy(error = null)
        }
    }

    private fun parseExpiresAt(expiresAt: String): Long? {
        return try {
            val format = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.getDefault())
            format.timeZone = TimeZone.getTimeZone("UTC")
            format.parse(expiresAt)?.time
        } catch (e: Exception) {
            null
        }
    }

    fun isTokenExpired(): Boolean {
        val current = _authState.value
        return current.expiresAt?.let { System.currentTimeMillis() >= it } ?: true
    }

    fun shouldRefreshToken(): Boolean {
        val current = _authState.value
        return current.expiresAt?.let {
            // Refresh if token expires within 5 minutes
            System.currentTimeMillis() >= (it - 5 * 60 * 1000)
        } ?: false
    }
}