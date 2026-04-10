package com.catalogizer.android.ui.player

import android.os.Bundle
import android.view.WindowManager
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Slider
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.catalogizer.android.ui.theme.CatalogizerTheme

/**
 * MediaPlayerActivity provides a full-screen media playback experience.
 * Handles video and audio playback with native Android MediaPlayer.
 */
class MediaPlayerActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Keep screen on during playback
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        
        val mediaId = intent.getStringExtra(EXTRA_MEDIA_ID)
        val mediaTitle = intent.getStringExtra(EXTRA_MEDIA_TITLE) ?: "Unknown"
        val mediaUrl = intent.getStringExtra(EXTRA_MEDIA_URL)
        
        if (mediaId == null || mediaUrl == null) {
            Toast.makeText(this, "Invalid media", Toast.LENGTH_SHORT).show()
            finish()
            return
        }
        
        setContent {
            CatalogizerTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = Color.Black
                ) {
                    MediaPlayerScreen(
                        title = mediaTitle,
                        mediaUrl = mediaUrl,
                        onBackPressed = { finish() }
                    )
                }
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        // Clear keep screen on flag
        window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
    }

    companion object {
        const val EXTRA_MEDIA_ID = "media_id"
        const val EXTRA_MEDIA_TITLE = "media_title"
        const val EXTRA_MEDIA_URL = "media_url"
    }
}

@Composable
fun MediaPlayerScreen(
    title: String,
    mediaUrl: String,
    onBackPressed: () -> Unit
) {
    var isPlaying by remember { mutableStateOf(true) }
    var progress by remember { mutableStateOf(0f) }
    var currentTime by remember { mutableStateOf("0:00") }
    var duration by remember { mutableStateOf("0:00") }
    
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
    ) {
        // Placeholder for video surface
        // In production, this would use ExoPlayer or MediaPlayer
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.DarkGray),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = "Playing: $title",
                color = Color.White,
                style = MaterialTheme.typography.headlineMedium
            )
        }
        
        // Top controls
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            IconButton(onClick = onBackPressed) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                    tint = Color.White
                )
            }
            
            Text(
                text = title,
                color = Color.White,
                modifier = Modifier.align(Alignment.Center),
                style = MaterialTheme.typography.titleMedium
            )
        }
        
        // Bottom controls
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomCenter)
                .padding(16.dp)
        ) {
            // Progress slider
            Slider(
                value = progress,
                onValueChange = { progress = it },
                modifier = Modifier.fillMaxWidth()
            )
            
            // Time display
            Text(
                text = "$currentTime / $duration",
                color = Color.White,
                modifier = Modifier.padding(bottom = 8.dp)
            )
            
            // Play/Pause button
            IconButton(
                onClick = { isPlaying = !isPlaying },
                modifier = Modifier.align(Alignment.CenterHorizontally)
            ) {
                Icon(
                    imageVector = if (isPlaying) Icons.Default.Pause else Icons.Default.PlayArrow,
                    contentDescription = if (isPlaying) "Pause" else "Play",
                    tint = Color.White,
                    modifier = Modifier.padding(16.dp)
                )
            }
        }
    }
}
