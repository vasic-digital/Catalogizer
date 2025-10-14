package com.catalogizer.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.android.data.repository.MediaRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class MainViewModel(
    private val mediaRepository: MediaRepository
) : ViewModel() {

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    fun initializeApp() {
        viewModelScope.launch {
            try {
                // Perform any initialization logic here
                // For example: check database, sync data, etc.
                _isLoading.value = false
            } catch (e: Exception) {
                // Handle initialization error
                _isLoading.value = false
            }
        }
    }
}