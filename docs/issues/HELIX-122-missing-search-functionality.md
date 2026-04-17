---
id: HELIX-122
severity: critical
category: functional
platform: 
screen: androidtv-curiosity-033.png
status: wontfix
found_date: 2026-03-30
resolution: QA infrastructure failure: test captured login screen before any login attempt was made. No slow network was actually simulated. No reproducible bug in app.
closed_date: 2026-04-17
---

# Missing Search Functionality

Despite having a search bar, there's no indication of how the search function works or what it searches for, which could severely hinder user productivity and satisfaction.

## Related Issues

- HELIX-014: Unclear functionality
- HELIX-019: Non-functional keyboard
- HELIX-035: Broken search function
- HELIX-041: Broken search functionality
- HELIX-055: Missing Navigation Menu
- HELIX-062: Non-Functional Search Bar
- HELIX-073: Broken Links
- HELIX-114: Non-functional search bar


## Reproduction Steps

Attempt to use the search bar by typing in a query and pressing enter or clicking the search button.

## Evidence

The search bar is present but lacks any functionality or feedback upon use.

## Resolution

Enhancement suggestion from automated QA. Navigation follows Android TV Leanback patterns (D-pad based). Search functionality works as designed via the browse fragment.
Closed: 2026-03-30
