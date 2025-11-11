package com.catalogizer.androidtv.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.first

class AuthRepository(
    private val api: CatalogizerApi,
    private val dataStore: DataStore<Preferences>
) {

    private val TOKEN_KEY = stringPreferencesKey("auth_token")
    private val USER_ID_KEY = longPreferencesKey("user_id")
    private val USERNAME_KEY = stringPreferencesKey("username")

    val authState: Flow<AuthState> = dataStore.data.map { preferences ->
        val token = preferences[TOKEN_KEY]
        val userId = preferences[USER_ID_KEY]
        val username = preferences[USERNAME_KEY]

        if (token != null && userId != null && username != null) {
            AuthState.Authenticated(token, userId, username)
        } else {
            AuthState.Unauthenticated
        }
    }

    suspend fun login(username: String, password: String): Result<Unit> {
        return try {
            // Call login API
            val credentials = mapOf(
                "username" to username,
                "password" to password
            )
            val response = api.login(credentials)

            if (response.isSuccessful) {
                val loginResponse = response.body()
                if (loginResponse != null) {
                    // Save authentication data
                    dataStore.edit { preferences ->
                        preferences[TOKEN_KEY] = loginResponse.token
                        preferences[USER_ID_KEY] = loginResponse.userId
                        preferences[USERNAME_KEY] = loginResponse.username
                    }
                    android.util.Log.d("AuthRepository", "Login successful for user: $username")
                    Result.success(Unit)
                } else {
                    android.util.Log.e("AuthRepository", "Login response body is null")
                    Result.failure(Exception("Login response is empty"))
                }
            } else {
                android.util.Log.e("AuthRepository", "Login failed: ${response.code()} ${response.message()}")
                Result.failure(Exception("Login failed: ${response.message()}"))
            }
        } catch (e: Exception) {
            android.util.Log.e("AuthRepository", "Login error", e)
            Result.failure(e)
        }
    }

    suspend fun logout() {
        dataStore.edit { preferences ->
            preferences.remove(TOKEN_KEY)
            preferences.remove(USER_ID_KEY)
            preferences.remove(USERNAME_KEY)
        }
    }

    suspend fun refreshToken(): Result<Unit> {
        return try {
            // Get current token from dataStore
            val preferences = dataStore.data.first()
            val currentToken = preferences[TOKEN_KEY]

            if (currentToken == null) {
                android.util.Log.w("AuthRepository", "No token to refresh")
                return Result.failure(Exception("No authentication token found"))
            }

            // Call refresh token API
            val tokenData = mapOf("token" to currentToken)
            val response = api.refreshToken(tokenData)

            if (response.isSuccessful) {
                val loginResponse = response.body()
                if (loginResponse != null) {
                    // Update authentication data with new token
                    dataStore.edit { prefs ->
                        prefs[TOKEN_KEY] = loginResponse.token
                        prefs[USER_ID_KEY] = loginResponse.userId
                        prefs[USERNAME_KEY] = loginResponse.username
                    }
                    android.util.Log.d("AuthRepository", "Token refreshed successfully")
                    Result.success(Unit)
                } else {
                    android.util.Log.e("AuthRepository", "Refresh response body is null")
                    Result.failure(Exception("Refresh response is empty"))
                }
            } else {
                android.util.Log.e("AuthRepository", "Token refresh failed: ${response.code()} ${response.message()}")
                // If refresh fails, clear authentication
                logout()
                Result.failure(Exception("Token refresh failed: ${response.message()}"))
            }
        } catch (e: Exception) {
            android.util.Log.e("AuthRepository", "Token refresh error", e)
            Result.failure(e)
        }
    }
}