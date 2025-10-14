package com.catalogizer.androidtv.ui.screens.login

import androidx.compose.foundation.layout.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.*
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel

@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun LoginScreen(
    authViewModel: AuthViewModel,
    onLoginSuccess: () -> Unit
) {
    val authState by authViewModel.authState.collectAsStateWithLifecycle()

    // Auto-navigate if already authenticated
    LaunchedEffect(authState) {
        if (authState.isAuthenticated) {
            onLoginSuccess()
        }
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
                style = MaterialTheme.typography.headlineLarge
            )

            Text(
                text = "Please log in to continue",
                style = MaterialTheme.typography.bodyLarge
            )

            Button(
                onClick = {
                    // For demo purposes, simulate login
                    authViewModel.login("demo", "demo")
                }
            ) {
                Text("Demo Login")
            }
        }
    }
}