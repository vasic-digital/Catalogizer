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
import java.net.HttpURLConnection
import java.net.InetAddress
import java.net.MulticastSocket
import java.net.URL

/**
 * Discovers Catalogizer API instances on the local network.
 *
 * Strategy:
 * 1. HTTP probe on common ports across the LAN subnet (most reliable)
 * 2. UDP multicast listening with WiFi MulticastLock (may be blocked by router)
 */
class NetworkDiscoveryService(private val context: Context? = null) {

    companion object {
        private const val TAG = "Discovery"
        private const val MULTICAST_GROUP = "239.42.42.42"
        private const val MULTICAST_PORT = 42069
        private const val HTTP_PROBE_TIMEOUT_MS = 2000
        private val COMMON_PORTS = listOf(8080, 8081, 8082, 80)
    }

    /**
     * Primary discovery: HTTP probe on common ports across LAN subnet.
     * Tries /discovery endpoint on each IP in the subnet.
     * This works reliably on all networks (no multicast needed).
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
                    withTimeoutOrNull(HTTP_PROBE_TIMEOUT_MS.toLong()) {
                        probeServer("http://$prefix.$host:$port")
                    }
                }
            }
        }

        // Process results as they come in, with overall timeout
        withTimeoutOrNull(timeoutMs) {
            jobs.awaitAll().filterNotNull().forEach { results.add(it) }
        }

        Log.d(TAG, "HTTP probe found ${results.size} servers")
        results
    }

    /**
     * Probe a single URL to check if it's a Catalogizer API.
     */
    suspend fun probeServer(baseUrl: String): ServerEntry? = withContext(Dispatchers.IO) {
        try {
            val url = URL("${baseUrl.trimEnd('/')}/discovery")
            val conn = url.openConnection() as HttpURLConnection
            conn.connectTimeout = HTTP_PROBE_TIMEOUT_MS
            conn.readTimeout = HTTP_PROBE_TIMEOUT_MS
            conn.requestMethod = "GET"

            if (conn.responseCode == 200) {
                val body = conn.inputStream.bufferedReader().readText()
                val obj = JSONObject(body)
                if (obj.optString("service") == "catalogizer-api") {
                    val version = obj.optString("version", "?")
                    return@withContext ServerEntry(
                        url = baseUrl.trimEnd('/'),
                        name = "Catalogizer v$version",
                        isDiscovered = true,
                        lastConnected = System.currentTimeMillis()
                    )
                }
            }
            conn.disconnect()
        } catch (_: Exception) { }
        null
    }

    /**
     * UDP multicast discovery with WiFi MulticastLock.
     * Requires android.permission.CHANGE_WIFI_MULTICAST_STATE.
     */
    suspend fun discoverViaMulticast(timeoutMs: Long = 5000L): List<ServerEntry> =
        withContext(Dispatchers.IO) {
            val results = mutableMapOf<String, ServerEntry>()
            var multicastLock: WifiManager.MulticastLock? = null

            try {
                // Acquire multicast lock (Android blocks multicast by default for power saving)
                val wifiManager = context?.applicationContext?.getSystemService(Context.WIFI_SERVICE) as? WifiManager
                multicastLock = wifiManager?.createMulticastLock("catalogizer_discovery")
                multicastLock?.setReferenceCounted(true)
                multicastLock?.acquire()
                Log.d(TAG, "MulticastLock acquired")

                val group = InetAddress.getByName(MULTICAST_GROUP)
                val socket = MulticastSocket(MULTICAST_PORT)
                socket.joinGroup(group)
                socket.soTimeout = 500

                val deadline = System.currentTimeMillis() + timeoutMs
                val buf = ByteArray(4096)

                while (System.currentTimeMillis() < deadline) {
                    try {
                        val packet = DatagramPacket(buf, buf.size)
                        socket.receive(packet)
                        val json = String(packet.data, 0, packet.length)
                        val obj = JSONObject(json)

                        if (obj.optString("type") == "catalogizer-announce") {
                            val host = obj.getString("host")
                            val port = obj.getInt("port")
                            val key = "$host:$port"
                            results[key] = ServerEntry(
                                url = "http://$host:$port",
                                name = obj.optString("name", "Catalogizer API"),
                                isDiscovered = true,
                                lastConnected = System.currentTimeMillis()
                            )
                            Log.d(TAG, "Multicast discovered: $key")
                        }
                    } catch (_: java.net.SocketTimeoutException) { }
                }

                socket.leaveGroup(group)
                socket.close()
            } catch (e: Exception) {
                Log.w(TAG, "Multicast discovery failed: ${e.message}")
            } finally {
                multicastLock?.release()
                Log.d(TAG, "MulticastLock released")
            }
            results.values.toList()
        }

    /**
     * Full discovery: try multicast first (fast if supported), then HTTP probe.
     */
    suspend fun discoverAll(timeoutMs: Long = 12000L): List<ServerEntry> {
        // Try multicast first (3 seconds)
        val multicastResults = discoverViaMulticast(3000L)
        if (multicastResults.isNotEmpty()) {
            Log.d(TAG, "Found ${multicastResults.size} via multicast")
            return multicastResults
        }

        // Fallback: HTTP probe (more reliable, scans subnet)
        Log.d(TAG, "Multicast found nothing, trying HTTP probe...")
        return discoverViaHttpProbe(timeoutMs)
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
