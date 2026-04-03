package com.catalogizer.androidtv.ui.splash

import android.content.Context
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.catalogizer.androidtv.BuildConfig
import com.catalogizer.androidtv.R
import kotlinx.coroutines.delay

/**
 * TV splash screen with version display and a minimum display duration
 * (longer on first launch). Transitions to the main content when both the
 * timer completes and [isAppReady] is true.
 */
@Composable
fun SplashContent(
    isAppReady: Boolean,
    onSplashComplete: () -> Unit
) {
    val context = LocalContext.current
    val prefs = remember { context.getSharedPreferences("catalogizer_prefs", Context.MODE_PRIVATE) }
    val isFirstLaunch = remember { !prefs.getBoolean("has_launched", false) }
    val minDuration = remember { if (isFirstLaunch) 5000L else 2500L }

    LaunchedEffect(Unit) {
        prefs.edit().putBoolean("has_launched", true).apply()
    }

    var timerComplete by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        delay(minDuration)
        timerComplete = true
    }

    LaunchedEffect(timerComplete, isAppReady) {
        if (timerComplete && isAppReady) {
            onSplashComplete()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(Color(0xFF101214), Color(0xFF1A1C1E))
                )
            )
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Image(
                painter = painterResource(R.drawable.app_icon),
                contentDescription = "Catalogizer",
                modifier = Modifier.size(120.dp),
                contentScale = ContentScale.Fit
            )
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "Catalogizer",
                fontSize = 36.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Advanced Multi-Protocol Media\nCollection Management System",
                fontSize = 16.sp,
                color = Color(0xFF94A3B8),
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(32.dp))
            CircularProgressIndicator(
                modifier = Modifier.size(40.dp),
                color = Color(0xFF4A90E2),
                strokeWidth = 3.dp
            )
        }

        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .padding(bottom = 32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = "Made with \u2764\uFE0F by Vasic Digital",
                fontSize = 12.sp,
                color = Color(0xFF64748B)
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = "v${BuildConfig.VERSION_NAME}",
                fontSize = 11.sp,
                color = Color(0xFF475569)
            )
        }
    }
}
