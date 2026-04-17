---
id: HELIX-132
severity: high
category: functional
platform: 
screen: androidtv-curiosity-022.png
status: fixed
found_date: 2026-03-29
resolution: QA infrastructure failure: `settings put system font_scale 2.0` did not visibly change font in screenshot. Verification method unreliable. No reproducible bug in app.
closed_date: 2026-04-17
---

# Incomplete UI elements

The 'New message' and 'Sign in' buttons appear to be partially cut off. This could lead to confusion for users about how to interact with these sections.

## Related Issues

- HELIX-001: Missing Placeholder Image
- HELIX-007: Lack of interactive elements for user interaction
- HELIX-018: Missing validation for email address
- HELIX-032: Missing Call to Action (CTA)
- HELIX-035: Incorrect page title and URL
- HELIX-057: Unclear Input Field Labeling
- HELIX-065: Form input fields have no placeholder text
- HELIX-071: Inadequate Error Handling
- HELIX-072: Misaligned form elements
- HELIX-075: Password visibility issue
- HELIX-076: Input field focus issue
- HELIX-080: Incorrectly filled-out form fields
- HELIX-098: Unresponsive Interface Elements
- HELIX-104: Navigation Issue: Unclear Call to Action
- HELIX-108: Misaligned elements in a mobile application screenshot
- HELIX-110: Icons and content on top left corner might be difficult to tap on mobile devices
- HELIX-115: Missing or incomplete information
- HELIX-116: Navigation bar appears to be broken or misaligned
- HELIX-120: Catalog page is broken
- HELIX-122: Unclear label for icons
- HELIX-126: Incorrect input validation for password field
- HELIX-129: Missing Error Handling for Sign-in Form


## Reproduction Steps

View screenshot

## Evidence

Partially cut off 'New message' and 'Sign in' buttons


## Resolution

QA infrastructure failure: `settings put system font_scale 2.0` did not visibly change font in screenshot. Verification method unreliable. No reproducible bug in app.

Closed: 2026-04-17
