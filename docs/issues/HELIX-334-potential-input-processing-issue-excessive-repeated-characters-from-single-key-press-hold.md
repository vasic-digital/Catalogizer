---
id: HELIX-334
severity: medium
category: functional
platform: 
screen: androidtv-curiosity-018.png
status: open
found_date: 2026-04-04
---

# Potential input processing issue: Excessive repeated characters from single key press/hold.

The search input field displays an unusually long sequence of the digit '7' (seven '7's) following 'mcv cw 2. 8'. Concurrently, the '7' key on the virtual keyboard is highlighted, indicating active interaction. This could suggest a functional defect where a single or brief press of the '7' key results in multiple '7's being registered, or that holding the key down generates characters at an excessively rapid rate, making it difficult for users to input single characters accurately. This can lead to incorrect search queries and a frustrating user experience.

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


## Reproduction Steps

1. Navigate to the search screen where the virtual keyboard is displayed.
2. Type 'mcv cw 2. 8'.
3. Lightly tap the '7' key on the virtual keyboard once, or hold it for a very brief duration (e.g., less than 0.5 seconds).
4. Observe the number of '7's that appear in the search bar. (Expected: a single '7' or a controlled, slower repetition rate if holding the key).

## Evidence

The text 'mcv cw 2. 8777777' visible in the search bar, combined with the prominent highlighting of the '7' key on the virtual keyboard.
