---
id: HELIX-026
severity: medium
category: visual
platform: 
screen: androidtv-curiosity-002.png
status: wontfix
found_date: 2026-03-30
---

# Inadequate Contrast Between Field Text and Background

The server connection field has white text on a black background, which may cause difficulty for users with visual impairments to read the text.

## Related Issues

- HELIX-005: Inconsistent Text Color
- HELIX-018: Inconsistent Font Styles
- HELIX-024: Inadequate Contrast Between Text and Background
- HELIX-025: Inadequate Contrast Between Button Text and Background


## Reproduction Steps

N/A

## Evidence

The screenshot shows the server connection field with white text on a black background.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
