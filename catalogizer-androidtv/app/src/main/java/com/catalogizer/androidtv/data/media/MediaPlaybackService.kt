package com.catalogizer.androidtv.data.media

import android.content.Intent
import androidx.media3.common.MediaItem
import androidx.media3.common.util.UnstableApi
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.datasource.DefaultDataSource
import androidx.media3.session.MediaSession
import androidx.media3.session.MediaSessionService
import com.google.common.util.concurrent.Futures
import com.google.common.util.concurrent.ListenableFuture

@UnstableApi
class MediaPlaybackService : MediaSessionService() {
    private var mediaSession: MediaSession? = null
    private var player: ExoPlayer? = null

    // Create your Player and MediaSession in the onCreate lifecycle event
    override fun onCreate() {
        super.onCreate()
        initializePlayerAndSession()
    }

    private fun initializePlayerAndSession() {
        val dataSourceFactory = DefaultDataSource.Factory(this)
        
        player = ExoPlayer.Builder(this)
            .setMediaSourceFactory(ProgressiveMediaSource.Factory(dataSourceFactory))
            .build()
            .also { exoPlayer ->
                mediaSession = MediaSession.Builder(this, exoPlayer)
                    .setCallback(object : MediaSession.Callback {
                        override fun onAddMediaItems(
                            mediaSession: MediaSession,
                            controller: MediaSession.ControllerInfo,
                            mediaItems: MutableList<MediaItem>
                        ): ListenableFuture<MutableList<MediaItem>> {
                            // This is called when a controller wants to add media items
                            return Futures.immediateFuture(mediaItems)
                        }
                    })
                    .build()
            }
    }

    // The user dismissed the app from recent tasks
    override fun onTaskRemoved(rootIntent: Intent?) {
        val player = player ?: return
        if (player.playbackState != androidx.media3.common.Player.STATE_IDLE) {
            player.stop()
        }
        player.release()
        super.onTaskRemoved(rootIntent)
    }

    // Remember to release player and media session in onDestroy
    override fun onDestroy() {
        mediaSession?.run {
            player?.release()
            release()
            mediaSession = null
        }
        player = null
        super.onDestroy()
    }

    // This is the only required callback of a MediaSessionService
    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? {
        return mediaSession
    }
}