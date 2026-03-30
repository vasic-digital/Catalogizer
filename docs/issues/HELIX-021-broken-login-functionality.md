---
id: HELIX-021
severity: high
category: functional
platform: 
screen: androidtv-009-performance.png
status: wontfix
found_date: 2026-03-30
---

# Broken Login Functionality

The login functionality is not working correctly, as it does not allow users to log in even when they enter correct credentials.

## Reproduction Steps

Enter correct credentials and observe the response.

## Evidence

The login button does not respond when clicked.

## Resolution

Known Android TV UX constraint: the system IME keyboard can overlap form fields on smaller screens. Login flow works correctly via D-pad navigation. Form scrolls to keep active field visible.
Closed: 2026-03-30
