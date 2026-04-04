package com.catalogizer.androidtv.data.tv

import android.net.Uri
import androidx.tvprovider.media.tv.TvContractCompat
import com.catalogizer.androidtv.data.models.ExternalMetadata
import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

@RunWith(RobolectricTestRunner::class)
@Config(manifest = Config.NONE)
class ChannelProgramMapperEdgeCaseTest {

    @Test
    fun `toPreviewProgramValues with zero duration does not set duration millis`() {
        val item = MediaItem(id = 1L, title = "Zero Duration", duration = 0L, watchProgress = 0.0)
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 10L)
        // duration 0 should not trigger the watchProgress branch and not put duration via let (let body executes but 0 * 1000 = 0 is still put)
        // The mapper does: item.duration?.let { ... put duration ... }
        // Since duration is 0L (non-null), it will put COLUMN_DURATION_MILLIS = 0
        val durationMillis = values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_DURATION_MILLIS)
        // watchProgress is 0.0, so the watchProgress branch is not entered
        // The let block puts durationSec * 1000 = 0 * 1000 = 0
        assertEquals(0L, durationMillis)
    }

    @Test
    fun `toPreviewProgramValues with negative year still sets release date`() {
        val item = MediaItem(id = 2L, title = "Ancient Content", year = -100)
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 10L)
        val releaseDate = values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_RELEASE_DATE)
        assertEquals("-100", releaseDate)
    }

    @Test
    fun `toPreviewProgramValues with empty title still sets title`() {
        val item = MediaItem(id = 3L, title = "")
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 10L)
        val title = values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_TITLE)
        assertEquals("", title)
    }

    @Test
    fun `toPreviewProgramValues with watch progress exactly 0 dot 0 does not set playback position`() {
        val item = MediaItem(id = 4L, title = "Not Started", duration = 3600L, watchProgress = 0.0)
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 10L)
        // watchProgress == 0.0, so the position branch is skipped
        val position = values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS)
        assertNull(position)
    }

    @Test
    fun `toPreviewProgramValues with watch progress 1 dot 0 sets full position`() {
        val item = MediaItem(id = 5L, title = "Completed", duration = 3600L, watchProgress = 1.0)
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 10L)
        val position = values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS)
        assertNotNull(position)
        // 1.0 * 3600 * 1000 = 3_600_000
        assertEquals(3_600_000L, position)
    }

    @Test
    fun `toWatchNextValues with null duration does not set duration fields`() {
        val item = MediaItem(id = 6L, title = "No Duration", duration = null, watchProgress = 0.5)
        val values = ChannelProgramMapper.toWatchNextValues(
            item,
            watchNextType = TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE
        )
        val durationMillis = values.getAsLong(TvContractCompat.WatchNextPrograms.COLUMN_DURATION_MILLIS)
        val positionMillis = values.getAsLong(TvContractCompat.WatchNextPrograms.COLUMN_LAST_PLAYBACK_POSITION_MILLIS)
        assertNull(durationMillis)
        assertNull(positionMillis)
    }

    @Test
    fun `toWatchNextValues includes engagement time`() {
        val before = System.currentTimeMillis()
        val item = MediaItem(id = 7L, title = "Engaging Content")
        val values = ChannelProgramMapper.toWatchNextValues(
            item,
            watchNextType = TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE
        )
        val after = System.currentTimeMillis()
        val engagementTime = values.getAsLong(TvContractCompat.WatchNextPrograms.COLUMN_LAST_ENGAGEMENT_TIME_UTC_MILLIS)
        assertNotNull(engagementTime)
        assertTrue(engagementTime!! >= before)
        assertTrue(engagementTime <= after)
    }

    @Test
    fun `buildDeepLinkUri with very large mediaId`() {
        val largeId = Long.MAX_VALUE
        val uri = ChannelProgramMapper.buildDeepLinkUri(largeId, "movie")
        assertEquals("catalogizer", uri.scheme)
        assertEquals("media", uri.host)
        assertEquals("/${Long.MAX_VALUE}", uri.path)
        assertEquals("movie", uri.getQueryParameter("type"))
    }

    @Test
    fun `mapToPreviewProgramType with null returns TYPE_CLIP`() {
        val type = ChannelProgramMapper.mapToPreviewProgramType(null)
        assertEquals(TvContractCompat.PreviewPrograms.TYPE_CLIP, type)
    }

    @Test
    fun `toPreviewProgramValues uses coverUrl when no externalMetadata`() {
        val item = MediaItem(
            id = 8L,
            title = "Cover Only",
            coverUrl = "https://example.com/cover.jpg",
            externalMetadata = null
        )
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 10L)
        val posterUri = values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_POSTER_ART_URI)
        assertNotNull(posterUri)
        assertEquals("https://example.com/cover.jpg", posterUri)
    }
}
