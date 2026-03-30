package com.catalogizer.androidtv.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaTypeTest {

    @Test
    fun `fromValue returns MOVIE for movie`() {
        assertEquals(MediaType.MOVIE, MediaType.fromValue("movie"))
    }

    @Test
    fun `fromValue returns TV_SHOW for tv_show`() {
        assertEquals(MediaType.TV_SHOW, MediaType.fromValue("tv_show"))
    }

    @Test
    fun `fromValue returns OTHER for unknown`() {
        assertEquals(MediaType.OTHER, MediaType.fromValue("unknown"))
    }

    @Test
    fun `fromValue returns OTHER for empty string`() {
        assertEquals(MediaType.OTHER, MediaType.fromValue(""))
    }

    @Test
    fun `getAllTypes returns all 16 types`() {
        assertEquals(16, MediaType.getAllTypes().size)
    }

    @Test
    fun `getVideoTypes returns correct types`() {
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
        assertEquals(9, videoTypes.size)
    }

    @Test
    fun `getAudioTypes returns correct types`() {
        val audioTypes = MediaType.getAudioTypes()
        assertTrue(audioTypes.contains(MediaType.MUSIC))
        assertTrue(audioTypes.contains(MediaType.AUDIOBOOK))
        assertTrue(audioTypes.contains(MediaType.PODCAST))
        assertEquals(3, audioTypes.size)
    }

    @Test
    fun `video and audio types do not overlap`() {
        val videoTypes = MediaType.getVideoTypes().toSet()
        val audioTypes = MediaType.getAudioTypes().toSet()
        assertTrue(videoTypes.intersect(audioTypes).isEmpty())
    }

    @Test
    fun `display names are human readable`() {
        assertEquals("Movies", MediaType.MOVIE.displayName)
        assertEquals("TV Shows", MediaType.TV_SHOW.displayName)
        assertEquals("Music", MediaType.MUSIC.displayName)
        assertEquals("Games", MediaType.GAME.displayName)
        assertEquals("E-books", MediaType.EBOOK.displayName)
    }

    @Test
    fun `values are lowercase`() {
        for (type in MediaType.values()) {
            assertEquals(type.value, type.value.lowercase())
        }
    }
}
