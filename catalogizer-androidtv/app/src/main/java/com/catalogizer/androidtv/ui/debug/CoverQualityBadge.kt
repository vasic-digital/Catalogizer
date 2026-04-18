package com.catalogizer.androidtv.ui.debug

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.catalogizer.androidtv.BuildConfig
import com.catalogizer.androidtv.DependencyContainer
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.Request

/**
 * Android TV port of catalog-web's CoverQualityBadge. Reads the
 * X-Cover-Quality / X-Cover-Source headers that catalog-api emits on
 * /api/v1/cover/:id and renders a tiny pill. Debug-only so release
 * builds pay no network cost.
 */
data class CoverQualitySignal(val quality: String, val source: String)

private val coverQualityCache = mutableMapOf<String, CoverQualitySignal>()

@Composable
fun CoverQualityBadge(
    coverId: Any?,
    enabled: Boolean,
    modifier: Modifier = Modifier,
) {
    if (!BuildConfig.DEBUG || !enabled || coverId == null) return

    val context = LocalContext.current
    val key = coverId.toString()
    var signal by remember(key) { mutableStateOf(coverQualityCache[key]) }

    LaunchedEffect(key, enabled) {
        if (coverQualityCache.containsKey(key)) return@LaunchedEffect
        val container = DependencyContainer.getInstance(context)
        val base = container.getServerUrl().trimEnd('/')
        if (base.isEmpty()) return@LaunchedEffect
        val url = "$base/api/v1/cover/$key"
        val result = withContext(Dispatchers.IO) {
            runCatching {
                val req = Request.Builder().url(url).head().build()
                container.httpClient().newCall(req).execute().use { resp ->
                    CoverQualitySignal(
                        quality = resp.header("X-Cover-Quality") ?: "unknown",
                        source = resp.header("X-Cover-Source") ?: "",
                    )
                }
            }.getOrElse { CoverQualitySignal("unknown", "") }
        }
        coverQualityCache[key] = result
        signal = result
    }

    val current = signal ?: return
    val bg = verdictColor(current.quality)

    Box(
        modifier = modifier
            .background(bg, RoundedCornerShape(999.dp))
            .padding(horizontal = 6.dp, vertical = 2.dp)
    ) {
        val label = if (current.source.isNotEmpty()) {
            "${current.quality} · ${current.source}"
        } else {
            current.quality
        }
        Text(
            text = label,
            fontSize = 10.sp,
            color = Color.White,
        )
    }
}

/** Clears the module-level probe cache. Tests only. */
fun resetCoverQualityCache() {
    coverQualityCache.clear()
}

private fun verdictColor(verdict: String): Color = when (verdict) {
    "pass" -> Color(0xFF10B981)
    "fail_lowres", "fail_wrong_aspect" -> Color(0xFFF59E0B)
    "fail_blurry", "fail_small_bytes" -> Color(0xFFD97706)
    "fail_corrupt" -> Color(0xFFE11D48)
    "fail_too_large" -> Color(0xFFF43F5E)
    "placeholder_fallback" -> Color(0xFF64748B)
    else -> Color(0xFF94A3B8)
}
