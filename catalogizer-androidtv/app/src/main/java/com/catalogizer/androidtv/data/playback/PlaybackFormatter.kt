package com.catalogizer.androidtv.data.playback

/**
 * Formats position / amount / duration values into user-facing
 * strings that respect the position unit ("seconds" for video
 * and audio, "pages" for books and comics, "events" for games
 * and software). Pure functions — no Android dependencies so
 * they can be shared between the TV and phone Compose apps
 * without paying for a Context.
 */
object PlaybackFormatter {

    /**
     * Formats a duration or position in the given [unit]. For
     * seconds returns "1h 23m", "23m", or "45s"; for pages
     * returns "120 pages"; for events returns "3 launches".
     * Zero always renders as "-" so the badge collapses to a
     * placeholder instead of "0s" / "0 pages".
     */
    fun formatAmount(value: Long, unit: String): String {
        if (value <= 0L) return "-"
        return when (unit) {
            "seconds" -> formatSeconds(value)
            "pages" -> "$value ${if (value == 1L) "page" else "pages"}"
            "events" -> "$value ${if (value == 1L) "run" else "runs"}"
            else -> "$value $unit"
        }
    }

    /**
     * Renders "N / M unit" for a progress-in-total display. If
     * total is null or zero we fall back to just the current
     * value so the badge still says something meaningful.
     */
    fun formatProgress(current: Long, total: Long?, unit: String): String {
        if (total == null || total <= 0) return formatAmount(current, unit)
        return "${formatAmount(current, unit)} / ${formatAmount(total, unit)}"
    }

    /**
     * Clamps current/total to 0..1 for a progress bar. Returns
     * 0 when we don't know the total — the caller decides
     * whether to hide the bar or show a fully-empty track.
     */
    fun progressFraction(current: Long, total: Long?): Float {
        if (total == null || total <= 0L) return 0f
        val pct = current.toFloat() / total.toFloat()
        return pct.coerceIn(0f, 1f)
    }

    /**
     * Seconds → "H h M m" / "M m" / "S s". Uses the common
     * abbreviated form so the badge stays compact on 10-foot UI.
     */
    fun formatSeconds(totalSeconds: Long): String {
        if (totalSeconds <= 0L) return "-"
        val h = totalSeconds / 3600L
        val m = (totalSeconds % 3600L) / 60L
        val s = totalSeconds % 60L
        return buildString {
            if (h > 0) append("${h}h ")
            if (h > 0 || m > 0) append("${m}m")
            if (h == 0L && m == 0L) append("${s}s")
        }.trim()
    }

    /** Renders "N×" with the right multiplier suffix. */
    fun formatReproductionCount(count: Long): String = "${count}×"
}
