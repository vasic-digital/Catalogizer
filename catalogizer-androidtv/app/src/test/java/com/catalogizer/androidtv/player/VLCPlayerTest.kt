package com.catalogizer.androidtv.player

import android.content.Context
import io.mockk.clearAllMocks
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkConstructor
import io.mockk.mockkStatic
import io.mockk.unmockkConstructor
import io.mockk.unmockkStatic
import io.mockk.verify
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.After
import org.junit.AfterClass
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.BeforeClass
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.MediaPlayer

/**
 * Unit tests for [VLCPlayer]. Because LibVLC loads native .so files
 * that are unavailable in a JVM-only test environment, the
 * constructor's `initialize()` will catch the resulting exception
 * and set state to ERROR. Tests verify that the player handles
 * this gracefully and that all public API methods are safe to call
 * in the error / uninitialized state.
 *
 * Integration tests with a real LibVLC run on-device via HelixQA.
 */
@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class VLCPlayerTest {

    private lateinit var context: Context
    private lateinit var player: VLCPlayer

    companion object {
        private var originalSecurityManager: SecurityManager? = null

        @JvmStatic
        @BeforeClass
        fun installSecurityManager() {
            // LibVLC.loadLibraries() calls System.exit(1) when native
            // library loading fails. Intercept System.exit so the
            // SecurityException propagates to initialize()'s catch block.
            originalSecurityManager = System.getSecurityManager()
            System.setSecurityManager(object : SecurityManager() {
                override fun checkExit(status: Int) {
                    throw SecurityException("System.exit($status) blocked")
                }
                override fun checkPermission(perm: java.security.Permission?) {}
            })
        }

        @JvmStatic
        @AfterClass
        fun restoreSecurityManager() {
            System.setSecurityManager(originalSecurityManager)
        }
    }

    @Before
    fun setup() {
        context = mockk(relaxed = true)
        // In JVM tests, LibVLC's native libraries (built for Android)
        // are incompatible with the Linux host. loadLibraries() will
        // fail and call System.exit(1). The SecurityManager installed
        // above converts that into a SecurityException which is caught
        // by VLCPlayer.initialize(), leaving the player in ERROR state.
        player = VLCPlayer(context)
    }

    @After
    fun tearDown() {
        player.release()
        clearAllMocks()
    }

    // ------------------------------------------------------------------
    // Construction and initial state
    // ------------------------------------------------------------------

    @Test
    fun `player can be instantiated without crashing`() {
        assertNotNull(player)
    }

    @Test
    fun `initial state is ERROR when native libs unavailable`() = runTest {
        // In JVM test, LibVLC native init fails => ERROR state
        val state = player.playbackState.first()
        assertEquals(VLCPlayer.PlaybackState.ERROR, state)
    }

    @Test
    fun `initial isPlaying is false`() = runTest {
        val playing = player.isPlaying.first()
        assertFalse(playing)
    }

    @Test
    fun `initial currentPosition is 0`() = runTest {
        val pos = player.currentPosition.first()
        assertEquals(0L, pos)
    }

    @Test
    fun `initial duration is 0`() = runTest {
        val dur = player.duration.first()
        assertEquals(0L, dur)
    }

    @Test
    fun `initial volume is 100`() = runTest {
        val vol = player.volume.first()
        assertEquals(100, vol)
    }

    @Test
    fun `initial audioTracks is empty`() = runTest {
        val tracks = player.audioTracks.first()
        assertTrue(tracks.isEmpty())
    }

    @Test
    fun `initial subtitleTracks is empty`() = runTest {
        val tracks = player.subtitleTracks.first()
        assertTrue(tracks.isEmpty())
    }

    @Test
    fun `initial videoTracks is empty`() = runTest {
        val tracks = player.videoTracks.first()
        assertTrue(tracks.isEmpty())
    }

    @Test
    fun `initial currentAudioTrack is null`() = runTest {
        val track = player.currentAudioTrack.first()
        assertNull(track)
    }

    @Test
    fun `initial currentSubtitleTrack is null`() = runTest {
        val track = player.currentSubtitleTrack.first()
        assertNull(track)
    }

    // ------------------------------------------------------------------
    // play — when libVLC is null (native init failed)
    // ------------------------------------------------------------------

    @Test
    fun `play with null libVLC sets ERROR state`() = runTest {
        // attachView first so play() doesn't just defer
        val layout = mockk<org.videolan.libvlc.util.VLCVideoLayout>(relaxed = true)
        player.attachView(layout)

        player.play("http://example.com/video.mp4")

        val state = player.playbackState.first()
        assertEquals(VLCPlayer.PlaybackState.ERROR, state)
    }

    @Test
    fun `play without attachView defers URI`() = runTest {
        // play before attachView — should not crash, just store URI
        player.play("http://example.com/video.mp4")
        // State should remain as-is (ERROR from init failure)
        val state = player.playbackState.first()
        assertEquals(VLCPlayer.PlaybackState.ERROR, state)
    }

    // ------------------------------------------------------------------
    // pause, resume, stop — safe when mediaPlayer is null
    // ------------------------------------------------------------------

    @Test
    fun `pause does not crash when mediaPlayer is null`() {
        player.pause()
    }

    @Test
    fun `resume does not crash when mediaPlayer is null`() {
        player.resume()
    }

    @Test
    fun `stop sets STOPPED state and isPlaying false`() = runTest {
        player.stop()
        val state = player.playbackState.first()
        assertEquals(VLCPlayer.PlaybackState.STOPPED, state)
        assertFalse(player.isPlaying.first())
    }

    // ------------------------------------------------------------------
    // seek operations — safe when mediaPlayer is null
    // ------------------------------------------------------------------

    @Test
    fun `seek does not crash when mediaPlayer is null`() {
        player.seek(0.5f)
    }

    @Test
    fun `seekForward does not crash when mediaPlayer is null`() {
        player.seekForward(10000)
    }

    @Test
    fun `seekBackward does not crash when mediaPlayer is null`() {
        player.seekBackward(10000)
    }

    @Test
    fun `seekForward with default parameter does not crash`() {
        player.seekForward()
    }

    @Test
    fun `seekBackward with default parameter does not crash`() {
        player.seekBackward()
    }

    // ------------------------------------------------------------------
    // setSpeed — clamping behavior
    // ------------------------------------------------------------------

    @Test
    fun `setSpeed does not crash when mediaPlayer is null`() {
        player.setSpeed(2.0f)
    }

    @Test
    fun `setSpeed with value below 0_25 is safe`() {
        player.setSpeed(0.1f)
    }

    @Test
    fun `setSpeed with value above 4_0 is safe`() {
        player.setSpeed(10.0f)
    }

    // ------------------------------------------------------------------
    // setVolume — clamping and state update
    // ------------------------------------------------------------------

    @Test
    fun `setVolume updates volume StateFlow`() = runTest {
        player.setVolume(50)
        val vol = player.volume.first()
        assertEquals(50, vol)
    }

    @Test
    fun `setVolume clamps below 0 to 0`() = runTest {
        player.setVolume(-10)
        val vol = player.volume.first()
        assertEquals(0, vol)
    }

    @Test
    fun `setVolume clamps above 100 to 100`() = runTest {
        player.setVolume(200)
        val vol = player.volume.first()
        assertEquals(100, vol)
    }

    @Test
    fun `setVolume 0 is valid mute`() = runTest {
        player.setVolume(0)
        assertEquals(0, player.volume.first())
    }

    @Test
    fun `setVolume 100 is valid max`() = runTest {
        player.setVolume(100)
        assertEquals(100, player.volume.first())
    }

    // ------------------------------------------------------------------
    // Track selection — safe when mediaPlayer is null
    // ------------------------------------------------------------------

    @Test
    fun `setAudioTrack does not crash when mediaPlayer is null`() {
        player.setAudioTrack(1)
    }

    @Test
    fun `setSubtitleTrack does not crash when mediaPlayer is null`() {
        player.setSubtitleTrack(1)
    }

    @Test
    fun `setSubtitleTrack with -1 disables subtitles safely`() {
        player.setSubtitleTrack(-1)
    }

    // ------------------------------------------------------------------
    // Delay settings — safe when mediaPlayer is null
    // ------------------------------------------------------------------

    @Test
    fun `setSubtitleDelay does not crash when mediaPlayer is null`() {
        player.setSubtitleDelay(500L)
    }

    @Test
    fun `setAudioDelay does not crash when mediaPlayer is null`() {
        player.setAudioDelay(200L)
    }

    @Test
    fun `setSubtitleDelay with zero is valid`() {
        player.setSubtitleDelay(0L)
    }

    @Test
    fun `setAudioDelay with negative value is valid`() {
        player.setAudioDelay(-100L)
    }

    // ------------------------------------------------------------------
    // Aspect ratio and scale — safe when mediaPlayer is null
    // ------------------------------------------------------------------

    @Test
    fun `setAspectRatio does not crash when mediaPlayer is null`() {
        player.setAspectRatio("16:9")
    }

    @Test
    fun `setVideoScale does not crash when mediaPlayer is null`() {
        player.setVideoScale(1.5f)
    }

    // ------------------------------------------------------------------
    // takeScreenshot
    // ------------------------------------------------------------------

    @Test
    fun `takeScreenshot returns null when no video layout attached`() {
        val result = player.takeScreenshot()
        assertNull(result)
    }

    @Test
    fun `takeScreenshot returns null with attached layout but no player`() {
        val layout = mockk<org.videolan.libvlc.util.VLCVideoLayout>(relaxed = true)
        player.attachView(layout)
        val result = player.takeScreenshot()
        assertNull(result)
    }

    // ------------------------------------------------------------------
    // release
    // ------------------------------------------------------------------

    @Test
    fun `release resets state to IDLE`() = runTest {
        player.release()
        val state = player.playbackState.first()
        assertEquals(VLCPlayer.PlaybackState.IDLE, state)
    }

    @Test
    fun `release sets isPlaying to false`() = runTest {
        player.release()
        assertFalse(player.isPlaying.first())
    }

    @Test
    fun `release can be called multiple times safely`() = runTest {
        player.release()
        player.release()
        player.release()
        assertEquals(VLCPlayer.PlaybackState.IDLE, player.playbackState.first())
    }

    @Test
    fun `operations after release do not crash`() = runTest {
        player.release()
        player.pause()
        player.resume()
        player.stop()
        player.seek(0.5f)
        player.seekForward()
        player.seekBackward()
        player.setSpeed(1.5f)
        player.setVolume(50)
        player.setAudioTrack(0)
        player.setSubtitleTrack(-1)
        player.setSubtitleDelay(0)
        player.setAudioDelay(0)
        player.setAspectRatio("4:3")
        player.setVideoScale(1.0f)
        assertNull(player.takeScreenshot())
    }

    // ------------------------------------------------------------------
    // attachView
    // ------------------------------------------------------------------

    @Test
    fun `attachView does not crash when mediaPlayer is null`() {
        val layout = mockk<org.videolan.libvlc.util.VLCVideoLayout>(relaxed = true)
        player.attachView(layout)
    }

    @Test
    fun `attachView triggers deferred play when URI was pending`() = runTest {
        // Play first (deferred), then attach view
        player.play("http://example.com/test.mp4")
        val layout = mockk<org.videolan.libvlc.util.VLCVideoLayout>(relaxed = true)
        player.attachView(layout)
        // In JVM, play will set ERROR because libVLC is null
        val state = player.playbackState.first()
        assertEquals(VLCPlayer.PlaybackState.ERROR, state)
    }

    // ------------------------------------------------------------------
    // PlaybackState enum values
    // ------------------------------------------------------------------

    @Test
    fun `PlaybackState enum has all expected values`() {
        val states = VLCPlayer.PlaybackState.values()
        val expected = setOf("IDLE", "BUFFERING", "PLAYING", "PAUSED", "STOPPED", "COMPLETED", "ERROR")
        assertEquals(expected, states.map { it.name }.toSet())
    }

    @Test
    fun `PlaybackState valueOf works for all values`() {
        for (name in listOf("IDLE", "BUFFERING", "PLAYING", "PAUSED", "STOPPED", "COMPLETED", "ERROR")) {
            val state = VLCPlayer.PlaybackState.valueOf(name)
            assertEquals(name, state.name)
        }
    }

    // ------------------------------------------------------------------
    // TrackType enum values
    // ------------------------------------------------------------------

    @Test
    fun `TrackType enum has all expected values`() {
        val types = VLCPlayer.TrackType.values()
        val expected = setOf("AUDIO", "VIDEO", "SUBTITLE")
        assertEquals(expected, types.map { it.name }.toSet())
    }

    // ------------------------------------------------------------------
    // Track data class
    // ------------------------------------------------------------------

    @Test
    fun `Track data class holds correct values`() {
        val track = VLCPlayer.Track(id = 1, name = "English", type = VLCPlayer.TrackType.AUDIO)
        assertEquals(1, track.id)
        assertEquals("English", track.name)
        assertEquals(VLCPlayer.TrackType.AUDIO, track.type)
    }

    @Test
    fun `Track equality works for identical values`() {
        val a = VLCPlayer.Track(id = 2, name = "French", type = VLCPlayer.TrackType.SUBTITLE)
        val b = VLCPlayer.Track(id = 2, name = "French", type = VLCPlayer.TrackType.SUBTITLE)
        assertEquals(a, b)
        assertEquals(a.hashCode(), b.hashCode())
    }

    @Test
    fun `Track copy changes only specified fields`() {
        val original = VLCPlayer.Track(id = 1, name = "Track 1", type = VLCPlayer.TrackType.VIDEO)
        val copy = original.copy(name = "Track 1 HD")
        assertEquals("Track 1 HD", copy.name)
        assertEquals(original.id, copy.id)
        assertEquals(original.type, copy.type)
    }

    @Test
    fun `Track with negative id does not crash`() {
        val track = VLCPlayer.Track(id = -1, name = "Disabled", type = VLCPlayer.TrackType.SUBTITLE)
        assertEquals(-1, track.id)
    }
}
