@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.login

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.focusable
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.*
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel

@Composable
fun LoginScreen(
    authViewModel: AuthViewModel,
    onLoginSuccess: () -> Unit
) {
    val authState by authViewModel.authState.collectAsStateWithLifecycle()
    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var isLoading by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    
    val usernameFocusRequester = remember { FocusRequester() }
    val passwordFocusRequester = remember { FocusRequester() }
    val keyboardController = LocalSoftwareKeyboardController.current

    // Auto-navigate if already authenticated
    LaunchedEffect(authState) {
        if (authState.isAuthenticated) {
            onLoginSuccess()
        }
    }

    // Watch for login errors
    LaunchedEffect(authState.error) {
        errorMessage = authState.error
        isLoading = false
    }

    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp),
            modifier = Modifier.padding(48.dp)
        ) {
            Text(
                text = "Welcome to Catalogizer TV",
                style = androidx.tv.material3.MaterialTheme.typography.headlineLarge
            )

            Text(
                text = "Please enter your credentials",
                style = androidx.tv.material3.MaterialTheme.typography.bodyLarge
            )

            // Username field
            TextField(
                value = username,
                onValueChange = { newValue: String -> 
                    username = newValue
                    errorMessage = null
                },
                label = { Text("Username") },
                modifier = Modifier
                    .width(300.dp)
                    .focusRequester(usernameFocusRequester)
                    .focusable(),
                keyboardOptions = KeyboardOptions(
                    imeAction = ImeAction.Next
                ),
                keyboardActions = KeyboardActions(
                    onNext = {
                        passwordFocusRequester.requestFocus()
                    }
                ),
                singleLine = true,
                enabled = !isLoading
            )

            // Password field
            TextField(
                value = password,
                onValueChange = { newValue: String -> 
                    password = newValue
                    errorMessage = null
                },
                label = { Text("Password") },
                modifier = Modifier
                    .width(300.dp)
                    .focusRequester(passwordFocusRequester)
                    .focusable(),
                keyboardOptions = KeyboardOptions(
                    imeAction = ImeAction.Done
                ),
                keyboardActions = KeyboardActions(
                    onDone = {
                        keyboardController?.hide()
                        performLogin(username, password, authViewModel, { isLoading = it }, { errorMessage = it })
                    }
                ),
                visualTransformation = PasswordVisualTransformation(),
                singleLine = true,
                enabled = !isLoading
            )

            // Error message
            errorMessage?.let { error ->
                Surface(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(8.dp),
                    shape = androidx.tv.material3.MaterialTheme.shapes.medium,
                    color = androidx.tv.material3.MaterialTheme.colorScheme.errorContainer,
                    onClick = {} // Empty onClick for compatibility
                ) {
                    Text(
                        text = error,
                        modifier = Modifier.padding(16.dp),
                        style = androidx.tv.material3.MaterialTheme.typography.bodyMedium
                    )
                }
            }

            // Login button
            Button(
                onClick = {
                    keyboardController?.hide()
                    performLogin(username, password, authViewModel, { isLoading = it }, { errorMessage = it })
                },
                modifier = Modifier.width(300.dp),
                enabled = !isLoading && username.isNotBlank() && password.isNotBlank()
            ) {
                if (isLoading) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        strokeWidth = 2.dp
                    )
                } else {
                    Text("Login")
                }
            }

            // Demo login button (for development)
            Button(
                onClick = {
                    username = "demo"
                    password = "demo"
                },
                modifier = Modifier.width(300.dp),
                enabled = !isLoading
            ) {
                Text("Use Demo Credentials")
            }
        }
    }

    // Focus username field on launch
    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(100) // Small delay to ensure layout is ready
        usernameFocusRequester.requestFocus()
    }

    DisposableEffect(Unit) {
        onDispose {
            // Clear any state when screen is disposed
            authViewModel.clearError()
        }
    }
}

private fun performLogin(
    username: String, 
    password: String, 
    authViewModel: AuthViewModel, 
    setIsLoading: (Boolean) -> Unit, 
    setErrorMessage: (String?) -> Unit
) {
    if (username.isBlank() || password.isBlank()) {
        setErrorMessage("Please enter username and password")
        return
    }

    setIsLoading(true)
    setErrorMessage(null)
    
    authViewModel.login(username, password)
}