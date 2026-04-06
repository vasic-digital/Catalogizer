package com.catalogizer.androidtv.data.repository

import android.content.Context
import android.util.Log
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone

private const val TAG = "AuthRepository"

/**
 * Manages authentication state for the Android TV app including login, logout,
 * token refresh with mutex-based concurrency, and token expiration checks.
 * Exposes [authState] as a [StateFlow] for reactive UI binding.
 */
class AuthRepository(private val context: Context, private var api: CatalogizerApi?) {

    fun setApi(api: CatalogizerApi) {
        this.api = api
    }
    
    private val refreshMutex = Mutex()
    
    private val _authState = MutableStateFlow<AuthState>(AuthState.Unauthenticated)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()

    suspend fun login(username: String, password: String) {
        try {
            val currentApi = api
            if (currentApi == null) {
                _authState.value = AuthState(
                    isAuthenticated = false,
                    error = "API not initialized. Please configure server URL."
                )
                return
            }
            
            val credentials = mapOf("username" to username, "password" to password)
            val response = currentApi.login(credentials)

            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    _authState.value = AuthState(
                        isAuthenticated = true,
                        token = body.token,
                        username = body.username,
                        userId = body.userId,
                        expiresAt = body.expiresAt?.let { parseExpiresAt(it) }
                    )
                } else {
                    _authState.value = AuthState(
                        isAuthenticated = false,
                        error = "Login failed: Invalid response from server"
                    )
                }
            } else {
                val errorMsg = try {
                    response.errorBody()?.string() ?: "Login failed: ${response.message()}"
                } catch (_: Exception) {
                    "Login failed: ${response.code()}"
                }
                _authState.value = AuthState(
                    isAuthenticated = false,
                    error = errorMsg
                )
            }
        } catch (e: Exception) {
            Log.e(TAG, "Login error: ${e.message}", e)
            _authState.value = AuthState(
                isAuthenticated = false,
                error = "Login failed: ${e.message ?: "Unknown error"}"
            )
        }
    }

    suspend fun logout() {
        try {
            val currentApi = api
            if (currentApi != null) {
                try {
                    currentApi.logout()
                } catch (e: Exception) {
                    Log.w(TAG, "Logout API call failed: ${e.message}")
                }
            }
        } catch (e: Exception) {
            Log.w(TAG, "Logout error: ${e.message}")
        } finally {
            _authState.value = AuthState.Unauthenticated
        }
    }

    suspend fun refreshToken() {
        refreshMutex.withLock {
            try {
                val currentApi = api
                if (currentApi == null) {
                    _authState.value = AuthState.Unauthenticated
                    return
                }
                
                val current = _authState.value
                if (current.isAuthenticated && current.token != null) {
                    val tokenBody = mapOf("token" to current.token)
                    val response = currentApi.refreshToken(tokenBody)

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
                Log.e(TAG, "Token refresh error: ${e.message}")
                // If refresh fails, logout user
                _authState.value = AuthState.Unauthenticated
            }
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
            Log.w(TAG, "Failed to parse expiresAt: $expiresAt")
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
