package com.catalogizer.android.ui.theme

import androidx.compose.ui.graphics.Color

/**
 * Catalogizer Blue — OpenDesign-token color palette for the Android PHONE app
 * (§11.4.162 OpenDesign UI design system mandate).
 *
 * Single source of truth for the phone client's brand palette (light + dark).
 * The OpenDesign token *taxonomy* (color.brand / color.semantic / color.neutral
 * roles, each with light + dark variants) is mirrored here and mapped onto the
 * Material 3 ColorScheme roles in Theme.kt.
 *
 * BRAND-CONSISTENCY SOURCE (§11.4.6 no-guessing — values are NOT invented):
 *   catalog-web/src/styles/tokens.ts  (commit e748bba5, "Catalogizer-Blue"
 *   OpenDesign token module, itself traced to docs/design/
 *   OPENDESIGN_INTEGRATION_AUDIT.md §3). The web/desktop clients are the
 *   canonical brand reference. The brand-role hexes below are copied verbatim
 *   from that module so all three clients share one palette:
 *
 *     OpenDesign role            web token (tokens.ts)        phone hex
 *     color.brand.primary        light #1565C0 / dark #9ECAFF  (identical)
 *     color.brand.primaryHover   light #1976D2 / dark #7FB4F5  (identical)
 *     color.brand.onPrimary      light #FFFFFF / dark #003258  (identical)
 *     color.brand.secondary      light #535F70 / dark #BBC7DB  (identical)
 *     color.brand.onSecondary    light #FFFFFF / dark #253140  (identical)
 *     color.brand.accent         light #6B5778 / dark #D6BEE4  (identical, M3 "tertiary")
 *     color.brand.onAccent       light #FFFFFF / dark #3B2948  (light identical; dark = M3-standard onTertiary tonal, NOT web onAccent.dark #251431 which is an onTertiaryContainer-class deep tone unsuited as a foreground)
 *     color.semantic.error       M3-tonal light #BA1A1A / dark #FFB4AB
 *       (NOT web #DC2626/#EF4444 — those are FOREGROUND tokens; on the M3
 *        error SURFACE they fail WCAG AA, dark dropping to 3.48:1)
 *     color.neutral.background   light #FFFFFF / dark #020817  (identical)
 *     color.neutral.surface      light #F8FAFC / dark #0F172A  (identical)
 *     color.neutral.foreground   light #020817 / dark #F8FAFC  (identical)
 *     color.neutral.border       light #E2E8F0 / dark #1E293B  (identical)  (M3 "outline")
 *
 * Material-3-only roles that the web token module does NOT define — the *Container
 * and on*Container tonal pairs, plus onError / onSurfaceVariant — are kept from
 * the platform Material 3 tonal palette generated around the brand hues (the
 * standard M3 "blue source #1565C0 / mauve tertiary #6B5778" tonal set), so the
 * phone keeps proper M3 elevation + container semantics while the *brand-defining*
 * roles match web/desktop exactly. These are intentionally not in the web tokens
 * because the web (shadcn/Tailwind) model has no container concept.
 *
 * Light + dark are BOTH fully defined (§11.4.162). on-* foreground colors are
 * paired with every background/container role so text/labels never overlay an
 * insufficient-contrast surface and elements never collide.
 */

// region OpenDesign color.brand.primary  (web: light #1565C0 / dark #9ECAFF)
val CatalogizerLightPrimary = Color(0xFF1565C0)
val CatalogizerLightOnPrimary = Color(0xFFFFFFFF)
val CatalogizerLightPrimaryContainer = Color(0xFFD1E4FF)
val CatalogizerLightOnPrimaryContainer = Color(0xFF001D36)

val CatalogizerDarkPrimary = Color(0xFF9ECAFF)
val CatalogizerDarkOnPrimary = Color(0xFF003258)
val CatalogizerDarkPrimaryContainer = Color(0xFF00497D)
val CatalogizerDarkOnPrimaryContainer = Color(0xFFD1E4FF)
// endregion

// region OpenDesign color.brand.secondary  (web: light #535F70 / dark #BBC7DB)
val CatalogizerLightSecondary = Color(0xFF535F70)
val CatalogizerLightOnSecondary = Color(0xFFFFFFFF)
val CatalogizerLightSecondaryContainer = Color(0xFFD7E3F7)
val CatalogizerLightOnSecondaryContainer = Color(0xFF101C2B)

val CatalogizerDarkSecondary = Color(0xFFBBC7DB)
val CatalogizerDarkOnSecondary = Color(0xFF253140)
val CatalogizerDarkSecondaryContainer = Color(0xFF3B4858)
val CatalogizerDarkOnSecondaryContainer = Color(0xFFD7E3F7)
// endregion

// region OpenDesign color.brand.accent -> Material 3 "tertiary"  (web: light #6B5778 / dark #D6BEE4)
val CatalogizerLightTertiary = Color(0xFF6B5778)
val CatalogizerLightOnTertiary = Color(0xFFFFFFFF)
val CatalogizerLightTertiaryContainer = Color(0xFFF2DAFF)
val CatalogizerLightOnTertiaryContainer = Color(0xFF251431)

val CatalogizerDarkTertiary = Color(0xFFD6BEE4)
val CatalogizerDarkOnTertiary = Color(0xFF3B2948)
val CatalogizerDarkTertiaryContainer = Color(0xFF523F5F)
val CatalogizerDarkOnTertiaryContainer = Color(0xFFF2DAFF)
// endregion

// region OpenDesign color.semantic.error  (web: light #DC2626 / dark #EF4444)
// error SURFACE role stays M3-tonal (#BA1A1A), NOT web semantic.error #DC2626:
// the web token is a FOREGROUND token; on the M3 error SURFACE (onError text on
// top) #DC2626 is borderline and the dark #EF4444 dropped to 3.48:1 (< WCAG AA).
// Mirrors the TV client; proven by ThemeTest's contrast oracle.
val CatalogizerLightError = Color(0xFFBA1A1A)
val CatalogizerLightOnError = Color(0xFFFFFFFF)
val CatalogizerLightErrorContainer = Color(0xFFFFDAD6)
val CatalogizerLightOnErrorContainer = Color(0xFF410002)

val CatalogizerDarkError = Color(0xFFFFB4AB)
val CatalogizerDarkOnError = Color(0xFF690005)
val CatalogizerDarkErrorContainer = Color(0xFF93000A)
val CatalogizerDarkOnErrorContainer = Color(0xFFFFDAD6)
// endregion

// region OpenDesign color.neutral.background / surface / foreground / border
// (web: bg light #FFFFFF / dark #020817 ; surface light #F8FAFC / dark #0F172A ;
//  foreground light #020817 / dark #F8FAFC ; border light #E2E8F0 / dark #1E293B)
val CatalogizerLightBackground = Color(0xFFFFFFFF)
val CatalogizerLightOnBackground = Color(0xFF020817)
val CatalogizerLightSurface = Color(0xFFFFFFFF)
val CatalogizerLightOnSurface = Color(0xFF020817)
val CatalogizerLightSurfaceVariant = Color(0xFFF8FAFC)
val CatalogizerLightOnSurfaceVariant = Color(0xFF43474E)
val CatalogizerLightOutline = Color(0xFFE2E8F0)

val CatalogizerDarkBackground = Color(0xFF020817)
val CatalogizerDarkOnBackground = Color(0xFFF8FAFC)
val CatalogizerDarkSurface = Color(0xFF0F172A)
val CatalogizerDarkOnSurface = Color(0xFFF8FAFC)
val CatalogizerDarkSurfaceVariant = Color(0xFF43474E)
val CatalogizerDarkOnSurfaceVariant = Color(0xFFC3C7CF)
val CatalogizerDarkOutline = Color(0xFF1E293B)
// endregion
