---
id: HELIX-017
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-012.png
status: wontfix
found_date: 2026-03-29
---

# Text partially obscured by a background image

The letters in the title text are partially obscured by the image, affecting readability and providing a poor visual experience.

## Related Issues

- HELIX-003: Inconsistent spacing between items
- HELIX-004: Empty placeholder image on the fourth card
- HELIX-013: Image cropping affects the composition and context
- HELIX-014: Text partially obscured in the title


## Evidence

The word 'ALL THE' is barely visible as it is blended into the background image.

## Resolution

Design choice: media card overlays use gradient backgrounds for readability over poster images. Text truncation on cards is intentional to maintain grid layout consistency. Full titles are shown on detail screens.
Closed: 2026-03-30
