@file:SuppressLint("RestrictedApi")

package com.catalogizer.androidtv.data.tv

import android.annotation.SuppressLint
import android.content.ContentValues
import android.net.Uri
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.MediaItem

/**
 * Maps [MediaItem] instances to [ContentValues] suitable for inserting into the
 * system TvProvider as [TvContractCompat.PreviewPrograms] or [TvContractCompat.WatchNextPrograms].
 */
object ChannelProgramMapper {

    private val TYPE_MAP = mapOf(
        "movie" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
        "tv_show" to TvContractCompat.PreviewPrograms.TYPE_TV_SERIES,
        "tv_season" to TvContractCompat.PreviewPrograms.TYPE_TV_SEASON,
        "tv_episode" to TvContractCompat.PreviewPrograms.TYPE_TV_EPISODE,
        "music" to TvContractCompat.PreviewPrograms.TYPE_TRACK,
        "music_artist" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "music_album" to TvContractCompat.PreviewPrograms.TYPE_ALBUM,
        "song" to TvContractCompat.PreviewPrograms.TYPE_TRACK,
        "anime" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
        "documentary" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
        "concert" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
        "game" to TvContractCompat.PreviewPrograms.TYPE_GAME,
        "software" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "ebook" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "book" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "comic" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "audiobook" to TvContractCompat.PreviewPrograms.TYPE_ALBUM,
        "podcast" to TvContractCompat.PreviewPrograms.TYPE_CHANNEL,
        "training" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "sports" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
        "news" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
        "other" to TvContractCompat.PreviewPrograms.TYPE_CLIP
    )

    fun mapToPreviewProgramType(mediaType: String?): Int {
        return TYPE_MAP[mediaType] ?: TvContractCompat.PreviewPrograms.TYPE_CLIP
    }

    fun buildDeepLinkUri(mediaId: Long, mediaType: String?, action: String? = null): Uri {
        val builder = Uri.Builder()
            .scheme("catalogizer")
            .authority("media")
            .appendPath(mediaId.toString())
        mediaType?.let { builder.appendQueryParameter("type", it) }
        action?.let { builder.appendQueryParameter("action", it) }
        return builder.build()
    }

    /**
     * Resolves a potentially relative image URL to an absolute URL.
     * The Android TV system channel infrastructure fetches images externally
     * and cannot resolve relative paths like `/api/v1/image-proxy?url=...`.
     */
    fun resolveImageUrl(url: String?, serverBaseUrl: String?): String? {
        if (url == null) return null
        // Relative URLs (including /api/v1/image-proxy) need the
        // server base to become absolute for the TV channel system.
        if (url.startsWith("/") && serverBaseUrl != null) {
            return serverBaseUrl.trimEnd('/') + url
        }
        // Direct TMDB URLs get routed through the API image proxy
        // to bypass Mi Box DNS blocking of TMDB CDN.
        if (url.contains("image.tmdb.org") && serverBaseUrl != null) {
            return try {
                val encoded = java.net.URLEncoder.encode(url, "UTF-8")
                serverBaseUrl.trimEnd('/') + "/api/v1/image-proxy?url=$encoded"
            } catch (_: Exception) { url }
        }
        if (url.startsWith("http://") || url.startsWith("https://")) return url
        return url
    }

    fun toPreviewProgramValues(item: MediaItem, channelId: Long, serverBaseUrl: String? = null): ContentValues {
        val values = ContentValues()
        values.put(TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID, channelId)
        values.put(TvContractCompat.PreviewPrograms.COLUMN_TITLE, item.title)
        values.put(TvContractCompat.PreviewPrograms.COLUMN_TYPE, mapToPreviewProgramType(item.mediaType))
        values.put(TvContractCompat.PreviewPrograms.COLUMN_INTENT_URI, buildDeepLinkUri(item.id, item.mediaType).toString())

        item.description?.let { values.put(TvContractCompat.PreviewPrograms.COLUMN_SHORT_DESCRIPTION, it) }
        resolveImageUrl(item.posterUrl, serverBaseUrl)?.let { values.put(TvContractCompat.PreviewPrograms.COLUMN_POSTER_ART_URI, it) }
        resolveImageUrl(item.backdropUrl, serverBaseUrl)?.let { values.put(TvContractCompat.PreviewPrograms.COLUMN_THUMBNAIL_URI, it) }
        item.duration?.let { durationSec -> values.put(TvContractCompat.PreviewPrograms.COLUMN_DURATION_MILLIS, durationSec * 1000) }
        item.year?.let { values.put(TvContractCompat.PreviewPrograms.COLUMN_RELEASE_DATE, it.toString()) }
        if (item.genres.isNotEmpty()) {
            values.put(TvContractCompat.PreviewPrograms.COLUMN_GENRE, item.genres.joinToString(", "))
        }
        if (item.watchProgress > 0.0 && item.duration != null) {
            val positionMs = (item.watchProgress * item.duration!! * 1000).toLong()
            values.put(TvContractCompat.PreviewPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS, positionMs)
            values.put(TvContractCompat.PreviewPrograms.COLUMN_DURATION_MILLIS, item.duration!! * 1000)
        }
        return values
    }

    fun toWatchNextValues(item: MediaItem, watchNextType: Int, serverBaseUrl: String? = null): ContentValues {
        val values = ContentValues()
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_TITLE, item.title)
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_TYPE, mapToPreviewProgramType(item.mediaType))
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_WATCH_NEXT_TYPE, watchNextType)
        // CONTINUE items resume playback directly; NEXT items open detail for the new episode
        val deepLinkAction = if (watchNextType == TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE) "play" else null
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_INTENT_URI, buildDeepLinkUri(item.id, item.mediaType, deepLinkAction).toString())
        values.put(TvContractCompat.WatchNextPrograms.COLUMN_LAST_ENGAGEMENT_TIME_UTC_MILLIS, System.currentTimeMillis())

        item.description?.let { values.put(TvContractCompat.WatchNextPrograms.COLUMN_SHORT_DESCRIPTION, it) }
        resolveImageUrl(item.posterUrl, serverBaseUrl)?.let { values.put(TvContractCompat.WatchNextPrograms.COLUMN_POSTER_ART_URI, it) }
        resolveImageUrl(item.backdropUrl, serverBaseUrl)?.let { values.put(TvContractCompat.WatchNextPrograms.COLUMN_THUMBNAIL_URI, it) }
        item.duration?.let { durationSec ->
            values.put(TvContractCompat.WatchNextPrograms.COLUMN_DURATION_MILLIS, durationSec * 1000)
            if (item.watchProgress > 0.0) {
                val positionMs = (item.watchProgress * durationSec * 1000).toLong()
                values.put(TvContractCompat.WatchNextPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS, positionMs)
            }
        }
        return values
    }
}
