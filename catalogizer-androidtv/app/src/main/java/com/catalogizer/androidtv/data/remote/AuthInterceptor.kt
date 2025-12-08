package com.catalogizer.androidtv.data.remote

import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor(private val authRepository: AuthRepository) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Skip auth for login endpoint
        if (originalRequest.url.encodedPath.contains("/auth/login")) {
            return chain.proceed(originalRequest)
        }

        // Check if we need to refresh token
        if (authRepository.shouldRefreshToken()) {
            runBlocking {
                authRepository.refreshToken()
            }
        }

        // Add authorization header if we have a token
        val authState = runBlocking { authRepository.authState.value }
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