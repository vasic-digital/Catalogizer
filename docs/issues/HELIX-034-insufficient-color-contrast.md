---
id: HELIX-034
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-009.png
status: wontfix
found_date: 2026-03-30
---

# Insufficient color contrast

The color contrast between the text and background is insufficient, which can make it difficult for users with visual impairments to read the text.

## Related Issues

- HELIX-007: Insufficient Color Contrast
- HELIX-019: Inaccessible Color Scheme


## Reproduction Steps

Use a color contrast analyzer tool to evaluate the color contrast.

## Evidence

The color contrast ratio is less than 4.5:1.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
