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
class AuthRepository(
    private val context: Context,
    private var api: CatalogizerApi?,
    private val tokenStore: com.catalogizer.androidtv.data.auth.TokenStore =
        com.catalogizer.androidtv.data.auth.TokenStore.safe(context),
) {

    fun setApi(api: CatalogizerApi) {
        this.api = api
    }

    private val refreshMutex = Mutex()

    private val _authState = MutableStateFlow<AuthState>(AuthState.Unauthenticated)
    val authState: StateFlow<AuthState> = _authState.asStateFlow()

    /**
     * Hydrates [_authState] from disk once at application start.
     * Called by MainViewModel before the Login vs Home decision
     * in MainActivity. Safe to call repeatedly — after the first
     * invocation subsequent calls are cheap no-ops because the
     * authState flow already carries the restored value.
     */
    suspend fun hydrateFromDisk() {
        val rec = tokenStore.load() ?: return
        if (rec.expiresAtMs != 0L && rec.expiresAtMs <= System.currentTimeMillis()) {
            tokenStore.clear()
            return
        }
        _authState.value = AuthState(
            isAuthenticated = true,
            token = rec.token,
            username = rec.username,
            userId = rec.userId,
            expiresAt = rec.expiresAtMs.takeIf { it > 0 },
        )
    }

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
                    val expiresAtMs = body.expiresAt?.let { parseExpiresAt(it) }
                    _authState.value = AuthState(
                        isAuthenticated = true,
                        token = body.token,
                        username = body.username,
                        userId = body.userId,
                        expiresAt = expiresAtMs
                    )
                    tokenStore.save(
                        com.catalogizer.androidtv.data.auth.TokenStore.Record(
                            token = body.token,
                            username = body.username,
                            userId = body.userId,
                            expiresAtMs = expiresAtMs ?: 0L,
                        )
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
            tokenStore.clear()
            _authState.value = AuthState.Unauthenticated
        }
    }

    suspend fun refreshToken() {
        refreshMutex.withLock {
            try {
                val currentApi = api
                if (currentApi == null) {
                    tokenStore.clear()
                    _authState.value = AuthState.Unauthenticated
                    return
                }

                val current = _authState.value
                if (current.isAuthenticated && current.token != null) {
                    val tokenBody = mapOf("token" to current.token)
                    val response = currentApi.refreshToken(tokenBody)

                    if (response.isSuccessful) {
                        response.body()?.let { loginResponse ->
                            val newExpiresMs = loginResponse.expiresAt?.let { parseExpiresAt(it) }
                            _authState.value = current.copy(
                                token = loginResponse.token,
                                expiresAt = newExpiresMs
                            )
                            tokenStore.save(
                                com.catalogizer.androidtv.data.auth.TokenStore.Record(
                                    token = loginResponse.token,
                                    username = loginResponse.username,
                                    userId = loginResponse.userId,
                                    expiresAtMs = newExpiresMs ?: 0L,
                                )
                            )
                        }
                    } else {
                        // Refresh rejected: wipe on-disk token and sign out.
                        tokenStore.clear()
                        _authState.value = AuthState.Unauthenticated
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Token refresh error: ${e.message}")
                tokenStore.clear()
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
        // The server emits RFC 3339 / ISO 8601 with either a
        // 'Z' suffix (UTC) or a numeric offset ("+03:00"), and
        // sub-second precision ranges from none to nanoseconds.
        // Use java.time.OffsetDateTime which accepts all of
        // these shapes; fall back to the legacy SimpleDateFormat
        // for the strictly-'Z' case as a belt-and-braces safety
        // net on very old devices.
        return try {
            java.time.OffsetDateTime.parse(expiresAt)
                .toInstant()
                .toEpochMilli()
        } catch (primary: Throwable) {
            try {
                val legacy = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.getDefault())
                legacy.timeZone = TimeZone.getTimeZone("UTC")
                legacy.parse(expiresAt)?.time
            } catch (fallback: Throwable) {
                Log.w(TAG, "Failed to parse expiresAt ($primary -> $fallback): $expiresAt")
                null
            }
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
