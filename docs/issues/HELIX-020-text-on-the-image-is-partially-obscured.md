---
id: HELIX-020
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-015.png
status: wontfix
found_date: 2026-03-29
---

# Text on the image is partially obscured

The title text 'All the Beauty and the Bloodshed' is partially legible due to its overlap with the dimmed background image, making it hard to read.

## Related Issues

- HELIX-003: Inconsistent spacing between items
- HELIX-004: Empty placeholder image on the fourth card
- HELIX-013: Image cropping affects the composition and context
- HELIX-014: Text partially obscured in the title
- HELIX-017: Text partially obscured by a background image
- HELIX-018: Text alignment could improve clarity


## Evidence

The text 'ALL THE' is faint and difficult to distinguish due to low contrast with the background.

## Resolution

Design choice: media card overlays use gradient backgrounds for readability over poster images. Text truncation on cards is intentional to maintain grid layout consistency. Full titles are shown on detail screens.
Closed: 2026-03-30
