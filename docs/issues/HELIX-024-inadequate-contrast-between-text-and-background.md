---
id: HELIX-024
severity: critical
category: visual
platform: 
screen: androidtv-curiosity-002.png
status: wontfix
found_date: 2026-03-30
---

# Inadequate Contrast Between Text and Background

The username and password fields have white text on a black background, which may cause difficulty for users with visual impairments to read the text.

## Related Issues

- HELIX-005: Inconsistent Text Color
- HELIX-018: Inconsistent Font Styles


## Reproduction Steps

N/A

## Evidence

The screenshot shows the username and password fields with white text on a black background.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (TV Material 3 color schemes), NOT a Leanback app — there is no browse fragment. The login fields render with theme `onSurface`/`onBackground` `#E2E2E6` on `background`/`surface` `#101214` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/theme/Theme.kt:33-36`). The in-source WCAG audit comment (`Theme.kt:42-46`) documents a computed ~13:1 ratio (AAA) — well above the WCAG 2.1 AA 4.5:1 floor. "White text on black" reading difficulty is contradicted by the actual near-13:1 light-on-dark contrast.

Runtime verification still outstanding (NEEDS-RUNTIME): pixel/contrast-ratio runtime proof on the rendered login fields. `ui/theme/ThemeTest.kt` only asserts proxy color properties (e.g. `onBackground.red > 0.8f`, `background.red < 0.1f`), NOT a computed WCAG contrast ratio. Closed wontfix on code evidence; a rendered-pixel contrast measurement remains to be captured.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
