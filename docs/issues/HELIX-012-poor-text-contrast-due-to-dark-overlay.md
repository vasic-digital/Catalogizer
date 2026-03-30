---
id: HELIX-012
severity: medium
category: accessibility
platform: 
screen: androidtv-curiosity-006.png
status: wontfix
found_date: 2026-03-29
---

# Poor text contrast due to dark overlay

The text partially overlaps with the background image and the overlay, resulting in poor contrast which may be difficult for users with visual impairments to read.

## Related Issues

- HELIX-007: Missing alt text for the images
- HELIX-008: Low contrast text in the movie title


## Evidence

The text over the person's back is not easily distinguishable for readability due to low contrast.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
