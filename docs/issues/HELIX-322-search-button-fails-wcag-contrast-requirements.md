---
id: HELIX-322
severity: high
category: accessibility
platform: 
screen: androidtv-curiosity-002-after.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Search button fails WCAG contrast requirements.

Due to the extremely low contrast between the 'Search' button text and its background, the button is not accessible to users with visual impairments. This likely violates WCAG (Web Content Accessibility Guidelines) contrast ratio requirements for readable text.

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


## Reproduction Steps

Observe the search button element on the top right of the screen and try to read its label.

## Evidence

The 'Search' button in the top right corner, showing text that is nearly invisible due to low contrast.
