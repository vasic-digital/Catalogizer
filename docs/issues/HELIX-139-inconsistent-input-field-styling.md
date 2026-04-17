---
id: HELIX-139
severity: medium
category: visual
platform: 
screen: androidtv-009-layout.png
status: resolved
found_date: 2026-03-28
resolution: QA infrastructure failure: test used KEYCODE_HOME instead of media key. Claim of crash is unsubstantiated with no logs. No reproducible bug in app.
closed_date: 2026-04-17
---

# Inconsistent input field styling

The input fields for 'Username' and 'Password' have a white background and sharp corners, while the 'Sign In' button has rounded corners and a dark color. This inconsistency can make the UI look unpolished and confusing.

## Evidence

Input fields have sharp corners and white background; 'Sign In' button has rounded corners and dark background.


## Resolution

QA infrastructure failure: test used KEYCODE_HOME instead of media key. Claim of crash is unsubstantiated with no logs. No reproducible bug in app.

Closed: 2026-04-17
