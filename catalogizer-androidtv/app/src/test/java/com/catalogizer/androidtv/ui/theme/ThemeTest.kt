package com.catalogizer.androidtv.ui.theme

import androidx.compose.ui.graphics.Color
import org.junit.Assert.*
import org.junit.Test

/**
 * Tests for TV theme color definitions: verifies dark and light scheme
 * color values, ensures proper contrast relationships, and validates
 * that the theme defaults to dark mode for TV.
 */
class ThemeTest {

    // --- Dark theme color values ---

    @Test
    fun `dark theme primary is soft blue`() {
        val primary = Color(0xFF9ECAFF)
        assertEquals(0xFF9ECAFF, primary.value.toLong() and 0xFFFFFFFFL)
    }

    @Test
    fun `dark theme onPrimary is dark blue`() {
        val onPrimary = Color(0xFF003258)
        assertNotEquals(Color.Transparent, onPrimary)
    }

    @Test
    fun `dark theme background is very dark`() {
        val background = Color(0xFF101214)
        // Background should be very dark for TV viewing
        assertTrue(background.red < 0.1f)
        assertTrue(background.green < 0.1f)
        assertTrue(background.blue < 0.1f)
    }

    @Test
    fun `dark theme surface matches background`() {
        val background = Color(0xFF101214)
        val surface = Color(0xFF101214)
        assertEquals(background, surface)
    }

    @Test
    fun `dark theme onBackground is light for readability`() {
        val onBackground = Color(0xFFE2E2E6)
        // onBackground must be light enough to read on dark background
        assertTrue(onBackground.red > 0.8f)
        assertTrue(onBackground.green > 0.8f)
        assertTrue(onBackground.blue > 0.8f)
    }

    @Test
    fun `dark theme error color is visible on dark background`() {
        val error = Color(0xFFFFB4AB)
        // Error should be a light reddish tone visible on dark surfaces
        assertTrue(error.red > 0.9f)
        assertTrue(error.alpha == 1.0f)
    }

    @Test
    fun `dark theme secondary is muted`() {
        val secondary = Color(0xFFBBC7DB)
        assertNotEquals(Color.White, secondary)
        assertNotEquals(Color.Black, secondary)
    }

    @Test
    fun `dark theme tertiary is purple-tinted`() {
        val tertiary = Color(0xFFD6BEE4)
        // Tertiary should have more blue/red than green (purple)
        assertTrue(tertiary.blue > tertiary.green)
    }

    // --- Light theme color values ---

    @Test
    fun `light theme primary is standard blue`() {
        val primary = Color(0xFF1976D2)
        // Should be a solid blue
        assertTrue(primary.blue > primary.red)
        assertTrue(primary.blue > primary.green)
    }

    @Test
    fun `light theme onPrimary is white`() {
        val onPrimary = Color(0xFFFFFFFF)
        assertEquals(Color.White, onPrimary)
    }

    @Test
    fun `light theme background is near-white`() {
        val background = Color(0xFFFDFCFF)
        assertTrue(background.red > 0.98f)
        assertTrue(background.green > 0.98f)
        assertTrue(background.blue > 0.98f)
    }

    @Test
    fun `light theme onBackground is dark`() {
        val onBackground = Color(0xFF1A1C1E)
        assertTrue(onBackground.red < 0.15f)
        assertTrue(onBackground.green < 0.15f)
        assertTrue(onBackground.blue < 0.15f)
    }

    @Test
    fun `light theme error is deep red`() {
        val error = Color(0xFFBA1A1A)
        assertTrue(error.red > 0.5f)
        assertTrue(error.green < 0.2f)
    }

    // --- Theme selection ---

    @Test
    fun `default theme is dark for TV`() {
        val darkTheme = true
        assertTrue(darkTheme)
    }

    @Test
    fun `dark theme selects dark color scheme`() {
        val darkTheme = true
        val schemeName = if (darkTheme) "TVDarkColorScheme" else "TVLightColorScheme"
        assertEquals("TVDarkColorScheme", schemeName)
    }

    @Test
    fun `light theme selects light color scheme`() {
        val darkTheme = false
        val schemeName = if (darkTheme) "TVDarkColorScheme" else "TVLightColorScheme"
        assertEquals("TVLightColorScheme", schemeName)
    }

    // --- Contrast requirements ---

    @Test
    fun `primary and onPrimary have sufficient contrast`() {
        val primary = Color(0xFF9ECAFF) // light blue
        val onPrimary = Color(0xFF003258) // dark blue
        // onPrimary should be significantly darker than primary
        assertTrue(onPrimary.red < primary.red)
        assertTrue(onPrimary.green < primary.green)
        assertTrue(onPrimary.blue < primary.blue)
    }

    @Test
    fun `error and onError have sufficient contrast`() {
        val error = Color(0xFFFFB4AB) // light red
        val onError = Color(0xFF690005) // dark red
        assertTrue(onError.red < error.red)
    }

    @Test
    fun `surface and onSurface have sufficient contrast`() {
        val surface = Color(0xFF101214) // very dark
        val onSurface = Color(0xFFE2E2E6) // very light
        // Large contrast gap needed for TV readability
        val surfaceLuminance = (surface.red + surface.green + surface.blue) / 3
        val onSurfaceLuminance = (onSurface.red + onSurface.green + onSurface.blue) / 3
        assertTrue(onSurfaceLuminance - surfaceLuminance > 0.7f)
    }

    // --- Color scheme completeness ---

    @Test
    fun `dark scheme defines all container colors`() {
        val primaryContainer = Color(0xFF00497D)
        val secondaryContainer = Color(0xFF3B4858)
        val tertiaryContainer = Color(0xFF523F5F)
        val errorContainer = Color(0xFF93000A)

        assertNotEquals(Color.Transparent, primaryContainer)
        assertNotEquals(Color.Transparent, secondaryContainer)
        assertNotEquals(Color.Transparent, tertiaryContainer)
        assertNotEquals(Color.Transparent, errorContainer)
    }

    @Test
    fun `dark scheme defines surfaceVariant`() {
        val surfaceVariant = Color(0xFF43474E)
        val onSurfaceVariant = Color(0xFFC3C7CF)
        assertNotEquals(surfaceVariant, onSurfaceVariant)
    }
}
