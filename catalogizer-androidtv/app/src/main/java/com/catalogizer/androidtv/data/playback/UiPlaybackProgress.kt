package com.catalogizer.androidtv.data.playback

/**
 * View-layer shape of the backend `media_progress` row. Kept
 * distinct from the JSON DTO so the UI never has to touch the
 * kotlinx.serialization types and can handle null/missing
 * fields with the usual Kotlin safety.
 */
data class UiPlaybackProgress(
    val mediaItemId: Long,
    val positionUnit: String,          // "seconds" | "pages" | "events"
    val durationTotal: Long?,          // null when not yet learned
    val lastPosition: Long,
    val lastSessionAmount: Long,
    val totalReproductions: Long,
    val aggregateAmount: Long,
    val lastSessionEndedAtMs: Long?,   // epoch millis, null when never played
)

/**
 * One historical session row used by [UiHistoryEntry] lists
 * rendered inside HistoryDialog.
 */
data class UiPlaybackSession(
    val id: Long,
    val positionUnit: String,
    val startPosition: Long,
    val endPosition: Long,
    val totalAmount: Long,
    val startedAtMs: Long,
    val endedAtMs: Long?,
    val completed: Boolean,
)
