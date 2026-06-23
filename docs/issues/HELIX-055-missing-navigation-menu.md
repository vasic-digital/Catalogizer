---
id: HELIX-055
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-007.png
status: wontfix
found_date: 2026-03-30
---

# Missing Navigation Menu

There is no navigation menu visible on the page, which means users cannot easily navigate to other parts of the application.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality


## Reproduction Steps

None

## Evidence

There is no navigation menu visible on the page.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app, NOT a Leanback app — there is no browse fragment. Navigation is present via the top bar: Search and Settings actions are rendered and click-wired in `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/components/TopBar.kt:36-151`, and wired into the home screen at `ui/screens/home/HomeScreen.kt:105-107` (`onSearchClick`/`onSettingsClick` → NavHost routes). The "no navigation menu" claim is contradicted by the implemented, wired TopBar.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited a nonexistent Leanback browse fragment).
