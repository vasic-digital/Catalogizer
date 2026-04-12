package com.catalogizer.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.android.data.models.AuthState
import com.catalogizer.android.data.repository.AuthRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * Manages authentication state including login, logout, and initial auth status check.
 * Exposes [authState] as a [StateFlow] for reactive UI updates.
 */
class AuthViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _authState = MutableStateFlow(AuthState())
    val authState: StateFlow<AuthState> = _authState.asStateFlow()

    init {
        checkAuthStatus()
    }

    private fun checkAuthStatus() {
        viewModelScope.launch {
            val isAuthenticated = authRepository.isAuthenticated()
            _authState.value = AuthState(isAuthenticated = isAuthenticated)
        }
    }

    fun login(username: String, password: String) {
        viewModelScope.launch {
            try {
                _authState.value = _authState.value.copy(isLoading = true, error = null)
                val result = authRepository.login(username, password)
                if (result.isSuccess) {
                    _authState.value = AuthState(isAuthenticated = true)
                } else {
                    _authState.value = _authState.value.copy(
                        isLoading = false,
                        error = result.error ?: "Login failed"
                    )
                }
            } catch (e: Exception) {
                _authState.value = _authState.value.copy(
                    isLoading = false,
                    error = e.message ?: "Login failed"
                )
            }
        }
    }

    /** Re-check auth status from the repository. Called by MainActivity
     *  after the QA auto-login completes on the IO dispatcher so the
     *  Compose-observed [authState] reflects the new token. */
    fun checkAuthState() {
        viewModelScope.launch {
            val isAuthenticated = authRepository.isAuthenticated()
            _authState.value = AuthState(isAuthenticated = isAuthenticated)
        }
    }

    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
            _authState.value = AuthState(isAuthenticated = false)
        }
    }
}