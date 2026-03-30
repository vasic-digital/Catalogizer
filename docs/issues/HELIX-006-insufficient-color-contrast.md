---
id: HELIX-006
severity: medium
category: accessibility
platform: 
screen: androidtv-curiosity-002.png
status: wontfix
found_date: 2026-03-30
---

# Insufficient Color Contrast

The color contrast between the text and background is insufficient, making it difficult for users with visual impairments to read the text.

## Reproduction Steps

Use a color contrast analyzer tool to measure the contrast ratio.

## Evidence

The text color is a light gray (#ccc) on a dark gray (#333) background, resulting in a low contrast ratio.

## Resolution

Accessibility enhancement suggestion from automated QA. The app uses Material Design 3 color system which meets WCAG 2.1 AA contrast ratios for the primary theme. Advanced accessibility features are tracked in the product backlog.
Closed: 2026-03-30
