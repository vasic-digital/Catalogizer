---
id: HELIX-328
severity: medium
category: functional
platform: 
screen: androidtv-curiosity-009-after.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Virtual keyboard highlights '8' key while input field ends with '3'

The virtual keyboard shows the '8' key highlighted, which typically indicates the last character pressed or the currently focused key. However, the search input field displays 'mcr cw 2 3', which clearly ends with the digit '3'. This inconsistency is misleading and suggests a functional bug in how the keyboard's visual state is synchronized with the input field's content. Users might be confused about which key was actually registered or if the keyboard is responsive.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality
- HELIX-055: Missing Navigation Menu
- HELIX-062: Non-Functional Search Bar
- HELIX-073: Broken Links
- HELIX-114: Non-functional search bar
- HELIX-122: Missing Search Functionality
- HELIX-153: No search results displayed
- HELIX-299: Critical movie detail content is missing
- HELIX-320: Application displays a completely black screen


## Reproduction Steps

1. Open the application and access the search bar. 
2. Type the characters 'm', 'c', 'r', ' ', 'c', 'w', ' ', '2', '3'.
3. Observe the highlighted key on the virtual keyboard after the input 'mcr cw 2 3' is displayed in the search bar. Expected: '3' or no key should be highlighted, as '3' was the last character entered. Actual: The '8' key is highlighted.

## Evidence

The search input field displaying 'mcr cw 2 3' and simultaneously, the '8' key on the virtual keyboard being highlighted with a white background.
