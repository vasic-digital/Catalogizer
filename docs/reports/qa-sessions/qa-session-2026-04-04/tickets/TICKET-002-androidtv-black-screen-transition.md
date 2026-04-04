# [MAJOR] Android TV: Black screen during app transition

**Platform**: Android TV (Mi Box 192.168.0.214:5555)
**Screen**: Transition between login and home
**Severity**: MAJOR
**Discovered by**: HelixQA autonomous session (curiosity-001 screenshot)

## Description

A completely black screen appears during transition from login to the home screen. The screenshot `androidtv-curiosity-001.png` shows a fully black frame with no visible content.

## Reproduction Steps

1. Launch Catalogizer app on Android TV
2. Login with valid credentials
3. Observe the screen immediately after login

## Expected Behavior

Smooth transition with either a loading indicator or immediate home screen display.

## Actual Behavior

Fully black screen captured between login and home screen.

## Evidence

- Screenshot: `androidtv-curiosity-001.png` (completely black)
- This may be a brief transition frame, but it should show a loading state instead.

## Fix Suggestion

Add a splash/loading screen during the authentication → home transition. Ensure `setContent` is called before any async operations complete.
