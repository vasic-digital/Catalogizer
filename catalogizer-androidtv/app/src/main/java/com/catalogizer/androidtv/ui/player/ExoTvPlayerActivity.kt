package com.catalogizer.androidtv.ui.player

import android.annotation.SuppressLint
import android.os.Bundle
import android.util.Log
import android.view.KeyEvent
import android.view.View
import android.widget.ProgressBar
import androidx.activity.ComponentActivity
import androidx.lifecycle.lifecycleScope
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.ui.PlayerView
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.R
import kotlinx.coroutines.launch
import kotlinx.serialization.json.jsonPrimitive

/**
 * Media3 / ExoPlayer replacement for the crashing libVLC-based
 * player. Uses DefaultHttpDataSource with an
 * "Authorization: Bearer <token>" header so the authenticated
 * /api/v1/stream/:id endpoint is reachable without leaking the
 * JWT into the URL. Pure Kotlin, no native-library init work —
 * eliminates the entire class of SIGSEGVs we hit in
 * VLCJniObject_newFromJavaLibVlc on armeabi-v7a.
 *
 * The activity reads its media id from the "media_id" intent
 * extra, calls container.api.getEntityStream(id) to resolve the
 * stream URL, and hands the result to ExoPlayer. Errors bubble
 * up via Player.Listener.onPlayerError and finish() the activity
 * so HelixQA sees the app return to the previous screen instead
 * of a frozen black frame.
 */
@SuppressLint("UnsafeOptInUsageError")
class ExoTvPlayerActivity : ComponentActivity() {

    private var player: ExoPlayer? = null
    private lateinit var playerView: PlayerView
    private lateinit var buffering: ProgressBar
    private var mediaId: Long = 0L

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_exo_tv_player)

        playerView = findViewById(R.id.player_view)
        buffering = findViewById(R.id.buffering_indicator)

        mediaId = intent.getLongExtra(EXTRA_MEDIA_ID, 0L)
        if (mediaId <= 0L) {
            Log.e(TAG, "Missing media_id extra, finishing")
            finish()
            return
        }

        initialisePlayer()
        resolveAndPlay()
    }

    private fun initialisePlayer() {
        val container = DependencyContainer.getInstance(applicationContext)
        val token = container.authRepository.authState.value.token.orEmpty()
        val httpFactory = DefaultHttpDataSource.Factory()
            .setAllowCrossProtocolRedirects(true)
            .setConnectTimeoutMs(15_000)
            .setReadTimeoutMs(20_000)
            .setDefaultRequestProperties(
                mapOf("Authorization" to "Bearer $token")
            )
        val mediaSourceFactory = DefaultMediaSourceFactory(this)
            .setDataSourceFactory(httpFactory)

        player = ExoPlayer.Builder(this)
            .setMediaSourceFactory(mediaSourceFactory)
            .setSeekBackIncrementMs(10_000)
            .setSeekForwardIncrementMs(10_000)
            .build()
            .also { exo ->
                playerView.player = exo
                exo.addListener(object : Player.Listener {
                    override fun onPlaybackStateChanged(state: Int) {
                        buffering.visibility =
                            if (state == Player.STATE_BUFFERING) View.VISIBLE else View.GONE
                    }

                    override fun onPlayerError(error: PlaybackException) {
                        Log.e(TAG, "Playback error: ${error.errorCodeName}", error)
                        finish()
                    }
                })
            }
    }

    private fun resolveAndPlay() {
        val container = DependencyContainer.getInstance(applicationContext)
        lifecycleScope.launch {
            try {
                val resp = container.api.getEntityStream(mediaId)
                if (!resp.isSuccessful) {
                    Log.e(TAG, "Stream resolve HTTP ${resp.code()}")
                    finish()
                    return@launch
                }
                val body = resp.body()
                val streamPath = body?.get("stream_url")?.jsonPrimitive?.content
                if (streamPath.isNullOrBlank()) {
                    Log.e(TAG, "Empty stream_url in response body")
                    finish()
                    return@launch
                }
                val base = container.getServerUrl().trimEnd('/')
                val url = if (streamPath.startsWith("/")) "$base$streamPath" else streamPath
                Log.d(TAG, "Playing $url")
                val exo = player ?: return@launch
                exo.setMediaItem(MediaItem.fromUri(url))
                exo.prepare()
                exo.playWhenReady = true
            } catch (e: Exception) {
                Log.e(TAG, "Failed to resolve stream", e)
                finish()
            }
        }
    }

    override fun onStop() {
        super.onStop()
        player?.pause()
    }

    override fun onDestroy() {
        super.onDestroy()
        player?.release()
        player = null
    }

    override fun onKeyDown(keyCode: Int, event: KeyEvent?): Boolean {
        val exo = player ?: return super.onKeyDown(keyCode, event)
        return when (keyCode) {
            KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
            KeyEvent.KEYCODE_DPAD_CENTER -> {
                if (exo.isPlaying) exo.pause() else exo.play()
                true
            }
            KeyEvent.KEYCODE_MEDIA_REWIND,
            KeyEvent.KEYCODE_DPAD_LEFT -> {
                exo.seekBack()
                true
            }
            KeyEvent.KEYCODE_MEDIA_FAST_FORWARD,
            KeyEvent.KEYCODE_DPAD_RIGHT -> {
                exo.seekForward()
                true
            }
            KeyEvent.KEYCODE_BACK -> {
                finish()
                true
            }
            else -> super.onKeyDown(keyCode, event)
        }
    }

    companion object {
        private const val TAG = "ExoTvPlayerActivity"
        const val EXTRA_MEDIA_ID = "media_id"
    }
}
