---
id: HELIX-007
severity: low
category: ux
platform: 
screen: androidtv-curiosity-002.png
status: wontfix
found_date: 2026-03-30
---

# Inconsistent Label Alignment

The labels for the username and password fields are not aligned consistently, which can make the form look cluttered and difficult to read.

## Related Issues

- HELIX-001: Password Input Field Does Not Display Password Strength Indicator
- HELIX-002: Sign In Button Does Not Display Loading Indicator
- HELIX-003: Server Connection URL Input Field Does Not Display Placeholder Text
- HELIX-004: Username Input Field Does Not Display Placeholder Text
- HELIX-005: Lack of Feedback on Form Submission


## Reproduction Steps

Compare the alignment of the labels for the username and password fields.

## Evidence

The label for the username field is aligned to the left, while the label for the password field is aligned to the right.

## Resolution

Enhancement suggestion from automated QA vision analysis. Form fields follow standard Jetpack Compose Material Design conventions for Android TV. Noted for future UX polish iteration.
Closed: 2026-03-30
