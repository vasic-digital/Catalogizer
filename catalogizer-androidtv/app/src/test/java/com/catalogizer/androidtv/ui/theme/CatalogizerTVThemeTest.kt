package com.catalogizer.androidtv.ui.theme

import androidx.compose.ui.graphics.Color
import androidx.tv.material3.ExperimentalTvMaterial3Api
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

/**
 * §11.4.162 OpenDesign UI design-system guard for the Android-TV client +
 * §11.4.135 standing regression guard for its WCAG contrast.
 *
 * This test reads the REAL [TVDarkColorScheme] / [TVLightColorScheme] from
 * Theme.kt (they are `internal`, same module) — NOT a replicated copy — so a
 * drift in any role away from the shared OpenDesign Catalogizer-Blue palette,
 * or a contrast regression, FAILs here.
 *
 * Two independent oracles:
 *  (1) ROLE-HEX: every brand / semantic / neutral role resolves to the exact
 *      token from catalog-web/src/styles/tokens.ts (commit e748bba5) so web +
 *      phone + TV share ONE palette. Values are NOT invented (§11.4.6).
 *  (2) WCAG CONTRAST: the relative-luminance contrast ratio of every text pair
 *      is computed (WCAG 2.1 formula) and asserted >= 4.5:1 (AA normal text) in
 *      BOTH schemes. This is a MECHANICAL, device-independent replacement for
 *      the prior manual HELIX-175 audit — it re-proves contrast on every build,
 *      which matters because MIBOX4 (the only TV target) can be offline.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@RunWith(RobolectricTestRunner::class)
class CatalogizerTVThemeTest {

    // ---- WCAG 2.1 relative luminance + contrast ratio -----------------------
    // Compose Color channel accessors return sRGB-encoded Floats in [0,1].
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
        val hi = maxOf(la, lb)
        val lo = minOf(la, lb)
        return (hi + 0.05) / (lo + 0.05)
    }

    private fun assertAa(label: String, fg: Color, bg: Color) {
        val ratio = contrastRatio(fg, bg)
        assertTrue(
            "WCAG AA fail: $label contrast ${"%.2f".format(ratio)}:1 < 4.5:1",
            ratio >= 4.5,
        )
    }

    // ---- (1) ROLE-HEX: shared OpenDesign Catalogizer-Blue tokens ------------

    @Test
    fun darkSchemeBrandAndNeutralRolesMatchOpenDesignTokens() {
        val s = TVDarkColorScheme
        assertEquals(Color(0xFF9ECAFF), s.primary)
        assertEquals(Color(0xFF003258), s.onPrimary)
        assertEquals(Color(0xFF00497D), s.primaryContainer)
        assertEquals(Color(0xFFD1E4FF), s.onPrimaryContainer)
        assertEquals(Color(0xFFBBC7DB), s.secondary)
        assertEquals(Color(0xFF253140), s.onSecondary)
        assertEquals(Color(0xFF3B4858), s.secondaryContainer)
        assertEquals(Color(0xFFD7E3F7), s.onSecondaryContainer)
        assertEquals(Color(0xFFD6BEE4), s.tertiary)
        assertEquals(Color(0xFF3B2948), s.onTertiary)
        assertEquals(Color(0xFF523F5F), s.tertiaryContainer)
        assertEquals(Color(0xFFF2DAFF), s.onTertiaryContainer)
        // error stays M3-tonal (#FFB4AB) — web semantic.error #EF4444 is a
        // FOREGROUND token; on the M3 error SURFACE it would drop onError
        // contrast to 3.48:1 (< AA). See Theme.kt rationale.
        assertEquals(Color(0xFFFFB4AB), s.error)
        assertEquals(Color(0xFF690005), s.onError)
        assertEquals(Color(0xFF93000A), s.errorContainer)
        assertEquals(Color(0xFFFFDAD6), s.onErrorContainer)
        // neutral.background/surface — aligned to web tokens.
        assertEquals(Color(0xFF020817), s.background)
        assertEquals(Color(0xFFF8FAFC), s.onBackground)
        assertEquals(Color(0xFF0F172A), s.surface)
        assertEquals(Color(0xFFF8FAFC), s.onSurface)
        assertEquals(Color(0xFF43474E), s.surfaceVariant)
        assertEquals(Color(0xFFC3C7CF), s.onSurfaceVariant)
    }

    @Test
    fun lightSchemeBrandAndNeutralRolesMatchOpenDesignTokens() {
        val s = TVLightColorScheme
        assertEquals(Color(0xFF1565C0), s.primary)
        assertEquals(Color(0xFFFFFFFF), s.onPrimary)
        assertEquals(Color(0xFFD1E4FF), s.primaryContainer)
        assertEquals(Color(0xFF001D36), s.onPrimaryContainer)
        assertEquals(Color(0xFF535F70), s.secondary)
        assertEquals(Color(0xFFFFFFFF), s.onSecondary)
        assertEquals(Color(0xFFD7E3F7), s.secondaryContainer)
        assertEquals(Color(0xFF101C2B), s.onSecondaryContainer)
        assertEquals(Color(0xFF6B5778), s.tertiary)
        assertEquals(Color(0xFFFFFFFF), s.onTertiary)
        assertEquals(Color(0xFFF2DAFF), s.tertiaryContainer)
        assertEquals(Color(0xFF251431), s.onTertiaryContainer)
        // error stays M3-tonal (#BA1A1A) for the same surface-vs-foreground
        // reason as the dark scheme (see Theme.kt rationale).
        assertEquals(Color(0xFFBA1A1A), s.error)
        assertEquals(Color(0xFFFFFFFF), s.onError)
        assertEquals(Color(0xFFFFDAD6), s.errorContainer)
        assertEquals(Color(0xFF410002), s.onErrorContainer)
        // neutral.background/surface — aligned to web tokens.
        assertEquals(Color(0xFFFFFFFF), s.background)
        assertEquals(Color(0xFF020817), s.onBackground)
        assertEquals(Color(0xFFFFFFFF), s.surface)
        assertEquals(Color(0xFF020817), s.onSurface)
        assertEquals(Color(0xFFF8FAFC), s.surfaceVariant)
        assertEquals(Color(0xFF43474E), s.onSurfaceVariant)
    }

    // ---- (2) WCAG CONTRAST: every text pair >= 4.5:1 AA in both schemes -----

    @Test
    fun darkSchemeTextPairsMeetWcagAa() {
        val s = TVDarkColorScheme
        assertAa("dark onBackground/background", s.onBackground, s.background)
        assertAa("dark onSurface/surface", s.onSurface, s.surface)
        assertAa("dark onPrimary/primary", s.onPrimary, s.primary)
        assertAa("dark onSecondary/secondary", s.onSecondary, s.secondary)
        assertAa("dark onTertiary/tertiary", s.onTertiary, s.tertiary)
        assertAa("dark onError/error", s.onError, s.error)
        assertAa("dark onSurfaceVariant/surfaceVariant", s.onSurfaceVariant, s.surfaceVariant)
    }

    @Test
    fun lightSchemeTextPairsMeetWcagAa() {
        val s = TVLightColorScheme
        assertAa("light onBackground/background", s.onBackground, s.background)
        assertAa("light onSurface/surface", s.onSurface, s.surface)
        assertAa("light onPrimary/primary", s.onPrimary, s.primary)
        assertAa("light onSecondary/secondary", s.onSecondary, s.secondary)
        assertAa("light onTertiary/tertiary", s.onTertiary, s.tertiary)
        assertAa("light onError/error", s.onError, s.error)
        assertAa("light onSurfaceVariant/surfaceVariant", s.onSurfaceVariant, s.surfaceVariant)
    }

    @Test
    fun sanityContrastFormulaIsCorrect() {
        // Black-on-white is the WCAG reference maximum, 21:1.
        val ratio = contrastRatio(Color(0xFF000000), Color(0xFFFFFFFF))
        assertTrue("black/white should be ~21:1, got $ratio", ratio in 20.9..21.1)
        // Identical colors are 1:1 — guards against an inverted formula.
        val same = contrastRatio(Color(0xFF123456), Color(0xFF123456))
        assertTrue("identical colors should be 1:1, got $same", same in 0.99..1.01)
    }
}
