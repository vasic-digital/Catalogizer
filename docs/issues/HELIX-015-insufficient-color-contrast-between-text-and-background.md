---
id: HELIX-015
severity: medium
category: accessibility
platform: 
screen: androidtv-curiosity-009.png
status: wontfix
found_date: 2026-03-29
---

# Insufficient color contrast between text and background

The white text on the bright part of the background is not easily readable due to a lack of sufficient contrast, potentially making it difficult for users with visual impairments to read.

## Related Issues

- HELIX-007: Missing alt text for the images
- HELIX-008: Low contrast text in the movie title
- HELIX-012: Poor text contrast due to dark overlay


## Evidence

The portion of the title 'All the Beauty and the Bloodshed' blends into the background, reducing readability.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
