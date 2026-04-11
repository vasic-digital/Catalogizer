package com.catalogizer.androidtv.data.auth

import android.content.Context
import android.content.SharedPreferences
import android.util.Log
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

/**
 * Persists the JWT session returned by /api/v1/auth/login so the
 * app can reopen into the home screen instead of the login form
 * after a cold start or force-stop.
 *
 * Production uses [encrypted] which backs the store with
 * [EncryptedSharedPreferences] — AES-256-GCM encrypted, keyed via
 * the Android keystore. Tests inject a plain [SharedPreferences]
 * (Robolectric's in-memory backing) so the save/load/clear logic
 * can run without requiring the keystore.
 *
 * I/O happens on [Dispatchers.IO] because EncryptedSharedPreferences
 * blocks on first access while unwrapping the master key.
 */
class TokenStore internal constructor(
    private val prefs: SharedPreferences,
) {

    data class Record(
        val token: String,
        val username: String,
        val userId: Long,
        val expiresAtMs: Long,
    )

    suspend fun save(record: Record) = withContext(Dispatchers.IO) {
        try {
            prefs.edit()
                .putString(KEY_TOKEN, record.token)
                .putString(KEY_USERNAME, record.username)
                .putLong(KEY_USER_ID, record.userId)
                .putLong(KEY_EXPIRES_AT, record.expiresAtMs)
                .commit()
        } catch (e: Throwable) {
            Log.e("TokenStore", "save failed: ${e.message}", e)
        }
        Unit
    }

    suspend fun load(): Record? = withContext(Dispatchers.IO) {
        val token = try {
            prefs.getString(KEY_TOKEN, null)
        } catch (e: Throwable) {
            Log.e("TokenStore", "load failed: ${e.message}", e)
            null
        } ?: return@withContext null
        val username = prefs.getString(KEY_USERNAME, "") ?: ""
        val userId = prefs.getLong(KEY_USER_ID, 0)
        val expiresAt = prefs.getLong(KEY_EXPIRES_AT, 0)
        Record(token = token, username = username, userId = userId, expiresAtMs = expiresAt)
    }

    suspend fun clear() = withContext(Dispatchers.IO) {
        prefs.edit().clear().commit()
    }

    /**
     * Returns true iff a non-blank token is stored and its expiry
     * (if known) is in the future. A zero expiresAtMs is treated
     * as "unknown" and reported as unauthenticated, forcing a
     * fresh login on next launch. Callers pass `nowMs` explicitly
     * so tests can inject a fixed clock.
     */
    suspend fun isAuthenticated(nowMs: Long = System.currentTimeMillis()): Boolean {
        val rec = load() ?: return false
        if (rec.token.isBlank()) return false
        if (rec.expiresAtMs == 0L) return false
        return rec.expiresAtMs > nowMs
    }

    companion object {
        private const val DEFAULT_FILE = "catalogizer_token_store"
        private const val KEY_TOKEN = "token"
        private const val KEY_USERNAME = "username"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_EXPIRES_AT = "expires_at_ms"

        /**
         * Factory for production: wraps [EncryptedSharedPreferences]
         * with a keystore-backed master key. Must run on an Android
         * device or emulator — the Android keystore is not
         * available from Robolectric.
         */
        fun encrypted(
            context: Context,
            fileName: String = DEFAULT_FILE,
        ): TokenStore {
            val masterKey = MasterKey.Builder(context)
                .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                .build()
            val prefs = EncryptedSharedPreferences.create(
                context,
                fileName,
                masterKey,
                EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
                EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
            )
            return TokenStore(prefs)
        }

        /**
         * Factory for tests: uses a plain [SharedPreferences] so
         * unit tests can exercise the persistence logic without
         * needing the Android keystore.
         */
        internal fun inMemory(prefs: SharedPreferences): TokenStore = TokenStore(prefs)
    }
}
