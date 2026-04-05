---
id: HELIX-340
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-001-after.png
status: open
found_date: 2026-04-05
---

# Text input cursor is incorrectly displayed on the virtual keyboard instead of the search input field.

The white vertical cursor, which indicates the current typing position, is visible on the period key of the virtual keyboard. It should logically be within the 'Search' input field, at the end of the text 'mov cs..', to allow the user to continue typing. This prevents the user from knowing where their input will appear and suggests a fundamental disconnect between the keyboard input and the text field.

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
- HELIX-328: Virtual keyboard highlights '8' key while input field ends with '3'
- HELIX-330: Input field content does not reflect active keyboard key press.
- HELIX-334: Potential input processing issue: Excessive repeated characters from single key press/hold.
- HELIX-335: Keyboard key highlight does not correspond to search bar input


## Reproduction Steps

N/A (observed state in screenshot)

## Evidence

A white vertical bar (cursor) is positioned over the period key on the virtual keyboard, while the search input field shows 'mov cs..' without a cursor.
