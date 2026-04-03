package com.catalogizer.androidtv.ui.player

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.runtime.*
import androidx.tv.material3.MaterialTheme
import com.catalogizer.androidtv.ui.screens.player.MediaPlayerScreen
import com.catalogizer.androidtv.ui.theme.CatalogizerTVTheme

/**
 * Standalone activity for media playback launched via intent with MEDIA_ID, MEDIA_URL,
 * and MEDIA_TITLE extras. Hosts [MediaPlayerScreen] within the TV theme.
 */
class MediaPlayerActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Get media information from intent extras
        val mediaId = intent.getLongExtra("MEDIA_ID", 0L)
        val mediaUrl = intent.getStringExtra("MEDIA_URL") ?: ""
        val mediaTitle = intent.getStringExtra("MEDIA_TITLE") ?: "Media $mediaId"
        
        setContent {
            CatalogizerTVTheme {
                MediaPlayerScreen(
                    mediaId = mediaId,
                    mediaUrl = mediaUrl,
                    mediaTitle = mediaTitle,
                    onNavigateBack = {
                        // Close the activity when back is pressed
                        finish()
                    }
                )
            }
        }
    }
}