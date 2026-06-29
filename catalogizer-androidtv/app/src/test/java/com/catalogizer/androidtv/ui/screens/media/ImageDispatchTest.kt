package com.catalogizer.androidtv.ui.screens.media

import org.junit.Assert.assertEquals
import org.junit.Test

/**
 * Tests the leaf-action dispatch classifier that decides whether the
 * MediaDetailScreen primary button opens the image viewer or the video player.
 * This is the fix for the defect where selecting an image leaf routed to the
 * video player. Pure + device-free (§11.4.6 — the routing decision is proven,
 * not guessed; §11.4.1 — an unknown type falls back to the existing working
 * video route, never a dead end).
 */
class ImageDispatchTest {

    // --- authoritative backend media_type ---

    @Test
    fun `image media_type dispatches to IMAGE`() {
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor("image", null))
    }

    @Test
    fun `movie media_type dispatches to VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor("movie", null))
    }

    @Test
    fun `tv_episode media_type dispatches to VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor("tv_episode", null))
    }

    @Test
    fun `null media_type with no path falls back to VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor(null, null))
    }

    // --- file-extension fallback when media_type is absent/unknown ---

    @Test
    fun `jpg path dispatches to IMAGE when type unknown`() {
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor(null, "/photos/holiday.jpg"))
    }

    @Test
    fun `png jpeg webp gif bmp heic paths all dispatch to IMAGE`() {
        for (ext in listOf("png", "jpeg", "webp", "gif", "bmp", "heic", "heif", "tiff")) {
            assertEquals(
                "ext .$ext should classify as IMAGE",
                LeafActionKind.IMAGE,
                leafActionKindFor(null, "/x/file.$ext")
            )
        }
    }

    @Test
    fun `extension match is case-insensitive`() {
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor(null, "/x/PHOTO.JPG"))
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor("unknown", "/x/PHOTO.PnG"))
    }

    @Test
    fun `mp4 path stays VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor(null, "/movies/film.mp4"))
    }

    @Test
    fun `path without extension stays VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor(null, "/movies/film"))
    }

    @Test
    fun `explicit image media_type wins even with a video-looking path`() {
        // The backend type is authoritative; a misleading path never downgrades it.
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor("image", "/weird/name.mp4"))
    }
}
