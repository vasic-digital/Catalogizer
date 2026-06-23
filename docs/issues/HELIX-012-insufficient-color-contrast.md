---
id: HELIX-012
severity: critical
category: accessibility
platform: 
screen: androidtv-curiosity-001.png
status: wontfix
found_date: 2026-03-30
---

# Insufficient color contrast

The color contrast between the background and text is insufficient, which can make it difficult for users with visual impairments to read the content.

## Reproduction Steps

None

## Evidence

The background is dark gray and the text is light gray, which does not provide sufficient contrast.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (TV Material 3 color schemes), NOT a Leanback app — there is no browse fragment. The dark theme sets `onBackground`/`onSurface` to `#E2E2E6` on `background`/`surface` `#101214` (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/theme/Theme.kt:33-36`). The in-source WCAG audit comment (`Theme.kt:42-46`) documents a computed ~13:1 ratio (AAA) for `onBackground`-on-`background` and `onSurface`-on-`surface` — well above the WCAG 2.1 AA 4.5:1 floor. The "light gray on dark gray is insufficient" claim is contradicted by the actual color values.

Runtime verification still outstanding (NEEDS-RUNTIME): pixel/contrast-ratio runtime proof on the rendered screen. `ui/theme/ThemeTest.kt` only asserts proxy color properties (e.g. `onBackground.red > 0.8f`, `background.red < 0.1f`), NOT a computed WCAG contrast ratio. Closed wontfix on code evidence; a rendered-pixel contrast measurement remains to be captured.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
