package com.catalogizer.androidtv.ui.screens.media

import org.junit.Assert.assertEquals
import org.junit.Test

/**
 * Leaf-action dispatch tests for the NEW book branch of [leafActionKindFor].
 *
 * Kept in its own file (not folded into MediaDetailScreenTest) so it documents
 * the BOOK extension point on its own and does not disturb the existing helper
 * tests. The pre-existing VIDEO / IMAGE / COMIC behaviour is re-asserted here as
 * an anti-regression guard (§11.4.92 blast-radius) — the book case is purely
 * ADDITIVE and must not steal an image, a comic, or a movie.
 */
class MediaDetailBookDispatchTest {

    // --- the NEW book case ---

    @Test
    fun `book media_type dispatches to BOOK`() {
        assertEquals(
            LeafActionKind.BOOK,
            leafActionKindFor(mediaType = "book", path = "/books/Clean Code.pdf")
        )
    }

    @Test
    fun `pdf path dispatches to BOOK when media_type is absent`() {
        assertEquals(LeafActionKind.BOOK, leafActionKindFor(mediaType = null, path = "/books/manual.pdf"))
    }

    @Test
    fun `book extension match is case-insensitive`() {
        assertEquals(LeafActionKind.BOOK, leafActionKindFor(mediaType = null, path = "/books/MANUAL.PDF"))
    }

    @Test
    fun `book media_type wins even with a non-pdf path`() {
        // A "book" entity whose primary file is a non-.pdf still routes to the
        // reader (which then shows an honest 400 "not a PDF" / 501 message).
        assertEquals(LeafActionKind.BOOK, leafActionKindFor(mediaType = "book", path = "/books/manual.epub"))
    }

    // --- pre-existing behaviour is unchanged (anti-regression) ---

    @Test
    fun `comic media_type still dispatches to COMIC`() {
        assertEquals(LeafActionKind.COMIC, leafActionKindFor(mediaType = "comic", path = "/comics/Saga 001.cbz"))
    }

    @Test
    fun `cbz path still dispatches to COMIC`() {
        assertEquals(LeafActionKind.COMIC, leafActionKindFor(mediaType = null, path = "/comics/issue.cbz"))
    }

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
    fun `book and comic and image and video are four distinct kinds`() {
        val kinds = setOf(
            leafActionKindFor("book"),
            leafActionKindFor("comic"),
            leafActionKindFor("image"),
            leafActionKindFor("movie")
        )
        assertEquals(4, kinds.size)
    }
}
