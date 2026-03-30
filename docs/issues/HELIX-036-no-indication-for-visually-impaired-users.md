---
id: HELIX-036
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-029.png
status: wontfix
found_date: 2026-03-29
---

# No indication for visually impaired users

The lack of visible content or auditory feedback makes the application completely inaccessible for visually impaired users using assistive technologies.

## Related Issues

- HELIX-007: Missing alt text for the images
- HELIX-008: Low contrast text in the movie title
- HELIX-012: Poor text contrast due to dark overlay
- HELIX-015: Insufficient color contrast between text and background
- HELIX-016: Insufficient contrast for some text elements
- HELIX-021: Low contrast between text and background
- HELIX-024: No visible keyboard or screen reader focus indicators
- HELIX-029: No readable text or navigable content
- HELIX-032: No accessible feedback provided


## Evidence

No contrast, elements, or signals are present for accessibility tools or screen readers to function.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
