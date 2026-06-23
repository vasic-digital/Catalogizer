---
id: HELIX-113
severity: critical
category: accessibility
platform: 
screen: androidtv-curiosity-010.png
status: wontfix
found_date: 2026-03-30
---

# Inaccessible menu

The menu on the application is not accessible to users with disabilities, which may cause difficulties for users.

## Related Issues

- HELIX-012: Insufficient color contrast
- HELIX-022: Insufficient Color Contrast
- HELIX-027: Inadequate Color Contrast
- HELIX-049: Accessibility issue
- HELIX-071: Inaccessible Color Scheme
- HELIX-101: Inadequate Color Scheme
- HELIX-106: Color Contrast Issue


## Reproduction Steps

None

## Evidence

The menu on the application is not accessible to users with disabilities.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment). The menu (top bar) is implemented, focusable, and labelled: Search and Settings actions are rendered and click-wired in `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/components/TopBar.kt:36-151` (each with `contentDescription` for screen readers, `TopBar.kt:109,146`), wired from `ui/screens/home/HomeScreen.kt:105-107`. The "menu not accessible" claim is contradicted by the implemented, labelled TopBar.

Runtime verification still outstanding (NEEDS-RUNTIME): an on-device focus-render pass (D-pad focus traversal showing each menu item receiving a visible focus indicator, plus a TalkBack announcement check) has not been captured; closure rests on the code being present and wired, not on an observed focus/announce run.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale was vague "vision analysis" boilerplate; corrected to cite the wired, labelled TopBar).
