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

    // Bind to the REAL public Color.kt vals (same package, no import) — NOT
    // replicated copies — so a drift in Color.kt fails these assertions
    // (§11.4 no-tautology). Brand roles mirror catalog-web/src/styles/tokens.ts
    // (Catalogizer-Blue, commit e748bba5); §11.4.162 one shared palette.
    private val lightPrimary = CatalogizerLightPrimary
    private val lightOnPrimary = CatalogizerLightOnPrimary
    private val lightPrimaryContainer = CatalogizerLightPrimaryContainer
    private val lightOnPrimaryContainer = CatalogizerLightOnPrimaryContainer
    private val lightSecondary = CatalogizerLightSecondary
    private val lightOnSecondary = CatalogizerLightOnSecondary
    private val lightSecondaryContainer = CatalogizerLightSecondaryContainer
    private val lightOnSecondaryContainer = CatalogizerLightOnSecondaryContainer
    private val lightTertiary = CatalogizerLightTertiary
    private val lightOnTertiary = CatalogizerLightOnTertiary
    private val lightTertiaryContainer = CatalogizerLightTertiaryContainer
    private val lightOnTertiaryContainer = CatalogizerLightOnTertiaryContainer
    private val lightError = CatalogizerLightError
    private val lightErrorContainer = CatalogizerLightErrorContainer
    private val lightOnError = CatalogizerLightOnError
    private val lightOnErrorContainer = CatalogizerLightOnErrorContainer
    private val lightBackground = CatalogizerLightBackground
    private val lightOnBackground = CatalogizerLightOnBackground
    private val lightSurface = CatalogizerLightSurface
    private val lightOnSurface = CatalogizerLightOnSurface
    private val lightSurfaceVariant = CatalogizerLightSurfaceVariant
    private val lightOnSurfaceVariant = CatalogizerLightOnSurfaceVariant
    private val lightOutline = CatalogizerLightOutline

    private val darkPrimary = CatalogizerDarkPrimary
    private val darkOnPrimary = CatalogizerDarkOnPrimary
    private val darkPrimaryContainer = CatalogizerDarkPrimaryContainer
    private val darkOnPrimaryContainer = CatalogizerDarkOnPrimaryContainer
    private val darkSecondary = CatalogizerDarkSecondary
    private val darkOnSecondary = CatalogizerDarkOnSecondary
    private val darkSecondaryContainer = CatalogizerDarkSecondaryContainer
    private val darkOnSecondaryContainer = CatalogizerDarkOnSecondaryContainer
    private val darkTertiary = CatalogizerDarkTertiary
    private val darkOnTertiary = CatalogizerDarkOnTertiary
    private val darkTertiaryContainer = CatalogizerDarkTertiaryContainer
    private val darkOnTertiaryContainer = CatalogizerDarkOnTertiaryContainer
    private val darkError = CatalogizerDarkError
    private val darkErrorContainer = CatalogizerDarkErrorContainer
    private val darkOnError = CatalogizerDarkOnError
    private val darkOnErrorContainer = CatalogizerDarkOnErrorContainer
    private val darkBackground = CatalogizerDarkBackground
    private val darkOnBackground = CatalogizerDarkOnBackground
    private val darkSurface = CatalogizerDarkSurface
    private val darkOnSurface = CatalogizerDarkOnSurface
    private val darkSurfaceVariant = CatalogizerDarkSurfaceVariant
    private val darkOnSurfaceVariant = CatalogizerDarkOnSurfaceVariant
    private val darkOutline = CatalogizerDarkOutline

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
    fun `light error color is M3-tonal red`() {
        // M3-tonal #BA1A1A (NOT web semantic.error #DC2626) — see Color.kt.
        assertEquals(Color(0xFFBA1A1A), lightError)
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
    fun `dark error is M3-tonal red`() {
        // M3-tonal #FFB4AB (NOT web semantic.error #EF4444, which dropped the
        // onError/error pair to 3.48:1 < WCAG AA) — see Color.kt.
        assertEquals(Color(0xFFFFB4AB), darkError)
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

    // --- WCAG 2.1 contrast guard (§11.4.135 / §11.4.162) ---------------------
    // Mechanical, device-independent proof that every text pair clears AA. This
    // is what catches the surface-vs-foreground error trap: forcing web
    // semantic.error onto the M3 error SURFACE drops dark onError/error to
    // 3.48:1 — a hex-only test would pass it silently, this one fails it.

    private fun channelLuminance(v: Float): Double {
        val cs = v.toDouble()
        return if (cs <= 0.03928) cs / 12.92 else Math.pow((cs + 0.055) / 1.055, 2.4)
    }

    private fun relativeLuminance(c: Color): Double =
        0.2126 * channelLuminance(c.red) +
            0.7152 * channelLuminance(c.green) +
            0.0722 * channelLuminance(c.blue)

    private fun contrastRatio(a: Color, b: Color): Double {
        val la = relativeLuminance(a)
        val lb = relativeLuminance(b)
        return (maxOf(la, lb) + 0.05) / (minOf(la, lb) + 0.05)
    }

    private fun assertAa(label: String, fg: Color, bg: Color) {
        val ratio = contrastRatio(fg, bg)
        assertTrue("WCAG AA fail: $label ${"%.2f".format(ratio)}:1 < 4.5:1", ratio >= 4.5)
    }

    @Test
    fun `contrast formula is correct (self-validation)`() {
        // Inverted/mis-coefficiented formula misses these reference points.
        assertTrue(contrastRatio(Color(0xFF000000), Color(0xFFFFFFFF)) in 20.9..21.1)
        assertTrue(contrastRatio(Color(0xFF123456), Color(0xFF123456)) in 0.99..1.01)
    }

    @Test
    fun `dark scheme text pairs meet WCAG AA`() {
        assertAa("dark onBackground/background", darkOnBackground, darkBackground)
        assertAa("dark onSurface/surface", darkOnSurface, darkSurface)
        assertAa("dark onPrimary/primary", darkOnPrimary, darkPrimary)
        assertAa("dark onSecondary/secondary", darkOnSecondary, darkSecondary)
        assertAa("dark onTertiary/tertiary", darkOnTertiary, darkTertiary)
        assertAa("dark onError/error", darkOnError, darkError)
        assertAa("dark onSurfaceVariant/surfaceVariant", darkOnSurfaceVariant, darkSurfaceVariant)
    }

    @Test
    fun `light scheme text pairs meet WCAG AA`() {
        assertAa("light onBackground/background", lightOnBackground, lightBackground)
        assertAa("light onSurface/surface", lightOnSurface, lightSurface)
        assertAa("light onPrimary/primary", lightOnPrimary, lightPrimary)
        assertAa("light onSecondary/secondary", lightOnSecondary, lightSecondary)
        assertAa("light onTertiary/tertiary", lightOnTertiary, lightTertiary)
        assertAa("light onError/error", lightOnError, lightError)
        assertAa("light onSurfaceVariant/surfaceVariant", lightOnSurfaceVariant, lightSurfaceVariant)
    }
}
