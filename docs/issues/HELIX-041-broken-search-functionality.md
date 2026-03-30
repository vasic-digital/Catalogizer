---
id: HELIX-041
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-001.png
status: wontfix
found_date: 2026-03-30
---

# Broken search functionality

The search functionality is not working as expected, which may prevent users from finding what they are looking for.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function


## Reproduction Steps

Enter a search query and press the 'Search' button.

## Evidence

The search results are not displayed, and an error message is shown instead.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
