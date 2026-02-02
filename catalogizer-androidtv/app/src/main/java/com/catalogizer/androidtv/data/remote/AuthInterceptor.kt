package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor(private val authRepository: AuthRepository) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip auth for login endpoint
        if (originalRequest.url.encodedPath.contains("/auth/login")) {
            return chain.proceed(originalRequest)
        }

        // Refresh token if needed, with timeout to prevent blocking indefinitely
        if (authRepository.shouldRefreshToken()) {
            runBlocking(Dispatchers.IO) {
                withTimeout(10_000L) {
                    authRepository.refreshToken()
                }
            }
        }

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