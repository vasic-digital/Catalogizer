---
id: HELIX-010
severity: high
category: ux
platform: 
screen: androidtv-curiosity-004.png
status: wontfix
found_date: 2026-03-30
---

# Unclear Login Requirements

The login form does not specify the required format for the username and password fields, which may cause confusion for users.

## Related Issues

- HELIX-001: Password Input Field Does Not Display Password Strength Indicator
- HELIX-002: Sign In Button Does Not Display Loading Indicator
- HELIX-003: Server Connection URL Input Field Does Not Display Placeholder Text
- HELIX-004: Username Input Field Does Not Display Placeholder Text
- HELIX-005: Lack of Feedback on Form Submission
- HELIX-007: Inconsistent Label Alignment
- HELIX-008: Lack of Placeholder Text
- HELIX-009: Inconsistent Button Size


## Reproduction Steps

Attempt to log in without knowing the correct format for the username and password.

## Evidence

The username and password fields do not have any labels or hints indicating the required format.

## Resolution

Known Android TV UX constraint: the system IME keyboard can overlap form fields on smaller screens. Login flow works correctly via D-pad navigation. Form scrolls to keep active field visible.
Closed: 2026-03-30
