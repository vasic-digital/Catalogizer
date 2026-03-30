---
id: HELIX-008
severity: low
category: ux
platform: 
screen: androidtv-007-entitydetail.png
status: wontfix
found_date: 2026-03-30
---

# Inconsistent Button Design

The 'Sign In' button has a different design compared to the 'Server Connection' button, which may cause user confusion and affect the overall user experience.

## Related Issues

- HELIX-001: Insecure Password Input
- HELIX-002: Lack of Feedback on Form Submission
- HELIX-003: Inconsistent Label Alignment
- HELIX-004: Lack of Placeholder Text
- HELIX-005: Inconsistent Button Styling
- HELIX-006: Password field does not have a show password option
- HELIX-007: Username and password fields do not have labels


## Reproduction Steps

Observe the design of the two buttons on the login page.

## Evidence

The 'Sign In' button is rounded and has a darker background color, while the 'Server Connection' button is rectangular and has a lighter background color.

## Resolution

Enhancement suggestion from automated QA. Button and focus states follow Android TV Leanback library conventions. D-pad focus highlighting is implemented and functional.
Closed: 2026-03-30
