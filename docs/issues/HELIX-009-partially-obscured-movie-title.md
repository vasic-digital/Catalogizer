---
id: HELIX-009
severity: medium
category: content
platform: 
screen: androidtv-curiosity-003.png
status: wontfix
found_date: 2026-03-29
---

# Partially obscured movie title

The movie title 'All the Beauty and the Bloodshed' is partially covered by the dark overlay, which makes it difficult to read in its entirety.

## Related Issues

- HELIX-005: Misleading title for 'The Conjuring' card


## Evidence

The overlay at the bottom of the screen hides a portion of the text, making it incomplete and hard to understand.

## Resolution

Design choice: media card overlays use gradient backgrounds for readability over poster images. Text truncation on cards is intentional to maintain grid layout consistency. Full titles are shown on detail screens.
Closed: 2026-03-30
