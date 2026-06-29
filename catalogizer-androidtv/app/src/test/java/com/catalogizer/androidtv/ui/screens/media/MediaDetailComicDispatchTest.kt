package com.catalogizer.androidtv.ui.screens.media

import org.junit.Assert.assertEquals
import org.junit.Test

/**
 * Leaf-action dispatch tests for the NEW comic branch of [leafActionKindFor].
 *
 * Kept in its own file (not folded into MediaDetailScreenTest) so it documents
 * the COMIC extension point on its own and does not disturb the existing helper
 * tests. The pre-existing VIDEO / IMAGE behaviour is re-asserted here as an
 * anti-regression guard (§11.4.92 blast-radius) — the comic case is purely
 * ADDITIVE and must not steal an image or a movie.
 */
class MediaDetailComicDispatchTest {

    // --- the NEW comic case ---

    @Test
    fun `comic media_type dispatches to COMIC`() {
        assertEquals(
            LeafActionKind.COMIC,
            leafActionKindFor(mediaType = "comic", path = "/comics/Saga 001.cbz")
        )
    }

    @Test
    fun `cbz path dispatches to COMIC when media_type is absent`() {
        assertEquals(LeafActionKind.COMIC, leafActionKindFor(mediaType = null, path = "/comics/issue.cbz"))
    }

    @Test
    fun `cbr path dispatches to COMIC so the reader can show the honest 501 message`() {
        // .cbr is routed to the reader (not the player) precisely so the user
        // sees "not yet supported" rather than the player silently failing.
        assertEquals(LeafActionKind.COMIC, leafActionKindFor(mediaType = null, path = "/comics/issue.cbr"))
    }

    @Test
    fun `comic extension match is case-insensitive`() {
        assertEquals(LeafActionKind.COMIC, leafActionKindFor(mediaType = null, path = "/comics/ISSUE.CBZ"))
    }

    @Test
    fun `comic media_type wins even with a non-archive path`() {
        // A "comic" entity whose primary file is a .pdf/.epub still routes to the
        // reader (which then shows an honest 400 "not a supported archive").
        assertEquals(LeafActionKind.COMIC, leafActionKindFor(mediaType = "comic", path = "/comics/issue.pdf"))
    }

    // --- pre-existing behaviour is unchanged (anti-regression) ---

    @Test
    fun `image media_type still dispatches to IMAGE`() {
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor(mediaType = "image", path = "/p/photo.jpg"))
    }

    @Test
    fun `image-extension path still dispatches to IMAGE`() {
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor(mediaType = null, path = "/p/photo.png"))
        assertEquals(LeafActionKind.IMAGE, leafActionKindFor(mediaType = null, path = "/p/photo.webp"))
    }

    @Test
    fun `movie still dispatches to VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor(mediaType = "movie", path = "/m/heat.mkv"))
    }

    @Test
    fun `unknown leaf still falls back to VIDEO`() {
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor(mediaType = null, path = "/x/file.mp4"))
        assertEquals(LeafActionKind.VIDEO, leafActionKindFor(mediaType = null, path = null))
    }

    @Test
    fun `comic and image and video are three distinct kinds`() {
        val kinds = setOf(
            leafActionKindFor("comic"),
            leafActionKindFor("image"),
            leafActionKindFor("movie")
        )
        assertEquals(3, kinds.size)
    }
}
