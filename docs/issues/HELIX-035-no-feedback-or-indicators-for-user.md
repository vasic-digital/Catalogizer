---
id: HELIX-035
severity: critical
category: ux
platform: 
screen: androidtv-curiosity-029.png
status: wontfix
found_date: 2026-03-29
---

# No feedback or indicators for user

A lack of any loading indicators or error messages prevents the user from understanding whether the application is working or encountering an issue, leading to confusion.

## Related Issues

- HELIX-006: Lack of clear hover or focus states
- HELIX-010: Action icons lack clear labeling
- HELIX-028: No feedback or error message


## Evidence

The screen remains blank without communicating the application's state.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment). Loading, error, and connection feedback are implemented and wired: the search screen shows a `CircularProgressIndicator` and an error/empty-state block (`catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/search/SearchScreen.kt:494-525,341-369`); the login screen shows a loading spinner (`ui/screens/login/LoginScreen.kt:511-512`), an error message block (`LoginScreen.kt:440-465`), and an insecure-connection warning (`LoginScreen.kt:600-611`). A momentarily blank screenshot reflects a transient state (e.g. nothing loaded yet on an unreachable backend), not absent feedback UI.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale cited generic "Android TV conventions" boilerplate; corrected to cite the implemented loading/error/connection UI).
