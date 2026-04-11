package com.catalogizer.androidtv.data.playback

import android.util.Log
import com.catalogizer.androidtv.data.remote.CatalogizerApi
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.longOrNull

/**
 * PlaybackRepository adapts the /api/v1/entities/:id/progress
 * and /api/v1/entities/:id/history JSON responses into the
 * [UiPlaybackProgress] / [UiPlaybackSession] shapes the TV
 * composables consume.
 *
 * Kept intentionally thin: no caching, no StateFlow. Callers
 * (HomeViewModel, MediaDetailViewModel, HistoryDialog) are
 * responsible for fetching on focus/mount and holding state
 * themselves. That way a single repo instance in the
 * DependencyContainer can be shared across screens.
 */
class PlaybackRepository(
    private val api: CatalogizerApi,
) {
    /**
     * Fetches the rolled-up progress summary for one entity.
     * Returns null when the user has never reproduced the item
     * (the backend responds with `{"progress": null}`).
     */
    suspend fun getProgress(mediaItemId: Long): UiPlaybackProgress? {
        return try {
            val resp = api.getEntityProgress(mediaItemId)
            if (!resp.isSuccessful) {
                Log.w(TAG, "getProgress HTTP ${resp.code()} for $mediaItemId")
                return null
            }
            val body = resp.body() ?: return null
            val progEl = body["progress"] ?: return null
            if (progEl.toString() == "null") return null
            val prog = progEl.jsonObject
            UiPlaybackProgress(
                mediaItemId = mediaItemId,
                positionUnit = prog["position_unit"]?.jsonPrimitive?.content ?: "seconds",
                durationTotal = prog["duration_total"]?.jsonPrimitive?.longOrNull,
                lastPosition = prog["last_position"]?.jsonPrimitive?.longOrNull ?: 0L,
                lastSessionAmount = prog["last_session_amount"]?.jsonPrimitive?.longOrNull ?: 0L,
                totalReproductions = prog["total_reproductions"]?.jsonPrimitive?.longOrNull ?: 0L,
                aggregateAmount = prog["aggregate_amount"]?.jsonPrimitive?.longOrNull ?: 0L,
                lastSessionEndedAtMs = parseIso8601(prog["last_session_ended_at"]?.jsonPrimitive?.content),
            )
        } catch (t: Throwable) {
            Log.w(TAG, "getProgress exception: ${t.message}")
            null
        }
    }

    /**
     * Fetches up to [limit] session rows for the history dialog.
     * Returns an empty list when the user has no sessions yet.
     */
    suspend fun listHistory(mediaItemId: Long, limit: Int = 50): List<UiPlaybackSession> {
        return try {
            val resp = api.getEntityHistory(mediaItemId, limit)
            if (!resp.isSuccessful) return emptyList()
            val body = resp.body() ?: return emptyList()
            val sessionsEl = body["sessions"] ?: return emptyList()
            sessionsEl.jsonArray.map { it.jsonObject.toSession() }
        } catch (t: Throwable) {
            Log.w(TAG, "listHistory exception: ${t.message}")
            emptyList()
        }
    }

    private fun JsonObject.toSession(): UiPlaybackSession = UiPlaybackSession(
        id = this["id"]?.jsonPrimitive?.longOrNull ?: 0L,
        positionUnit = this["position_unit"]?.jsonPrimitive?.content ?: "seconds",
        startPosition = this["start_position"]?.jsonPrimitive?.longOrNull ?: 0L,
        endPosition = this["end_position"]?.jsonPrimitive?.longOrNull ?: 0L,
        totalAmount = this["total_amount"]?.jsonPrimitive?.longOrNull ?: 0L,
        startedAtMs = parseIso8601(this["started_at"]?.jsonPrimitive?.content) ?: 0L,
        endedAtMs = parseIso8601(this["ended_at"]?.jsonPrimitive?.content),
        completed = this["completed"]?.jsonPrimitive?.booleanOrNull ?: false,
    )

    private fun parseIso8601(raw: String?): Long? {
        if (raw.isNullOrBlank() || raw == "null") return null
        return try {
            java.time.OffsetDateTime.parse(raw).toInstant().toEpochMilli()
        } catch (_: Throwable) {
            null
        }
    }

    companion object {
        private const val TAG = "PlaybackRepository"
    }
}
