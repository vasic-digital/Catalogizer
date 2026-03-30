---
id: HELIX-106
severity: low
category: accessibility
platform: 
screen: androidtv-curiosity-038.png
status: wontfix
found_date: 2026-03-30
---

# Color Contrast Issue

The color contrast between the text and background may not be sufficient for users with visual impairments.

## Related Issues

- HELIX-012: Insufficient color contrast
- HELIX-022: Insufficient Color Contrast
- HELIX-027: Inadequate Color Contrast
- HELIX-049: Accessibility issue
- HELIX-071: Inaccessible Color Scheme
- HELIX-101: Inadequate Color Scheme


## Reproduction Steps

Use a color contrast analyzer tool to evaluate the color scheme.

## Evidence

The text color is light gray, and the background color is dark gray, which may not meet the minimum contrast ratio recommended by accessibility guidelines.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
