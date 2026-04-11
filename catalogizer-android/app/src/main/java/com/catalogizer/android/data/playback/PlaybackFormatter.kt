package com.catalogizer.android.data.playback

/**
 * Mirror of catalogizer-androidtv's PlaybackFormatter. Pure functions with
 * no Android dependencies so badge rendering on the phone produces the
 * exact same strings as the TV app.
 */
object PlaybackFormatter {

    fun formatAmount(value: Long, unit: String): String {
        if (value <= 0L) return "-"
        return when (unit) {
            "seconds" -> formatSeconds(value)
            "pages" -> "$value ${if (value == 1L) "page" else "pages"}"
            "events" -> "$value ${if (value == 1L) "run" else "runs"}"
            else -> "$value $unit"
        }
    }

    fun formatProgress(current: Long, total: Long?, unit: String): String {
        if (total == null || total <= 0) return formatAmount(current, unit)
        return "${formatAmount(current, unit)} / ${formatAmount(total, unit)}"
    }

    fun progressFraction(current: Long, total: Long?): Float {
        if (total == null || total <= 0L) return 0f
        val pct = current.toFloat() / total.toFloat()
        return pct.coerceIn(0f, 1f)
    }

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

    fun formatReproductionCount(count: Long): String = "${count}×"
}
