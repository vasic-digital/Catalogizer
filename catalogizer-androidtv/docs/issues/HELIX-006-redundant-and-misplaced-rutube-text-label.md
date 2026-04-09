---
id: HELIX-006
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-017-after.png
status: open
found_date: 2026-04-08
---

# Redundant and misplaced 'RUTUBE' text label

The 'RUTUBE' app is listed under 'Favorite Apps' with its icon and a text label 'RUTUBE' directly below it. However, an additional 'RUTUBE' text label is visible further down and to the left of the app row, floating without clear association to any UI element. This creates visual clutter and appears to be a rendering error or an unintentional duplicate, negatively impacting the clean presentation of the interface.

## Related Issues

- HELIX-003: App name text truncated or too close to bounding box.
- HELIX-004: Inconsistent spacing and sizing of app icons in 'Favorite Apps' row.


## Reproduction Steps

Observe the 'Favorite Apps' section on the Home screen, specifically around the 'RUTUBE' app icon.

## Evidence

The screenshot shows the 'RUTUBE' app icon with its text label directly beneath it, and a separate, unaligned 'RUTUBE' text string positioned below and to the left of the app row.
