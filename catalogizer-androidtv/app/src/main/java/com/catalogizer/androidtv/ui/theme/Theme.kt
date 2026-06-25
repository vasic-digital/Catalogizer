package com.catalogizer.androidtv.ui.theme

import androidx.compose.runtime.Composable
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.MaterialTheme
import androidx.tv.material3.darkColorScheme
import androidx.tv.material3.lightColorScheme
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

// TV-optimized color schemes.
// `internal` (not private) so the same-module CatalogizerTVThemeTest reads the
// REAL scheme values (binds the §11.4.162 / WCAG guard to this file, not a copy).
@OptIn(ExperimentalTvMaterial3Api::class)
internal val TVDarkColorScheme = darkColorScheme(
    primary = Color(0xFF9ECAFF),
    onPrimary = Color(0xFF003258),
    primaryContainer = Color(0xFF00497D),
    onPrimaryContainer = Color(0xFFD1E4FF),
    secondary = Color(0xFFBBC7DB),
    onSecondary = Color(0xFF253140),
    secondaryContainer = Color(0xFF3B4858),
    onSecondaryContainer = Color(0xFFD7E3F7),
    tertiary = Color(0xFFD6BEE4),
    onTertiary = Color(0xFF3B2948),
    tertiaryContainer = Color(0xFF523F5F),
    onTertiaryContainer = Color(0xFFF2DAFF),
    // neutral.background/surface aligned to the shared OpenDesign
    // Catalogizer-Blue tokens (catalog-web/src/styles/tokens.ts) so web +
    // phone + TV share ONE palette (§11.4.162). Darker neutrals only RAISE the
    // light-text contrast — proven mechanically in CatalogizerTVThemeTest
    // (onBackground/background ≈ 17:1, AAA).
    // error STAYS M3-tonal (#FFB4AB), NOT web semantic.error #EF4444: the web
    // token is a FOREGROUND token (error text on a neutral bg), and forcing it
    // onto the M3 error SURFACE role drops onError(#690005)/error contrast to
    // 3.48:1 — below WCAG AA. The mechanical guard caught this regression.
    error = Color(0xFFFFB4AB),
    errorContainer = Color(0xFF93000A),
    onError = Color(0xFF690005),
    onErrorContainer = Color(0xFFFFDAD6),
    background = Color(0xFF020817),
    onBackground = Color(0xFFF8FAFC),
    surface = Color(0xFF0F172A),
    onSurface = Color(0xFFF8FAFC),
    surfaceVariant = Color(0xFF43474E),
    onSurfaceVariant = Color(0xFFC3C7CF)
)

// HELIX-175 audit (2026-04-22) — contrast ratios verified against
// WCAG 2.1 AA (minimum 4.5 : 1 normal text, 3 : 1 large text):
//   dark.onBackground on background  ≈ 13 : 1  (AAA)
//   dark.onSurface on surface         ≈ 13 : 1  (AAA)
//   dark.onPrimary on primary         ≈  6 : 1  (AAA)
//   light.onBackground on background  ≈ 14 : 1  (AAA)
//   light.onPrimary on primary        ≈  5.3 : 1 (AA) — bumped the
//       primary from #1976D2 to #1565C0 (darker blue) to raise
//       the white-on-blue ratio closer to AAA. Previous value
//       landed at 4.5 : 1 which is exactly the AA floor, so any
//       antialiased strokes at the edge of a button risked
//       perceptual drop-below. #1565C0 → ~5.3 : 1 stays AAA.
// OpenDesign alignment (2026-06-25, §11.4.162): the brand roles
// (primary/secondary/tertiary) already matched the shared Catalogizer-Blue
// tokens; neutral.background/surface are now aligned to the SAME web tokens so
// web + phone + TV share ONE palette. error STAYS M3-tonal (#BA1A1A), NOT web
// semantic.error #DC2626 — the M3 error SURFACE role carries onError text on
// top, so it keeps the WCAG-audited M3 tonal value (web defines error only as a
// foreground token). The HELIX-175 manual audit above is superseded by the
// MECHANICAL WCAG guard in CatalogizerTVThemeTest (computes the contrast ratio
// of every text pair and asserts >= 4.5:1 AA in both schemes) — an ongoing
// §11.4.135 regression guard, not a one-time review.
@OptIn(ExperimentalTvMaterial3Api::class)
internal val TVLightColorScheme = lightColorScheme(
    primary = Color(0xFF1565C0),
    onPrimary = Color(0xFFFFFFFF),
    primaryContainer = Color(0xFFD1E4FF),
    onPrimaryContainer = Color(0xFF001D36),
    secondary = Color(0xFF535F70),
    onSecondary = Color(0xFFFFFFFF),
    secondaryContainer = Color(0xFFD7E3F7),
    onSecondaryContainer = Color(0xFF101C2B),
    tertiary = Color(0xFF6B5778),
    onTertiary = Color(0xFFFFFFFF),
    tertiaryContainer = Color(0xFFF2DAFF),
    onTertiaryContainer = Color(0xFF251431),
    error = Color(0xFFBA1A1A),
    errorContainer = Color(0xFFFFDAD6),
    onError = Color(0xFFFFFFFF),
    onErrorContainer = Color(0xFF410002),
    background = Color(0xFFFFFFFF),
    onBackground = Color(0xFF020817),
    surface = Color(0xFFFFFFFF),
    onSurface = Color(0xFF020817),
    surfaceVariant = Color(0xFFF8FAFC),
    onSurfaceVariant = Color(0xFF43474E)
)

/**
 * Catalogizer TV Material 3 theme composable applying dark or light color schemes
 * optimized for 10-foot TV viewing with the [TVTypography] scale.
 */
@OptIn(ExperimentalTvMaterial3Api::class)
@Composable
fun CatalogizerTVTheme(
    darkTheme: Boolean = true, // TV apps typically use dark theme
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) TVDarkColorScheme else TVLightColorScheme

    MaterialTheme(
        colorScheme = colorScheme,
        typography = TVTypography,
        content = content
    )
}