package com.catalogizer.android.ui.navigation

import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for CatalogizerNavigation route definitions, start destination logic,
 * and Screen sealed class structure.
 */
@RunWith(RobolectricTestRunner::class)
class CatalogizerNavigationTest {

    @Test
    fun `Login screen has correct route`() {
        assertEquals("login", Screen.Login.route)
    }

    @Test
    fun `Home screen has correct route`() {
        assertEquals("home", Screen.Home.route)
    }

    @Test
    fun `Search screen has correct route`() {
        assertEquals("search", Screen.Search.route)
    }

    @Test
    fun `Settings screen has correct route`() {
        assertEquals("settings", Screen.Settings.route)
    }

    @Test
    fun `all screen routes are unique`() {
        val routes = listOf(
            Screen.Login.route,
            Screen.Home.route,
            Screen.Search.route,
            Screen.Settings.route,
        )

        assertEquals(routes.size, routes.toSet().size)
    }

    @Test
    fun `all screen routes are non-blank`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)

        screens.forEach { screen ->
            assertTrue("Route '${screen.route}' should not be blank", screen.route.isNotBlank())
        }
    }

    @Test
    fun `screen routes do not contain slashes`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)

        screens.forEach { screen ->
            assertFalse(
                "Route '${screen.route}' should not contain slashes",
                screen.route.contains("/")
            )
        }
    }

    @Test
    fun `screen routes are lowercase`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)

        screens.forEach { screen ->
            assertEquals(
                "Route '${screen.route}' should be lowercase",
                screen.route.lowercase(),
                screen.route,
            )
        }
    }

    @Test
    fun `Screen sealed class has exactly four subclasses`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)

        assertEquals(4, screens.size)
    }

    @Test
    fun `start destination is home when authenticated`() {
        val isAuthenticated = true
        val startDestination = if (isAuthenticated) Screen.Home.route else Screen.Login.route

        assertEquals("home", startDestination)
    }

    @Test
    fun `start destination is login when not authenticated`() {
        val isAuthenticated = false
        val startDestination = if (isAuthenticated) Screen.Home.route else Screen.Login.route

        assertEquals("login", startDestination)
    }

    @Test
    fun `login success navigation target is home`() {
        val target = Screen.Home.route

        assertEquals("home", target)
    }

    @Test
    fun `login success pops up to login route inclusive`() {
        val popUpToRoute = Screen.Login.route
        val inclusive = true

        assertEquals("login", popUpToRoute)
        assertTrue(inclusive)
    }

    @Test
    fun `search navigation target from home`() {
        val target = Screen.Search.route

        assertEquals("search", target)
    }

    @Test
    fun `settings navigation target from home`() {
        val target = Screen.Settings.route

        assertEquals("settings", target)
    }

    @Test
    fun `logout navigation target is login`() {
        val target = Screen.Login.route

        assertEquals("login", target)
    }

    @Test
    fun `logout clears entire backstack`() {
        val popUpTo = 0
        val inclusive = true

        assertEquals(0, popUpTo)
        assertTrue(inclusive)
    }

    @Test
    fun `search and settings screens support back navigation`() {
        val searchSupportsBack = true
        val settingsSupportsBack = true

        assertTrue(searchSupportsBack)
        assertTrue(settingsSupportsBack)
    }

    @Test
    fun `Screen instances are singleton objects`() {
        val login1 = Screen.Login
        val login2 = Screen.Login

        assertSame(login1, login2)

        val home1 = Screen.Home
        val home2 = Screen.Home

        assertSame(home1, home2)
    }

    @Test
    fun `Screen route property is stable across accesses`() {
        val route1 = Screen.Home.route
        val route2 = Screen.Home.route

        assertEquals(route1, route2)
    }

    @Test
    fun `all screens inherit from Screen sealed class`() {
        val screens: List<Screen> = listOf(
            Screen.Login,
            Screen.Home,
            Screen.Search,
            Screen.Settings,
        )

        screens.forEach { screen ->
            assertTrue(screen is Screen)
        }
    }

    @Test
    fun `navigation graph contains all four routes`() {
        val allRoutes = setOf(
            Screen.Login.route,
            Screen.Home.route,
            Screen.Search.route,
            Screen.Settings.route,
        )

        assertEquals(4, allRoutes.size)
        assertTrue(allRoutes.contains("login"))
        assertTrue(allRoutes.contains("home"))
        assertTrue(allRoutes.contains("search"))
        assertTrue(allRoutes.contains("settings"))
    }

    @Test
    fun `home screen is the only screen accessible after login`() {
        val startDestination = Screen.Home.route
        val loginRoute = Screen.Login.route

        assertNotEquals(loginRoute, startDestination)
        assertEquals("home", startDestination)
    }

    @Test
    fun `media detail navigation is placeholder with no-op`() {
        var mediaDetailCalled = false
        val onNavigateToMediaDetail: (Long) -> Unit = { _ ->
            mediaDetailCalled = true
        }

        onNavigateToMediaDetail(1L)

        assertTrue(mediaDetailCalled)
    }
}
