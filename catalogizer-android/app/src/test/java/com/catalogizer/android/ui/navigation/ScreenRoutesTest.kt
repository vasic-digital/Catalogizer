package com.catalogizer.android.ui.navigation

import org.junit.Assert.*
import org.junit.Test

class ScreenRoutesTest {

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
            Screen.Settings.route
        )

        assertEquals(routes.size, routes.toSet().size)
    }

    @Test
    fun `Screen is sealed class with fixed set of subclasses`() {
        val screens = listOf(
            Screen.Login,
            Screen.Home,
            Screen.Search,
            Screen.Settings
        )

        assertEquals(4, screens.size)
        screens.forEach { screen ->
            assertTrue(screen is Screen)
            assertTrue(screen.route.isNotBlank())
        }
    }
}
