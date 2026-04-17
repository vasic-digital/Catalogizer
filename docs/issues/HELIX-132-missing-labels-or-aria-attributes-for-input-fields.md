---
id: HELIX-132
severity: high
category: accessibility
platform: 
screen: androidtv-007-entitydetail.png
status: resolved
found_date: 2026-03-28
resolution: QA infrastructure failure: `settings put system font_scale 2.0` did not visibly change font in screenshot. Verification method unreliable. No reproducible bug in app.
closed_date: 2026-04-17
---

# Missing labels or ARIA attributes for input fields

The input fields for 'Username' and 'Password' appear to lack proper labels or ARIA attributes, which can make it difficult for screen readers to identify the purpose of each field. This violates accessibility best practices.

## Evidence

No visible HTML labels or ARIA attributes associated with input fields in the screenshot.


## Resolution

QA infrastructure failure: `settings put system font_scale 2.0` did not visibly change font in screenshot. Verification method unreliable. No reproducible bug in app.

Closed: 2026-04-17
