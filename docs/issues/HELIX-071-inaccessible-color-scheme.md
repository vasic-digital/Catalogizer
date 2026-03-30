---
id: HELIX-071
severity: high
category: accessibility
platform: 
screen: androidtv-008-navigate.png
status: wontfix
found_date: 2026-03-30
---

# Inaccessible Color Scheme

The screenshot uses a color scheme that may be difficult for users with visual impairments to read.

## Related Issues

- HELIX-012: Insufficient color contrast
- HELIX-022: Insufficient Color Contrast
- HELIX-027: Inadequate Color Contrast
- HELIX-049: Accessibility issue


## Reproduction Steps

Use a color contrast analyzer tool to evaluate the color scheme.

## Evidence

The screenshot shows a dark background with light-colored text, which may not provide sufficient contrast for users with visual impairments.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
