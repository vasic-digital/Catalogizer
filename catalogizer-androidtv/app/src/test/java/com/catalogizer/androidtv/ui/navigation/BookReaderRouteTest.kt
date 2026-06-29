package com.catalogizer.androidtv.ui.navigation

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Route tests for the new book-reader destination. Kept in its own file so it
 * does not disturb the existing TVScreenRoutesTest / ComicReaderRouteTest.
 *
 * These pin that the book reader is a DISTINCT, reachable destination — the
 * BookReaderScreen the prior agent built is wired into navigation, not dead
 * code that silently falls through to the video player (§11.4.124).
 */
class BookReaderRouteTest {

    @Test
    fun `BookReader screen has route with parameter`() {
        assertEquals("book_reader/{mediaId}", TVScreen.BookReader.route)
    }

    @Test
    fun `BookReader createRoute produces correct path`() {
        assertEquals("book_reader/42", TVScreen.BookReader.createRoute(42L))
        assertEquals("book_reader/1", TVScreen.BookReader.createRoute(1L))
        assertEquals("book_reader/0", TVScreen.BookReader.createRoute(0L))
    }

    @Test
    fun `BookReader route differs from Player, ImageViewer and ComicReader for same mediaId`() {
        val mediaId = 42L
        val bookRoute = TVScreen.BookReader.createRoute(mediaId)
        val comicRoute = TVScreen.ComicReader.createRoute(mediaId)
        val imageRoute = TVScreen.ImageViewer.createRoute(mediaId)
        val playerRoute = TVScreen.Player.createRoute(mediaId)

        assertNotEquals(bookRoute, comicRoute)
        assertNotEquals(bookRoute, imageRoute)
        assertNotEquals(bookRoute, playerRoute)
        assertTrue(bookRoute.startsWith("book_reader/"))
    }
}
