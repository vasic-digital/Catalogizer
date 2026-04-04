# [MINOR] Android TV: Entity title truncated on detail screen

**Platform**: Android TV (Mi Box 192.168.0.214:5555)
**Screen**: Media Detail
**Severity**: MINOR
**Discovered by**: HelixQA autonomous session screenshot analysis

## Description

On the media detail screen, the title "Lucky Luke The" appears truncated. The full title should be "Lucky Luke: The Adventures" or similar but is cut off without an ellipsis indicator.

## Reproduction Steps

1. Navigate to a media item with a long title
2. Open the detail screen

## Expected Behavior

Long titles should either wrap to a second line or show ellipsis (...) to indicate truncation.

## Actual Behavior

Title is abruptly cut off: "Lucky Luke The" with no indication of truncation.

## Evidence

- Screenshots: `androidtv-curiosity-004.png`, `androidtv-curiosity-012.png`

## Fix Suggestion

Add `maxLines = 2` with `overflow = TextOverflow.Ellipsis` to the title `Text` composable in the detail screen.
