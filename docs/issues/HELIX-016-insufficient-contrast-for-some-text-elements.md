---
id: HELIX-016
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-012.png
status: wontfix
found_date: 2026-03-29
---

# Insufficient contrast for some text elements

The title text 'All the Beauty and the Bloodshed' and some other UI elements have insufficient contrast against the background image, which makes them hard to read and does not comply with accessibility standards.

## Related Issues

- HELIX-007: Missing alt text for the images
- HELIX-008: Low contrast text in the movie title
- HELIX-012: Poor text contrast due to dark overlay
- HELIX-015: Insufficient color contrast between text and background


## Evidence

The title text blends into the background of the image, causing readability issues.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
