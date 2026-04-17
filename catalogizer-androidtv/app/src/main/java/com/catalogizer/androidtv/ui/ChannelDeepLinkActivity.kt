package com.catalogizer.androidtv.ui

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.lifecycle.lifecycleScope
import com.catalogizer.androidtv.DependencyContainer
import com.catalogizer.androidtv.data.tv.LaunchAction
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

/**
 * Parsed deep link data from a `catalogizer://media/{id}` URI or channel header URI.
 */
data class DeepLinkResult(
    val mediaId: Long?,
    val mediaType: String?,
    val action: String? = null,
    val browseCategory: String? = null,
    val isHome: Boolean = false
)

/**
 * Parses `catalogizer://media/{id}?type={type}`, `catalogizer://home`, and
 * `catalogizer://browse/{type}` URIs from TV home screen channel clicks.
 */
object DeepLinkParser {
    private val AUDIO_TYPES = setOf("music", "audiobook", "podcast")

    fun parse(uri: Uri?): DeepLinkResult {
        if (uri == null) return DeepLinkResult(null, null)
        val host = uri.host
        val pathSegments = uri.pathSegments ?: emptyList()
        return when (host) {
            "home" -> DeepLinkResult(null, null, isHome = true)
            "browse" -> {
                val category = pathSegments.firstOrNull()
                DeepLinkResult(null, null, browseCategory = category)
            }
            else -> {
                val mediaId = pathSegments.firstOrNull()?.toLongOrNull()
                val mediaType = uri.getQueryParameter("type")
                val action = uri.getQueryParameter("action")
                DeepLinkResult(mediaId, mediaType, action)
            }
        }
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
        const val EXTRA_BROWSE_CATEGORY = "deep_link_browse_category"
        const val EXTRA_IS_HOME = "deep_link_is_home"
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val deepLink = DeepLinkParser.parse(intent?.data)
        Log.d(TAG, "Deep link received: mediaId=${deepLink.mediaId}, type=${deepLink.mediaType}, action=${deepLink.action}, browse=${deepLink.browseCategory}, isHome=${deepLink.isHome}")

        when {
            deepLink.isHome -> {
                launchMainActivity()
                finish()
                return
            }
            deepLink.browseCategory != null -> {
                lifecycleScope.launch {
                    launchBrowse(deepLink.browseCategory)
                    finish()
                }
                return
            }
            deepLink.mediaId == null -> {
                Log.w(TAG, "Invalid deep link — no media ID")
                launchMainActivity()
                finish()
                return
            }
        }

        lifecycleScope.launch {
            resolveAndRoute(deepLink)
            finish()
        }
    }

    private suspend fun launchBrowse(category: String) {
        val container = DependencyContainer.getInstance(this)
        val authState = container.authRepository.authState.first()

        val mainIntent = Intent(this, MainActivity::class.java).apply {
            if (!authState.isAuthenticated) {
                // Not authenticated — open app and let it handle login; after login
                // the pending browse category can be consumed.
                putExtra(EXTRA_BROWSE_CATEGORY, category)
            } else {
                putExtra(EXTRA_BROWSE_CATEGORY, category)
            }
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(mainIntent)
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

        // URI action parameter takes precedence (e.g., Watch Next "play" for resume)
        val uriAction = deepLink.action

        val action: String = if (uriAction == "play" || uriAction == "detail") {
            uriAction
        } else {
            // Fall back to per-category setting
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

            if (shouldPlayImmediately) "play" else "detail"
        }

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
