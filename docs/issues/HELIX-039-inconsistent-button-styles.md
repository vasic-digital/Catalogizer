---
id: HELIX-039
severity: low
category: visual
platform: 
screen: androidtv-curiosity-001.png
status: wontfix
found_date: 2026-03-30
---

# Inconsistent button styles

The buttons in the top-right corner have different styles, which may cause visual inconsistency and make it harder for users to understand their purpose.

## Related Issues

- HELIX-009: Misaligned Text
- HELIX-010: Logo and text are not aligned
- HELIX-017: Inconsistent font sizes
- HELIX-020: Inconsistent Font Sizes


## Reproduction Steps

None

## Evidence

The 'Search' button has a magnifying glass icon, while the 'Settings' button has a gear icon. The 'Search' button is also slightly larger than the 'Settings' button.

## Resolution

Enhancement suggestion from automated QA. Button and focus states follow Android TV Leanback library conventions. D-pad focus highlighting is implemented and functional.
Closed: 2026-03-30
