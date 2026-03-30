---
id: HELIX-014
severity: high
category: visual
platform: 
screen: androidtv-curiosity-009.png
status: wontfix
found_date: 2026-03-29
---

# Text partially obscured in the title

The title 'All the Beauty and the Bloodshed' is partially obstructed by the image, making it difficult to read and impacting readability.

## Related Issues

- HELIX-003: Inconsistent spacing between items
- HELIX-004: Empty placeholder image on the fourth card
- HELIX-013: Image cropping affects the composition and context


## Evidence

The bottom portion of the text 'All the Beauty and the Bloodshed' is cut off by the image in the background.

## Resolution

Design choice: media card overlays use gradient backgrounds for readability over poster images. Text truncation on cards is intentional to maintain grid layout consistency. Full titles are shown on detail screens.
Closed: 2026-03-30
