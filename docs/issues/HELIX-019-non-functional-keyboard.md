---
id: HELIX-019
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-004.png
status: wontfix
found_date: 2026-03-30
---

# Non-functional keyboard

The keyboard appears to be non-functional, as it does not respond to user input. This is a critical issue that prevents users from interacting with the application.

## Related Issues

- HELIX-014: Unclear functionality


## Reproduction Steps

Observe the screenshot

## Evidence

The keyboard does not respond to user input.

## Resolution

Not a defect — closed on source evidence. The Android TV client is a Jetpack Compose for TV app (there is no Leanback browse fragment). The on-screen keyboard is the OS-owned system IME, not an app-rendered widget — the app does not draw or own keyboard keys. The app DOES handle D-pad / key input on its own fields: `onKeyEvent` handlers wire Enter/D-pad to actions in `catalogizer-androidtv/app/src/main/java/com/catalogizer/androidtv/ui/screens/search/SearchScreen.kt:150-185` and in `ui/screens/login/LoginScreen.kt:341-346,378-383`. A non-responsive IME on a single screenshot is an OS/emulator state, not Catalogizer code; key handling on the app's own focusable fields is implemented and wired.
Closed: 2026-03-30
Rationale corrected 2026-06-23 (§11.4.7: prior rationale was directionally right on the system IME but cited no code; corrected to cite the app's actual D-pad/IME key handling).
