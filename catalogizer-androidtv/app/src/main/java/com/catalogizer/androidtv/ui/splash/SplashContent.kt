package com.catalogizer.androidtv.ui.splash

import android.content.Context
import android.util.Log
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
 * TV splash screen with version display and a minimum display duration.
 * Transitions to the main content when both the timer completes and [isAppReady] is true.
 * 
 * CRITICAL: Uses short timeouts to prevent ANR (Android kills app after 5s of no response).
 * - Min display: 1.5s (normal) / 2.5s (first launch)
 * - Safety timeout: 5s max (must be < Android ANR threshold)
 */
@Composable
fun SplashContent(
    isAppReady: Boolean,
    onSplashComplete: () -> Unit
) {
    val context = LocalContext.current
    val prefs = remember { context.getSharedPreferences("catalogizer_prefs", Context.MODE_PRIVATE) }
    val isFirstLaunch = remember { !prefs.getBoolean("has_launched", false) }
    // Reduced durations for faster startup (critical for QA testing)
    val minDuration = remember { if (isFirstLaunch) 2500L else 1500L }
    val startTime = remember { System.currentTimeMillis() }

    LaunchedEffect(Unit) {
        prefs.edit().putBoolean("has_launched", true).apply()
    }

    var timerComplete by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        delay(minDuration)
        timerComplete = true
        Log.d("SplashContent", "Min duration timer complete (${minDuration}ms)")
    }

    // Safety timeout: force complete after 5 seconds regardless of isAppReady
    // Android ANR threshold is ~5 seconds for input, so we must complete before that
    var safetyTimeout by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        delay(5000L)
        safetyTimeout = true
        val elapsed = System.currentTimeMillis() - startTime
        Log.w("SplashContent", "SAFETY TIMEOUT triggered after ${elapsed}ms - forcing splash completion")
    }

    LaunchedEffect(timerComplete, isAppReady, safetyTimeout) {
        if ((timerComplete && isAppReady) || safetyTimeout) {
            val elapsed = System.currentTimeMillis() - startTime
            Log.i("SplashContent", "Splash completing after ${elapsed}ms (timer=$timerComplete, ready=$isAppReady, safety=$safetyTimeout)")
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
                modifier = Modifier.size(80.dp),
                contentScale = ContentScale.Fit
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = "Catalogizer",
                fontSize = 10.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White
            )
            Spacer(modifier = Modifier.height(2.dp))
            Text(
                text = "Advanced Multi-Protocol Media\nCollection Management System",
                fontSize = 7.sp,
                color = Color(0xFF94A3B8),
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(20.dp))
            CircularProgressIndicator(
                modifier = Modifier.size(28.dp),
                color = Color(0xFF4A90E2),
                strokeWidth = 2.dp
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
                fontSize = 7.sp,
                color = Color(0xFF64748B)
            )
            Text(
                text = "v${BuildConfig.VERSION_NAME}",
                fontSize = 7.sp,
                color = Color(0xFF475569)
            )
        }
    }
}
