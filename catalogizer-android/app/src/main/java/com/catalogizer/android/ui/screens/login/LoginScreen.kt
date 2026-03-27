package com.catalogizer.android.ui.screens.login

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.material3.Divider
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.catalogizer.android.DependencyContainer
import com.catalogizer.android.ui.viewmodel.AuthViewModel
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.net.HttpURLConnection
import java.net.URL

@Composable
fun LoginScreen(
    authViewModel: AuthViewModel,
    onLoginSuccess: () -> Unit
) {
    val authState by authViewModel.authState.collectAsStateWithLifecycle()
    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var passwordVisible by remember { mutableStateOf(false) }
    val passwordFocusRequester = remember { FocusRequester() }

    val container = DependencyContainer.getInstance(LocalContext.current)
    var serverUrl by remember { mutableStateOf(container.getServerUrl()) }
    var serverConnected by remember { mutableStateOf(container.getServerUrl().isNotBlank()) }
    var isDiscovering by remember { mutableStateOf(false) }
    var discoveredServers by remember { mutableStateOf<List<String>>(emptyList()) }
    var serverError by remember { mutableStateOf<String?>(null) }
    val coroutineScope = rememberCoroutineScope()
    val scrollState = rememberScrollState()

    LaunchedEffect(authState.isAuthenticated) {
        if (authState.isAuthenticated) {
            onLoginSuccess()
        }
    }

    // Load persisted server URL on first composition
    LaunchedEffect(Unit) {
        val savedUrl = container.loadServerUrl()
        if (!savedUrl.isNullOrBlank()) {
            serverUrl = savedUrl
            if (container.getServerUrl() != savedUrl) {
                container.switchServer(savedUrl)
            }
            serverConnected = true
        }
    }

    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.Center
    ) {
        Card(
            modifier = Modifier
                .widthIn(max = 400.dp)
                .padding(24.dp)
        ) {
            Column(
                modifier = Modifier
                    .padding(24.dp)
                    .verticalScroll(scrollState),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Text(
                    text = "Catalogizer",
                    style = MaterialTheme.typography.headlineLarge,
                    color = MaterialTheme.colorScheme.primary
                )

                Text(
                    text = "Sign in to your account",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Spacer(modifier = Modifier.height(8.dp))

                OutlinedTextField(
                    value = username,
                    onValueChange = { username = it },
                    label = { Text("Username") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Next),
                    keyboardActions = KeyboardActions(
                        onNext = { passwordFocusRequester.requestFocus() }
                    ),
                    enabled = !authState.isLoading && serverConnected
                )

                OutlinedTextField(
                    value = password,
                    onValueChange = { password = it },
                    label = { Text("Password") },
                    singleLine = true,
                    modifier = Modifier
                        .fillMaxWidth()
                        .focusRequester(passwordFocusRequester),
                    visualTransformation = if (passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
                    trailingIcon = {
                        IconButton(onClick = { passwordVisible = !passwordVisible }) {
                            Icon(
                                imageVector = if (passwordVisible) Icons.Filled.VisibilityOff else Icons.Filled.Visibility,
                                contentDescription = if (passwordVisible) "Hide password" else "Show password"
                            )
                        }
                    },
                    keyboardOptions = KeyboardOptions(
                        keyboardType = KeyboardType.Password,
                        imeAction = ImeAction.Done
                    ),
                    keyboardActions = KeyboardActions(
                        onDone = {
                            if (username.isNotBlank() && password.isNotBlank() && serverConnected) {
                                authViewModel.login(username, password)
                            }
                        }
                    ),
                    enabled = !authState.isLoading && serverConnected
                )

                authState.error?.let { error ->
                    Text(
                        text = error,
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodySmall
                    )
                }

                Button(
                    onClick = { authViewModel.login(username, password) },
                    modifier = Modifier.fillMaxWidth(),
                    enabled = !authState.isLoading && username.isNotBlank() && password.isNotBlank() && serverConnected
                ) {
                    if (authState.isLoading) {
                        CircularProgressIndicator(
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                            color = MaterialTheme.colorScheme.onPrimary
                        )
                    } else {
                        Text("Sign In")
                    }
                }

                // --- Server Configuration ---
                Divider(color = MaterialTheme.colorScheme.outlineVariant)

                Text(
                    text = "Server Connection",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                OutlinedTextField(
                    value = serverUrl,
                    onValueChange = { serverUrl = it; serverError = null },
                    label = { Text("Server URL") },
                    placeholder = { Text("http://192.168.0.100:8080") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    enabled = !authState.isLoading,
                    trailingIcon = {
                        if (isDiscovering) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(20.dp),
                                strokeWidth = 2.dp
                            )
                        }
                    }
                )

                serverError?.let { error ->
                    Text(
                        text = error,
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodySmall
                    )
                }

                Row(
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    // Discover button — scans local network
                    OutlinedButton(
                        onClick = {
                            coroutineScope.launch {
                                isDiscovering = true
                                discoveredServers = emptyList()
                                serverError = null
                                try {
                                    val results = discoverServers()
                                    discoveredServers = results
                                    if (results.isEmpty()) {
                                        serverError = "No servers found on the network"
                                    }
                                } catch (e: Exception) {
                                    serverError = "Discovery failed: ${e.message}"
                                }
                                isDiscovering = false
                            }
                        },
                        modifier = Modifier.weight(1f),
                        enabled = !isDiscovering && !authState.isLoading
                    ) {
                        Icon(Icons.Default.Search, contentDescription = null, modifier = Modifier.size(18.dp))
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Discover")
                    }

                    // Connect button — validate and save
                    Button(
                        onClick = {
                            if (serverUrl.isNotBlank()) {
                                coroutineScope.launch {
                                    val reachable = container.probeServer(serverUrl)
                                    if (reachable) {
                                        container.switchServer(serverUrl)
                                        container.saveServerUrl(serverUrl)
                                        serverConnected = true
                                        serverError = null
                                    } else {
                                        // Still save and use — the server may not have /health yet
                                        container.switchServer(serverUrl)
                                        container.saveServerUrl(serverUrl)
                                        serverConnected = true
                                        serverError = "Server saved (could not verify — check URL if login fails)"
                                    }
                                }
                            }
                        },
                        modifier = Modifier.weight(1f),
                        enabled = serverUrl.isNotBlank() && !authState.isLoading
                    ) {
                        Icon(Icons.Default.Settings, contentDescription = null, modifier = Modifier.size(18.dp))
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Connect")
                    }
                }

                // Discovered servers list
                discoveredServers.forEach { url ->
                    OutlinedButton(
                        onClick = {
                            serverUrl = url
                            coroutineScope.launch {
                                container.switchServer(url)
                                container.saveServerUrl(url)
                                serverConnected = true
                                serverError = null
                            }
                            discoveredServers = emptyList()
                        },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text(
                            text = url,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                    }
                }

                // Connection status
                if (serverConnected && serverUrl.isNotBlank()) {
                    Text(
                        text = "Connected: $serverUrl",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                Spacer(modifier = Modifier.height(32.dp))

                // ─── Vasic Digital Branding Footer ─────────────────────
                Divider(color = MaterialTheme.colorScheme.outlineVariant)
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = "Made with \u2764 by Vasic Digital",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "\u00A9 ${java.util.Calendar.getInstance().get(java.util.Calendar.YEAR)} Vasic Digital. All rights reserved.",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f)
                )
                Spacer(modifier = Modifier.height(16.dp))
            }
        }
    }
}

/**
 * Discover Catalogizer API servers on the local network.
 * Probes common ports on the device's subnet gateway and nearby IPs.
 */
private suspend fun discoverServers(): List<String> = withContext(Dispatchers.IO) {
    val found = mutableListOf<String>()
    val ports = listOf(8080, 8081, 3000)

    // Detect subnet prefix from network interfaces
    val prefix = detectSubnetPrefix() ?: return@withContext found

    // Probe gateway (.1) and common server IPs
    val hosts = listOf(1) + (100..120).toList() + (200..220).toList()

    for (host in hosts) {
        for (port in ports) {
            val url = "http://$prefix.$host:$port"
            try {
                val conn = URL("$url/health").openConnection() as HttpURLConnection
                conn.connectTimeout = 1500
                conn.readTimeout = 1500
                conn.requestMethod = "GET"
                if (conn.responseCode == 200) {
                    found.add(url)
                }
                conn.disconnect()
            } catch (_: Exception) { }
        }
        if (found.isNotEmpty()) break // Stop after first subnet with results
    }
    found
}

private fun detectSubnetPrefix(): String? {
    return try {
        val interfaces = java.net.NetworkInterface.getNetworkInterfaces()
        while (interfaces.hasMoreElements()) {
            val iface = interfaces.nextElement()
            if (iface.isLoopback || !iface.isUp) continue
            if (iface.name.startsWith("tun") || iface.name.startsWith("vpn")) continue
            val addrs = iface.inetAddresses
            while (addrs.hasMoreElements()) {
                val addr = addrs.nextElement()
                if (addr is java.net.Inet4Address && !addr.isLoopbackAddress) {
                    val parts = addr.hostAddress?.split(".") ?: continue
                    if (parts.size == 4) return "${parts[0]}.${parts[1]}.${parts[2]}"
                }
            }
        }
        null
    } catch (_: Exception) { null }
}
