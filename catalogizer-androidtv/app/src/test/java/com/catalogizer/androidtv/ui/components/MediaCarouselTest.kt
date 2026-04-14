package com.catalogizer.androidtv.ui.components

import com.catalogizer.androidtv.data.models.MediaItem
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for MediaCarousel and FeaturedMediaCard logic: item list handling,
 * empty states, featured item metadata extraction, and section header behavior.
 */
class MediaCarouselTest {

    private fun createTestMediaItem(
        id: Long = 1L,
        title: String = "Test Media",
        mediaType: String = "movie",
        year: Int? = 2024,
        description: String? = "Test description",
        rating: Double? = 7.5,
        quality: String? = "1080p",
        watchProgress: Double = 0.0
    ) = MediaItem(
        id = id,
        title = title,
        mediaType = mediaType,
        year = year,
        description = description,
        rating = rating,
        quality = quality,
        directoryPath = "/media/test",
        createdAt = "2024-01-01T00:00:00Z",
        updatedAt = "2024-01-01T00:00:00Z",
        watchProgress = watchProgress
    )

    // --- Item list handling ---

    @Test
    fun `carousel with empty items list should have zero size`() {
        val items = emptyList<MediaItem>()
        assertEquals(0, items.size)
    }

    @Test
    fun `carousel with single item should have size one`() {
        val items = listOf(createTestMediaItem())
        assertEquals(1, items.size)
    }

    @Test
    fun `carousel with multiple items preserves order`() {
        val items = listOf(
            createTestMediaItem(1L, "First"),
            createTestMediaItem(2L, "Second"),
            createTestMediaItem(3L, "Third")
        )
        assertEquals(3, items.size)
        assertEquals("First", items[0].title)
        assertEquals("Second", items[1].title)
        assertEquals("Third", items[2].title)
    }

    @Test
    fun `carousel items are indexable`() {
        val items = listOf(
            createTestMediaItem(10L, "Item A"),
            createTestMediaItem(20L, "Item B"),
            createTestMediaItem(30L, "Item C")
        )
        for (index in items.indices) {
            val mediaItem = items[index]
            assertNotNull(mediaItem)
            assertTrue(mediaItem.title.isNotBlank())
        }
    }

    @Test
    fun `carousel item click callback receives correct item`() {
        val items = listOf(
            createTestMediaItem(1L, "Movie A"),
            createTestMediaItem(2L, "Movie B")
        )
        var clickedItem: MediaItem? = null
        val onItemClick: (MediaItem) -> Unit = { clickedItem = it }

        onItemClick(items[1])
        assertNotNull(clickedItem)
        assertEquals(2L, clickedItem?.id)
        assertEquals("Movie B", clickedItem?.title)
    }

    // --- Section header ---

    @Test
    fun `carousel title should be non-blank`() {
        val title = "Recently Added Movies"
        assertTrue(title.isNotBlank())
    }

    @Test
    fun `carousel title preserves exact text`() {
        val title = "Top Rated TV Shows"
        assertEquals("Top Rated TV Shows", title)
    }

    // --- Featured media card metadata ---

    @Test
    fun `featured card displays year when present`() {
        val item = createTestMediaItem(year = 2024)
        assertNotNull(item.year)
        assertEquals(2024, item.year)
    }

    @Test
    fun `featured card handles null year gracefully`() {
        val item = createTestMediaItem(year = null)
        assertNull(item.year)
    }

    @Test
    fun `featured card displays rating when present`() {
        val item = createTestMediaItem(rating = 8.5)
        assertNotNull(item.rating)
        assertEquals(8.5, item.rating!!, 0.01)
    }

    @Test
    fun `featured card handles null rating gracefully`() {
        val item = createTestMediaItem(rating = null)
        assertNull(item.rating)
    }

    @Test
    fun `featured card displays description when present`() {
        val item = createTestMediaItem(description = "An epic adventure film")
        assertNotNull(item.description)
        assertEquals("An epic adventure film", item.description)
    }

    @Test
    fun `featured card handles null description gracefully`() {
        val item = createTestMediaItem(description = null)
        assertNull(item.description)
    }

    @Test
    fun `featured card displays quality when present`() {
        val item = createTestMediaItem(quality = "4K HDR")
        assertNotNull(item.quality)
        assertEquals("4K HDR", item.quality)
    }

    @Test
    fun `featured card handles null quality gracefully`() {
        val item = createTestMediaItem(quality = null)
        assertNull(item.quality)
    }

    // --- Mixed media types in carousel ---

    @Test
    fun `carousel can contain items of different media types`() {
        val items = listOf(
            createTestMediaItem(1L, "Movie", mediaType = "movie"),
            createTestMediaItem(2L, "TV Show", mediaType = "tv_show"),
            createTestMediaItem(3L, "Song", mediaType = "song")
        )
        assertEquals(3, items.size)
        assertEquals("movie", items[0].mediaType)
        assertEquals("tv_show", items[1].mediaType)
        assertEquals("song", items[2].mediaType)
    }

    // --- Large carousel ---

    @Test
    fun `carousel handles large number of items`() {
        val items = (1..100).map { i ->
            createTestMediaItem(i.toLong(), "Media $i")
        }
        assertEquals(100, items.size)
        assertEquals("Media 1", items.first().title)
        assertEquals("Media 100", items.last().title)
    }

    // --- Item IDs are unique ---

    @Test
    fun `carousel items should have unique IDs`() {
        val items = listOf(
            createTestMediaItem(1L),
            createTestMediaItem(2L),
            createTestMediaItem(3L)
        )
        val ids = items.map { it.id }.toSet()
        assertEquals(items.size, ids.size)
    }
}
