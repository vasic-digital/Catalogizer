package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.models.AuthState
import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

class AuthViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    val authState = authRepository.authState.stateIn(
        viewModelScope,
        SharingStarted.WhileSubscribed(5000),
        AuthState.Unauthenticated
    )

    fun login(username: String, password: String) {
        viewModelScope.launch {
            try {
                authRepository.login(username, password)
            } catch (e: Exception) {
                // Handle login error - the repository will update the authState with error
            }
        }
    }

    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
        }
    }

    fun refreshToken() {
        viewModelScope.launch {
            authRepository.refreshToken()
        }
    }

    fun clearError() {
        viewModelScope.launch {
            authRepository.clearError()
        }
    }
}