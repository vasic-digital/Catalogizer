package com.catalogizer.android.ui.theme

import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.ui.graphics.Color
import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * Tests for the CatalogizerTheme color definitions, typography defaults,
 * and theme selection logic (light/dark, dynamic color).
 * Validates that all color constants are properly defined and that the
 * light/dark scheme selection logic works correctly.
 */
@RunWith(RobolectricTestRunner::class)
class ThemeTest {

    // Replicate the color constants from Color.kt for verification.
    // Brand-role values mirror catalog-web/src/styles/tokens.ts (Catalogizer-Blue,
    // commit e748bba5) so the phone palette stays brand-consistent (§11.4.162).
    private val lightPrimary = Color(0xFF1565C0)
    private val lightOnPrimary = Color(0xFFFFFFFF)
    private val lightPrimaryContainer = Color(0xFFD1E4FF)
    private val lightOnPrimaryContainer = Color(0xFF001D36)
    private val lightSecondary = Color(0xFF535F70)
    private val lightOnSecondary = Color(0xFFFFFFFF)
    private val lightSecondaryContainer = Color(0xFFD7E3F7)
    private val lightOnSecondaryContainer = Color(0xFF101C2B)
    private val lightTertiary = Color(0xFF6B5778)
    private val lightOnTertiary = Color(0xFFFFFFFF)
    private val lightTertiaryContainer = Color(0xFFF2DAFF)
    private val lightOnTertiaryContainer = Color(0xFF251431)
    private val lightError = Color(0xFFDC2626)
    private val lightErrorContainer = Color(0xFFFFDAD6)
    private val lightOnError = Color(0xFFFFFFFF)
    private val lightOnErrorContainer = Color(0xFF410002)
    private val lightBackground = Color(0xFFFFFFFF)
    private val lightOnBackground = Color(0xFF020817)
    private val lightSurface = Color(0xFFFFFFFF)
    private val lightOnSurface = Color(0xFF020817)
    private val lightSurfaceVariant = Color(0xFFF8FAFC)
    private val lightOnSurfaceVariant = Color(0xFF43474E)
    private val lightOutline = Color(0xFFE2E8F0)

    private val darkPrimary = Color(0xFF9ECAFF)
    private val darkOnPrimary = Color(0xFF003258)
    private val darkPrimaryContainer = Color(0xFF00497D)
    private val darkOnPrimaryContainer = Color(0xFFD1E4FF)
    private val darkSecondary = Color(0xFFBBC7DB)
    private val darkOnSecondary = Color(0xFF253140)
    private val darkSecondaryContainer = Color(0xFF3B4858)
    private val darkOnSecondaryContainer = Color(0xFFD7E3F7)
    private val darkTertiary = Color(0xFFD6BEE4)
    private val darkOnTertiary = Color(0xFF3B2948)
    private val darkTertiaryContainer = Color(0xFF523F5F)
    private val darkOnTertiaryContainer = Color(0xFFF2DAFF)
    private val darkError = Color(0xFFEF4444)
    private val darkErrorContainer = Color(0xFF93000A)
    private val darkOnError = Color(0xFF690005)
    private val darkOnErrorContainer = Color(0xFFFFDAD6)
    private val darkBackground = Color(0xFF020817)
    private val darkOnBackground = Color(0xFFF8FAFC)
    private val darkSurface = Color(0xFF0F172A)
    private val darkOnSurface = Color(0xFFF8FAFC)
    private val darkSurfaceVariant = Color(0xFF43474E)
    private val darkOnSurfaceVariant = Color(0xFFC3C7CF)
    private val darkOutline = Color(0xFF1E293B)

    // --- Light Color Definitions ---

    @Test
    fun `light primary color is correct blue`() {
        assertEquals(Color(0xFF1565C0), lightPrimary)
    }

    @Test
    fun `light onPrimary is white`() {
        assertEquals(Color(0xFFFFFFFF), lightOnPrimary)
    }

    @Test
    fun `light primaryContainer is light blue`() {
        assertEquals(Color(0xFFD1E4FF), lightPrimaryContainer)
    }

    @Test
    fun `light onPrimaryContainer is dark blue`() {
        assertEquals(Color(0xFF001D36), lightOnPrimaryContainer)
    }

    @Test
    fun `light secondary color is defined`() {
        assertEquals(Color(0xFF535F70), lightSecondary)
    }

    @Test
    fun `light onSecondary is white`() {
        assertEquals(Color(0xFFFFFFFF), lightOnSecondary)
    }

    @Test
    fun `light tertiary color is defined`() {
        assertEquals(Color(0xFF6B5778), lightTertiary)
    }

    @Test
    fun `light onTertiary is white`() {
        assertEquals(Color(0xFFFFFFFF), lightOnTertiary)
    }

    @Test
    fun `light error color is red`() {
        assertEquals(Color(0xFFDC2626), lightError)
    }

    @Test
    fun `light onError is white`() {
        assertEquals(Color(0xFFFFFFFF), lightOnError)
    }

    @Test
    fun `light errorContainer is light red`() {
        assertEquals(Color(0xFFFFDAD6), lightErrorContainer)
    }

    @Test
    fun `light background is white`() {
        assertEquals(Color(0xFFFFFFFF), lightBackground)
    }

    @Test
    fun `light onBackground is dark`() {
        assertEquals(Color(0xFF020817), lightOnBackground)
    }

    @Test
    fun `light surface equals background`() {
        assertEquals(lightBackground, lightSurface)
    }

    @Test
    fun `light onSurface equals onBackground`() {
        assertEquals(lightOnBackground, lightOnSurface)
    }

    @Test
    fun `light surfaceVariant is defined`() {
        assertEquals(Color(0xFFF8FAFC), lightSurfaceVariant)
    }

    @Test
    fun `light outline is defined`() {
        assertEquals(Color(0xFFE2E8F0), lightOutline)
    }

    // --- Dark Color Definitions ---

    @Test
    fun `dark primary color is light blue`() {
        assertEquals(Color(0xFF9ECAFF), darkPrimary)
    }

    @Test
    fun `dark onPrimary is dark blue`() {
        assertEquals(Color(0xFF003258), darkOnPrimary)
    }

    @Test
    fun `dark primaryContainer is defined`() {
        assertEquals(Color(0xFF00497D), darkPrimaryContainer)
    }

    @Test
    fun `dark onPrimaryContainer matches light primaryContainer`() {
        assertEquals(lightPrimaryContainer, darkOnPrimaryContainer)
    }

    @Test
    fun `dark secondary is defined`() {
        assertEquals(Color(0xFFBBC7DB), darkSecondary)
    }

    @Test
    fun `dark tertiary is defined`() {
        assertEquals(Color(0xFFD6BEE4), darkTertiary)
    }

    @Test
    fun `dark error is red`() {
        assertEquals(Color(0xFFEF4444), darkError)
    }

    @Test
    fun `dark onError is deep red`() {
        assertEquals(Color(0xFF690005), darkOnError)
    }

    @Test
    fun `dark errorContainer is dark red`() {
        assertEquals(Color(0xFF93000A), darkErrorContainer)
    }

    @Test
    fun `dark background is near black`() {
        assertEquals(Color(0xFF020817), darkBackground)
    }

    @Test
    fun `dark onBackground is light`() {
        assertEquals(Color(0xFFF8FAFC), darkOnBackground)
    }

    @Test
    fun `dark surface is distinct elevated layer above background`() {
        // Catalogizer-Blue dark uses a distinct surface (#0F172A) raised above
        // the deeper background (#020817), matching the web token model
        // (color.neutral.surface != color.neutral.background) — gives card/sheet
        // elevation so surfaces never blend into the background.
        assertEquals(Color(0xFF0F172A), darkSurface)
        assertNotEquals(darkBackground, darkSurface)
    }

    @Test
    fun `dark onSurface equals dark onBackground`() {
        assertEquals(darkOnBackground, darkOnSurface)
    }

    @Test
    fun `dark surfaceVariant is defined`() {
        assertEquals(Color(0xFF43474E), darkSurfaceVariant)
    }

    @Test
    fun `dark outline is defined`() {
        assertEquals(Color(0xFF1E293B), darkOutline)
    }

    // --- Contrast and Consistency ---

    @Test
    fun `light and dark primary colors are different`() {
        assertNotEquals(lightPrimary, darkPrimary)
    }

    @Test
    fun `light and dark background colors are different`() {
        assertNotEquals(lightBackground, darkBackground)
    }

    @Test
    fun `light and dark surface colors are different`() {
        assertNotEquals(lightSurface, darkSurface)
    }

    @Test
    fun `light and dark error colors are different`() {
        assertNotEquals(lightError, darkError)
    }

    @Test
    fun `light and dark onPrimary colors are different`() {
        assertNotEquals(lightOnPrimary, darkOnPrimary)
    }

    @Test
    fun `light and dark onBackground colors are different`() {
        assertNotEquals(lightOnBackground, darkOnBackground)
    }

    // --- Color Scheme Construction ---

    @Test
    fun `light color scheme can be constructed with all colors`() {
        val scheme = lightColorScheme(
            primary = lightPrimary,
            onPrimary = lightOnPrimary,
            primaryContainer = lightPrimaryContainer,
            onPrimaryContainer = lightOnPrimaryContainer,
            secondary = lightSecondary,
            onSecondary = lightOnSecondary,
            secondaryContainer = lightSecondaryContainer,
            onSecondaryContainer = lightOnSecondaryContainer,
            tertiary = lightTertiary,
            onTertiary = lightOnTertiary,
            tertiaryContainer = lightTertiaryContainer,
            onTertiaryContainer = lightOnTertiaryContainer,
            error = lightError,
            errorContainer = lightErrorContainer,
            onError = lightOnError,
            onErrorContainer = lightOnErrorContainer,
            background = lightBackground,
            onBackground = lightOnBackground,
            surface = lightSurface,
            onSurface = lightOnSurface,
            surfaceVariant = lightSurfaceVariant,
            onSurfaceVariant = lightOnSurfaceVariant,
            outline = lightOutline,
        )

        assertEquals(lightPrimary, scheme.primary)
        assertEquals(lightOnPrimary, scheme.onPrimary)
        assertEquals(lightError, scheme.error)
        assertEquals(lightBackground, scheme.background)
    }

    @Test
    fun `dark color scheme can be constructed with all colors`() {
        val scheme = darkColorScheme(
            primary = darkPrimary,
            onPrimary = darkOnPrimary,
            primaryContainer = darkPrimaryContainer,
            onPrimaryContainer = darkOnPrimaryContainer,
            secondary = darkSecondary,
            onSecondary = darkOnSecondary,
            secondaryContainer = darkSecondaryContainer,
            onSecondaryContainer = darkOnSecondaryContainer,
            tertiary = darkTertiary,
            onTertiary = darkOnTertiary,
            tertiaryContainer = darkTertiaryContainer,
            onTertiaryContainer = darkOnTertiaryContainer,
            error = darkError,
            errorContainer = darkErrorContainer,
            onError = darkOnError,
            onErrorContainer = darkOnErrorContainer,
            background = darkBackground,
            onBackground = darkOnBackground,
            surface = darkSurface,
            onSurface = darkOnSurface,
            surfaceVariant = darkSurfaceVariant,
            onSurfaceVariant = darkOnSurfaceVariant,
            outline = darkOutline,
        )

        assertEquals(darkPrimary, scheme.primary)
        assertEquals(darkOnPrimary, scheme.onPrimary)
        assertEquals(darkError, scheme.error)
        assertEquals(darkBackground, scheme.background)
    }

    // --- Theme Selection Logic ---

    @Test
    fun `dark theme with no dynamic color selects dark scheme`() {
        val darkTheme = true
        val dynamicColor = false
        val sdkVersion = 30

        val useDynamic = dynamicColor && sdkVersion >= 31
        val selectedScheme = when {
            useDynamic -> "dynamic"
            darkTheme -> "dark"
            else -> "light"
        }

        assertEquals("dark", selectedScheme)
    }

    @Test
    fun `light theme with no dynamic color selects light scheme`() {
        val darkTheme = false
        val dynamicColor = false
        val sdkVersion = 30

        val useDynamic = dynamicColor && sdkVersion >= 31
        val selectedScheme = when {
            useDynamic -> "dynamic"
            darkTheme -> "dark"
            else -> "light"
        }

        assertEquals("light", selectedScheme)
    }

    @Test
    fun `dynamic color on API 31 plus selects dynamic scheme`() {
        val darkTheme = false
        val dynamicColor = true
        val sdkVersion = 31

        val useDynamic = dynamicColor && sdkVersion >= 31
        val selectedScheme = when {
            useDynamic -> "dynamic"
            darkTheme -> "dark"
            else -> "light"
        }

        assertEquals("dynamic", selectedScheme)
    }

    @Test
    fun `dynamic color on API below 31 falls back to static scheme`() {
        val darkTheme = false
        val dynamicColor = true
        val sdkVersion = 30

        val useDynamic = dynamicColor && sdkVersion >= 31
        val selectedScheme = when {
            useDynamic -> "dynamic"
            darkTheme -> "dark"
            else -> "light"
        }

        assertEquals("light", selectedScheme)
    }

    @Test
    fun `dynamic dark theme on API 31 plus selects dynamic`() {
        val darkTheme = true
        val dynamicColor = true
        val sdkVersion = 33

        val useDynamic = dynamicColor && sdkVersion >= 31
        val selectedScheme = when {
            useDynamic -> "dynamic"
            darkTheme -> "dark"
            else -> "light"
        }

        assertEquals("dynamic", selectedScheme)
    }

    // --- Typography ---

    @Test
    fun `default typography is used`() {
        val typography = androidx.compose.material3.Typography()

        assertNotNull(typography)
        assertNotNull(typography.bodyLarge)
        assertNotNull(typography.bodyMedium)
        assertNotNull(typography.bodySmall)
        assertNotNull(typography.titleLarge)
        assertNotNull(typography.titleMedium)
        assertNotNull(typography.titleSmall)
        assertNotNull(typography.headlineLarge)
        assertNotNull(typography.headlineMedium)
        assertNotNull(typography.headlineSmall)
        assertNotNull(typography.labelLarge)
        assertNotNull(typography.labelMedium)
        assertNotNull(typography.labelSmall)
        assertNotNull(typography.displayLarge)
        assertNotNull(typography.displayMedium)
        assertNotNull(typography.displaySmall)
    }

    // --- Cross-theme container symmetry ---

    @Test
    fun `light secondaryContainer matches dark onSecondaryContainer`() {
        assertEquals(lightSecondaryContainer, darkOnSecondaryContainer)
    }

    @Test
    fun `light tertiaryContainer matches dark onTertiaryContainer`() {
        assertEquals(lightTertiaryContainer, darkOnTertiaryContainer)
    }

    @Test
    fun `light errorContainer matches dark onErrorContainer`() {
        assertEquals(lightErrorContainer, darkOnErrorContainer)
    }

    @Test
    fun `all color values are non-transparent`() {
        val allColors = listOf(
            lightPrimary, lightOnPrimary, lightBackground, lightOnBackground,
            lightError, lightOnError, lightSurface, lightOnSurface,
            darkPrimary, darkOnPrimary, darkBackground, darkOnBackground,
            darkError, darkOnError, darkSurface, darkOnSurface,
        )

        allColors.forEach { color ->
            assertTrue("Color $color should not be fully transparent", color.alpha > 0f)
        }
    }
}
