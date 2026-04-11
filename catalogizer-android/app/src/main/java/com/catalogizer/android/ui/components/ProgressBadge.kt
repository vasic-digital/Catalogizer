package com.catalogizer.android.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.catalogizer.android.data.playback.PlaybackFormatter
import com.catalogizer.android.data.playback.UiPlaybackProgress

/**
 * Material 3 variant of the reproduction badge used on phone media cards.
 * Shows total duration / pages, current position, last-session amount, and
 * total play count. When [progress] is null the composable renders nothing
 * so callers can safely overlay it on every card.
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
                        .background(MaterialTheme.colorScheme.primary),
                )
            }
        }
    }
}
