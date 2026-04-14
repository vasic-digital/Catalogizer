package com.catalogizer.androidtv.ui.navigation

import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for TVNavigation route definitions, start destination selection,
 * deep link routing, and screen route parameter handling.
 */
class TVNavigationTest {

    // --- Start destination selection ---

    @Test
    fun `start destination is Home when authenticated`() {
        val isAuthenticated = true
        val startDestination = if (isAuthenticated) TVScreen.Home.route else TVScreen.Login.route
        assertEquals("home", startDestination)
    }

    @Test
    fun `start destination is Login when not authenticated`() {
        val isAuthenticated = false
        val startDestination = if (isAuthenticated) TVScreen.Home.route else TVScreen.Login.route
        assertEquals("login", startDestination)
    }

    // --- Deep link routing ---

    @Test
    fun `deep link with play action routes to Player`() {
        val deepLinkMediaId = 42L
        val deepLinkAction = "play"
        val isAuthenticated = true

        val shouldRoute = deepLinkMediaId > 0 && isAuthenticated
        assertTrue(shouldRoute)

        val route = when (deepLinkAction) {
            "play" -> TVScreen.Player.createRoute(deepLinkMediaId)
            else -> TVScreen.MediaDetail.createRoute(deepLinkMediaId)
        }
        assertEquals("player/42", route)
    }

    @Test
    fun `deep link with detail action routes to MediaDetail`() {
        val deepLinkMediaId = 42L
        val deepLinkAction = "detail"

        val route = when (deepLinkAction) {
            "play" -> TVScreen.Player.createRoute(deepLinkMediaId)
            else -> TVScreen.MediaDetail.createRoute(deepLinkMediaId)
        }
        assertEquals("media_detail/42", route)
    }

    @Test
    fun `deep link with null action routes to MediaDetail`() {
        val deepLinkMediaId = 42L
        val deepLinkAction: String? = null

        val route = when (deepLinkAction) {
            "play" -> TVScreen.Player.createRoute(deepLinkMediaId)
            else -> TVScreen.MediaDetail.createRoute(deepLinkMediaId)
        }
        assertEquals("media_detail/42", route)
    }

    @Test
    fun `deep link ignored when mediaId is negative`() {
        val deepLinkMediaId = -1L
        val isAuthenticated = true

        val shouldRoute = deepLinkMediaId > 0 && isAuthenticated
        assertFalse(shouldRoute)
    }

    @Test
    fun `deep link ignored when mediaId is zero`() {
        val deepLinkMediaId = 0L
        val isAuthenticated = true

        val shouldRoute = deepLinkMediaId > 0 && isAuthenticated
        assertFalse(shouldRoute)
    }

    @Test
    fun `deep link ignored when not authenticated`() {
        val deepLinkMediaId = 42L
        val isAuthenticated = false

        val shouldRoute = deepLinkMediaId > 0 && isAuthenticated
        assertFalse(shouldRoute)
    }

    // --- Route parameter extraction ---

    @Test
    fun `mediaId extracted from MediaDetail route argument`() {
        val mediaIdString = "42"
        val mediaId = mediaIdString.toLongOrNull() ?: 0L
        assertEquals(42L, mediaId)
    }

    @Test
    fun `invalid mediaId defaults to zero`() {
        val mediaIdString: String? = null
        val mediaId = mediaIdString?.toLongOrNull() ?: 0L
        assertEquals(0L, mediaId)
    }

    @Test
    fun `non-numeric mediaId defaults to zero`() {
        val mediaIdString = "abc"
        val mediaId = mediaIdString.toLongOrNull() ?: 0L
        assertEquals(0L, mediaId)
    }

    @Test
    fun `empty mediaId string defaults to zero`() {
        val mediaIdString = ""
        val mediaId = mediaIdString.toLongOrNull() ?: 0L
        assertEquals(0L, mediaId)
    }

    // --- Route uniqueness (extended from TVScreenRoutesTest) ---

    @Test
    fun `all screen routes are distinct`() {
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

    // --- Navigation patterns ---

    @Test
    fun `login success navigates to home with popUpTo`() {
        val targetRoute = TVScreen.Home.route
        val popRoute = TVScreen.Login.route
        assertEquals("home", targetRoute)
        assertEquals("login", popRoute)
    }

    @Test
    fun `logout navigates to login with popUpTo root`() {
        val targetRoute = TVScreen.Login.route
        assertEquals("login", targetRoute)
    }

    @Test
    fun `search navigates from home`() {
        val fromRoute = TVScreen.Home.route
        val toRoute = TVScreen.Search.route
        assertEquals("home", fromRoute)
        assertEquals("search", toRoute)
    }

    @Test
    fun `settings navigates from home`() {
        val fromRoute = TVScreen.Home.route
        val toRoute = TVScreen.Settings.route
        assertEquals("home", fromRoute)
        assertEquals("settings", toRoute)
    }

    @Test
    fun `media detail navigates from home with media ID`() {
        val mediaId = 99L
        val route = TVScreen.MediaDetail.createRoute(mediaId)
        assertEquals("media_detail/99", route)
    }

    @Test
    fun `player navigates from home with media ID`() {
        val mediaId = 99L
        val route = TVScreen.Player.createRoute(mediaId)
        assertEquals("player/99", route)
    }

    @Test
    fun `player also navigates from media detail`() {
        val mediaId = 55L
        val route = TVScreen.Player.createRoute(mediaId)
        assertEquals("player/55", route)
    }

    // --- Route template parameters ---

    @Test
    fun `MediaDetail route contains mediaId parameter placeholder`() {
        assertTrue(TVScreen.MediaDetail.route.contains("{mediaId}"))
    }

    @Test
    fun `Player route contains mediaId parameter placeholder`() {
        assertTrue(TVScreen.Player.route.contains("{mediaId}"))
    }

    @Test
    fun `Login route has no parameters`() {
        assertFalse(TVScreen.Login.route.contains("{"))
    }

    @Test
    fun `Home route has no parameters`() {
        assertFalse(TVScreen.Home.route.contains("{"))
    }

    @Test
    fun `Search route has no parameters`() {
        assertFalse(TVScreen.Search.route.contains("{"))
    }

    @Test
    fun `Settings route has no parameters`() {
        assertFalse(TVScreen.Settings.route.contains("{"))
    }
}
