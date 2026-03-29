---
id: HELIX-126
severity: low
category: functional
platform: 
screen: androidtv-curiosity-009.png
status: fixed
found_date: 2026-03-29
---

# Incorrect input validation for password field

The input box for the password does not show an indication that it is a password field. This can create confusion and may not be immediately clear to users that this field requires a password.

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


## Reproduction Steps

Click on 'Sign in with a different account'. Enter an email address and attempt to create a new account.

## Evidence

The password input box lacks common visual indicators such as asterisks or icons to represent a password field, which are typically used to inform users that the entered text is hidden for privacy reasons. This could lead to confusion or misunderstanding regarding the type of input expected by the application.
