---
id: HELIX-021
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-015.png
status: wontfix
found_date: 2026-03-29
---

# Low contrast between text and background

The rating and other text elements blend into the dimmed background, making them hard to read for users with visual impairments.

## Related Issues

- HELIX-007: Missing alt text for the images
- HELIX-008: Low contrast text in the movie title
- HELIX-012: Poor text contrast due to dark overlay
- HELIX-015: Insufficient color contrast between text and background
- HELIX-016: Insufficient contrast for some text elements


## Evidence

The yellow rating score (7.2) is difficult to read against the dark and uneven background.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
