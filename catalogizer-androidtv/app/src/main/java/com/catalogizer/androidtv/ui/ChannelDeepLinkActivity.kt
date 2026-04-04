package com.catalogizer.androidtv.ui

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.tv.LaunchAction
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Parsed deep link data from a `catalogizer://media/{id}` URI.
 */
data class DeepLinkResult(
    val mediaId: Long?,
    val mediaType: String?
)

/**
 * Parses `catalogizer://media/{id}?type={type}` URIs from TV home screen channel clicks.
 */
object DeepLinkParser {
    private val AUDIO_TYPES = setOf("music", "audiobook", "podcast")

    fun parse(uri: Uri?): DeepLinkResult {
        if (uri == null) return DeepLinkResult(null, null)
        val mediaId = uri.pathSegments?.firstOrNull()?.toLongOrNull()
        val mediaType = uri.getQueryParameter("type")
        return DeepLinkResult(mediaId, mediaType)
    }

    fun isAudioWithoutContext(mediaType: String?, hasExternalMetadata: Boolean): Boolean {
        return mediaType in AUDIO_TYPES && !hasExternalMetadata
    }
}

/**
 * Transparent activity that handles deep link intents from Android TV home screen channels.
 * Routes to MediaDetailScreen or Player based on per-category launch settings.
 * If the user is not authenticated, redirects to LoginScreen with the pending deep link.
 */
class ChannelDeepLinkActivity : ComponentActivity() {
    companion object {
        private const val TAG = "ChannelDeepLink"
        const val EXTRA_MEDIA_ID = "deep_link_media_id"
        const val EXTRA_MEDIA_TYPE = "deep_link_media_type"
        const val EXTRA_ACTION = "deep_link_action"
    }

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val deepLink = DeepLinkParser.parse(intent?.data)
        Log.d(TAG, "Deep link received: mediaId=${deepLink.mediaId}, type=${deepLink.mediaType}")

        if (deepLink.mediaId == null) {
            Log.w(TAG, "Invalid deep link — no media ID")
            launchMainActivity()
            finish()
            return
        }

        scope.launch {
            resolveAndRoute(deepLink)
            finish()
        }
    }

    private suspend fun resolveAndRoute(deepLink: DeepLinkResult) {
        val container = DependencyContainer.getInstance(this)
        val authState = container.authRepository.authState.first()

        if (!authState.isAuthenticated) {
            val mainIntent = Intent(this, MainActivity::class.java).apply {
                putExtra(EXTRA_MEDIA_ID, deepLink.mediaId)
                putExtra(EXTRA_MEDIA_TYPE, deepLink.mediaType)
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            startActivity(mainIntent)
            return
        }

        val launchAction = if (deepLink.mediaType != null) {
            container.settingsRepository.getLaunchAction(deepLink.mediaType)
        } else {
            LaunchAction.DETAIL
        }

        // Check audio-without-context override
        val shouldPlayImmediately = if (DeepLinkParser.isAudioWithoutContext(
                deepLink.mediaType, hasExternalMetadata = false
            )) {
            try {
                val item = container.mediaRepository.getMediaById(deepLink.mediaId!!).first()
                val hasMetadata = item?.externalMetadata?.isNotEmpty() == true
                if (!hasMetadata) true else launchAction == LaunchAction.IMMEDIATE_PLAY
            } catch (e: Exception) {
                launchAction == LaunchAction.IMMEDIATE_PLAY
            }
        } else {
            launchAction == LaunchAction.IMMEDIATE_PLAY
        }

        val action = if (shouldPlayImmediately) "play" else "detail"

        val mainIntent = Intent(this, MainActivity::class.java).apply {
            putExtra(EXTRA_MEDIA_ID, deepLink.mediaId)
            putExtra(EXTRA_MEDIA_TYPE, deepLink.mediaType)
            putExtra(EXTRA_ACTION, action)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(mainIntent)
    }

    private fun launchMainActivity() {
        val mainIntent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(mainIntent)
    }
}
