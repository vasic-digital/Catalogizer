package com.catalogizer.android.ui.navigation

import org.junit.Assert.*
import org.junit.Test

class ScreenRoutesTest2 {

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
    fun `all routes are unique`() {
        val routes = listOf(
            Screen.Login.route,
            Screen.Home.route,
            Screen.Search.route,
            Screen.Settings.route
        )
        assertEquals(routes.size, routes.toSet().size)
    }

    @Test
    fun `all routes are non-empty`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)
        for (screen in screens) {
            assertTrue("Screen ${screen.javaClass.simpleName} route should not be empty",
                screen.route.isNotEmpty())
        }
    }

    @Test
    fun `routes do not contain slashes`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)
        for (screen in screens) {
            assertFalse("Screen ${screen.javaClass.simpleName} route should not contain '/'",
                screen.route.contains("/"))
        }
    }

    @Test
    fun `routes are lowercase`() {
        val screens = listOf(Screen.Login, Screen.Home, Screen.Search, Screen.Settings)
        for (screen in screens) {
            assertEquals(screen.route, screen.route.lowercase())
        }
    }
}
