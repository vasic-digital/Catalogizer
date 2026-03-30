package com.catalogizer.android.data.models

import org.junit.Assert.*
import org.junit.Test

class MediaTypeTest {

    @Test
    fun `fromValue returns MOVIE for movie string`() {
        assertEquals(MediaType.MOVIE, MediaType.fromValue("movie"))
    }

    @Test
    fun `fromValue returns TV_SHOW for tv_show string`() {
        assertEquals(MediaType.TV_SHOW, MediaType.fromValue("tv_show"))
    }

    @Test
    fun `fromValue returns OTHER for unknown string`() {
        assertEquals(MediaType.OTHER, MediaType.fromValue("unknown_type"))
    }

    @Test
    fun `fromValue returns OTHER for empty string`() {
        assertEquals(MediaType.OTHER, MediaType.fromValue(""))
    }

    @Test
    fun `getAllTypes returns all 16 media types`() {
        val allTypes = MediaType.getAllTypes()
        assertEquals(16, allTypes.size)
    }

    @Test
    fun `all media types have non-empty values`() {
        for (type in MediaType.values()) {
            assertTrue("MediaType ${type.name} should have non-empty value", type.value.isNotEmpty())
        }
    }

    @Test
    fun `all media types have non-empty display names`() {
        for (type in MediaType.values()) {
            assertTrue("MediaType ${type.name} should have non-empty displayName", type.displayName.isNotEmpty())
        }
    }

    @Test
    fun `display names are human readable`() {
        assertEquals("Movies", MediaType.MOVIE.displayName)
        assertEquals("TV Shows", MediaType.TV_SHOW.displayName)
        assertEquals("Documentaries", MediaType.DOCUMENTARY.displayName)
        assertEquals("Anime", MediaType.ANIME.displayName)
        assertEquals("Music", MediaType.MUSIC.displayName)
        assertEquals("Games", MediaType.GAME.displayName)
        assertEquals("Software", MediaType.SOFTWARE.displayName)
    }

    @Test
    fun `values are lowercase with underscores`() {
        for (type in MediaType.values()) {
            assertEquals(
                "MediaType ${type.name} value should be lowercase",
                type.value,
                type.value.lowercase()
            )
        }
    }
}
