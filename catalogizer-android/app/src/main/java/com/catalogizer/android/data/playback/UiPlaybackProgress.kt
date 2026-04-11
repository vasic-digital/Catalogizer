package com.catalogizer.android.data.playback

data class UiPlaybackProgress(
    val mediaItemId: Long,
    val positionUnit: String,
    val durationTotal: Long?,
    val lastPosition: Long,
    val lastSessionAmount: Long,
    val totalReproductions: Long,
    val aggregateAmount: Long,
    val lastSessionEndedAtMs: Long?,
)

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
