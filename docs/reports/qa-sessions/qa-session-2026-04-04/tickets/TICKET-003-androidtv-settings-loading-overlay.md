# [MINOR] Android TV: Settings screen visible behind loading spinner

**Platform**: Android TV (Mi Box 192.168.0.214:5555)
**Screen**: Settings / Home
**Severity**: MINOR
**Discovered by**: HelixQA autonomous session (curiosity-016 screenshot)

## Description

The settings screen (Auto Play Next Episode, Streaming Quality, Subtitle Settings) is partially visible behind the "Loading your media collection..." spinner overlay. Settings controls appear functional even during loading.

## Reproduction Steps

1. Login and navigate to Settings
2. While media collection is still loading, observe the settings screen

## Expected Behavior

Either the loading overlay should be fully opaque, or the settings should only be accessible after loading completes.

## Actual Behavior

Settings controls are visible and interactive behind a semi-transparent loading overlay.

## Evidence

- Screenshot: `androidtv-curiosity-016.png`

## Fix Suggestion

Make the loading overlay fully opaque or disable settings interaction until loading completes.
