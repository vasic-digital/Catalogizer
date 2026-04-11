@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Text
import com.catalogizer.androidtv.data.playback.PlaybackFormatter
import com.catalogizer.androidtv.data.playback.UiPlaybackProgress

/**
 * Indicator overlay rendered on top of a [MediaCard] showing
 * how much of the media has been reproduced. Shows five pieces
 * of information that the user asked for on every card:
 *
 *   1. Total duration/pages (e.g. "2h 0m" / "320 pages")
 *   2. Current position (e.g. "30m" / "140 pages")
 *   3. Last-session amount (e.g. "20m" / "15 pages")
 *   4. Total reproduction count (e.g. "3×")
 *   5. Visual progress bar (hidden when duration unknown)
 *
 * The composable is designed to sit inside the bottom of the
 * card's Box as an absolutely-positioned overlay. Tap/click
 * wiring is done by the parent — this component stays a pure
 * display layer so it's trivially unit-testable.
 */
@Composable
fun ProgressBadge(
    progress: UiPlaybackProgress?,
    modifier: Modifier = Modifier,
) {
    if (progress == null) return

    val duration = PlaybackFormatter.formatAmount(
        progress.durationTotal ?: 0L,
        progress.positionUnit,
    )
    val currentLabel = PlaybackFormatter.formatProgress(
        progress.lastPosition,
        progress.durationTotal,
        progress.positionUnit,
    )
    val lastSessionLabel = PlaybackFormatter.formatAmount(
        progress.lastSessionAmount,
        progress.positionUnit,
    ) + " last"
    val reps = PlaybackFormatter.formatReproductionCount(progress.totalReproductions)
    val pct = PlaybackFormatter.progressFraction(
        progress.lastPosition,
        progress.durationTotal,
    )

    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 6.dp, vertical = 4.dp)
            .clip(RoundedCornerShape(6.dp))
            .background(Color.Black.copy(alpha = 0.75f))
            .padding(horizontal = 6.dp, vertical = 4.dp),
        verticalArrangement = Arrangement.spacedBy(2.dp),
    ) {
        Text(
            text = duration,
            color = Color.White,
            fontSize = 11.sp,
            fontWeight = FontWeight.SemiBold,
        )
        Text(
            text = currentLabel,
            color = Color.White.copy(alpha = 0.95f),
            fontSize = 10.sp,
        )
        Text(
            text = lastSessionLabel,
            color = Color.White.copy(alpha = 0.75f),
            fontSize = 10.sp,
        )
        Text(
            text = "$reps played",
            color = Color.White.copy(alpha = 0.6f),
            fontSize = 10.sp,
        )
        if (pct > 0f) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(3.dp)
                    .clip(RoundedCornerShape(2.dp))
                    .background(Color.White.copy(alpha = 0.25f)),
                contentAlignment = Alignment.CenterStart,
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(pct)
                        .height(3.dp)
                        .clip(RoundedCornerShape(2.dp))
                        .background(Color(0xFF4FC3F7)),
                )
            }
        }
    }
}
