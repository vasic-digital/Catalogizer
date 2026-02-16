package com.catalogizer.androidtv.data.sync

import android.content.Intent
import org.junit.After
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import io.mockk.*

class SyncServiceTest {

    private lateinit var syncService: SyncService

    @Before
    fun setup() {
        syncService = SyncService()
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    @Test
    fun `onBind should return null`() {
        val intent = mockk<Intent>(relaxed = true)

        val binder = syncService.onBind(intent)

        assertNull(binder)
    }

    @Test
    fun `onStartCommand with SYNC_NOW action should return START_STICKY`() {
        val intent = mockk<Intent>(relaxed = true)
        every { intent.action } returns "SYNC_NOW"

        val result = syncService.onStartCommand(intent, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `onStartCommand with SCHEDULED_SYNC action should return START_STICKY`() {
        val intent = mockk<Intent>(relaxed = true)
        every { intent.action } returns "SCHEDULED_SYNC"

        val result = syncService.onStartCommand(intent, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `onStartCommand with null action should return START_STICKY`() {
        val intent = mockk<Intent>(relaxed = true)
        every { intent.action } returns null

        val result = syncService.onStartCommand(intent, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `onStartCommand with null intent should return START_STICKY`() {
        val result = syncService.onStartCommand(null, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `onStartCommand with unknown action should return START_STICKY`() {
        val intent = mockk<Intent>(relaxed = true)
        every { intent.action } returns "UNKNOWN_ACTION"

        val result = syncService.onStartCommand(intent, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `service can be instantiated`() {
        val service = SyncService()

        assertNotNull(service)
    }

    @Test
    fun `onBind with null intent should return null`() {
        val binder = syncService.onBind(null)

        assertNull(binder)
    }

    @Test
    fun `onStartCommand handles different start IDs`() {
        val intent = mockk<Intent>(relaxed = true)
        every { intent.action } returns "SYNC_NOW"

        val result1 = syncService.onStartCommand(intent, 0, 1)
        val result2 = syncService.onStartCommand(intent, 0, 2)
        val result3 = syncService.onStartCommand(intent, 0, 99)

        assertEquals(android.app.Service.START_STICKY, result1)
        assertEquals(android.app.Service.START_STICKY, result2)
        assertEquals(android.app.Service.START_STICKY, result3)
    }
}
