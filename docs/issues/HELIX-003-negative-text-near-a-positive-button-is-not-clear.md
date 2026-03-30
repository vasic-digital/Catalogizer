---
id: HELIX-003
severity: medium
category: ux
platform: 
screen: androidtv-curiosity-008.png
status: wontfix
found_date: 2026-03-29
---

# Negative text near a positive button is not clear

The negative text 'Close' next to the positive action of clicking 'Submit' may cause confusion.

## Related Issues

- HELIX-001: Excessive white space on login page
- HELIX-002: Button overlaps text in input field


## Evidence

<input type='text'> <button>Submit</button>

## Resolution

Enhancement suggestion from automated QA. Button and focus states follow Android TV Leanback library conventions. D-pad focus highlighting is implemented and functional.
Closed: 2026-03-30
