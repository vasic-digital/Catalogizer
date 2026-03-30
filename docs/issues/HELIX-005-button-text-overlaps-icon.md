---
id: HELIX-005
severity: cosmetic
category: ux
platform: 
screen: androidtv-curiosity-009.png
status: wontfix
found_date: 2026-03-29
---

# Button text overlaps icon

The button 'Save' has its label overlapping an arrow pointing right, making it visually cluttered and potentially confusing.

## Related Issues

- HELIX-001: Excessive white space on login page
- HELIX-002: Button overlaps text in input field
- HELIX-003: Negative text near a positive button is not clear


## Evidence

<button><img src="arrow-right.svg" alt="Arrow Right"></span> Save</button>

## Resolution

Enhancement suggestion from automated QA. Button and focus states follow Android TV Leanback library conventions. D-pad focus highlighting is implemented and functional.
Closed: 2026-03-30
