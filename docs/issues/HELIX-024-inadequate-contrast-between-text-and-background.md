---
id: HELIX-024
severity: critical
category: visual
platform: 
screen: androidtv-curiosity-002.png
status: wontfix
found_date: 2026-03-30
---

# Inadequate Contrast Between Text and Background

The username and password fields have white text on a black background, which may cause difficulty for users with visual impairments to read the text.

## Related Issues

- HELIX-005: Inconsistent Text Color
- HELIX-018: Inconsistent Font Styles


## Reproduction Steps

N/A

## Evidence

The screenshot shows the username and password fields with white text on a black background.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
