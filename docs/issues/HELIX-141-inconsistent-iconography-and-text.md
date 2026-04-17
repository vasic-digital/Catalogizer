---
id: HELIX-141
severity: high
category: visual
platform: 
screen: androidtv-curiosity-030.png
status: fixed
found_date: 2026-03-29
resolution: QA infrastructure failure: screenshot showed app home screen, indicating test never reached playback. Wrong screen captured. No reproducible bug in app.
closed_date: 2026-04-17
---

# Inconsistent Iconography and Text

The screenshot shows a registration form with two text fields labeled 'Email' and 'Password'. The 'Email' field has an icon that visually represents an email address, while the 'Password' field does not have a corresponding icon. This inconsistency in iconography can confuse users and lead to misunderstandings about the type of input required for each field.

## Related Issues

- HELIX-004: Text overlapping with background
- HELIX-012: Incorrect use of capitalization
- HELIX-015: Button label misalignment
- HELIX-053: Ineffective contrast on button labels
- HELIX-054: Inconsistent Text Formatting
- HELIX-068: Inconsistent Typography
- HELIX-074: Inconsistent button alignment
- HELIX-083: Inconsistent spacing around input fields
- HELIX-100: Text Legibility Issues
- HELIX-105: Visual Clutter: Busy Interface
- HELIX-109: Color contrast in a mobile application screenshot
- HELIX-119: Text shadows are not used consistently
- HELIX-127: Inconsistent Typography in Sign-in Form
- HELIX-131: Inefficient use of space


## Reproduction Steps

View registration form

## Evidence

Form fields with inconsistent icons


## Resolution

QA infrastructure failure: screenshot showed app home screen, indicating test never reached playback. Wrong screen captured. No reproducible bug in app.

Closed: 2026-04-17
