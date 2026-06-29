package com.catalogizer.androidtv.ui.navigation

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Route tests for the new comic-reader destination. Kept in its own file so it
 * does not disturb the existing TVScreenRoutesTest / ImageViewerRouteTest.
 */
class ComicReaderRouteTest {

    @Test
    fun `ComicReader screen has route with parameter`() {
        assertEquals("comic_reader/{mediaId}", TVScreen.ComicReader.route)
    }

    @Test
    fun `ComicReader createRoute produces correct path`() {
        assertEquals("comic_reader/42", TVScreen.ComicReader.createRoute(42L))
        assertEquals("comic_reader/1", TVScreen.ComicReader.createRoute(1L))
        assertEquals("comic_reader/0", TVScreen.ComicReader.createRoute(0L))
    }

    @Test
    fun `ComicReader route differs from Player and ImageViewer for same mediaId`() {
        val mediaId = 42L
        val comicRoute = TVScreen.ComicReader.createRoute(mediaId)
        val imageRoute = TVScreen.ImageViewer.createRoute(mediaId)
        val playerRoute = TVScreen.Player.createRoute(mediaId)

        assertNotEquals(comicRoute, imageRoute)
        assertNotEquals(comicRoute, playerRoute)
        assertTrue(comicRoute.startsWith("comic_reader/"))
    }
}
