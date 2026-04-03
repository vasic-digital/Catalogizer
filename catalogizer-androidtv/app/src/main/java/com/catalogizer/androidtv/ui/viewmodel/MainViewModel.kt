package com.catalogizer.androidtv.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * Controls application initialization state used by the TV splash screen.
 * Exposes [isLoading] as a [StateFlow] to signal when the app is ready.
 */
class MainViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    fun initializeApp() {
        viewModelScope.launch {
            // Perform any initialization logic here
            // For now, just mark as loaded
            _isLoading.value = false
        }
    }
}