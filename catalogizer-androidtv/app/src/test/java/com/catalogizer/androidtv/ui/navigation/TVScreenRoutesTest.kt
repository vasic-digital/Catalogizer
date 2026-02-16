package com.catalogizer.androidtv.ui.navigation

import org.junit.Assert.*
import org.junit.Test

class TVScreenRoutesTest {

    @Test
    fun `Login screen has correct route`() {
        assertEquals("login", TVScreen.Login.route)
    }

    @Test
    fun `Home screen has correct route`() {
        assertEquals("home", TVScreen.Home.route)
    }

    @Test
    fun `Search screen has correct route`() {
        assertEquals("search", TVScreen.Search.route)
    }

    @Test
    fun `Settings screen has correct route`() {
        assertEquals("settings", TVScreen.Settings.route)
    }

    @Test
    fun `MediaDetail screen has route with parameter`() {
        assertEquals("media_detail/{mediaId}", TVScreen.MediaDetail.route)
    }

    @Test
    fun `MediaDetail createRoute produces correct route`() {
        assertEquals("media_detail/42", TVScreen.MediaDetail.createRoute(42L))
        assertEquals("media_detail/1", TVScreen.MediaDetail.createRoute(1L))
        assertEquals("media_detail/0", TVScreen.MediaDetail.createRoute(0L))
    }

    @Test
    fun `Player screen has route with parameter`() {
        assertEquals("player/{mediaId}", TVScreen.Player.route)
    }

    @Test
    fun `Player createRoute produces correct route`() {
        assertEquals("player/42", TVScreen.Player.createRoute(42L))
        assertEquals("player/1", TVScreen.Player.createRoute(1L))
        assertEquals("player/0", TVScreen.Player.createRoute(0L))
    }

    @Test
    fun `all screen routes are unique`() {
        val routes = listOf(
            TVScreen.Login.route,
            TVScreen.Home.route,
            TVScreen.Search.route,
            TVScreen.Settings.route,
            TVScreen.MediaDetail.route,
            TVScreen.Player.route
        )

        assertEquals(routes.size, routes.toSet().size)
    }

    @Test
    fun `TVScreen is sealed class with six subclasses`() {
        val screens = listOf(
            TVScreen.Login,
            TVScreen.Home,
            TVScreen.Search,
            TVScreen.Settings,
            TVScreen.MediaDetail,
            TVScreen.Player
        )

        assertEquals(6, screens.size)
        screens.forEach { screen ->
            assertTrue(screen is TVScreen)
            assertTrue(screen.route.isNotBlank())
        }
    }

    @Test
    fun `MediaDetail and Player have different routes for same mediaId`() {
        val mediaId = 42L
        val detailRoute = TVScreen.MediaDetail.createRoute(mediaId)
        val playerRoute = TVScreen.Player.createRoute(mediaId)

        assertNotEquals(detailRoute, playerRoute)
        assertTrue(detailRoute.startsWith("media_detail/"))
        assertTrue(playerRoute.startsWith("player/"))
    }
}
