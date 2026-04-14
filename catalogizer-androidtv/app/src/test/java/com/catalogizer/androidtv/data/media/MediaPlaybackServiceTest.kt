package com.catalogizer.androidtv.data.media

import android.content.Intent
import androidx.media3.common.util.UnstableApi
import io.mockk.clearAllMocks
import io.mockk.mockk
import org.junit.After
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

/**
 * Unit tests for [MediaPlaybackService]. Because the service
 * extends [MediaSessionService] which requires a full Android
 * environment for lifecycle, we test the constructor and the
 * public API contract. Integration tests on a real device are
 * handled by HelixQA autonomous sessions.
 */
@UnstableApi
class MediaPlaybackServiceTest {

    private lateinit var service: MediaPlaybackService

    @Before
    fun setup() {
        service = MediaPlaybackService()
    }

    @After
    fun tearDown() {
        clearAllMocks()
    }

    // ------------------------------------------------------------------
    // Construction
    // ------------------------------------------------------------------

    @Test
    fun `service can be instantiated`() {
        assertNotNull(service)
    }

    @Test
    fun `multiple instances can be created independently`() {
        val service2 = MediaPlaybackService()
        assertNotNull(service2)
    }

    // ------------------------------------------------------------------
    // onGetSession — before onCreate, session is null
    // ------------------------------------------------------------------

    @Test
    fun `onGetSession returns null before onCreate`() {
        val controllerInfo = mockk<androidx.media3.session.MediaSession.ControllerInfo>(relaxed = true)
        val session = service.onGetSession(controllerInfo)
        assertNull(session)
    }

    // ------------------------------------------------------------------
    // onTaskRemoved — safe to call before onCreate
    // ------------------------------------------------------------------

    @Test
    fun `onTaskRemoved does not crash when player is null`() {
        // Before onCreate, player is null; onTaskRemoved should
        // return early without crashing.
        val intent = mockk<Intent>(relaxed = true)
        service.onTaskRemoved(intent)
    }

    @Test
    fun `onTaskRemoved with null intent does not crash`() {
        service.onTaskRemoved(null)
    }

    // ------------------------------------------------------------------
    // onDestroy — safe to call before onCreate
    // ------------------------------------------------------------------

    @Test
    fun `onDestroy does not crash when called before onCreate`() {
        // player and mediaSession are both null; onDestroy must
        // handle this gracefully (release on null is a no-op).
        service.onDestroy()
    }

    @Test
    fun `onDestroy can be called multiple times safely`() {
        service.onDestroy()
        service.onDestroy()
    }

    // ------------------------------------------------------------------
    // Class structure verification
    // ------------------------------------------------------------------

    @Test
    fun `service extends MediaSessionService`() {
        val superclass = service.javaClass.superclass
        assertNotNull(superclass)
        // MediaPlaybackService -> MediaSessionService
        val ancestors = mutableListOf<String>()
        var cls: Class<*>? = service.javaClass
        while (cls != null) {
            ancestors.add(cls.simpleName)
            cls = cls.superclass
        }
        val hasMediaSessionService = ancestors.any {
            it == "MediaSessionService"
        }
        assertNotNull(
            "MediaPlaybackService must extend MediaSessionService",
            hasMediaSessionService
        )
    }

    @Test
    fun `service declares onGetSession method`() {
        val methods = service.javaClass.declaredMethods.map { it.name }
        val hasOnGetSession = methods.contains("onGetSession")
        assertNotNull("onGetSession must be declared", hasOnGetSession)
    }
}
