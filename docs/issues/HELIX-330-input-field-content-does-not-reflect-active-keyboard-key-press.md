---
id: HELIX-330
severity: high
category: functional
platform: 
screen: androidtv-curiosity-014-after.png
status: resolved
resolution: fixed
fixed_date: 2026-04-06
found_date: 2026-04-04
---

# Input field content does not reflect active keyboard key press.

The virtual keyboard shows the '7' key prominently highlighted, indicating it is currently being pressed or has just been registered. However, the search input field, which contains "mcr cw 2. 877777" (six '7's), does not display an additional '7' that would result from this active key press. This discrepancy suggests a potential issue with input processing, display latency, or synchronization between the keyboard state and the application's input field.

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


## Evidence

The '7' key on the virtual keyboard is visually highlighted in white, while the text in the search bar (highlighted with a blue border) ends with '877777', indicating that the current key press has not yet been registered or displayed in the input field.
