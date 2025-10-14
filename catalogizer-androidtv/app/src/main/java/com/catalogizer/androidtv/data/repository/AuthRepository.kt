package com.catalogizer.androidtv.data.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

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
            // TODO: Implement login API call
            // For now, simulate successful login
            dataStore.edit { preferences ->
                preferences[TOKEN_KEY] = "mock_token"
                preferences[USER_ID_KEY] = 1L
                preferences[USERNAME_KEY] = username
            }
            Result.success(Unit)
        } catch (e: Exception) {
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
            // TODO: Implement token refresh
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}