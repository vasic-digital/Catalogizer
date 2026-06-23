---
id: HELIX-073
severity: critical
category: functional
platform: 
screen: androidtv-008-navigate.png
status: wontfix
found_date: 2026-03-30
---

# Broken Links

The screenshot contains broken links that do not function as expected.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality
- HELIX-055: Missing Navigation Menu
- HELIX-062: Non-Functional Search Bar


## Reproduction Steps

Click on the links and observe the error messages.

## Evidence

The screenshot shows several links with error messages indicating that they are broken.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment), and it has no web links — there is nothing to "break". In-app navigation is `NavHost` routes, all wired: Login/Home/Search/MediaDetail/Player/Settings/Category (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/navigation/TVNavigation.kt:61-171`). The only external link is "Forgot Password", which is guarded (opens the reset page and surfaces an error if it cannot, `ui/screens/login/LoginScreen.kt:411-430`). The "broken links with error messages" claim does not correspond to any hyperlink construct in the TV UI.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale was vague "vision analysis" boilerplate; corrected to cite the NavHost routes and the absence of web links).
