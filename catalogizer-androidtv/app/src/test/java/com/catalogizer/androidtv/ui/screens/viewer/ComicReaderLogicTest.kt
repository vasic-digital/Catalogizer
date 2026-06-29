package com.catalogizer.androidtv.ui.screens.viewer

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Pure-logic tests for the comic reader — page-index clamping, page-image URL
 * construction, first/last jumps, the page indicator, and the honest HTTP-error
 * mapping (incl. the §11.4.1 `.cbr`-not-supported message). No device, no
 * Robolectric, no Compose: these exercise the deterministic helpers that drive
 * the screen's D-pad page transitions and its error states (§11.4.52 autonomous,
 * §11.4.6 proven-not-guessed).
 *
 * Comics read LINEARLY — unlike the image viewer's sibling wraparound, the comic
 * helpers CLAMP at both ends (you cannot turn past the back cover or before the
 * cover). These tests pin that difference explicitly.
 */
class ComicReaderLogicTest {

    // --- next/prev page: clamp at both ends, NO wraparound ---

    @Test
    fun `next page advances within range`() {
        assertEquals(1, nextComicPage(current = 0, total = 12))
        assertEquals(6, nextComicPage(current = 5, total = 12))
    }

    @Test
    fun `next page clamps at the last page (no wrap)`() {
        // current = 11 is the back cover of a 12-page comic; next stays put.
        assertEquals(11, nextComicPage(current = 11, total = 12))
    }

    @Test
    fun `prev page retreats within range`() {
        assertEquals(4, prevComicPage(current = 5, total = 12))
        assertEquals(0, prevComicPage(current = 1, total = 12))
    }

    @Test
    fun `prev page clamps at the first page (no wrap)`() {
        // current = 0 is the cover; prev stays put rather than wrapping to 11.
        assertEquals(0, prevComicPage(current = 0, total = 12))
    }

    @Test
    fun `single-page comic stays on page zero in both directions`() {
        assertEquals(0, nextComicPage(current = 0, total = 1))
        assertEquals(0, prevComicPage(current = 0, total = 1))
    }

    @Test
    fun `page math is safe for an empty or unknown page set`() {
        assertEquals(0, nextComicPage(current = 0, total = 0))
        assertEquals(0, prevComicPage(current = 0, total = 0))
        assertEquals(0, nextComicPage(current = 3, total = 0))
    }

    // --- first / last jumps ---

    @Test
    fun `first page is always zero`() {
        assertEquals(0, firstComicPage())
    }

    @Test
    fun `last page is total minus one`() {
        assertEquals(11, lastComicPage(total = 12))
        assertEquals(0, lastComicPage(total = 1))
    }

    @Test
    fun `last page is safe for an empty set`() {
        assertEquals(0, lastComicPage(total = 0))
    }

    // --- page-image URL construction (matches comic_pages_handler.go: 0-based) ---

    @Test
    fun `page url is built from base, id and zero-based index`() {
        assertEquals(
            "http://host:8080/api/v1/entities/42/pages/0",
            comicPageUrl("http://host:8080", mediaId = 42L, pageIndex = 0)
        )
        assertEquals(
            "http://host:8080/api/v1/entities/42/pages/7",
            comicPageUrl("http://host:8080", mediaId = 42L, pageIndex = 7)
        )
    }

    @Test
    fun `page url does not double the base trailing slash`() {
        assertEquals(
            "http://host:8080/api/v1/entities/9/pages/3",
            comicPageUrl("http://host:8080/", mediaId = 9L, pageIndex = 3)
        )
    }

    // --- page indicator ---

    @Test
    fun `page label shows one-based page over total`() {
        assertEquals("Page 1 / 12", comicPageLabel(currentIndex = 0, total = 12))
        assertEquals("Page 12 / 12", comicPageLabel(currentIndex = 11, total = 12))
    }

    @Test
    fun `page label shows for a single-page comic`() {
        assertEquals("Page 1 / 1", comicPageLabel(currentIndex = 0, total = 1))
    }

    @Test
    fun `page label is null when there are no readable pages or index out of range`() {
        assertNull(comicPageLabel(currentIndex = 0, total = 0))
        assertNull(comicPageLabel(currentIndex = -1, total = 12))
        assertNull(comicPageLabel(currentIndex = 12, total = 12))
    }

    // --- honest HTTP-error mapping (§11.4.1) ---

    @Test
    fun `cbr returns the honest not-yet-supported message naming the format`() {
        // 501 is exactly what comic_pages_handler.go returns for a .cbr archive.
        val msg = comicErrorForHttp(code = 501, ext = "cbr")
        assertTrue("must name .cbr: $msg", msg.contains(".cbr"))
        assertTrue("must say not yet supported: $msg", msg.contains("not yet supported"))
    }

    @Test
    fun `501 without a known extension still says not yet supported`() {
        val msg = comicErrorForHttp(code = 501, ext = null)
        assertTrue(msg.contains("not yet supported"))
    }

    @Test
    fun `non-archive 400 explains the file is not a supported comic`() {
        val msg = comicErrorForHttp(code = 400, ext = "pdf")
        assertTrue(msg.contains("not a supported comic archive"))
    }

    @Test
    fun `auth errors prompt re-sign-in`() {
        assertTrue(comicErrorForHttp(401, null).contains("Authentication required"))
        assertTrue(comicErrorForHttp(403, null).contains("Authentication required"))
    }

    @Test
    fun `404 explains no file is linked`() {
        assertTrue(comicErrorForHttp(404, "cbz").contains("No file linked"))
    }

    @Test
    fun `unknown code is surfaced honestly with the code`() {
        assertTrue(comicErrorForHttp(418, null).contains("418"))
    }
}
