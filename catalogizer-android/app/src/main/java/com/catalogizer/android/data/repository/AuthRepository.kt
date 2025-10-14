package com.catalogizer.android.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import com.catalogizer.android.data.models.*
import com.catalogizer.android.data.remote.CatalogizerApi
import com.catalogizer.android.data.remote.ApiResult
import com.catalogizer.android.data.remote.toApiResult
import kotlinx.coroutines.flow.*
import kotlinx.serialization.json.Json
import kotlinx.serialization.encodeToString
import kotlinx.serialization.decodeFromString
class AuthRepository(
    private val api: CatalogizerApi,
    private val dataStore: DataStore<Preferences>
) {

    private val json = Json { ignoreUnknownKeys = true }

    companion object {
        private val AUTH_TOKEN_KEY = stringPreferencesKey("auth_token")
        private val REFRESH_TOKEN_KEY = stringPreferencesKey("refresh_token")
        private val USER_KEY = stringPreferencesKey("user_data")
        private val TOKEN_EXPIRY_KEY = longPreferencesKey("token_expiry")
        private val SERVER_URL_KEY = stringPreferencesKey("server_url")
        private val REMEMBER_ME_KEY = booleanPreferencesKey("remember_me")
        private val BIOMETRIC_ENABLED_KEY = booleanPreferencesKey("biometric_enabled")
    }

    // Authentication state flows
    val authToken: Flow<String?> = dataStore.data.map { preferences ->
        preferences[AUTH_TOKEN_KEY]
    }

    val refreshToken: Flow<String?> = dataStore.data.map { preferences ->
        preferences[REFRESH_TOKEN_KEY]
    }

    val currentUser: Flow<User?> = dataStore.data.map { preferences ->
        preferences[USER_KEY]?.let { userJson ->
            try {
                json.decodeFromString<User>(userJson)
            } catch (e: Exception) {
                null
            }
        }
    }

    val isAuthenticated: Flow<Boolean> = combine(
        authToken,
        dataStore.data.map { it[TOKEN_EXPIRY_KEY] ?: 0L }
    ) { token, expiry ->
        !token.isNullOrBlank() && System.currentTimeMillis() < expiry
    }

    val serverUrl: Flow<String?> = dataStore.data.map { preferences ->
        preferences[SERVER_URL_KEY]
    }

    val rememberMe: Flow<Boolean> = dataStore.data.map { preferences ->
        preferences[REMEMBER_ME_KEY] ?: false
    }

    val biometricEnabled: Flow<Boolean> = dataStore.data.map { preferences ->
        preferences[BIOMETRIC_ENABLED_KEY] ?: false
    }

    // Authentication operations
    suspend fun login(
        username: String,
        password: String,
        rememberMe: Boolean = false
    ): ApiResult<LoginResponse> {
        return try {
            val loginRequest = LoginRequest(username, password)
            val result = api.login(loginRequest).toApiResult()

            if (result.isSuccess && result.data != null) {
                val response = result.data
                saveAuthData(response, rememberMe)
            }

            result
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Login failed")
        }
    }

    suspend fun register(
        username: String,
        email: String,
        password: String,
        firstName: String,
        lastName: String
    ): ApiResult<User> {
        return try {
            val registerRequest = RegisterRequest(username, email, password, firstName, lastName)
            api.register(registerRequest).toApiResult()
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Registration failed")
        }
    }

    suspend fun logout(): ApiResult<Unit> {
        return try {
            // Call API logout if possible
            val result = try {
                api.logout().toApiResult()
            } catch (e: Exception) {
                // Continue with local logout even if API call fails
                ApiResult.success(Unit)
            }

            // Clear local auth data
            clearAuthData()
            result
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Logout failed")
        }
    }

    suspend fun refreshAuthToken(): ApiResult<String> {
        return try {
            val currentRefreshToken = refreshToken.first()
            if (currentRefreshToken.isNullOrBlank()) {
                return ApiResult.error("No refresh token available")
            }

            // Note: Implement refresh token endpoint in API
            // For now, this is a placeholder
            val result = api.getAuthStatus().toApiResult()
            if (result.isSuccess) {
                ApiResult.success(authToken.first() ?: "")
            } else {
                ApiResult.error("Token refresh failed")
            }
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Token refresh failed")
        }
    }

    suspend fun getProfile(): ApiResult<User> {
        return try {
            val result = api.getProfile().toApiResult()
            if (result.isSuccess && result.data != null) {
                saveUser(result.data)
            }
            result
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Failed to get profile")
        }
    }

    suspend fun updateProfile(updateRequest: UpdateProfileRequest): ApiResult<User> {
        return try {
            val result = api.updateProfile(updateRequest).toApiResult()
            if (result.isSuccess && result.data != null) {
                saveUser(result.data)
            }
            result
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Failed to update profile")
        }
    }

    suspend fun changePassword(
        currentPassword: String,
        newPassword: String
    ): ApiResult<Unit> {
        return try {
            val request = ChangePasswordRequest(currentPassword, newPassword)
            api.changePassword(request).toApiResult()
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Failed to change password")
        }
    }

    suspend fun checkAuthStatus(): ApiResult<AuthStatus> {
        return try {
            api.getAuthStatus().toApiResult()
        } catch (e: Exception) {
            ApiResult.error(e.message ?: "Failed to check auth status")
        }
    }

    // Settings operations
    suspend fun setServerUrl(url: String) {
        dataStore.edit { preferences ->
            preferences[SERVER_URL_KEY] = url
        }
    }

    suspend fun setBiometricEnabled(enabled: Boolean) {
        dataStore.edit { preferences ->
            preferences[BIOMETRIC_ENABLED_KEY] = enabled
        }
    }

    suspend fun setRememberMe(remember: Boolean) {
        dataStore.edit { preferences ->
            preferences[REMEMBER_ME_KEY] = remember
        }
    }

    // Permission checking
    suspend fun hasPermission(permission: String): Boolean {
        val user = currentUser.first()
        return user?.let {
            it.isAdmin || it.permissions?.contains(permission) == true
        } ?: false
    }

    suspend fun canAccess(resource: String, action: String): Boolean {
        val permission = "$action:$resource"
        return hasPermission(permission) || hasPermission("admin:system")
    }

    // Private helper methods
    private suspend fun saveAuthData(loginResponse: LoginResponse, rememberMe: Boolean) {
        val expiryTime = System.currentTimeMillis() + (loginResponse.expiresIn * 1000)

        dataStore.edit { preferences ->
            preferences[AUTH_TOKEN_KEY] = loginResponse.token
            preferences[REFRESH_TOKEN_KEY] = loginResponse.refreshToken
            preferences[USER_KEY] = json.encodeToString(loginResponse.user)
            preferences[TOKEN_EXPIRY_KEY] = expiryTime
            preferences[REMEMBER_ME_KEY] = rememberMe
        }
    }

    private suspend fun saveUser(user: User) {
        dataStore.edit { preferences ->
            preferences[USER_KEY] = json.encodeToString(user)
        }
    }

    private suspend fun clearAuthData() {
        dataStore.edit { preferences ->
            preferences.remove(AUTH_TOKEN_KEY)
            preferences.remove(REFRESH_TOKEN_KEY)
            preferences.remove(USER_KEY)
            preferences.remove(TOKEN_EXPIRY_KEY)
            // Keep server URL and remember me settings
        }
    }

    // Token validation
    suspend fun isTokenValid(): Boolean {
        val token = authToken.first()
        val expiry = dataStore.data.map { it[TOKEN_EXPIRY_KEY] ?: 0L }.first()
        return !token.isNullOrBlank() && System.currentTimeMillis() < expiry
    }

    suspend fun isAuthenticated(): Boolean {
        return isAuthenticated.first()
    }

    suspend fun getValidToken(): String? {
        return if (isTokenValid()) {
            authToken.first()
        } else {
            // Try to refresh token
            val refreshResult = refreshAuthToken()
            if (refreshResult.isSuccess) {
                authToken.first()
            } else {
                null
            }
        }
    }

    // Utility methods for UI
    suspend fun getUserDisplayName(): String {
        val user = currentUser.first()
        return user?.let {
            if (it.firstName.isNotBlank() || it.lastName.isNotBlank()) {
                it.fullName
            } else {
                it.username
            }
        } ?: "Unknown User"
    }

    suspend fun getUserInitials(): String {
        val user = currentUser.first()
        return user?.let {
            val first = it.firstName.firstOrNull()?.toString() ?: ""
            val last = it.lastName.firstOrNull()?.toString() ?: ""
            if (first.isNotBlank() || last.isNotBlank()) {
                "$first$last".uppercase()
            } else {
                it.username.take(2).uppercase()
            }
        } ?: "?"
    }

    // Auto-logout on token expiry
    fun observeTokenExpiry(): Flow<Boolean> {
        return dataStore.data.map { preferences ->
            val expiry = preferences[TOKEN_EXPIRY_KEY] ?: 0L
            System.currentTimeMillis() >= expiry
        }.distinctUntilChanged()
    }
}