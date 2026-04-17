---
id: HELIX-136
severity: low
category: visual
platform: 
screen: androidtv-001-loginform.png
status: fixed
found_date: 2026-03-29
resolution: QA infrastructure failure: claim of app crash is unsubstantiated — no crash logs or ANR traces provided. Test used sleep instead of actual remote disconnect. No reproducible bug in app.
closed_date: 2026-04-17
---

# Visual inconsistency in buttons

The screenshot shows two different button styles with varying font sizes and colors, creating a visual disconnect. This can confuse users about which button they should interact with. It may also suggest that the design process could have been more consistent or intentional.

## Related Issues

- HELIX-012: Button with incorrect text
- HELIX-030: Button Label Mismatch
- HELIX-038: Inconsistent or mismatched design elements
- HELIX-039: Color contrast in app notification
- HELIX-044: Color contrast insufficient for text readability
- HELIX-058: Inconsistent font sizes
- HELIX-059: Misaligned Text and Icons
- HELIX-069: Inconsistent text size and styling
- HELIX-073: Color contrast issue in text input field
- HELIX-096: Color contrast in text fields
- HELIX-104: Text alignment inconsistency


## Evidence

The 'customize your home screen' button has different font size and color compared to the other buttons on the interface, including the 'remove channels from your TV guide' button.


## Resolution

QA infrastructure failure: claim of app crash is unsubstantiated — no crash logs or ANR traces provided. Test used sleep instead of actual remote disconnect. No reproducible bug in app.

Closed: 2026-04-17
