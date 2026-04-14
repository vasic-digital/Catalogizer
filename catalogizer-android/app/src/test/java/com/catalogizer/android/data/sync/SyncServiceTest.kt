package com.catalogizer.android.data.sync

import android.content.Context
import android.content.Intent
import com.catalogizer.android.data.sync.SyncService.SyncState
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.Robolectric
import org.robolectric.RobolectricTestRunner
import org.robolectric.RuntimeEnvironment
import org.robolectric.annotation.Config

@OptIn(ExperimentalCoroutinesApi::class)
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class SyncServiceTest {

    private lateinit var context: Context

    @Before
    fun setup() {
        context = RuntimeEnvironment.getApplication()
    }

    // --- SyncState sealed class ---

    @Test
    fun `SyncState Idle is a singleton`() {
        val s1 = SyncState.Idle
        val s2 = SyncState.Idle
        assertSame(s1, s2)
    }

    @Test
    fun `SyncState Completed is a singleton`() {
        val s1 = SyncState.Completed
        val s2 = SyncState.Completed
        assertSame(s1, s2)
    }

    @Test
    fun `SyncState InProgress holds progress value`() {
        val state = SyncState.InProgress(50)
        assertEquals(50, state.progress)
    }

    @Test
    fun `SyncState InProgress at zero percent`() {
        val state = SyncState.InProgress(0)
        assertEquals(0, state.progress)
    }

    @Test
    fun `SyncState InProgress at 100 percent`() {
        val state = SyncState.InProgress(100)
        assertEquals(100, state.progress)
    }

    @Test
    fun `SyncState InProgress equality`() {
        val s1 = SyncState.InProgress(25)
        val s2 = SyncState.InProgress(25)
        assertEquals(s1, s2)
    }

    @Test
    fun `SyncState InProgress inequality`() {
        val s1 = SyncState.InProgress(25)
        val s2 = SyncState.InProgress(75)
        assertNotEquals(s1, s2)
    }

    @Test
    fun `SyncState Error holds message`() {
        val state = SyncState.Error("Network timeout")
        assertEquals("Network timeout", state.message)
    }

    @Test
    fun `SyncState Error with empty message`() {
        val state = SyncState.Error("")
        assertEquals("", state.message)
    }

    @Test
    fun `SyncState Error equality`() {
        val s1 = SyncState.Error("error")
        val s2 = SyncState.Error("error")
        assertEquals(s1, s2)
    }

    @Test
    fun `SyncState Error inequality`() {
        val s1 = SyncState.Error("error1")
        val s2 = SyncState.Error("error2")
        assertNotEquals(s1, s2)
    }

    @Test
    fun `SyncState subtypes are distinct`() {
        val idle: SyncState = SyncState.Idle
        val inProgress: SyncState = SyncState.InProgress(50)
        val completed: SyncState = SyncState.Completed
        val error: SyncState = SyncState.Error("fail")

        assertNotEquals(idle, inProgress)
        assertNotEquals(idle, completed)
        assertNotEquals(idle, error)
        assertNotEquals(inProgress, completed)
        assertNotEquals(inProgress, error)
        assertNotEquals(completed, error)
    }

    @Test
    fun `SyncState can be checked with is operator`() {
        val idle: SyncState = SyncState.Idle
        val inProgress: SyncState = SyncState.InProgress(50)
        val completed: SyncState = SyncState.Completed
        val error: SyncState = SyncState.Error("fail")

        assertTrue(idle is SyncState.Idle)
        assertTrue(inProgress is SyncState.InProgress)
        assertTrue(completed is SyncState.Completed)
        assertTrue(error is SyncState.Error)
    }

    @Test
    fun `SyncState InProgress copy updates progress`() {
        val original = SyncState.InProgress(25)
        val updated = original.copy(progress = 75)
        assertEquals(75, updated.progress)
    }

    @Test
    fun `SyncState Error copy updates message`() {
        val original = SyncState.Error("first error")
        val updated = original.copy(message = "second error")
        assertEquals("second error", updated.message)
    }

    // --- Companion object constants ---

    @Test
    fun `ACTION_START_SYNC has correct value`() {
        assertEquals("com.catalogizer.android.action.START_SYNC", SyncService.ACTION_START_SYNC)
    }

    @Test
    fun `ACTION_STOP_SYNC has correct value`() {
        assertEquals("com.catalogizer.android.action.STOP_SYNC", SyncService.ACTION_STOP_SYNC)
    }

    @Test
    fun `CHANNEL_ID has correct value`() {
        assertEquals("sync_service_channel", SyncService.CHANNEL_ID)
    }

    @Test
    fun `NOTIFICATION_ID has correct value`() {
        assertEquals(1001, SyncService.NOTIFICATION_ID)
    }

    // --- Service lifecycle ---

    @Test
    fun `initial syncState is Idle`() = runTest {
        val service = Robolectric.buildService(SyncService::class.java).create().get()

        val state = service.syncState.first()

        assertTrue(state is SyncState.Idle)
    }

    @Test
    fun `onBind returns LocalBinder`() {
        val service = Robolectric.buildService(SyncService::class.java).create().get()
        val intent = Intent()

        val binder = service.onBind(intent)

        assertNotNull(binder)
        assertTrue(binder is SyncService.LocalBinder)
    }

    @Test
    fun `LocalBinder getService returns the service instance`() {
        val service = Robolectric.buildService(SyncService::class.java).create().get()
        val intent = Intent()
        val binder = service.onBind(intent) as SyncService.LocalBinder

        val retrievedService = binder.getService()

        assertSame(service, retrievedService)
    }

    @Test
    fun `onStartCommand returns START_STICKY`() {
        val service = Robolectric.buildService(SyncService::class.java).create().get()
        val intent = Intent()

        val result = service.onStartCommand(intent, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `onStartCommand with null intent returns START_STICKY`() {
        val service = Robolectric.buildService(SyncService::class.java).create().get()

        val result = service.onStartCommand(null, 0, 1)

        assertEquals(android.app.Service.START_STICKY, result)
    }

    @Test
    fun `onStartCommand with STOP_SYNC action resets state to Idle`() = runTest {
        val service = Robolectric.buildService(SyncService::class.java).create().get()
        val intent = Intent().apply {
            action = SyncService.ACTION_STOP_SYNC
        }

        service.onStartCommand(intent, 0, 1)

        val state = service.syncState.first()
        assertTrue(state is SyncState.Idle)
    }

    // --- Static start/stop helpers ---

    @Test
    fun `start creates intent with START_SYNC action`() {
        // Verifying the static method does not throw
        SyncService.start(context)
    }

    @Test
    fun `stop creates intent with STOP_SYNC action`() {
        // Verifying the static method does not throw
        SyncService.stop(context)
    }

    @Test
    fun `service can be destroyed without error`() {
        val controller = Robolectric.buildService(SyncService::class.java).create()
        val service = controller.get()

        // Should not throw, cancels the coroutine scope
        controller.destroy()
    }
}
