package com.catalogizer.androidtv.ui.viewmodel

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.catalogizer.androidtv.data.repository.AuthRepository
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeoutOrNull

/**
 * Controls application initialization state used by the TV splash screen.
 * Exposes [isLoading] as a [StateFlow] to signal when the app is ready.
 *
 * When QA auto-login is pending ([qaLoginGate] is set), the splash screen
 * stays visible until login completes or times out, preventing the empty
 * library race condition where HomeScreen loads before auth token is available.
 */
class MainViewModel(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    /**
     * Gate that blocks [initializeApp] until QA auto-login completes.
     * Set via [setQaLoginGate] before [initializeApp] is called.
     * Null when no QA login is pending (normal startup).
     */
    private var qaLoginGate: CompletableDeferred<Unit>? = null

    /**
     * Sets a gate that [initializeApp] will await before marking loading
     * complete. Call [completeQaLogin] after login succeeds or fails.
     */
    fun setQaLoginGate(gate: CompletableDeferred<Unit>) {
        qaLoginGate = gate
    }

    /**
     * Completes the QA login gate, unblocking [initializeApp].
     */
    fun completeQaLogin() {
        qaLoginGate?.complete(Unit)
    }

    fun initializeApp() {
        viewModelScope.launch {
            try {
                // If QA auto-login is pending, wait for it (max 10s) before
                // showing the app. This prevents HomeScreen from loading with
                // no auth token and displaying an empty library.
                qaLoginGate?.let { gate ->
                    withTimeoutOrNull(10_000L) { gate.await() }
                }
            } catch (e: Exception) {
                Log.w("MainViewModel", "Error during QA login wait: ${e.message}")
            }
            // ALWAYS set loading to false to prevent ANR
            // Splash screen has its own minimum duration timer
            _isLoading.value = false
            Log.d("MainViewModel", "App initialization complete, isLoading=false")
        }
    }
}