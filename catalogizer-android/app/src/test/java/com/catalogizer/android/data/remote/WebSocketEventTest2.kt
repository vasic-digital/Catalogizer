package com.catalogizer.android.data.remote

import com.catalogizer.android.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

class WebSocketEventTest2 {

    @Test
    fun `MediaUpdate event holds correct data`() {
        val event = WebSocketEvent.MediaUpdate(
            action = "created",
            mediaId = 42L,
            media = null
        )
        assertEquals("created", event.action)
        assertEquals(42L, event.mediaId)
        assertNull(event.media)
    }

    @Test
    fun `MediaUpdate event with media item`() {
        val item = MediaItem(
            id = 1L, title = "Test", mediaType = "movie",
            directoryPath = "/path", createdAt = "2024-01-01", updatedAt = "2024-01-01"
        )
        val event = WebSocketEvent.MediaUpdate(
            action = "updated",
            mediaId = 1L,
            media = item
        )
        assertNotNull(event.media)
        assertEquals("Test", event.media?.title)
    }

    @Test
    fun `SystemUpdate event holds correct data`() {
        val event = WebSocketEvent.SystemUpdate(
            action = "restart",
            component = "scanner",
            status = "running",
            message = "Scan in progress"
        )
        assertEquals("restart", event.action)
        assertEquals("scanner", event.component)
        assertEquals("running", event.status)
        assertEquals("Scan in progress", event.message)
    }

    @Test
    fun `SystemUpdate event with null message`() {
        val event = WebSocketEvent.SystemUpdate(
            action = "stop",
            component = "cache",
            status = "stopped",
            message = null
        )
        assertNull(event.message)
    }

    @Test
    fun `AnalysisComplete event holds correct data`() {
        val event = WebSocketEvent.AnalysisComplete(
            analysisId = "scan-123",
            itemsProcessed = 1000,
            newItems = 50,
            updatedItems = 200
        )
        assertEquals("scan-123", event.analysisId)
        assertEquals(1000, event.itemsProcessed)
        assertEquals(50, event.newItems)
        assertEquals(200, event.updatedItems)
    }

    @Test
    fun `Notification event holds correct data`() {
        val event = WebSocketEvent.Notification(
            type = "info",
            title = "Scan Complete",
            message = "Found 50 new items",
            level = "success"
        )
        assertEquals("info", event.type)
        assertEquals("Scan Complete", event.title)
        assertEquals("Found 50 new items", event.message)
        assertEquals("success", event.level)
    }

    @Test
    fun `Connected is a singleton object`() {
        val event1 = WebSocketEvent.Connected
        val event2 = WebSocketEvent.Connected
        assertSame(event1, event2)
    }

    @Test
    fun `Disconnected is a singleton object`() {
        val event1 = WebSocketEvent.Disconnected
        val event2 = WebSocketEvent.Disconnected
        assertSame(event1, event2)
    }

    @Test
    fun `Error event holds message`() {
        val event = WebSocketEvent.Error("Connection lost")
        assertEquals("Connection lost", event.message)
    }

    @Test
    fun `all WebSocketEvent subtypes are instances of WebSocketEvent`() {
        val events: List<WebSocketEvent> = listOf(
            WebSocketEvent.MediaUpdate("a", 1L, null),
            WebSocketEvent.SystemUpdate("a", "b", "c", null),
            WebSocketEvent.AnalysisComplete("id", 0, 0, 0),
            WebSocketEvent.Notification("t", "ti", "m", "l"),
            WebSocketEvent.Connected,
            WebSocketEvent.Disconnected,
            WebSocketEvent.Error("err")
        )
        assertEquals(7, events.size)
    }
}
