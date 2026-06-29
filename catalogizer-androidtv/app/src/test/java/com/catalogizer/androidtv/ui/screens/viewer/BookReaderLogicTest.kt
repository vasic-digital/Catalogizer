package com.catalogizer.androidtv.ui.screens.viewer

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Pure-logic tests for the PDF book reader — page-index clamping, page-image URL
 * construction, first/last jumps, the page indicator, and the honest HTTP-error
 * mapping (incl. the §11.4.1 not-a-PDF and unsupported-format messages). No
 * device, no Robolectric, no Compose: these exercise the deterministic helpers
 * that drive the screen's D-pad page transitions and its error states
 * (§11.4.52 autonomous, §11.4.6 proven-not-guessed).
 *
 * Books read LINEARLY — like the comic reader (and unlike the image viewer's
 * wraparound), the book helpers CLAMP at both ends (you cannot turn past the
 * back cover or before the cover). These tests pin that difference explicitly.
 */
class BookReaderLogicTest {

    // --- next/prev page: clamp at both ends, NO wraparound ---

    @Test
    fun `next page advances within range`() {
        assertEquals(1, nextPdfPage(current = 0, total = 58))
        assertEquals(6, nextPdfPage(current = 5, total = 58))
    }

    @Test
    fun `next page clamps at the last page (no wrap)`() {
        // current = 57 is the back cover of a 58-page book; next stays put.
        assertEquals(57, nextPdfPage(current = 57, total = 58))
    }

    @Test
    fun `prev page retreats within range`() {
        assertEquals(4, prevPdfPage(current = 5, total = 58))
        assertEquals(0, prevPdfPage(current = 1, total = 58))
    }

    @Test
    fun `prev page clamps at the first page (no wrap)`() {
        // current = 0 is the cover; prev stays put rather than wrapping to 57.
        assertEquals(0, prevPdfPage(current = 0, total = 58))
    }

    @Test
    fun `single-page book stays on page zero in both directions`() {
        assertEquals(0, nextPdfPage(current = 0, total = 1))
        assertEquals(0, prevPdfPage(current = 0, total = 1))
    }

    @Test
    fun `page math is safe for an empty or unknown page set`() {
        assertEquals(0, nextPdfPage(current = 0, total = 0))
        assertEquals(0, prevPdfPage(current = 0, total = 0))
        assertEquals(0, nextPdfPage(current = 3, total = 0))
        assertEquals(0, prevPdfPage(current = 3, total = -1))
    }

    // --- first / last jumps ---

    @Test
    fun `first page is always zero`() {
        assertEquals(0, firstPdfPage())
    }

    @Test
    fun `last page is total minus one`() {
        assertEquals(57, lastPdfPage(total = 58))
        assertEquals(0, lastPdfPage(total = 1))
    }

    @Test
    fun `last page is safe for an empty set`() {
        assertEquals(0, lastPdfPage(total = 0))
    }

    // --- page-image URL construction (matches pdf_pages_handler.go: 0-based) ---

    @Test
    fun `page url is built from base, id and zero-based index`() {
        assertEquals(
            "http://host:8080/api/v1/entities/42/pdf-pages/0",
            pdfPageUrl("http://host:8080", mediaId = 42L, pageIndex = 0)
        )
        assertEquals(
            "http://host:8080/api/v1/entities/42/pdf-pages/7",
            pdfPageUrl("http://host:8080", mediaId = 42L, pageIndex = 7)
        )
    }

    @Test
    fun `page url does not double the base trailing slash`() {
        assertEquals(
            "http://host:8080/api/v1/entities/9/pdf-pages/3",
            pdfPageUrl("http://host:8080/", mediaId = 9L, pageIndex = 3)
        )
    }

    // --- page indicator ---

    @Test
    fun `page label shows one-based page over total`() {
        assertEquals("Page 1 / 58", pdfPageLabel(currentIndex = 0, total = 58))
        assertEquals("Page 58 / 58", pdfPageLabel(currentIndex = 57, total = 58))
    }

    @Test
    fun `page label shows for a single-page book`() {
        assertEquals("Page 1 / 1", pdfPageLabel(currentIndex = 0, total = 1))
    }

    @Test
    fun `page label is null when there are no readable pages or index out of range`() {
        assertNull(pdfPageLabel(currentIndex = 0, total = 0))
        assertNull(pdfPageLabel(currentIndex = -1, total = 58))
        assertNull(pdfPageLabel(currentIndex = 58, total = 58))
    }

    // --- honest HTTP-error mapping (§11.4.1) ---

    @Test
    fun `501 with a known extension names the format as not yet supported`() {
        val msg = bookErrorForHttp(code = 501, ext = "epub")
        assertTrue("must name .epub: $msg", msg.contains(".epub"))
        assertTrue("must say not yet supported: $msg", msg.contains("not yet supported"))
    }

    @Test
    fun `501 without a known extension still says not yet supported`() {
        val msg = bookErrorForHttp(code = 501, ext = null)
        assertTrue(msg.contains("not yet supported"))
    }

    @Test
    fun `400 explains the file is not a PDF document`() {
        val msg = bookErrorForHttp(code = 400, ext = "txt")
        assertTrue("must say not a PDF: $msg", msg.contains("not a PDF document"))
    }

    @Test
    fun `auth errors prompt re-sign-in`() {
        assertTrue(bookErrorForHttp(401, null).contains("Authentication required"))
        assertTrue(bookErrorForHttp(403, null).contains("Authentication required"))
    }

    @Test
    fun `404 explains no file is linked`() {
        assertTrue(bookErrorForHttp(404, "pdf").contains("No file linked"))
    }

    @Test
    fun `500 explains a server-side open failure`() {
        assertTrue(bookErrorForHttp(500, "pdf").contains("Server error"))
    }

    @Test
    fun `unknown code is surfaced honestly with the code`() {
        assertTrue(bookErrorForHttp(418, null).contains("418"))
    }
}
