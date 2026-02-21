package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.*
import kotlinx.coroutines.sync.Mutex
import okhttp3.Interceptor
import okhttp3.Response

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
        val newRequest = if (authState.isAuthenticated && authState.token != null) {
            originalRequest.newBuilder()
                .addHeader("Authorization", "Bearer ${authState.token}")
                .build()
        } else {
            originalRequest
        }

        return chain.proceed(newRequest)
    }
}