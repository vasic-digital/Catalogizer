@file:OptIn(ExperimentalTvMaterial3Api::class)
package com.catalogizer.androidtv.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.Text
import com.catalogizer.androidtv.data.playback.PlaybackFormatter
import com.catalogizer.androidtv.data.playback.PlaybackRepository
import com.catalogizer.androidtv.data.playback.UiPlaybackProgress
import com.catalogizer.androidtv.data.playback.UiPlaybackSession
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone

/**
 * Full detailed history dialog opened when the user clicks a
 * [ProgressBadge]. Shows the aggregate summary at the top
 * (total reproductions, aggregate duration, learned total
 * duration) and a scrollable list of every session row
 * returned by /api/v1/entities/:id/history.
 *
 * Load happens via [PlaybackRepository] inside a LaunchedEffect
 * so the dialog pulls fresh data every time it opens — a
 * reproduction finished 10 seconds ago always shows up.
 */
@Composable
fun HistoryDialog(
    mediaItemId: Long,
    mediaTitle: String,
    repository: PlaybackRepository,
    onDismiss: () -> Unit,
) {
    var progress by remember { mutableStateOf<UiPlaybackProgress?>(null) }
    var sessions by remember { mutableStateOf<List<UiPlaybackSession>>(emptyList()) }
    var loaded by remember { mutableStateOf(false) }

    LaunchedEffect(mediaItemId) {
        progress = repository.getProgress(mediaItemId)
        sessions = repository.listHistory(mediaItemId, limit = 100)
        loaded = true
    }

    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false),
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.92f))
                .padding(48.dp),
            contentAlignment = Alignment.Center,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth(0.75f)
                    .fillMaxHeight(0.85f)
                    .clip(RoundedCornerShape(16.dp))
                    .background(Color(0xFF141414))
                    .padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                Text(
                    text = "Reproduction History",
                    color = Color.White,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.Bold,
                )
                Text(
                    text = mediaTitle,
                    color = Color.White.copy(alpha = 0.75f),
                    fontSize = 14.sp,
                )
                HistorySummary(progress)
                HistorySessionList(sessions, loaded)
            }
        }
    }
}

@Composable
private fun HistorySummary(progress: UiPlaybackProgress?) {
    val summaryRows = if (progress == null) {
        listOf("Status" to "Never reproduced")
    } else {
        buildList {
            add(
                "Total duration" to PlaybackFormatter.formatAmount(
                    progress.durationTotal ?: 0L,
                    progress.positionUnit,
                ),
            )
            add(
                "Current position" to PlaybackFormatter.formatProgress(
                    progress.lastPosition,
                    progress.durationTotal,
                    progress.positionUnit,
                ),
            )
            add(
                "Last session" to PlaybackFormatter.formatAmount(
                    progress.lastSessionAmount,
                    progress.positionUnit,
                ),
            )
            add(
                "Times played" to PlaybackFormatter.formatReproductionCount(
                    progress.totalReproductions,
                ),
            )
            add(
                "Total reproduced" to PlaybackFormatter.formatAmount(
                    progress.aggregateAmount,
                    progress.positionUnit,
                ),
            )
        }
    }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(Color.White.copy(alpha = 0.05f))
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        summaryRows.forEach { (label, value) ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(
                    text = label,
                    color = Color.White.copy(alpha = 0.6f),
                    fontSize = 13.sp,
                )
                Text(
                    text = value,
                    color = Color.White,
                    fontSize = 13.sp,
                    fontWeight = FontWeight.SemiBold,
                )
            }
        }
    }
}

@Composable
private fun HistorySessionList(
    sessions: List<UiPlaybackSession>,
    loaded: Boolean,
) {
    when {
        !loaded -> Text(
            text = "Loading…",
            color = Color.White.copy(alpha = 0.6f),
            fontSize = 14.sp,
        )
        sessions.isEmpty() -> Text(
            text = "No sessions yet. Press Play to record your first.",
            color = Color.White.copy(alpha = 0.6f),
            fontSize = 14.sp,
        )
        else -> {
            Text(
                text = "Sessions",
                color = Color.White,
                fontSize = 15.sp,
                fontWeight = FontWeight.Bold,
            )
            LazyColumn(
                modifier = Modifier.fillMaxWidth(),
                verticalArrangement = Arrangement.spacedBy(6.dp),
                contentPadding = PaddingValues(bottom = 4.dp),
            ) {
                items(sessions, key = { it.id }) { session ->
                    SessionRow(session)
                }
            }
        }
    }
}

@Composable
private fun SessionRow(session: UiPlaybackSession) {
    val dateLabel = formatInstant(session.startedAtMs)
    val amount = PlaybackFormatter.formatAmount(session.totalAmount, session.positionUnit)
    val status = if (session.completed) "Completed" else "In progress"
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .background(Color.White.copy(alpha = 0.05f))
            .padding(horizontal = 12.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = dateLabel,
            color = Color.White,
            fontSize = 13.sp,
            modifier = Modifier.width(200.dp),
        )
        Text(
            text = amount,
            color = Color.White,
            fontSize = 13.sp,
            modifier = Modifier.width(140.dp),
        )
        Text(
            text = status,
            color = if (session.completed) Color(0xFF81C784) else Color(0xFFFFB74D),
            fontSize = 12.sp,
        )
    }
}

private fun formatInstant(ms: Long): String {
    if (ms <= 0L) return "-"
    val fmt = SimpleDateFormat("yyyy-MM-dd HH:mm", Locale.getDefault())
    fmt.timeZone = TimeZone.getDefault()
    return fmt.format(Date(ms))
}
