package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.*
import kotlinx.coroutines.sync.Mutex
import okhttp3.Interceptor
import okhttp3.Response

/**
 * OkHttp [Interceptor] that attaches Bearer authentication tokens to outgoing requests
 * and proactively refreshes tokens nearing expiration. Skips auth for login endpoints.
 */
class AuthInterceptor(private val authRepository: AuthRepository) : Interceptor {
    private val scope = CoroutineScope(Dispatchers.IO)
    private val refreshMutex = Mutex()
    private var refreshJob: Job? = null

    private fun refreshTokenIfNeeded() {
        if (!authRepository.shouldRefreshToken()) return
        synchronized(this) {
            if (refreshJob?.isActive == true) return
            refreshJob = scope.launch {
                try {
                    withTimeout(10_000L) {
                        authRepository.refreshToken()
                    }
                } catch (e: Exception) {
                    // Token refresh failed, will retry on next request
                } finally {
                    synchronized(this@AuthInterceptor) {
                        refreshJob = null
                    }
                }
            }
        }
    }

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip auth for login endpoint
        if (originalRequest.url.encodedPath.contains("/auth/login")) {
            return chain.proceed(originalRequest)
        }

        // Refresh token if needed, without blocking the network thread
        refreshTokenIfNeeded()

        // Add authorization header if we have a token (synchronous StateFlow access)
        val authState = authRepository.authState.value
        android.util.Log.d("AuthInterceptor", "isAuthenticated=${authState.isAuthenticated}, hasToken=${authState.token != null}, url=${originalRequest.url.encodedPath}")
        val newRequest = if (authState.isAuthenticated && authState.token != null) {
            originalRequest.newBuilder()
                .addHeader("Authorization", "Bearer ${authState.token}")
                .build()
        } else {
            originalRequest
        }

        val response = chain.proceed(newRequest)

        // HELIX-157 fix — mid-playback session expiry handling. When
        // the server returns 401 (access token expired after our
        // proactive-refresh heuristic missed the window), block for
        // up to 5 s waiting for a refresh, then retry the request
        // ONCE with the new bearer. Prevents the "abruptly cut to
        // login screen mid-scene" scenario the bank test flagged.
        //
        // Bounded: one retry max, 5 s total. Never recurses — if the
        // refresh also fails or the retry also 401s, we return the
        // 401 to the caller (which ExoPlayer surfaces as a playback
        // error + the UI layer routes to login).
        if (response.code == 401 && !originalRequest.url.encodedPath.contains("/auth/")) {
            response.close()
            val refreshed = runBlocking {
                withTimeoutOrNull(5_000L) {
                    try {
                        authRepository.refreshToken()
                        true
                    } catch (_: Exception) {
                        false
                    }
                } ?: false
            }
            if (refreshed) {
                val retryState = authRepository.authState.value
                if (retryState.isAuthenticated && retryState.token != null) {
                    val retryRequest = originalRequest.newBuilder()
                        .addHeader("Authorization", "Bearer ${retryState.token}")
                        .build()
                    return chain.proceed(retryRequest)
                }
            }
            // Refresh failed or produced no token — let the 401 bubble
            // up so the UI layer can route to the login screen.
            // Re-execute the original request to re-create a Response
            // object the caller can consume (the old one was closed).
            return chain.proceed(newRequest)
        }

        return response
    }
}