---
id: HELIX-100
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-008.png
status: wontfix
found_date: 2026-03-30
---

# Insufficient Contrast

The text in the top-left corner is light gray on a black background, which may be difficult to read for some users. This is particularly problematic for users with visual impairments.

## Related Issues

- HELIX-009: Misaligned Text
- HELIX-010: Logo and text are not aligned
- HELIX-017: Inconsistent font sizes
- HELIX-020: Inconsistent Font Sizes
- HELIX-039: Inconsistent button styles
- HELIX-048: Color scheme issue
- HELIX-053: Inconsistent Text Color
- HELIX-063: Inconsistent font size
- HELIX-077: Inconsistent Spacing Between Elements


## Reproduction Steps

Open the application and observe the text in the top-left corner.

## Evidence

The screenshot shows light gray text on a black background.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
