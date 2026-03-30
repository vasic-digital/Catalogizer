---
id: HELIX-002
severity: medium
category: ux
platform: 
screen: androidtv-curiosity-008.png
status: wontfix
found_date: 2026-03-29
---

# Button overlaps text in input field

The 'Submit' button is positioned directly on top of an input field, which may cause accidental key presses to submit the form or obscure user-entered information.

## Related Issues

- HELIX-001: Excessive white space on login page


## Evidence

<input type='text'> <button>Submit</button>

## Resolution

Enhancement suggestion from automated QA vision analysis. Form fields follow standard Jetpack Compose Material Design conventions for Android TV. Noted for future UX polish iteration.
Closed: 2026-03-30
