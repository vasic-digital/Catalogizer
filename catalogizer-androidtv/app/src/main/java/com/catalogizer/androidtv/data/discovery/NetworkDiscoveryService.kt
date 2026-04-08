package com.catalogizer.androidtv.data.discovery

import android.content.Context
import android.net.wifi.WifiManager
import android.util.Log
import com.catalogizer.androidtv.data.models.ServerEntry
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import org.json.JSONObject
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.HttpURLConnection
import java.net.InetAddress
import java.net.MulticastSocket
import java.net.URL

/**
 * Discovers Catalogizer API instances on the local network.
 *
 * Strategy (ordered by speed and reliability):
 * 1. UDP broadcast to the responder port (fastest — direct request/reply)
 * 2. UDP multicast listening with WiFi MulticastLock (passive, may be blocked)
 * 3. HTTP probe on common ports across the LAN subnet (most reliable fallback)
 */
class NetworkDiscoveryService(private val context: Context? = null) {

    companion object {
        private const val TAG = "Discovery"
        private const val MULTICAST_GROUP = "239.42.42.42"
        private const val MULTICAST_PORT = 42069
        private const val RESPONDER_PORT = 19820
        private const val DISCOVERY_MESSAGE = "CATALOGIZER_DISCOVER"
        // Default timeout for HTTP probes - increased from 2000ms to 5000ms
        // to accommodate slower servers or networks with higher latency.
        // Can be overridden via SettingsRepository for custom deployments.
        private const val DEFAULT_HTTP_PROBE_TIMEOUT_MS = 5000
        private const val MAX_RETRIES = 2
        private const val RETRY_DELAY_MS = 500L
        private val COMMON_PORTS = listOf(8080, 8081, 8082, 80)
    }

    /**
     * Fastest discovery method: send a UDP broadcast with "CATALOGIZER_DISCOVER"
     * to port 19820. The catalog-api Responder replies with JSON ServiceInfo
     * directly to the sender. Works on any network that allows local broadcast.
     */
    suspend fun discoverViaBroadcast(timeoutMs: Long = 3000L): List<ServerEntry> =
        withContext(Dispatchers.IO) {
            val results = mutableMapOf<String, ServerEntry>()
            var socket: DatagramSocket? = null
            try {
                socket = DatagramSocket()
                socket.broadcast = true
                socket.soTimeout = 500

                // Send discovery request to the broadcast address on responder port
                val message = DISCOVERY_MESSAGE.toByteArray()
                val broadcastAddr = InetAddress.getByName("255.255.255.255")
                val packet = DatagramPacket(message, message.size, broadcastAddr, RESPONDER_PORT)
                socket.send(packet)
                Log.d(TAG, "Sent UDP broadcast discovery to 255.255.255.255:$RESPONDER_PORT")

                // Also try subnet-directed broadcast if we can detect the subnet
                val prefix = detectSubnetPrefix()
                if (prefix != null) {
                    val subnetBroadcast = InetAddress.getByName("$prefix.255")
                    val subnetPacket = DatagramPacket(message, message.size, subnetBroadcast, RESPONDER_PORT)
                    socket.send(subnetPacket)
                    Log.d(TAG, "Sent UDP broadcast discovery to $prefix.255:$RESPONDER_PORT")
                }

                val deadline = System.currentTimeMillis() + timeoutMs
                val buf = ByteArray(4096)

                while (System.currentTimeMillis() < deadline) {
                    try {
                        val response = DatagramPacket(buf, buf.size)
                        val remaining = deadline - System.currentTimeMillis()
                        if (remaining <= 0) break
                        socket.soTimeout = remaining.coerceAtMost(500).toInt()
                        socket.receive(response)
                        val json = String(response.data, 0, response.length)
                        val entry = parseServiceInfoJson(json)
                        if (entry != null) {
                            val key = entry.url
                            results[key] = entry
                            Log.d(TAG, "UDP broadcast discovered: ${entry.url} (${entry.name})")
                        }
                    } catch (_: java.net.SocketTimeoutException) {
                        // Expected: socket timeout means no more broadcasts received
                    }
                }
            } catch (e: Exception) {
                Log.w(TAG, "UDP broadcast discovery failed: ${e.message}")
            } finally {
                try {
                    socket?.close()
                } catch (_: Exception) { }
            }
            results.values.toList()
        }

    /**
     * HTTP probe on common ports across LAN subnet.
     * Tries /discovery endpoint on each IP in the subnet.
     * This works reliably on all networks (no multicast/broadcast needed).
     */
    suspend fun discoverViaHttpProbe(timeoutMs: Long = 10000L): List<ServerEntry> = coroutineScope {
        val prefix = detectSubnetPrefix()
        if (prefix == null) {
            Log.w(TAG, "Could not detect subnet prefix")
            return@coroutineScope emptyList()
        }
        Log.d(TAG, "Scanning subnet $prefix.0/24 on ports $COMMON_PORTS")

        val results = mutableListOf<ServerEntry>()

        // Scan common IPs first (.1, .100-120, .200-254) then the rest
        val priorityHosts = (listOf(1) + (100..120).toList() + (200..254).toList() + (2..99).toList() + (121..199).toList())

        val jobs = priorityHosts.flatMap { host ->
            COMMON_PORTS.map { port ->
                async(Dispatchers.IO) {
                    withTimeoutOrNull(DEFAULT_HTTP_PROBE_TIMEOUT_MS.toLong() * 2) {
                        probeServer("http://$prefix.$host:$port")
                    }
                }
            }
        }

        // Process results as they come in, with overall timeout
        try {
            withTimeoutOrNull(timeoutMs) {
                jobs.awaitAll().filterNotNull().forEach { results.add(it) }
            }
        } catch (e: Exception) {
            Log.w(TAG, "HTTP probe discovery interrupted: ${e.message}")
        }

        Log.d(TAG, "HTTP probe found ${results.size} servers")
        results
    }

    /**
     * Probe a single URL to check if it's a Catalogizer API.
     */
    /**
     * Probes a single URL to check if it's a Catalogizer API.
     * 
     * @param baseUrl The base URL to probe (e.g., "http://192.168.1.100:8080")
     * @param timeoutMs Timeout in milliseconds (default: 5000ms)
     * @param retryCount Number of retries on failure (default: 1)
     * @return ServerEntry if successful, null otherwise
     */
    suspend fun probeServer(
        baseUrl: String, 
        timeoutMs: Int = DEFAULT_HTTP_PROBE_TIMEOUT_MS,
        retryCount: Int = MAX_RETRIES
    ): ServerEntry? = withContext(Dispatchers.IO) {
        val safeBaseUrl = baseUrl.trimEnd('/')
        
        // Retry loop for reliability
        repeat(retryCount + 1) { attempt ->
            var connection: HttpURLConnection? = null
            try {
                val url = URL("$safeBaseUrl/discovery")
                connection = url.openConnection() as HttpURLConnection
                connection.connectTimeout = timeoutMs
                connection.readTimeout = timeoutMs
                connection.requestMethod = "GET"
                // Disable following redirects for faster failure detection
                connection.instanceFollowRedirects = false

            val responseCode = connection.responseCode
            if (responseCode == 200) {
                val body = connection.inputStream.bufferedReader().use { it.readText() }
                val obj = JSONObject(body)
                if (obj.optString("service") == "catalogizer-api") {
                    val version = obj.optString("version", "?")
                    val host = obj.optString("host", "")
                    val port = obj.optInt("port", 8080)
                    val protocol = obj.optString("protocol", "http")
                    val name = obj.optString("name", "Catalogizer API")
                    // Use the actual host IP from the discovery response
                    // so the URL shows the real server address, not localhost
                    val resolvedUrl = if (host.isNotBlank() && host != "0.0.0.0") {
                        "$protocol://$host:$port"
                    } else {
                        safeBaseUrl
                    }
                    return@withContext ServerEntry(
                        url = resolvedUrl,
                        name = "$name v$version",
                        isDiscovered = true,
                        lastConnected = System.currentTimeMillis()
                    )
                }
            }
        } catch (e: Exception) {
            Log.d(TAG, "HTTP probe failed for $safeBaseUrl: ${e.message}")
        } finally {
            try {
                connection?.disconnect()
            } catch (_: Exception) { }
        }
        null
    }

    /**
     * UDP multicast discovery with WiFi MulticastLock.
     * Listens for periodic announcements from catalog-api on 239.42.42.42:42069.
     * Requires android.permission.CHANGE_WIFI_MULTICAST_STATE.
     */
    suspend fun discoverViaMulticast(timeoutMs: Long = 5000L): List<ServerEntry> =
        withContext(Dispatchers.IO) {
            val results = mutableMapOf<String, ServerEntry>()
            var multicastLock: WifiManager.MulticastLock? = null
            var socket: MulticastSocket? = null

            try {
                // Acquire multicast lock (Android blocks multicast by default for power saving)
                val wifiManager = context?.applicationContext?.getSystemService(Context.WIFI_SERVICE) as? WifiManager
                multicastLock = wifiManager?.createMulticastLock("catalogizer_discovery")
                multicastLock?.setReferenceCounted(true)
                multicastLock?.acquire()
                Log.d(TAG, "MulticastLock acquired")

                val group = InetAddress.getByName(MULTICAST_GROUP)
                socket = MulticastSocket(MULTICAST_PORT)
                socket.joinGroup(group)
                socket.soTimeout = 500

                val deadline = System.currentTimeMillis() + timeoutMs
                val buf = ByteArray(4096)

                while (System.currentTimeMillis() < deadline) {
                    try {
                        val packet = DatagramPacket(buf, buf.size)
                        socket.receive(packet)
                        val json = String(packet.data, 0, packet.length)
                        val entry = parseServiceInfoJson(json)
                        if (entry != null) {
                            results[entry.url] = entry
                            Log.d(TAG, "Multicast discovered: ${entry.url}")
                        }
                    } catch (_: java.net.SocketTimeoutException) {
                        // Expected: socket timeout for multicast receive
                    }
                }

                try {
                    socket.leaveGroup(group)
                } catch (_: Exception) { }
            } catch (e: Exception) {
                Log.w(TAG, "Multicast discovery failed: ${e.message}")
            } finally {
                try {
                    socket?.close()
                } catch (_: Exception) { }
                try {
                    multicastLock?.release()
                } catch (_: Exception) { }
                Log.d(TAG, "MulticastLock released")
            }
            results.values.toList()
        }

    /**
     * Full discovery: tries all methods in order of speed, returning as soon as
     * any method finds at least one server.
     *
     * Order:
     * 1. Localhost probe (instant — works when ADB reverse proxy is active)
     * 2. UDP broadcast (fast — direct request/reply, ~1-2s)
     * 3. UDP multicast (passive listen for announcements, ~3s)
     * 4. HTTP subnet probe (slow but reliable fallback, scans entire /24)
     */
    suspend fun discoverAll(timeoutMs: Long = 12000L): List<ServerEntry> {
        // 1. Try localhost first (ADB reverse proxy scenario)
        val localhostResult = probeServer("http://localhost:8080")
        if (localhostResult != null) {
            Log.d(TAG, "Found server via localhost probe")
            return listOf(localhostResult)
        }

        // 2. Try UDP broadcast (fastest network discovery — 2 seconds)
        val broadcastResults = discoverViaBroadcast(2000L)
        if (broadcastResults.isNotEmpty()) {
            Log.d(TAG, "Found ${broadcastResults.size} via UDP broadcast")
            return broadcastResults
        }

        // 3. Try multicast (3 seconds)
        val multicastResults = discoverViaMulticast(3000L)
        if (multicastResults.isNotEmpty()) {
            Log.d(TAG, "Found ${multicastResults.size} via multicast")
            return multicastResults
        }

        // 4. Fallback: HTTP probe (more reliable, scans subnet)
        Log.d(TAG, "No servers found via broadcast/multicast, trying HTTP probe...")
        return discoverViaHttpProbe(timeoutMs)
    }

    /**
     * Parse a JSON service info response (from either multicast announce or
     * broadcast responder reply) into a [ServerEntry].
     */
    private fun parseServiceInfoJson(json: String): ServerEntry? {
        return try {
            val obj = JSONObject(json)
            val type = obj.optString("type", "")
            if (type != "catalogizer-announce") return null

            val host = obj.optString("host", "")
            val port = obj.optInt("port", 0)
            if (host.isBlank() || port <= 0) return null

            val protocol = obj.optString("protocol", "http")
            val version = obj.optString("version", "?")
            val name = obj.optString("name", "Catalogizer API")

            ServerEntry(
                url = "$protocol://$host:$port",
                name = "$name v$version",
                isDiscovered = true,
                lastConnected = System.currentTimeMillis()
            )
        } catch (e: Exception) {
            Log.w(TAG, "Failed to parse service info JSON: ${e.message}")
            null
        }
    }

    private fun detectSubnetPrefix(): String? {
        return try {
            val interfaces = java.net.NetworkInterface.getNetworkInterfaces()
            while (interfaces.hasMoreElements()) {
                val iface = interfaces.nextElement()
                if (iface.isLoopback || !iface.isUp) continue
                if (iface.name.startsWith("tun") || iface.name.startsWith("vpn")) continue // Skip VPN
                val addrs = iface.inetAddresses
                while (addrs.hasMoreElements()) {
                    val addr = addrs.nextElement()
                    if (addr is java.net.Inet4Address && !addr.isLoopbackAddress) {
                        val parts = addr.hostAddress?.split(".") ?: continue
                        if (parts.size == 4 && parts[0] != "10") { // Prefer non-10.x (LAN over VPN)
                            return "${parts[0]}.${parts[1]}.${parts[2]}"
                        }
                    }
                }
            }
            // Second pass: accept any non-loopback
            val interfaces2 = java.net.NetworkInterface.getNetworkInterfaces()
            while (interfaces2.hasMoreElements()) {
                val iface = interfaces2.nextElement()
                if (iface.isLoopback || !iface.isUp) continue
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
}
