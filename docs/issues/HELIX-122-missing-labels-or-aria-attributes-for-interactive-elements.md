---
id: HELIX-122
severity: medium
category: accessibility
platform: 
screen: androidtv-005-layout.png
status: resolved
found_date: 2026-03-28
resolution: QA infrastructure failure: test captured login screen before any login attempt was made. No slow network was actually simulated. No reproducible bug in app.
closed_date: 2026-04-17
---

# Missing labels or ARIA attributes for interactive elements

The 'Discover' and 'Connect' buttons lack clear labels or ARIA attributes to assist screen readers in identifying their purpose, which can hinder accessibility for visually impaired users.

## Evidence

No visible ARIA labels or additional context for the Discover and Connect buttons.


## Resolution

QA infrastructure failure: test captured login screen before any login attempt was made. No slow network was actually simulated. No reproducible bug in app.

Closed: 2026-04-17
