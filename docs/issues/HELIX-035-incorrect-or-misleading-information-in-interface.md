---
id: HELIX-035
severity: medium
category: functional
platform: 
screen: androidtv-curiosity-005.png
status: fixed
found_date: 2026-03-29
---

# Incorrect or misleading information in interface

The screenshot shows a customization screen with an error message indicating that the user's home screen cannot be updated. This is incorrect because the user can still interact with the app, and they have already made selections for their home screen. The issue here is that the error message does not accurately reflect the current state of the application.

## Related Issues

- HELIX-006: Incorrect Default Login Credentials
- HELIX-015: Potential security risk in 'Username' field
- HELIX-025: Accessibility Issues: Incomplete Form Fields
- HELIX-026: Potential UX Issue: Button Text and Icon Clarity


## Reproduction Steps

The error message appears after the user has selected elements to be added to their home screen. Reproducing this issue would involve navigating to the customization page and selecting items to add, then attempting to save or update the home screen.

## Evidence

The presence of a 'customize' button suggests that users can make selections for their home screen. The error message is visible just below this button.
