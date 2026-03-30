---
id: HELIX-013
severity: low
category: visual
platform: 
screen: androidtv-curiosity-006.png
status: wontfix
found_date: 2026-03-29
---

# Image cropping affects the composition and context

The background image appears cropped in a way that may not effectively complement or add context to the movie being showcased.

## Related Issues

- HELIX-003: Inconsistent spacing between items
- HELIX-004: Empty placeholder image on the fourth card


## Evidence

The image primarily highlights a person's back, which may not be visually rich or informative.

## Resolution

Design choice: media card overlays use gradient backgrounds for readability over poster images. Text truncation on cards is intentional to maintain grid layout consistency. Full titles are shown on detail screens.
Closed: 2026-03-30
