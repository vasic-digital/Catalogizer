package com.catalogizer.android.data.remote

import com.catalogizer.android.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

class WebSocketEventTest {

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
    fun `SystemUpdate event holds correct data`() {
        val event = WebSocketEvent.SystemUpdate(
            action = "status_change",
            component = "smb_scanner",
            status = "healthy",
            message = "All sources connected"
        )

        assertEquals("status_change", event.action)
        assertEquals("smb_scanner", event.component)
        assertEquals("healthy", event.status)
        assertEquals("All sources connected", event.message)
    }

    @Test
    fun `AnalysisComplete event holds correct data`() {
        val event = WebSocketEvent.AnalysisComplete(
            analysisId = "analysis-123",
            itemsProcessed = 100,
            newItems = 25,
            updatedItems = 10
        )

        assertEquals("analysis-123", event.analysisId)
        assertEquals(100, event.itemsProcessed)
        assertEquals(25, event.newItems)
        assertEquals(10, event.updatedItems)
    }

    @Test
    fun `Notification event holds correct data`() {
        val event = WebSocketEvent.Notification(
            type = "info",
            title = "Scan Complete",
            message = "Found 50 new items",
            level = "info"
        )

        assertEquals("info", event.type)
        assertEquals("Scan Complete", event.title)
        assertEquals("Found 50 new items", event.message)
        assertEquals("info", event.level)
    }

    @Test
    fun `Connected event is singleton`() {
        val event1 = WebSocketEvent.Connected
        val event2 = WebSocketEvent.Connected

        assertSame(event1, event2)
    }

    @Test
    fun `Disconnected event is singleton`() {
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
    fun `WebSocketEvent subtypes are distinct`() {
        val events: List<WebSocketEvent> = listOf(
            WebSocketEvent.Connected,
            WebSocketEvent.Disconnected,
            WebSocketEvent.Error("error"),
            WebSocketEvent.MediaUpdate("create", 1L, null),
            WebSocketEvent.SystemUpdate("update", "comp", "ok", null),
            WebSocketEvent.AnalysisComplete("id", 0, 0, 0),
            WebSocketEvent.Notification("info", "title", "msg", "info")
        )

        assertEquals(7, events.size)
        assertTrue(events[0] is WebSocketEvent.Connected)
        assertTrue(events[1] is WebSocketEvent.Disconnected)
        assertTrue(events[2] is WebSocketEvent.Error)
        assertTrue(events[3] is WebSocketEvent.MediaUpdate)
        assertTrue(events[4] is WebSocketEvent.SystemUpdate)
        assertTrue(events[5] is WebSocketEvent.AnalysisComplete)
        assertTrue(events[6] is WebSocketEvent.Notification)
    }
}
