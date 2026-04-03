package com.catalogizer.android.data.repository

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import org.java_websocket.client.WebSocketClient
import org.java_websocket.handshake.ServerHandshake
import java.net.URI

/**
 * Manages a WebSocket connection to the Catalogizer API for real-time event streaming.
 * Automatically subscribes to media and system update channels on connection and
 * exposes events as a [StateFlow] of [WebSocketEvent].
 */
class WebSocketRepository(
    private val baseUrl: String,
    private val tokenProvider: () -> String?
) {
    sealed class WebSocketEvent {
        data class Connected(val serverInfo: String) : WebSocketEvent()
        data object Disconnected : WebSocketEvent()
        data class MessageReceived(val type: String, val payload: String) : WebSocketEvent()
        data class Error(val message: String) : WebSocketEvent()
    }

    private val _events = MutableStateFlow<WebSocketEvent>(WebSocketEvent.Disconnected)
    val events: StateFlow<WebSocketEvent> = _events.asStateFlow()

    private var client: WebSocketClient? = null

    fun connect() {
        val token = tokenProvider() ?: return
        val wsUrl = baseUrl.replace("http://", "ws://").replace("https://", "wss://")
        val uri = URI("$wsUrl/ws?token=$token")

        client?.close()
        client = object : WebSocketClient(uri) {
            override fun onOpen(handshakedata: ServerHandshake?) {
                _events.value = WebSocketEvent.Connected(handshakedata?.httpStatusMessage ?: "")
                send("""{"type":"subscribe","channel":"media_updates"}""")
                send("""{"type":"subscribe","channel":"system_updates"}""")
            }

            override fun onMessage(message: String?) {
                message?.let {
                    // Extract type from JSON simply
                    val typeMatch = Regex(""""type"\s*:\s*"([^"]+)"""").find(it)
                    val type = typeMatch?.groupValues?.getOrNull(1) ?: "unknown"
                    _events.value = WebSocketEvent.MessageReceived(type, it)
                }
            }

            override fun onClose(code: Int, reason: String?, remote: Boolean) {
                _events.value = WebSocketEvent.Disconnected
            }

            override fun onError(ex: Exception?) {
                _events.value = WebSocketEvent.Error(ex?.message ?: "Unknown WebSocket error")
            }
        }

        try {
            client?.connect()
        } catch (e: Exception) {
            _events.value = WebSocketEvent.Error(e.message ?: "Connection failed")
        }
    }

    fun disconnect() {
        client?.close()
        client = null
        _events.value = WebSocketEvent.Disconnected
    }
}
