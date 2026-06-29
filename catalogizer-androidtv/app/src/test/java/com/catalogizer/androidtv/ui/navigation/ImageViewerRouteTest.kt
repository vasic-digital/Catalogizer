package com.catalogizer.androidtv.ui.navigation

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Route tests for the new image-viewer destination. Kept in its own file so it
 * does not disturb the existing TVScreenRoutesTest's hand-built screen lists.
 */
class ImageViewerRouteTest {

    @Test
    fun `ImageViewer screen has route with parameter`() {
        assertEquals("image_viewer/{mediaId}", TVScreen.ImageViewer.route)
    }

    @Test
    fun `ImageViewer createRoute produces correct route`() {
        assertEquals("image_viewer/42", TVScreen.ImageViewer.createRoute(42L))
        assertEquals("image_viewer/1", TVScreen.ImageViewer.createRoute(1L))
        assertEquals("image_viewer/0", TVScreen.ImageViewer.createRoute(0L))
    }

    @Test
    fun `ImageViewer route differs from Player route for same mediaId`() {
        val mediaId = 42L
        val imageRoute = TVScreen.ImageViewer.createRoute(mediaId)
        val playerRoute = TVScreen.Player.createRoute(mediaId)

        assertNotEquals(imageRoute, playerRoute)
        assertTrue(imageRoute.startsWith("image_viewer/"))
        assertTrue(playerRoute.startsWith("player/"))
    }
}
