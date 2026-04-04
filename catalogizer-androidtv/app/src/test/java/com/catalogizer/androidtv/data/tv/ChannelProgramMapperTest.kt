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
class ChannelProgramMapperTest {

    @Test
    fun `mapToPreviewProgramType maps movie correctly`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            ChannelProgramMapper.mapToPreviewProgramType("movie")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps tv_show correctly`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_TV_SERIES,
            ChannelProgramMapper.mapToPreviewProgramType("tv_show")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps game correctly`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_GAME,
            ChannelProgramMapper.mapToPreviewProgramType("game")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps music to TYPE_TRACK`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_TRACK,
            ChannelProgramMapper.mapToPreviewProgramType("music")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps unknown type to TYPE_CLIP`() {
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_CLIP,
            ChannelProgramMapper.mapToPreviewProgramType("unknown_type")
        )
    }

    @Test
    fun `mapToPreviewProgramType maps all 16 types`() {
        val expectedMappings = mapOf(
            "movie" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            "tv_show" to TvContractCompat.PreviewPrograms.TYPE_TV_SERIES,
            "tv_episode" to TvContractCompat.PreviewPrograms.TYPE_TV_EPISODE,
            "music" to TvContractCompat.PreviewPrograms.TYPE_TRACK,
            "anime" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            "documentary" to TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            "concert" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
            "game" to TvContractCompat.PreviewPrograms.TYPE_GAME,
            "software" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "ebook" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "audiobook" to TvContractCompat.PreviewPrograms.TYPE_ALBUM,
            "podcast" to TvContractCompat.PreviewPrograms.TYPE_CHANNEL,
            "training" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "sports" to TvContractCompat.PreviewPrograms.TYPE_EVENT,
            "news" to TvContractCompat.PreviewPrograms.TYPE_CLIP,
            "other" to TvContractCompat.PreviewPrograms.TYPE_CLIP
        )
        expectedMappings.forEach { (mediaType, expectedType) ->
            assertEquals(
                "Failed for mediaType=$mediaType",
                expectedType,
                ChannelProgramMapper.mapToPreviewProgramType(mediaType)
            )
        }
    }

    @Test
    fun `buildDeepLinkUri creates correct URI`() {
        val uri = ChannelProgramMapper.buildDeepLinkUri(42L, "movie")
        assertEquals("catalogizer", uri.scheme)
        assertEquals("media", uri.host)
        assertEquals("/42", uri.path)
        assertEquals("movie", uri.getQueryParameter("type"))
    }

    @Test
    fun `buildDeepLinkUri with null type omits type param`() {
        val uri = ChannelProgramMapper.buildDeepLinkUri(99L, null)
        assertEquals("catalogizer", uri.scheme)
        assertEquals("/99", uri.path)
        assertNull(uri.getQueryParameter("type"))
    }

    @Test
    fun `toPreviewProgramValues maps required fields`() {
        val item = MediaItem(
            id = 1L,
            title = "Test Movie",
            mediaType = "movie",
            description = "A test movie",
            year = 2025,
            duration = 7200L,
            rating = 8.5,
            coverUrl = "https://example.com/poster.jpg"
        )
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 100L)
        assertEquals("Test Movie", values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_TITLE))
        assertEquals("A test movie", values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_SHORT_DESCRIPTION))
        assertEquals(100L, values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID))
        assertEquals(
            TvContractCompat.PreviewPrograms.TYPE_MOVIE,
            values.getAsInteger(TvContractCompat.PreviewPrograms.COLUMN_TYPE)
        )
    }

    @Test
    fun `toPreviewProgramValues handles missing optional fields`() {
        val item = MediaItem(id = 2L, title = "Minimal Item")
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 100L)
        assertEquals("Minimal Item", values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_TITLE))
        assertEquals(100L, values.getAsLong(TvContractCompat.PreviewPrograms.COLUMN_CHANNEL_ID))
    }

    @Test
    fun `toPreviewProgramValues includes poster from externalMetadata`() {
        val item = MediaItem(
            id = 3L,
            title = "Movie With Metadata",
            mediaType = "movie",
            externalMetadata = listOf(
                ExternalMetadata(
                    id = 1L, mediaId = 3L, provider = "tmdb", externalId = "123",
                    title = "Movie With Metadata",
                    posterUrl = "https://tmdb.org/poster.jpg",
                    backdropUrl = "https://tmdb.org/backdrop.jpg"
                )
            )
        )
        val values = ChannelProgramMapper.toPreviewProgramValues(item, channelId = 100L)
        val posterUri = values.getAsString(TvContractCompat.PreviewPrograms.COLUMN_POSTER_ART_URI)
        assertNotNull(posterUri)
        assertTrue(posterUri!!.contains("poster.jpg"))
    }

    @Test
    fun `toWatchNextValues sets WATCH_NEXT_TYPE_CONTINUE for in-progress`() {
        val item = MediaItem(id = 10L, title = "Watching Movie", mediaType = "movie", duration = 7200L, watchProgress = 0.5)
        val values = ChannelProgramMapper.toWatchNextValues(item, watchNextType = TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE)
        assertEquals("Watching Movie", values.getAsString(TvContractCompat.WatchNextPrograms.COLUMN_TITLE))
        assertEquals(TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_CONTINUE, values.getAsInteger(TvContractCompat.WatchNextPrograms.COLUMN_WATCH_NEXT_TYPE))
    }

    @Test
    fun `toWatchNextValues sets WATCH_NEXT_TYPE_NEXT for next episode`() {
        val item = MediaItem(id = 11L, title = "Next Episode", mediaType = "tv_episode", duration = 3600L)
        val values = ChannelProgramMapper.toWatchNextValues(item, watchNextType = TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_NEXT)
        assertEquals(TvContractCompat.WatchNextPrograms.WATCH_NEXT_TYPE_NEXT, values.getAsInteger(TvContractCompat.WatchNextPrograms.COLUMN_WATCH_NEXT_TYPE))
    }
}
