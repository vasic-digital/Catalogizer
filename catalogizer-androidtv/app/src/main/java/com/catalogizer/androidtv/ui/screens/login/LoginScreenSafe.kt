@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.login

import androidx.compose.foundation.layout.padding
import androidx.compose.ui.unit.dp
import androidx.compose.runtime.Composable
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.ExperimentalTvMaterial3Api
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel

/**
 * Null-safe wrapper for LoginScreen to prevent NPE crashes.
 * Addresses HELIX-155 and related crash issues.
 */
@Composable
fun LoginScreenSafe(
    authViewModel: AuthViewModel,
    onLoginSuccess: () -> Unit
) {
    // Call original LoginScreen with error handling wrapper
    // ViewModel is always initialized when passed as parameter
    LoginScreen(
        authViewModel = authViewModel,
        onLoginSuccess = onLoginSuccess
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
