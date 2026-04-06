---
id: HELIX-335
severity: medium
category: functional
platform: 
screen: androidtv-curiosity-019-after.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Keyboard key highlight does not correspond to search bar input

The '6' key on the virtual keyboard is highlighted, indicating it might be currently pressed or was the last key interacted with. However, the text in the search bar, 'mcu_chG', does not contain the character '6'. This inconsistency creates confusion about the current input state and whether user input is being registered correctly, potentially leading to incorrect entries or frustration.

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


## Reproduction Steps

1. Open the search interface.
2. Type 'mcu_chG'.
3. Observe the virtual keyboard state. The '6' key is highlighted without a corresponding '6' in the search input.

## Evidence

The '6' key on the virtual keyboard has a distinct white background highlight, while the search bar text 'mcu_chG' does not include the number '6'. The cursor is positioned after the 'G', but the 'G' key is not highlighted.
