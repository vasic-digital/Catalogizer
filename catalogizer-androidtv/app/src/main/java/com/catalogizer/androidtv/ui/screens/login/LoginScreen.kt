@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.login

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.focusable
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.TextField
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.Button
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.Surface
import androidx.tv.material3.Text
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.models.ServerEntry
import com.catalogizer.androidtv.ui.viewmodel.AuthViewModel
import kotlinx.coroutines.launch

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

    // Server configuration state
    val container = DependencyContainer.getInstance(androidx.compose.ui.platform.LocalContext.current)
    var serverUrl by remember { mutableStateOf(container.getServerUrl()) }
    var showServerConfig by remember { mutableStateOf(false) }
    var isDiscovering by remember { mutableStateOf(false) }
    var discoveredServers by remember { mutableStateOf<List<ServerEntry>>(emptyList()) }
    val coroutineScope = rememberCoroutineScope()

    val usernameFocusRequester = remember { FocusRequester() }
    val passwordFocusRequester = remember { FocusRequester() }
    val keyboardController = LocalSoftwareKeyboardController.current

    // Auto-navigate if already authenticated
    LaunchedEffect(authState) {
        if (authState.isAuthenticated) {
            onLoginSuccess()
        }
    }

    // Auto-discover on first launch
    LaunchedEffect(Unit) {
        val settings = container.settingsRepository.getSettingsAsync()
        if (settings.autoDiscovery) {
            isDiscovering = true
            try {
                val results = container.discoveryService.discoverViaMulticast(3000L)
                discoveredServers = results
                if (results.size == 1) {
                    // Auto-select single discovered server
                    val server = results.first()
                    serverUrl = server.url
                    container.switchServer(server.url)
                    container.settingsRepository.updateServerUrl(server.url)
                    container.settingsRepository.addServer(server)
                }
            } catch (_: Exception) { }
            isDiscovering = false
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

            // ─── Server URL section ─────────────────────────────────────
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.width(400.dp)
            ) {
                TextField(
                    value = serverUrl,
                    onValueChange = { newUrl ->
                        serverUrl = newUrl
                    },
                    label = { Text("Server URL") },
                    modifier = Modifier.weight(1f).focusable(),
                    singleLine = true,
                    enabled = !isLoading,
                    trailingIcon = {
                        if (isDiscovering) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(20.dp),
                                strokeWidth = 2.dp
                            )
                        }
                    }
                )

                // Discover button
                IconButton(
                    onClick = {
                        coroutineScope.launch {
                            isDiscovering = true
                            discoveredServers = emptyList()
                            try {
                                val results = container.discoveryService.discoverViaMulticast(5000L)
                                discoveredServers = results
                                if (results.isEmpty()) {
                                    errorMessage = "No servers found on the network"
                                }
                            } catch (e: Exception) {
                                errorMessage = "Discovery failed: ${e.message}"
                            }
                            isDiscovering = false
                        }
                    },
                    enabled = !isDiscovering && !isLoading
                ) {
                    Icon(Icons.Default.Search, contentDescription = "Discover servers")
                }

                // Apply server URL button
                IconButton(
                    onClick = {
                        if (serverUrl.isNotBlank()) {
                            container.switchServer(serverUrl)
                            coroutineScope.launch {
                                container.settingsRepository.updateServerUrl(serverUrl)
                                container.settingsRepository.addServer(
                                    ServerEntry(url = serverUrl, name = "Manual entry")
                                )
                            }
                            errorMessage = null
                        }
                    },
                    enabled = serverUrl.isNotBlank() && !isLoading
                ) {
                    Icon(Icons.Default.Settings, contentDescription = "Apply server")
                }
            }

            // Discovered servers list
            if (discoveredServers.isNotEmpty()) {
                Text(
                    text = "Discovered servers:",
                    style = androidx.tv.material3.MaterialTheme.typography.bodySmall
                )
                discoveredServers.forEach { server ->
                    Button(
                        onClick = {
                            serverUrl = server.url
                            container.switchServer(server.url)
                            coroutineScope.launch {
                                container.settingsRepository.updateServerUrl(server.url)
                                container.settingsRepository.addServer(server)
                            }
                            discoveredServers = emptyList()
                        },
                        modifier = Modifier.width(400.dp)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Text(
                                text = server.name.ifBlank { "Catalogizer API" },
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                                modifier = Modifier.weight(1f)
                            )
                            Text(text = server.url)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // ─── Username field ──────────────────────────────────────────
            TextField(
                value = username,
                onValueChange = { newValue: String ->
                    username = newValue
                    errorMessage = null
                },
                label = { Text("Username") },
                modifier = Modifier
                    .width(400.dp)
                    .focusRequester(usernameFocusRequester)
                    .focusable(),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
                keyboardActions = KeyboardActions(
                    onNext = { passwordFocusRequester.requestFocus() }
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
                    .width(400.dp)
                    .focusRequester(passwordFocusRequester)
                    .focusable(),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
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
                Box(
                    modifier = Modifier
                        .width(400.dp)
                        .padding(8.dp)
                        .background(
                            MaterialTheme.colorScheme.errorContainer,
                            MaterialTheme.shapes.medium
                        )
                ) {
                    Text(
                        text = error,
                        modifier = Modifier.padding(16.dp),
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }

            // Login button
            Button(
                onClick = {
                    keyboardController?.hide()
                    performLogin(username, password, authViewModel, { isLoading = it }, { errorMessage = it })
                },
                modifier = Modifier.width(400.dp),
                enabled = !isLoading && username.isNotBlank() && password.isNotBlank()
            ) {
                if (isLoading) {
                    CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp)
                } else {
                    Text("Login")
                }
            }

            // Connected server indicator
            Text(
                text = "Connected to: $serverUrl",
                style = androidx.tv.material3.MaterialTheme.typography.bodySmall
            )
        }
    }

    // Focus username field on launch
    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(100)
        usernameFocusRequester.requestFocus()
    }

    DisposableEffect(Unit) {
        onDispose { authViewModel.clearError() }
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
