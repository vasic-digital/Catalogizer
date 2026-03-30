---
id: HELIX-054
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-015.png
status: wontfix
found_date: 2026-03-30
---

# Inadequate color contrast

The text and background colors have insufficient contrast, potentially causing readability issues for users with visual impairments.

## Related Issues

- HELIX-006: Insufficient Color Contrast
- HELIX-051: Lack of ARIA Attributes for Screen Readers


## Reproduction Steps

Evaluate the color contrast using accessibility tools or guidelines.

## Evidence

The dark gray text on a black background does not meet accessibility standards.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
