@file:OptIn(ExperimentalTvMaterial3Api::class, ExperimentalComposeUiApi::class)
package com.catalogizer.androidtv.ui.screens.login

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.focusable
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Divider
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.min
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.tv.material3.Button
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
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

    val container = DependencyContainer.getInstance(androidx.compose.ui.platform.LocalContext.current)
    val settings by container.settingsRepository.settingsFlow.collectAsStateWithLifecycle(
        initialValue = com.catalogizer.androidtv.data.models.Settings(
            enableNotifications = true, enableAutoPlay = false, streamingQuality = "Auto",
            enableSubtitles = true, subtitleLanguage = "English"
        )
    )
    var serverUrl by remember { mutableStateOf(container.getServerUrl()) }
    var isDiscovering by remember { mutableStateOf(false) }
    var discoveredServers by remember { mutableStateOf<List<ServerEntry>>(emptyList()) }
    val coroutineScope = rememberCoroutineScope()

    // Sync server URL from DataStore when settings load
    LaunchedEffect(settings.serverUrl) {
        if (settings.serverUrl.isNotBlank() && settings.serverUrl != serverUrl) {
            serverUrl = settings.serverUrl
            container.switchServer(settings.serverUrl)
        }
    }

    val usernameFocusRequester = remember { FocusRequester() }
    val passwordFocusRequester = remember { FocusRequester() }
    val keyboardController = LocalSoftwareKeyboardController.current
    val scrollState = rememberScrollState()

    // Muted text field colors for dark TV theme
    val textFieldColors = OutlinedTextFieldDefaults.colors(
        focusedTextColor = Color.White,
        unfocusedTextColor = Color.White.copy(alpha = 0.9f),
        focusedContainerColor = Color.White.copy(alpha = 0.06f),
        unfocusedContainerColor = Color.White.copy(alpha = 0.04f),
        focusedBorderColor = MaterialTheme.colorScheme.primary,
        unfocusedBorderColor = Color.White.copy(alpha = 0.3f),
        focusedLabelColor = MaterialTheme.colorScheme.primary,
        unfocusedLabelColor = Color.White.copy(alpha = 0.7f),
        cursorColor = MaterialTheme.colorScheme.primary
    )

    LaunchedEffect(authState) {
        if (authState.isAuthenticated) onLoginSuccess()
    }

    LaunchedEffect(Unit) {
        val initialSettings = container.settingsRepository.getSettingsAsync()
        if (initialSettings.autoDiscovery) {
            isDiscovering = true
            try {
                val results = container.discoveryService.discoverAll(8000L)
                discoveredServers = results
                if (results.size == 1) {
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

    LaunchedEffect(authState.error) {
        errorMessage = authState.error
        isLoading = false
    }

    // Responsive sizing: adapt to screen width
    val configuration = LocalConfiguration.current
    val screenWidthDp = configuration.screenWidthDp.dp
    val isCompact = screenWidthDp < 600.dp  // Phone
    val formWidth = if (isCompact) screenWidthDp - 48.dp else min(520.dp, screenWidthDp * 0.5f)
    val horizontalPad = if (isCompact) 24.dp else 80.dp
    val verticalPad = if (isCompact) 24.dp else 40.dp

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFF121212)),
        contentAlignment = Alignment.TopCenter
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier
                .verticalScroll(scrollState)
                .padding(horizontal = horizontalPad, vertical = verticalPad)
                .widthIn(max = 600.dp)
        ) {
            // ─── Branding ───────────────────────────────────────────
            Spacer(modifier = Modifier.height(if (isCompact) 8.dp else 24.dp))

            Text(
                text = "CATALOGIZER",
                style = if (isCompact) MaterialTheme.typography.headlineMedium else MaterialTheme.typography.headlineLarge,
                color = Color.White,
                letterSpacing = androidx.compose.ui.unit.TextUnit(if (isCompact) 2f else 4f, androidx.compose.ui.unit.TextUnitType.Sp)
            )

            Spacer(modifier = Modifier.height(4.dp))

            Text(
                text = "Media Collection Manager",
                style = MaterialTheme.typography.bodyMedium,
                color = Color.White.copy(alpha = 0.7f)
            )

            Spacer(modifier = Modifier.height(if (isCompact) 16.dp else 32.dp))

            // ─── Credentials ────────────────────────────────────────
            OutlinedTextField(
                value = username,
                onValueChange = { username = it; errorMessage = null },
                label = { Text("Username", color = Color.White.copy(alpha = 0.7f)) },
                modifier = Modifier
                    .width(formWidth)
                    .focusRequester(usernameFocusRequester)
                    .focusable(),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
                keyboardActions = KeyboardActions(onNext = { passwordFocusRequester.requestFocus() }),
                singleLine = true,
                enabled = !isLoading,
                colors = textFieldColors
            )

            Spacer(modifier = Modifier.height(12.dp))

            OutlinedTextField(
                value = password,
                onValueChange = { password = it; errorMessage = null },
                label = { Text("Password", color = Color.White.copy(alpha = 0.7f)) },
                modifier = Modifier
                    .width(formWidth)
                    .focusRequester(passwordFocusRequester)
                    .focusable(),
                keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
                keyboardActions = KeyboardActions(onDone = {
                    keyboardController?.hide()
                    performLogin(username, password, authViewModel, { isLoading = it }, { errorMessage = it })
                }),
                visualTransformation = PasswordVisualTransformation(),
                singleLine = true,
                enabled = !isLoading,
                colors = textFieldColors
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Error message
            errorMessage?.let { error ->
                Box(
                    modifier = Modifier
                        .width(formWidth)
                        .background(Color(0xFF442222), MaterialTheme.shapes.small)
                        .padding(12.dp)
                ) {
                    Text(
                        text = error,
                        style = MaterialTheme.typography.bodySmall,
                        color = Color(0xFFFF8888)
                    )
                }
                Spacer(modifier = Modifier.height(12.dp))
            }

            // Login button
            Button(
                onClick = {
                    keyboardController?.hide()
                    performLogin(username, password, authViewModel, { isLoading = it }, { errorMessage = it })
                },
                modifier = Modifier.width(formWidth).height(if (isCompact) 48.dp else 52.dp),
                enabled = !isLoading && username.isNotBlank() && password.isNotBlank()
            ) {
                if (isLoading) {
                    CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp, color = Color.White)
                } else {
                    Text("Sign In", style = MaterialTheme.typography.titleMedium)
                }
            }

            Spacer(modifier = Modifier.height(if (isCompact) 16.dp else 28.dp))

            // ─── Server Configuration (collapsible section) ─────────
            Divider(color = Color.White.copy(alpha = 0.1f), modifier = Modifier.width(formWidth))

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "SERVER CONNECTION",
                style = MaterialTheme.typography.labelSmall,
                color = Color.White.copy(alpha = 0.6f),
                letterSpacing = androidx.compose.ui.unit.TextUnit(2f, androidx.compose.ui.unit.TextUnitType.Sp)
            )

            Spacer(modifier = Modifier.height(12.dp))

            OutlinedTextField(
                value = serverUrl,
                onValueChange = { serverUrl = it },
                label = { Text("Server URL", color = Color.White.copy(alpha = 0.7f)) },
                modifier = Modifier.width(formWidth).focusable(),
                singleLine = true,
                enabled = !isLoading,
                colors = textFieldColors,
                trailingIcon = {
                    if (isDiscovering) {
                        CircularProgressIndicator(modifier = Modifier.size(20.dp), strokeWidth = 2.dp, color = MaterialTheme.colorScheme.primary)
                    }
                }
            )

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                modifier = Modifier.width(formWidth)
            ) {
                Button(
                    onClick = {
                        coroutineScope.launch {
                            isDiscovering = true
                            discoveredServers = emptyList()
                            try {
                                val results = container.discoveryService.discoverAll(10000L)
                                discoveredServers = results
                                if (results.isEmpty()) errorMessage = "No servers found on the network"
                            } catch (e: Exception) {
                                errorMessage = "Discovery failed: ${e.message}"
                            }
                            isDiscovering = false
                        }
                    },
                    modifier = Modifier.weight(1f).height(44.dp),
                    enabled = !isDiscovering && !isLoading
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(Icons.Default.Search, contentDescription = "Discover servers", modifier = Modifier.size(18.dp))
                        Text("Discover", style = MaterialTheme.typography.bodyMedium)
                    }
                }

                Button(
                    onClick = {
                        if (serverUrl.isNotBlank()) {
                            container.switchServer(serverUrl)
                            coroutineScope.launch {
                                container.settingsRepository.updateServerUrl(serverUrl)
                                container.settingsRepository.addServer(ServerEntry(url = serverUrl, name = "Manual"))
                            }
                            errorMessage = null
                        }
                    },
                    modifier = Modifier.weight(1f).height(44.dp),
                    enabled = serverUrl.isNotBlank() && !isLoading
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(Icons.Default.Settings, contentDescription = "Connect to server", modifier = Modifier.size(18.dp))
                        Text("Connect", style = MaterialTheme.typography.bodyMedium)
                    }
                }
            }

            // Discovered servers
            if (discoveredServers.isNotEmpty()) {
                Spacer(modifier = Modifier.height(12.dp))
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
                        modifier = Modifier.width(formWidth).height(44.dp)
                    ) {
                        Text(
                            text = "${server.name.ifBlank { "API" }} — ${server.url}",
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                    }
                    Spacer(modifier = Modifier.height(8.dp))
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Connection status
            Text(
                text = serverUrl,
                style = MaterialTheme.typography.bodySmall,
                color = Color.White.copy(alpha = 0.5f)
            )

            Spacer(modifier = Modifier.height(32.dp))

            // ─── Vasic Digital Branding Footer ─────────────────────────
            Divider(
                color = Color.White.copy(alpha = 0.1f),
                modifier = Modifier.width(formWidth)
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "Made with \u2764 by Vasic Digital",
                style = MaterialTheme.typography.bodySmall,
                color = Color.White.copy(alpha = 0.5f)
            )
            Spacer(modifier = Modifier.height(16.dp))
        }
    }

    LaunchedEffect(Unit) {
        kotlinx.coroutines.delay(200)
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
