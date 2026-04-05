---
id: HELIX-348
severity: medium
category: accessibility
platform: 
screen: androidtv-curiosity-005-after.png
status: open
found_date: 2026-04-05
---

# Refresh button's background lacks sufficient contrast with the page background.

The 'Refresh' button has a dark gray background that provides insufficient contrast against the overall dark page background. If the button background is similar to #4D4D4D and the page background is #181818, the contrast ratio is approximately 2.4:1. This is below the WCAG 2.1 AA requirement of 3:1 for non-text components (like button boundaries that indicate a clickable area), which can make the button less discernible as an interactive element for all users, especially those with low vision.

## Related Issues

- HELIX-012: Insufficient color contrast
- HELIX-022: Insufficient Color Contrast
- HELIX-027: Inadequate Color Contrast
- HELIX-049: Accessibility issue
- HELIX-071: Inaccessible Color Scheme
- HELIX-101: Inadequate Color Scheme
- HELIX-106: Color Contrast Issue
- HELIX-113: Inaccessible menu
- HELIX-174: Inaccessible Keyboard Navigation
- HELIX-177: Keyboard is not accessible for users with disabilities
- HELIX-307: Low contrast for unselected category labels
- HELIX-322: Search button fails WCAG contrast requirements.
- HELIX-339: Insufficient contrast for suggestion words in keyboard
- HELIX-344: Insufficient contrast for 'Search' button text.
- HELIX-347: Insufficient contrast for secondary instructional text.


## Evidence

The 'Refresh' button element, specifically its dark gray background.
