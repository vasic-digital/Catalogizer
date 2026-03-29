---
id: HELIX-129
severity: medium
category: functional
platform: 
screen: androidtv-curiosity-030.png
status: fixed
found_date: 2026-03-29
---

# Incorrect data validation for Catalog number input field

The 'Catalog number' input field should enforce that the user enters a valid number format, such as 000-000 or 123456789. In the screenshot, no such validation seems to be in place.

## Related Issues

- HELIX-006: Incorrect Default Login Credentials
- HELIX-015: Potential security risk in 'Username' field
- HELIX-025: Accessibility Issues: Incomplete Form Fields
- HELIX-026: Potential UX Issue: Button Text and Icon Clarity
- HELIX-035: Incorrect or misleading information in interface
- HELIX-036: Lack of clear, concise instructions or guidance
- HELIX-040: Unclear call-to-action in settings menu
- HELIX-066: Potential security vulnerabilities
- HELIX-080: Password visibility
- HELIX-085: Incomplete password change flow
- HELIX-120: Input field error message placement


## Reproduction Steps

Navigate to the 'Create new catalog' form and attempt to enter an invalid catalog number format.

## Evidence

The input field is allowing non-numeric characters or any other format, which suggests a lack of proper data validation.
