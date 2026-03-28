---
id: HELIX-248
severity: medium
category: accessibility
platform: 
screen: androidtv-curiosity-006.png
status: resolved
found_date: 2026-03-28
---

# Missing screen reader support for loading state

There is no ARIA label, alt text, or screen reader announcement to indicate that the app is in a loading state. This makes the app less accessible to users who rely on assistive technologies.

## Reproduction Steps

Use a screen reader to navigate the loading screen.

## Evidence

No audible or programmatic feedback is provided to indicate loading status.
