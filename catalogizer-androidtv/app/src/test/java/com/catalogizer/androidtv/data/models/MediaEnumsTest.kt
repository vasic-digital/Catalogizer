package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaEnumsTest {

    // --- MediaType ---

    @Test
    fun `MediaType fromValue returns correct type for known values`() {
        assertEquals(MediaType.MOVIE, MediaType.fromValue("movie"))
        assertEquals(MediaType.TV_SHOW, MediaType.fromValue("tv_show"))
        assertEquals(MediaType.DOCUMENTARY, MediaType.fromValue("documentary"))
        assertEquals(MediaType.ANIME, MediaType.fromValue("anime"))
        assertEquals(MediaType.MUSIC, MediaType.fromValue("music"))
        assertEquals(MediaType.AUDIOBOOK, MediaType.fromValue("audiobook"))
        assertEquals(MediaType.PODCAST, MediaType.fromValue("podcast"))
    }

    @Test
    fun `MediaType fromValue returns OTHER for unknown values`() {
        assertEquals(MediaType.OTHER, MediaType.fromValue("unknown"))
        assertEquals(MediaType.OTHER, MediaType.fromValue(""))
    }

    @Test
    fun `MediaType getAllTypes returns all types`() {
        val allTypes = MediaType.getAllTypes()
        assertEquals(MediaType.values().size, allTypes.size)
    }

    @Test
    fun `MediaType getVideoTypes returns correct video types`() {
        val videoTypes = MediaType.getVideoTypes()

        assertTrue(videoTypes.contains(MediaType.MOVIE))
        assertTrue(videoTypes.contains(MediaType.TV_SHOW))
        assertTrue(videoTypes.contains(MediaType.DOCUMENTARY))
        assertTrue(videoTypes.contains(MediaType.ANIME))
        assertTrue(videoTypes.contains(MediaType.CONCERT))
        assertTrue(videoTypes.contains(MediaType.YOUTUBE_VIDEO))
        assertTrue(videoTypes.contains(MediaType.SPORTS))
        assertTrue(videoTypes.contains(MediaType.NEWS))
        assertTrue(videoTypes.contains(MediaType.TRAINING))
        assertFalse(videoTypes.contains(MediaType.MUSIC))
        assertFalse(videoTypes.contains(MediaType.AUDIOBOOK))
    }

    @Test
    fun `MediaType getAudioTypes returns correct audio types`() {
        val audioTypes = MediaType.getAudioTypes()

        assertTrue(audioTypes.contains(MediaType.MUSIC))
        assertTrue(audioTypes.contains(MediaType.AUDIOBOOK))
        assertTrue(audioTypes.contains(MediaType.PODCAST))
        assertFalse(audioTypes.contains(MediaType.MOVIE))
        assertFalse(audioTypes.contains(MediaType.TV_SHOW))
    }

    @Test
    fun `MediaType has correct display names`() {
        assertEquals("Movies", MediaType.MOVIE.displayName)
        assertEquals("TV Shows", MediaType.TV_SHOW.displayName)
        assertEquals("Music", MediaType.MUSIC.displayName)
        assertEquals("E-books", MediaType.EBOOK.displayName)
    }

    // --- QualityLevel ---

    @Test
    fun `QualityLevel fromValue returns correct quality`() {
        assertEquals(QualityLevel.HD_720P, QualityLevel.fromValue("720p"))
        assertEquals(QualityLevel.HD_1080P, QualityLevel.fromValue("1080p"))
        assertEquals(QualityLevel.UHD_4K, QualityLevel.fromValue("4k"))
        assertEquals(QualityLevel.DOLBY_VISION, QualityLevel.fromValue("dolby_vision"))
    }

    @Test
    fun `QualityLevel fromValue returns null for unknown`() {
        assertNull(QualityLevel.fromValue("unknown"))
        assertNull(QualityLevel.fromValue(""))
    }

    @Test
    fun `QualityLevel getAllQualities returns all levels`() {
        assertEquals(QualityLevel.values().size, QualityLevel.getAllQualities().size)
    }

    // --- SortOption ---

    @Test
    fun `SortOption fromValue returns correct option`() {
        assertEquals(SortOption.TITLE, SortOption.fromValue("title"))
        assertEquals(SortOption.YEAR, SortOption.fromValue("year"))
        assertEquals(SortOption.RATING, SortOption.fromValue("rating"))
        assertEquals(SortOption.DURATION, SortOption.fromValue("duration"))
    }

    @Test
    fun `SortOption fromValue returns UPDATED_AT for unknown`() {
        assertEquals(SortOption.UPDATED_AT, SortOption.fromValue("unknown"))
    }

    @Test
    fun `SortOption has DURATION option for TV`() {
        assertEquals("duration", SortOption.DURATION.value)
        assertEquals("Duration", SortOption.DURATION.displayName)
    }

    // --- SortOrder ---

    @Test
    fun `SortOrder fromValue returns correct order`() {
        assertEquals(SortOrder.ASC, SortOrder.fromValue("asc"))
        assertEquals(SortOrder.DESC, SortOrder.fromValue("desc"))
    }

    @Test
    fun `SortOrder fromValue returns DESC for unknown`() {
        assertEquals(SortOrder.DESC, SortOrder.fromValue("unknown"))
    }

    // --- PlaybackProgress ---

    @Test
    fun `PlaybackProgress calculates percentage correctly`() {
        val progress = PlaybackProgress(mediaId = 1L, position = 3000, duration = 6000)
        assertEquals(0.5, progress.progressPercentage, 0.01)
    }

    @Test
    fun `PlaybackProgress with zero duration returns zero percentage`() {
        val progress = PlaybackProgress(mediaId = 1L, position = 3000, duration = 0)
        assertEquals(0.0, progress.progressPercentage, 0.01)
    }

    @Test
    fun `PlaybackProgress at start returns zero percentage`() {
        val progress = PlaybackProgress(mediaId = 1L, position = 0, duration = 6000)
        assertEquals(0.0, progress.progressPercentage, 0.01)
    }

    @Test
    fun `PlaybackProgress at end returns full percentage`() {
        val progress = PlaybackProgress(mediaId = 1L, position = 6000, duration = 6000)
        assertEquals(1.0, progress.progressPercentage, 0.01)
    }

    @Test
    fun `PlaybackProgress has default timestamp`() {
        val progress = PlaybackProgress(mediaId = 1L, position = 0, duration = 0)
        assertTrue(progress.timestamp > 0)
    }
}
