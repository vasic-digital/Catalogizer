package com.catalogizer.androidtv.data.discovery

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
 * Discovers Catalogizer API instances on the local network via:
 * 1. UDP multicast listening (primary — matches backend's broadcast announcer)
 * 2. HTTP probe on common ports (fallback — tries /discovery on each candidate)
 */
class NetworkDiscoveryService {

    companion object {
        private const val MULTICAST_GROUP = "239.42.42.42"
        private const val MULTICAST_PORT = 42069
        private const val MULTICAST_TIMEOUT_MS = 5000L
        private const val HTTP_PROBE_TIMEOUT_MS = 3000
        private val COMMON_PORTS = listOf(8080, 8081, 8082, 3000, 80, 443)
    }

    /**
     * Discover API instances via UDP multicast.
     * Listens for announcements from the backend's broadcast.Announcer.
     */
    suspend fun discoverViaMulticast(timeoutMs: Long = MULTICAST_TIMEOUT_MS): List<ServerEntry> =
        withContext(Dispatchers.IO) {
            val results = mutableMapOf<String, ServerEntry>()
            try {
                val group = InetAddress.getByName(MULTICAST_GROUP)
                val socket = MulticastSocket(MULTICAST_PORT)
                socket.joinGroup(group)
                socket.soTimeout = 500 // Short timeout for responsive context checking

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
                        }
                    } catch (_: java.net.SocketTimeoutException) {
                        // Normal — continue polling until deadline
                    }
                }

                socket.leaveGroup(group)
                socket.close()
            } catch (e: Exception) {
                // Multicast may not be available on all networks
                android.util.Log.w("Discovery", "Multicast discovery failed: ${e.message}")
            }
            results.values.toList()
        }

    /**
     * Probe a specific URL to check if it's a Catalogizer API.
     * Returns a ServerEntry if the /discovery endpoint responds correctly.
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
                    return@withContext ServerEntry(
                        url = baseUrl.trimEnd('/'),
                        name = "Catalogizer v${obj.optString("version", "?")}",
                        isDiscovered = true,
                        lastConnected = System.currentTimeMillis()
                    )
                }
            }
            conn.disconnect()
        } catch (_: Exception) {
            // Connection failed — server not available at this URL
        }
        null
    }

    /**
     * Scan the local network subnet on common ports.
     * Tries to find Catalogizer API instances by probing /discovery.
     */
    suspend fun discoverViaHttpProbe(
        subnetPrefix: String? = null,
        ports: List<Int> = COMMON_PORTS
    ): List<ServerEntry> = coroutineScope {
        val prefix = subnetPrefix ?: detectSubnetPrefix() ?: return@coroutineScope emptyList()
        val results = mutableListOf<ServerEntry>()

        // Scan .1 to .254 on each port (limited concurrency)
        val jobs = (1..254).flatMap { host ->
            ports.map { port ->
                async(Dispatchers.IO) {
                    withTimeoutOrNull(HTTP_PROBE_TIMEOUT_MS.toLong()) {
                        probeServer("http://$prefix.$host:$port")
                    }
                }
            }
        }

        jobs.awaitAll().filterNotNull().forEach { results.add(it) }
        results
    }

    /**
     * Full discovery: multicast first (fast), then HTTP probe as fallback.
     */
    suspend fun discoverAll(): List<ServerEntry> {
        // Try multicast first (5 seconds)
        val multicastResults = discoverViaMulticast()
        if (multicastResults.isNotEmpty()) return multicastResults

        // Fallback: HTTP probe (slower, scans common ports on subnet)
        return discoverViaHttpProbe()
    }

    private fun detectSubnetPrefix(): String? {
        return try {
            val interfaces = java.net.NetworkInterface.getNetworkInterfaces()
            while (interfaces.hasMoreElements()) {
                val iface = interfaces.nextElement()
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
        } catch (_: Exception) {
            null
        }
    }
}
