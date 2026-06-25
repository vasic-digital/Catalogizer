# OpenDesign Phone Palette — Catalogizer Android (Catalogizer-Blue)

**Revision:** 1
**Last modified:** 2026-06-25T00:00:00Z
**Scope:** `catalogizer-android` (Android PHONE app — NOT `catalogizer-androidtv`)
**Mandate:** §11.4.162 (OpenDesign UI design system)

## Summary

The Android phone app's color palette is now an OpenDesign-token-structured
light + dark palette branded **Catalogizer-Blue**, kept consistent with the
web and desktop clients.

- New file: `app/src/main/java/com/catalogizer/android/ui/theme/Color.kt` —
  the OpenDesign token palette (single source of truth for phone colors).
- `Theme.kt` now consumes those tokens and maps them onto the Material 3
  `ColorScheme` roles (light + dark schemes + dynamic-color on Android 12+).
- `ThemeTest.kt` updated to the new token values.

## Brand-consistency source (§11.4.6 — values are not invented)

The canonical brand reference is the web/desktop OpenDesign token module landed
in commit **e748bba5**:

> `catalog-web/src/styles/tokens.ts` — the "Catalogizer-Blue" OpenDesign token
> module, itself traced to `docs/design/OPENDESIGN_INTEGRATION_AUDIT.md §3`.

Every brand-defining role in the phone palette is copied verbatim from that
module so all three clients (web, desktop, phone) share one palette. Per
`docs/design/OPENDESIGN_CONSUMABILITY_VERDICT.md`, `nexu-io/open-design` is not
npm/Gradle-consumable (it ships no token package), so — exactly as web and
desktop did — the phone mirrors OpenDesign's token *taxonomy* (color.brand /
color.semantic / color.neutral, each with light + dark variants) in hand-authored
Kotlin rather than importing a package.

## Token → Material 3 mapping (light / dark)

Brand-defining roles — **identical to web tokens.ts**:

| OpenDesign token            | M3 role(s)                | Light     | Dark      |
|-----------------------------|---------------------------|-----------|-----------|
| color.brand.primary         | primary                   | `#1565C0` | `#9ECAFF` |
| color.brand.onPrimary       | onPrimary                 | `#FFFFFF` | `#003258` |
| color.brand.secondary       | secondary                 | `#535F70` | `#BBC7DB` |
| color.brand.onSecondary     | onSecondary               | `#FFFFFF` | `#253140` |
| color.brand.accent          | tertiary                  | `#6B5778` | `#D6BEE4` |
| color.brand.onAccent        | onTertiary                | `#FFFFFF` | `#3B2948` |
| color.semantic.error        | error                     | `#DC2626` | `#EF4444` |
| color.neutral.background    | background                | `#FFFFFF` | `#020817` |
| color.neutral.foreground    | onBackground / onSurface  | `#020817` | `#F8FAFC` |
| color.neutral.surface       | surface (dark) / surfaceVariant (light) | `#F8FAFC` (light surfaceVariant) | `#0F172A` (dark surface) |
| color.neutral.border        | outline                   | `#E2E8F0` | `#1E293B` |

Material-3-only roles — **not present in the web token model** (shadcn/Tailwind
has no "container" concept); kept from the platform M3 tonal palette generated
around the Catalogizer-Blue source hues (`#1565C0` primary, `#6B5778` tertiary)
so the phone retains proper M3 elevation + container semantics:

| M3 role                   | Light     | Dark      |
|---------------------------|-----------|-----------|
| primaryContainer          | `#D1E4FF` | `#00497D` |
| onPrimaryContainer        | `#001D36` | `#D1E4FF` |
| secondaryContainer        | `#D7E3F7` | `#3B4858` |
| onSecondaryContainer      | `#101C2B` | `#D7E3F7` |
| tertiaryContainer         | `#F2DAFF` | `#523F5F` |
| onTertiaryContainer       | `#251431` | `#F2DAFF` |
| errorContainer            | `#FFDAD6` | `#93000A` |
| onError                   | `#FFFFFF` | `#690005` |
| onErrorContainer          | `#410002` | `#FFDAD6` |
| onSurfaceVariant          | `#43474E` | `#C3C7CF` |

## Light + dark both defined; labels never overlay (§11.4.162)

- Both `lightColorScheme(...)` and `darkColorScheme(...)` are fully populated.
- Every background/container role is paired with an explicit `on*` foreground so
  text/labels always have a defined, sufficient-contrast color and never collide
  with or overlay an undefined surface.
- Dark mode uses a distinct elevated `surface` (`#0F172A`) above the deeper
  `background` (`#020817`) — matching the web token model where
  `color.neutral.surface != color.neutral.background` — so cards/sheets read as
  raised layers instead of blending into the background.

## What changed vs the previous inline palette

| Role            | Old (inline Theme.kt) | New (Catalogizer-Blue) | Reason |
|-----------------|-----------------------|------------------------|--------|
| light primary   | `#1976D2`             | `#1565C0`              | web `brand.primary.light` (old value was web's `primaryHover`) |
| light error     | `#BA1A1A`             | `#DC2626`              | web `semantic.error.light` |
| dark error      | `#FFB4AB`             | `#EF4444`              | web `semantic.error.dark` |
| light background| `#FDFCFF`             | `#FFFFFF`              | web `neutral.background.light` |
| light onBackground/onSurface | `#1A1C1E` | `#020817`           | web `neutral.foreground.light` |
| light surfaceVariant | `#DFE2EB`        | `#F8FAFC`              | web `neutral.surface.light` |
| light outline   | `#73777F`             | `#E2E8F0`              | web `neutral.border.light` |
| dark background | `#101214`             | `#020817`              | web `neutral.background.dark` |
| dark onBackground/onSurface | `#E2E2E6` | `#F8FAFC`            | web `neutral.foreground.dark` |
| dark surface    | `#101214` (== bg)     | `#0F172A` (elevated)   | web `neutral.surface.dark` (distinct from bg) |
| dark outline    | `#8D9199`             | `#1E293B`              | web `neutral.border.dark` |

Unchanged brand roles (already matched web): secondary, tertiary/accent, all
`on*` brand foregrounds, and the M3 container tonal pairs.

## Honest gaps

- OpenDesign is not consumed as a Gradle dependency (upstream ships no
  consumable token package — see `OPENDESIGN_CONSUMABILITY_VERDICT.md`); the
  token taxonomy is mirrored in `Color.kt`, consistent with how web/desktop
  handled the same constraint.
- The M3 `*Container` / `onError` / `onSurfaceVariant` tonal roles have no web
  counterpart (the web model lacks a container concept); they are retained from
  the platform Material 3 tonal palette around the brand hues, not from a web
  token. This is a deliberate platform-fit gap, not a brand divergence.
- This change authors source only. Build, on-device run, and visual-regression
  capture are performed by the conductor (this subagent does not build or run a
  device).
