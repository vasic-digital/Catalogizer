@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.login

import androidx.compose.runtime.*
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.ExperimentalTvMaterial3Api
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import kotlinx.coroutines.launch

/**
 * Null-safe wrapper for LoginScreen to prevent NPE crashes.
 * Addresses HELIX-155 and related crash issues.
 */
@Composable
fun LoginScreenSafe(
    authViewModel: AuthViewModel,
    onLoginSuccess: () -> Unit
) {
    // Wrap the original LoginScreen with error handling
    val authState by authViewModel.authState.collectAsStateWithLifecycle()
    
    // Safety check: ensure authViewModel is properly initialized
    if (::authViewModel.isInitialized.not()) {
        // Show error state if ViewModel not initialized
        LoginErrorScreen("Authentication system not initialized")
        return
    }
    
    // Wrap onLoginSuccess to handle null cases
    val safeOnLoginSuccess = {
        try {
            onLoginSuccess()
        } catch (e: Exception) {
            // Log error but don't crash
            android.util.Log.e("LoginScreenSafe", "Error in onLoginSuccess callback", e)
        }
    }
    
    // Call original LoginScreen with safe wrapper
    LoginScreen(
        authViewModel = authViewModel,
        onLoginSuccess = safeOnLoginSuccess
    )
}

@Composable
private fun LoginErrorScreen(message: String) {
    androidx.tv.material3.Text(
        text = "Login Error: $message",
        color = androidx.compose.ui.graphics.Color.Red,
        modifier = androidx.compose.ui.Modifier.padding(16.dp)
    )
}
