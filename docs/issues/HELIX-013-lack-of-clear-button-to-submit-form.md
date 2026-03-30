---
id: HELIX-013
severity: high
category: ux
platform: 
screen: androidtv-005-layout.png
status: wontfix
found_date: 2026-03-30
---

# Lack of clear button to submit form

The form is missing a clear button to submit the form, which can cause confusion for users and make it difficult for them to complete the login process.

## Related Issues

- HELIX-001: Unclear Form Field Labels
- HELIX-002: Missing Password Masking
- HELIX-003: Lack of Form Validation Feedback
- HELIX-004: Inconsistent Button Styling
- HELIX-006: Lack of Feedback on Form Submission
- HELIX-008: Inconsistent Input Field Lengths
- HELIX-009: Lack of Clear Error Messages
- HELIX-010: Insufficient Feedback on Form Submission
- HELIX-011: Missing Password Visibility Toggle
- HELIX-012: Unclear Server Connection Status


## Reproduction Steps

Attempt to login without a clear submit button.

## Evidence

The form does not have a clear submit button.

## Resolution

Enhancement suggestion from automated QA. Button and focus states follow Android TV Leanback library conventions. D-pad focus highlighting is implemented and functional.
Closed: 2026-03-30
