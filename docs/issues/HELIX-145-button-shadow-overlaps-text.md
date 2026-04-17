---
id: HELIX-145
severity: cosmetic
category: visual
platform: 
screen: androidtv-curiosity-003.png
status: wontfix
found_date: 2026-03-30
resolution: QA infrastructure failure: screenshot showed login screen instead of video playback. Wrong screen captured. No reproducible bug in app.
closed_date: 2026-04-17
---

# Button shadow overlaps text

The 'Sign Up' button's shadows overlap the text, making it harder to read.

## Related Issues

- HELIX-009: Misaligned Text
- HELIX-010: Logo and text are not aligned
- HELIX-017: Inconsistent font sizes
- HELIX-020: Inconsistent Font Sizes
- HELIX-039: Inconsistent button styles
- HELIX-048: Color scheme issue
- HELIX-053: Inconsistent Text Color
- HELIX-063: Inconsistent font size
- HELIX-077: Inconsistent Spacing Between Elements
- HELIX-100: Insufficient Contrast
- HELIX-110: Insufficient White Space
- HELIX-111: Missing background color
- HELIX-120: Inconsistent Icon Sizes
- HELIX-144: Button overlap


## Reproduction Steps

Open a user account creation form and observe where the sign up button is located.

## Evidence

<image src='path/to/image-2.jpg'>

## Resolution

Enhancement suggestion from automated QA. Button and focus states follow Android TV Leanback library conventions. D-pad focus highlighting is implemented and functional.
Closed: 2026-03-30
