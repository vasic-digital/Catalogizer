---
id: HELIX-019
severity: critical
category: accessibility
platform: 
screen: androidtv-009-performance.png
status: wontfix
found_date: 2026-03-30
---

# Inaccessible Color Scheme

The color scheme used in the interface is not accessible for users with visual impairments, as it does not provide sufficient contrast between the background and text.

## Related Issues

- HELIX-007: Insufficient Color Contrast


## Reproduction Steps

Observe the color scheme used in the interface.

## Evidence

The background color is too similar to the text color, making it difficult to read.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (TV Material 3 color schemes), NOT a Leanback app — there is no browse fragment. The dark theme sets `onBackground`/`onSurface` to `#E2E2E6` on `background`/`surface` `#101214` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/theme/Theme.kt:33-36`). The in-source WCAG audit comment (`Theme.kt:42-46`) documents a computed ~13:1 ratio (AAA) for `onBackground`-on-`background` and `onSurface`-on-`surface` — well above the WCAG 2.1 AA 4.5:1 floor. The "background too similar to text" claim is contradicted by the actual color values.

Runtime verification still outstanding (NEEDS-RUNTIME): pixel/contrast-ratio runtime proof on the rendered screen. `ui/theme/ThemeTest.kt` only asserts proxy color properties (e.g. `onBackground.red > 0.8f`, `background.red < 0.1f`), NOT a computed WCAG contrast ratio. Closed wontfix on code evidence; a rendered-pixel contrast measurement remains to be captured.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
