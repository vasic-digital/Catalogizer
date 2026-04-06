---
id: HELIX-339
severity: low
category: accessibility
platform: 
screen: androidtv-001-loginform.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-05
---

# Insufficient contrast for suggestion words in keyboard

The text for the suggested words ('with', 'to', 'in', 'like') in the keyboard's suggestion bar appears to have low contrast against their background. This can make the words difficult to read for users with visual impairments or in bright lighting conditions, hindering usability and accessibility.

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


## Reproduction Steps

1. Open the search screen.
2. Focus on the search input field to bring up the keyboard.
3. Observe the contrast of the suggestion words in the top bar of the keyboard.

## Evidence

In the suggestion bar of the on-screen keyboard, the words 'with', 'to', 'in', 'like' are light grey on a slightly darker grey background, exhibiting poor contrast compared to the main keyboard keys.
